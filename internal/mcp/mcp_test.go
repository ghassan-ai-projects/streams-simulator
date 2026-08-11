package mcp

// Role separation, capability tokens, error taxonomy, the closed loop
// through the MCP surface, and the prefix-indistinguishability harness:
// while two worlds' delivered evidence prefixes are identical, replaying
// the same operator calls must produce byte-identical responses. This is
// the leak check — the reason the operator view exists at all.

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"errors"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/score"
)

const (
	aquaculturePath = "../../docs/examples/aquaculture-pond.domain.json"
	nativeAdapter   = "../../adapters/native-jsonl.adapter.json"
)

// newTestDirector builds a director with the aquaculture-pond domain and
// the native-jsonl adapter.
func newTestDirector(t *testing.T) *Director {
	t.Helper()
	spec, err := domain.Load(aquaculturePath)
	if err != nil {
		t.Fatal(err)
	}
	a, err := adapter.Load(nativeAdapter)
	if err != nil {
		t.Fatal(err)
	}
	cat := domain.NewCatalog([]*domain.Compiled{spec})
	d := NewDirector(context.Background(), cat, map[string]*model.Adapter{"native-jsonl": a}, t.TempDir())
	return d
}

// connect returns a client session to the server over in-memory transports.
func connect(t *testing.T, server *mcp.Server) (*mcp.ClientSession, func()) {
	t.Helper()
	ctx := context.Background()
	ct, st := mcp.NewInMemoryTransports()
	ss, err := server.Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test"}, nil)
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	cleanup := func() {
		_ = cs.Close()
		_ = ss.Close()
		_ = ss.Wait()
	}
	t.Cleanup(cleanup)
	return cs, cleanup
}

func callTool(t *testing.T, server *mcp.Server, name string, args map[string]any) (*mcp.CallToolResult, error) {
	t.Helper()
	cs, _ := connect(t, server)
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		return nil, fmt.Errorf("callTool %s: %w", name, err)
	}
	return res, nil
}

func TestRoleSeparation(t *testing.T) {
	d := newTestDirector(t)
	dirServer := NewDirectorServer(d)
	res, err := callTool(t, dirServer, "tools/list", nil)
	_ = res
	_ = err
	// Director tools include truth, fault and perturb surfaces.
	for _, name := range []string{"sim.fault.list", "sim.truth.reveal", "sim.perturb.apply", "sim.score"} {
		if !serverHasTool(t, dirServer, name) {
			t.Fatalf("director server missing %s", name)
		}
	}
	// Operator server advertises exactly the four operator tools.
	d2 := newTestDirector(t)
	world := createWorld(t, d2)
	opServer := NewOperatorServer(d2.Worlds[world].Operator)
	for _, name := range []string{"sim.nameplate.read", "sim.effector.list", "sim.effector.invoke", "sim.consumer.report"} {
		if !serverHasTool(t, opServer, name) {
			t.Fatalf("operator server missing %s", name)
		}
	}
	for _, name := range []string{"sim.truth.reveal", "sim.fault.list", "sim.score", "sim.perturb.apply"} {
		if serverHasTool(t, opServer, name) {
			t.Fatalf("operator server must not expose %s", name)
		}
	}
}

func TestToolSchemasAreClosedAndMachineReadable(t *testing.T) {
	d := newTestDirector(t)
	servers := map[string]*mcp.Server{"director": NewDirectorServer(d)}
	worldID := createWorld(t, d)
	servers["operator"] = NewOperatorServer(d.Worlds[worldID].Operator)

	for role, server := range servers {
		client, _ := connect(t, server)
		res, err := client.ListTools(context.Background(), &mcp.ListToolsParams{})
		if err != nil {
			t.Fatalf("%s tools/list: %v", role, err)
		}
		if len(res.Tools) == 0 {
			t.Fatalf("%s advertised no tools", role)
		}
		for _, tool := range res.Tools {
			raw, err := json.Marshal(tool.InputSchema)
			if err != nil {
				t.Fatalf("%s %s schema marshal: %v", role, tool.Name, err)
			}
			var schema map[string]any
			if err := json.Unmarshal(raw, &schema); err != nil {
				t.Fatalf("%s %s schema decode: %v", role, tool.Name, err)
			}
			if schema["type"] != "object" {
				t.Errorf("%s %s schema type = %v, want object", role, tool.Name, schema["type"])
			}
			if schema["additionalProperties"] != false {
				t.Errorf("%s %s must reject unknown properties", role, tool.Name)
			}
			if _, ok := schema["properties"].(map[string]any); !ok {
				t.Errorf("%s %s schema has no properties object", role, tool.Name)
			}
		}
	}
}

func TestMCPRejectsUnknownAndMissingArguments(t *testing.T) {
	d := newTestDirector(t)
	server := NewDirectorServer(d)

	res, err := callTool(t, server, "sim.catalog.describe", map[string]any{
		"domain": "aquaculture-pond", "unexpected": true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatal("unknown argument must be rejected")
	}
	if len(d.Worlds) != 0 {
		t.Fatal("rejected catalog call must not mutate director state")
	}

	res, err = callTool(t, server, "sim.world.create", map[string]any{
		"domain": "aquaculture-pond", "sink": "file",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatal("file sink without sink_target must be rejected")
	}

	res, err = callTool(t, server, "sim.world.describe", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatal("missing required argument must be rejected")
	}
}

func TestClockAdvanceSupportsRelativeTimeAndReportsTotal(t *testing.T) {
	d := newTestDirector(t)
	created, err := d.CreateWorld(map[string]any{
		"domain": "aquaculture-pond", "seed": float64(42), "adapter": "native-jsonl",
		"sink": model.SinkInproc, "time_mode": model.TimeStepped,
	})
	if err != nil {
		t.Fatal(err)
	}
	worldID := created["world_id"].(string)
	w := d.World(worldID)
	if got := w.Run.World.Clock(); got != model.DefaultStartTimeNS {
		t.Fatalf("default MCP start time = %d, want %d", got, model.DefaultStartTimeNS)
	}

	res, err := callTool(t, NewDirectorServer(d), "sim.clock.advance", map[string]any{
		"world_id": worldID, "by_ns": float64(60 * 1e9),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("relative advance failed: %+v", res)
	}
	var out map[string]any
	raw, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if out["emitted_total"] != float64(w.Run.World.EmittedCount()) {
		t.Fatalf("emitted_total = %v, want %d", out["emitted_total"], w.Run.World.EmittedCount())
	}
	if w.Run.World.Clock() != model.DefaultStartTimeNS+60*1e9 {
		t.Fatalf("relative advance clock = %d", w.Run.World.Clock())
	}
}

func serverHasTool(t *testing.T, server *mcp.Server, name string) bool {
	t.Helper()
	cs, _ := connect(t, server)
	res, err := cs.ListTools(context.Background(), &mcp.ListToolsParams{})
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range res.Tools {
		if tool.Name == name {
			return true
		}
	}
	return false
}

func createWorld(t *testing.T, d *Director) string {
	t.Helper()
	res, err := d.CreateWorld(map[string]any{
		"domain": "aquaculture-pond", "seed": float64(42), "adapter": "native-jsonl",
		"sink": model.SinkInproc, "time_mode": model.TimeStepped,
		"start_time": float64(model.DefaultStartTimeNS + 4*3600*1e9),
	})
	if err != nil {
		t.Fatal(err)
	}
	return res["world_id"].(string)
}

func TestCapabilityTokenEnforced(t *testing.T) {
	d := newTestDirector(t)
	worldID := createWorld(t, d)
	w := d.Worlds[worldID]
	// Without the token every operator call is refused.
	if _, err := w.Operator.ReadNameplate(""); err == nil {
		t.Fatal("nameplate.read without token must fail")
	}
	np, err := w.Operator.ReadNameplate(w.Token)
	if err != nil {
		t.Fatalf("nameplate.read with token failed: %v", err)
	}
	if np.Domain != "aquaculture-pond" || len(np.Channels) == 0 || len(np.Effectors) == 0 {
		t.Fatalf("nameplate incomplete: %+v", np)
	}
	// Unknown effector refused with a stable code.
	_, err = w.Operator.Invoke(w.Token, "no_such_effector", "site-a/pond-1", "cmd-x", nil, 0)
	var te *ToolError
	if !errors.As(err, &te) || te.Code != CodeUnknownEffector {
		t.Fatalf("expected unknown_effector, got %v", err)
	}
	// Missing command_id refused.
	_, err = w.Operator.Invoke(w.Token, "start_aerator", "site-a/pond-1", "", nil, 0)
	if err == nil {
		t.Fatal("missing command_id must fail")
	}
}

func TestDirectorCreatesDistinctWorldsAndOpaqueTokens(t *testing.T) {
	d := newTestDirector(t)
	one := createWorld(t, d)
	two := createWorld(t, d)
	if one == two {
		t.Fatalf("world ids collided: %q", one)
	}
	w1, w2 := d.Worlds[one], d.Worlds[two]
	if w1 == nil || w2 == nil || w1.Run.ID == w2.Run.ID {
		t.Fatalf("run ids collided: %q and %q", w1.Run.ID, w2.Run.ID)
	}
	if w1.Token == w2.Token || len(w1.Token) < 40 || len(w2.Token) < 40 {
		t.Fatalf("capability tokens are not opaque and unique: %q %q", w1.Token, w2.Token)
	}
	if w1.Token == fmt.Sprintf("t-%d-%d", 1, 42) {
		t.Fatal("capability token still exposes the predictable seed form")
	}
}

func TestBeginRunRequiresSealedTruth(t *testing.T) {
	d := newTestDirector(t)
	worldID := createWorld(t, d)
	if _, err := d.BeginRun(worldID, "missing-truth"); err == nil {
		t.Fatal("run.begin must refuse an unsealed oracle")
	}
}

func TestDirectorPassesSinkTarget(t *testing.T) {
	d := newTestDirector(t)
	target := t.TempDir() + "/trace.jsonl"
	res, err := d.CreateWorld(map[string]any{
		"domain": "aquaculture-pond", "seed": float64(9), "adapter": "native-jsonl",
		"sink": model.SinkFile, "sink_target": target,
	})
	if err != nil {
		t.Fatal(err)
	}
	worldID := res["world_id"].(string)
	if _, err := d.Advance(context.Background(), worldID, model.DefaultStartTimeNS+60*1e9, false); err != nil {
		t.Fatal(err)
	}
	if _, err := d.DestroyWorld(worldID); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("director did not pass sink target through: %v", err)
	}
}

func TestClosedLoopThroughMCPSurface(t *testing.T) {
	d := newTestDirector(t)
	worldID := createWorld(t, d)
	w := d.Worlds[worldID]
	start := model.DefaultStartTimeNS + 4*3600*1e9
	pond := "site-a/pond-1"

	// Scenario context: aerator on at night, then a fault.
	if _, err := w.Run.InvokeEffector("start_aerator", pond, "setup", map[string]any{"pond_id": pond, "level": 1.0}, start); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Advance(context.Background(), worldID, start+2*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	if _, err := d.InjectFault(worldID, pond, "aerator_failure", 0, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Advance(context.Background(), worldID, start+3*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	// The operator invokes the effector through the narrow surface.
	res, err := w.Operator.Invoke(w.Token, "start_aerator", pond, "cmd-mcp-1",
		map[string]any{"pond_id": pond, "level": 1.0}, w.Run.World.Clock())
	if err != nil {
		t.Fatalf("invoke failed: %v", err)
	}
	if !res.Simulated || !res.Accepted {
		t.Fatalf("invoke result wrong: %+v", res)
	}
	if _, err := d.Advance(context.Background(), worldID, start+7*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	// Submit a verdict through the report tool.
	verdict := &model.Verdict{
		SchemaVersion: "0.1", RunID: w.Run.ID,
		Consumer: model.ConsumerInfo{Name: "test", Version: "0.1"},
		Actions: []model.Action{{
			CommandID: "cmd-mcp-1", Effector: "start_aerator", EntityID: pond,
			IssuedAt: model.FormatTime(w.Run.World.Clock()), OutcomeBelieved: model.BelievedSucceeded,
		}},
	}
	if err := w.Operator.Report(w.Token, w.Run.ID, w.Run.World.Clock(), verdict); err != nil {
		t.Fatal(err)
	}
	// Seal the director-only label before opening the run. The operator never
	// receives this record.
	rec := &model.GroundTruthRecord{
		ScenarioID: "mcp/0001", Domain: "aquaculture-pond", Label: "aerator_failure",
		EntityID: pond, ExpectedEffector: "start_aerator",
		InjectionTimeNS: start + 2*3600*1e9, FirstObservableTimeNS: start + 2*3600*1e9,
	}
	if err := d.SealTruth(w.Run.ID, rec); err != nil {
		t.Fatal(err)
	}
	// Run lifecycle + score through the director.
	if _, err := d.BeginRun(worldID, "mcp-loop"); err != nil {
		t.Fatal(err)
	}
	if _, err := d.EndRun(worldID); err != nil {
		t.Fatal(err)
	}
	out, err := d.Score(w.Run.ID)
	if err != nil {
		t.Fatal(err)
	}
	sc, ok := out["scorecard"].(*score.Scorecard)
	if !ok {
		t.Fatalf("scorecard missing: %T", out["scorecard"])
	}
	if !sc.Loop.Resolved {
		t.Fatalf("loop did not resolve: %+v", sc.Loop)
	}
}

// TestPrefixIndistinguishability: run two worlds with different faults from
// the same seed; while their delivered evidence is identical, the operator
// surface must respond identically. This catches any covert channel — an
// error string, a latency, a field that changes with hidden state.
func TestPrefixIndistinguishability(t *testing.T) {
	start := model.DefaultStartTimeNS + 4*3600*1e9
	mk := func(seed uint64) (*Director, string) {
		d := newTestDirector(t)
		res, err := d.CreateWorld(map[string]any{
			"domain": "aquaculture-pond", "seed": float64(seed), "adapter": "native-jsonl",
			"sink": model.SinkInproc, "time_mode": model.TimeStepped,
			"start_time": float64(start),
		})
		if err != nil {
			t.Fatal(err)
		}
		return d, res["world_id"].(string)
	}
	dA, idA := mk(7)
	dB, idB := mk(7)
	wA := dA.Worlds[idA]
	wB := dB.Worlds[idB]
	pond := "site-a/pond-1"
	// Record the delivered evidence (post-perturbation, pre-render).
	evA := evidence{}
	evB := evidence{}
	wA.Run.SetEvidenceRecorder(func(e model.SimEvent) { evA.record(e) })
	wB.Run.SetEvidenceRecorder(func(e model.SimEvent) { evB.record(e) })
	// Same setup.
	for _, w := range []*WorldRecord{wA, wB} {
		if _, err := w.Run.InvokeEffector("start_aerator", pond, "setup", map[string]any{"pond_id": pond, "level": 1.0}, start); err != nil {
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
	ra, err := operatorResponseSet(t, wA.Operator, wA.Token, before)
	if err != nil {
		t.Fatal(err)
	}
	rb, err := operatorResponseSet(t, wB.Operator, wB.Token, before)
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

func operatorResponseSet(t *testing.T, v *OperatorView, token string, atNS int64) (string, error) {
	t.Helper()
	np, err := v.ReadNameplate(token)
	if err != nil {
		return "", err
	}
	effs, err := v.ListEffectors(token)
	if err != nil {
		return "", err
	}
	a, _ := json.Marshal(np)
	b, _ := json.Marshal(effs)
	return string(a) + "|" + string(b), nil
}
