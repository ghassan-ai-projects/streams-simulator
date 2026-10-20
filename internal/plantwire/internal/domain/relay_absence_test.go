package domain

import "testing"

func TestASilentGatewayIsKeptBeforeTheFirstHostAndWatchedOnceTheHostAnswers(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.relay.GatewayAttached()
	g.gw(hardwareBirth(g.t, nil))
	if offline := g.relay.HostUnreachable(); len(offline.Down) != 1 || offline.Down[0]["online"] != false {
		t.Fatalf("unreachable host = %+v, want the gateway told its host is offline", offline)
	}
	silentFor(t, g.relay, 200)
	g.relay.HostAttached()
	if moves := g.hostSends(hostState(7, 0)); !hasType(moves.Down, "STATE") {
		t.Fatalf("the first host STATE = %+v, want it forwarded to the gateway", moves)
	}
	silentFor(t, g.relay, 119)
	if moves := tickOnce(g.relay); !moves.CloseGateway {
		t.Fatal("the watchdog did not re-arm when the host answered")
	}
}

func TestASilentGatewayIsKeptDuringAHostLossAndGetsAFullWindowWhenTheHostReturns(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	silentFor(t, g.relay, 116)
	if lost := g.relay.HostLost(); !lost.CloseHost || !hasType(lost.Down, "STATE") {
		t.Fatalf("host loss = %+v, want the host closed and the gateway told it is offline", lost)
	}
	silentFor(t, g.relay, 400)
	g.relay.HostAttached()
	g.newHostSession("relay-h2")
	moves := g.hostSends(hostState(7, 0))
	if !hasType(moves.Down, "STATE") || !hasType(moves.Down, "NCMD") {
		t.Fatalf("the returning host = %+v, want its STATE and a rebirth request forwarded", moves)
	}
	silentFor(t, g.relay, 119)
	if moves := tickOnce(g.relay); !moves.CloseGateway {
		t.Fatal("the watchdog did not re-arm when the host returned")
	}
}

func TestHostHeartbeatsDoNotKeepASilentGatewayAlive(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	for beat := 1; beat < 30; beat++ {
		silentFor(t, g.relay, 4)
		g.hostSends(hostState(7, beat))
	}
	silentFor(t, g.relay, 3)
	if moves := tickOnce(g.relay); !moves.CloseGateway {
		t.Fatal("forwarded host heartbeats refreshed the gateway watchdog")
	}
}

func silentFor(t *testing.T, r *Relay, ticks int) {
	t.Helper()
	for range ticks {
		if moves := tickOnce(r); moves.CloseGateway {
			t.Fatal("the gateway watchdog closed a gateway it should keep")
		}
	}
}
