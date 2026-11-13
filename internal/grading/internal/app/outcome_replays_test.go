package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestGradingAnOutcomeOnOneWorkerOrManyGivesTheSameDocumentByteForByte(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	weakPlantRun(t, dir)
	var documents [2][]byte
	for i, workers := range []int{1, 6} {
		options := outcomeOptions(dir)
		options.Workers = workers
		graded, err := GradeOutcome(t.Context(), options)
		if err != nil {
			t.Fatal(err)
		}
		if documents[i], err = json.Marshal(graded); err != nil {
			t.Fatal(err)
		}
	}
	if !bytes.Equal(documents[0], documents[1]) {
		t.Fatalf("one worker and six differ:\n%s\n%s", documents[0], documents[1])
	}
}

func TestOutcomeReplaysReportTheFirstFailureByModeOrderAndCancelTheLaterOnes(t *testing.T) {
	t.Parallel()
	modes := []model.OutcomeMode{model.OutcomeAsRun, model.OutcomeDoNothing, model.PolicyOutcome("b1"), model.PolicyOutcome("b2")}
	b2Started, b2Canceled := make(chan struct{}), make(chan struct{})
	replay := func(ctx context.Context, mode model.OutcomeMode) (model.OutcomeRun, error) {
		switch mode {
		case model.OutcomeDoNothing:
			<-b2Canceled
			return model.OutcomeRun{}, fmt.Errorf("%s failed", mode)
		case model.PolicyOutcome("b1"):
			<-b2Started
			return model.OutcomeRun{}, fmt.Errorf("%s failed", mode)
		case model.PolicyOutcome("b2"):
			close(b2Started)
			<-ctx.Done()
			close(b2Canceled)
			return model.OutcomeRun{}, ctx.Err()
		}
		return model.OutcomeRun{Mode: mode}, nil
	}
	runs, err := replayOutcomes(t.Context(), modes, len(modes), replay)
	if runs != nil || err == nil || err.Error() != "do_nothing failed" {
		t.Fatalf("replayOutcomes = %v, %v; want do_nothing's failure although b1 failed first and canceled b2", runs, err)
	}
}

func TestOutcomeReplaysKeepModeOrderWhateverOrderTheyFinishIn(t *testing.T) {
	t.Parallel()
	modes := []model.OutcomeMode{model.OutcomeAsRun, model.OutcomeDoNothing, model.PolicyOutcome("b1")}
	release := make(chan struct{})
	replay := func(_ context.Context, mode model.OutcomeMode) (model.OutcomeRun, error) {
		if mode == model.OutcomeAsRun {
			<-release
		} else if mode == model.PolicyOutcome("b1") {
			defer close(release)
		}
		return model.OutcomeRun{Mode: mode}, nil
	}
	runs, err := replayOutcomes(t.Context(), modes, 3, replay)
	if err != nil || len(runs) != 3 || runs[0].Mode != modes[0] || runs[1].Mode != modes[1] || runs[2].Mode != modes[2] {
		t.Fatalf("replayOutcomes = %+v, %v; want the runs in mode order", runs, err)
	}
}

func TestOutcomeWorkersDefaultToAtMostFourAndHonourARequest(t *testing.T) {
	t.Parallel()
	if got, want := outcomeWorkers(0), min(4, runtime.GOMAXPROCS(0)); got != want {
		t.Fatalf("default workers %d, want %d", got, want)
	}
	if got := outcomeWorkers(2); got != 2 {
		t.Fatalf("requested two workers, got %d", got)
	}
	if _, err := replayOutcomes(t.Context(), nil, 0, nil); err != nil {
		t.Fatalf("no modes: %v", err)
	}
}
