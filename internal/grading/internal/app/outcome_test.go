package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/grading/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func outcomeOptions(dir string) OutcomeOptions {
	return OutcomeOptions{RunDir: dir, DomainsDir: testsupport.DomainsDir(), AdaptersDir: testsupport.AdaptersDir()}
}

func weakPlantRun(t *testing.T, dir string) {
	t.Helper()
	spec, err := domain.Load(testsupport.BeltConveyorDomain("b"))
	if err != nil {
		t.Fatal(err)
	}
	native, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	r, err := run.New(t.Context(), run.Config{Domain: spec, Adapter: native, Seed: 1, SinkName: model.SinkInproc, TimeMode: model.TimeStepped,
		StartTimeNS: model.DefaultStartTimeNS, StartTimeSet: true})
	if err != nil {
		t.Fatal(err)
	}
	at := func(s int64) int64 { return model.DefaultStartTimeNS + s*1e9 }
	if _, err := r.InjectFault("D-3", "conveyor_drive__vfd_fault", at(60), nil); err != nil {
		t.Fatal(err)
	}
	if _, err := r.InvokeEffector("bulk_feeder__set_feed", "F-1", "c-1", map[string]any{"percent": 0.0}, at(70)); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(t.Context(), at(600), false); err != nil {
		t.Fatal(err)
	}
	if _, err := r.End(dir); err != nil {
		t.Fatal(err)
	}
}

func TestGradeOutcomeScoresTheRunAgainstDoingNothingAndEveryPolicyInOrder(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	weakPlantRun(t, dir)
	graded, err := GradeOutcome(t.Context(), outcomeOptions(dir))
	if err != nil {
		t.Fatal(err)
	}
	if len(graded.Runs) != 3 || graded.Runs[0].Mode != model.OutcomeAsRun || graded.Runs[1].Mode != model.OutcomeDoNothing ||
		graded.Runs[2].Mode != model.PolicyOutcome("b4") {
		t.Fatalf("runs %+v, want as_run, do_nothing and b4 in that order", graded.Runs)
	}
	if !graded.Valid || len(graded.Verdicts) != 2 || graded.Verdicts[0].Verdict != model.VerdictBeats || graded.Domain != "belt-conveyor-drives-b" {
		t.Fatalf("graded %+v, want a valid story the early feed cut beats", graded.OutcomeScore)
	}
}

func TestGradingAnOutcomeTwiceGivesTheSameDocumentByteForByte(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	weakPlantRun(t, dir)
	var documents [2][]byte
	for i := range documents {
		graded, err := GradeOutcome(t.Context(), outcomeOptions(dir))
		if err != nil {
			t.Fatal(err)
		}
		if documents[i], err = json.Marshal(graded); err != nil {
			t.Fatal(err)
		}
	}
	if !bytes.Equal(documents[0], documents[1]) {
		t.Fatalf("two gradings differ:\n%s\n%s", documents[0], documents[1])
	}
}

func TestGradeOutcomeNeedsNoVerdictOrLabelAndRefusesWhatGradingRefuses(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	closedLoopRun(t, dir)
	graded, err := GradeOutcome(t.Context(), outcomeOptions(dir))
	if err != nil || len(graded.Runs) != 2 || len(graded.Runs[0].Facts.Calls) != 2 || graded.Verdicts != nil {
		t.Fatalf("GradeOutcome = %+v, %v; want facts only for a domain without weights", graded, err)
	}
	tamperTraceDigest(t, filepath.Join(dir, "run.json"))
	var mismatch *rules.ReplayMismatchError
	if _, err := GradeOutcome(t.Context(), outcomeOptions(dir)); !errors.As(err, &mismatch) {
		t.Fatalf("a run that does not reproduce: err = %v", err)
	}
	stampUnblinded(t, filepath.Join(dir, "run.json"))
	if _, err := GradeOutcome(t.Context(), outcomeOptions(dir)); !errors.Is(err, rules.ErrRunUnblinded) {
		t.Fatalf("an unblinded run: err = %v", err)
	}
}

func stampUnblinded(t *testing.T, artifact string) {
	t.Helper()
	raw, err := os.ReadFile(artifact)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	doc["unblinded"], doc["mode"] = true, "explore"
	stamped, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(artifact, stamped, 0o600); err != nil {
		t.Fatal(err)
	}
}
