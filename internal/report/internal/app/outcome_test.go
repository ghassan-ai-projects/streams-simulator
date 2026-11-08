package app

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/report/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func plantRun(t *testing.T, end bool) string {
	t.Helper()
	spec, err := domain.Load(testsupport.BeltConveyorDomain("b"))
	if err != nil {
		t.Fatal(err)
	}
	native, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	r, err := run.New(t.Context(), run.Config{Domain: spec, Adapter: native, Seed: 1, SinkName: model.SinkInproc, TimeMode: model.TimeStepped,
		StartTimeNS: model.DefaultStartTimeNS, StartTimeSet: true, JournalDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.InjectFault("D-3", "conveyor_drive__vfd_fault", model.DefaultStartTimeNS+30e9, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(t.Context(), model.DefaultStartTimeNS+300e9, false); err != nil {
		t.Fatal(err)
	}
	if end {
		if _, err := r.End(dir); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func editArtifact(t *testing.T, dir string, edit func(doc map[string]any)) {
	t.Helper()
	rewrite(t, dir, model.FileRunArtifact, func(raw []byte) []byte {
		var doc map[string]any
		if err := json.Unmarshal(raw, &doc); err != nil {
			t.Fatal(err)
		}
		edit(doc)
		edited, err := json.Marshal(doc)
		if err != nil {
			t.Fatal(err)
		}
		return edited
	})
}

func outcomeOf(t *testing.T, dir string, asked bool) *rules.OutcomeScoring {
	t.Helper()
	document, err := Produce(t.Context(), Request{Dir: dir, Outcome: asked})
	if err != nil {
		t.Fatal(err)
	}
	return document.(*rules.Report).OutcomeScore
}

func TestTheOutcomeScoreIsComputedOnlyWhenAskedForAndOnlyForAnEndedRunWithWeights(t *testing.T) {
	t.Parallel()
	ended := plantRun(t, true)
	if got := outcomeOf(t, ended, false); got != nil {
		t.Fatalf("a report not asked for the outcome carries %+v", got)
	}
	if got := outcomeOf(t, ended, true); got == nil || got.Status != rules.Scored || got.Result == nil || len(got.Result.Runs) != 3 {
		t.Fatalf("outcome score %+v, want a scored result with as_run, do_nothing and b4", got)
	}
	if got := outcomeOf(t, plantRun(t, false), true); got != nil {
		t.Fatalf("an open run carries %+v, want no outcome score before it ends", got)
	}
	if got := outcomeOf(t, realRun(t, ""), true); got != nil {
		t.Fatalf("a run of a domain without outcome weights carries %+v", got)
	}
}

func TestAnOutcomeThatCannotBeGradedSaysWhyWithoutTheFolderPath(t *testing.T) {
	t.Parallel()
	unblinded, broken := plantRun(t, true), plantRun(t, true)
	editArtifact(t, unblinded, func(doc map[string]any) { doc["unblinded"] = true })
	editArtifact(t, broken, func(doc map[string]any) { doc["ledger_digest"] = "sha256:not-a-digest" })
	if got := outcomeOf(t, unblinded, true); got == nil || got.Status != rules.NotScored || !strings.Contains(got.Reason, "unblinded") {
		t.Fatalf("an unblinded run: %+v, want not_scored with the reason", got)
	}
	got := outcomeOf(t, broken, true)
	if got == nil || got.Status != rules.Failed || got.Result != nil || !strings.Contains(got.Reason, "/ledger_digest") ||
		strings.Contains(got.Reason, filepath.Clean(broken)) {
		t.Fatalf("a run grading refuses: %+v, want failed, naming the member at fault and not the folder", got)
	}
}
