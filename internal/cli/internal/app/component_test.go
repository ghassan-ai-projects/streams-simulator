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

func builtInComponents() string { return testsupport.ComponentsDir() }

func pilotComposition() string { return testsupport.PilotComposition() }

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
	folder := filepath.Join(dir, "equipment")
	if err := os.MkdirAll(folder, 0o750); err != nil {
		t.Fatal(err)
	}
	testsupport.WriteFileIn(t, folder, "ahu.component.json", []byte(custom))
	return dir
}

func TestComponentListPrintsEveryLibraryEntry(t *testing.T) {
	t.Parallel()
	doc := invoke(t, "component", "list").mustSucceed(t).document(t, "component_list")
	equipment, instruments := builtInFileCounts(t)
	if entries, _ := doc["components"].([]any); len(entries) != equipment+instruments {
		t.Fatalf("components = %d, want %d: %v", len(entries), equipment+instruments, doc)
	}
}

func TestComponentExamplesListTheShippedExamplePlantsAndPrintOne(t *testing.T) {
	t.Parallel()
	list := invoke(t, "component", "examples").mustSucceed(t).document(t, "component_examples")
	entries, _ := list["examples"].([]any)
	first, _ := entries[0].(map[string]any)
	if len(entries) < 2 || first["id"] != "chiller-pilot" || first["title"] != "Chiller pilot" {
		t.Fatalf("examples = %v", list)
	}
	shown := invoke(t, "component", "example", "factory-utilities").mustSucceed(t).document(t, "component_example")
	if example, _ := shown["example"].(map[string]any); example["id"] != "factory-utilities" || example["equipment"] == nil {
		t.Fatalf("example = %v", shown)
	}
}

func TestComponentWithoutADirectoryListsTheBuiltInLibraryWhateverTheWorkingDirectory(t *testing.T) {
	t.Parallel()
	doc := invoke(t, "component").mustSucceed(t).document(t, "component_list")
	equipment, instruments := builtInFileCounts(t)
	if entries, _ := doc["components"].([]any); len(entries) != equipment+instruments {
		t.Fatalf("components = %d, want the %d built-in entries: %v", len(entries), equipment+instruments, doc)
	}
}

func TestComponentShowPrintsTheEntryAndItsDocument(t *testing.T) {
	t.Parallel()
	doc := invoke(t, "component", "show", "streamsim/rtd").mustSucceed(t).document(t, "component_show")
	entry, _ := doc["entry"].(map[string]any)
	document, _ := doc["document"].(map[string]any)
	if entry["id"] != "streamsim/rtd" || entry["kind"] != "instrument" || document["tag"] != "TT" {
		t.Fatalf("show = %v, want the RTD instrument", doc)
	}
}

func TestComponentValidateCountsTheLibrary(t *testing.T) {
	t.Parallel()
	doc := invoke(t, "component", "validate").mustSucceed(t).document(t, "component_validate")
	equipment, instruments := builtInFileCounts(t)
	if doc["equipment"] != float64(equipment) || doc["instruments"] != float64(instruments) {
		t.Fatalf("validate = %v, want %d equipment and %d instruments", doc, equipment, instruments)
	}
}

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

func TestComponentDirectoriesFromTheEnvironmentComeAfterTheFlags(t *testing.T) {
	t.Parallel()
	flagged := t.TempDir()
	environment := map[string]string{"STREAMSIM_COMPONENTS_DIRS": customerLibrary(t)}
	doc := invokeWithEnvironment(t, environment, "component", "validate", "--components-dir", flagged).
		mustSucceed(t).document(t, "component_validate")
	directories, _ := doc["directories"].([]any)
	builtIn, _ := builtInFileCounts(t)
	if len(directories) != 2 || directories[0] != flagged || directories[1] != environment["STREAMSIM_COMPONENTS_DIRS"] || doc["equipment"] != float64(builtIn+1) {
		t.Fatalf("validate = %v, want the flag directory then the environment directory", doc)
	}
}

func TestComponentEmptyDirectoryEntriesAreIgnored(t *testing.T) {
	t.Parallel()
	custom := customerLibrary(t)
	environment := map[string]string{"STREAMSIM_COMPONENTS_DIRS": custom + string(os.PathListSeparator)}
	doc := invokeWithEnvironment(t, environment, "component", "validate", "--components-dir", "").
		mustSucceed(t).document(t, "component_validate")
	directories, _ := doc["directories"].([]any)
	builtIn, _ := builtInFileCounts(t)
	if len(directories) != 1 || directories[0] != custom || doc["equipment"] != float64(builtIn+1) {
		t.Fatalf("validate = %v, want only the named directory", doc)
	}
}

func TestComponentEnvironmentListsSplitAtThePathListSeparator(t *testing.T) {
	t.Parallel()
	first, second := customerLibrary(t), t.TempDir()
	environment := map[string]string{"STREAMSIM_COMPONENTS_DIRS": first + string(os.PathListSeparator) + second}
	doc := invokeWithEnvironment(t, environment, "component", "validate").
		mustSucceed(t).document(t, "component_validate")
	directories, _ := doc["directories"].([]any)
	if len(directories) != 2 || directories[0] != first || directories[1] != second {
		t.Fatalf("directories = %v, want both environment entries", directories)
	}
}

func TestComponentCommandsRefuseWhatTheyCannotDo(t *testing.T) {
	t.Parallel()
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
		{"examples with an argument", []string{"component", "examples", "extra"}, "usage_invalid", 2},
		{"example without an id", []string{"component", "example"}, "usage_invalid", 2},
		{"an example that is not shipped", []string{"component", "example", "nothing"}, "component_failed", 1},
		{"a directory that does not exist", []string{"component", "list", "--components-dir", "no-such-dir"}, "file_not_found", 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := invoke(t, tc.args...)
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
	testsupport.WriteFileIn(t, filepath.Dir(path), filepath.Base(path), []byte(broken))
	failure := invoke(t, "component", "validate", "--components-dir", dir).failure(t)
	pointers := map[any]bool{}
	issues, _ := failure["issues"].([]any)
	for _, issue := range issues {
		pointers[issue.(map[string]any)["pointer"]] = true
	}
	if failure["code"] != "component_invalid" || failure["source"] != path || !pointers["/measurement_points/0/unit_family"] {
		t.Fatalf("failure = %v, want component_invalid at the unit family of %s", failure, path)
	}
}

func TestComponentNewCopiesOrEmptiesAPlantAndChecksItsID(t *testing.T) {
	t.Parallel()
	copied := invoke(t, "component", "new", "lab-copy", "--title", "Lab copy", "--from", testsupport.PilotComposition()).
		mustSucceed(t).document(t, "component_new")
	plant, _ := copied["composition"].(map[string]any)
	if plant["id"] != "lab-copy" || plant["title"] != "Lab copy" || plant["equipment"] == nil {
		t.Fatalf("copy = %v", copied)
	}
	emptied := invoke(t, "component", "new", "lab-empty", "--title", "Empty", "--example", "chiller-pilot", "--empty").
		mustSucceed(t).document(t, "component_new")
	plant, _ = emptied["composition"].(map[string]any)
	profiles, _ := plant["profiles"].([]any)
	if plant["equipment"] != nil || plant["connections"] != nil || len(profiles) == 0 {
		t.Fatalf("empty = %v", emptied)
	}
	for _, profile := range profiles {
		if entry, _ := profile.(map[string]any); entry["fault_weights"] != nil || entry["not_applicable"] != nil {
			t.Fatalf("empty plant kept %v", entry)
		}
	}
	refused := invoke(t, "component", "new", "Cool_Hall", "--title", "Cool hall", "--example", "chiller-pilot")
	if refused.code == 0 || !strings.Contains(refused.stdout, `"Cool_Hall\" is not a plant id`) {
		t.Fatalf("bad id: exit %d %s", refused.code, refused.stdout)
	}
	if usage := invoke(t, "component", "new", "lab", "--title", "Lab"); usage.code == 0 {
		t.Fatalf("no base accepted: %s", usage.stdout)
	}
}
