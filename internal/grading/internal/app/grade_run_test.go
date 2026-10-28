package app

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/grading/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func writeLabels(t *testing.T, dir, name string, labels ...*model.GroundTruthRecord) string {
	t.Helper()
	raw, err := json.Marshal(append([]*model.GroundTruthRecord{}, labels...))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func labelOnAnother(gt *model.GroundTruthRecord, entity string) *model.GroundTruthRecord {
	other := *gt
	other.EntityID, other.ScenarioID = entity, "parity/0002"
	return &other
}

func TestARunWithTwoLabelsGetsOneScorecardEachJudgedOnItsOwnEntity(t *testing.T) {
	t.Parallel()
	testsupport.SkipUnderRace(t, "runs, replays and grades one world on one goroutine")
	dir := t.TempDir()
	_, gt := closedLoopRun(t, dir)
	writeLabels(t, dir, "labels.json", gt, labelOnAnother(gt, "site-a/pond-2"))
	graded, err := GradeRun(t.Context(), gradeOptions(dir, ""))
	if err != nil || len(graded.Scorecards) != 2 {
		t.Fatalf("GradeRun = %+v, %v, want two scorecards", graded, err)
	}
	treated, untreated := graded.Scorecards[0], graded.Scorecards[1]
	if treated.GroundTruth.EntityID != "site-a/pond-1" || !treated.Loop.ActionAppropriate || treated.Loop.EffectCalls == 0 {
		t.Fatalf("the treated pond = %+v", treated.Loop)
	}
	if untreated.GroundTruth.EntityID != "site-a/pond-2" || untreated.Loop.ActionAppropriate || untreated.Loop.EffectCalls != 0 {
		t.Fatalf("the untreated pond = %+v, want no call credited from pond-1", untreated.Loop)
	}
	if unexpected := graded.Unexpected; len(unexpected.Actions) != 0 || unexpected.EffectCalls != 0 {
		t.Fatalf("unexpected = %+v, want nothing: every pond with a call has a label", unexpected)
	}
}

func TestActionsOnAnUnlabelledPondAreUnexpectedAndNoLabelIsAFileOfItsOwn(t *testing.T) {
	t.Parallel()
	testsupport.SkipUnderRace(t, "runs, replays and grades one world on one goroutine")
	dir := t.TempDir()
	_, gt := closedLoopRun(t, dir)
	none := writeLabels(t, dir, "none.json")
	graded, err := GradeRun(t.Context(), gradeOptions(dir, none))
	if err != nil || len(graded.Scorecards) != 0 || len(graded.Unexpected.Actions) != 1 || graded.Unexpected.EffectCalls == 0 {
		t.Fatalf("GradeRun = %+v, %v, want no scorecard and the action on pond-1 unexpected", graded, err)
	}
	several := writeLabels(t, dir, "labels.json", gt, labelOnAnother(gt, "site-a/pond-2"))
	if _, err := Grade(t.Context(), gradeOptions(dir, several)); !errors.Is(err, rules.ErrNotOneLabel) {
		t.Fatalf("Grade of two labels = %v, want %v", err, rules.ErrNotOneLabel)
	}
}

func TestTheLabelsBesideTheRunAreFoundByTheirFileName(t *testing.T) {
	t.Parallel()
	testsupport.SkipUnderRace(t, "runs, replays and grades one world on one goroutine")
	dir := t.TempDir()
	_, gt := closedLoopRun(t, dir)
	if _, err := GradeRun(t.Context(), gradeOptions(dir, "")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("a folder without a label = %v, want a missing file", err)
	}
	writeLabels(t, dir, "labels.json", gt)
	graded, err := GradeRun(t.Context(), gradeOptions(dir, ""))
	if err != nil || len(graded.Scorecards) != 1 {
		t.Fatalf("GradeRun of labels.json = %+v, %v", graded, err)
	}
}
