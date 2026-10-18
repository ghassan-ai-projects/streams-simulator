package domain

import "testing"

func TestALateTickCountsTheTimeThatPassedForTheGatewayWatchdog(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	if moves := g.relay.TickBy(gatewayIdleMS - 1); moves.CloseGateway {
		t.Fatal("gateway closed before the idle limit")
	}
	if moves := g.relay.TickBy(1); !moves.CloseGateway {
		t.Fatal("one late tick covering the idle limit did not close the silent gateway")
	}
}

func TestALateTickCountsTheTimeThatPassedForTheHostHeartbeat(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	g.hostSends(hostState(7, 1))
	if moves := g.relay.TickBy(2999); moves.CloseHost {
		t.Fatal("host lost before three missed heartbeats")
	}
	if moves := g.relay.TickBy(1); !moves.CloseHost {
		t.Fatal("a single late tick covering three heartbeats did not lose the host")
	}
}
