package domain

import (
	"encoding/json"
	"reflect"
	"slices"
	"testing"
)

func TestTheOriginOfADeviceIsTheBoardThatAnnouncedItAndNothingElse(t *testing.T) {
	t.Parallel()
	g := newMixedRig(t)
	g.bringUpMixed()
	for device, want := range map[string]Origin{"P1": OriginGateway, "V1": OriginGateway, "S1": OriginSimulated, "S2": OriginSimulated, "ghost-9": OriginUnknown, "": OriginUnknown} {
		if got := g.relay.OriginOf(device); got != want {
			t.Errorf("OriginOf(%q) = %d, want %d", device, got, want)
		}
	}
}

func TestACommandIsRoutedByTheBoardOfItsDeviceAndOnlyThatWay(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, device       string
		down, toSim, up    []string
		unknown, forwarded int
	}{
		{"a hardware device goes to the gateway untouched", "P1", []string{"DCMD"}, nil, nil, 0, 0},
		{"a simulated device goes to the simulated world", "S1", nil, []string{"DCMD"}, nil, 0, 1},
		{"a device on no board is refused as unknown_device", "ghost-9", nil, nil, []string{"DRECEIPT"}, 1, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := newMixedRig(t)
			g.bringUpMixed()
			moves := g.hostSends(commandFor(tc.device, "c-1"))
			if !slices.Equal(typesOf(moves.Down), tc.down) || !slices.Equal(typesOf(moves.ToSim), tc.toSim) || !slices.Equal(typesOf(moves.Up), tc.up) {
				t.Fatalf("down %v to the world %v up %v, want %v %v %v", typesOf(moves.Down), typesOf(moves.ToSim), typesOf(moves.Up), tc.down, tc.toSim, tc.up)
			}
			if g.relay.Stats().UnknownDevice != tc.unknown {
				t.Fatalf("stats = %+v, want %d refusals as unknown_device", g.relay.Stats(), tc.unknown)
			}
			for _, routed := range append(slices.Clone(moves.Down), moves.ToSim...) {
				sameContent(t, routed, g.hostSent)
			}
			if tc.forwarded == 1 && (moves.ToSim[0]["session_id"] != "sim-1" || moves.ToSim[0]["seq"] != uint64(1)) {
				t.Fatalf("command to the world = %v, want the relay's own session with seq 1 after the STATE", moves.ToSim[0])
			}
		})
	}
}

func TestACommandForAnUnknownDeviceIsRefusedAsTheNodeWould(t *testing.T) {
	t.Parallel()
	g := newMixedRig(t)
	g.bringUpMixed()
	receipt := g.hostSends(commandFor("ghost-9", "c-7")).Up[0]
	for member, want := range map[string]any{"command_id": "c-7", "status": "refused", "reason": "unknown_device", "boot_id": "none"} {
		if receipt[member] != want {
			t.Errorf("%s = %v, want %v", member, receipt[member], want)
		}
	}
	if moves := g.hostSends(Record{Type: "DQUERY", Lane: LaneControl, Fields: map[string]any{"device": "ghost-9"}}); len(moves.Up)+len(moves.Down)+len(moves.ToSim) != 0 {
		t.Fatalf("a query for an unknown device moved %v, want it dropped: nothing answers a query for a device that is not there", moves)
	}
}

func TestAQueryGoesToTheWorldOfTheDevice(t *testing.T) {
	t.Parallel()
	g := newMixedRig(t)
	g.bringUpMixed()
	if moves := g.hostSends(Record{Type: "DQUERY", Lane: LaneControl, Fields: map[string]any{"device": "S2"}}); len(moves.Down) != 0 || !slices.Equal(typesOf(moves.ToSim), []string{"DQUERY"}) {
		t.Fatalf("a query for S2: down %v to the world %v, want it with the world alone", typesOf(moves.Down), typesOf(moves.ToSim))
	}
	if moves := g.hostSends(Record{Type: "DQUERY", Lane: LaneControl, Fields: map[string]any{"device": "V1"}}); len(moves.ToSim) != 0 || !slices.Equal(typesOf(moves.Down), []string{"DQUERY"}) {
		t.Fatalf("a query for V1: down %v to the world %v, want it with the gateway alone", typesOf(moves.Down), typesOf(moves.ToSim))
	}
}

func TestNothingTheSimulatedWorldDoesCanReachTheGateway(t *testing.T) {
	t.Parallel()
	if _, present := reflect.TypeFor[SimMoves]().FieldByName("Down"); present {
		t.Fatal("SimMoves has a Down member: the answer to a simulated record could then be a record for the gateway")
	}
	g := newMixedRig(t)
	g.bringUpMixed()
	before := g.relay.Stats()
	for _, record := range []Record{{Type: "NBIRTH", Lane: LaneControl, Fields: simulatedBirthRecord(t, "s-1", []string{"S1", "S2"}, birthKinds()).Fields},
		simDeviceBirth("S1"), DeviceData("S1", "sim-boot", false, []map[string]any{Sample(1, 3.5, "", 1)}), DeviceState("S1", "sim-boot", []any{}),
		Receipt("c-1", "accepted", "", "sim-boot"), Result("c-1", "succeeded", "sim-boot"), Sync(3, 1), NodeDeath(0, "shutdown")} {
		moves := g.simSends(record)
		if len(moves.Trace) > 0 && moves.Trace[0].Direction == TraceRelay {
			t.Fatalf("%s from the world was traced as sent to the gateway", record.Type)
		}
	}
	tickOnce(g.relay)
	after := g.relay.Stats()
	if after.Down != before.Down || after.Up == before.Up {
		t.Fatalf("down %d -> %d up %d -> %d, want the world's records to count up and nothing down", before.Down, after.Down, before.Up, after.Up)
	}
}

func TestTheWorldKeepsReachingTheHostWhenTheGatewayIsGone(t *testing.T) {
	t.Parallel()
	g := newMixedRig(t)
	g.bringUpMixed()
	g.simSends(Record{Type: "NBIRTH", Lane: LaneControl, Fields: simulatedBirthRecord(t, "s-1", []string{"S1", "S2"}, birthKinds()).Fields})
	g.relay.GatewayLost()
	down := g.relay.Stats().Down
	moves := g.simSends(simDeviceBirth("S1"))
	if !slices.Equal(typesOf(moves.Up), []string{"DBIRTH"}) || g.relay.Stats().Down != down {
		t.Fatalf("up %v down %d -> %d, want the birth carried to the host and nothing for the lost gateway", typesOf(moves.Up), down, g.relay.Stats().Down)
	}
}

func TestARecordIsCarriedOnlyFromTheSideThatOwnsTheDevice(t *testing.T) {
	t.Parallel()
	g := newMixedRig(t)
	g.bringUpMixed()
	g.simSends(Record{Type: "NBIRTH", Lane: LaneControl, Fields: simulatedBirthRecord(t, "s-1", []string{"S1", "S2"}, birthKinds()).Fields})
	dropped := g.relay.Stats().Dropped
	if moves := g.gw(simDeviceBirth("S1")); len(moves.Up) != 0 {
		t.Fatalf("up = %v, want a gateway birth of a simulated device id dropped, or the host would see two devices called S1", typesOf(moves.Up))
	}
	if moves := g.simSends(deviceBirthOf("P1", "boot-1")); len(moves.Up) != 0 {
		t.Fatalf("up = %v, want a simulated birth of a gateway device id dropped", typesOf(moves.Up))
	}
	if got := g.relay.Stats().Dropped; got != dropped+2 {
		t.Fatalf("dropped %d, want %d: both impostors counted", got, dropped+2)
	}
}

func TestTheWorldsRecordsWaitForTheMergedBirthAndAreSequencedByTheRelay(t *testing.T) {
	t.Parallel()
	early := newMixedRig(t)
	if moves := early.simSends(simDeviceBirth("S1")); len(moves.Up) != 0 || early.relay.Stats().Dropped != 1 {
		t.Fatalf("up = %v dropped %d, want a record from the world dropped while no host session has the merged birth", typesOf(moves.Up), early.relay.Stats().Dropped)
	}
	g := newMixedRig(t)
	g.bringUpMixed()
	g.simSends(Record{Type: "NBIRTH", Lane: LaneControl, Fields: simulatedBirthRecord(t, "s-1", []string{"S1", "S2"}, birthKinds()).Fields})
	birth := g.simSends(simDeviceBirth("S1"))
	data := g.simSends(DeviceData("S1", "sim-boot", false, []map[string]any{Sample(1, 3.5, "", 1)}))
	if len(birth.Up) != 1 || len(data.Up) != 1 || data.Up[0]["seq"].(uint64) != birth.Up[0]["seq"].(uint64)+1 || data.Up[0]["session_id"] != "relay-h1" {
		t.Fatalf("birth %v data %v, want consecutive seq in the relay's own host session", birth.Up, data.Up)
	}
}

func TestABarrierTheWorldRaisesIsAnsweredToTheWorldAndNeverToTheGateway(t *testing.T) {
	t.Parallel()
	g := newMixedRig(t)
	g.bringUpMixed()
	g.simSends(Record{Type: "NBIRTH", Lane: LaneControl, Fields: simulatedBirthRecord(t, "s-1", []string{"S1", "S2"}, birthKinds()).Fields})
	g.simSends(simDeviceBirth("S1"))
	g.simSends(DeviceData("S1", "sim-boot", false, []map[string]any{Sample(1, 3.5, "", 1)}))
	sync := g.simSends(Sync(2, 1767225660000000000))
	if len(sync.Up) != 1 || sync.Up[0]["type"] != "NSYNC" || sync.Up[0]["through_seq"] != uint64(4) || sync.Up[0]["seq"] != uint64(5) {
		t.Fatalf("up = %v, want the world's NSYNC in the relay's sequence with through_seq 4 (the relay's seq of the world's seq 2)", sync.Up)
	}
	answer := g.hostSends(hostSync(5))
	if len(answer.Down) != 0 || len(answer.ToSim) != 1 || answer.ToSim[0]["type"] != "HSYNC" || answer.ToSim[0]["through_seq"] != uint64(3) {
		t.Fatalf("down %v to the world %v, want HSYNC through_seq 3 (the world's own seq of its NSYNC) for the world alone", answer.Down, answer.ToSim)
	}
}

func TestABarrierTheGatewayRaisesIsAnsweredToTheGatewayAndNeverToTheWorld(t *testing.T) {
	t.Parallel()
	g := newMixedRig(t)
	g.bringUpMixed()
	g.gw(validData())
	g.gw(barrierSync(3))
	answer := g.hostSends(hostSync(4))
	if len(answer.ToSim) != 0 || len(answer.Down) != 1 || answer.Down[0]["type"] != "HSYNC" || answer.Down[0]["through_seq"] != uint64(4) {
		t.Fatalf("down %v to the world %v, want HSYNC through_seq 4 (the gateway's seq of its NSYNC) for the gateway alone", answer.Down, answer.ToSim)
	}
}

func TestAnAnswerNoBarrierWaitsForIsDropped(t *testing.T) {
	t.Parallel()
	g := newMixedRig(t)
	g.bringUpMixed()
	dropped := g.relay.Stats().Dropped
	if moves := g.hostSends(hostSync(2)); len(moves.Down)+len(moves.ToSim) != 0 || g.relay.Stats().Dropped != dropped+1 {
		t.Fatalf("down %v to the world %v dropped %d, want a stray HSYNC counted and carried nowhere", moves.Down, moves.ToSim, g.relay.Stats().Dropped)
	}
}

func TestWhenTheWorldEndsItsBornDevicesAreReportedOfflineInOrder(t *testing.T) {
	t.Parallel()
	g := newMixedRig(t)
	g.bringUpMixed()
	g.simSends(Record{Type: "NBIRTH", Lane: LaneControl, Fields: simulatedBirthRecord(t, "s-1", []string{"S1", "S2"}, birthKinds()).Fields})
	g.simSends(simDeviceBirth("S2"))
	g.simSends(simDeviceBirth("S1"))
	ended := g.simSends(NodeDeath(0, "shutdown"))
	if types := typesOf(ended.Up); !slices.Equal(types, []string{"DDEATH", "DDEATH"}) || ended.Up[0]["device"] != "S1" || ended.Up[1]["device"] != "S2" || ended.Up[0]["reason"] != "offline" {
		t.Fatalf("up = %v, want DDEATH offline for S1 and then S2 and no NDEATH of the world's own", ended.Up)
	}
	if again := g.simSends(NodeDeath(0, "shutdown")); len(again.Up) != 0 {
		t.Fatalf("up = %v, want nothing to die twice", again.Up)
	}
}

func TestTheWorldIsAskedToBirthAgainWheneverTheHostSessionIsNewOrTheGatewayBirthsAgain(t *testing.T) {
	t.Parallel()
	g := newMixedRig(t)
	g.bringUpMixed()
	rebirth := g.gw(hardwareBirth(t, nil))
	if !slices.Equal(typesOf(rebirth.Up), []string{"NBIRTH"}) || !slices.Equal(typesOf(rebirth.ToSim), []string{"NCMD"}) || rebirth.ToSim[0]["command"] != "rebirth" {
		t.Fatalf("up %v to the world %v, want the merged rebirth and then the world's own rebirth request", typesOf(rebirth.Up), typesOf(rebirth.ToSim))
	}
	g.relay.HostLost()
	g.relay.HostAttached()
	g.newHostSession("relay-h2")
	again := g.hostSends(hostState(7, 0))
	if !slices.Equal(typesOf(again.ToSim), []string{"STATE", "NCMD"}) {
		t.Fatalf("to the world %v, want the new host's STATE and then a rebirth request so that its devices are born on the new session", typesOf(again.ToSim))
	}
}

func TestAHostGoodbyeIsPassedToTheWorldSoItsSessionEnds(t *testing.T) {
	t.Parallel()
	g := newMixedRig(t)
	g.bringUpMixed()
	goodbye := hostState(7, 0)
	goodbye.Fields["online"] = false
	moves := g.hostSends(goodbye)
	if len(moves.ToSim) != 1 || moves.ToSim[0]["online"] != false || moves.ToSim[0]["authority_epoch"] != json.Number("7") {
		t.Fatalf("to the world = %v, want the host's goodbye, which ends the world's session", moves.ToSim)
	}
}
