package contract

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

func TestEveryBoardNamesItsEnvironmentAndOnlyTheNodeCanBeMixed(t *testing.T) {
	t.Parallel()
	schema := readJSON(t, filepath.Join(plantWireDir(), "schemas", "NBIRTH.schema.json"))
	if !slices.Contains(member(schema, "properties", "boards", "items", "required").([]any), any("environment")) {
		t.Error("NBIRTH no longer requires the environment of every board")
	}
	boardValues := member(schema, "properties", "boards", "items", "properties", "environment", "enum").([]any)
	nodeValues := member(schema, "properties", "environment", "enum").([]any)
	for _, value := range []any{"simulated", "hardware"} {
		if !slices.Contains(boardValues, value) || !slices.Contains(nodeValues, value) {
			t.Errorf("%v is no longer an environment of a board and of a node", value)
		}
	}
	if slices.Contains(boardValues, any("mixed")) || !slices.Contains(nodeValues, any("mixed")) {
		t.Errorf("a board is simulated or hardware and a node may be mixed, got boards %v, node %v", boardValues, nodeValues)
	}
}

func TestTheFixturesAndTheSpecSpeakTheSameContractVersion(t *testing.T) {
	t.Parallel()
	version, _ := member(readJSON(t, filepath.Join(plantWireDir(), "fixtures", "valid", "NBIRTH.json")), "contract_version").(string)
	spec, err := os.ReadFile(filepath.Join(plantWireDir(), "SPEC.md"))
	if err != nil {
		t.Fatal(err)
	}
	if version == "" || !strings.Contains(string(spec), "contract v"+version+",") {
		t.Errorf("the NBIRTH fixture announces contract %q, the SPEC status line does not say v%s", version, version)
	}
	for _, file := range fixtures(t, "exchanges", "*.jsonl") {
		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for _, announced := range announcedVersion.FindAllStringSubmatch(string(raw), -1) {
			if announced[1] != version {
				t.Errorf("%s announces contract %s, the NBIRTH fixture %s", filepath.Base(file), announced[1], version)
			}
		}
	}
}

var announcedVersion = regexp.MustCompile(`"contract_version":\s*"([^"]*)"`)
