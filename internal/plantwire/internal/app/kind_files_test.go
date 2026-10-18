package app

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
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
		text string
	}{
		{"a directory holding the kind file", []string{pump}, nil, ""},
		{"a kind file whose unit drifted, beside an empty directory", []string{drifted, kindsDirectoryWith(t, nil)}, wire.ErrKindDrift, " at /metrics/1/unit:"},
		{"no directory at all", nil, wire.ErrKindFileMissing, "(0 loaded)"},
		{"a directory without the kind file", []string{kindsDirectoryWith(t, nil)}, wire.ErrKindFileMissing, "(0 loaded)"},
		{"a directory that does not exist", []string{filepath.Join(pump, "absent")}, fs.ErrNotExist, "absent"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			checked, err := CheckImplements(pumpImplementing(t, "pump"), tc.dirs)
			if tc.want == nil {
				if err != nil || len(checked) != 1 || checked[0].Kind != "pump" {
					t.Fatalf("CheckImplements = %v, %v, want the pump kind checked", checked, err)
				}
				return
			}
			if checked != nil || !errors.Is(err, tc.want) || !strings.Contains(err.Error(), tc.text) {
				t.Fatalf("CheckImplements = %v, %v, want %v mentioning %q and nothing checked", checked, err, tc.want, tc.text)
			}
		})
	}
}

func TestCheckImplementsReadsNothingForADomainWithoutImplements(t *testing.T) {
	t.Parallel()
	malformed := kindsDirectoryWith(t, map[string]string{"thin.kind.json": `{"kind": "thin"}`, "broken.kind.json": "{"})
	pipes := t.TempDir()
	if err := syscall.Mkfifo(filepath.Join(pipes, "pipe.kind.json"), 0o600); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name    string
		dirs    []string
		touched func(err error) bool
	}{
		{"no directory", nil, func(err error) bool { return errors.Is(err, wire.ErrKindFileMissing) }},
		{"a directory holding malformed kind files", []string{malformed}, func(err error) bool {
			invalid, ok := model.AsValidationError(err)
			return ok && invalid.Contract == model.ContractKindFile && strings.HasSuffix(invalid.Source, ".kind.json")
		}},
		{"a directory holding a named pipe", []string{pipes}, func(err error) bool { return errors.Is(err, kernel.ErrNotRegular) }},
		{"a directory that does not exist", []string{filepath.Join(malformed, "absent")}, func(err error) bool { return errors.Is(err, fs.ErrNotExist) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			checked, err := CheckImplements(pumpImplementing(t, ""), tc.dirs)
			if err != nil || checked == nil || len(checked) != 0 {
				t.Fatalf("CheckImplements = %v, %v, want an empty list and no error for a domain that implements nothing", checked, err)
			}
			if _, err := CheckImplements(pumpImplementing(t, "pump"), tc.dirs); !tc.touched(err) {
				t.Fatalf("the same directories checked for a domain that implements = %v: they would have been read and refused, so reading none above proves the skip", err)
			}
		})
	}
}

func TestCheckImplementsRefusesAMalformedKindFileInAnyDirectoryOfADomainThatImplements(t *testing.T) {
	t.Parallel()
	dir := kindsDirectoryWith(t, map[string]string{"thin.kind.json": `{"kind": "thin"}`})
	_, err := CheckImplements(pumpImplementing(t, "pump"), []string{dir})
	if err == nil || !strings.Contains(err.Error(), "thin.kind.json: plantwire: kind file fails its schema") {
		t.Fatalf("CheckImplements = %v, want the schema refusal naming the file", err)
	}
}

func TestValidateKindFilesSummarisesEachFileAndNamesTheOneItCannotRead(t *testing.T) {
	t.Parallel()
	dir := kindsDirectoryWith(t, map[string]string{"pump.kind.json": pumpKindContent(t), "bad.kind.json": "{"})
	summaries, err := ValidateKindFiles([]string{filepath.Join(dir, "pump.kind.json")})
	if err != nil || len(summaries) != 1 || summaries[0].Kind != "pump" || summaries[0].Metrics != 2 || summaries[0].Commands != 3 {
		t.Fatalf("ValidateKindFiles = %+v, %v, want the pump kind file summarized", summaries, err)
	}
	bad := filepath.Join(dir, "bad.kind.json")
	_, err = ValidateKindFiles([]string{bad})
	invalid, ok := model.AsValidationError(err)
	if !ok || invalid.Contract != model.ContractKindFile || invalid.Source != bad || len(invalid.Issues) != 1 || invalid.Issues[0].Located ||
		!strings.Contains(err.Error(), "plantwire: kind file "+bad) {
		t.Fatalf("a file that is not JSON: err = %v, want one unlocated kind-file issue naming %s", err, bad)
	}
	if _, err := ValidateKindFiles([]string{filepath.Join(dir, "absent.kind.json")}); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a missing file: err = %v, want fs.ErrNotExist", err)
	}
}
