package mcp

// Slice E (G3): the operator role over a real HTTP endpoint with a real SDK
// client — the deployable process boundary. The reference consumer closes
// the loop out-of-process through operator capabilities: fault -> evidence
// -> detection -> effector -> world effect -> verdict -> score. Director
// tools are unreachable from the operator endpoint.

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/refconsumer"
	"github.com/ghassan-ai-projects/streams-simulator/internal/score"
)

// startOperatorEndpoint binds the operator role over streamable HTTP, the
// same wiring the CLI's --operator-addr uses.
func startOperatorEndpoint(t *testing.T, d *Director) string {
	t.Helper()
	opServer := NewOperatorServerResolver(d)
	handler := mcpsdk.NewStreamableHTTPHandler(func(*http.Request) *mcpsdk.Server { return opServer }, nil)
	ln, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := &http.Server{Handler: handler, ReadHeaderTimeout: 10 * time.Second}
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(func() { _ = srv.Close() })
	endpoint := "http://" + ln.Addr().String()
	d.SetOperatorEndpoint(endpoint)
	return endpoint
}

// connectOperatorClient opens a real MCP client session to the operator
// endpoint over HTTP.
func connectOperatorClient(t *testing.T, endpoint string) *mcpsdk.ClientSession {
	t.Helper()
	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "e2e-test", Version: "0.0.1"}, nil)
	session, err := client.Connect(context.Background(), &mcpsdk.StreamableClientTransport{Endpoint: endpoint}, nil)
	if err != nil {
		t.Fatalf("operator client connect: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session
}

func TestOperatorEndpointServesOnlyOperatorTools(t *testing.T) {
	d := newTestDirector(t)
	endpoint := startOperatorEndpoint(t, d)
	worldID := createWorld(t, d)
	_ = worldID
	session := connectOperatorClient(t, endpoint)
	res, err := session.ListTools(context.Background(), &mcpsdk.ListToolsParams{})
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, tool := range res.Tools {
		names[tool.Name] = true
	}
	for _, want := range []string{"sim.nameplate.read", "sim.effector.list", "sim.effector.invoke", "sim.consumer.report"} {
		if !names[want] {
			t.Fatalf("operator endpoint missing %s", want)
		}
	}
	for _, forbidden := range []string{"sim.world.create", "sim.fault.inject", "sim.truth.reveal", "sim.score", "sim.clock.advance", "sim.perturb.apply"} {
		if names[forbidden] {
			t.Fatalf("operator endpoint must not expose %s", forbidden)
		}
	}
}

// nativeTrace renders captured delivered events as native-format JSONL, the
// bytes a consumer would read from the sink.
func nativeTrace(evs []model.SimEvent) []byte {
	var out []byte
	for i := range evs {
		b, _ := json.Marshal(evs[i])
		out = append(out, b...)
		out = append(out, '\n')
	}
	return out
}

// TestGoldenClosedLoopOverOperatorEndpoint: the reference consumer closes
// the loop out-of-process. The harness drives the world; the consumer reads
// evidence, detects, actuates over MCP, asserts quiescence, submits a
// verdict; the harness scores and the loop resolves.
func TestGoldenClosedLoopOverOperatorEndpoint(t *testing.T) {
	d := newTestDirector(t)
	endpoint := startOperatorEndpoint(t, d)
	start := model.DefaultStartTimeNS + 4*3600*1e9
	created, err := d.CreateWorld(map[string]any{
		"domain": "aquaculture-pond", "seed": float64(5), "adapter": "native-jsonl",
		"sink": model.SinkInproc, "time_mode": model.TimeStepped,
		"start_time": float64(start),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := created["operator_endpoint"]; got != endpoint {
		t.Fatalf("world.create must return the operator endpoint, got %v", got)
	}
	worldID := created["world_id"].(string)
	w := d.Worlds[worldID]
	pond := "site-a/pond-1"

	// Harness setup: aerator on at night, then a fault.
	if _, err := w.Run.InvokeEffector("start_aerator", pond, "setup", map[string]any{"pond_id": pond, "level": 1.0}, start); err != nil {
		t.Fatal(err)
	}
	faultAt := start + 2*3600*1e9
	if _, err := d.InjectFault(worldID, pond, "aerator_failure", faultAt, nil); err != nil {
		t.Fatal(err)
	}
	if err := d.SealTruth(w.Run.ID, &model.GroundTruthRecord{
		ScenarioID: "e2e/0001", Domain: "aquaculture-pond", Label: "aerator_failure",
		EntityID: pond, ExpectedEffector: "start_aerator", ExpectedEpisode: true,
		InjectionTimeNS: faultAt, FirstObservableTimeNS: faultAt, UnavoidableTimeNS: start + 3*3600*1e9,
		TrivialBaselineVerdict: model.TrivialNonTrivial,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := d.BeginRun(worldID, "e2e-golden"); err != nil {
		t.Fatal(err)
	}

	// Capture the delivered evidence and park the quiescence waits.
	var evidence []model.SimEvent
	w.Run.SetEvidenceRecorder(func(e model.SimEvent) { evidence = append(evidence, e) })
	parked := make(chan struct{})
	w.Run.SetQuiesceParkedHook(func() { parked <- struct{}{} })

	// The out-of-process consumer.
	op, err := refconsumer.NewMCPOperator(endpoint, w.Token, w.Run.ID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = op.Close() })
	np, err := op.Nameplate()
	if err != nil {
		t.Fatal(err)
	}
	cfg := refconsumer.DefaultConfig()
	cfg.Threshold = 3 // the crash is loud; the reference detector is not clairvoyant
	cfg.OnDetectionEffector = "start_aerator"
	consumer := refconsumer.New(cfg, np, op, op, w.Run.ID)

	// park blocks until the quiescence wait parks, with a bound so a
	// fast-path wait (watermark already satisfied) cannot hang the test.
	park := func() {
		select {
		case <-parked:
		case <-time.After(5 * time.Second):
			t.Fatal("quiescence wait did not park")
		}
	}

	// Advance 1: the fault develops. The consumer processes the batch and
	// actuates; the quiescence assertion resolves the await.
	t1 := start + 5*3600*1e9
	adv1 := make(chan error, 1)
	go func() {
		_, err := d.Advance(context.Background(), worldID, t1, true)
		adv1 <- err
	}()
	park()
	v1, err := consumer.Process(nativeTrace(evidence), t1)
	if err != nil {
		t.Fatal(err)
	}
	if err := <-adv1; err != nil {
		t.Fatalf("advance 1 did not resolve: %v", err)
	}
	if len(v1.Detections) == 0 {
		t.Fatal("the reference consumer must detect the fault")
	}
	if len(v1.Actions) == 0 || v1.Actions[0].Effector != "start_aerator" {
		t.Fatalf("the reference consumer must actuate the aerator: %+v", v1.Actions)
	}

	// Advance 2: the effect propagates. The consumer processes the new
	// evidence and reports quiescence through the new horizon.
	t2 := t1 + 5*3600*1e9
	adv2 := make(chan error, 1)
	go func() {
		_, err := d.Advance(context.Background(), worldID, t2, true)
		adv2 <- err
	}()
	park()
	if _, err := consumer.Process(nativeTrace(evidence), t2); err != nil {
		t.Fatal(err)
	}
	if err := <-adv2; err != nil {
		t.Fatalf("advance 2 did not resolve: %v", err)
	}

	// The run closes and scores; the loop must resolve. Resolution is
	// asserted on the physical state too: Loop.Resolved alone would pass
	// even without the consumer's mid-wait invoke, because the setup call
	// precedes the fault.
	if got := w.Run.World.StateValue(pond, "aerator_output", w.Run.World.Clock()); got <= 0.9 {
		t.Fatalf("the aerator must have recovered after the consumer's actuation: %v", got)
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
		t.Fatalf("closed loop did not resolve: %+v", sc.Loop)
	}
}

// TestOperatorEndpointPerShippedDomain: every shipped domain can be driven
// through the operator endpoint — nameplate, effector list, an invocation
// with arguments derived from the declared schema, quiescence and a verdict.
func TestOperatorEndpointPerShippedDomain(t *testing.T) {
	specs, err := domain.LoadAll("../../domains")
	if err != nil {
		t.Fatal(err)
	}
	adap, err := adapter.Load(nativeAdapter)
	if err != nil {
		t.Fatal(err)
	}
	d := NewDirector(context.Background(), domain.NewCatalog(specs), map[string]*model.Adapter{"native-jsonl": adap}, t.TempDir())
	endpoint := startOperatorEndpoint(t, d)

	for _, spec := range specs {
		id := spec.Spec.ID
		t.Run(id, func(t *testing.T) {
			created, err := d.CreateWorld(map[string]any{
				"domain": id, "seed": float64(7), "adapter": "native-jsonl",
				"sink": model.SinkInproc, "time_mode": model.TimeStepped,
			})
			if err != nil {
				t.Fatalf("world.create: %v", err)
			}
			worldID := created["world_id"].(string)
			w := d.Worlds[worldID]
			op, err := refconsumer.NewMCPOperator(endpoint, w.Token, w.Run.ID)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = op.Close() }()

			np, err := op.Nameplate()
			if err != nil {
				t.Fatal(err)
			}
			if len(np.Entities) == 0 || len(np.Channels) == 0 {
				t.Fatalf("nameplate incomplete for %s: %+v", id, np)
			}
			effs, err := op.ListEffectors()
			if err != nil {
				t.Fatal(err)
			}
			entity := np.Entities[0].ID
			if len(effs) > 0 {
				res, err := op.InvokeEffector(effs[0].Name, entity, "e2e-"+id, argsForSchema(effs[0].Schema, entity), 0)
				if err != nil {
					t.Fatalf("invoke %s on %s: %v", effs[0].Name, id, err)
				}
				if !res.Accepted {
					t.Fatalf("invoke %s on %s not accepted: %+v", effs[0].Name, id, res)
				}
			}
			if err := op.ReportQuiesced(w.Run.World.Clock()); err != nil {
				t.Fatalf("quiescence report: %v", err)
			}
			if err := op.SubmitVerdict(&model.Verdict{
				SchemaVersion: "0.1", RunID: w.Run.ID,
				Consumer: model.ConsumerInfo{Name: "e2e", Version: "0.0.1"},
			}); err != nil {
				t.Fatalf("verdict submission: %v", err)
			}
		})
	}
}

// argsForSchema synthesizes arguments for an effector from its declared
// argument schema: required *_id strings bind to the entity, numbers to 1.
func argsForSchema(schema map[string]any, entity string) map[string]any {
	out := map[string]any{}
	if schema == nil {
		return out
	}
	props, _ := schema["properties"].(map[string]any)
	required, _ := schema["required"].([]any)
	for _, r := range required {
		name, _ := r.(string)
		if name == "" {
			continue
		}
		p, _ := props[name].(map[string]any)
		switch p["type"] {
		case "string":
			if strings.HasSuffix(name, "_id") {
				out[name] = entity
			} else {
				out[name] = "x"
			}
		case "number", "integer":
			out[name] = 1
		case "boolean":
			out[name] = true
		default:
			out[name] = ""
		}
	}
	return out
}
