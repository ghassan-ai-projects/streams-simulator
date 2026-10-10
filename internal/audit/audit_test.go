package audit_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/audit"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestNewPanelRefusesAMissingSpec(t *testing.T) {
	t.Parallel()
	if _, err := audit.NewPanel(nil, 1, 60e9); !errors.Is(err, audit.ErrNoSpec) {
		t.Fatalf("err = %v, want ErrNoSpec", err)
	}
}

func TestAuditReportsEveryDetectorScoreAndABestDetector(t *testing.T) {
	t.Parallel()
	spec, err := domain.Load(filepath.Join("..", "..", "domains", "aquaculture-pond.domain.json"))
	if err != nil {
		t.Fatal(err)
	}
	panel, err := audit.NewPanel(spec, 42, 60e9)
	if err != nil {
		t.Fatal(err)
	}
	start := model.DefaultStartTimeNS + 4*3600e9
	ids := []string{"site-a/pond-1", "site-a/pond-2"}
	setup := []model.SetupCall{{
		Effector: "start_aerator", EntityID: ids[0], CommandID: "setup",
		Args: map[string]any{"pond_id": ids[0], "level": 1.0}, AtNS: start,
	}}
	verdict, err := panel.Audit(ids[0], "aerator_failure", start+2*3600e9, start, ids, 12*3600e9, setup,
		[]audit.Perturbation{{Name: "drop", Params: map[string]any{"rate": 0.1}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(verdict.Scores) == 0 || verdict.Best == "" || verdict.Samples == 0 || len(verdict.Channels) == 0 {
		t.Fatalf("verdict = %+v", verdict)
	}
	if verdict.Scores[verdict.Best] != verdict.BestScore {
		t.Fatalf("best %q scores %v, BestScore %v", verdict.Best, verdict.Scores[verdict.Best], verdict.BestScore)
	}
}

func TestAuditRefusesAnUnknownFault(t *testing.T) {
	t.Parallel()
	spec, err := domain.Load(filepath.Join("..", "..", "domains", "aquaculture-pond.domain.json"))
	if err != nil {
		t.Fatal(err)
	}
	panel, err := audit.NewPanel(spec, 1, 60e9)
	if err != nil {
		t.Fatal(err)
	}
	_, err = panel.Audit("site-a/pond-1", "no_such_fault", 0, 0, []string{"site-a/pond-1"}, 3600e9, nil, nil)
	if err == nil || !strings.Contains(err.Error(), `unknown fault "no_such_fault"`) {
		t.Fatalf("err = %v, want an unknown-fault refusal", err)
	}
}

func TestAuditRefusesAMissingPanel(t *testing.T) {
	t.Parallel()
	var nilPanel *audit.Panel
	if _, err := nilPanel.Audit("e", "f", 0, 0, nil, 1, nil, nil); !errors.Is(err, audit.ErrNoPanel) {
		t.Fatalf("nil panel: %v", err)
	}
	var zero audit.Panel
	if _, err := zero.Audit("e", "f", 0, 0, nil, 1, nil, nil); !errors.Is(err, audit.ErrNoPanel) {
		t.Fatalf("zero panel: %v", err)
	}
}
