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
		CheckDomain: func(spec *model.DomainSpec) error {
			seen = spec.ID
			return refused
		}}
	_, err := run.RunScenario(t.Context(), scenarioFor(model.DefaultStartTimeNS), options)
	written, readErr := os.ReadDir(out)
	if !errors.Is(err, refused) || seen != "aquaculture-pond" || readErr != nil || len(written) != 0 {
		t.Fatalf("RunScenario = %v after checking %q and writing %d entries (%v), want the refusal before any run", err, seen, len(written), readErr)
	}
}

func TestRunScenarioRunsWhenItsDomainCheckAccepts(t *testing.T) {
	t.Parallel()
	checked := false
	options := run.ScenarioRunOptions{DomainsDir: testsupport.DomainsDir(), AdaptersDir: testsupport.AdaptersDir(), SinkName: model.SinkInproc, OutDir: t.TempDir(),
		CheckDomain: func(*model.DomainSpec) error {
			checked = true
			return nil
		}}
	summary, err := run.RunScenario(t.Context(), scenarioFor(model.DefaultStartTimeNS), options)
	if err != nil || !checked || summary.RunID == "" {
		t.Fatalf("RunScenario = %+v, %v with the check run=%v, want a run", summary, err, checked)
	}
}
