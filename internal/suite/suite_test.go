package suite_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/suite"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestScenarioPerturbationsKeepTheAuditJSONShape(t *testing.T) {
	t.Parallel()
	raw, err := json.Marshal(suite.Perturbation{Name: "drop", Params: map[string]any{"rate": 0.5}, FromNS: 1, UntilNS: 2})
	if err != nil || string(raw) != `{"name":"drop","params":{"rate":0.5},"from_ns":1,"until_ns":2}` {
		t.Fatalf("perturbation JSON = %s (%v)", raw, err)
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

func TestPublishCampaignWritesTheCampaignAndRefusesASpecThatBreaksARule(t *testing.T) {
	t.Parallel()
	options := suite.CampaignOptions{DomainsDir: testsupport.PumpStationDir(), SpecPath: testsupport.FirstSliceCampaignSpec(), Out: t.TempDir()}
	published, err := suite.PublishCampaign(options)
	if err != nil || published.Runs != 160 {
		t.Fatalf("published = %+v, %v", published, err)
	}
	options.DomainsDir = testsupport.DomainsDir()
	if _, err := suite.PublishCampaign(options); !errors.Is(err, domain.ErrUnknownDomain) {
		t.Fatalf("a campaign naming a domain the directory lacks = %v, want ErrUnknownDomain", err)
	}
	options.DomainsDir, options.SpecPath = testsupport.PumpStationDir(), testsupport.PumpStationDomain()
	var invalid *suite.CampaignSpecError
	if _, err := suite.PublishCampaign(options); !errors.As(err, &invalid) {
		t.Fatalf("err = %v, want a CampaignSpecError", err)
	}
}
