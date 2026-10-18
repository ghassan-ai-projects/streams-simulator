package contract

import (
	"maps"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
)

func kindEntrySchema(t *testing.T) map[string]any {
	t.Helper()
	nbirth := readJSON(t, filepath.Join(plantWireDir(), "schemas", "NBIRTH.schema.json"))
	return member(nbirth, "properties", "kinds", "items").(map[string]any)
}

func TestTheKindFileAndTheBirthShareOneKindDefinition(t *testing.T) {
	t.Parallel()
	entry := kindEntrySchema(t)
	kindFile := readJSON(t, filepath.Join(plantWireDir(), "schemas", "KIND.schema.json")).(map[string]any)
	if !slices.Contains(entry["required"].([]any), any("kind_digest")) {
		t.Fatal("NBIRTH.kinds entries no longer require kind_digest")
	}
	properties := maps.Clone(entry["properties"].(map[string]any))
	delete(properties, "kind_digest")
	filed := maps.Clone(kindFile["properties"].(map[string]any))
	if forbidden, _ := filed["kind_digest"].(map[string]any); forbidden["not"] == nil {
		t.Error("the kind file schema no longer forbids kind_digest")
	}
	delete(filed, "kind_digest")
	required := slices.DeleteFunc(slices.Clone(entry["required"].([]any)), func(name any) bool { return name == "kind_digest" })
	if !reflect.DeepEqual(properties, filed) || !reflect.DeepEqual(required, kindFile["required"]) || entry["type"] != kindFile["type"] {
		t.Error("the kind file schema and the NBIRTH.kinds entry schema have diverged: generate both from one definition")
	}
}

func TestTheKindFixtureIsTheBirthEntryWithoutItsDigestAndTheDigestIsRight(t *testing.T) {
	t.Parallel()
	birth := readJSON(t, filepath.Join(plantWireDir(), "fixtures", "valid", "NBIRTH.json"))
	kindFile := readJSON(t, filepath.Join(plantWireDir(), "fixtures", "valid", "KIND.json")).(map[string]any)
	for _, kind := range member(birth, "kinds").([]any) {
		entry := maps.Clone(kind.(map[string]any))
		declared := entry["kind_digest"]
		delete(entry, "kind_digest")
		got, err := canonical.Digest(entry)
		if err != nil || got != declared {
			t.Errorf("kind %v announces %v, its content gives %s (%v)", entry["kind"], declared, got, err)
		}
		if entry["kind"] == kindFile["kind"] && !reflect.DeepEqual(entry, kindFile) {
			t.Error("valid/KIND.json differs from the NBIRTH fixture's entry of that kind")
		}
	}
}
