package domain

import (
	"encoding/json"
	"testing"
)

func TestARelayCountsWhatItCarriedInBothDirections(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	stats := g.relay.Stats()
	if stats.GatewaySessions != 1 || stats.HostSessions != 1 || stats.Up != 3 || stats.Down != 1 || stats.Dropped != 0 {
		t.Fatalf("stats = %+v, want 1 gateway session, 1 host session, 3 records up (the birth and two devices), 1 down (the STATE) and no drop", stats)
	}
}

func TestAHostAttachedBeforeAnyGatewayBirthSendsNothing(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	moves := g.relay.HostAttached()
	if moves.Up != nil || moves.Down != nil || g.relay.WantsHost() {
		t.Fatalf("moves = %+v wants host %v, want silence until a gateway is born", moves, g.relay.WantsHost())
	}
}

func TestAGatewayLostBeforeTheHostAnswersClosesTheHalfOpenHostSession(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.relay.GatewayAttached()
	g.gw(hardwareBirth(t, nil))
	g.relay.HostAttached()
	moves := g.relay.GatewayLost()
	if !moves.CloseHost || moves.Down != nil || moves.Up != nil || g.relay.WantsHost() {
		t.Fatalf("moves = %+v wants host %v, want the host link closed with nobody told", moves, g.relay.WantsHost())
	}
	g.relay.GatewayAttached()
	g.gateway = NewTransmitter("gw-2")
	g.gw(hardwareBirth(t, nil))
	if !g.relay.WantsHost() {
		t.Fatal("a new gateway born after that wants no host, want a fresh connection")
	}
}

func TestADeviceThatDiedOnTheGatewayIsNotKilledAgainWhenTheGatewayIsLost(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	death := Record{Type: "DDEATH", Lane: LaneControl, Fields: map[string]any{"device": "P1", "boot_id": "boot-1", "reason": "offline", "last_seq": 1}}
	if moves := g.gw(death); len(moves.Up) != 1 || moves.Up[0]["reason"] != "offline" {
		t.Fatalf("up = %v, want the gateway's own DDEATH passed on", moves.Up)
	}
	moves := g.relay.GatewayLost()
	if len(moves.Up) != 1 || moves.Up[0]["device"] != "V1" || moves.Up[0]["boot_id"] != "boot-7" {
		t.Fatalf("up = %v, want only the device still born to be killed", moves.Up)
	}
}

func TestAHostStateWhileNoGatewayIsConnectedIsNotedAndGivenToTheNextOne(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	g.relay.GatewayLost()
	if moves := g.hostSends(hostState(7, 1)); moves.Down != nil {
		t.Fatalf("down = %v, want nothing for a gateway that is gone", moves.Down)
	}
	g.relay.GatewayAttached()
	g.gateway = NewTransmitter("gw-2")
	moves := g.gw(hardwareBirth(t, nil))
	if len(moves.Down) != 1 || moves.Down[0]["heartbeat"] != json.Number("1") {
		t.Fatalf("down = %v, want the host's latest heartbeat, not the one from the handshake", moves.Down)
	}
}

func TestALostHostCannotBeLostTwice(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	g.relay.HostLost()
	if again := g.relay.HostLost(); again.CloseHost || again.Down != nil {
		t.Fatalf("a second loss produced %+v, want nothing", again)
	}
}
