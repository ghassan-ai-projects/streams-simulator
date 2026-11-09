package app

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/suite/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

// firstSliceDigest is the SHA-256 of the committed first-slice campaign's files, concatenated in name order, as generated
// before fault weights could carry a severity range.
const firstSliceDigest = "86bb2ea6330ef755656b98c660511fb26c2756a587b6beea4adc9768253a8094"

func folderDigest(t *testing.T, dir string) string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.New()
	for _, name := range slices.Sorted(slices.Values(entryNames(entries))) {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		sum.Write(raw)
	}
	return hex.EncodeToString(sum.Sum(nil))
}

func entryNames(entries []os.DirEntry) []string {
	names := make([]string, len(entries))
	for i, entry := range entries {
		names[i] = entry.Name()
	}
	return names
}

func TestTheFirstSliceCampaignDrawsBitForBitAsBeforeSeverityRanges(t *testing.T) {
	t.Parallel()
	options, _ := firstSlice(t)
	if got := folderDigest(t, options.Out); got != firstSliceDigest {
		t.Fatalf("first-slice campaign digest %s, want %s", got, firstSliceDigest)
	}
}

func publishEdited(t *testing.T, from, to string) (CampaignOptions, error) {
	t.Helper()
	spec, err := os.ReadFile(testsupport.FirstSliceCampaignSpec())
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	testsupport.WriteFileIn(t, dir, "spec.json", bytes.ReplaceAll(spec, []byte(from), []byte(to)))
	options := CampaignOptions{DomainsDir: testsupport.PumpStationDir(), SpecPath: filepath.Join(dir, "spec.json"), Out: filepath.Join(dir, "out")}
	_, err = PublishCampaign(options)
	return options, err
}

func TestASpecFileGivesAFaultASeverityRangeThatEveryPlanOfItDrawsFrom(t *testing.T) {
	t.Parallel()
	options, err := publishEdited(t, `"bearing_wear": 1`, `"bearing_wear": {"weight": 1, "severity_range": [0.8, 1.25]}`)
	if err != nil {
		t.Fatal(err)
	}
	campaign, drawn := readCampaign(t, filepath.Join(options.Out, rules.CampaignFileName)), 0
	for _, run := range campaign.Runs {
		raw, err := os.ReadFile(filepath.Join(options.Out, run.SessionPlan))
		if err != nil {
			t.Fatal(err)
		}
		plan, err := model.DecodeSessionPlan(raw)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range plan.Disturbances {
			if (d.Fault == "bearing_wear") != (d.Severity != nil) || (d.Severity != nil && (*d.Severity < 0.8 || *d.Severity > 1.25)) {
				t.Fatalf("run %d: %s with severity %v, want a severity in [0.8, 1.25] exactly on bearing_wear", run.Index, d.Fault, d.Severity)
			}
			if d.Severity != nil {
				drawn++
			}
		}
	}
	if drawn != 32 {
		t.Fatalf("%d bearing_wear runs drew a severity, want the 32 the weights give (20 + 10 + 2: a third of the positive runs of each phase)", drawn)
	}
}

func TestAFaultWeightThatIsNeitherANumberNorTheClosedObjectIsRefused(t *testing.T) {
	t.Parallel()
	for name, to := range map[string]string{
		"unknown member":   `"bearing_wear": {"weight": 1, "range": [1, 2]}`,
		"a reversed range": `"bearing_wear": {"weight": 1, "severity_range": [2, 1]}`,
		"one bound":        `"bearing_wear": {"weight": 1, "severity_range": [2]}`,
		"a negative bound": `"bearing_wear": {"weight": 1, "severity_range": [-1, 1]}`,
		"no weight":        `"bearing_wear": {"severity_range": [1, 2]}`,
		"a string":         `"bearing_wear": "1"`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := publishEdited(t, `"bearing_wear": 1`, to)
			var invalid *rules.CampaignSpecError
			if !errors.As(err, &invalid) {
				t.Fatalf("PublishCampaign = %v, want a CampaignSpecError", err)
			}
		})
	}
}
