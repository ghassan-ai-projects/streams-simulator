package acceptance

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/refconsumer"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

// director drives the director role of a running streamsim process.
type director struct {
	t       *testing.T
	session *mcpsdk.ClientSession
}

func startDirector(t *testing.T, out string) *director {
	t.Helper()
	root := testsupport.RepositoryRoot()
	bin := buildBinary(t, root, "")
	// #nosec G204 -- the test runs the freshly built binary with fixed args
	cmd := exec.CommandContext(t.Context(), bin, "mcp", "--role", "director", "--operator-addr", "127.0.0.1:0",
		"--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir(), "--out", out)
	cmd.Dir = root
	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "acceptance"}, nil)
	session, err := client.Connect(t.Context(), &mcpsdk.CommandTransport{Command: cmd}, nil)
	if err != nil {
		t.Fatalf("connect to the director process: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return &director{t: t, session: session}
}

// call invokes a director tool and returns its structured result.
func (d *director) call(name string, args map[string]any) map[string]any {
	d.t.Helper()
	res, err := d.session.CallTool(context.Background(), &mcpsdk.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		d.t.Fatalf("%s: %v", name, err)
	}
	if res.IsError {
		d.t.Fatalf("%s refused: %+v", name, res.Content)
	}
	raw, err := json.Marshal(res.StructuredContent)
	if err != nil {
		d.t.Fatal(err)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		d.t.Fatalf("%s result: %v", name, err)
	}
	return out
}

func TestOutOfProcessConsumerClosesTheLoopThroughTheBinary(t *testing.T) {
	t.Parallel()
	out := t.TempDir()
	trace := filepath.Join(out, "trace.jsonl")
	d := startDirector(t, out)
	start := model.DefaultStartTimeNS + 4*3600*1e9
	const pond = "site-a/pond-1"

	created := d.call("sim.world.create", map[string]any{
		"domain": "aquaculture-pond", "seed": 5, "adapter": "native-jsonl", "time_mode": model.TimeStepped,
		"start_time": start, "sink": model.SinkFile, "sink_target": trace,
	})
	worldID, _ := created["world_id"].(string)
	token, _ := created["token"].(string)
	endpoint, _ := created["operator_endpoint"].(string)
	runID, _ := created["run_id"].(string)
	if !strings.HasPrefix(endpoint, "http://127.0.0.1:") {
		t.Fatalf("the director must advertise its operator endpoint, got %q", endpoint)
	}

	operator, err := refconsumer.NewMCPOperator(endpoint, token, runID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = operator.Close() })
	setup, err := operator.InvokeEffector("start_aerator", pond, "setup", map[string]any{"pond_id": pond, "level": 1.0}, start)
	if err != nil || !setup.Accepted {
		t.Fatalf("setup actuation: %+v, %v", setup, err)
	}

	faultAt := start + 2*3600*1e9
	d.call("sim.fault.inject", map[string]any{"world_id": worldID, "entity_id": pond, "fault": "aerator_failure", "onset_ns": faultAt})
	d.call("sim.truth.seal", map[string]any{"run_id": runID, "ground_truth": map[string]any{
		"scenario_id": "acceptance/0001", "domain": "aquaculture-pond", "seed": 5, "entity_id": pond,
		"label": "aerator_failure", "expected_episode": true, "expected_effector": "start_aerator",
		"injection_time_ns": faultAt, "first_observable_time_ns": faultAt, "unavoidable_time_ns": start + 3*3600*1e9,
		"trivial_baseline_verdict": model.TrivialNonTrivial,
		"observability": map[string]any{"detector_form": "single_channel_snr", "channels": []any{"dissolved_oxygen"},
			"effective_sigma": 0.1, "first_observable_snr": 3.0, "unavoidable_snr": 6.0},
	}})
	d.call("sim.run.begin", map[string]any{"world_id": worldID, "label": "acceptance"})

	horizon := start + 5*3600*1e9
	d.call("sim.clock.advance", map[string]any{"world_id": worldID, "to_ns": horizon})
	raw, err := os.ReadFile(trace)
	if err != nil || len(raw) == 0 {
		t.Fatalf("the director must have delivered a trace to %s: %d bytes, %v", trace, len(raw), err)
	}
	nameplate, err := operator.Nameplate()
	if err != nil {
		t.Fatal(err)
	}
	cfg := refconsumer.DefaultConfig()
	cfg.Threshold = 3
	cfg.OnDetectionEffector = "start_aerator"
	verdict, err := refconsumer.New(cfg, nameplate, operator, operator, runID).Process(raw, horizon)
	if err != nil {
		t.Fatal(err)
	}
	if len(verdict.Detections) == 0 || len(verdict.Actions) == 0 {
		t.Fatalf("the consumer must detect the fault and actuate: %+v", verdict)
	}

	d.call("sim.clock.advance", map[string]any{"world_id": worldID, "by_ns": 5 * 3600 * 1e9})
	ended := d.call("sim.run.end", map[string]any{"world_id": worldID})
	if ended["reproducible"] != true {
		t.Fatalf("run.end = %v", ended)
	}
	scored := d.call("sim.score", map[string]any{"run_id": runID})
	card, _ := scored["scorecard"].(map[string]any)
	judgment, _ := card["judgment"].(map[string]any)
	loop, _ := card["loop"].(map[string]any)
	if judgment["detected"] != true || loop["resolved"] != true || card["reproducible"] != true {
		t.Fatalf("the closed loop must detect, actuate and resolve reproducibly: %v", card)
	}
}
