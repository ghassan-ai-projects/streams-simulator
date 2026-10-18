package component

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func thermalKinds(t *testing.T) map[string]map[string]any {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(testsupport.ThermalBoardKindsDir(), "*.kind.json"))
	if err != nil || len(paths) != 3 {
		t.Fatalf("thermal kind files = %v, %v", paths, err)
	}
	kinds := map[string]map[string]any{}
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		doc := kindOf(t, string(raw))
		kinds[doc["kind"].(string)] = doc
	}
	return kinds
}

func choiceOf(t *testing.T, choices map[string]any, kind string) map[string]any {
	t.Helper()
	for _, raw := range choices["kinds"].([]any) {
		if choice := raw.(map[string]any); choice["kind"] == kind {
			return choice
		}
	}
	t.Fatalf("no choice for kind %s in %v", kind, choices)
	return nil
}

func TestSuggestChoosesTheBinarysDefaultsForTheThermalKindFilesAndTheTwinScaffoldAcceptsThem(t *testing.T) {
	t.Parallel()
	library, kinds := builtInLibrary(t), thermalKinds(t)
	choices, err := Suggest(library, kinds)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(choices)
	if err != nil {
		t.Fatal(err)
	}
	twin, err := Twin(library, kinds, raw, "suggested.twin.json")
	if err != nil || len(twin.Implements) != 3 {
		t.Fatalf("Twin of the suggested choices = %v, %v, want a twin of the three kind files", twin, err)
	}
	if choices["id"] != "fan-indicator-thermal-sensor-twin" || choices["version"] != "0.1.0" || choices["schema_version"] != "0.1" {
		t.Fatalf("suggested document = %v", choices)
	}
}

func TestSuggestPicksAnInstrumentByUnitAndNoneWhenNothingIsCompatible(t *testing.T) {
	t.Parallel()
	choices, err := Suggest(builtInLibrary(t), thermalKinds(t))
	if err != nil {
		t.Fatal(err)
	}
	instrumentOf := func(kind, metric string) any {
		return choiceOf(t, choices, kind)["metrics"].(map[string]any)[metric].(map[string]any)["instrument"]
	}
	cases := []struct {
		kind, metric string
		want         any
	}{
		{"thermal_sensor", "temperature", "streamsim/rtd@1"},
		{"thermal_sensor", "humidity", "streamsim/humidity@1"},
		{"indicator", "brightness_permille", "none"},
		{"indicator", "pattern", "none"},
		{"fan", "duty_permille", "none"},
	}
	for _, tc := range cases {
		if got := instrumentOf(tc.kind, tc.metric); got != tc.want {
			t.Errorf("%s.%s instrument = %v, want %v", tc.kind, tc.metric, got, tc.want)
		}
	}
	if instance := choiceOf(t, choices, "thermal_sensor")["instance"]; instance != "thermal-sensor-01" {
		t.Errorf("instance = %v, want thermal-sensor-01", instance)
	}
}

func TestSuggestAnswersACommandDirectlyOnlyWhenTheArgumentsNameWhatTheyAreTheOnlyToSet(t *testing.T) {
	t.Parallel()
	choices, err := Suggest(builtInLibrary(t), thermalKinds(t))
	if err != nil {
		t.Fatal(err)
	}
	led := choiceOf(t, choices, "indicator")["commands"].(map[string]any)
	want := map[string]any{"set_led": map[string]any{"response": "direct", "sets": map[string]any{"brightness_permille": "brightness_permille", "pattern": "pattern"}}}
	if !jsonEqual(led, want) {
		t.Fatalf("indicator commands = %v, want %v: the safe stop is not asked for", led, want)
	}
	const rivals = `{"kind":"valve","metrics":[{"name":"opening","data_type":"integer","unit":"permille","eu_range":{"min":0,"max":1000},"resolution":1,"writable":true}],
	 "commands":[
	  {"name":"open_to","args_schema":{"type":"object","properties":{"opening":{"type":"integer"}},"required":["opening"],"additionalProperties":false},"risk_class":"low","safe_stop":false},
	  {"name":"trim_to","args_schema":{"type":"object","properties":{"opening":{"type":"integer"}},"required":["opening"],"additionalProperties":false},"risk_class":"low","safe_stop":false},
	  {"name":"ping","args_schema":` + noArguments + `,"risk_class":"low","safe_stop":false}]}`
	ambiguous, err := Suggest(builtInLibrary(t), map[string]map[string]any{"valve": kindOf(t, rivals)})
	if err != nil {
		t.Fatal(err)
	}
	none := map[string]any{"response": "none"}
	if got := choiceOf(t, ambiguous, "valve")["commands"]; !jsonEqual(got, map[string]any{"open_to": none, "trim_to": none, "ping": none}) {
		t.Fatalf("two commands that set one metric: %v, want response none for all", got)
	}
}

func TestSuggestRefusesWhatItCannotSuggestFor(t *testing.T) {
	t.Parallel()
	if _, err := Suggest(builtInLibrary(t), nil); !errors.Is(err, ErrNoKindFiles) {
		t.Fatalf("Suggest of no kind file = %v, want ErrNoKindFiles", err)
	}
	if _, err := Suggest(nil, thermalKinds(t)); !errors.Is(err, ErrNoLibrary) {
		t.Fatalf("Suggest without a library = %v, want ErrNoLibrary", err)
	}
	thin := map[string]map[string]any{"probe": kindOf(t, `{"kind":"probe","metrics":[{"name":"m","data_type":"number"}],"commands":[]}`)}
	_, err := Suggest(builtInLibrary(t), thin)
	invalid, ok := model.AsValidationError(err)
	if !ok || invalid.Contract != model.ContractKindFile || invalid.Source != "probe" || len(invalid.Issues) != 3 {
		t.Fatalf("Suggest of a kind a twin cannot reproduce = %v, want the three violations at once", err)
	}
}

func jsonEqual(got, want any) bool {
	a, errA := json.Marshal(got)
	b, errB := json.Marshal(want)
	return errA == nil && errB == nil && string(a) == string(b)
}

func TestAKindFileThatPassesTheTwinRequirementsScaffoldsATwinWithoutRefusal(t *testing.T) {
	t.Parallel()
	library := builtInLibrary(t)
	for name, doc := range thermalKinds(t) {
		if issues := TwinRequirements(doc); len(issues) != 0 {
			t.Fatalf("kind %s fails the twin requirements: %+v", name, issues)
		}
		choices, err := Suggest(library, map[string]map[string]any{name: doc})
		if err != nil {
			t.Fatal(err)
		}
		raw, err := json.Marshal(choices)
		if err != nil {
			t.Fatal(err)
		}
		if twin, err := Twin(library, map[string]map[string]any{name: doc}, raw, name+".twin.json"); err != nil || twin == nil {
			t.Errorf("kind %s: Twin = %v, %v, want no refusal for a kind that passes the requirements", name, twin, err)
		}
	}
}
