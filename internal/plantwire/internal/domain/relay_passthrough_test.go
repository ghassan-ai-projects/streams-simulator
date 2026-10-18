package domain

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestARelayPresentsTheGatewaysBirthToTheHostAsItsOwnSessionWithTheSameKinds(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.relay.GatewayAttached()
	g.gw(hardwareBirth(t, func(fields map[string]any) { fields["bd_seq"], fields["prior_last_seq"] = 9, 41 }))
	birth := g.relay.HostAttached().Up[0]
	sameMembers(t, birth, g.sent, "kinds", "boards", "capability_digest", "node_metrics", "environment", "contract_version", "contract_digest", "protocol_min", "protocol_max")
	if birth["type"] != "NBIRTH" || birth["session_id"] != "relay-h1" || birth["seq"] != uint64(0) || birth["lane"] != LaneControl {
		t.Fatalf("envelope = %v, want the relay's own first NBIRTH of session relay-h1", birth)
	}
	if birth["bd_seq"] != uint64(0) || birth["prior_last_seq"] != int64(-1) {
		t.Fatalf("bd_seq %v prior_last_seq %v, want the relay's own 0 and -1, not the gateway's 9 and 41", birth["bd_seq"], birth["prior_last_seq"])
	}
}

func TestAHostCommissionsTheSameKindDigestsAsTheGatewayAnnounced(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.relay.GatewayAttached()
	g.gw(hardwareBirth(t, nil))
	announced := relayDecode(t, g.sent).Kinds
	received := relayDecode(t, g.relay.HostAttached().Up[0])
	if KindsMisdigested(received.Kinds) || len(received.Kinds) != len(announced) {
		t.Fatalf("kinds the host received = %v, want the %d kinds announced with digests that hold", received.Kinds, len(announced))
	}
	for i, kind := range announced {
		want := kind.(map[string]any)["kind_digest"]
		if got := received.Kinds[i].(map[string]any)["kind_digest"]; got != want {
			t.Errorf("kind %d digest = %v, want the gateway's %v", i, got, want)
		}
	}
	if received.CapabilityDigest != relayDecode(t, g.sent).CapabilityDigest {
		t.Errorf("capability_digest = %s, want the gateway's", received.CapabilityDigest)
	}
}

func TestARelayPassesDeviceRecordsUpWithContentUnchangedAndOwnEnvelopeInOrder(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	var seqs []uint64
	for i, quality := range []string{"good", "suspect", "stale", "disconnected"} {
		moves := g.gw(DeviceData("P1", "boot-1", false, []map[string]any{Sample(1, float64(i)+0.25, quality, 1767225600123456789+int64(i))}))
		if len(moves.Up) != 1 || len(moves.Down) != 0 {
			t.Fatalf("a data record produced up %d down %d records, want exactly one up", len(moves.Up), len(moves.Down))
		}
		sameContent(t, moves.Up[0], g.sent)
		if moves.Up[0]["session_id"] != "relay-h1" || moves.Up[0]["lane"] != LaneData || moves.Up[0]["type"] != "DDATA" {
			t.Fatalf("envelope = %v, want the relay's session and the data lane", moves.Up[0])
		}
		seqs = append(seqs, Seq(moves.Up[0]["seq"]))
	}
	if want := []uint64{3, 4, 5, 6}; !slices.Equal(seqs, want) {
		t.Fatalf("seq of the relayed records = %v, want %v assigned in the order they arrived", seqs, want)
	}
}

func TestARelayKeepsEveryQualityAndTimeOfADataRecordExactly(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	sample := Sample(1, 0.0, "warming_up", 1767225600123456789)
	got := g.gw(DeviceData("P1", "boot-1", true, []map[string]any{sample})).Up[0]
	metric := got["metrics"].([]any)[0].(map[string]any)
	if metric["quality"] != "warming_up" || metric["value"] != json.Number("0") || got["historical"] != true || got["boot_id"] != "boot-1" {
		t.Fatalf("metric %v historical %v boot %v, want the quality, the value, the flag and the boot as sent", metric, got["historical"], got["boot_id"])
	}
	if metric["event_time_ns"].(json.Number).String() != "1767225600123456789" {
		t.Fatalf("event_time_ns = %v, want the exact 64-bit time", metric["event_time_ns"])
	}
}

func TestAHostStateReachesTheGatewayUnchangedSoTheFirmwareSeesTheRealAuthority(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	moves := g.hostSends(hostState(8, 5))
	if len(moves.Down) != 1 || moves.Up != nil {
		t.Fatalf("a heartbeat produced down %d up %d records, want one STATE down", len(moves.Down), len(moves.Up))
	}
	state := moves.Down[0]
	sameContent(t, state, g.hostSent)
	if state["session_id"] != "gw-1" || state["authority_epoch"] != json.Number("8") || state["heartbeat"] != json.Number("5") || state["host_id"] != "host-a" {
		t.Fatalf("STATE = %v, want the host's own epoch 8, heartbeat 5 and id in the gateway's session", state)
	}
}

func TestARelayForwardsACommandToTheGatewayWithEverySafetyMemberUntouched(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	moves := g.hostSends(hostCommand())
	if len(moves.Down) != 1 || moves.Down[0]["type"] != "DCMD" {
		t.Fatalf("down = %v, want the DCMD alone", typesOf(moves.Down))
	}
	command := moves.Down[0]
	sameContent(t, command, g.hostSent)
	want := map[string]string{"expected_boot_id": "boot-1", "deadline_ns": "1767225660000000000", "not_before_ns": "0", "authority_epoch": "7",
		"fallback_on_lease_loss": "safe_off", "command_id": "c-1", "device": "P1", "command": "start"}
	for member, text := range want {
		if got := command[member]; got != json.Number(text) && got != text {
			t.Errorf("%s = %v (%T), want %s as the host sent it", member, got, got, text)
		}
	}
	if command["policy_digest"] != g.hostSent["policy_digest"] || command["idempotency_key"] != g.hostSent["idempotency_key"] {
		t.Errorf("policy_digest %v idempotency_key %v, want the host's", command["policy_digest"], command["idempotency_key"])
	}
}

func TestAReceiptAndAResultReachTheHostWithTheirContentUnchanged(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	for _, record := range []Record{Receipt("c-1", "refused", "wrong_boot", "boot-9"), Result("c-1", "unknown", "boot-1"), DeviceState("P1", "boot-1", []any{})} {
		moves := g.gw(record)
		if len(moves.Up) != 1 {
			t.Fatalf("%s produced %d records up, want 1", record.Type, len(moves.Up))
		}
		sameContent(t, moves.Up[0], g.sent)
		if moves.Up[0]["type"] != record.Type || moves.Up[0]["lane"] != LaneControl {
			t.Errorf("relayed %v, want a %s on the control lane", moves.Up[0], record.Type)
		}
	}
}

func TestARelayForwardsADeviceQueryAndARebirthRequestToTheGateway(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	query := Record{Type: "DQUERY", Lane: LaneControl, Fields: map[string]any{"device": "P1"}}
	rebirth := Record{Type: "NCMD", Lane: LaneControl, Fields: map[string]any{"command": "rebirth"}}
	for _, record := range []Record{query, rebirth} {
		moves := g.hostSends(record)
		if len(moves.Down) != 1 || moves.Down[0]["type"] != record.Type {
			t.Fatalf("down = %v, want the %s unchanged", typesOf(moves.Down), record.Type)
		}
		sameContent(t, moves.Down[0], g.hostSent)
	}
}

func TestARelayNeverOriginatesACommandOrAQuery(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	var down []map[string]any
	take := func(moves Moves) { down = append(down, moves.Down...) }
	g.bringUp()
	take(g.relay.HostLost())
	take(g.relay.HostUnreachable())
	take(g.relay.HostAttached())
	g.newHostSession("relay-h2")
	take(g.hostSends(hostState(7, 0)))
	take(g.gwEdited(DeviceData("P1", "boot-1", false, []map[string]any{Sample(1, 1.0, "", 1)}), func(record map[string]any) { record["seq"] = 40 }))
	take(g.relay.GatewayLost())
	take(g.relay.Shutdown())
	if hasType(down, "DCMD") || hasType(down, "DQUERY") || hasType(down, "HSYNC") {
		t.Fatalf("the relay sent %v to the gateway, want only STATE and NCMD rebirth of its own", typesOf(down))
	}
	if !hasType(down, "NCMD") || !hasType(down, "STATE") {
		t.Fatalf("the relay sent %v, want the offline STATE and the rebirth request it owes the gateway", typesOf(down))
	}
}
