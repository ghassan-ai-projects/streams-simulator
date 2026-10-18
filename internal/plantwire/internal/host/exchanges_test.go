package host

import (
	"bufio"
	"bytes"
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

type exchangeStep struct {
	From   string         `json:"from"`
	Record map[string]any `json:"record"`
	Expect *struct {
		From   string         `json:"from"`
		Type   string         `json:"type"`
		Fields map[string]any `json:"fields"`
	} `json:"expect"`
}

const (
	exchangeStartNS = int64(1767225660000000000)
	exchangeWindow  = int64(60e9)
	contractPolicy  = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

var exchangeChecks = map[string]*Verify{
	"session-start":          nil,
	"environment-refusal":    nil,
	"kind-digest-mismatch":   nil,
	"version-negotiation":    nil,
	"gap-rebirth":            nil,
	"verified-result":        {Metric: "speed_setpoint", Equals: 80, Tolerance: 0.5, WithinNS: exchangeWindow},
	"lying-result-setpoint":  {Metric: "speed_setpoint", Equals: 80, Tolerance: 0.5, WithinNS: exchangeWindow},
	"lying-result-start":     {Metric: "flow", Equals: 90, Tolerance: 40, WithinNS: exchangeWindow},
	"lying-result-safe-stop": {Metric: "flow", Equals: 0, Tolerance: 0.5, WithinNS: exchangeWindow},
}

func TestTheContractsExchangesDriveTheReferenceHost(t *testing.T) {
	t.Parallel()
	for name, verify := range exchangeChecks {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			steps := readExchange(t, name)
			h := New(Options{Environment: "simulated", ProtocolMax: 1, Epoch: 1, Policy: contractPolicy,
				Inventory: map[string]string{"P1": "pump"}, Script: scriptOf(steps, verify)})
			sent := playNode(t, h, steps)
			expectHostRecords(t, h, steps, sent)
		})
	}
}

func readExchange(t *testing.T, name string) []exchangeStep {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(testsupport.RepositoryRoot(), "contracts", "plant-wire", "fixtures", "exchanges", name+".jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	var steps []exchangeStep
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	for scanner.Scan() {
		var step exchangeStep
		if err := json.Unmarshal(scanner.Bytes(), &step); err != nil {
			t.Fatal(err)
		}
		steps = append(steps, step)
	}
	return steps[1:]
}

func scriptOf(steps []exchangeStep, verify *Verify) []Scripted {
	var script []Scripted
	for _, step := range steps {
		if step.From == "host" && step.Record["type"] == "DCMD" {
			r := step.Record
			script = append(script, Scripted{ID: r["command_id"].(string), Device: r["device"].(string), Command: r["command"].(string),
				Args: r["args"].(map[string]any), Verify: verify})
		}
	}
	return script
}

func playNode(t *testing.T, h *Host, steps []exchangeStep) []map[string]any {
	t.Helper()
	var sent []map[string]any
	offset := uint64(0)
	feed := func(record map[string]any) {
		line, _ := json.Marshal(record)
		in, ok := domain.Decode(line)
		if !ok || in.Invalid != nil {
			t.Fatalf("an exchange record is invalid: %s: %v", line, in.Invalid)
		}
		sent = append(sent, h.Receive(in)...)
	}
	sync := func(seq uint64, worldNS int64) map[string]any {
		return map[string]any{"type": "NSYNC", "protocol_version": 1, "session_id": "s-1", "seq": seq, "lane": "control",
			"through_seq": seq - 1, "world_time_ns": worldNS}
	}
	last := uint64(0)
	for _, step := range steps {
		switch {
		case step.From == "node":
			record := maps.Clone(step.Record)
			last = uint64(record["seq"].(float64)) + offset
			record["seq"] = last
			feed(record)
		case step.From == "host" && step.Record["type"] == "DCMD" && offset == 0:
			offset++
			feed(sync(last+1, exchangeStartNS))
		}
	}
	if offset > 0 {
		feed(sync(last+1, exchangeStartNS+exchangeWindow))
	}
	return sent
}

func expectHostRecords(t *testing.T, h *Host, steps []exchangeStep, sent []map[string]any) {
	t.Helper()
	remaining := sent
	for _, step := range steps {
		switch {
		case step.From == "host" && step.Record["type"] == "DCMD":
			index := slices.IndexFunc(sent, func(r map[string]any) bool { return r["command_id"] == step.Record["command_id"] })
			if index < 0 || !sameCommand(sent[index], step.Record) {
				t.Fatalf("the host's command differs from the contract's:\n got %v\nwant %v", sent, step.Record)
			}
		case step.Expect == nil || step.Expect.From != "host":
		case step.Expect.Type == "verdict":
			verdict := h.Report().Commands[step.Expect.Fields["command_id"].(string)]
			if verdict == nil || verdict.Verified == nil || *verdict.Verified != step.Expect.Fields["verified"] {
				t.Fatalf("verdict %+v, want %v", verdict, step.Expect.Fields)
			}
		default:
			index := slices.IndexFunc(remaining, func(r map[string]any) bool { return r["type"] == step.Expect.Type && contains(r, step.Expect.Fields) })
			if index < 0 {
				t.Fatalf("the host sent no further %s with %v; it sent %v", step.Expect.Type, step.Expect.Fields, sent)
			}
			remaining = remaining[index+1:]
		}
	}
}

func sameCommand(got, want map[string]any) bool {
	for _, key := range []string{"command_id", "idempotency_key", "device", "command", "expected_boot_id", "authority_epoch", "policy_digest", "fallback_on_lease_loss"} {
		if !sameValue(got[key], want[key]) {
			return false
		}
	}
	return sameValue(got["args"], want["args"])
}

func contains(record, fields map[string]any) bool {
	for key, want := range fields {
		if !sameValue(record[key], want) {
			return false
		}
	}
	return true
}

func sameValue(a, b any) bool {
	left, _ := json.Marshal(a)
	right, _ := json.Marshal(b)
	return bytes.Equal(left, right)
}
