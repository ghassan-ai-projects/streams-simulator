package domain

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// examplePath is the committed closed-loop showcase domain.
var examplePath = filepath.Join("..", "..", "docs", "examples", "aquaculture-pond.domain.json")

func TestLoadAquaculturePond(t *testing.T) {
	c, err := Load(examplePath)
	if err != nil {
		t.Fatalf("load example domain: %v", err)
	}
	if c.Spec.ID != "aquaculture-pond" {
		t.Fatalf("wrong id: %s", c.Spec.ID)
	}
	if !c.HasState("dissolved_oxygen_true") {
		t.Fatal("missing state")
	}
	if !c.HasChannel("pond.dissolved_oxygen") {
		t.Fatal("missing channel")
	}
	if !c.HasFault("do_probe_fouling") {
		t.Fatal("missing fault")
	}
	if !c.HasEffector("start_aerator") {
		t.Fatal("missing effector")
	}
	for _, p := range []string{"nominal", "correlated_cascade", "sensor_pathology"} {
		if !c.HasProfile(p) {
			t.Fatalf("missing profile %s", p)
		}
	}
	if c.ChannelGain("pond.aerator_current") != 18 {
		t.Fatalf("channel gain default wrong: %v", c.ChannelGain("pond.aerator_current"))
	}
	if c.ChannelGain("pond.dissolved_oxygen") != 1 {
		t.Fatalf("implicit gain must default to 1: %v", c.ChannelGain("pond.dissolved_oxygen"))
	}
}

func TestDigestStable(t *testing.T) {
	a, err := Load(examplePath)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Load(examplePath)
	if err != nil {
		t.Fatal(err)
	}
	if a.Digest != b.Digest {
		t.Fatalf("digest unstable: %s != %s", a.Digest, b.Digest)
	}
	if !strings.HasPrefix(a.Digest, "sha256:") {
		t.Fatalf("bad digest shape: %s", a.Digest)
	}
}

func TestCrossCheckRejectsBadReferences(t *testing.T) {
	// Copy the example and break one cross-reference.
	c, err := Load(examplePath)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(c.Spec)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	channels := doc["channels"].([]any)
	ch0 := channels[0].(map[string]any)
	ch0["observes"] = "no_such_state"
	b, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Parse(b, "broken"); err == nil || !strings.Contains(err.Error(), "undeclared state") {
		t.Fatalf("expected undeclared-state error, got %v", err)
	}
}

func TestSchemaValidationRejectsMalformed(t *testing.T) {
	// A spec missing required fields must fail the schema, not the
	// cross-check.
	bad := `{"id": "x"}`
	if _, err := Parse([]byte(bad), "bad"); err == nil {
		t.Fatal("malformed spec accepted")
	}
	// A numeric channel without a unit must fail.
	noUnit := `{
		"id": "test-domain", "version": "0.1.0", "title": "t", "stresses": "` + strings.Repeat("x", 40) + `",
		"axes": {"rate": "low", "cardinality": "singleton", "value_shape": ["scalar"], "cadence": ["periodic"], "lateness": "none", "absence": "signal", "time_reference": "wall", "correlation": ["independent"], "seasonality": ["none"], "actuation": "observe_only", "consequence": ["cost"], "fidelity": ["F0"]},
		"entities": {"id_template": "e-{n}", "count": {"default": 1}},
		"channels": [{"name": "ch", "value_type": "number", "resolution": 0.1, "fidelity": "F0", "absence": "signal", "cadence": {"mode": "periodic", "period_s": 60}, "noise": {"model": "gaussian", "sigma": 0.1}}],
		"faults": [{"id": "f1", "onset": {"shape": "step"}, "affects": [], "observability": {"detector": {"form": "single_channel_snr", "channel": "ch"}}}],
		"profiles": [
			{"name": "nominal", "description": "x"},
			{"name": "correlated_cascade", "description": "x", "not_applicable": "` + strings.Repeat("y", 20) + `"},
			{"name": "sensor_pathology", "description": "x"}
		],
		"ground_truth": {"negative_class_fraction": 0.4}
	}`
	if _, err := Parse([]byte(noUnit), "nounit"); err == nil || !strings.Contains(err.Error(), "unit") {
		t.Fatalf("expected unit requirement error, got %v", err)
	}
}

func TestCatalogCoverage(t *testing.T) {
	c, err := Load(examplePath)
	if err != nil {
		t.Fatal(err)
	}
	cat := NewCatalog([]*Compiled{c})
	entries := cat.List("")
	if len(entries) != 1 || entries[0].ID != "aquaculture-pond" {
		t.Fatalf("list wrong: %+v", entries)
	}
	if _, err := cat.Describe("aquaculture-pond"); err != nil {
		t.Fatal(err)
	}
	if _, err := cat.Describe("nope"); err == nil {
		t.Fatal("describe of unknown domain must fail")
	}
	// The twelve axis groups must all be present.
	cov := cat.Coverage()
	if len(cov.ByAxis) != 12 {
		t.Fatalf("expected 12 axes, got %d", len(cov.ByAxis))
	}
	if cov.ByAxis["rate"]["medium"] != 1 {
		t.Fatalf("rate coverage wrong: %+v", cov.ByAxis["rate"])
	}
	if len(cov.Thin) == 0 {
		t.Fatal("single-domain catalog must report thin coverage")
	}
}

// TestF2Rejected: F2 reference models are not built; a domain that declares
// them must fail loudly, never silently emit a static state.
func TestF2Rejected(t *testing.T) {
	spec := minimalSpec()
	spec.Dynamics = []model.Dynamics{{
		Target: "x",
		Tier:   "F2",
		DTMs:   1000,
		F2:     &model.F2Dyn{Model: "soil_water_balance"},
	}}
	raw, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Parse(raw, "f2"); err == nil || !strings.Contains(err.Error(), "not_implemented") {
		t.Fatalf("expected F2 not_implemented rejection, got %v", err)
	}
}

func TestDeadTimeDoubleApplicationRejected(t *testing.T) {
	// An effector with dead_time_s driving a dead_time state must be refused.
	spec := minimalSpec()
	spec.Effectors = []model.Effector{{
		Name:       "do_it",
		ArgsSchema: map[string]any{"type": "object"},
		Ack:        model.Ack{LatencyMS: model.Latency{Mean: 100}},
		Effect: model.Effect{
			StateDeltas:   []model.StateDelta{{State: "x", Delta: 1}},
			TimeConstantS: 100,
			DeadTimeS:     50,
		},
	}}
	spec.Dynamics = []model.Dynamics{{
		Target: "x",
		Tier:   "F1",
		DTMs:   1000,
		F1:     &model.F1Dyn{Form: "dead_time", TimeConstantS: 100, DeadTimeS: 0},
	}}
	raw, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Parse(raw, "dt"); err == nil || !strings.Contains(err.Error(), "twice") {
		t.Fatalf("expected double-dead-time rejection, got %v", err)
	}
}

// minimalSpec builds a valid tiny spec for cross-check tests.
func minimalSpec() *model.DomainSpec {
	return &model.DomainSpec{
		ID:       "minimal",
		Version:  "0.1.0",
		Title:    "minimal",
		Stresses: strings.Repeat("z", 40),
		Axes: model.Axes{
			Rate: "low", Cardinality: "singleton", ValueShape: []string{"scalar"},
			Cadence: []string{"periodic"}, Lateness: "none", Absence: "signal",
			TimeRef: "wall", Correlation: []string{"independent"},
			Seasonality: []string{"none"}, Actuation: "observe_only",
			Consequence: []string{"cost"}, Fidelity: []string{"F0"},
		},
		Entities: model.Entities{IDTemplate: "e-{n}", Count: struct {
			Default int `json:"default"`
			Min     int `json:"min,omitempty"`
			Max     int `json:"max,omitempty"`
		}{Default: 1}},
		State: []model.State{{Name: "x", Initial: 0}},
		Channels: []model.Channel{{
			Name: "ch", ValueType: "number", Unit: "u", Resolution: 0.1,
			Fidelity: "F0", Absence: "signal", Observes: "x",
			Cadence: model.Cadence{Mode: "periodic", PeriodS: 60},
			Noise:   model.Noise{Model: "gaussian", Sigma: 0.1},
		}},
		Faults: []model.Fault{{
			ID: "f1", Onset: model.FaultOnset{Shape: "step"},
			Affects: []model.FaultEffect{{State: "x", Delta: 1}},
			Observability: model.Observability{
				Detector: model.Detector{Form: "single_channel_snr", Channel: "ch"},
			},
		}},
		Profiles: []model.Profile{
			{Name: "nominal", Description: "x"},
			{Name: "correlated_cascade", Description: "x", NotApplicable: strings.Repeat("y", 20)},
			{Name: "sensor_pathology", Description: "x"},
		},
		GroundTruth: model.GroundTruth{NegativeClassFraction: 0.4},
	}
}
