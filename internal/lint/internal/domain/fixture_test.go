package domain

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func compile(t *testing.T, mutate func(*model.DomainSpec)) *domain.Compiled {
	t.Helper()
	raw, err := os.ReadFile("testdata/sane.domain.json")
	if err != nil {
		t.Fatal(err)
	}
	var spec model.DomainSpec
	if err := json.Unmarshal(raw, &spec); err != nil {
		t.Fatal(err)
	}
	if mutate != nil {
		mutate(&spec)
	}
	if raw, err = json.Marshal(spec); err != nil {
		t.Fatal(err)
	}
	compiled, err := domain.Parse(raw, "lint-test")
	if err != nil {
		t.Fatal(err)
	}
	return compiled
}

func lintOf(t *testing.T, hours float64, mutate func(*model.DomainSpec)) *Report {
	t.Helper()
	report, err := Check(compile(t, mutate), Options{Seed: 1, Hours: hours})
	if err != nil {
		t.Fatal(err)
	}
	return report
}

func findingFor(report *Report, code, entity, subject string) (Finding, bool) {
	for _, f := range report.Findings {
		if f.Code == code && f.Entity == entity && f.subject() == subject {
			return f, true
		}
	}
	return Finding{}, false
}

func plantKind(spec *model.DomainSpec) *model.Kind {
	return &spec.Kinds[0]
}

func lagOf(spec *model.DomainSpec) *model.F1Dyn {
	return plantKind(spec).Dynamics[0].F1
}
