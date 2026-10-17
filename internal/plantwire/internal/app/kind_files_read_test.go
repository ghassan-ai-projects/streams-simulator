package app

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	wire "github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

const fanKindContent = `{"kind": "fan", "metrics": [], "commands": []}`

func TestReadKindFilesGathersFilesAndDirectoriesIntoOneSetByKindName(t *testing.T) {
	t.Parallel()
	pumps := kindsDirectoryWith(t, map[string]string{"pump.kind.json": pumpKindContent(t), "notes.txt": "not a kind file"})
	fans := kindsDirectoryWith(t, map[string]string{"fan.kind.json": fanKindContent})
	kinds, err := ReadKindFiles([]string{filepath.Join(fans, "fan.kind.json")}, []string{pumps})
	if err != nil || len(kinds) != 2 || kinds["pump"]["kind"] != "pump" || kinds["fan"]["kind"] != "fan" {
		t.Fatalf("ReadKindFiles = %v, %v, want the pump and fan kind files by name", kinds, err)
	}
}

func TestReadKindFilesRefusesWhatCannotBeReadAsKindFiles(t *testing.T) {
	t.Parallel()
	pumps := kindsDirectoryWith(t, map[string]string{"pump.kind.json": pumpKindContent(t)})
	thin := kindsDirectoryWith(t, map[string]string{"thin.kind.json": `{"kind": "thin"}`})
	var invalid *model.ValidationError
	if _, err := ReadKindFiles([]string{filepath.Join(pumps, "pump.kind.json")}, []string{pumps}); !errors.Is(err, wire.ErrKindFileDuplicate) {
		t.Fatalf("one kind in a file and a directory: err = %v, want ErrKindFileDuplicate", err)
	}
	if _, err := ReadKindFiles(nil, []string{thin}); !errors.As(err, &invalid) || invalid.Contract != model.ContractKindFile {
		t.Fatalf("a kind file that breaks its schema: err = %v, want a kind-file validation error", err)
	}
	if _, err := ReadKindFiles(nil, []string{filepath.Join(pumps, "absent")}); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a directory that does not exist: err = %v, want fs.ErrNotExist", err)
	}
	if _, err := ReadKindFiles([]string{filepath.Join(pumps, "absent.kind.json")}, nil); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a file that does not exist: err = %v, want fs.ErrNotExist", err)
	}
}

func TestCheckImplementsKindsComparesTheDomainWithTheKindFilesAlreadyRead(t *testing.T) {
	t.Parallel()
	pumps := kindsDirectoryWith(t, map[string]string{"pump.kind.json": pumpKindContent(t)})
	drifted := kindsDirectoryWith(t, map[string]string{"pump.kind.json": strings.Replace(pumpKindContent(t), `"%"`, `"pct"`, 1)})
	cases := []struct {
		name string
		dir  string
		want error
	}{
		{"a kind file that matches", pumps, nil},
		{"a kind file whose unit drifted", drifted, wire.ErrKindDrift},
		{"a directory without the kind file", kindsDirectoryWith(t, nil), wire.ErrKindFileMissing},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			kinds, err := ReadKindFiles(nil, []string{tc.dir})
			if err != nil {
				t.Fatal(err)
			}
			got := CheckImplementsKinds(pumpImplementing(t, "pump"), kinds)
			if (tc.want == nil && got != nil) || (tc.want != nil && !errors.Is(got, tc.want)) {
				t.Fatalf("CheckImplementsKinds = %v, want %v", got, tc.want)
			}
		})
	}
}
