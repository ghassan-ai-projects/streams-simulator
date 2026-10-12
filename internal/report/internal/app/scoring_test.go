package app

import (
	"bytes"
	"encoding/json"
	"os"
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

const pond = "site-a/pond-1"

func finishedRun(t *testing.T, dir string, explore bool) {
	t.Helper()
	spec, err := domain.Load(testsupport.Domain("aquaculture-pond"))
	if err != nil {
		t.Fatal(err)
	}
	native, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	start := model.DefaultStartTimeNS + 4*3600*1e9
	r, err := run.New(t.Context(), run.Config{Domain: spec, Adapter: native, Seed: 5, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start, StartTimeSet: true, Explore: explore, JournalDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	faultAt := start + 2*3600*1e9
	if _, err := r.InjectFault(pond, "aerator_failure", faultAt, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(t.Context(), start+3*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	verdict := &model.Verdict{SchemaVersion: "0.1", RunID: r.ID, Consumer: model.ConsumerInfo{Name: "scoring-test", Version: "1"}}
	if err := r.SubmitVerdict(verdict); err != nil {
		t.Fatal(err)
	}
	if _, err := r.End(dir); err != nil {
		t.Fatal(err)
	}
	writeLabel(t, dir, faultAt)
}

func writeLabel(t *testing.T, dir string, faultAt int64) {
	t.Helper()
	label, err := json.Marshal(&model.GroundTruthRecord{ScenarioID: "report/0001", Domain: "aquaculture-pond", Seed: 5, EntityID: pond,
		Label: "aerator_failure", ExpectedEpisode: true, InjectionTimeNS: faultAt, FirstObservableTimeNS: faultAt,
		UnavoidableTimeNS: faultAt + 3600e9,
		Observability: model.ObservabilityInfo{DetectorForm: "single_channel_snr", Channels: []string{"pond.dissolved_oxygen"},
			EffectiveSigma: 0.1, FirstObservableSNR: 3, UnavoidableSNR: 10},
		TrivialBaselineVerdict: model.TrivialNonTrivial})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "label.json"), label, 0o600); err != nil {
		t.Fatal(err)
	}
}

func scoringOf(t *testing.T, dir string) *rules.Scoring {
	t.Helper()
	built, err := Build(t.Context(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if built.Scoring == nil {
		t.Fatal("report has no scoring section")
	}
	return built.Scoring
}

func TestReportScoresARunWithAVerdictAndALabelFromItsEmbeddedSpecs(t *testing.T) {
	t.Parallel()
	testsupport.SkipUnderRace(t, "runs, replays and grades one world on one goroutine")
	dir := t.TempDir()
	finishedRun(t, dir, false)
	scoring := scoringOf(t, dir)
	var card struct {
		RunID string `json:"run_id"`
	}
	if err := json.Unmarshal(scoring.Scorecard, &card); err != nil || scoring.Status != rules.Scored || card.RunID == "" {
		t.Fatalf("scoring = %+v (%v), want a scorecard for the run", scoring, err)
	}
}

func TestReportShowsAnExploreRunUnblindedAndNotScored(t *testing.T) {
	t.Parallel()
	testsupport.SkipUnderRace(t, "runs one world on one goroutine")
	dir := t.TempDir()
	finishedRun(t, dir, true)
	built, err := Build(t.Context(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if built.Run.Mode != "explore" || !built.Run.Unblinded || built.Truth == nil || built.Scoring.Status != rules.NotScored {
		t.Fatalf("explore report = %+v scoring %+v, want explore, unblinded, truth shown, not scored", built.Run, built.Scoring)
	}
}

func TestReportShowsAScoringFailureWhenTheReplayDoesNotReproduce(t *testing.T) {
	t.Parallel()
	testsupport.SkipUnderRace(t, "runs, replays and grades one world on one goroutine")
	dir := t.TempDir()
	finishedRun(t, dir, false)
	path := filepath.Join(dir, "run.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var artifact map[string]any
	if err := json.Unmarshal(raw, &artifact); err != nil {
		t.Fatal(err)
	}
	artifact["expected_trace_digest"] = "sha256:" + strings.Repeat("0", 64)
	edited, _ := json.Marshal(artifact)
	if err := os.WriteFile(path, edited, 0o600); err != nil {
		t.Fatal(err)
	}
	if scoring := scoringOf(t, dir); scoring.Status != rules.Failed || scoring.Reason == "" {
		t.Fatalf("scoring = %+v, want failed with a reason", scoring)
	}
}

func TestReportOfAnExploreRunKeepsItsCommandsWhenTheArtifactIsMissing(t *testing.T) {
	t.Parallel()
	testsupport.SkipUnderRace(t, "runs one world on one goroutine")
	dir := t.TempDir()
	finishedRun(t, dir, true)
	if err := os.Remove(filepath.Join(dir, "run.json")); err != nil {
		t.Fatal(err)
	}
	built, err := Build(t.Context(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if built.Run.Ended || built.Truth == nil || len(built.Truth.Faults) != 1 || built.Run.Seed == nil || *built.Run.Seed != 5 {
		t.Fatalf("report = %+v truth %+v, want an open explore run with its one fault and seed 5", built.Run, built.Truth)
	}
}

func TestReportOfAScoredRunIsTheSameBytesInACopy(t *testing.T) {
	t.Parallel()
	testsupport.SkipUnderRace(t, "runs, replays and grades one world on one goroutine")
	dir := t.TempDir()
	finishedRun(t, dir, false)
	if first, second := render(t, dir, JSONFormat), render(t, copyFolder(t, dir), JSONFormat); !bytes.Equal(first, second) {
		t.Fatal("the scored report differs between a folder and its copy")
	}
}
