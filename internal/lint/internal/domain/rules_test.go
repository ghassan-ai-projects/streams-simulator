package domain

import (
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestASanePlantHasNoFindings(t *testing.T) {
	t.Parallel()
	for _, hours := range []float64{0.5, 24} {
		report := lintOf(t, hours, nil)
		if len(report.Findings) != 0 || report.Errors != 0 || report.Warnings != 0 {
			t.Fatalf("hours %v: findings %+v, want none", hours, report.Findings)
		}
	}
}

func TestEachRuleFiresOnItsMinimalMutation(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, code, severity, subject, detail string
		hours                                 float64
		mutate                                func(*model.DomainSpec)
	}{
		{"reading outside its range", codeOutOfRange, severityError, "flow_in", "reading 2 outside the declared range [0, 1] (worst), 61 of 61 samples outside", 1,
			func(s *model.DomainSpec) { plantKind(s).Channels[0].Range.Max = 1 }},
		{"state saturated at its bound", codePinnedAtBound, severityWarning, "x", "state sits at its declared max 3 in", 2,
			func(s *model.DomainSpec) {
				plantKind(s).State[0].Initial, plantKind(s).State[1].Initial = 5, 0
				plantKind(s).State[1].Max = new(float64)
				*plantKind(s).State[1].Max = 3
				lagOf(s).TimeConstantS = 60
			}},
		{"state still ramping at the end", codeNotSettled, severityWarning, "x", "still moves", 2,
			func(s *model.DomainSpec) {
				plantKind(s).State[1].Initial = 0
				lagOf(s).TimeConstantS = 3_600_000
			}},
		{"balance that does not close", codeConservation, severityWarning, "leak", "inputs minus outputs reaches 2 at 0 s", 2,
			func(s *model.DomainSpec) { plantKind(s).State[1].Initial = 0 }},
		{"state that overflows", codeNonFinite, severityError, "x", "samples are not finite", 1,
			func(s *model.DomainSpec) { lagOf(s).Gain = 1e308 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			report := lintOf(t, tc.hours, tc.mutate)
			found, ok := findingFor(report, tc.code, "p-1", tc.subject)
			if !ok || found.Severity != tc.severity || !strings.Contains(found.Detail, tc.detail) {
				t.Fatalf("findings %+v, want %s %s on p-1 %s containing %q", report.Findings, tc.severity, tc.code, tc.subject, tc.detail)
			}
		})
	}
}

func TestAFindingIsReportedOncePerSeriesAtItsFirstSample(t *testing.T) {
	t.Parallel()
	report := lintOf(t, 1, func(s *model.DomainSpec) { plantKind(s).Channels[0].Range.Max = 1 })
	count := 0
	for _, f := range report.Findings {
		if f.Code == codeOutOfRange && f.Entity == "p-2" && f.Channel == "flow_in" {
			count++
			if f.AtS != 0 {
				t.Fatalf("at_s %v, want the first sample 0", f.AtS)
			}
		}
	}
	if count != 1 || report.Errors != 2 {
		t.Fatalf("p-2 findings %d, errors %d, want one finding for p-2 and one error per plant", count, report.Errors)
	}
}

func TestAnIdleStateAtItsBoundIsNotPinned(t *testing.T) {
	t.Parallel()
	report := lintOf(t, 2, func(s *model.DomainSpec) {
		plantKind(s).State[1].Initial, plantKind(s).State[0].Initial = 0, 0
		plantKind(s).State[1].Min = new(float64)
	})
	if _, pinned := findingFor(report, codePinnedAtBound, "p-1", "x"); pinned {
		t.Fatalf("a state that never moves was reported as pinned: %+v", report.Findings)
	}
}

func TestACyclicStateIsNotExpectedToSettle(t *testing.T) {
	t.Parallel()
	report := lintOf(t, 24, func(s *model.DomainSpec) {
		kind := plantKind(s)
		kind.Dynamics = append(kind.Dynamics, model.Dynamics{
			Target: "u", Tier: "F0", F0: &model.F0Dyn{Baseline: 2, Seasonality: []model.Seasonality{{PeriodS: 86400, Amplitude: 1}}},
		})
		lagOf(s).TimeConstantS = 3600
	})
	if _, drifting := findingFor(report, codeNotSettled, "p-1", "x"); drifting {
		t.Fatalf("a state driven by a daily cycle was reported as unsettled: %+v", report.Findings)
	}
}

func TestAnUnbalancedFaultNeedsAllItsChannelsToBeNumeric(t *testing.T) {
	t.Parallel()
	report := lintOf(t, 1, func(s *model.DomainSpec) {
		plantKind(s).State[1].Initial = 0
		plantKind(s).Channels[0].ValueType, plantKind(s).Channels[0].Unit, plantKind(s).Channels[0].Range = "boolean", "", nil
	})
	if _, unbalanced := findingFor(report, codeConservation, "p-1", "leak"); unbalanced {
		t.Fatalf("a balance over an unsampled channel was judged: %+v", report.Findings)
	}
}
