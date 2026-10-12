package grading_test

import (
	"errors"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/grading"
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
