package app

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/grading/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/score"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

// closedLoopRun executes a faulted aquaculture run with a consumer actuation,
// ends it into dir and returns the finished run and its sealed label.
func closedLoopRun(t *testing.T, dir string) (*run.Run, *model.GroundTruthRecord) {
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
		TimeMode: model.TimeStepped, StartTimeNS: start, StartTimeSet: true})
	if err != nil {
		t.Fatal(err)
	}
	const pond = "site-a/pond-1"
	faultAt := start + 2*3600*1e9
	invocation := map[string]any{"pond_id": pond, "level": 1.0}
	if _, err := r.InvokeEffector("start_aerator", pond, "setup", invocation, start); err != nil {
		t.Fatal(err)
	}
	if _, err := r.InjectFault(pond, "aerator_failure", faultAt, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := r.InvokeEffector("start_aerator", pond, "act", invocation, faultAt+600e9); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(t.Context(), start+5*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	verdict := &model.Verdict{SchemaVersion: "0.1", RunID: r.ID, Consumer: model.ConsumerInfo{Name: "parity", Version: "1"},
		Actions: []model.Action{{CommandID: "act", Effector: "start_aerator", EntityID: pond,
			IssuedAt: model.FormatTime(faultAt + 600e9), OutcomeBelieved: model.BelievedSucceeded}}}
	if err := r.SubmitVerdict(verdict); err != nil {
		t.Fatal(err)
	}
	if _, err := r.End(dir); err != nil {
		t.Fatal(err)
	}
	gt := &model.GroundTruthRecord{ScenarioID: "parity/0001", Domain: "aquaculture-pond", Seed: 5, EntityID: pond,
		Label: "aerator_failure", ExpectedEpisode: true, ExpectedEffector: "start_aerator",
		InjectionTimeNS: faultAt, FirstObservableTimeNS: faultAt, UnavoidableTimeNS: start + 3*3600*1e9,
		Observability: model.ObservabilityInfo{DetectorForm: "single_channel_snr", Channels: []string{"pond.dissolved_oxygen"},
			EffectiveSigma: 0.1, FirstObservableSNR: 3, UnavoidableSNR: 10},
		TrivialBaselineVerdict: model.TrivialNonTrivial}
	return r, gt
}

// gradeOptions grade the artifact published in dir against the label.
func gradeOptions(dir, label string) Options {
	return Options{Artifact: filepath.Join(dir, "run.json"), Label: label,
		DomainsDir: testsupport.DomainsDir(), AdaptersDir: testsupport.AdaptersDir()}
}

func writeLabel(t *testing.T, dir string, gt *model.GroundTruthRecord) string {
	t.Helper()
	label, err := json.Marshal(gt)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "label.json")
	if err := os.WriteFile(path, label, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestOfflineScoreEqualsTheOnlineScoreOfTheSameRun(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	r, gt := closedLoopRun(t, dir)
	online, err := score.Score(score.Evidence{
		RunID: r.ID, Domain: r.Domain(), Verdict: r.Verdict(), Ledger: r.Ledger(), Calls: r.World.EffectorCalls(),
		Perturbations: r.AppliedPerturbations(), Emitted: r.World.EmittedCount(), History: r.History(),
		Reproducible: r.Reproducible(), Unblinded: r.UnblindedStamp(),
	}, gt)
	if err != nil {
		t.Fatal(err)
	}
	offline, err := Grade(t.Context(), gradeOptions(dir, writeLabel(t, dir, gt)))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(offline, online) {
		t.Fatalf("offline scorecard differs from the online one\noffline: %+v\nonline:  %+v", offline, online)
	}
	if offline.Loop.EffectCalls == 0 || !offline.Loop.Resolved || !offline.Reproducible {
		t.Fatalf("the offline path must grade the closed loop: %+v", offline)
	}
}

func TestGradeRefusesAnArtifactWhoseReplayDoesNotReproduce(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, gt := closedLoopRun(t, dir)
	label := writeLabel(t, dir, gt)
	tamperTraceDigest(t, filepath.Join(dir, "run.json"))
	_, err := Grade(t.Context(), gradeOptions(dir, label))
	var mismatch *rules.ReplayMismatchError
	if !errors.As(err, &mismatch) || mismatch.Result.Matches {
		t.Fatalf("err = %v, want a ReplayMismatchError carrying the replay report", err)
	}
}

// tamperTraceDigest replaces the artifact's expected trace digest so a replay
// can no longer reproduce it.
func tamperTraceDigest(t *testing.T, artifact string) {
	t.Helper()
	raw, err := os.ReadFile(artifact)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	doc["expected_trace_digest"] = "sha256:" + strings.Repeat("0", 64)
	tampered, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(artifact, tampered, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestGradeChecksItsInputsAgainstTheirContractsAndNamesTheFileAtFault(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, gt := closedLoopRun(t, dir)
	label := writeLabel(t, dir, gt)
	if _, err := Grade(t.Context(), gradeOptions(dir, filepath.Join(dir, "absent.json"))); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("a missing label: err = %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "verdict.json"), []byte(`{}`), 0o600); err != nil {
		t.Fatal(err)
	}
	var invalid *model.ValidationError
	if _, err := Grade(t.Context(), gradeOptions(dir, label)); !errors.As(err, &invalid) || invalid.Source != filepath.Join(dir, "verdict.json") {
		t.Fatalf("a verdict that breaks its contract: err = %v", err)
	}
}

func TestGradeNamesAMissingArtifactFirst(t *testing.T) {
	t.Parallel()
	artifact := filepath.Join(t.TempDir(), "run.json")
	var pathFailure *os.PathError
	_, err := Grade(t.Context(), Options{Artifact: artifact})
	if !errors.As(err, &pathFailure) || pathFailure.Path != artifact {
		t.Fatalf("err = %v, want a path failure naming %s", err, artifact)
	}
}

func TestLoadLedgerAndHistoryNameAMissingFileOrADirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	missing := filepath.Join(dir, "absent.jsonl")
	if _, err := loadLedger(missing); err == nil || !strings.HasPrefix(err.Error(), "open ledger "+missing+": ") {
		t.Fatalf("a missing ledger: err = %v", err)
	}
	if _, err := loadLedger(dir); err == nil || !strings.HasPrefix(err.Error(), "open ledger "+dir+": ") || !strings.Contains(err.Error(), "is a directory") {
		t.Fatalf("a ledger that is a directory: err = %v", err)
	}
	if _, err := loadHistory(missing); err == nil || !strings.HasPrefix(err.Error(), "open state history "+missing+": ") {
		t.Fatalf("a missing history: err = %v", err)
	}
}
