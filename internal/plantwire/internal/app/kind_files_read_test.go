package app

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	wire "github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/files"
)

const fanKindContent = `{"kind": "fan", "metrics": [], "commands": []}`

func TestReadKindFilesGathersFilesAndDirectoriesIntoOneSetByKindName(t *testing.T) {
	t.Parallel()
	pumps := kindsDirectoryWith(t, map[string]string{"pump.kind.json": pumpKindContent(t), "notes.txt": "not a kind file"})
	fans := kindsDirectoryWith(t, map[string]string{"fan.kind.json": fanKindContent})
	kinds, err := readKindFiles([]string{filepath.Join(fans, "fan.kind.json")}, []string{pumps})
	if err != nil || len(kinds) != 2 || kinds["pump"]["kind"] != "pump" || kinds["fan"]["kind"] != "fan" {
		t.Fatalf("readKindFiles = %v, %v, want the pump and fan kind files by name", kinds, err)
	}
}

func TestReadKindFilesRefusesWhatCannotBeReadAsKindFiles(t *testing.T) {
	t.Parallel()
	pumps := kindsDirectoryWith(t, map[string]string{"pump.kind.json": pumpKindContent(t)})
	thin := kindsDirectoryWith(t, map[string]string{"thin.kind.json": `{"kind": "thin"}`})
	var invalid *model.ValidationError
	if _, err := readKindFiles([]string{filepath.Join(pumps, "pump.kind.json")}, []string{pumps}); !errors.Is(err, wire.ErrKindFileDuplicate) {
		t.Fatalf("one kind in a file and a directory: err = %v, want ErrKindFileDuplicate", err)
	}
	if _, err := readKindFiles(nil, []string{thin}); !errors.As(err, &invalid) || invalid.Contract != model.ContractKindFile {
		t.Fatalf("a kind file that breaks its schema: err = %v, want a kind-file validation error", err)
	}
	if _, err := readKindFiles(nil, []string{filepath.Join(pumps, "absent")}); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a directory that does not exist: err = %v, want fs.ErrNotExist", err)
	}
	if _, err := readKindFiles([]string{filepath.Join(pumps, "absent.kind.json")}, nil); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a file that does not exist: err = %v, want fs.ErrNotExist", err)
	}
}

func TestReadKindFilesRefusesWhatIsNotARegularFileOfReasonableSize(t *testing.T) {
	t.Parallel()
	dir := kindsDirectoryWith(t, map[string]string{"pump.kind.json": pumpKindContent(t)})
	oversized := strings.Repeat(" ", files.MaxKindFileBytes) + fanKindContent
	fixtures := []struct {
		name, file string
		create     func(path string) error
		want       error
	}{
		{"a symbolic link named like a kind file", "link.kind.json", func(path string) error { return os.Symlink(filepath.Join(dir, "pump.kind.json"), path) }, kernel.ErrNotRegular},
		{"a named pipe named like a kind file, which must not hang the read", "pipe.kind.json", func(path string) error { return syscall.Mkfifo(path, 0o600) }, kernel.ErrNotRegular},
		{"a kind file larger than the limit", "big.kind.json", func(path string) error { return os.WriteFile(path, []byte(oversized), 0o600) }, kernel.ErrTooLarge},
	}
	for _, tc := range fixtures {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			folder := t.TempDir()
			if err := tc.create(filepath.Join(folder, tc.file)); err != nil {
				t.Fatal(err)
			}
			if _, err := readKindFiles(nil, []string{folder}); !errors.Is(err, tc.want) || !strings.Contains(err.Error(), tc.file) {
				t.Fatalf("readKindFiles = %v, want %v naming %s", err, tc.want, tc.file)
			}
		})
	}
}

func TestScaffoldChecksTheScaffoldedDomainAgainstTheKindFilesItWasGiven(t *testing.T) {
	t.Parallel()
	pumps := kindsDirectoryWith(t, map[string]string{"pump.kind.json": pumpKindContent(t)})
	drifted := kindsDirectoryWith(t, map[string]string{"pump.kind.json": strings.Replace(pumpKindContent(t), `"%"`, `"pct"`, 1)})
	cases := []struct {
		name string
		dir  string
		want error
	}{
		{"a twin that equals its kind file", pumps, nil},
		{"a twin whose unit drifted from its kind file", drifted, wire.ErrKindDrift},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			built, err := Scaffold(nil, []string{tc.dir}, "choices.json", func(kinds map[string]map[string]any) (string, *model.DomainSpec, error) {
				return "twin of " + kinds["pump"]["kind"].(string), pumpImplementing(t, "pump"), nil
			})
			if tc.want == nil && (err != nil || built != "twin of pump") {
				t.Fatalf("Scaffold = %q, %v, want the scaffolded twin", built, err)
			}
			var invalid *model.ValidationError
			if tc.want != nil && (!errors.Is(err, tc.want) || !errors.As(err, &invalid) || invalid.Source != "choices.json" || built != "") {
				t.Fatalf("Scaffold = %q, %v, want %v stamped with the choices file and no twin", built, err, tc.want)
			}
		})
	}
}

func TestScaffoldStopsAtTheFirstThingItCannotDo(t *testing.T) {
	t.Parallel()
	pumps := kindsDirectoryWith(t, map[string]string{"pump.kind.json": pumpKindContent(t)})
	refused := errors.New("the choices do not fit the kind files")
	called := false
	scaffold := func(map[string]map[string]any) (int, *model.DomainSpec, error) {
		called = true
		return 0, nil, refused
	}
	if _, err := Scaffold(nil, []string{filepath.Join(pumps, "absent")}, "choices.json", scaffold); !errors.Is(err, fs.ErrNotExist) || called {
		t.Fatalf("Scaffold over a missing directory = %v (scaffold called: %v), want fs.ErrNotExist before any scaffolding", err, called)
	}
	if _, err := Scaffold(nil, []string{pumps}, "choices.json", scaffold); !errors.Is(err, refused) || !called {
		t.Fatalf("Scaffold = %v (scaffold called: %v), want the scaffold's own refusal", err, called)
	}
}
