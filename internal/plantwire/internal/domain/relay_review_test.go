package domain

import (
	"slices"
	"testing"
)

func TestAGatewayCannotAnswerACommandThatWasRoutedToTheSimulatedWorld(t *testing.T) {
	t.Parallel()
	g := newMixedRig(t)
	g.bringUpMixed()
	g.hostSends(commandFor("S1", "c-sim"))
	g.hostSends(commandFor("P1", "c-hw"))
	for _, forged := range []Record{Receipt("c-sim", "accepted", "", "boot-1"), Result("c-sim", "succeeded", "boot-1"), Receipt("c-never", "accepted", "", "boot-1")} {
		if moves := g.gw(forged); len(moves.Up) != 0 {
			t.Errorf("a gateway %s for a command it was not given went up as %v, want it dropped", forged.Type, typesOf(moves.Up))
		}
	}
	if stats := g.relay.Stats(); stats.Forged != 3 {
		t.Fatalf("stats = %+v, want the three forged answers counted", stats)
	}
	if moves := g.gw(Receipt("c-hw", "accepted", "", "boot-1")); !slices.Equal(typesOf(moves.Up), []string{"DRECEIPT"}) {
		t.Fatalf("up = %v, want the gateway's own answer for its own command carried", typesOf(moves.Up))
	}
}

func TestTheSimulatedWorldCannotAnswerACommandThatWasRoutedToTheGateway(t *testing.T) {
	t.Parallel()
	g := newMixedRig(t)
	g.bringUpMixed()
	g.hostSends(commandFor("P1", "c-hw"))
	g.hostSends(commandFor("S1", "c-sim"))
	forged := g.simSends(Receipt("c-hw", "accepted", "", "sim-boot"))
	if len(forged.Up) != 0 || g.relay.Stats().Forged != 1 {
		t.Fatalf("up = %v stats %+v, want the world's answer to a gateway command dropped and counted", typesOf(forged.Up), g.relay.Stats())
	}
	if own := g.simSends(Result("c-sim", "succeeded", "sim-boot")); !slices.Equal(typesOf(own.Up), []string{"DRESULT"}) {
		t.Fatalf("up = %v, want the world's answer to its own command carried", typesOf(own.Up))
	}
}

func TestADeviceBirthThatContradictsTheNodeBirthIsDroppedAndCounted(t *testing.T) {
	t.Parallel()
	cases := map[string]func(fields map[string]any){
		"another board than the one that announced the device": func(f map[string]any) { f["board"] = "b-9" },
		"a kind the node birth never announced":                func(f map[string]any) { f["kind"] = "ghost" },
	}
	for name, edit := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := newRelayRig(t)
			g.bringUp()
			birth := deviceBirthOf("P1", "boot-2")
			edit(birth.Fields)
			if moves := g.gw(birth); len(moves.Up) != 0 || g.relay.Stats().Forged != 1 {
				t.Fatalf("up = %v stats %+v, want the birth dropped and reported", typesOf(moves.Up), g.relay.Stats())
			}
		})
	}
	g := newRelayRig(t)
	g.bringUp()
	other := deviceBirthOf("P1", "boot-2")
	other.Fields["kind"] = "valve"
	if moves := g.gw(other); len(moves.Up) != 1 {
		t.Fatalf("up = %v, want a birth of another announced kind on the right board carried", typesOf(moves.Up))
	}
}

func TestOnlyWhatDecodedInTheAdmittedSessionIsMarkedForTheFieldTrace(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.relay.GatewayAttached()
	refused := g.gwEdited(hardwareBirth(t, func(f map[string]any) { f["contract_version"] = "9.9.9" }), nil)
	if refused.Trace[0].Field {
		t.Fatal("a refused birth is marked for the field trace")
	}
	g = newRelayRig(t)
	g.bringUp()
	cases := map[string]struct {
		edit func(map[string]any)
		want bool
	}{
		"a valid record":              {nil, true},
		"a record failing its schema": {func(r map[string]any) { delete(r, "metrics") }, false},
		"a record of another session": {func(r map[string]any) { r["session_id"] = "intruder" }, false},
	}
	for name, tc := range cases {
		if got := g.gwEdited(validData(), tc.edit).Trace[0].Field; got != tc.want {
			t.Errorf("%s: marked %v, want %v", name, got, tc.want)
		}
	}
	rebirth := g.gwEdited(hardwareBirth(t, func(f map[string]any) { f["contract_version"] = "9.9.9" }), nil)
	if rebirth.Trace[0].Field {
		t.Fatal("a refused rebirth is marked for the field trace")
	}
	if accepted := g.gw(hardwareBirth(t, nil)); !accepted.Trace[0].Field {
		t.Fatal("an accepted rebirth is not marked for the field trace")
	}
}

func TestACommandForADeviceOfALostGatewayIsRefusedNotReady(t *testing.T) {
	t.Parallel()
	pure := newRelayRig(t)
	pure.bringUp()
	pure.relay.GatewayLost()
	mixed := newMixedRig(t)
	mixed.bringUpMixed()
	mixed.relay.GatewayLost()
	for name, g := range map[string]*relayRig{"relay": pure, "mixed relay": mixed.relayRig} {
		receipt := g.hostSends(commandFor("P1", "c-lost")).Up
		if len(receipt) != 1 || receipt[0]["type"] != "DRECEIPT" || receipt[0]["status"] != "refused" || receipt[0]["reason"] != "not_ready" || receipt[0]["command_id"] != "c-lost" {
			t.Errorf("%s: up = %v, want one refused not_ready receipt", name, receipt)
		}
	}
}

func TestAGatewayThatSaysGoodbyeCanBirthAgainOnTheSameLinkAndLeavesNothingBehind(t *testing.T) {
	t.Parallel()
	g := newMixedRig(t)
	g.bringUpMixed()
	g.gw(validData())
	g.gw(barrierSync(3))
	if len(g.relay.host.gatewaySyncs.ups) != 1 {
		t.Fatalf("pending barriers = %v, want the gateway's NSYNC waiting for its answer", g.relay.host.gatewaySyncs.ups)
	}
	g.gw(NodeDeath(0, "shutdown"))
	if g.relay.GatewayAdmitted() || g.relay.OriginOf("P1") != OriginUnknown || len(g.relay.host.gatewaySyncs.ups) != 0 {
		t.Fatalf("admitted %v origin %d syncs %v, want the gateway forgotten", g.relay.GatewayAdmitted(), g.relay.OriginOf("P1"), g.relay.host.gatewaySyncs.ups)
	}
	g.gateway = NewTransmitter("gw-1")
	if moves := g.gw(hardwareBirth(t, nil)); !g.relay.GatewayAdmitted() || !hasType(moves.Up, "NBIRTH") {
		t.Fatalf("up = %v admitted %v, want a birth at seq 0 admitted after the goodbye", typesOf(moves.Up), g.relay.GatewayAdmitted())
	}
	if g.relay.OriginOf("P1") != OriginGateway {
		t.Fatal("the reborn gateway's devices are not routed to it")
	}
}

func TestALostGatewayLeavesNoBarrierOrDeviceBehind(t *testing.T) {
	t.Parallel()
	g := newMixedRig(t)
	g.bringUpMixed()
	g.gw(validData())
	g.gw(barrierSync(3))
	g.relay.GatewayLost()
	if g.relay.OriginOf("V1") != OriginUnknown || len(g.relay.host.gatewaySyncs.ups) != 0 || g.relay.gateway.in != (Receiver{}) {
		t.Fatalf("origin %d syncs %v, want devices, barriers and sequence forgotten", g.relay.OriginOf("V1"), g.relay.host.gatewaySyncs.ups)
	}
}

func TestAHostRebirthRequestInAMixedRelayReachesTheGatewayAndTheWorld(t *testing.T) {
	t.Parallel()
	g := newMixedRig(t)
	g.bringUpMixed()
	rebirth := Record{Type: "NCMD", Lane: LaneControl, Fields: map[string]any{"command": "rebirth"}}
	moves := g.hostSends(rebirth)
	if !slices.Equal(typesOf(moves.Down), []string{"NCMD"}) || !slices.Equal(typesOf(moves.ToSim), []string{"NCMD"}) || moves.ToSim[0]["command"] != "rebirth" {
		t.Fatalf("down %v to the world %v, want the rebirth request with both", typesOf(moves.Down), typesOf(moves.ToSim))
	}
	g.relay.GatewayLost()
	if gone := g.hostSends(rebirth); len(gone.Down) != 0 || !slices.Equal(typesOf(gone.ToSim), []string{"NCMD"}) {
		t.Fatalf("down %v to the world %v, want the world reborn even when the gateway is gone", typesOf(gone.Down), typesOf(gone.ToSim))
	}
}

func TestALostSimulatedWorldKillsItsDevicesOnceAndRefusesItsCommands(t *testing.T) {
	t.Parallel()
	g := newMixedRig(t)
	g.bringUpMixed()
	g.simSends(simDeviceBirth("S1"))
	lost := g.relay.SimLost()
	if len(lost.Up) != 1 || lost.Up[0]["type"] != "DDEATH" || lost.Up[0]["device"] != "S1" || g.relay.Stats().SimulatedLost != 1 {
		t.Fatalf("up = %v stats %+v, want S1 reported offline and the loss counted", lost.Up, g.relay.Stats())
	}
	if again := g.relay.SimLost(); len(again.Up) != 0 || g.relay.Stats().SimulatedLost != 1 {
		t.Fatalf("a second loss produced %v, want nothing", again.Up)
	}
	receipt := g.hostSends(commandFor("S1", "c-dead")).Up
	if len(receipt) != 1 || receipt[0]["status"] != "refused" || receipt[0]["reason"] != "not_ready" {
		t.Fatalf("up = %v, want a command for a dead world refused not_ready", receipt)
	}
}

func TestAPlanSocketThatIsRelativeHasADotDotOrAControlCharacterIsRefusedAndOtherwiseCleaned(t *testing.T) {
	t.Parallel()
	cases := map[string]struct{ host, gateway, pointer, keyword string }{
		"a relative host socket":   {"h.sock", "", "/host_socket", "socket_not_absolute"},
		"a parent component":       {"/tmp/../etc/h.sock", "", "/host_socket", "socket_parent_component"},
		"a control character":      {"/tmp/h\x07.sock", "", "/host_socket", "socket_control_character"},
		"a gateway with a parent":  {"/tmp/h.sock", "/tmp/a/../g.sock", "/gateway_socket", "socket_parent_component"},
		"the same socket cleaned":  {"/tmp/h.sock", "/tmp//./h.sock", "/gateway_socket", "same_socket"},
		"a relative gateway":       {"/tmp/h.sock", "g.sock", "/gateway_socket", "socket_not_absolute"},
		"a gateway with a control": {"/tmp/h.sock", "/tmp/g\n.sock", "/gateway_socket", "socket_control_character"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			plan := benchPlan()
			plan.HostSocket, plan.GatewaySocket = tc.host, tc.gateway
			_, issues := CheckPlan(plan, benchInstances())
			if len(issues) == 0 || issues[0].Pointer != tc.pointer || issues[0].Keyword != tc.keyword {
				t.Fatalf("issues = %+v, want %s at %s", issues, tc.keyword, tc.pointer)
			}
		})
	}
	plan := benchPlan()
	plan.HostSocket, plan.GatewaySocket = "/tmp//run/./h.sock", "/tmp/g.sock/"
	checked, issues := CheckPlan(plan, benchInstances())
	if len(issues) != 0 || checked.HostSocket != "/tmp/run/h.sock" || checked.GatewaySocket != "/tmp/g.sock" {
		t.Fatalf("plan %+v issues %+v, want the sockets cleaned", checked, issues)
	}
}
