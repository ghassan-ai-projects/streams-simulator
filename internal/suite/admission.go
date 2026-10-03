package suite

import (
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/audit"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (g *suiteGeneration) attemptScenario() error {
	g.suite.Attempts++
	// Rejected attempts retain distinct identities and seeds.
	attempt := g.attemptIndex
	g.attemptIndex++
	seed := g.cfg.Seed + attempt*0x9e3779b97f4a7c15
	sc, label, verdict, err := g.buildScenario(seed, attempt, g.negativeSamplingFraction())
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	g.admitCandidate(sc, label, verdict)
	return nil
}

func (g *suiteGeneration) countComposition(sc *Scenario, label *model.GroundTruthRecord) {
	if label.IsNegativeClass {
		g.negatives++
	}
	if sc.PreDegraded {
		g.preDegraded++
	}
	if sc.Profile == "correlated_cascade" {
		g.cascadeCount++
	}
	if sc.Profile == "sensor_pathology" {
		g.pathologyCount++
	}
}

func (g *suiteGeneration) compositionShortfalls(shortfalls []string) []string {
	count := len(g.suite.Scenarios)
	share := float64(g.negatives) / float64(count)
	if count > 0 && (share < 0.35 || share > 0.45) {
		shortfalls = append(shortfalls, fmt.Sprintf("negative-class fraction %.2f outside [0.35, 0.45]", share))
	}
	if count > 0 && g.preDegraded*100/count < 10 {
		shortfalls = append(shortfalls, "pre-degraded fraction below 10%")
	}
	return shortfalls
}

func (g *suiteGeneration) profileShortfalls(shortfalls []string) []string {
	if g.cascadeCount < 3 {
		shortfalls = append(shortfalls, fmt.Sprintf("correlated_cascade scenarios: %d < 3", g.cascadeCount))
	}
	if g.pathologyCount < 10 {
		shortfalls = append(shortfalls, fmt.Sprintf("sensor_pathology scenarios: %d < 10", g.pathologyCount))
	}
	return shortfalls
}

func (g *suiteGeneration) admitCandidate(sc *Scenario, label *model.GroundTruthRecord, verdict *audit.Verdict) {
	if verdict != nil && verdict.Trivial {
		g.suite.TrivialExcluded = append(g.suite.TrivialExcluded, TrivialCase{ScenarioID: sc.ID, Fault: sc.Fault, Scores: verdict.Scores, Best: verdict.Best})
		return
	}
	g.admitScenario(sc, label)
}
