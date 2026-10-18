package plantwire

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func kindsFolder(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		testsupport.WriteFileIn(t, dir, name, []byte(content))
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

func TestScaffoldHandsTheKindFilesToTheScaffoldAndChecksWhatItBuilt(t *testing.T) {
	t.Parallel()
	compiled, err := domain.Load(testsupport.PumpBoardDomain())
	if err != nil {
		t.Fatal(err)
	}
	compiled.Spec.Kinds[0].Implements = "pump"
	drifted := strings.Replace(pumpKindContent(t), `"%"`, `"pct"`, 1)
	cases := []struct {
		name  string
		paths []string
		dirs  []string
		fails error
	}{
		{"a kind file that matches", nil, []string{kindsFolder(t, map[string]string{"pump.kind.json": pumpKindContent(t)})}, nil},
		{"a kind file whose unit drifted", nil, []string{kindsFolder(t, map[string]string{"pump.kind.json": drifted})}, ErrKindDrift},
		{"a folder without the kind file", nil, []string{kindsFolder(t, nil)}, ErrKindFileMissing},
		{"a folder that does not exist", nil, []string{filepath.Join(t.TempDir(), "absent")}, fs.ErrNotExist},
		{"a file that does not exist", []string{filepath.Join(t.TempDir(), "absent.kind.json")}, nil, fs.ErrNotExist},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			built, err := Scaffold(tc.paths, tc.dirs, "choices.json", func(map[string]map[string]any) (string, *model.DomainSpec, error) {
				return "twin", compiled.Spec, nil
			})
			if tc.fails == nil {
				if err != nil || built != "twin" {
					t.Fatalf("Scaffold = %q, %v, want the twin handed back", built, err)
				}
				return
			}
			if !errors.Is(err, tc.fails) || built != "" {
				t.Fatalf("Scaffold = %q, %v, want %v and no twin", built, err, tc.fails)
			}
		})
	}
}
