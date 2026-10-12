package app

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	wire "github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

type scriptedHost struct {
	t          *testing.T
	sent       []map[string]any
	pending    []wire.Inbound
	hangUp     error
	birthState map[string]any
	lenient    bool
	bySync     map[int][]map[string]any
	syncing    int
	seq        uint64
	refusal    string
	silent     bool
	afterSync  func(h *scriptedHost, syncs int)
	syncs      int
	transcript []byte
}

func (h *scriptedHost) Send(records []map[string]any) error {
	for _, record := range records {
		raw, err := json.Marshal(record)
		if err != nil {
			h.t.Fatal(err)
		}
		if _, err := wire.Parse(raw); err != nil && !h.lenient {
			h.t.Fatalf("the node sent an invalid record: %v", err)
		}
		h.transcript = append(append(h.transcript, raw...), '\n')
		h.sent = append(h.sent, record)
		h.react(record)
	}
	return nil
}

func (h *scriptedHost) react(record map[string]any) {
	switch record["type"] {
	case "NBIRTH":
		if len(h.sent) == 1 && h.birthState != nil {
			h.reply("STATE", h.birthState)
		} else if len(h.sent) == 1 {
			h.reply("STATE", h.state())
		}
	case "NSYNC":
		h.syncs++
		h.syncing = h.syncs
		if h.afterSync != nil {
			h.afterSync(h, h.syncs)
		}
		h.syncing = 0
		if !h.silent {
			h.reply("HSYNC", map[string]any{"through_seq": record["seq"]})
		}
	}
}

func (h *scriptedHost) state() map[string]any {
	state := map[string]any{"host_id": "host-a", "online": h.refusal == "", "heartbeat": 0, "heartbeat_interval_ms": 1000,
		"authority_epoch": 1, "protocol_min": 1, "protocol_max": 1}
	if h.refusal != "" {
		state["refusal"] = h.refusal
	}
	return state
}

func (h *scriptedHost) reply(kind string, fields map[string]any) {
	record := map[string]any{"type": kind, "protocol_version": 1, "session_id": "s-1", "seq": h.seq, "lane": "control"}
	for key, value := range fields {
		record[key] = value
	}
	h.seq++
	h.pending = append(h.pending, wire.Inbound{Record: record})
	if h.syncing > 0 && kind != "HSYNC" {
		if h.bySync == nil {
			h.bySync = map[int][]map[string]any{}
		}
		h.bySync[h.syncing] = append(h.bySync[h.syncing], record)
	}
}

func (h *scriptedHost) replyInvalid(record map[string]any) {
	record["seq"], record["protocol_version"], record["session_id"], record["lane"] = h.seq, 1, "s-1", "control"
	h.seq++
	h.pending = append(h.pending, wire.Inbound{Record: record, Invalid: wire.Check(record)})
}

func (h *scriptedHost) Await(_ context.Context, until func(map[string]any) bool) ([]wire.Inbound, bool, error) {
	for i, arrival := range h.pending {
		if arrival.Invalid == nil && until(arrival.Record) {
			got := h.pending[:i+1]
			h.pending = h.pending[i+1:]
			return got, true, nil
		}
	}
	got := h.pending
	h.pending = nil
	return got, false, h.hangUp
}

func (h *scriptedHost) types() []string {
	var out []string
	for _, record := range h.sent {
		out = append(out, record["type"].(string))
	}
	return out
}

func utilityRun(t *testing.T) *run.Run {
	t.Helper()
	return runOf(t, testsupport.UtilityDomain())
}

func structuredRun(t *testing.T) *run.Run {
	t.Helper()
	return runOf(t, testsupport.StructuredDomain())
}

func runOf(t *testing.T, path string) *run.Run {
	t.Helper()
	spec, err := domain.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	a, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	r, err := run.New(context.Background(), run.Config{Domain: spec, Adapter: a, Seed: 7, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: model.DefaultStartTimeNS})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func session(t *testing.T, host *scriptedHost, slices int64) (*run.Run, error) {
	t.Helper()
	r := utilityRun(t)
	return r, NewSession(r, host, sliceOptions(slices)).Run(context.Background())
}

func sliceOptions(slices int64) Options {
	return Options{Environment: "simulated", ProtocolMax: 1, ContractVersion: "1.0.0", SessionID: "s-1",
		SliceNS: 60e9, UntilNS: model.DefaultStartTimeNS + slices*60e9}
}
