package run_test

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestLoadInputsSelectsTheNamedDomainAndAdapterAndNamesAnUnknownOne(t *testing.T) {
	t.Parallel()
	spec, adapterSpec, err := run.LoadInputs(testsupport.DomainsDir(), testsupport.AdaptersDir(), "aquaculture-pond", "native-jsonl")
	if err != nil || spec.Spec.ID != "aquaculture-pond" || adapterSpec.ID != "native-jsonl" {
		t.Fatalf("LoadInputs = %v, %v, %v", spec, adapterSpec, err)
	}
	if _, _, err := run.LoadInputs(testsupport.DomainsDir(), testsupport.AdaptersDir(), "no-such", "native-jsonl"); !errors.Is(err, domain.ErrUnknownDomain) {
		t.Fatalf("an unknown domain: err = %v, want ErrUnknownDomain", err)
	}
	if _, _, err := run.LoadInputs(testsupport.DomainsDir(), testsupport.AdaptersDir(), "aquaculture-pond", "no-such"); !errors.Is(err, adapter.ErrUnknownAdapter) {
		t.Fatalf("an unknown adapter: err = %v, want ErrUnknownAdapter", err)
	}
	emptyDomains, emptyAdapters := t.TempDir(), t.TempDir()
	if _, _, err := run.LoadInputs(emptyDomains, testsupport.AdaptersDir(), "aquaculture-pond", "native-jsonl"); err == nil || !strings.HasSuffix(err.Error(), "no domains found in "+emptyDomains) {
		t.Fatalf("a domains directory without a domain: err = %v", err)
	}
	if _, _, err := run.LoadInputs(testsupport.DomainsDir(), emptyAdapters, "aquaculture-pond", "native-jsonl"); err == nil || !strings.HasSuffix(err.Error(), "no adapters found in "+emptyAdapters) {
		t.Fatalf("an adapters directory without an adapter: err = %v", err)
	}
}

func TestExportScenarioFileReturnsTheRunAsAScenarioAndRefusesChangedInputs(t *testing.T) {
	t.Parallel()
	spec, adapterSpec, err := run.LoadInputs(testsupport.DomainsDir(), testsupport.AdaptersDir(), "aquaculture-pond", "native-jsonl")
	if err != nil {
		t.Fatal(err)
	}
	start := model.DefaultStartTimeNS
	scenario := &model.Scenario{SchemaVersion: model.ScenarioSchemaVersion, ID: "export", Domain: "aquaculture-pond", Adapter: "native-jsonl",
		Seed: 3, World: model.ScenarioWorld{StartTimeNS: start}, Commands: []model.ScenarioCommand{
			{Op: model.OpClockAdvance, Args: map[string]any{"to_ns": start + 600e9}}}}
	dir := t.TempDir()
	r, err := run.New(t.Context(), run.Config{Domain: spec, Adapter: adapterSpec, Seed: scenario.Seed, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start, StartTimeSet: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := r.ExecuteScenario(t.Context(), scenario); err != nil {
		t.Fatal(err)
	}
	if _, err := r.End(dir); err != nil {
		t.Fatal(err)
	}
	dirs := run.Directories{Domains: testsupport.DomainsDir(), Adapters: testsupport.AdaptersDir()}
	artifact := filepath.Join(dir, "run.json")
	exported, err := run.ExportScenarioFile(t.Context(), artifact, dirs)
	if err != nil || exported.Omitted != 0 || exported.Refused != 0 || len(exported.Scenario.Commands) != 1 || model.ValidateScenario(exported.Raw) != nil {
		t.Fatalf("export = %+v, %v", exported, err)
	}
	tamperDomainDigest(t, artifact)
	if _, err := run.ExportScenarioFile(t.Context(), artifact, dirs); !errors.Is(err, run.ErrDigestMismatch) || !strings.HasPrefix(err.Error(), "streamsim: scenario from-run: ") {
		t.Fatalf("a changed domain: err = %v, want ErrDigestMismatch", err)
	}
	if _, err := run.ExportScenarioFile(t.Context(), filepath.Join(dir, "absent.json"), dirs); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a missing artifact: err = %v", err)
	}
}

func tamperDomainDigest(t *testing.T, artifact string) {
	t.Helper()
	raw, err := os.ReadFile(artifact)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	domainRecord, _ := doc["domain"].(map[string]any)
	domainRecord["digest"] = "sha256:" + strings.Repeat("0", 64)
	tampered, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(artifact, tampered, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestReplayFileReproducesAPublishedRunAndNamesWhatItCannotRead(t *testing.T) {
	t.Parallel()
	dirs := run.Directories{Domains: testsupport.DomainsDir(), Adapters: testsupport.AdaptersDir()}
	scenario := scenarioFor(model.DefaultStartTimeNS)
	dir := t.TempDir()
	summary, err := run.RunScenario(t.Context(), scenario, run.ScenarioRunOptions{DomainsDir: dirs.Domains, AdaptersDir: dirs.Adapters,
		SinkName: model.SinkInproc, OutDir: dir})
	if err != nil || summary.RunID == "" || !summary.Reproducible || summary.TraceDigest == "" {
		t.Fatalf("summary = %+v, %v", summary, err)
	}
	res, err := run.ReplayFile(t.Context(), filepath.Join(dir, "run.json"), dirs)
	if err != nil || !res.Matches || res.GotDigest != summary.TraceDigest {
		t.Fatalf("replay = %+v, %v; want a match with digest %s", res, err, summary.TraceDigest)
	}
	if _, err := run.ReplayFile(t.Context(), filepath.Join(dir, "absent.json"), dirs); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a missing artifact: err = %v", err)
	}
}

func TestRunScenarioRefusesAScenarioThatBreaksItsDomainBeforeAnySinkOpens(t *testing.T) {
	t.Parallel()
	scenario := scenarioFor(model.DefaultStartTimeNS)
	scenario.World.Profile = "no-such-profile"
	options := run.ScenarioRunOptions{DomainsDir: testsupport.DomainsDir(), AdaptersDir: testsupport.AdaptersDir(), SinkName: model.SinkInproc}
	_, err := run.RunScenario(t.Context(), scenario, options)
	var check *run.ScenarioCheckError
	var invalid *model.ValidationError
	if !errors.As(err, &check) || !errors.As(check.Err, &invalid) || invalid.Issues[0].Pointer != "/world/profile" {
		t.Fatalf("err = %v, want a ScenarioCheckError holding the located violation", err)
	}
	scenario.World.Profile = ""
	scenario.Domain = "no-such"
	if _, err := run.RunScenario(t.Context(), scenario, options); !errors.Is(err, domain.ErrUnknownDomain) || errors.As(err, &check) {
		t.Fatalf("an unknown domain: err = %v, want ErrUnknownDomain that is not a check violation", err)
	}
}

func scenarioFor(start int64) *model.Scenario {
	return &model.Scenario{SchemaVersion: model.ScenarioSchemaVersion, ID: "file", Domain: "aquaculture-pond", Adapter: "native-jsonl",
		Seed: 3, World: model.ScenarioWorld{StartTimeNS: start}, Commands: []model.ScenarioCommand{
			{Op: model.OpClockAdvance, Args: map[string]any{"to_ns": start + 600e9}}}}
}
