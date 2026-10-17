package director_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/director"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func writeTemp(t *testing.T, dir, name string, content []byte) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), content, 0o600); err != nil {
		t.Fatal(err)
	}
}

func implementingDomains(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(testsupport.PumpBoardDomain())
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	doc["kinds"].([]any)[0].(map[string]any)["implements"] = "pump"
	out, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	writeTemp(t, dir, "pump-board.domain.json", out)
	return dir
}

func pumpKindsFolder(t *testing.T, withKindFile bool) string {
	t.Helper()
	dir := t.TempDir()
	if withKindFile {
		raw, err := os.ReadFile(testsupport.PumpKindFile())
		if err != nil {
			t.Fatal(err)
		}
		writeTemp(t, dir, "pump.kind.json", raw)
	}
	return dir
}

func TestTheDirectorChecksAnImplementingDomainWithTheCheckItWasGiven(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name         string
		withKindFile bool
		want         string
	}{
		{"a kinds folder that holds the kind file", true, ""},
		{"a kinds folder without the kind file", false, "no kind file of that name was loaded"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			folder := pumpKindsFolder(t, tc.withKindFile)
			check := func(spec *model.DomainSpec) error { return plantwire.CheckImplements(spec, []string{folder}) }
			d, err := director.NewFromDirs(t.Context(), implementingDomains(t), testsupport.AdaptersDir(), t.TempDir(), nil, check)
			if err != nil {
				t.Fatal(err)
			}
			_, err = d.CreateWorld(director.WorldRequest{Domain: "pump-board", Seed: 1})
			if tc.want == "" && err != nil {
				t.Fatalf("CreateWorld = %v, want a world", err)
			}
			if tc.want != "" && (err == nil || !strings.HasPrefix(err.Error(), string(director.CodeDomainInvalid)) || !strings.Contains(err.Error(), tc.want)) {
				t.Fatalf("CreateWorld = %v, want domain_invalid mentioning %q", err, tc.want)
			}
		})
	}
}

func TestADirectorWithoutAKindsCheckRefusesAnImplementingDomain(t *testing.T) {
	t.Parallel()
	d, err := director.NewFromDirs(t.Context(), implementingDomains(t), testsupport.AdaptersDir(), t.TempDir(), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.CreateWorld(director.WorldRequest{Domain: "pump-board", Seed: 1}); err == nil || !strings.HasPrefix(err.Error(), string(director.CodeDomainInvalid)) {
		t.Fatalf("CreateWorld = %v, want domain_invalid", err)
	}
}
