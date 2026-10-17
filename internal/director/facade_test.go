package director_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/director"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestFacadeDrivesAWorldThroughItsWholeLifecycle(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	created, err := d.CreateWorld(director.WorldRequest{Domain: "aquaculture-pond", Seed: 3})
	if err != nil {
		t.Fatal(err)
	}
	worldID, runID := created["world_id"].(string), created["run_id"].(string)
	entity := created["entity_ids"].([]string)[0]
	if _, err := d.DescribeWorld(worldID); err != nil {
		t.Fatal(err)
	}
	if _, err := d.AdvanceBy(t.Context(), worldID, 60e9, false); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Advance(t.Context(), worldID, model.DefaultStartTimeNS+120e9, false); err != nil {
		t.Fatal(err)
	}
	if state, err := d.ClockState(worldID); err != nil || state["pending_effects"] == nil {
		t.Fatalf("clock state = %v, %v", state, err)
	}
	injected, err := d.InjectFault(worldID, entity, "aerator_failure", model.DefaultStartTimeNS+300e9, nil)
	if err != nil {
		t.Fatal(err)
	}
	if listed, err := d.ListFaults(worldID); err != nil || listed["faults"] == nil {
		t.Fatalf("faults = %v, %v", listed, err)
	}
	if _, err := d.ClearFault(worldID, injected["fault_id"].(string)); err != nil {
		t.Fatal(err)
	}
	applied, err := d.ApplyPerturb(worldID, "drop", nil, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.ClearPerturb(worldID, applied["perturb_id"].(string)); err != nil {
		t.Fatal(err)
	}
	if _, err := d.EnvInject(worldID, "mains", "outage", nil); err == nil || !strings.HasPrefix(err.Error(), "domain_invalid: ") {
		t.Fatalf("env inject without a target: %v", err)
	}
	if _, err := d.RetireEntity(worldID, entity, "decommissioned"); err != nil {
		t.Fatal(err)
	}
	sealAndFinish(t, d, worldID, runID, entity, created["token"].(string))
}

func sealAndFinish(t *testing.T, d *director.Director, worldID, runID, entity, token string) {
	t.Helper()
	truth := &model.GroundTruthRecord{ScenarioID: "f/1", Domain: "aquaculture-pond", Label: "l", EntityID: entity}
	if err := d.SealTruth(runID, truth); err != nil {
		t.Fatal(err)
	}
	if status, err := d.SealStatus(runID); err != nil || status["sealed"] != true {
		t.Fatalf("seal status = %v, %v", status, err)
	}
	if _, err := d.BeginRun(worldID, "facade"); err != nil {
		t.Fatal(err)
	}
	verdict := &model.Verdict{SchemaVersion: "0.1", RunID: runID, Consumer: model.ConsumerInfo{Name: "facade", Version: "0.0.1"}}
	if _, err := d.OperatorReport(token, runID, 0, verdict); err != nil {
		t.Fatal(err)
	}
	ended, err := d.EndRun(worldID)
	if err != nil {
		t.Fatal(err)
	}
	if verified, err := d.VerifyRun(ended["run_artifact_path"].(string)); err != nil || verified["matches"] != true {
		t.Fatalf("verify = %v, %v", verified, err)
	}
	if _, err := d.Score(runID); err != nil {
		t.Fatal(err)
	}
	if revealed, err := d.RevealTruth(runID, false); err != nil || revealed["ground_truth"] == nil {
		t.Fatalf("reveal = %v, %v", revealed, err)
	}
	if _, err := d.DestroyWorld(worldID); err != nil {
		t.Fatal(err)
	}
}

func TestFacadeAuditsAndDescribesTheCatalog(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	if listed := d.CatalogList(""); listed["domains"] == nil {
		t.Fatalf("catalog = %v", listed)
	}
	if described, err := d.CatalogDescribe("aquaculture-pond"); err != nil || described["digest"] == nil {
		t.Fatalf("describe = %v, %v", described, err)
	}
	if d.CatalogCoverage() == nil {
		t.Fatal("coverage must be reported")
	}
	if spec, err := d.DomainSpec("aquaculture-pond"); err != nil || spec == nil {
		t.Fatalf("spec = %v, %v", spec, err)
	}
	adapters := d.ListAdapters()["adapters"].([]map[string]any)
	if len(adapters) != 1 || adapters[0]["id"] != "native-jsonl" {
		t.Fatalf("adapters = %v", adapters)
	}
	// A short window: the facade only delegates, the audit length is proved in app.
	if _, err := d.AuditScenario("aquaculture-pond", "site-a/pond-1", "aerator_failure", 3600e9, 0, 6*3600e9); err != nil {
		t.Fatal(err)
	}
}

func TestFacadeOperatorOperationsNeedTheWorldsCapabilityToken(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	created, err := d.CreateWorld(director.WorldRequest{Domain: "aquaculture-pond"})
	if err != nil {
		t.Fatal(err)
	}
	token, runID := created["token"].(string), created["run_id"].(string)
	if err := d.RequireOperator(token); err != nil {
		t.Fatal(err)
	}
	if np, err := d.OperatorNameplate(token); err != nil || np.Domain != "aquaculture-pond" {
		t.Fatalf("nameplate = %+v, %v", np, err)
	}
	effectors, err := d.OperatorEffectors(token)
	if err != nil || len(effectors) == 0 {
		t.Fatalf("effectors = %v, %v", effectors, err)
	}
	entity := created["entity_ids"].([]string)[0]
	ack, err := d.OperatorInvoke(token, effectors[0].Name, entity, "c-1", testsupport.ArgsForSchema(effectors[0].ArgsSchema, entity))
	if err != nil || !ack.Accepted {
		t.Fatalf("ack = %+v, %v", ack, err)
	}
	if worldID, err := d.OperatorReport(token, runID, 0, nil); err != nil || worldID != created["world_id"] {
		t.Fatalf("report = %q, %v", worldID, err)
	}
	for name, err := range map[string]error{"require": d.RequireOperator("t-wrong"), "report": errOf(d.OperatorReport("t-wrong", runID, 0, nil))} {
		if err == nil || !strings.HasPrefix(err.Error(), string(director.CodeCapabilityDenied)+": ") {
			t.Errorf("%s with a wrong token: %v", name, err)
		}
	}
}

func TestFacadeWorldEffectorsListAndInvokeForTheDirector(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	created, err := d.CreateWorld(director.WorldRequest{Domain: "aquaculture-pond"})
	if err != nil {
		t.Fatal(err)
	}
	worldID, entity := created["world_id"].(string), created["entity_ids"].([]string)[0]
	listed, err := d.WorldEffectors(worldID)
	effectors, _ := listed["effectors"].([]director.EffectorInfo)
	if err != nil || len(effectors) == 0 {
		t.Fatalf("effectors = %v, %v", listed, err)
	}
	ack, err := d.InvokeWorldEffector(worldID, effectors[0].Name, entity, "", testsupport.ArgsForSchema(effectors[0].ArgsSchema, entity))
	if err != nil || !ack.Accepted || ack.CommandID == "" {
		t.Fatalf("ack = %+v, %v", ack, err)
	}
}

func errOf(_ string, err error) error { return err }

func TestFacadeConstructorsAndRefusals(t *testing.T) {
	t.Parallel()
	if _, err := director.NewFromDirs(t.Context(), t.TempDir(), testsupport.AdaptersDir(), t.TempDir(), nil, nil); err == nil || !strings.Contains(err.Error(), "no domains found in") {
		t.Fatalf("an empty domains directory: %v", err)
	}
	d, err := director.NewFromDirs(t.Context(), testsupport.DomainsDir(), testsupport.AdaptersDir(), t.TempDir(), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	d.SetOperatorEndpoint("http://127.0.0.1:9/operator")
	created, err := d.CreateWorld(director.WorldRequest{Domain: "aquaculture-pond"})
	if err != nil || created["operator_endpoint"] != "http://127.0.0.1:9/operator" {
		t.Fatalf("created = %v, %v", created, err)
	}
	refusal := director.ToolErrorf(director.CodeNotImplemented, "no %s", "x")
	if refusal.Error() != "not_implemented: no x" || errors.Unwrap(refusal) != nil {
		t.Fatalf("refusal = %v", refusal)
	}
}
