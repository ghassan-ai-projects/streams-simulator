package suite_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/suite"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func shipped(t *testing.T, id string) *domain.Compiled {
	t.Helper()
	spec, err := domain.Load(testsupport.Domain(id))
	if err != nil {
		t.Fatal(err)
	}
	return spec
}

func TestGenerateRefusesAMissingDomain(t *testing.T) {
	t.Parallel()
	if _, err := suite.Generate(suite.Config{Profile: "nominal", N: 1}); !errors.Is(err, suite.ErrNoDomain) {
		t.Fatalf("err = %v, want ErrNoDomain", err)
	}
}

func TestScenarioPerturbationsKeepTheAuditJSONShape(t *testing.T) {
	t.Parallel()
	raw, err := json.Marshal(suite.Perturbation{Name: "drop", Params: map[string]any{"rate": 0.5}, FromNS: 1, UntilNS: 2})
	if err != nil || string(raw) != `{"name":"drop","params":{"rate":0.5},"from_ns":1,"until_ns":2}` {
		t.Fatalf("perturbation JSON = %s (%v)", raw, err)
	}
}

func TestGenerateRefusesAnUnknownProfile(t *testing.T) {
	t.Parallel()
	_, err := suite.Generate(suite.Config{Domain: shipped(t, "rotating-machinery"), Profile: "no_such_profile", N: 1, Seed: 1})
	if err == nil || !strings.Contains(err.Error(), `unknown profile "no_such_profile"`) {
		t.Fatalf("err = %v", err)
	}
}

func TestGenerateIsDeterministicAndLabelsEveryScenario(t *testing.T) {
	t.Parallel()
	testsupport.SkipUnderRace(t, "generates and audits a suite on one goroutine")
	cfg := suite.Config{Domain: shipped(t, "shipment-exception"), Profile: "nominal", N: 1, Seed: 5}
	first, err := suite.Generate(cfg)
	if err != nil {
		t.Fatal(err)
	}
	second, err := suite.Generate(cfg)
	if err != nil {
		t.Fatal(err)
	}
	a, errA := json.Marshal(first)
	b, errB := json.Marshal(second)
	if errA != nil || errB != nil || string(a) != string(b) {
		t.Fatalf("the same configuration produced different suites (%v, %v)", errA, errB)
	}
	if first.DomainID != "shipment-exception" || first.Profile != "nominal" || first.Seed != 5 {
		t.Fatalf("identity = %q %q %d", first.DomainID, first.Profile, first.Seed)
	}
	if len(first.Labels) != len(first.Scenarios) {
		t.Fatalf("%d labels for %d scenarios", len(first.Labels), len(first.Scenarios))
	}
}

func TestPublishWritesASuiteFileAndRefusesAnUnknownProfile(t *testing.T) {
	t.Parallel()
	testsupport.SkipUnderRace(t, "generates and audits a suite on one goroutine")
	options := suite.PublishOptions{DomainsDir: testsupport.DomainsDir(), DomainID: "shipment-exception", Profile: "nominal",
		Out: t.TempDir(), N: 1, Seed: 1}
	published, err := suite.Publish(options)
	if err != nil || published.Scenarios == 0 {
		t.Fatalf("published = %+v, %v", published, err)
	}
	options.Profile = "no-such-profile"
	var unknown *suite.UnknownProfileError
	if _, err := suite.Publish(options); !errors.As(err, &unknown) || unknown.Profile != "no-such-profile" {
		t.Fatalf("err = %v, want an UnknownProfileError", err)
	}
}
