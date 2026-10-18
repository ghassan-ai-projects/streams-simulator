package domain

import (
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func simulatedBirthRecord(t *testing.T, board string, members []string, kinds []any) Record {
	t.Helper()
	record, err := NodeBirthOf(NodeIdentity{Environment: EnvironmentSimulated, ProtocolMax: 1, PriorLastSeq: -1}, kinds,
		[]model.Board{{Name: board, Members: members, Environment: EnvironmentSimulated}})
	if err != nil {
		t.Fatal(err)
	}
	return record
}

func simulatedBirthInbound(t *testing.T, board string, members []string, kinds []any) Inbound {
	t.Helper()
	in, err := BirthInbound(simulatedBirthRecord(t, board, members, kinds), "sim-1")
	if err != nil {
		t.Fatal(err)
	}
	return in
}

type mixedRig struct {
	*relayRig
	sim *Transmitter
}

func mixedRigWith(t *testing.T, simulated Inbound) *mixedRig {
	t.Helper()
	relay, err := NewMixedRelay(250, simulated)
	if err != nil {
		t.Fatal(err)
	}
	rig := newRelayRig(t)
	rig.relay = relay
	return &mixedRig{relayRig: rig, sim: NewTransmitter("sim-1")}
}

func newMixedRig(t *testing.T) *mixedRig {
	t.Helper()
	return mixedRigWith(t, simulatedBirthInbound(t, "s-1", []string{"S1", "S2"}, birthKinds()))
}

func (m *mixedRig) bringUpMixed() (attached, accepted Moves) {
	m.t.Helper()
	m.relay.GatewayAttached()
	m.gw(hardwareBirth(m.t, nil))
	attached = m.relay.HostAttached()
	accepted = m.hostSends(hostState(7, 0))
	m.gw(deviceBirthOf("P1", "boot-1"))
	m.gw(deviceBirthOf("V1", "boot-7"))
	return attached, accepted
}

func (m *mixedRig) simSends(record Record) SimMoves {
	m.t.Helper()
	m.sim.Queue(record)
	return m.relay.FromSim(relayDecode(m.t, m.sim.Drain()[0]))
}

func simDeviceBirth(device string) Record {
	record := deviceBirthOf(device, "sim-boot")
	record.Fields["board"] = "s-1"
	return record
}

func commandFor(device, id string) Record {
	record := hostCommand()
	record.Fields["device"], record.Fields["command_id"] = device, id
	return record
}

func TestAMixedNodeAnnouncesTheGatewaysBoardsAndTheWorldsInOneBirth(t *testing.T) {
	t.Parallel()
	g := newMixedRig(t)
	attached, accepted := g.bringUpMixed()
	if g.relay.sim == nil || len(attached.Up) != 1 || attached.Up[0]["type"] != "NBIRTH" {
		t.Fatalf("up = %v, want the one merged NBIRTH", typesOf(attached.Up))
	}
	birth := relayDecode(t, attached.Up[0])
	if birth.Invalid != nil || birth.Environment != EnvironmentMixed || EnvironmentClaimFault(birth.Message) != nil {
		t.Fatalf("NBIRTH %v (%v) environment %q, want a valid mixed node", attached.Up[0], birth.Invalid, birth.Environment)
	}
	boards := map[string]string{}
	for _, board := range birth.Boards {
		boards[board.Board] = board.Environment
	}
	if want := map[string]string{"b-1": EnvironmentHardware, "s-1": EnvironmentSimulated}; !reflect.DeepEqual(boards, want) {
		t.Fatalf("boards = %v, want each board to keep the environment of its own origin %v", boards, want)
	}
	assertKindsKeepTheirDigests(t, birth)
	if types := typesOf(accepted.ToSim); !slices.Equal(types, []string{"STATE"}) || accepted.ToSim[0]["authority_epoch"] != json.Number("7") {
		t.Fatalf("to the simulated node = %v, want the host's STATE alone, with the host's epoch", accepted.ToSim)
	}
}

func assertKindsKeepTheirDigests(t *testing.T, birth Inbound) {
	t.Helper()
	announced := map[string]string{}
	for _, entry := range birth.Kinds {
		announced[kindNamed(entry)] = digestOfKind(entry)
		if KindMisdigested(entry) {
			t.Errorf("kind %s was announced with a digest that is not its own", kindNamed(entry))
		}
	}
	for _, fresh := range birthKinds() {
		digest, err := KindDigest(fresh.(map[string]any))
		if err != nil || announced[kindNamed(fresh)] != digest {
			t.Errorf("kind %s announced as %q, its file digests to %q (%v)", kindNamed(fresh), announced[kindNamed(fresh)], digest, err)
		}
	}
	if len(announced) != 2 || birth.CapabilityDigest != CapabilityDigest(birth.Kinds) {
		t.Errorf("kinds %v capability_digest %q, want the two shared kinds once and a capability digest over them", announced, birth.CapabilityDigest)
	}
}

func TestAGatewayThatCollidesWithTheWorldIsRefusedAtTheHandshakeWithTheReason(t *testing.T) {
	t.Parallel()
	otherPump := []any{map[string]any{"kind": "pump", "metrics": []any{}, "commands": []any{}}}
	cases := map[string]struct {
		world  Inbound
		reason string
	}{
		"a device on both sides":  {simulatedBirthInbound(t, "s-1", []string{"P1"}, birthKinds()), `device "P1"`},
		"a board on both sides":   {simulatedBirthInbound(t, "b-1", []string{"S1"}, birthKinds()), `board "b-1"`},
		"a kind with two digests": {simulatedBirthInbound(t, "s-1", []string{"S1"}, otherPump), `kind "pump"`},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := mixedRigWith(t, tc.world)
			g.relay.GatewayAttached()
			moves := g.gw(hardwareBirth(t, nil))
			stats := g.relay.Stats()
			if len(moves.Down) != 1 || moves.Down[0]["refusal"] != "inventory_mismatch" || !moves.CloseGateway || g.relay.GatewayAdmitted() {
				t.Fatalf("down = %v close %v, want an inventory_mismatch refusal and the gateway closed", moves.Down, moves.CloseGateway)
			}
			if stats.Refused != "inventory_mismatch" || !strings.Contains(stats.Reason, tc.reason) {
				t.Fatalf("stats = %+v, want the refusal with a reason naming %s", stats, tc.reason)
			}
		})
	}
}

func TestAGatewayThatIsItselfMixedOrSimulatedIsRefusedEvenBesideAWorld(t *testing.T) {
	t.Parallel()
	for _, environment := range []string{EnvironmentMixed, EnvironmentSimulated} {
		g := newMixedRig(t)
		g.relay.GatewayAttached()
		moves := g.gw(hardwareBirth(t, func(fields map[string]any) { fields["environment"] = environment }))
		if len(moves.Down) != 1 || moves.Down[0]["refusal"] != "environment_mismatch" {
			t.Fatalf("a %s gateway: down = %v, want environment_mismatch", environment, moves.Down)
		}
	}
}

func TestAMixedRelayNeedsAValidBirthOfASimulatedNode(t *testing.T) {
	t.Parallel()
	hardware, err := BirthInbound(hardwareBirth(t, nil), "sim-1")
	if err != nil {
		t.Fatal(err)
	}
	broken := simulatedBirthInbound(t, "s-1", []string{"S1"}, birthKinds())
	broken.Invalid = errors.New("fails its schema")
	notBirth := relayDecode(t, map[string]any{"type": "STATE", "lane": "control", "session_id": "x", "protocol_version": 1, "seq": 0,
		"host_id": "h", "online": true, "heartbeat": 0, "heartbeat_interval_ms": 1000, "authority_epoch": 1, "protocol_min": 1, "protocol_max": 1})
	for name, birth := range map[string]Inbound{"a hardware node": hardware, "an invalid birth": broken, "a record that is no birth": notBirth} {
		if _, err := NewMixedRelay(250, birth); !errors.Is(err, ErrSimulatedBirth) {
			t.Errorf("%s: NewMixedRelay = %v, want ErrSimulatedBirth", name, err)
		}
	}
}

func TestARecordTheSimulatedNodeSendsIsBuiltIntoAnInboundOrRefused(t *testing.T) {
	t.Parallel()
	if _, err := InboundOf(nil); !errors.Is(err, ErrNotRecord) {
		t.Errorf("InboundOf(nil) = %v, want ErrNotRecord", err)
	}
	if _, err := InboundOf(map[string]any{"type": make(chan int)}); !errors.Is(err, ErrNotRecord) {
		t.Errorf("InboundOf(a channel) = %v, want ErrNotRecord", err)
	}
	if _, err := BirthInbound(Record{Type: "NBIRTH", Fields: map[string]any{"bad": make(chan int)}}, "sim-1"); !errors.Is(err, ErrNotRecord) {
		t.Errorf("BirthInbound of an unencodable birth = %v, want ErrNotRecord", err)
	}
}

func TestAGatewayRebirthThatNowCollidesWithTheWorldIsDroppedNotCarried(t *testing.T) {
	t.Parallel()
	g := newMixedRig(t)
	g.bringUpMixed()
	before := g.relay.Stats().Dropped
	moves := g.gw(Record{Type: "NBIRTH", Lane: LaneControl, Fields: hardwareBirthMembers(t, "b-1", "P1", "S1")})
	if len(moves.Up) != 0 || g.relay.Stats().Dropped != before+1 {
		t.Fatalf("up = %v dropped %d, want the colliding rebirth dropped", typesOf(moves.Up), g.relay.Stats().Dropped)
	}
}

func hardwareBirthMembers(t *testing.T, board string, members ...string) map[string]any {
	t.Helper()
	record, err := NodeBirthOf(NodeIdentity{Environment: EnvironmentHardware, ProtocolMax: 1, PriorLastSeq: -1, ClockMode: "wall"}, birthKinds(),
		[]model.Board{{Name: board, Members: members, Environment: EnvironmentHardware}})
	if err != nil {
		t.Fatal(err)
	}
	return record.Fields
}

func TestAHardwareOnlyRelayStillPassesACommandForAnyDeviceToTheGateway(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	moves := g.hostSends(commandFor("ghost-9", "c-9"))
	if g.relay.sim != nil || !slices.Equal(typesOf(moves.Down), []string{"DCMD"}) || g.relay.Stats().UnknownDevice != 0 {
		t.Fatalf("down = %v, want the gateway to decide about a device the relay does not know", typesOf(moves.Down))
	}
}
