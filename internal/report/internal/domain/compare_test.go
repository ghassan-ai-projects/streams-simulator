package domain

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const startNS = int64(1_800_000_000_000_000_000)

func seeded(seed uint64) *uint64 { return &seed }

func ns(seconds int64) int64 { return startNS + seconds*1_000_000_000 }

func sealedRun(seed uint64) *Report {
	observed, emitted := int64(5), int64(5)
	return &Report{
		SchemaVersion: SchemaVersion,
		Run: Run{RunID: "r-1", Mode: "explore", Complete: true, Ended: true, Unblinded: true, SimVersion: "9",
			Domain: &model.ArtifactRef{ID: "plant", Version: "1", Digest: "sha256:d1"}, Adapter: &model.ArtifactRef{ID: "native", Version: "1", Digest: "sha256:a1"},
			Seed: seeded(seed), TimeMode: "stepped", StartTimeNS: startNS, EndTimeNS: ns(600), RequestedEndNS: ns(600),
			Termination: &model.Termination{Reason: model.TerminationCompleted, AtNS: ns(600)}},
		Verification: Verification{Status: FilesMatch, Checks: []Check{}},
		Counts:       Counts{Emitted: &emitted, Observed: &observed},
		Truth: &Truth{
			Faults:        []model.FaultInterval{faultInterval("chiller-1", "fan_failure", ns(60), nil)},
			Perturbations: []Perturbation{{Seq: 1, AtNS: ns(0), Action: "apply", Perturbation: "drop", FromNS: pointer(ns(100)), UntilNS: pointer(ns(200)), Params: map[string]any{"rate": 0.5}}},
			EffectorCalls: []EffectorCall{{Seq: 2, AtNS: ns(120), CommandID: "act-1", Effector: "set_setpoint", EntityID: "chiller-1",
				Args: map[string]any{"value": 7.0}, Outcome: &Outcome{Mode: "ack", Accepted: boolean(true), EffectApplied: boolean(true)}}},
			KPIs: &model.KPITruth{FromNS: startNS, ToNS: ns(600), Entities: []model.EntityKPIs{{EntityID: "chiller-1",
				KPIs:   []model.KPIValue{{State: "cop", Mean: 5, Min: 4, Max: 6, Samples: 60}},
				Alarms: []model.AlarmKPI{{State: "high_temp", Priority: "high", Activations: 2, ActiveS: 30, LongestActiveS: 20}}}},
				AlarmSystem: model.AlarmSystemKPIs{Activations: 2, PeakPer10Min: 2}},
		},
	}
}

func pointer(n int64) *int64 { return &n }

func boolean(b bool) *bool { return &b }

func faultInterval(entity, fault string, onset int64, cleared *int64) model.FaultInterval {
	return model.FaultInterval{FaultID: "f-0", EntityID: entity, Fault: fault, InjectedAtNS: startNS, OnsetNS: onset, ClearedNS: cleared}
}

func limitCodes(c *Comparison) []string {
	codes := []string{}
	for _, limit := range c.Limits {
		codes = append(codes, limit.Code+limit.Side)
	}
	return codes
}

func TestComparingARunWithItselfFindsNoDifferenceAndOnlyTheSingleRunLimit(t *testing.T) {
	t.Parallel()
	c := Compare(sealedRun(7), sealedRun(7))
	for name, count := range map[string]int{"faults": len(c.Plan.Faults.Different) + len(c.Plan.Faults.OnlyInA) + len(c.Plan.Faults.OnlyInB),
		"perturbations": len(c.Plan.Perturbations.Different) + len(c.Plan.Perturbations.OnlyInA) + len(c.Plan.Perturbations.OnlyInB),
		"calls":         len(c.Plan.EffectorCalls.Different) + len(c.Plan.EffectorCalls.OnlyInA) + len(c.Plan.EffectorCalls.OnlyInB)} {
		if count != 0 {
			t.Errorf("%s differ for identical runs: %+v", name, c.Plan)
		}
	}
	if len(c.Plan.Faults.Identical) != 1 || len(c.Plan.Perturbations.Identical) != 1 || len(c.Plan.EffectorCalls.Identical) != 1 {
		t.Fatalf("plan = %+v, want each item identical", c.Plan)
	}
	if got := limitCodes(c); !slices.Equal(got, []string{"single_run_per_side"}) {
		t.Fatalf("limits = %v, want only single_run_per_side", got)
	}
	row := c.Results.KPIs[0]
	if row.Presence != PresenceBoth || row.Delta == nil || row.Delta.Mean != 0 || row.Scope.Kind != "whole_run" || *row.Scope.ASimulatedLengthNS != 600e9 {
		t.Fatalf("kpi row = %+v, want both sides, a zero delta and the whole-run scope", row)
	}
}

func TestTheSamePlanWithAnotherSeedIsComparableWithTheSeedNamedAsALimit(t *testing.T) {
	t.Parallel()
	c := Compare(sealedRun(7), sealedRun(8))
	if got := limitCodes(c); !slices.Equal(got, []string{"different_seed", "single_run_per_side"}) {
		t.Fatalf("limits = %v, want different_seed first", got)
	}
	if len(c.Plan.Faults.Identical) != 1 || len(c.Plan.Faults.Different) != 0 {
		t.Fatalf("faults = %+v, want the same fault", c.Plan.Faults)
	}
}

func TestAFaultOnlyTheSecondRunInjectedIsOnlyInB(t *testing.T) {
	t.Parallel()
	b := sealedRun(7)
	b.Truth.Faults = append(b.Truth.Faults, faultInterval("pump-1", "bearing_wear", ns(90), nil))
	c := Compare(sealedRun(7), b)
	if len(c.Plan.Faults.OnlyInB) != 1 || c.Plan.Faults.OnlyInB[0].EntityID != "pump-1" || len(c.Plan.Faults.OnlyInA) != 0 || len(c.Plan.Faults.Identical) != 1 {
		t.Fatalf("faults = %+v, want pump-1 only in B beside the identical first fault", c.Plan.Faults)
	}
}

func TestAFaultClearedEarlierInTheSecondRunIsPairedAndNamesTheFieldThatDiffers(t *testing.T) {
	t.Parallel()
	a, b := sealedRun(7), sealedRun(7)
	a.Truth.Faults[0].ClearedNS = pointer(ns(300))
	b.Truth.Faults[0].ClearedNS = pointer(ns(200))
	c := Compare(a, b)
	if len(c.Plan.Faults.Different) != 1 || !slices.Equal(c.Plan.Faults.Different[0].Differences, []string{"cleared_ns"}) ||
		*c.Plan.Faults.Different[0].A.ClearedNS != ns(300) || *c.Plan.Faults.Different[0].B.ClearedNS != ns(200) {
		t.Fatalf("faults = %+v, want one pair differing in cleared_ns only", c.Plan.Faults)
	}
}

func TestIdenticalItemsPairBeforeNearMatchesSoALaterCallIsNotPairedWithAnEarlierOne(t *testing.T) {
	t.Parallel()
	a, b := sealedRun(7), sealedRun(7)
	early := a.Truth.EffectorCalls[0]
	late := early
	late.Seq, late.AtNS, late.CommandID = 3, ns(300), "act-2"
	a.Truth.EffectorCalls = []EffectorCall{early, late}
	b.Truth.EffectorCalls = []EffectorCall{late}
	b.Truth.EffectorCalls[0].CommandID = "other-id"
	c := Compare(a, b)
	calls := c.Plan.EffectorCalls
	if len(calls.Identical) != 1 || calls.Identical[0].AtNS != ns(300) || len(calls.OnlyInA) != 1 || calls.OnlyInA[0].AtNS != ns(120) || len(calls.Different) != 0 {
		t.Fatalf("calls = %+v, want the t=300 call identical regardless of its command id and the t=120 call only in A", calls)
	}
}

func TestARunStoppedEarlyAndARunOfAnotherLengthAreNamedWithTheirSide(t *testing.T) {
	t.Parallel()
	b := sealedRun(7)
	b.Run.EndTimeNS = ns(300)
	b.Run.Termination = &model.Termination{Reason: model.TerminationStoppedByOperator, AtNS: ns(300)}
	b.Truth.KPIs.ToNS = ns(300)
	c := Compare(sealedRun(7), b)
	if got := limitCodes(c); !slices.Equal(got, []string{"different_simulated_length", "stopped_earlyb", "single_run_per_side"}) {
		t.Fatalf("limits = %v, want different_simulated_length, stopped_early for b and the single run limit", got)
	}
	row := c.Results.KPIs[0]
	if *row.Scope.ASimulatedLengthNS != 600e9 || *row.Scope.BSimulatedLengthNS != 300e9 {
		t.Fatalf("scope = %+v, want each run's own simulated length", row.Scope)
	}
}

func TestADifferentDomainDigestAndAnUnverifiedIncompleteRunAreLimits(t *testing.T) {
	t.Parallel()
	b := sealedRun(7)
	b.Run.Domain = &model.ArtifactRef{ID: "plant", Version: "2", Digest: "sha256:d2"}
	b.Run.Complete, b.Run.IncompleteReason = false, "the sink failed"
	b.Verification.Status = FilesDiffer
	b.Run.Mode, b.Run.SimVersion, b.Run.TimeMode = "evaluate", "10", "realtime"
	c := Compare(sealedRun(7), b)
	want := []string{"different_mode", "different_simulator_version", "different_domain_digest", "different_time_mode", "incompleteb", "files_unverifiedb", "single_run_per_side"}
	if got := limitCodes(c); !slices.Equal(got, want) {
		t.Fatalf("limits = %v, want %v", got, want)
	}
}

func TestADifferentDomainIdAndRequestedDurationAreLimits(t *testing.T) {
	t.Parallel()
	b := sealedRun(7)
	b.Run.Domain = &model.ArtifactRef{ID: "other", Version: "1", Digest: "sha256:d1"}
	b.Run.RequestedEndNS = ns(900)
	got := limitCodes(Compare(sealedRun(7), b))
	if !slices.Contains(got, "different_domain") || !slices.Contains(got, "different_requested_duration") || slices.Contains(got, "different_domain_digest") {
		t.Fatalf("limits = %v, want different_domain and different_requested_duration", got)
	}
}

func TestAKPIOnlyOneRunHasIsPresentOnlyOnThatSideAndNamedAsALimit(t *testing.T) {
	t.Parallel()
	a, b := sealedRun(7), sealedRun(7)
	a.Truth.KPIs.Entities[0].KPIs = append(a.Truth.KPIs.Entities[0].KPIs, model.KPIValue{State: "power", Mean: 10, Min: 9, Max: 11, Samples: 60})
	b.Truth.KPIs.Entities[0].KPIs = append(b.Truth.KPIs.Entities[0].KPIs, model.KPIValue{State: "flow", Mean: 3, Min: 2, Max: 4, Samples: 60})
	c := Compare(a, b)
	presences := map[string]string{}
	for _, row := range c.Results.KPIs {
		presences[row.EntityID+"/"+row.KPI] = row.Presence
	}
	want := map[string]string{"chiller-1/cop": PresenceBoth, "chiller-1/flow": PresenceOnlyB, "chiller-1/power": PresenceOnlyA}
	if !slices.Equal(limitCodes(c)[len(limitCodes(c))-2:], []string{"kpi_only_in_one_run", "single_run_per_side"}) {
		t.Fatalf("limits = %v, want kpi_only_in_one_run", limitCodes(c))
	}
	for key, presence := range want {
		if presences[key] != presence {
			t.Errorf("presences = %v, want %s %s", presences, key, presence)
		}
	}
	for _, row := range c.Results.KPIs {
		if row.Presence != PresenceBoth && (row.Delta != nil || row.RelativeDelta != nil) {
			t.Errorf("row %+v of one side carries a delta", row)
		}
	}
}

func TestDeltasAreBMinusAAndTheRelativeDeltaIsUndefinedWhereAIsZero(t *testing.T) {
	t.Parallel()
	a, b := sealedRun(7), sealedRun(7)
	a.Truth.KPIs.Entities[0].KPIs = []model.KPIValue{{State: "cop", Mean: 5, Min: 0, Max: 6, Samples: 60}}
	b.Truth.KPIs.Entities[0].KPIs = []model.KPIValue{{State: "cop", Mean: 4, Min: 1, Max: 6, Samples: 60}}
	b.Truth.KPIs.Entities[0].Alarms[0].Activations = 5
	b.Truth.KPIs.AlarmSystem.Activations = 6
	c := Compare(a, b)
	row := c.Results.KPIs[0]
	if row.Delta.Mean != -1 || *row.RelativeDelta.Mean != -0.2 || row.RelativeDelta.Min != nil || *row.RelativeDelta.Max != 0 {
		t.Fatalf("delta %+v relative %+v, want mean -1 (-20 %%), min undefined, max 0", row.Delta, row.RelativeDelta)
	}
	if alarm := c.Results.Alarms[0]; alarm.Delta == nil || alarm.Delta.Activations != 3 || c.Results.AlarmSystem.Delta == nil || c.Results.AlarmSystem.Delta.Activations != 4 {
		t.Fatalf("alarms = %+v system = %+v, want activation deltas 3 and 4", c.Results.Alarms, c.Results.AlarmSystem)
	}
}

func TestAWindowedResultIsNotAvailableAndSaysWhy(t *testing.T) {
	t.Parallel()
	c := Compare(sealedRun(7), sealedRun(7))
	if c.WindowedResults.Status != "not_available" || !strings.Contains(c.WindowedResults.Reason, "whole-run") {
		t.Fatalf("windowed results = %+v, want not_available with a reason", c.WindowedResults)
	}
}

func TestARunWithoutKPITruthHasNoRowsAndIsNamedAsALimit(t *testing.T) {
	t.Parallel()
	b := sealedRun(7)
	b.Truth.KPIs = nil
	c := Compare(sealedRun(7), b)
	if !slices.Contains(limitCodes(c), "kpi_truth_absentb") || c.Results.KPIs[0].Presence != PresenceOnlyA || c.Results.AlarmSystem.B != nil {
		t.Fatalf("limits %v results %+v, want kpi_truth_absent for b and A-only rows", limitCodes(c), c.Results)
	}
}

func TestTheComparisonIsTheSameBytesEveryTimeAndNeverRanksTheRuns(t *testing.T) {
	t.Parallel()
	first, err := ComparisonJSON(Compare(sealedRun(7), sealedRun(8)))
	if err != nil {
		t.Fatal(err)
	}
	second, _ := ComparisonJSON(Compare(sealedRun(7), sealedRun(8)))
	if string(first) != string(second) {
		t.Fatal("two comparisons of the same reports differ")
	}
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(first, &doc); err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"winner", "better", "score", "rank", "verdict"} {
		if _, found := doc[forbidden]; found || strings.Contains(string(first), `"`+forbidden+`"`) {
			t.Errorf("the comparison names %q", forbidden)
		}
	}
}

func TestNotComparableNamesWhyARunCannotBeCompared(t *testing.T) {
	t.Parallel()
	sealed := &Report{Run: Run{Mode: "evaluate"}, TruthUnavailable: TruthSealed}
	foreignRun := &Report{Run: Run{Mode: "evaluate"}, TruthUnavailable: foreignArtifact + "run_id differ"}
	open := &Report{Run: Run{Mode: "explore"}, Truth: &Truth{}}
	for name, tc := range map[string]struct {
		report *Report
		want   string
	}{"sealed": {sealed, TruthSealed}, "foreign": {foreignRun, foreignArtifact}, "open": {open, "has not ended"}, "ended": {sealedRun(1), ""}} {
		if got := NotComparable(tc.report); !strings.Contains(got, tc.want) || (tc.want == "" && got != "") {
			t.Errorf("%s: NotComparable = %q, want it to contain %q", name, got, tc.want)
		}
	}
}
