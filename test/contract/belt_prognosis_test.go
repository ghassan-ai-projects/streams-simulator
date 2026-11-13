package contract

import (
	"math"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/grading"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/score"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
	"github.com/ghassan-ai-projects/streams-simulator/internal/truth"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func runScenarioInto(t *testing.T, scenario *model.Scenario, dir string) string {
	t.Helper()
	spec, native, err := run.LoadInputs(filepath.Join(testsupport.RepositoryRoot(), "examples"), testsupport.AdaptersDir(), scenario.Domain, scenario.Adapter)
	if err != nil {
		t.Fatal(err)
	}
	r, err := run.New(t.Context(), run.Config{Domain: spec, Adapter: native, Seed: scenario.Seed, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: scenario.World.StartTimeNS, StartTimeSet: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := r.ExecuteScenario(t.Context(), scenario); err != nil {
		t.Fatal(err)
	}
	if _, err := r.End(dir); err != nil {
		t.Fatal(err)
	}
	return r.ID
}

func bearingLabel(t *testing.T, scenario *model.Scenario) *model.GroundTruthRecord {
	t.Helper()
	spec := beltPlant(t, "a")
	solver, err := truth.NewSolver(spec, scenario.Seed, 120*1e9, 0)
	if err != nil {
		t.Fatal(err)
	}
	onset := nanoseconds(scenario.Commands[1].Args["onset_ns"])
	label, err := truth.BuildRecord(spec, solver, truth.Injection{ScenarioID: "belt-conveyor-drives-a/0001", Seed: scenario.Seed, EntityID: "D-3",
		FaultID: "conveyor_drive__lube_loss", OnsetNS: onset, StartNS: model.DefaultStartTimeNS, EntityIDs: world.PopulationIDs(spec, 0)})
	if err != nil {
		t.Fatal(err)
	}
	return label
}

func firstTripNS(entity model.EntityOutcome) int64 {
	first := int64(math.MaxInt64)
	for _, trip := range entity.Trips {
		first = min(first, trip.FirstNS)
	}
	return first
}

func firstInvokeNS(scenario *model.Scenario) int64 {
	for _, command := range scenario.Commands {
		if command.Op == model.OpEffectorInvoke {
			return nanoseconds(command.Args["at_ns"])
		}
	}
	return math.MaxInt64
}

// predictedAt is a prediction made at at whose failure time misses the true one by share times the true remaining life
// (positive: late).
func predictedAt(at, failure int64, share float64) model.Detection {
	predicted := failure + int64(share*float64(failure-at))
	return model.Detection{EntityID: "D-3", DetectedAt: model.FormatTime(at), Label: "conveyor_drive__lube_loss", PredictedFailureAt: model.FormatTime(predicted)}
}

func lambdaPoint(first, failure int64, lambda float64) int64 {
	return first + int64(lambda*float64(failure-first))
}

type s1aCase struct{ first, failure, cut int64 }

// gradedS1A runs S1-A row 1 (the named scenario) into a folder with its label and the verdict the predictions give, and
// grades it. D-3's true failure time comes from a direct world run of the do-nothing scenario, not from the grading path.
func gradedS1A(t *testing.T, name string, predictions func(c s1aCase) []model.Detection) (s1aCase, *score.PrognosisMetrics) {
	t.Helper()
	scenario, dir := storyScenario(t, name), t.TempDir()
	runID := runScenarioInto(t, scenario, dir)
	label := bearingLabel(t, scenario)
	c := s1aCase{first: label.FirstObservableTimeNS, cut: firstInvokeNS(scenario),
		failure: firstTripNS(entityOutcome(t, directOutcome(t, storyScenario(t, "s1-a.do-nothing")), "D-3"))}
	writeRunFile(t, dir, model.FileLabel, label)
	writeRunFile(t, dir, model.FileVerdict, model.Verdict{SchemaVersion: "0.1", RunID: runID, Consumer: model.ConsumerInfo{Name: "hand", Version: "1"},
		Detections: predictions(c)})
	graded, err := grading.GradeRun(t.Context(), grading.Options{Artifact: filepath.Join(dir, model.FileRunArtifact)})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("S1-A row 1 %s: first observable %.0f s, first call %.0f s, D-3 trips at %.0f s", name, secondsOf(c.first), secondsOf(c.cut), secondsOf(c.failure))
	return c, graded.Scorecards[0].Judgment.Prognosis
}

func accuracyIs(m score.LambdaMetric, at int64, inside bool, accuracy float64) bool {
	return m.AtNS == at && m.Graded && m.AlphaLambda == inside && m.RelativeAccuracy != nil && math.Abs(*m.RelativeAccuracy-accuracy) < 1e-9
}

// Nobody acts, so every prediction before the trip is graded: at the first observable time 15 % late (weighted 0.3,
// outside α 0.2), at the λ 0.5 point 10 % early (0.1, inside, RA 0.9), at the λ 0.75 point 5 % late (weighted 0.1, inside,
// RA 0.9), and one after the trip, which is not graded. PH runs from the λ 0.5 prediction.
func TestPredictionsOfAConveyorBearingFailureAreGradedAgainstTheDoNothingTrip(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("16-hour runs of plant A and the do-nothing replay")
	}
	c, got := gradedS1A(t, "s1-a.do-nothing", func(c s1aCase) []model.Detection {
		return []model.Detection{predictedAt(c.first, c.failure, 0.15), predictedAt(lambdaPoint(c.first, c.failure, 0.5), c.failure, -0.1),
			predictedAt(lambdaPoint(c.first, c.failure, 0.75), c.failure, 0.05), predictedAt(c.failure+60e9, c.failure, 0)}
	})
	half, late := lambdaPoint(c.first, c.failure, 0.5), lambdaPoint(c.first, c.failure, 0.75)
	if got == nil || got.FailureNS != c.failure || got.GradedUntilNS != c.failure || got.Predictions != 3 || got.HorizonNS != c.failure-half {
		t.Fatalf("prognosis %+v, want failure and window end at %.0f s, 3 predictions, PH %.0f s", got, secondsOf(c.failure), float64(c.failure-half)/1e9)
	}
	if !accuracyIs(got.Lambdas[0], half, true, 0.9) || !accuracyIs(got.Lambdas[1], late, true, 0.9) {
		t.Fatalf("λ points %+v, want both inside with RA 0.9 at %.0f s and %.0f s", got.Lambdas, secondsOf(half), secondsOf(late))
	}
}

// The reference policy cuts the feed 4 h after the onset: predictions after that call are not graded, and both λ points
// fall after it.
func TestTheFirstEffectiveCallEndsTheGradedWindowOnAConveyorStory(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("16-hour runs of plant A and the do-nothing replay")
	}
	c, got := gradedS1A(t, "s1-a.reference", func(c s1aCase) []model.Detection {
		return []model.Detection{predictedAt(c.first, c.failure, 0.15), predictedAt(c.cut-600e9, c.failure, -0.1), predictedAt(c.cut+60e9, c.failure, 0)}
	})
	if got == nil || got.FailureNS != c.failure || got.GradedUntilNS != c.cut || got.Predictions != 2 || got.HorizonNS != c.failure-(c.cut-600e9) {
		t.Fatalf("prognosis %+v, want failure at %.0f s, window end at the feed cut %.0f s, 2 predictions, PH from 600 s before the cut", got,
			secondsOf(c.failure), secondsOf(c.cut))
	}
	if got.Lambdas[0].Graded || got.Lambdas[1].Graded {
		t.Fatalf("λ points %+v, want both after the feed cut and ungraded", got.Lambdas)
	}
}
