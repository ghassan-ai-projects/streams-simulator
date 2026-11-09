package files

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/component/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestInstallingAnExamplesKindFilesWritesThemOnceAndThenFindsThemUnchanged(t *testing.T) {
	t.Parallel()
	kinds := filepath.Join(t.TempDir(), "kinds")
	first, err := InstallExampleKinds("belt-conveyor-drives-a", kinds)
	want := []string{"belt_conveyor.kind.json", "bulk_feeder.kind.json", "conveyor_drive.kind.json", "feeder-controller.board.json"}
	if err != nil || !slices.Equal(first.Written, want) || len(first.Unchanged) != 0 {
		t.Fatalf("first install = %+v, %v", first, err)
	}
	example, err := FindExample("belt-conveyor-drives-a")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range example.KindFiles {
		written, err := os.ReadFile(filepath.Join(kinds, file.Name))
		if err != nil || string(written) != string(file.Document) {
			t.Fatalf("%s written as %q, %v", file.Name, written, err)
		}
	}
	again, err := InstallExampleKinds("belt-conveyor-drives-b", kinds)
	if err != nil || len(again.Written) != 0 || !slices.Equal(again.Unchanged, want) {
		t.Fatalf("second install = %+v, %v", again, err)
	}
}

func TestInstallingAnExampleWithoutKindFilesCreatesNoKindsFolder(t *testing.T) {
	t.Parallel()
	kinds := filepath.Join(t.TempDir(), "kinds")
	installed, err := InstallExampleKinds("chiller-pilot", kinds)
	if err != nil || len(installed.Written) != 0 || installed.Example != "chiller-pilot" {
		t.Fatalf("install = %+v, %v", installed, err)
	}
	if _, err := os.Stat(kinds); !os.IsNotExist(err) {
		t.Fatalf("kinds folder created: %v", err)
	}
}

func TestInstallingOverADifferentKindFileLeavesTheFolderAsItIs(t *testing.T) {
	t.Parallel()
	kinds := t.TempDir()
	own := filepath.Join(kinds, "bulk_feeder.kind.json")
	if err := os.WriteFile(own, []byte(`{"kind":"bulk_feeder","metrics":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := InstallExampleKinds("belt-conveyor-drives-a", kinds)
	if _, ok := model.AsValidationError(err); !ok || !strings.Contains(err.Error(), "bulk_feeder.kind.json") {
		t.Fatalf("install error = %v", err)
	}
	if names, _ := os.ReadDir(kinds); len(names) != 1 {
		t.Fatalf("kinds folder = %v, want only the own file", names)
	}
}

func TestInstallingKindFilesReportsWhatTheFileSystemRefuses(t *testing.T) {
	t.Parallel()
	notAFolder := filepath.Join(t.TempDir(), "kinds")
	if err := os.WriteFile(notAFolder, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := InstallExampleKinds("belt-conveyor-drives-a", notAFolder); err == nil || !strings.Contains(err.Error(), "component: read") {
		t.Fatalf("a file as kinds folder = %v", err)
	}
	if _, err := InstallExampleKinds("nothing", t.TempDir()); !errors.Is(err, domain.ErrUnknownExample) {
		t.Fatalf("an unknown example = %v", err)
	}
	if err := writeKindFiles(filepath.Join(notAFolder, "sub"), []domain.KindFile{{Name: "a.kind.json"}}); err == nil || !strings.Contains(err.Error(), "kinds folder") {
		t.Fatalf("an uncreatable kinds folder = %v", err)
	}
	locked := t.TempDir()
	if err := os.Mkdir(filepath.Join(locked, "a.kind.json"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := writeKindFiles(locked, []domain.KindFile{{Name: "a.kind.json", Document: []byte("{}")}}); err == nil || !strings.Contains(err.Error(), "component: write") {
		t.Fatalf("an unwritable kinds folder = %v", err)
	}
}
