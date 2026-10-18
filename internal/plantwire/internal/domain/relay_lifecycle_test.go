package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestARelayRefusesAGatewayBirthTheWayAHostWouldAndForwardsNothing(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, refusal string
		edit          func(fields map[string]any)
	}{
		{"a contract it does not speak", "contract_unsupported", func(f map[string]any) { f["contract_version"] = "1.0.0" }},
		{"a kind digest that does not hold", "inventory_mismatch", func(f map[string]any) {
			f["kinds"].([]any)[0].(map[string]any)["kind_digest"] = "sha256:" + strings.Repeat("0", 64)
		}},
		{"a protocol range it cannot meet", "version_unsupported", func(f map[string]any) { f["protocol_min"], f["protocol_max"] = 2, 2 }},
		{"an inverted protocol range", "protocol_invalid", func(f map[string]any) { f["protocol_min"], f["protocol_max"] = 2, 1 }},
		{"a node that claims to be simulated", "environment_mismatch", func(f map[string]any) { f["environment"] = EnvironmentSimulated }},
		{"a mixed node", "environment_mismatch", func(f map[string]any) { f["environment"] = EnvironmentMixed }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := newRelayRig(t)
			g.relay.GatewayAttached()
			moves := g.gw(hardwareBirth(t, tc.edit))
			if len(moves.Down) != 1 || moves.Down[0]["refusal"] != tc.refusal || moves.Down[0]["online"] != false || !moves.CloseGateway || moves.Up != nil {
				t.Fatalf("moves = %+v, want one offline STATE refusing %s, the gateway closed and nothing up", moves, tc.refusal)
			}
			if stats := g.relay.Stats(); stats.Refused != tc.refusal || g.relay.GatewayAdmitted() || g.relay.WantsHost() {
				t.Fatalf("stats %+v admitted %v, want the refusal recorded and no host wanted", stats, g.relay.GatewayAdmitted())
			}
		})
	}
}

func TestARelayRefusalOfAnEnvironmentSaysWhichRuleBroke(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.relay.GatewayAttached()
	g.gw(hardwareBirth(t, func(f map[string]any) {
		f["environment"] = EnvironmentSimulated
		f["boards"].([]any)[0].(map[string]any)["environment"] = EnvironmentSimulated
	}))
	if reason := g.relay.Stats().Reason; reason != "the host accepts hardware nodes only, the node announces simulated" {
		t.Fatalf("reason = %q, want the rule the node broke", reason)
	}
}

func TestARelayRefusesAKindAnnouncedTwice(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.relay.GatewayAttached()
	twice := func(f map[string]any) {
		kinds := f["kinds"].([]any)
		f["kinds"] = append(kinds, map[string]any{"kind": "pump", "metrics": []any{}, "commands": []any{}, "kind_digest": kinds[0].(map[string]any)["kind_digest"]})
	}
	if moves := g.gw(hardwareBirth(t, twice)); len(moves.Down) != 1 || moves.Down[0]["refusal"] != "inventory_mismatch" {
		t.Fatalf("down = %v, want inventory_mismatch for a kind named twice", moves.Down)
	}
}

func TestARelayWithoutAHostTellsTheGatewayOnceThatItIsOfflineAndKeepsTheSession(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.relay.GatewayAttached()
	if moves := g.gw(hardwareBirth(t, nil)); moves.Up != nil || moves.Down != nil || !g.relay.WantsHost() {
		t.Fatalf("moves = %+v wants host %v, want the relay to ask for a host and say nothing yet", moves, g.relay.WantsHost())
	}
	moves := g.relay.HostUnreachable()
	if len(moves.Down) != 1 || moves.Down[0]["online"] != false || moves.Down[0]["host_id"] != "relay" || moves.Down[0]["refusal"] != nil || moves.CloseGateway {
		t.Fatalf("moves = %+v, want one offline STATE of the relay's own, no refusal, the gateway kept", moves)
	}
	if again := g.relay.HostUnreachable(); again.Down != nil {
		t.Fatalf("a second attempt said %v, want silence", again.Down)
	}
}

func TestARelayRepeatsTheHostsOwnStateAsOfflineWhenTheHostIsLost(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	moves := g.relay.HostLost()
	if !moves.CloseHost || len(moves.Down) != 1 {
		t.Fatalf("moves = %+v, want the host link closed and one record for the gateway", moves)
	}
	state := moves.Down[0]
	if state["type"] != "STATE" || state["online"] != false || state["authority_epoch"] != json.Number("7") || state["host_id"] != "host-a" {
		t.Fatalf("STATE = %v, want the real host's identity and epoch 7 with online false", state)
	}
	if later := g.gw(DeviceData("P1", "boot-1", false, []map[string]any{Sample(1, 1.0, "", 1)})); later.Up != nil {
		t.Fatalf("a record with no host went up as %v, want it dropped", later.Up)
	}
	if g.relay.Stats().Dropped != 1 {
		t.Fatalf("dropped = %d, want the one record the relay could not deliver", g.relay.Stats().Dropped)
	}
}

func TestARelayRebirthsTheHostSessionWhenTheHostComesBack(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	g.relay.HostLost()
	birth := g.relay.HostAttached().Up[0]
	if birth["session_id"] != "relay-h2" || birth["seq"] != uint64(0) || birth["bd_seq"] != uint64(1) || birth["prior_last_seq"] != int64(2) {
		t.Fatalf("NBIRTH = %v, want a new session relay-h2 from seq 0, bd_seq 1 and prior_last_seq 2", birth)
	}
	g.newHostSession("relay-h2")
	answered := g.hostSends(hostState(9, 0))
	if got := typesOf(answered.Down); len(got) != 2 || got[0] != "STATE" || got[1] != "NCMD" || answered.Down[1]["command"] != "rebirth" {
		t.Fatalf("down = %v, want the new host's STATE and then a rebirth request for the gateway", got)
	}
	if answered.Down[0]["authority_epoch"] != json.Number("9") {
		t.Fatalf("STATE = %v, want the new host's epoch", answered.Down[0])
	}
}

func TestARelayCarriesTheGatewaysRebirthUpInTheSameHostSession(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	moves := g.gw(hardwareBirth(t, func(f map[string]any) { f["bd_seq"] = 5 }))
	birth := moves.Up[0]
	if birth["session_id"] != "relay-h1" || birth["seq"] != uint64(3) || birth["bd_seq"] != uint64(1) || birth["prior_last_seq"] != int64(-1) {
		t.Fatalf("rebirth NBIRTH = %v, want seq 3 of session relay-h1 with the relay's next bd_seq 1", birth)
	}
	if moves := g.gw(deviceBirthOf("P1", "boot-2")); len(moves.Up) != 1 || moves.Up[0]["boot_id"] != "boot-2" {
		t.Fatalf("birth after the rebirth = %v, want the new boot passed up", moves.Up)
	}
}

func TestARelayDropsARebirthThatWouldHaveBeenRefusedAtTheHandshake(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	moves := g.gw(hardwareBirth(t, func(f map[string]any) { f["contract_version"] = "9.9.9" }))
	if moves.Up != nil || moves.Down != nil || g.relay.Stats().Dropped != 1 {
		t.Fatalf("moves = %+v dropped %d, want nothing relayed and the birth dropped", moves, g.relay.Stats().Dropped)
	}
}

func TestARelayThatLosesItsGatewayKillsEveryBornDeviceAndKeepsTheHostSession(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	moves := g.relay.GatewayLost()
	if len(moves.Up) != 2 || moves.Down != nil {
		t.Fatalf("moves = %+v, want a DDEATH for each of the two devices and nothing for the gateway", moves)
	}
	for i, want := range []struct{ device, boot string }{{"P1", "boot-1"}, {"V1", "boot-7"}} {
		death := moves.Up[i]
		if death["type"] != "DDEATH" || death["device"] != want.device || death["boot_id"] != want.boot || death["reason"] != "link_down" || death["session_id"] != "relay-h1" {
			t.Errorf("record %d = %v, want a link_down DDEATH of %s on boot %s in the host session", i, death, want.device, want.boot)
		}
	}
	if hasType(moves.Up, "NDEATH") || g.relay.GatewayAdmitted() {
		t.Fatalf("moves %v admitted %v, want no NDEATH and no gateway", typesOf(moves.Up), g.relay.GatewayAdmitted())
	}
	if again := g.relay.GatewayLost(); again.Up != nil {
		t.Fatalf("a second loss sent %v, want nothing", again.Up)
	}
}

func TestAReturningGatewayIsAnsweredWithTheHostsRealStateAndRebornUpward(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	g.relay.GatewayLost()
	g.relay.GatewayAttached()
	g.gateway = NewTransmitter("gw-2")
	moves := g.gw(hardwareBirth(t, nil))
	if got := typesOf(moves.Down); len(got) != 1 || got[0] != "STATE" || moves.Down[0]["authority_epoch"] != json.Number("7") || moves.Down[0]["online"] != true {
		t.Fatalf("down = %v, want the host's last STATE, online with epoch 7", moves.Down)
	}
	if len(moves.Up) != 1 || moves.Up[0]["type"] != "NBIRTH" || moves.Up[0]["bd_seq"] != uint64(1) {
		t.Fatalf("up = %v, want the rebirth NBIRTH with bd_seq 1", moves.Up)
	}
}

func TestAGatewayThatSaysGoodbyeIsTreatedAsLostNotForwardedAsANodeDeath(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	moves := g.gw(NodeDeath(0, "shutdown"))
	if got := typesOf(moves.Up); len(got) != 2 || got[0] != "DDEATH" || got[1] != "DDEATH" || g.relay.GatewayAdmitted() {
		t.Fatalf("up = %v admitted %v, want the devices killed and no NDEATH sent for the whole node", got, g.relay.GatewayAdmitted())
	}
}

func TestARelayStopsCountingHeartbeatsOnlyOnceTheHostHasSentOne(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	for range 100 {
		if moves := tickOnce(g.relay); moves.CloseHost {
			t.Fatal("a host that never sent a heartbeat was declared lost")
		}
	}
	g.gw(DeviceData("P1", "boot-1", false, []map[string]any{Sample(1, 20.0, "good", 0)}))
	g.hostSends(hostState(7, 1))
	for i := range 11 {
		if moves := tickOnce(g.relay); moves.CloseHost {
			t.Fatalf("host lost after %d ticks of 250 ms with a 1000 ms heartbeat, want three missed heartbeats", i+1)
		}
	}
	g.hostSends(hostState(7, 2))
	for range 11 {
		tickOnce(g.relay)
	}
	if moves := tickOnce(g.relay); !moves.CloseHost || len(moves.Down) != 1 || moves.Down[0]["online"] != false {
		t.Fatalf("moves = %+v, want the host lost on the twelfth silent tick after the last heartbeat", moves)
	}
}

func TestAHostThatRefusesTheRelayedSessionEndsTheRelayAndTellsTheGatewayWhy(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.relay.GatewayAttached()
	g.gw(hardwareBirth(t, nil))
	g.relay.HostAttached()
	refusal := hostState(7, 0)
	refusal.Fields["online"], refusal.Fields["refusal"] = false, "inventory_mismatch"
	moves := g.hostSends(refusal)
	if !errors.Is(moves.Err, ErrRefused) || !moves.CloseHost || !moves.CloseGateway {
		t.Fatalf("moves = %+v, want ErrRefused and both links closed", moves)
	}
	if len(moves.Down) != 1 || moves.Down[0]["refusal"] != "inventory_mismatch" || moves.Down[0]["online"] != false {
		t.Fatalf("down = %v, want the host's own refusal handed to the gateway", moves.Down)
	}
	sameContent(t, moves.Down[0], g.hostSent)
	if g.relay.Stats().HostRefused != "inventory_mismatch" {
		t.Fatalf("stats = %+v, want the host's refusal recorded", g.relay.Stats())
	}
}

func TestAHostStateTheRelayCannotNegotiateLosesTheHostAndWaitsForTheNext(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.relay.GatewayAttached()
	g.gw(hardwareBirth(t, nil))
	g.relay.HostAttached()
	moves := g.hostEdited(hostState(7, 0), func(record map[string]any) { record["protocol_min"] = 2 })
	if !moves.CloseHost || len(moves.Down) != 1 || moves.Down[0]["online"] != false || g.relay.WantsHost() != true {
		t.Fatalf("moves = %+v, want the host dropped, the gateway told it is offline and a new host wanted", moves)
	}
}

func TestARelayEndsItsNodeSessionWithAnNDeathOfTheLastBirth(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	if moves := g.relay.Shutdown(); moves.Up != nil {
		t.Fatalf("a relay with no host session sent %v", moves.Up)
	}
	g.bringUp()
	g.gw(hardwareBirth(t, nil))
	moves := g.relay.Shutdown()
	if len(moves.Up) != 1 || moves.Up[0]["type"] != "NDEATH" || moves.Up[0]["reason"] != "shutdown" || moves.Up[0]["bd_seq"] != uint64(1) {
		t.Fatalf("up = %v, want NDEATH shutdown of bd_seq 1, the last birth", moves.Up)
	}
}
