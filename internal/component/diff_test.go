package component

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func thermalVariants(t *testing.T) (before, after string) {
	t.Helper()
	library, dir := builtInLibrary(t), t.TempDir()
	choices := readFile(t, testsupport.TwinThermalChoices())
	for name, text := range map[string][]byte{"before": choices, "after": bytesReplace(t, choices, `"gain": -0.02`, `"gain": -0.03`)} {
		twin, err := Twin(library, thermalKinds(t), text, name+".twin.json")
		if err != nil {
			t.Fatal(err)
		}
		testsupport.WriteFileIn(t, dir, name+".domain.json", twin.Bytes)
	}
	return filepath.Join(dir, "before.domain.json"), filepath.Join(dir, "after.domain.json")
}

func relabelled(t *testing.T, document any, dir string) any {
	t.Helper()
	raw, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	var out any
	if err := json.Unmarshal(bytes.ReplaceAll(raw, []byte(filepath.ToSlash(dir)), []byte("workspace")), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestTheDiffOfTheThermalTwinWithAChangedGainMatchesItsGolden(t *testing.T) {
	t.Parallel()
	before, after := thermalVariants(t)
	got, err := DiffDomains(before, after)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Identity.SameIDAndVersionDifferentDigest || got.Summary.Changes != 1 || got.Kinds[0].Kind != "thermal_sensor" {
		t.Fatalf("diff = %+v, want one gain change on the sensor under an unchanged version", got.Summary)
	}
	checkGolden(t, relabelled(t, got, filepath.Dir(before)), "diff-thermal-gain.golden.json")
	markdown := DiffMarkdown(got)
	if !strings.Contains(string(markdown), "`gain_changed`") || !strings.Contains(string(markdown), "## Kind thermal_sensor (changed)") {
		t.Fatalf("markdown = %s", markdown)
	}
}

func TestADiffOfARunFolderAgainstTheDomainItWasMadeFromIsEmpty(t *testing.T) {
	t.Parallel()
	got, err := DiffDomains(testsupport.ReportRun(), testsupport.Domain("aquaculture-pond"))
	if err != nil {
		t.Fatal(err)
	}
	if got.A.Source.Kind != "run" || got.B.Source.Kind != "file" || !got.Identity.SameDigest || got.Summary.Changes != 0 {
		t.Fatalf("diff = %+v %+v, want the recorded domain equal to the file", got.A.Source, got.Identity)
	}
	if err := model.ValidateDomainDiff(mustMarshal(t, got)); err != nil {
		t.Fatal(err)
	}
}

func TestADiffRefusesAMissingSideAndABrokenDomain(t *testing.T) {
	t.Parallel()
	pond := testsupport.Domain("aquaculture-pond")
	if _, err := DiffDomains(pond, filepath.Join(t.TempDir(), "none.domain.json")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("DiffDomains = %v, want a missing file", err)
	}
	broken := filepath.Join(t.TempDir(), "broken.domain.json")
	if err := os.WriteFile(broken, []byte(`{"id":"X"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := DiffDomains(broken, pond)
	if code, known := failureCodeOf(err); !known || code != "domain_invalid" {
		t.Fatalf("DiffDomains = %v, want domain_invalid", err)
	}
}

func failureCodeOf(err error) (string, bool) {
	return model.FailureCodeOf(err)
}

func mustMarshal(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
