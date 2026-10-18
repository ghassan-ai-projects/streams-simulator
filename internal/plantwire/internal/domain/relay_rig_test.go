package domain

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type relayRig struct {
	t        *testing.T
	relay    *Relay
	gateway  *Transmitter
	host     *Transmitter
	sent     map[string]any
	hostSent map[string]any
}

func newRelayRig(t *testing.T) *relayRig {
	t.Helper()
	return &relayRig{t: t, relay: NewRelay(250), gateway: NewTransmitter("gw-1"), host: NewTransmitter("relay-h1")}
}

func relayDecode(t *testing.T, record map[string]any) Inbound {
	t.Helper()
	raw, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	in, ok := Decode(raw)
	if !ok {
		t.Fatalf("not a record: %s", raw)
	}
	return in
}

func (g *relayRig) gw(record Record) Moves {
	g.t.Helper()
	return g.gwEdited(record, nil)
}

func (g *relayRig) gwEdited(record Record, edit func(map[string]any)) Moves {
	g.t.Helper()
	g.sent = queueEdited(g.gateway, record, edit)
	return g.relay.FromGateway(relayDecode(g.t, g.sent))
}

func queueEdited(tx *Transmitter, record Record, edit func(map[string]any)) map[string]any {
	tx.Queue(record)
	sent := tx.Drain()[0]
	if edit != nil {
		edit(sent)
	}
	return sent
}

func (g *relayRig) hostSends(record Record) Moves {
	g.t.Helper()
	return g.hostEdited(record, nil)
}

func (g *relayRig) hostEdited(record Record, edit func(map[string]any)) Moves {
	g.t.Helper()
	g.hostSent = queueEdited(g.host, record, edit)
	return g.relay.FromHost(relayDecode(g.t, g.hostSent))
}

func (g *relayRig) bringUp() {
	g.t.Helper()
	g.relay.GatewayAttached()
	g.gw(hardwareBirth(g.t, nil))
	g.relay.HostAttached()
	g.hostSends(hostState(7, 0))
	g.gw(deviceBirthOf("P1", "boot-1"))
	g.gw(deviceBirthOf("V1", "boot-7"))
}

func (g *relayRig) newHostSession(name string) {
	g.host = NewTransmitter(name)
}

func hardwareBirth(t *testing.T, edit func(fields map[string]any)) Record {
	t.Helper()
	record, err := NodeBirthOf(NodeIdentity{Environment: EnvironmentHardware, ProtocolMax: 1, PriorLastSeq: -1, ClockMode: "wall"}, birthKinds(),
		[]model.Board{{Name: "b-1", Members: []string{"P1", "V1"}, Environment: EnvironmentHardware}})
	if err != nil {
		t.Fatal(err)
	}
	if edit != nil {
		edit(record.Fields)
	}
	return record
}

func deviceBirthOf(device, boot string) Record {
	return Record{Type: "DBIRTH", Lane: LaneData, Fields: map[string]any{"device": device, "kind": "pump", "board": "b-1", "boot_id": boot,
		"metadata_version": 1, "birth_time_ns": 0, "metrics": []any{map[string]any{"name": "flow", "alias": 1, "value": 3.5, "quality": "good"}}}}
}

func hostState(epoch int64, heartbeat int) Record {
	return Record{Type: "STATE", Lane: LaneControl, Fields: map[string]any{"host_id": "host-a", "online": true, "heartbeat": heartbeat,
		"heartbeat_interval_ms": 1000, "authority_epoch": epoch, "protocol_min": 1, "protocol_max": 1}}
}

func hostCommand() Record {
	key := Command{Device: "P1", Name: "start", Args: map[string]any{"speed": 2}, Boot: "boot-1"}.Content()
	return Record{Type: "DCMD", Lane: LaneControl, Fields: map[string]any{"command_id": "c-1", "idempotency_key": key, "device": "P1",
		"command": "start", "args": map[string]any{"speed": 2}, "expected_boot_id": "boot-1", "deadline_ns": int64(1767225660000000000),
		"not_before_ns": 0, "authority_epoch": 7, "policy_digest": key, "fallback_on_lease_loss": "safe_off"}}
}

func sameContent(t *testing.T, got, want map[string]any) {
	t.Helper()
	g, gotErr := json.Marshal(content(got))
	w, wantErr := json.Marshal(content(want))
	if gotErr != nil || wantErr != nil || string(g) != string(w) {
		t.Fatalf("content = %s, want %s", g, w)
	}
}

func sameMembers(t *testing.T, got, want map[string]any, members ...string) {
	t.Helper()
	for _, member := range members {
		g, _ := json.Marshal(got[member])
		w, _ := json.Marshal(want[member])
		if string(g) != string(w) {
			t.Errorf("%s = %s, want %s", member, g, w)
		}
	}
}

func typesOf(records []map[string]any) []string {
	names := make([]string, len(records))
	for i, record := range records {
		names[i], _ = record["type"].(string)
	}
	return names
}

func hasType(records []map[string]any, kind string) bool {
	return slices.Contains(typesOf(records), kind)
}
