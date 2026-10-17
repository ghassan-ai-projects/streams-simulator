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

func shippedDomains(t *testing.T) []string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(testsupport.DomainsDir(), "*.domain.json"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("shipped domains %v: %v", paths, err)
	}
	return append(paths, testsupport.Example(), filepath.Join(testsupport.RepositoryRoot(), "examples", "chiller-pilot.domain.json"), testsupport.ThermalBoardDomain())
}

func TestTheArduinoThermalTwinLintsCleanAtRest(t *testing.T) {
	t.Parallel()
	spec, err := domain.Load(testsupport.ThermalBoardDomain())
	if err != nil {
		t.Fatal(err)
	}
	report, err := lint.Check(spec, lint.Options{Seed: 1, Hours: 24})
	if err != nil {
		t.Fatal(err)
	}
	for _, finding := range report.Findings {
		t.Errorf("finding %+v, want none: the twin rests at its initial values", finding)
	}
}

func TestCheckRefusesAMissingSpec(t *testing.T) {
	t.Parallel()
	if _, err := lint.Check(nil, lint.Options{Seed: 1, Hours: 1}); !errors.Is(err, lint.ErrNoSpec) {
		t.Fatalf("error = %v, want ErrNoSpec", err)
	}
}

func TestEveryShippedDomainLintsWithoutErrors(t *testing.T) {
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
			if report.Errors != 0 {
				t.Fatalf("%s: %d error findings: %+v", spec.Spec.ID, report.Errors, report.Findings)
			}
		})
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

func TestTheComposedChillerPilotLintsCleanBecauseTheSmallChillerFollowsTheSeasonalAirHandlingUnit(t *testing.T) {
	t.Parallel()
	spec, err := domain.Load(filepath.Join(testsupport.RepositoryRoot(), "examples", "chiller-pilot.domain.json"))
	if err != nil {
		t.Fatal(err)
	}
	report, err := lint.Check(spec, lint.Options{Seed: 1, Hours: 24})
	if err != nil {
		t.Fatal(err)
	}
	for _, finding := range report.Findings {
		t.Errorf("finding %+v, want none on the pilot", finding)
	}
}

func TestFailedIsTrueOnlyForAReportWithErrorFindings(t *testing.T) {
	t.Parallel()
	if lint.Failed(&lint.Report{Warnings: 2}) || !lint.Failed(&lint.Report{Errors: 1}) {
		t.Fatal("Failed must follow the error count, never the warnings")
	}
}
