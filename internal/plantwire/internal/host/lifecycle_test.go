package host

import (
	"maps"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

func commissionedHost(t *testing.T) *Host {
	h := New(Options{Environment: "simulated", ProtocolMax: 1, Epoch: 1, Policy: contractPolicy, Inventory: map[string]string{"P1": "pump"},
		Script: []Scripted{{ID: "c-1", AtNS: 10, Device: "P1", Command: "start"}}})
	h.Receive(birth(t, "simulated"))
	h.Receive(pumpBirth(t, 1, "P1"))
	return h
}

func TestADeadOrRequalifiedDeviceGetsNoCommand(t *testing.T) {
	t.Parallel()
	for name, death := range map[string]func(t *testing.T) map[string]any{
		"device death": func(*testing.T) map[string]any { return map[string]any{"type": "DDEATH", "device": "P1"} },
		"node death":   func(*testing.T) map[string]any { return map[string]any{"type": "NDEATH", "bd_seq": 0} },
		"rebirth as valve": func(*testing.T) map[string]any {
			return map[string]any{"type": "DBIRTH", "device": "P1", "kind": "valve"}
		},
	} {
		h := commissionedHost(t)
		fields := death(t)
		kind := fields["type"].(string)
		delete(fields, "type")
		h.Receive(record(t, kind, 2, fields))
		if replies := h.Receive(sync(t, 3, 20)); len(replies) != 1 || h.Report().Commands["c-1"].Denied != "not_commissioned" {
			t.Fatalf("%s: replies %v, report %+v", name, replies, h.Report().Commands["c-1"])
		}
	}
}

func TestALateNodeDeathIsIgnored(t *testing.T) {
	t.Parallel()
	h := commissionedHost(t)
	h.Receive(record(t, "NBIRTH", 2, map[string]any{"bd_seq": 1, "contract_digest": birth(t, "simulated").ContractDigest}))
	h.Receive(pumpBirth(t, 3, "P1"))
	h.Receive(record(t, "NDEATH", 4, map[string]any{"bd_seq": 0}))
	if replies := h.Receive(sync(t, 5, 20)); len(replies) != 2 || replies[0]["type"] != "DCMD" {
		t.Fatalf("a death older than the birth removed the device: %v", replies)
	}
}

func TestTheHandshakeChecksTheInventoryAndTheContractVersion(t *testing.T) {
	t.Parallel()
	missing := New(Options{Environment: "simulated", ProtocolMax: 1, Inventory: map[string]string{"P7": "pump"}})
	if missing.Receive(birth(t, "simulated")); missing.Report().Refused != "inventory_mismatch" {
		t.Fatalf("a commissioned device missing from the birth: %q", missing.Report().Refused)
	}
	older := New(Options{Environment: "simulated", ProtocolMax: 1})
	if older.Receive(birth(t, "simulated", map[string]any{"contract_version": "0.9.0"})); older.Report().Refused != "contract_unsupported" {
		t.Fatalf("another contract version: %q", older.Report().Refused)
	}
}

func TestOnlyGoodCurrentDataOfTheBornBootCounts(t *testing.T) {
	t.Parallel()
	h := New(Options{Environment: "simulated", ProtocolMax: 1})
	h.Receive(birth(t, "simulated"))
	h.Receive(record(t, "DBIRTH", 1, map[string]any{"metrics": []any{map[string]any{"name": "flow", "alias": 1, "value": 5.0, "quality": "suspect"}}}))
	if _, seen := h.devices["P1"].values["flow"]; seen {
		t.Fatal("a bad birth value was taken as live")
	}
	h.Receive(record(t, "DDATA", 2, map[string]any{"boot_id": "board-01-boot-0", "metrics": []any{map[string]any{"alias": 1, "value": 7.0,
		"quality": "good", "event_time_ns": 1767225660000000000, "time_source": "device"}}}))
	if _, seen := h.devices["P1"].values["flow"]; seen {
		t.Fatal("data from another boot was taken as live")
	}
}

func TestARecordFromAnotherSessionIsIgnored(t *testing.T) {
	t.Parallel()
	h := New(Options{Environment: "simulated", ProtocolMax: 1})
	h.Receive(birth(t, "simulated"))
	if replies := h.Receive(record(t, "NSYNC", 1, map[string]any{"session_id": "s-other", "through_seq": 0})); len(replies) != 0 {
		t.Fatalf("a sync from another session was answered: %v", replies)
	}
}

func TestAKindWhoseContentDiffersFromItsDigestRefusesTheSession(t *testing.T) {
	t.Parallel()
	kinds := birth(t, "simulated").Kinds
	kinds[0].(map[string]any)["metrics"] = []any{}
	h := New(Options{Environment: "simulated", ProtocolMax: 1})
	replies := h.Receive(birth(t, "simulated", map[string]any{"kinds": kinds}))
	if h.Report().Refused != "inventory_mismatch" || len(replies) != 1 || replies[0]["online"] != false || replies[0]["refusal"] != "inventory_mismatch" {
		t.Fatalf("refused %q, replies %v", h.Report().Refused, replies)
	}
}

func TestAKindAnnouncedTwiceUnderOneNameOrItsCaseRefusesTheSession(t *testing.T) {
	t.Parallel()
	for name, rename := range map[string]func(string) string{"the same name": func(n string) string { return n }, "another case": strings.ToUpper} {
		kinds := birth(t, "simulated").Kinds
		twin := maps.Clone(kinds[0].(map[string]any))
		twin["kind"] = rename(twin["kind"].(string))
		digest, err := domain.KindDigest(twin)
		if err != nil {
			t.Fatal(err)
		}
		twin["kind_digest"] = digest
		h := New(Options{Environment: "simulated", ProtocolMax: 1})
		replies := h.Receive(birth(t, "simulated", map[string]any{"kinds": append(kinds, twin)}))
		if h.Report().Refused != "inventory_mismatch" || len(replies) != 1 || replies[0]["refusal"] != "inventory_mismatch" {
			t.Fatalf("%s: refused %q, replies %v, want inventory_mismatch", name, h.Report().Refused, replies)
		}
	}
}
