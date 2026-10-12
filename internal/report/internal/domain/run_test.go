package domain

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func endedArtifact(mutate func(*model.RunArtifact)) *model.RunArtifact {
	art := &model.RunArtifact{
		RunID: "r-7", SimVersion: "9", Seed: 4, Sink: "inproc", TimeMode: "stepped", Mode: "evaluate",
		Domain: model.ArtifactRef{ID: "d", Version: "1", Digest: "sha256:aa"}, Adapter: model.ArtifactRef{ID: "n", Version: "1", Digest: "sha256:bb"},
		WorldConfig: model.WorldConfig{StartTimeNS: 1000, EntityIDs: []string{"a"}},
		CommandLog: []model.Command{
			{Seq: 0, Op: model.OpClockAdvance, Args: map[string]any{"to_ns": json.Number("4000")}},
			{Seq: 1, Op: model.OpClockAdvance, Args: map[string]any{"to_ns": json.Number("9000")}},
		},
		ExpectedTraceDigest: "sha256:cc", WorldDigest: "sha256:dd", Counts: model.Counts{Emitted: 6},
	}
	mutate(art)
	return art
}

func TestAnEndedRunIsCompleteEndedAndEndsAtItsLastAdvance(t *testing.T) {
	t.Parallel()
	report, err := Assemble(Folder{Artifact: endedArtifact(func(*model.RunArtifact) {})}, nil)
	if err != nil {
		t.Fatal(err)
	}
	run := report.Run
	if !run.Complete || !run.Ended || run.Unblinded || run.EndTimeNS != 9000 || *run.Seed != 4 || report.Truth == nil || *report.Counts.Emitted != 6 {
		t.Fatalf("run = %+v truth %+v, want a complete ended run ending at 9000 with seed 4", run, report.Truth)
	}
}

func TestAnIncompleteArtifactRepeatsItsErrorOrSaysItEndedIncomplete(t *testing.T) {
	t.Parallel()
	cases := map[string]string{"sink failed": "sink failed", "": "the run ended incomplete"}
	for failure, want := range cases {
		report, err := Assemble(Folder{Artifact: endedArtifact(func(a *model.RunArtifact) { a.Incomplete, a.Error = true, failure })}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if report.Run.Complete || report.Run.IncompleteReason != want {
			t.Errorf("run = %+v, want incomplete with reason %q", report.Run, want)
		}
	}
}

func TestAnExploreArtifactIsUnblinded(t *testing.T) {
	t.Parallel()
	report, err := Assemble(Folder{Artifact: endedArtifact(func(a *model.RunArtifact) { a.Mode = "explore" })}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if report.Run.Mode != "explore" || !report.Run.Unblinded {
		t.Fatalf("run = %+v, want explore and unblinded", report.Run)
	}
}

func TestMarkdownRendersEveryRecordedSectionOfAnEndedRun(t *testing.T) {
	t.Parallel()
	art := endedArtifact(func(a *model.RunArtifact) {
		a.CommandLog = append(a.CommandLog,
			model.Command{Seq: 2, AtNS: 1500, Op: model.OpPerturbApply, Args: map[string]any{"perturbation": "drop", "from_ns": json.Number("2000")}},
			model.Command{Seq: 3, AtNS: 1500, Op: model.OpEnvInject, Args: map[string]any{"target": "t", "fault": "kill"}},
		)
	})
	ledger := doc(`{"delivery_id":1,"delivered":true,"delivery_reason":"ok"}`, `{"delivery_id":2,"delivered":false,"delivery_reason":"dropped_by_perturbation"}`)
	report, err := Assemble(Folder{Artifact: art, Ledger: ledger, Trace: Document{Raw: []byte("x"), Present: true}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	got := string(Markdown(report))
	for _, want := range []string{
		"| Domain | d 1 `sha256:aa` |", "| World | `sha256:dd` |", "Status: files_differ",
		"| trace | mismatch | `sha256:cc` | `sha256:", "### Delivered by reason", "### Dropped by reason",
		"| 2 | 1970-01-01T00:00:00.0000015Z (+0.0000005s) | apply | drop | 1970-01-01T00:00:00.000002Z (+0.000001s) |  |  |",
		"| 3 | 1970-01-01T00:00:00.0000015Z (+0.0000005s) | t | kill |  |",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("markdown lacks %q:\n%s", want, got)
		}
	}
}

func TestMarkdownSpellsOutAnAcknowledgedEffectorOutcome(t *testing.T) {
	t.Parallel()
	accepted, applied, latency := true, false, 12.5
	got := outcomeText(&Outcome{Mode: "slow", Accepted: &accepted, EffectApplied: &applied, AckLatencyMS: &latency})
	if got != "mode=slow accepted=yes effect_applied=no ack_latency_ms=12.5" || outcomeText(nil) != "" {
		t.Fatalf("outcome text = %q", got)
	}
}
