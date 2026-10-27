package plantwire

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/suite"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestEveryRunOfTheFirstSliceCampaignIsAPlanTheSessionAccepts(t *testing.T) {
	t.Parallel()
	out := filepath.Join(t.TempDir(), "campaign")
	published, err := suite.PublishCampaign(suite.CampaignOptions{DomainsDir: testsupport.PumpStationDir(), SpecPath: testsupport.FirstSliceCampaignSpec(), Out: out})
	if err != nil {
		t.Fatal(err)
	}
	var campaign struct {
		Runs []struct {
			Index       int    `json:"index"`
			Domain      string `json:"domain"`
			SessionPlan string `json:"session_plan"`
		} `json:"runs"`
	}
	raw, err := os.ReadFile(published.Path)
	if err != nil || json.Unmarshal(raw, &campaign) != nil || len(campaign.Runs) != published.Runs {
		t.Fatalf("campaign file: %v", err)
	}
	for _, run := range campaign.Runs {
		plan, err := os.ReadFile(filepath.Join(out, run.SessionPlan))
		if err != nil {
			t.Fatal(err)
		}
		check, err := CheckConnect(ConnectOptions{Plan: bytes.NewReader(plan), Source: run.SessionPlan, DomainsDir: testsupport.PumpStationDir(),
			AdaptersDir: testsupport.AdaptersDir(), SliceNS: 1e9})
		if err != nil || check.Domain != run.Domain || len(check.Instances) != 8 {
			t.Fatalf("run %d: CheckConnect = %+v, %v", run.Index, check, err)
		}
	}
}
