package component

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func checkGolden(t *testing.T, document any, golden string) {
	t.Helper()
	got, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	got = append(got, '\n')
	path := filepath.Join("testdata", golden)
	if os.Getenv("UPDATE_GOLDEN") != "" {
		if err := os.WriteFile(path, got, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if want := readFile(t, path); !bytes.Equal(got, want) {
		t.Fatalf("the document differs from %s:\n%s", golden, got)
	}
}

func thermalTwinInput(t *testing.T, library *Library) AssumptionsInput {
	t.Helper()
	choices, kinds := readFile(t, testsupport.TwinThermalChoices()), thermalKinds(t)
	twin, err := Twin(library, kinds, choices, "thermal-board.twin.json")
	if err != nil {
		t.Fatal(err)
	}
	return AssumptionsInput{Domain: twin.Bytes, Source: "arduino-thermal-twin.domain.json", Twin: &TwinFiles{Choices: choices, Source: "thermal-board.twin.json", Kinds: kinds}}
}

func TestAssumptionsOfTheShippedThermalTwinMatchTheirGolden(t *testing.T) {
	t.Parallel()
	library := builtInLibrary(t)
	got, err := Assumptions(library, thermalTwinInput(t, library))
	if err != nil {
		t.Fatal(err)
	}
	if got.Origin.Kind != "twin" || got.Origin.Twin.Status != "in_sync" {
		t.Fatalf("origin = %+v, want the twin in sync", got.Origin)
	}
	checkGolden(t, got, "assumptions-thermal-twin.golden.json")
}

func TestAssumptionsOfTheShippedChillerPilotNameEveryInstrumentItsCompositionRecorded(t *testing.T) {
	t.Parallel()
	got, err := Assumptions(builtInLibrary(t), AssumptionsInput{Domain: readFile(t, filepath.Join(testsupport.RepositoryRoot(), "examples", "chiller-pilot.domain.json")), Source: "chiller-pilot.domain.json"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Origin.Kind != "composition" || len(got.Kinds) == 0 {
		t.Fatalf("origin = %+v with %d kinds, want a composition", got.Origin, len(got.Kinds))
	}
	checkGolden(t, got, "assumptions-chiller-pilot.golden.json")
}

func TestAssumptionsNeedALibrary(t *testing.T) {
	t.Parallel()
	for name, library := range map[string]*Library{"no library": nil, "an empty library": {}} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := Assumptions(library, AssumptionsInput{}); !errors.Is(err, ErrNoLibrary) {
				t.Fatalf("Assumptions = %v, want ErrNoLibrary", err)
			}
		})
	}
}

func bytesReplace(t *testing.T, raw []byte, old, replacement string) []byte {
	t.Helper()
	if !bytes.Contains(raw, []byte(old)) {
		t.Fatalf("%q is not in the file", old)
	}
	return bytes.Replace(raw, []byte(old), []byte(replacement), 1)
}
