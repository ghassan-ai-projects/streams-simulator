package domain

import "testing"

func TestASilentConnectedGatewayIsLostWithoutAHostHeartbeat(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	for range 119 {
		if moves := g.relay.Tick(); moves.CloseGateway {
			t.Fatal("gateway watchdog ended before 30 seconds")
		}
	}
	moves := g.relay.Tick()
	if !moves.CloseGateway || g.relay.GatewayAdmitted() || len(moves.Up) != 2 {
		t.Fatalf("gateway watchdog = %+v; want both devices dead and gateway closed",moves)
	}
	for _, record := range moves.Up {
		if record["type"] != "DDEATH" || record["reason"] != "link_down" {
			t.Fatalf("watchdog death = %v",record)
		}
	}
	if len(moves.Notes) != 1 || moves.Notes[0].Fields["state"] != "offline" {
		t.Fatalf("watchdog health note = %+v",moves.Notes)
	}
}

func TestAdmittedGatewayTrafficResetsWatchdogButForgedTrafficDoesNot(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	for range 119 { g.relay.Tick() }
	g.gw(DeviceData("P1","boot-1",false,[]map[string]any{Sample(1,20.0,"good",0)}))
	if moves := g.relay.Tick(); moves.CloseGateway {
		t.Fatal("admitted data did not reset watchdog")
	}
	for range 118 { g.relay.Tick() }
	g.gw(deviceBirthOf("intruder","other"))
	if moves := g.relay.Tick(); !moves.CloseGateway {
		t.Fatal("forged device kept a silent gateway alive")
	}
}

func TestMalformedAndRepeatedGatewayRecordsDoNotRefreshWatchdog(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	for range 119 { g.relay.Tick() }
	g.relay.FromGateway(relayDecode(t,g.sent))
	bad := relayDecode(t,g.sent)
	bad.Seq++
	bad.Invalid = ErrNotSession
	g.relay.FromGateway(bad)
	if moves := g.relay.Tick(); !moves.CloseGateway {
		t.Fatal("malformed or duplicate input refreshed the watchdog")
	}
}
