package contract

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestEveryBoardRuleFixturePassesTheSchemaItGoesBeyond(t *testing.T) {
	t.Parallel()
	schemas := compiledSchemas(t)
	for _, file := range fixtures(t, "rules", "*.json") {
		fixture := readJSON(t, file)
		if member(fixture, "keyword") == "" || member(fixture, "fails_at") == nil {
			t.Errorf("%s: a rule fixture names its keyword and the member where it fails", filepath.Base(file))
		}
		if errs := schemas[member(fixture, "schema").(string)].Validate(member(fixture, "document")); len(errs) > 0 {
			t.Errorf("%s: a rule fixture must pass the schema, so that only the rule refuses it: %v", filepath.Base(file), errs)
		}
	}
}

var closedBoardObjects = [][]string{
	{},
	{"properties", "kinds", "additionalProperties"},
	{"properties", "kinds", "additionalProperties", "properties", "commands", "additionalProperties"},
	{"properties", "kinds", "additionalProperties", "properties", "commands", "additionalProperties", "properties", "lease"},
	{"properties", "kinds", "additionalProperties", "properties", "commands", "additionalProperties", "properties", "args", "additionalProperties"},
	{"properties", "kinds", "additionalProperties", "properties", "metrics", "additionalProperties"},
}

func TestTheBoardMappingIsClosedAndKeepsItsLeaseWhole(t *testing.T) {
	t.Parallel()
	board := readJSON(t, filepath.Join(plantWireDir(), "schemas", "BOARD.schema.json"))
	for _, path := range closedBoardObjects {
		if member(board, append(slices.Clone(path), "additionalProperties")...) != false {
			t.Errorf("BOARD %v no longer refuses unknown members: a misspelled lease would be dropped silently", path)
		}
	}
	lease := closedBoardObjects[3]
	required, _ := member(board, append(slices.Clone(lease), "required")...).([]any)
	if !slices.Contains(required, any("arg")) || !slices.Contains(required, any("ms")) {
		t.Errorf("a board lease no longer requires its raw argument and its duration: %v", required)
	}
	if member(board, "properties", "board_digest", "not") == nil {
		t.Error("the board mapping schema no longer forbids board_digest")
	}
}

func TestTheVersionFileTheFixturesAndTheSpecSpeakOneContractVersion(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(filepath.Join(plantWireDir(), "VERSION"))
	if err != nil {
		t.Fatal(err)
	}
	version, _ := member(readJSON(t, filepath.Join(plantWireDir(), "fixtures", "valid", "NBIRTH.json")), "contract_version").(string)
	if strings.TrimSpace(string(raw)) != version || string(raw) != version+"\n" {
		t.Errorf("VERSION holds %q, the NBIRTH fixture announces %q: one version, one line", raw, version)
	}
}
