package host

import (
	"encoding/json"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func record(t *testing.T, kind string, seq uint64, fields map[string]any) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(testsupport.RepositoryRoot(), "contracts", "plant-wire", "fixtures", "valid", kind+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var merged map[string]any
	if err := json.Unmarshal(raw, &merged); err != nil {
		t.Fatal(err)
	}
	maps.Copy(merged, fields)
	if merged["status"] == "accepted" {
		delete(merged, "reason")
	}
	merged["seq"] = seq
	line, _ := json.Marshal(merged)
	parsed, err := domain.Parse(line)
	if err != nil {
		t.Fatalf("%s: %v", kind, err)
	}
	return parsed
}

func birth(t *testing.T, environment string) map[string]any {
	return record(t, "NBIRTH", 0, map[string]any{"environment": environment, "contract_digest": domain.ContractDigest(), "protocol_min": 1, "protocol_max": 1})
}

func pumpBirth(t *testing.T, seq uint64, device string) map[string]any {
	return record(t, "DBIRTH", seq, map[string]any{"device": device})
}

func sample(t *testing.T, seq uint64, value float64, historical bool) map[string]any {
	return record(t, "DDATA", seq, map[string]any{"historical": historical, "metrics": []any{map[string]any{"alias": 1, "value": value, "quality": "good",
		"event_time_ns": 1767225660000000000, "time_source": "device", "uncertainty_ms": 5}}})
}

func sync(t *testing.T, seq uint64, worldNS int64) map[string]any {
	return record(t, "NSYNC", seq, map[string]any{"through_seq": seq - 1, "world_time_ns": worldNS})
}

func TestTheHandshakeAcceptsItsEnvironmentAndRefusesAnother(t *testing.T) {
	t.Parallel()
	for environment, online := range map[string]bool{"simulated": true, "hardware": false} {
		h := New(Options{Environment: "simulated", ProtocolMax: 1, Epoch: 2})
		replies := h.Receive(birth(t, environment))
		if len(replies) != 1 || replies[0]["type"] != "STATE" || replies[0]["online"] != online || domain.Check(replies[0]) != nil {
			t.Fatalf("%s: replies %v", environment, replies)
		}
	}
	for refusal, change := range map[string]map[string]any{
		"contract_unsupported": {"contract_digest": "sha256:" + "0000000000000000000000000000000000000000000000000000000000000000"},
		"protocol_invalid":     {"protocol_min": 2, "protocol_max": 1},
		"version_unsupported":  {"protocol_min": 2, "protocol_max": 2},
	} {
		h := New(Options{Environment: "simulated", ProtocolMax: 1})
		node := birth(t, "simulated")
		maps.Copy(node, change)
		if h.Receive(node); h.Report().Refused != refusal {
			t.Fatalf("refused %q, want %q", h.Report().Refused, refusal)
		}
		if replies := h.Receive(sync(t, 1, 0)); len(replies) != 0 {
			t.Fatalf("a refused host answered %v", replies)
		}
	}
}

func TestNothingButANodeBirthOpensTheSession(t *testing.T) {
	t.Parallel()
	h := New(Options{Environment: "simulated", ProtocolMax: 1})
	if replies := h.Receive(record(t, "NSYNC", 0, map[string]any{"through_seq": 0})); len(replies) != 0 || h.accepted {
		t.Fatalf("a sync before the birth was answered: %v", replies)
	}
}

func TestAGarbledRecordIsIgnoredButConsumesItsSeq(t *testing.T) {
	t.Parallel()
	h := New(Options{Environment: "simulated", ProtocolMax: 1})
	h.Receive(birth(t, "simulated"))
	for seq, kind := range []string{"DBIRTH", "DDATA", "DRECEIPT", "DRESULT", "NSYNC", "NBIRTH"} {
		if replies := h.Receive(map[string]any{"seq": json.Number(string(rune('1' + seq))), "type": kind}); len(replies) != 0 {
			t.Fatalf("%s: replies %v", kind, replies)
		}
	}
	if replies := h.Receive(sync(t, 7, 0)); len(replies) != 1 || replies[0]["type"] != "HSYNC" || h.Report().Rebirths != 0 {
		t.Fatalf("after garbled records: %v, rebirths %d", replies, h.Report().Rebirths)
	}
}

func TestAGapAsksForOneRebirthAndTheRebirthCompletes(t *testing.T) {
	t.Parallel()
	h := New(Options{Environment: "simulated", ProtocolMax: 1})
	h.Receive(birth(t, "simulated"))
	h.Receive(pumpBirth(t, 1, "P1"))
	h.Receive(sample(t, 2, 20, true))
	if h.devices["P1"].values["flow"] != 42 {
		t.Fatal("historical data updated the live value")
	}
	first := h.Receive(sample(t, 4, 30, false))
	if len(first) != 1 || first[0]["type"] != "NCMD" || h.devices["P1"].values["flow"] != 30 {
		t.Fatalf("on a gap: %v, flow %v", first, h.devices["P1"].values["flow"])
	}
	h.Receive(record(t, "NBIRTH", 5, map[string]any{"environment": "simulated", "contract_digest": domain.ContractDigest()}))
	h.Receive(pumpBirth(t, 6, "P1"))
	h.Receive(sample(t, 7, 40, false))
	if h.asked || h.Report().Rebirths != 1 || h.devices["P1"].values["flow"] != 40 {
		t.Fatalf("rebirths %d, asked %v, flow %v", h.Report().Rebirths, h.asked, h.devices["P1"].values["flow"])
	}
}

func scriptedHost(t *testing.T, script ...Scripted) *Host {
	h := New(Options{Environment: "simulated", ProtocolMax: 1, Epoch: 1, Policy: "sha256:" + "b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0",
		Inventory: map[string]string{"P1": "pump"}, Script: script})
	h.Receive(birth(t, "simulated"))
	h.Receive(pumpBirth(t, 1, "P1"))
	h.Receive(pumpBirth(t, 2, "P9"))
	return h
}

func succeed(t *testing.T, h *Host, seq uint64, ids ...string) uint64 {
	for _, id := range ids {
		h.Receive(record(t, "DRECEIPT", seq, map[string]any{"command_id": id, "status": "accepted"}))
		h.Receive(record(t, "DRESULT", seq+1, map[string]any{"command_id": id, "reported_outcome": "succeeded"}))
		seq += 2
	}
	return seq
}

func TestScriptedCommandsAreIssuedAtTheirTimeAndVerifiedByObservation(t *testing.T) {
	t.Parallel()
	verify := &Verify{Metric: "flow", Equals: 50, Tolerance: 1, WithinNS: 100}
	h := scriptedHost(t, Scripted{AtNS: 100, Device: "P1", Command: "start", Args: map[string]any{}, Verify: verify},
		Scripted{AtNS: 300, Device: "P9", Command: "start", Args: map[string]any{}},
		Scripted{AtNS: 100, Device: "P1", Command: "set_speed", Args: map[string]any{"pct": 50}, Verify: verify})
	if replies := h.Receive(sync(t, 3, 50)); len(replies) != 1 {
		t.Fatalf("a command left early: %v", replies)
	}
	replies := h.Receive(sync(t, 4, 100))
	if len(replies) != 3 || replies[0]["type"] != "DCMD" || domain.Check(replies[0]) != nil || replies[2]["type"] != "HSYNC" {
		t.Fatalf("replies %v", replies)
	}
	seq := succeed(t, h, 5, "ref-1", "ref-2")
	h.Receive(sample(t, seq, 50, false))
	if replies := h.Receive(sync(t, seq+1, 300)); len(replies) != 1 {
		t.Fatalf("a command went to a quarantined device: %v", replies)
	}
	report := h.Report()
	if !*report.Commands["ref-1"].Verified || report.Commands["ref-1"].Receipts[0] != "accepted" || len(report.Quarantined) != 1 ||
		report.Commands["ref-3"].Denied != "not_commissioned" {
		t.Fatalf("report %+v", report)
	}
}

func TestAValueThatAlreadyMatchedDoesNotVerifyACommand(t *testing.T) {
	t.Parallel()
	h := scriptedHost(t, Scripted{AtNS: 0, Device: "P1", Command: "start", Args: map[string]any{},
		Verify: &Verify{Metric: "flow", Equals: 42, Tolerance: 1, WithinNS: 100}})
	h.Receive(sync(t, 3, 0))
	seq := succeed(t, h, 4, "ref-1")
	h.Receive(sync(t, seq, 200))
	if verified := h.Report().Commands["ref-1"].Verified; verified == nil || *verified {
		t.Fatalf("verified %v from a value seen before the result", verified)
	}
}

func TestAScriptNamesADeviceACommandAndAPositiveWindow(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`[{"at_ns":0,"device":"","command":"start"}]`, `[{"at_ns":0,"device":"P1","command":"start","verify":{"metric":"flow","equals":1,"tolerance":0,"within_ns":0}}]`} {
		if _, err := ParseScript([]byte(raw)); !errors.Is(err, ErrInvalidStep) {
			t.Fatalf("ParseScript(%s) = %v", raw, err)
		}
	}
	script, err := ParseScript([]byte(`[{"at_ns":0,"device":"P1","command":"start"}]`))
	if err != nil || script[0].Args == nil {
		t.Fatalf("ParseScript = %+v, %v", script, err)
	}
}
