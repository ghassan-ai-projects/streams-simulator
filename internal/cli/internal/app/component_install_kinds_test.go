package app

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

var beltKindFiles = []string{"belt_conveyor.kind.json", "bulk_feeder.kind.json", "conveyor_drive.kind.json", "feeder-controller.board.json"}

func kindFileNamesOf(t *testing.T, doc map[string]any) []string {
	t.Helper()
	listed, present := doc["kind_files"].([]any)
	if !present {
		t.Fatalf("no kind_files in %v", doc)
	}
	names := []string{}
	for _, entry := range listed {
		fields, _ := entry.(map[string]any)
		if _, isObject := fields["document"].(map[string]any); !isObject {
			t.Fatalf("kind file %v has no document", fields["name"])
		}
		names = append(names, fields["name"].(string))
	}
	return names
}

func stringsOf(value any) []string {
	listed, _ := value.([]any)
	out := []string{}
	for _, item := range listed {
		out = append(out, item.(string))
	}
	return out
}

func TestAnExampleAndANewPlantFromItCarryTheKindFilesItsPlantImplements(t *testing.T) {
	t.Parallel()
	belt := invoke(t, "component", "example", "belt-conveyor-drives-a").mustSucceed(t).document(t, "component_example")
	if names := kindFileNamesOf(t, belt); !slices.Equal(names, beltKindFiles) {
		t.Fatalf("belt kind files = %v", names)
	}
	pilot := invoke(t, "component", "example", "chiller-pilot").mustSucceed(t).document(t, "component_example")
	if names := kindFileNamesOf(t, pilot); len(names) != 0 {
		t.Fatalf("pilot kind files = %v", names)
	}
	made := invoke(t, "component", "new", "belt", "--title", "Belt", "--example", "belt-conveyor-drives-b").mustSucceed(t).document(t, "component_new")
	if names := kindFileNamesOf(t, made); !slices.Equal(names, beltKindFiles) {
		t.Fatalf("new plant kind files = %v", names)
	}
	copied := invoke(t, "component", "new", "lab", "--title", "Lab", "--from", testsupport.BeltConveyorComposition("a")).mustSucceed(t).document(t, "component_new")
	if _, present := copied["kind_files"]; present {
		t.Fatalf("a copy of a file carries kind files: %v", copied)
	}
}

func TestInstallKindsWritesAnExamplesKindFilesSoItsPlantComposes(t *testing.T) {
	t.Parallel()
	kinds := filepath.Join(t.TempDir(), "kinds")
	first := invoke(t, "component", "install-kinds", "belt-conveyor-drives-a", "--kinds-dir", kinds).mustSucceed(t).document(t, "component_install_kinds")
	if first["example"] != "belt-conveyor-drives-a" || !slices.Equal(stringsOf(first["written"]), beltKindFiles) || len(stringsOf(first["unchanged"])) != 0 {
		t.Fatalf("first install = %v", first)
	}
	out := filepath.Join(t.TempDir(), "a.domain.json")
	invoke(t, "domain", "compose", testsupport.BeltConveyorComposition("a"), "--out", out, "--kinds-dir", kinds).mustSucceed(t)
	again := invoke(t, "component", "install-kinds", "--kinds-dir", kinds, "belt-conveyor-drives-b").mustSucceed(t).document(t, "component_install_kinds")
	if len(stringsOf(again["written"])) != 0 || !slices.Equal(stringsOf(again["unchanged"]), beltKindFiles) {
		t.Fatalf("second install = %v", again)
	}
}

func TestInstallKindsRefusesAKindsFolderWithADifferentKindFile(t *testing.T) {
	t.Parallel()
	kinds := t.TempDir()
	testsupport.WriteFileIn(t, kinds, "bulk_feeder.kind.json", []byte(`{"kind":"bulk_feeder","metrics":[],"commands":[]}`))
	got := invoke(t, "component", "install-kinds", "belt-conveyor-drives-a", "--kinds-dir", kinds)
	failure := got.failure(t)
	if issue := firstIssue(failure); got.code != 1 || failure["code"] != "kind_failed" || issue["keyword"] != "kind_file_differs" || issue["pointer"] != "/bulk_feeder.kind.json" {
		t.Fatalf("install over a different kind file = %d %v", got.code, failure)
	}
	if entries, _ := os.ReadDir(kinds); len(entries) != 1 {
		t.Fatalf("kinds folder = %v, want only the own file", entries)
	}
}

func TestInstallKindsRefusesWhatItCannotDo(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		args []string
		code string
		exit int
	}{
		{"no kinds folder", []string{"component", "install-kinds", "belt-conveyor-drives-a"}, "usage_invalid", 2},
		{"no example", []string{"component", "install-kinds", "--kinds-dir", "kinds"}, "usage_invalid", 2},
		{"an unknown flag", []string{"component", "install-kinds", "x", "--kinds-dir", "kinds", "--force"}, "usage_invalid", 2},
		{"an example that is not shipped", []string{"component", "install-kinds", "nothing", "--kinds-dir", "kinds"}, "component_failed", 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := invoke(t, tc.args...)
			if code := got.failure(t)["code"]; got.code != tc.exit || code != tc.code {
				t.Fatalf("exit %d code %v, want %d %s", got.code, code, tc.exit, tc.code)
			}
		})
	}
}
