package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/suite/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func firstSlice(t *testing.T) (CampaignOptions, PublishedCampaign) {
	t.Helper()
	options := CampaignOptions{DomainsDir: testsupport.PumpStationDir(), SpecPath: testsupport.FirstSliceCampaignSpec(), Out: filepath.Join(t.TempDir(), "campaign")}
	published, err := PublishCampaign(options)
	if err != nil {
		t.Fatal(err)
	}
	return options, published
}

func readCampaign(t *testing.T, path string) rules.Campaign {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var campaign rules.Campaign
	if err := json.Unmarshal(raw, &campaign); err != nil {
		t.Fatal(err)
	}
	return campaign
}

func TestTheFirstSliceCampaignHoldsItsPhasesAndOnePlanPerRun(t *testing.T) {
	t.Parallel()
	options, published := firstSlice(t)
	campaign := readCampaign(t, published.Path)
	phases := map[string]int{}
	for _, run := range campaign.Runs {
		phases[run.Phase]++
		raw, err := os.ReadFile(filepath.Join(options.Out, run.SessionPlan))
		if err != nil {
			t.Fatal(err)
		}
		plan, err := model.DecodeSessionPlan(raw)
		if err != nil || plan.Domain != run.Domain || *plan.Seed != run.Seed {
			t.Fatalf("run %d plan: %+v, %v", run.Index, plan, err)
		}
	}
	if published.Runs != 160 || phases["learn"] != 100 || phases["holdout"] != 50 || phases["disturb"] != 10 {
		t.Fatalf("published %d runs in phases %v", published.Runs, phases)
	}
	if last := campaign.Runs[159]; last.Domain != "pump-station-probe-b" || campaign.Runs[0].Domain != "pump-station" {
		t.Fatalf("first run on %q and last on %q, want the disturb phase on the probe variant", campaign.Runs[0].Domain, last.Domain)
	}
}

func TestTheSameSpecGivesTheSameFilesByteForByte(t *testing.T) {
	t.Parallel()
	first, published := firstSlice(t)
	second, _ := firstSlice(t)
	names, err := os.ReadDir(first.Out)
	if err != nil || len(names) != published.Runs+1 {
		t.Fatalf("%d files in the campaign folder, want %d: %v", len(names), published.Runs+1, err)
	}
	for _, name := range names {
		a, errA := os.ReadFile(filepath.Join(first.Out, name.Name()))
		b, errB := os.ReadFile(filepath.Join(second.Out, name.Name()))
		if errA != nil || errB != nil || !bytes.Equal(a, b) {
			t.Fatalf("%s differs between two publications (%v, %v)", name.Name(), errA, errB)
		}
	}
}

func TestTheCampaignFileCarriesNoTruthMember(t *testing.T) {
	t.Parallel()
	_, published := firstSlice(t)
	raw, err := os.ReadFile(published.Path)
	if err != nil {
		t.Fatal(err)
	}
	var document any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	allowed := map[string]bool{"schema_version": true, "domain": true, "seed": true, "runs": true, "index": true, "phase": true,
		"duration_s": true, "session_plan": true}
	for _, key := range memberNames(document) {
		if !allowed[key] {
			t.Errorf("campaign.json holds member %q, which is not part of what a run is told", key)
		}
	}
}

func memberNames(value any) []string {
	var names []string
	switch typed := value.(type) {
	case map[string]any:
		for key, member := range typed {
			names = append(append(names, key), memberNames(member)...)
		}
	case []any:
		for _, member := range typed {
			names = append(names, memberNames(member)...)
		}
	}
	return names
}

func TestACampaignSpecThatCannotBeUsedIsRefused(t *testing.T) {
	t.Parallel()
	options := CampaignOptions{DomainsDir: testsupport.PumpStationDir(), Out: t.TempDir()}
	dir := t.TempDir()
	testsupport.WriteFileIn(t, dir, "typo.json", []byte(`{"schema_version":"0.1","seeds":1}`))
	options.SpecPath = filepath.Join(dir, "typo.json")
	var invalid *rules.CampaignSpecError
	if _, err := PublishCampaign(options); !errors.As(err, &invalid) {
		t.Fatalf("a spec with an unknown member = %v, want a CampaignSpecError", err)
	}
	options.SpecPath = filepath.Join(dir, "absent.json")
	if _, err := PublishCampaign(options); err == nil || errors.As(err, &invalid) {
		t.Fatalf("a spec file that is absent = %v, want an ordinary read failure", err)
	}
}
