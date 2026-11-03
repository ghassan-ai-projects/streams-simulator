package grading_test

import (
	"errors"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/grading"
	"github.com/ghassan-ai-projects/streams-simulator/internal/score"
)

func TestGradeNamesAnArtifactItCannotRead(t *testing.T) {
	t.Parallel()
	artifact := filepath.Join(t.TempDir(), "run.json")
	_, err := grading.Grade(t.Context(), grading.Options{Artifact: artifact})
	var pathFailure *fs.PathError
	var mismatch *grading.ReplayMismatchError
	if !errors.As(err, &pathFailure) || pathFailure.Path != artifact || errors.As(err, &mismatch) {
		t.Fatalf("err = %v, want a path failure naming %s that is not a replay mismatch", err, artifact)
	}
}

func TestGradeRunNamesAnArtifactItCannotRead(t *testing.T) {
	t.Parallel()
	artifact := filepath.Join(t.TempDir(), "run.json")
	_, err := grading.GradeRun(t.Context(), grading.Options{Artifact: artifact})
	var pathFailure *fs.PathError
	if !errors.As(err, &pathFailure) || pathFailure.Path != artifact {
		t.Fatalf("err = %v, want a path failure naming %s", err, artifact)
	}
}

func TestDocumentIsTheCardForOneScorecardAndTheRunOtherwise(t *testing.T) {
	t.Parallel()
	card := &score.Scorecard{RunID: "r-1"}
	if got := grading.Document(&score.RunScore{RunID: "r-1", Scorecards: []*score.Scorecard{card}}); got != card {
		t.Fatalf("one scorecard: document = %v", got)
	}
	if _, ok := grading.Document(&score.RunScore{RunID: "r-1"}).(*score.RunScore); !ok {
		t.Fatal("no scorecard: document is not the run score")
	}
}
