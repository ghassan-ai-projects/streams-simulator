package domain

import (
	"strings"
	"testing"
)

func customerFiles(equipment, instruments []File, vocabulary string) testFiles {
	return testFiles{vocabulary: vocabulary, equipment: equipment, instruments: instruments}
}

func builtIn(t *testing.T) Directory { t.Helper(); return baseFiles().directory("builtin") }

func TestLoadRefusesTheSameReferenceTwiceWhateverTheContent(t *testing.T) {
	t.Parallel()
	same := customerFiles([]File{file("equipment/kiln.component.json", kilnJSON)}, nil, "")
	changed := mutate(t, kilnJSON, func(doc map[string]any) { doc["title"] = "Other title" })
	other := customerFiles([]File{file("equipment/kiln.component.json", changed)}, nil, "")
	for name, files := range map[string]testFiles{"identical content": same, "different content": other} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := LoadLibrary([]Directory{builtIn(t), files.directory("customer")})
			wantIssue(t, err, "component-v0.1", "/id", "acme/kiln@1 is already defined by builtin/equipment/kiln.component.json")
		})
	}
}

func TestLoadRefusesTheSameReferenceTwiceInOneDirectory(t *testing.T) {
	t.Parallel()
	files := baseFiles()
	files.equipment = append(files.equipment, file("equipment/kiln.copy.component.json", kilnJSON))
	_, err := LoadLibrary([]Directory{files.directory("lib")})
	wantIssue(t, err, "component-v0.1", "/id", "already defined by lib/equipment/kiln.component.json")
}

func TestLoadAllowsTwoMajorVersionsOfOneComponent(t *testing.T) {
	t.Parallel()
	second := mutate(t, kilnJSON, func(doc map[string]any) { doc["version"] = 2 })
	files := baseFiles()
	files.equipment = append(files.equipment, file("equipment/kiln-2.component.json", second))
	library := loadLibrary(t, files.directory("lib"))
	if _, present := library.equipment["acme/kiln@2"]; !present {
		t.Fatal("major 2 was not admitted beside major 1")
	}
}

func TestLoadRefusesTheStreamsimNamespaceOutsideTheBuiltInDirectory(t *testing.T) {
	t.Parallel()
	equipment := mutate(t, kilnJSON, func(doc map[string]any) { doc["id"] = "streamsim/kiln" })
	instrument := mutate(t, probeJSON, func(doc map[string]any) { doc["id"] = "streamsim/probe" })
	cases := map[string]testFiles{
		"equipment":  customerFiles([]File{file("equipment/kiln.component.json", equipment)}, nil, ""),
		"instrument": customerFiles(nil, []File{file("instruments/probe.instrument.json", instrument)}, ""),
	}
	for name, files := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := LoadLibrary([]Directory{builtIn(t), files.directory("customer")})
			if err == nil || !strings.Contains(err.Error(), `reserved "streamsim/" namespace`) {
				t.Fatalf("error = %v, want the reserved namespace refusal", err)
			}
		})
	}
}

func TestLoadAllowsTheStreamsimNamespaceInTheBuiltInDirectory(t *testing.T) {
	t.Parallel()
	equipment := mutate(t, kilnJSON, func(doc map[string]any) { doc["id"] = "streamsim/kiln" })
	files := baseFiles()
	files.equipment = []File{file("builtin/equipment/kiln.component.json", equipment)}
	loadLibrary(t, files.directory("builtin"))
}

func TestLoadLetsACustomerLibraryUseTheBuiltInInstruments(t *testing.T) {
	t.Parallel()
	custom := mutate(t, kilnJSON, func(doc map[string]any) { doc["id"] = "customer/furnace" })
	files := customerFiles([]File{file("equipment/furnace.component.json", custom)}, nil, "")
	library := loadLibrary(t, builtIn(t), files.directory("customer"))
	if got := len(library.List()); got != 4 {
		t.Fatalf("library holds %d entries, want the built-in three and one more equipment", got)
	}
}

func TestLoadChecksEveryInstrumentAgainstTheVocabulary(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, pointer, message string
		change                 func(doc map[string]any)
	}{
		{"a family the vocabulary lacks", "/unit_family", `unit family "gizmo" is not in the vocabulary`, func(doc map[string]any) { doc["unit_family"] = "gizmo" }},
		{"a unit outside the family", "/unit", `unit "kW" is not a unit of family "temperature"`, func(doc map[string]any) { doc["unit"] = "kW" }},
		{"a tag the vocabulary lacks", "/tag", `tag "ZZ" is not in the vocabulary`, func(doc map[string]any) { doc["tag"] = "ZZ" }},
		{"a tag that does not measure the family", "/tag", `tag "JT" does not measure unit family "temperature"`, func(doc map[string]any) { doc["tag"] = "JT" }},
		{"an inverted range", "/range", "must be below max", func(doc map[string]any) { doc["range"] = map[string]any{"min": 9, "max": 1} }},
		{"a fault declared twice", "/faults/1/id", `fault "offset" is declared twice`, func(doc map[string]any) {
			object(t, doc)["faults"].([]any)[1].(map[string]any)["id"] = "offset"
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			files := baseFiles()
			spare := mutate(t, probeJSON, func(doc map[string]any) { doc["id"] = "acme/spare"; tc.change(doc) })
			files.instruments = append(files.instruments, file("instruments/spare.instrument.json", spare))
			_, err := LoadLibrary([]Directory{files.directory("lib")})
			wantIssue(t, err, "instrument-v0.1", tc.pointer, tc.message)
		})
	}
}

func TestLoadRefusesAnInstrumentFaultThatLacksWhatItsFormNeeds(t *testing.T) {
	t.Parallel()
	for _, fault := range []map[string]any{
		{"id": "drift", "form": "drift", "magnitude": 1},
		{"id": "stuck", "form": "stuck"},
		{"id": "scale", "form": "scale"},
	} {
		raw := mutate(t, probeJSON, func(doc map[string]any) { doc["id"] = "acme/spare"; doc["faults"] = []any{fault} })
		files := baseFiles()
		files.instruments = append(files.instruments, file("instruments/spare.instrument.json", raw))
		_, err := LoadLibrary([]Directory{files.directory("lib")})
		wantViolation(t, err, "instrument-v0.1")
	}
}

func TestLoadAnEmptyListOfDirectoriesGivesAnEmptyLibrary(t *testing.T) {
	t.Parallel()
	library := loadLibrary(t)
	equipment, instruments := library.Counts()
	if equipment != 0 || instruments != 0 || len(library.List()) != 0 {
		t.Fatalf("library = %d/%d, want empty", equipment, instruments)
	}
}
