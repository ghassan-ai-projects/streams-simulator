package domain

import (
	"encoding/json"
	"errors"
	"math"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestCheckRefusesAMissingSpec(t *testing.T) {
	t.Parallel()
	if _, err := Check(nil, Options{Seed: 1, Hours: 1}); !errors.Is(err, ErrNoSpec) {
		t.Fatalf("error = %v, want ErrNoSpec", err)
	}
}

func TestCheckRefusesHoursOutsideTheNominalSpan(t *testing.T) {
	t.Parallel()
	for _, hours := range []float64{0, -1, math.NaN(), math.Inf(1), MaxHours + 1} {
		if _, err := Check(compile(t, nil), Options{Seed: 1, Hours: hours}); !errors.Is(err, ErrInvalidHours) {
			t.Errorf("hours %v: error = %v, want ErrInvalidHours", hours, err)
		}
	}
}

func TestSamplesAreEverySixtySecondsPlusTheEnd(t *testing.T) {
	t.Parallel()
	cases := []struct {
		hours   float64
		samples int
	}{{0.5, 31}, {0.51, 32}, {24, 1441}, {1.0 / 3600, 2}}
	for _, tc := range cases {
		if got := lintOf(t, tc.hours, nil); got.Samples != tc.samples || got.Hours != tc.hours {
			t.Errorf("hours %v: samples %d, want %d", tc.hours, got.Samples, tc.samples)
		}
	}
}

func TestReportNamesTheDomainAndCountsBySeverity(t *testing.T) {
	t.Parallel()
	spec := compile(t, func(s *model.DomainSpec) {
		plantKind(s).Channels[0].Range.Max = 1
		plantKind(s).State[1].Initial = 0
	})
	report, err := Check(spec, Options{Seed: 7, Hours: 1})
	if err != nil {
		t.Fatal(err)
	}
	if report.Domain != "lint-sane" || report.Digest != spec.Digest || report.Seed != 7 {
		t.Fatalf("identity = %s %s %d, want lint-sane %s 7", report.Domain, report.Digest, report.Seed, spec.Digest)
	}
	if report.Errors != 2 || report.Warnings != 2 || len(report.Findings) != 4 {
		t.Fatalf("errors %d warnings %d findings %d, want 2, 2 and 4: %+v", report.Errors, report.Warnings, len(report.Findings), report.Findings)
	}
}

func TestReportIsByteIdenticalForTheSameInputs(t *testing.T) {
	t.Parallel()
	encode := func() string {
		raw, err := json.Marshal(lintOf(t, 2, func(s *model.DomainSpec) { plantKind(s).State[1].Initial = 0 }))
		if err != nil {
			t.Fatal(err)
		}
		return string(raw)
	}
	if first, second := encode(), encode(); first != second {
		t.Fatalf("two reports differ:\n%s\n%s", first, second)
	}
}

func TestFindingsAreSortedByEntitySubjectCodeAndTime(t *testing.T) {
	t.Parallel()
	report := lintOf(t, 1, func(s *model.DomainSpec) {
		plantKind(s).Channels[0].Range.Max = 1
		plantKind(s).State[1].Initial = 0
	})
	if !slices.IsSortedFunc(report.Findings, compareFindings) {
		t.Fatalf("findings are not in order: %+v", report.Findings)
	}
	first, last := report.Findings[0], report.Findings[len(report.Findings)-1]
	if first.Entity != "p-1" || first.subject() != "flow_in" || last.Entity != "p-2" {
		t.Fatalf("first %+v, last %+v, want p-1 flow_in first and p-2 last", first, last)
	}
}

func TestLastTenthStartsAtTheFirstSampleInsideIt(t *testing.T) {
	t.Parallel()
	g, err := newGrid(0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if got := g.lastTenthStart(); got != 540 {
		t.Fatalf("last tenth starts at sample %d, want 540 (hour 9)", got)
	}
}
