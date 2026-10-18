package domain

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func TestKindAndComponentFilesAreWorkspacePathsTheCanvasMayWrite(t *testing.T) {
	t.Parallel()
	good := []struct {
		rel   string
		class FileClass
	}{
		{"kinds/fan.kind.json", ClassKind},
		{"kinds/e2e-fan.kind.json", ClassKind},
		{"components/equipment/e2e-board.component.json", ClassEquipment},
		{"components/equipment/arduino.component.json", ClassEquipment},
		{"components/instruments/dht11.instrument.json", ClassInstrument},
		{"components/equipment/a.b-c_1.component.json", ClassEquipment},
	}
	for _, tc := range good {
		path, err := ParsePath(tc.rel)
		if err != nil || path != (WorkspacePath{Rel: tc.rel, Class: tc.class}) || path.Writable() != nil {
			t.Errorf("ParsePath(%q) = %+v, %v; want a writable %s file", tc.rel, path, err, tc.class)
		}
	}
}

func TestKindAndComponentPathsRefuseTraversalWrongSuffixAndWrongFolder(t *testing.T) {
	t.Parallel()
	for _, rel := range []string{
		"kinds/fan.json", "kinds/.kind.json", "kinds/.hidden.kind.json", "kinds/a/b.kind.json", "kinds/../" + inDomains("x.domain.json"),
		"../kinds/fan.kind.json", "/kinds/fan.kind.json", "kinds//fan.kind.json", "Kinds/fan.kind.json", `kinds\fan.kind.json`,
		"kinds/fan.domain.json", inDomains("fan.kind.json"),
		"components/fan.component.json", "components/fan.instrument.json", "components/vocabulary.json",
		"components/equipment/fan.instrument.json", "components/instruments/fan.component.json",
		"components/equipment/.component.json", "components/equipment/a/b.component.json", "components/equipment/../x.component.json",
		"components/other/fan.component.json", "components/equipment/fan.json",
	} {
		if _, err := ParsePath(rel); !errors.Is(err, ErrBadPath) {
			t.Errorf("ParsePath(%q) = %v, want ErrBadPath", rel, err)
		}
	}
}

func TestEachDeviceFolderNamesItsClassAndSuffix(t *testing.T) {
	t.Parallel()
	if got := ComponentFolders(); !slices.Equal(got, []string{"components/equipment", "components/instruments"}) {
		t.Fatalf("ComponentFolders = %v", got)
	}
	want := []struct{ folder, class, suffix string }{
		{"kinds", "kind", ".kind.json"},
		{"components/equipment", "equipment", ".component.json"},
		{"components/instruments", "instrument", ".instrument.json"},
	}
	for _, tc := range want {
		class, suffix, known := ClassOfFolder(tc.folder)
		if !known || string(class) != tc.class || suffix != tc.suffix {
			t.Errorf("ClassOfFolder(%q) = %s, %s, %v", tc.folder, class, suffix, known)
		}
	}
	if slices.Contains(DefinitionFolders(), "kinds") {
		t.Fatal("kinds and components are listed beside the definitions, not ensured with them")
	}
}

func TestADeviceEntryReadsOnlyTheTopLevelNamesOfItsFile(t *testing.T) {
	t.Parallel()
	kind := KindEntryOf("kinds/fan.kind.json", []byte(`{"kind":"fan","title":"Fan","metrics":[{"name":"kind"}]}`))
	if kind != (KindEntry{Path: "kinds/fan.kind.json", Kind: "fan", Title: "Fan"}) {
		t.Fatalf("kind entry = %+v", kind)
	}
	if got := KindEntryOf("kinds/x.kind.json", []byte(`{`)); got != (KindEntry{Path: "kinds/x.kind.json"}) {
		t.Fatalf("a broken kind file = %+v", got)
	}
	if got := KindEntryOf("kinds/x.kind.json", []byte(`{"kind":{"state":[]}}`)); got != (KindEntry{Path: "kinds/x.kind.json"}) {
		t.Fatalf("a kind member that is not text = %+v", got)
	}
	component := ComponentEntryOf("components/equipment/a.component.json", ClassEquipment,
		[]byte(`{"id":"acme/board","title":"Board","class":"static/board","kind":{"state":[]}}`))
	wantComponent := ComponentEntry{Path: "components/equipment/a.component.json", ID: "acme/board", Title: "Board", Class: ClassEquipment}
	if !reflect.DeepEqual(component, wantComponent) {
		t.Fatalf("component entry = %+v", component)
	}
	if got := ComponentEntryOf("components/instruments/b.instrument.json", ClassInstrument, nil); got.ID != "" || got.Class != ClassInstrument {
		t.Fatalf("an unreadable instrument = %+v", got)
	}
}

func TestKindValidateTakesOneKindFileAndNothingElse(t *testing.T) {
	t.Parallel()
	validate, err := Lookup("kind.validate")
	if err != nil || validate.Kind != KindCLI || validate.Mutating || validate.Etag || validate.Kinds || validate.Library {
		t.Fatalf("kind.validate = %+v, %v", validate, err)
	}
	got, err := validate.Argv(map[string]any{"path": "kinds/fan.kind.json"})
	want := []ArgvPart{{Literal: "kind"}, {Literal: "validate"}, {Path: "kinds/fan.kind.json"}}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("Argv = %+v, %v", got, err)
	}
	refused := []struct {
		name string
		args map[string]any
	}{
		{"a domain file", map[string]any{"path": inDomains("fan.domain.json")}},
		{"a missing path", map[string]any{}},
		{"a path outside", map[string]any{"path": "../fan.kind.json"}},
		{"an unknown argument", map[string]any{"path": "kinds/fan.kind.json", "kinds_dir": "/"}},
	}
	for _, tc := range refused {
		if _, err := validate.Argv(tc.args); !errors.Is(err, ErrBadArguments) {
			t.Errorf("%s: %v, want ErrBadArguments", tc.name, err)
		}
	}
}

func TestKindTwinTakesAChoicesFileAndADomainToWriteAndGuardsTheDomain(t *testing.T) {
	t.Parallel()
	twin, err := Lookup("kind.twin")
	if err != nil || twin.Kind != KindCLI || !twin.Mutating || !twin.Etag || !twin.Kinds || !twin.Library {
		t.Fatalf("kind.twin = %+v, %v", twin, err)
	}
	args := map[string]any{"choices": "twins/fan.twin.json", "out": inDomains("fan.domain.json")}
	got, err := twin.Argv(args)
	want := []ArgvPart{{Literal: "kind"}, {Literal: "twin"}, {Literal: "--choices"}, {Path: "twins/fan.twin.json"}, {Literal: "--out"}, {Path: inDomains("fan.domain.json")}}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("Argv = %+v, %v", got, err)
	}
	if guard, guarded, err := twin.Guarded(args); err != nil || !guarded || guard.Path != inDomains("fan.domain.json") {
		t.Fatalf("Guarded = %+v, %v, %v, want the domain file guarded", guard, guarded, err)
	}
	refused := []struct {
		name string
		args map[string]any
	}{
		{"choices that are a kind file", map[string]any{"choices": "kinds/fan.kind.json", "out": inDomains("fan.domain.json")}},
		{"an output that is a twin file", map[string]any{"choices": "twins/fan.twin.json", "out": "twins/other.twin.json"}},
		{"no output", map[string]any{"choices": "twins/fan.twin.json"}},
		{"an unknown argument", map[string]any{"choices": "twins/fan.twin.json", "out": inDomains("fan.domain.json"), "kinds_dir": "/"}},
	}
	for _, tc := range refused {
		if _, err := twin.Argv(tc.args); !errors.Is(err, ErrBadArguments) {
			t.Errorf("%s: %v, want ErrBadArguments", tc.name, err)
		}
	}
}

func TestAChoicesFileOfTheWorkspaceIsATwinFileThatCanBeWritten(t *testing.T) {
	t.Parallel()
	path, err := ParsePath("twins/fan.twin.json")
	if err != nil || path.Class != ClassTwin || path.Writable() != nil {
		t.Fatalf("ParsePath = %+v, %v, want a writable twin file", path, err)
	}
	if _, err := ParsePath("twins/fan.json"); !errors.Is(err, ErrBadPath) {
		t.Fatalf("a twins file without its suffix: %v, want ErrBadPath", err)
	}
}

func TestComponentValidateChecksTheWholeLibraryAndTakesNoArguments(t *testing.T) {
	t.Parallel()
	validate, err := Lookup("component.validate")
	if err != nil || validate.Kind != KindCLI || !validate.Library || validate.Mutating {
		t.Fatalf("component.validate = %+v, %v", validate, err)
	}
	got, err := validate.Argv(nil)
	if err != nil || !reflect.DeepEqual(got, []ArgvPart{{Literal: "component"}, {Literal: "validate"}}) {
		t.Fatalf("Argv = %+v, %v", got, err)
	}
	if _, err := validate.Argv(map[string]any{"components_dir": "/"}); !errors.Is(err, ErrBadArguments) {
		t.Fatalf("an argument: %v", err)
	}
}

func TestOnlyDomainValidateComposeTwinAndSuggestCarryTheWorkspaceKinds(t *testing.T) {
	t.Parallel()
	for _, action := range Actions() {
		want := action.Name == "domain.validate" || action.Name == "domain.compose" || action.Name == "kind.twin" || action.Name == "kind.suggest"
		if action.Kinds != want {
			t.Errorf("%s: Kinds = %v, want %v", action.Name, action.Kinds, want)
		}
	}
}
