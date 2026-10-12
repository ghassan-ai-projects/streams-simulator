package domain

import (
	"errors"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func TestRequireReproducedRefusesAReplayThatDiverged(t *testing.T) {
	t.Parallel()
	if err := RequireReproduced(&run.ReplayEvidence{Result: &run.ReplayResult{Matches: true}}); err != nil {
		t.Fatalf("a reproduced replay must pass: %v", err)
	}
	err := RequireReproduced(&run.ReplayEvidence{Result: &run.ReplayResult{Matches: false, Detail: "trace digest differs"}})
	var mismatch *ReplayMismatchError
	if !errors.As(err, &mismatch) || mismatch.Result.Detail != "trace digest differs" {
		t.Fatalf("err = %v, want a ReplayMismatchError carrying the replay report", err)
	}
	want := "score: replay does not reproduce the artifact, so its effector calls cannot be graded: trace digest differs"
	if err.Error() != want {
		t.Fatalf("message = %q, want %q", err.Error(), want)
	}
}

func TestArtifactEvidenceCarriesTheStampsAndCountsOfTheArtifact(t *testing.T) {
	t.Parallel()
	art := &model.RunArtifact{RunID: "r-1", Reproducible: true, Unblinded: true, AppliedPerturbations: []string{"drop"}}
	art.Counts.Emitted = 7
	evidence := ArtifactEvidence(art, nil, nil)
	if evidence.RunID != "r-1" || evidence.Emitted != 7 || !evidence.Reproducible || !evidence.Unblinded || len(evidence.Perturbations) != 1 {
		t.Fatalf("evidence = %+v", evidence)
	}
}
