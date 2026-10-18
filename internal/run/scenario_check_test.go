package run_test

import (
	"errors"
	"os"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestRunScenarioHandsTheLoadedDomainToItsCheckAndStopsBeforeAnyRunWhenTheCheckRefuses(t *testing.T) {
	t.Parallel()
	refused := errors.New("the domain is refused")
	out := t.TempDir()
	var seen string
	options := run.ScenarioRunOptions{DomainsDir: testsupport.DomainsDir(), AdaptersDir: testsupport.AdaptersDir(), SinkName: model.SinkInproc, OutDir: out,
		CheckDomain: func(spec *model.DomainSpec) ([]model.ImplementsChecked, error) {
			seen = spec.ID
			return nil, refused
		}}
	_, err := run.RunScenario(t.Context(), scenarioFor(model.DefaultStartTimeNS), options)
	written, readErr := os.ReadDir(out)
	if !errors.Is(err, refused) || seen != "aquaculture-pond" || readErr != nil || len(written) != 0 {
		t.Fatalf("RunScenario = %v after checking %q and writing %d entries (%v), want the refusal before any run", err, seen, len(written), readErr)
	}
}

func TestRunScenarioRunsWhenItsDomainCheckAcceptsAndReportsWhatItChecked(t *testing.T) {
	t.Parallel()
	checked := false
	reported := []model.ImplementsChecked{{Kind: "pond", KindDigest: "sha256:0000"}}
	options := run.ScenarioRunOptions{DomainsDir: testsupport.DomainsDir(), AdaptersDir: testsupport.AdaptersDir(), SinkName: model.SinkInproc, OutDir: t.TempDir(),
		CheckDomain: func(*model.DomainSpec) ([]model.ImplementsChecked, error) {
			checked = true
			return reported, nil
		}}
	summary, err := run.RunScenario(t.Context(), scenarioFor(model.DefaultStartTimeNS), options)
	if err != nil || !checked || summary.RunID == "" || len(summary.ImplementsChecked) != 1 || summary.ImplementsChecked[0] != reported[0] {
		t.Fatalf("RunScenario = %+v, %v with the check run=%v, want a run that reports %v", summary, err, checked, reported)
	}
}

func TestRunScenarioWithoutACheckRunsADomainThatImplementsNothingAndReportsAnEmptyList(t *testing.T) {
	t.Parallel()
	options := run.ScenarioRunOptions{DomainsDir: testsupport.DomainsDir(), AdaptersDir: testsupport.AdaptersDir(), SinkName: model.SinkInproc, OutDir: t.TempDir()}
	summary, err := run.RunScenario(t.Context(), scenarioFor(model.DefaultStartTimeNS), options)
	if err != nil || summary.RunID == "" || summary.ImplementsChecked == nil || len(summary.ImplementsChecked) != 0 {
		t.Fatalf("RunScenario = %+v, %v, want a run that checked nothing", summary, err)
	}
}

func TestADomainThatImplementsAKindFileIsRefusedWhenNoCheckIsGiven(t *testing.T) {
	t.Parallel()
	spec := &model.DomainSpec{ID: "plant", Kinds: []model.Kind{{Name: "fan", Implements: "fan"}}}
	checked, err := run.CheckDomain(nil, spec)
	invalid, ok := model.AsValidationError(err)
	if !ok || invalid.Contract != model.ContractDomainSpec || len(invalid.Issues) != 1 || invalid.Issues[0].Pointer != "/kinds/0/implements" || checked != nil {
		t.Fatalf("CheckDomain(nil) = %v, %v, want a domain refusal at /kinds/0/implements", checked, err)
	}
	given, err := run.CheckDomain(func(*model.DomainSpec) ([]model.ImplementsChecked, error) { return nil, nil }, spec)
	if err != nil || given == nil || len(given) != 0 {
		t.Fatalf("CheckDomain(a check that found nothing) = %v, %v, want an empty list", given, err)
	}
}

func TestADomainThatImplementsNothingPassesWhenNoCheckIsGiven(t *testing.T) {
	t.Parallel()
	spec := &model.DomainSpec{ID: "plant", Kinds: []model.Kind{{Name: "fan"}, {Name: "pump"}}}
	checked, err := run.CheckDomain(nil, spec)
	if err != nil || checked == nil || len(checked) != 0 {
		t.Fatalf("CheckDomain(nil) = %v, %v, want an empty list and no refusal for a domain that implements nothing", checked, err)
	}
}
