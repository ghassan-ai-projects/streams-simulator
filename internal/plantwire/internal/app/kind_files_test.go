package app

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	wire "github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func pumpImplementing(t *testing.T, implements string) *model.DomainSpec {
	t.Helper()
	compiled, err := domain.Load(testsupport.PumpBoardDomain())
	if err != nil {
		t.Fatal(err)
	}
	compiled.Spec.Kinds[0].Implements = implements
	return compiled.Spec
}

func kindsDirectoryWith(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func pumpKindContent(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(testsupport.PumpKindFile())
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestCheckImplementsComparesTheDomainWithTheKindFilesOfItsDirectories(t *testing.T) {
	t.Parallel()
	pump := kindsDirectoryWith(t, map[string]string{"pump.kind.json": pumpKindContent(t), "readme.txt": "not a kind file"})
	drifted := kindsDirectoryWith(t, map[string]string{"pump.kind.json": strings.Replace(pumpKindContent(t), `"%"`, `"pct"`, 1)})
	cases := []struct {
		name string
		dirs []string
		want error
	}{
		{"a directory holding the kind file", []string{pump}, nil},
		{"a kind file whose unit drifted, beside an empty directory", []string{drifted, kindsDirectoryWith(t, nil)}, wire.ErrKindDrift},
		{"no directory at all", nil, wire.ErrKindFileMissing},
		{"a directory without the kind file", []string{kindsDirectoryWith(t, nil)}, wire.ErrKindFileMissing},
		{"a directory that does not exist", []string{filepath.Join(pump, "absent")}, fs.ErrNotExist},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := CheckImplements(pumpImplementing(t, "pump"), tc.dirs)
			if (tc.want == nil && err != nil) || (tc.want != nil && !errors.Is(err, tc.want)) {
				t.Fatalf("CheckImplements = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestCheckImplementsReadsNothingForADomainWithoutImplements(t *testing.T) {
	t.Parallel()
	if err := CheckImplements(pumpImplementing(t, ""), nil); err != nil {
		t.Fatalf("CheckImplements = %v, want a domain without implements unchanged", err)
	}
}

func TestCheckImplementsRefusesAMalformedKindFileInAnyDirectory(t *testing.T) {
	t.Parallel()
	dir := kindsDirectoryWith(t, map[string]string{"thin.kind.json": `{"kind": "thin"}`})
	err := CheckImplements(pumpImplementing(t, "pump"), []string{dir})
	if err == nil || !strings.Contains(err.Error(), "thin.kind.json: plantwire: kind file fails its schema") {
		t.Fatalf("CheckImplements = %v, want the schema refusal naming the file", err)
	}
}

func TestValidateKindFilesSummarisesEachFileAndNamesTheOneItCannotRead(t *testing.T) {
	t.Parallel()
	dir := kindsDirectoryWith(t, map[string]string{"pump.kind.json": pumpKindContent(t), "bad.kind.json": "{"})
	summaries, err := ValidateKindFiles([]string{filepath.Join(dir, "pump.kind.json")})
	if err != nil || len(summaries) != 1 || summaries[0].Kind != "pump" || summaries[0].Metrics != 2 || summaries[0].Commands != 3 {
		t.Fatalf("ValidateKindFiles = %+v, %v, want the pump kind file summarised", summaries, err)
	}
	if _, err := ValidateKindFiles([]string{filepath.Join(dir, "bad.kind.json")}); err == nil || !strings.Contains(err.Error(), "kind file ") {
		t.Fatalf("a file that is not JSON: err = %v", err)
	}
	if _, err := ValidateKindFiles([]string{filepath.Join(dir, "absent.kind.json")}); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a missing file: err = %v, want fs.ErrNotExist", err)
	}
}
