package acceptance

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	directormod "github.com/ghassan-ai-projects/streams-simulator/internal/director"
	"github.com/ghassan-ai-projects/streams-simulator/internal/mcp"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

// TestPrefixIndistinguishability: run two worlds with different faults from
// the same seed; while their delivered evidence is identical, the operator
// surface must respond identically. This catches any covert channel — an
// error string, a latency, a field that changes with hidden state.
func TestPrefixIndistinguishability(t *testing.T) {
	t.Parallel()
	start := model.DefaultStartTimeNS + 4*3600*1e9
	mk := func(seed int64) (*directormod.Director, string, string) {
		d := newCompositionDirector(t)
		res, err := d.CreateWorld(directormod.WorldRequest{
			Domain: "aquaculture-pond", Seed: seed, Adapter: "native-jsonl",
			Sink: model.SinkInproc, TimeMode: model.TimeStepped, StartTimeNS: &start,
		})
		if err != nil {
			t.Fatal(err)
		}
		return d, res["world_id"].(string), res["token"].(string)
	}
	dA, idA, tokenA := mk(7)
	dB, idB, tokenB := mk(7)
	wA := dA.RunOf(idA)
	wB := dB.RunOf(idB)
	pond := "site-a/pond-1"
	// Record the delivered evidence (post-perturbation, pre-render).
	evA := evidence{}
	evB := evidence{}
	wA.SetEvidenceRecorder(func(e model.SimEvent) { evA.record(e) })
	wB.SetEvidenceRecorder(func(e model.SimEvent) { evB.record(e) })
	// Same setup.
	for _, w := range []*run.Run{wA, wB} {
		if _, err := w.InvokeEffector("start_aerator", pond, "setup", map[string]any{"pond_id": pond, "level": 1.0}, start); err != nil {
			t.Fatal(err)
		}
	}
	// Different faults: fouling vs algae crash. Their early evidence is
	// identical (both subtle).
	if _, err := dA.InjectFault(idA, pond, "do_probe_fouling", start+2*3600*1e9, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := dB.InjectFault(idB, pond, "algae_bloom_crash", start+2*3600*1e9, nil); err != nil {
		t.Fatal(err)
	}

	step := int64(30 * 60 * 1000000000)
	divergedAt := int64(0)
	for now := start + step; now <= start+12*3600*1e9; now += step {
		if _, err := dA.Advance(context.Background(), idA, now, false); err != nil {
			t.Fatal(err)
		}
		if _, err := dB.Advance(context.Background(), idB, now, false); err != nil {
			t.Fatal(err)
		}
		if evA.equalPrefix(&evB) {
			continue
		}
		divergedAt = now
		break
	}
	if divergedAt == 0 {
		t.Fatal("evidence never diverged; the harness needs distinguishable faults")
	}
	// Before the divergence the operator responses were identical: replay
	// the same calls against the pre-divergence state.
	before := divergedAt - step
	ra, err := operatorResponseSet(t, dA, tokenA, before)
	if err != nil {
		t.Fatal(err)
	}
	rb, err := operatorResponseSet(t, dB, tokenB, before)
	if err != nil {
		t.Fatal(err)
	}
	if ra != rb {
		t.Fatalf("operator responses diverged before the evidence did:\nA: %s\nB: %s", ra, rb)
	}
}

// evidence is the delivered record stream in delivery order.
type evidence struct {
	entries []string // canonical rendering of each delivered record
}

func (e *evidence) record(ev model.SimEvent) {
	b, _ := json.Marshal(ev)
	e.entries = append(e.entries, string(b))
}

func (e *evidence) equalPrefix(o *evidence) bool {
	n := len(e.entries)
	if len(o.entries) < n {
		n = len(o.entries)
	}
	for i := 0; i < n; i++ {
		if e.entries[i] != o.entries[i] {
			return false
		}
	}
	return len(e.entries) == len(o.entries)
}

// operatorResponseSet renders the full operator surface over the server
// tool layer — nameplate, effector list, an invocation, a quiescence
// report, and the capability-denied error path — as one byte string. Two
// worlds with identical delivered prefixes must produce identical bytes.
func operatorResponseSet(t *testing.T, d *directormod.Director, token string, atNS int64) (string, error) {
	t.Helper()
	cs := connect(t, mcp.NewOperatorServerResolver(d))
	call := func(name string, args map[string]any) (string, error) {
		res, err := cs.CallTool(context.Background(), &mcpsdk.CallToolParams{Name: name, Arguments: args})
		if err != nil {
			return "", fmt.Errorf("operator call %s: %w", name, err)
		}
		raw, _ := json.Marshal(res.StructuredContent)
		return string(raw), nil
	}
	np, err := call("sim.nameplate.read", map[string]any{"token": token})
	if err != nil {
		return "", err
	}
	effs, err := call("sim.effector.list", map[string]any{"token": token})
	if err != nil {
		return "", err
	}
	invoked := "none"
	var effList []directormod.EffectorInfo
	if err := json.Unmarshal([]byte(effs), &effList); err == nil && len(effList) > 0 {
		invoked, err = call("sim.effector.invoke", map[string]any{
			"token": token, "effector": effList[0].Name,
			"entity_id": "site-a/pond-1", "command_id": "prefix-1",
			"args": testsupport.ArgsForSchema(effList[0].ArgsSchema, "site-a/pond-1"), "at_ns": atNS,
		})
		if err != nil {
			return "", err
		}
	}
	report, err := call("sim.consumer.report", map[string]any{
		"token": token, "run_id": "", "quiesced_through_ns": atNS,
	})
	if err != nil {
		return "", err
	}
	denied, err := call("sim.nameplate.read", map[string]any{"token": "bad-token"})
	if err != nil {
		return "", err
	}
	return np + "|" + effs + "|" + invoked + "|" + report + "|" + denied, nil
}
