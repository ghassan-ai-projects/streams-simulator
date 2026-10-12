package app

import (
	"bufio"
	"bytes"
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"

	wire "github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
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
	NowNS int64 `json:"now_ns"`
}

type gatedRecord struct {
	gate   int64
	record map[string]any
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

func hostRecordsOf(steps []exchangeStep) (map[string]any, []gatedRecord) {
	var state map[string]any
	var records []gatedRecord
	gate := int64(0)
	for _, step := range steps {
		switch {
		case step.NowNS != 0:
			gate = step.NowNS
		case step.From == "host" && state == nil && step.Record["type"] == "STATE":
			state = fieldsOf(step.Record)
		case step.From == "host":
			records = append(records, gatedRecord{gate: gate, record: step.Record})
		}
	}
	return state, records
}

func fieldsOf(record map[string]any) map[string]any {
	fields := maps.Clone(record)
	for _, envelope := range []string{"type", "seq", "session_id", "protocol_version", "lane"} {
		delete(fields, envelope)
	}
	return fields
}

func TestTheContractsCommandExchangesDriveTheNode(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"expired-command", "stale-epoch", "policy-not-allowed", "duplicate-command", "duplicate-different-content", "safe-stop-from-stale-epoch"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			steps := readExchange(t, name)
			sent := runExchange(t, steps)
			remaining := slices.Clone(sent)
			for _, step := range steps {
				if step.Expect == nil || step.Expect.From != "node" {
					continue
				}
				index := slices.IndexFunc(remaining, func(r map[string]any) bool { return r["type"] == step.Expect.Type && contains(r, step.Expect.Fields) })
				if index < 0 {
					t.Fatalf("the node sent no further %s with %v", step.Expect.Type, step.Expect.Fields)
				}
				remaining = remaining[index+1:]
			}
		})
	}
}

func runExchange(t *testing.T, steps []exchangeStep) []map[string]any {
	t.Helper()
	state, pending := hostRecordsOf(steps)
	host := &scriptedHost{t: t, birthState: state, afterSync: func(h *scriptedHost, syncs int) {
		for len(pending) > 0 && pending[0].gate <= sliceEnd(syncs) {
			h.reply(pending[0].record["type"].(string), fieldsOf(pending[0].record))
			pending = pending[1:]
		}
	}}
	options := sliceOptions(5)
	options.Policies, options.SafeStop = []string{policy}, []string{"safe_stop"}
	if err := NewSession(runOf(t, testsupport.PumpBoardDomain()), host, options).Run(t.Context()); err != nil {
		t.Fatal(err)
	}
	return host.sent
}

func contains(record, fields map[string]any) bool {
	for key, want := range fields {
		if wire.Seq(record[key]) != wire.Seq(want) || (wire.Seq(want) == 0 && record[key] != want) {
			return false
		}
	}
	return true
}
