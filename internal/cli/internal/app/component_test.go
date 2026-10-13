package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/cli/internal/process"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func builtInComponents() string { return filepath.Join(testsupport.RepositoryRoot(), "components") }

func builtInFileCounts(t *testing.T) (equipment, instruments int) {
	t.Helper()
	equipmentFiles, err := filepath.Glob(filepath.Join(builtInComponents(), "equipment", "*.component.json"))
	if err != nil {
		t.Fatal(err)
	}
	instrumentFiles, err := filepath.Glob(filepath.Join(builtInComponents(), "instruments", "*.instrument.json"))
	if err != nil {
		t.Fatal(err)
	}
	return len(equipmentFiles), len(instrumentFiles)
}

func pilotComposition() string {
	return filepath.Join(builtInComponents(), "compositions", "chiller-pilot.composition.json")
}

func invokeWithEnvironment(t *testing.T, environment map[string]string, args ...string) outcome {
	t.Helper()
	var stdout, stderr bytes.Buffer
	env := process.Env{Stdout: &stdout, Stderr: &stderr, Now: func() time.Time { return fixedNow },
		Getenv: func(name string) string { return environment[name] }}
	code := Run(append([]string{"streamsim"}, args...), Build{Version: "v-test", Commit: "c-test"}, env)
	return outcome{code: code, stdout: stdout.String(), stderr: stderr.String()}
}

func customerLibrary(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	raw, err := os.ReadFile(filepath.Join(builtInComponents(), "equipment", "ahu.component.json"))
	if err != nil {
		t.Fatal(err)
	}
	custom := strings.Replace(string(raw), `"id": "streamsim/ahu"`, `"id": "acme/ahu"`, 1)
	path := filepath.Join(dir, "equipment", "ahu.component.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(custom), 0o600); err != nil { //nolint:gosec // G703: the path is under t.TempDir and the name is a test constant
		t.Fatal(err)
	}
	return dir
}

func TestComponentListPrintsEveryLibraryEntry(t *testing.T) {
	t.Parallel()
	doc := invoke(t, "component", "list", "--components-dir", builtInComponents()).mustSucceed(t).document(t, "component_list")
	equipment, instruments := builtInFileCounts(t)
	if entries, _ := doc["components"].([]any); len(entries) != equipment+instruments {
		t.Fatalf("components = %d, want %d: %v", len(entries), equipment+instruments, doc)
	}
}

func TestComponentWithoutADirectoryReadsTheComponentsDirectoryOfTheWorkingDirectory(t *testing.T) {
	t.Parallel()
	failure := invoke(t, "component").failure(t)
	if failure["code"] != "file_not_found" || failure["source"] != "components" {
		t.Fatalf("failure = %v, want file_not_found for the default components directory", failure)
	}
}

func TestComponentShowPrintsTheEntryAndItsDocument(t *testing.T) {
	t.Parallel()
	doc := invoke(t, "component", "show", "streamsim/rtd", "--components-dir", builtInComponents()).mustSucceed(t).document(t, "component_show")
	entry, _ := doc["entry"].(map[string]any)
	document, _ := doc["document"].(map[string]any)
	if entry["id"] != "streamsim/rtd" || entry["kind"] != "instrument" || document["tag"] != "TT" {
		t.Fatalf("show = %v, want the RTD instrument", doc)
	}
}

func TestComponentValidateCountsTheLibrary(t *testing.T) {
	t.Parallel()
	doc := invoke(t, "component", "validate", "--components-dir", builtInComponents()).mustSucceed(t).document(t, "component_validate")
	equipment, instruments := builtInFileCounts(t)
	if doc["equipment"] != float64(equipment) || doc["instruments"] != float64(instruments) {
		t.Fatalf("validate = %v, want %d equipment and %d instruments", doc, equipment, instruments)
	}
}

func TestComponentDirectoriesFromTheEnvironmentComeAfterTheFlags(t *testing.T) {
	t.Parallel()
	environment := map[string]string{"STREAMSIM_COMPONENTS_DIRS": customerLibrary(t)}
	doc := invokeWithEnvironment(t, environment, "component", "validate", "--components-dir", builtInComponents()).
		mustSucceed(t).document(t, "component_validate")
	directories, _ := doc["directories"].([]any)
	builtInEquipment, _ := builtInFileCounts(t)
	if len(directories) != 2 || directories[0] != builtInComponents() || directories[1] != environment["STREAMSIM_COMPONENTS_DIRS"] || doc["equipment"] != float64(builtInEquipment+1) {
		t.Fatalf("validate = %v, want the flag directory then the environment directory", doc)
	}
}

func TestComponentEnvironmentListsSplitAtThePathListSeparator(t *testing.T) {
	t.Parallel()
	first, second := customerLibrary(t), t.TempDir()
	environment := map[string]string{"STREAMSIM_COMPONENTS_DIRS": first + string(os.PathListSeparator) + second}
	doc := invokeWithEnvironment(t, environment, "component", "validate", "--components-dir", builtInComponents()).
		mustSucceed(t).document(t, "component_validate")
	directories, _ := doc["directories"].([]any)
	if len(directories) != 3 || directories[1] != first || directories[2] != second {
		t.Fatalf("directories = %v, want the flag directory, then both environment entries", directories)
	}
}

func TestComponentCommandsRefuseWhatTheyCannotDo(t *testing.T) {
	t.Parallel()
	dir := []string{"--components-dir", builtInComponents()}
	cases := []struct {
		name string
		args []string
		code string
		exit int
	}{
		{"an unknown verb", []string{"component", "frobnicate"}, "usage_invalid", 2},
		{"show without a reference", []string{"component", "show"}, "usage_invalid", 2},
		{"list with an argument", []string{"component", "list", "extra"}, "usage_invalid", 2},
		{"validate with an argument", []string{"component", "validate", "extra"}, "usage_invalid", 2},
		{"show of an absent entry", []string{"component", "show", "streamsim/nothing"}, "component_failed", 1},
		{"a directory that does not exist", []string{"component", "list", "--components-dir", "no-such-dir"}, "file_not_found", 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := invoke(t, append(tc.args, dir...)...)
			if got.code != tc.exit {
				t.Fatalf("exit %d, want %d: %+v", got.code, tc.exit, got)
			}
			if code := got.failure(t)["code"]; code != tc.code {
				t.Fatalf("failure code = %v, want %s", code, tc.code)
			}
		})
	}
}

func TestComponentValidateReportsAnInvalidLibraryFileWithItsPointers(t *testing.T) {
	t.Parallel()
	dir := customerLibrary(t)
	path := filepath.Join(dir, "equipment", "ahu.component.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	broken := strings.Replace(string(raw), `"unit_family": "temperature"`, `"unit_family": "gizmo"`, 1)
	if err := os.WriteFile(path, []byte(broken), 0o600); err != nil { //nolint:gosec // G703: the path is under t.TempDir and the name is a test constant
		t.Fatal(err)
	}
	failure := invoke(t, "component", "validate", "--components-dir", builtInComponents(), "--components-dir", dir).failure(t)
	pointers := map[any]bool{}
	issues, _ := failure["issues"].([]any)
	for _, issue := range issues {
		pointers[issue.(map[string]any)["pointer"]] = true
	}
	if failure["code"] != "component_invalid" || failure["source"] != path || !pointers["/measurement_points/0/unit_family"] {
		t.Fatalf("failure = %v, want component_invalid at the unit family of %s", failure, path)
	}
}
