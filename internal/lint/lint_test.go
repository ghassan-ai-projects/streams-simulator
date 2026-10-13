package lint_test

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/lint"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

var knownErrors = map[string]int{"discrete-line-oee": 4}

func shippedDomains(t *testing.T) []string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(testsupport.DomainsDir(), "*.domain.json"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("shipped domains %v: %v", paths, err)
	}
	return append(paths, testsupport.Example(), filepath.Join(testsupport.RepositoryRoot(), "examples", "chiller-pilot.domain.json"))
}

func TestCheckRefusesAMissingSpec(t *testing.T) {
	t.Parallel()
	if _, err := lint.Check(nil, lint.Options{Seed: 1, Hours: 1}); !errors.Is(err, lint.ErrNoSpec) {
		t.Fatalf("error = %v, want ErrNoSpec", err)
	}
}

func TestEveryShippedDomainLintsWithoutErrorsOtherThanTheKnownOnes(t *testing.T) {
	t.Parallel()
	testsupport.SkipUnderRace(t, "simulates ten domains for a nominal day each")
	for _, path := range shippedDomains(t) {
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()
			spec, err := domain.Load(path)
			if err != nil {
				t.Fatal(err)
			}
			report, err := lint.Check(spec, lint.Options{Seed: 1, Hours: 24})
			if err != nil {
				t.Fatal(err)
			}
			if want := knownErrors[spec.Spec.ID]; report.Errors != want {
				t.Fatalf("%s: %d error findings, want %d: %+v", spec.Spec.ID, report.Errors, want, report.Findings)
			}
		})
	}
}

func TestTheKnownErrorsAreTheCycleTimeChannelOutsideItsRange(t *testing.T) {
	t.Parallel()
	spec, err := domain.Load(testsupport.Domain("discrete-line-oee"))
	if err != nil {
		t.Fatal(err)
	}
	report, err := lint.Check(spec, lint.Options{Seed: 1, Hours: 1})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range report.Findings {
		if f.Severity == "error" && (f.Code != "out_of_range" || f.Channel != "cycle_time_s") {
			t.Fatalf("unexpected error finding %+v", f)
		}
	}
}

func TestTheSameDomainSeedAndHoursGiveTheSameBytes(t *testing.T) {
	t.Parallel()
	spec, err := domain.Load(testsupport.Example())
	if err != nil {
		t.Fatal(err)
	}
	encode := func() string {
		report, err := lint.Check(spec, lint.Options{Seed: 3, Hours: 2})
		if err != nil {
			t.Fatal(err)
		}
		raw, err := json.Marshal(report)
		if err != nil {
			t.Fatal(err)
		}
		return string(raw)
	}
	if first, second := encode(), encode(); first != second {
		t.Fatalf("reports differ:\n%s\n%s", first, second)
	}
}
