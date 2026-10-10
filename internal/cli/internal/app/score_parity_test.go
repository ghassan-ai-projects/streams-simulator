package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
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
	gt := &model.GroundTruthRecord{ScenarioID: "parity/1", Domain: "aquaculture-pond", Seed: 5, EntityID: pond,
		Label: "aerator_failure", ExpectedEpisode: true, ExpectedEffector: "start_aerator",
		InjectionTimeNS: faultAt, FirstObservableTimeNS: faultAt, UnavoidableTimeNS: start + 3*3600*1e9,
		TrivialBaselineVerdict: model.TrivialNonTrivial}
	return r, gt
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
	label, err := json.Marshal(gt)
	if err != nil {
		t.Fatal(err)
	}
	labelPath := filepath.Join(dir, "label.json")
	if err := os.WriteFile(labelPath, label, 0o600); err != nil {
		t.Fatal(err)
	}

	out := invoke(t, "score", "--run", filepath.Join(dir, "run.json"), "--label", labelPath,
		"--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir()).mustSucceed(t)
	var offline, want map[string]any
	if err := json.Unmarshal([]byte(out.stdout), &offline); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(online)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(offline, want) {
		t.Fatalf("offline scorecard differs from the online one\noffline: %v\nonline:  %v", offline, want)
	}
	loop, _ := offline["loop"].(map[string]any)
	if loop["effect_calls"] == float64(0) || loop["resolved"] != true || offline["reproducible"] != true {
		t.Fatalf("the offline path must grade the closed loop: %v", offline)
	}
}
