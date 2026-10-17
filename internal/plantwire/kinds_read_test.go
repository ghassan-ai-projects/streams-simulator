package plantwire

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	wire "github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const emptyFanKind = `{"kind": "fan", "metrics": [], "commands": []}`

func kindsFolder(t *testing.T, files map[string]string) string {
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

func TestReadKindFilesGathersFilesAndFoldersIntoOneSetByKindName(t *testing.T) {
	t.Parallel()
	pumps := kindsFolder(t, map[string]string{"pump.kind.json": pumpKindContent(t), "notes.txt": "not a kind file"})
	fans := kindsFolder(t, map[string]string{"fan.kind.json": emptyFanKind})
	thin := kindsFolder(t, map[string]string{"thin.kind.json": `{"kind": "thin"}`})
	cases := []struct {
		name  string
		paths []string
		dirs  []string
		want  []string
		fails error
	}{
		{"a folder", nil, []string{pumps}, []string{"pump"}, nil},
		{"a file and a folder", []string{filepath.Join(fans, "fan.kind.json")}, []string{pumps}, []string{"fan", "pump"}, nil},
		{"nothing at all", nil, nil, nil, nil},
		{"one kind in a file and a folder", []string{filepath.Join(pumps, "pump.kind.json")}, []string{pumps}, nil, wire.ErrKindFileDuplicate},
		{"a folder that does not exist", nil, []string{filepath.Join(pumps, "absent")}, nil, fs.ErrNotExist},
		{"a file that does not exist", []string{filepath.Join(pumps, "absent.kind.json")}, nil, nil, fs.ErrNotExist},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			kinds, err := ReadKindFiles(tc.paths, tc.dirs)
			names := make([]string, 0, len(kinds))
			for name := range kinds {
				names = append(names, name)
			}
			slices.Sort(names)
			if tc.fails != nil && !errors.Is(err, tc.fails) {
				t.Fatalf("ReadKindFiles = %v, want %v", err, tc.fails)
			}
			if tc.fails == nil && (err != nil || !slices.Equal(names, tc.want)) {
				t.Fatalf("ReadKindFiles = %v, %v, want the kinds %v", names, err, tc.want)
			}
		})
	}
	var invalid *model.ValidationError
	if _, err := ReadKindFiles(nil, []string{thin}); !errors.As(err, &invalid) || invalid.Contract != model.ContractKindFile {
		t.Fatalf("a kind file that breaks its schema: err = %v, want a kind-file validation error", err)
	}
}

func TestCheckImplementsKindsComparesTheDomainWithKindFilesAlreadyRead(t *testing.T) {
	t.Parallel()
	compiled, err := domain.Load(testsupport.PumpBoardDomain())
	if err != nil {
		t.Fatal(err)
	}
	compiled.Spec.Kinds[0].Implements = "pump"
	cases := []struct {
		name  string
		dir   string
		fails error
	}{
		{"a kind file that matches", kindsFolder(t, map[string]string{"pump.kind.json": pumpKindContent(t)}), nil},
		{"a kind file whose unit drifted", kindsFolder(t, map[string]string{"pump.kind.json": strings.Replace(pumpKindContent(t), `"%"`, `"pct"`, 1)}), wire.ErrKindDrift},
		{"a folder without the kind file", kindsFolder(t, nil), wire.ErrKindFileMissing},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			kinds, err := ReadKindFiles(nil, []string{tc.dir})
			if err != nil {
				t.Fatal(err)
			}
			got := CheckImplementsKinds(compiled.Spec, kinds)
			if (tc.fails == nil && got != nil) || (tc.fails != nil && !errors.Is(got, tc.fails)) {
				t.Fatalf("CheckImplementsKinds = %v, want %v", got, tc.fails)
			}
		})
	}
}
