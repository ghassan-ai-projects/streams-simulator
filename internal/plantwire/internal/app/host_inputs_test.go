package app

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/host"
)

const oneScriptedCommand = `[{"at_ns": 1767225660000000000, "device": "P1", "command": "start", "args": {}}]`

func TestLoadHostReadsOnlyTheNamedInputs(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	script := filepath.Join(dir, "script.json")
	broken := filepath.Join(dir, "broken.json")
	writeHostInput(t, script, oneScriptedCommand)
	writeHostInput(t, broken, "not json")
	cases := map[string]struct {
		script, inventory string
		scripted          int
		refused           error
	}{
		"no inputs":          {},
		"a valid script":     {script: script, scripted: 1},
		"a malformed script": {script: broken, refused: ErrHostInput},
		"a missing script":   {script: filepath.Join(dir, "absent.json"), refused: fs.ErrNotExist},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			loaded, err := LoadHost(host.Options{}, tc.script, tc.inventory)
			if tc.refused == nil && (err != nil || len(loaded.Script) != tc.scripted) {
				t.Fatalf("LoadHost = %d scripted, %v, want %d and no error", len(loaded.Script), err, tc.scripted)
			}
			if tc.refused != nil && !errors.Is(err, tc.refused) {
				t.Fatalf("LoadHost error = %v, want %v", err, tc.refused)
			}
		})
	}
}

func writeHostInput(t *testing.T, path, text string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestTheHostScriptAndInventoryAreParsedStrictly(t *testing.T) {
	t.Parallel()
	script, err := host.ParseScript([]byte(`[{"at_ns":5,"device":"P1","command":"start","args":{}}]`))
	if err != nil || len(script) != 1 || script[0].Device != "P1" {
		t.Fatalf("ParseScript = %+v, %v", script, err)
	}
	if _, err := host.ParseScript([]byte(`[{"when":5}]`)); err == nil || !strings.Contains(err.Error(), "plantwire host script") {
		t.Fatalf("ParseScript with an unknown field = %v", err)
	}
	inventory, err := host.ParseInventory([]byte(`{"P1":"pump"}`))
	if err != nil || inventory["P1"] != "pump" {
		t.Fatalf("ParseInventory = %v, %v", inventory, err)
	}
	if _, err := host.ParseInventory([]byte(`[]`)); err == nil || !strings.Contains(err.Error(), "plantwire host inventory") {
		t.Fatalf("ParseInventory of an array = %v", err)
	}
}
