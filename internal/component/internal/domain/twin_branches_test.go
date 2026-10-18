package domain

import (
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var idPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{2,63}$`)

type located struct{ pointer, keyword, member string }

func locatedIn(issues []model.Issue) []located {
	if len(issues) == 0 {
		return nil
	}
	out := make([]located, len(issues))
	for i, issue := range issues {
		out[i] = located{issue.Pointer, issue.Keyword, issue.Member}
	}
	return out
}

func rawKind(t *testing.T, raw string) map[string]any {
	t.Helper()
	return kindDocuments(t, raw)[kindName(t, raw)]
}

func kindName(t *testing.T, raw string) string {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		t.Fatal(err)
	}
	name, _ := doc["kind"].(string)
	return name
}

func TestTwinRequirementsHandleAKindWithoutANameAndValuesOfTheWrongShape(t *testing.T) {
	t.Parallel()
	const metric = `{"name":"m","data_type":"number","unit":"degC","resolution":1,"writable":false}`
	cases := []struct {
		name string
		doc  map[string]any
		want []located
	}{
		{"a kind without a name", map[string]any{"metrics": []any{rawMetric(t, metric)}, "commands": []any{}}, []located{{"/kind", KeywordMemberRequired, "kind"}}},
		{"a risk class that is no text", map[string]any{"kind": "k", "metrics": []any{rawMetric(t, metric)}, "commands": []any{
			map[string]any{"name": "go", "args_schema": map[string]any{}, "risk_class": 7, "safe_stop": false}}}, []located{{"/commands/0/risk_class", KeywordRiskClass, "risk_class"}}},
		{"a command that is no object", map[string]any{"kind": "k", "metrics": []any{rawMetric(t, metric)}, "commands": []any{"go"}}, []located{{"/commands/0", KeywordNotObject, ""}}},
		{"a range bound that is no number", map[string]any{"kind": "k", "metrics": []any{map[string]any{"name": "m", "data_type": "number", "unit": "degC", "resolution": 1,
			"writable": false, "eu_range": map[string]any{"min": "low", "max": 5}}}, "commands": []any{}}, nil},
		{"a range bound beyond a double", map[string]any{"kind": "k", "metrics": []any{map[string]any{"name": "m", "data_type": "number", "unit": "degC", "resolution": 1,
			"writable": false, "eu_range": map[string]any{"min": json.Number("1e999"), "max": json.Number("5")}}}, "commands": []any{}}, nil},
		{"a range given as numbers of every kind", map[string]any{"kind": "k", "metrics": []any{map[string]any{"name": "m", "data_type": "number", "unit": "degC", "resolution": 1,
			"writable": false, "eu_range": map[string]any{"min": json.Number("9"), "max": 3.0}}}, "commands": []any{}}, []located{{"/metrics/0/eu_range", KeywordRange, "eu_range"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := locatedIn(TwinRequirements(tc.doc)); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("TwinRequirements = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func rawMetric(t *testing.T, raw string) map[string]any {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		t.Fatal(err)
	}
	return doc
}

func TestAKindFileWhoseValuesDoNotDecodeIsRefusedWithTheDecodeKeyword(t *testing.T) {
	t.Parallel()
	kinds := heaterKinds(t, func(doc map[string]any) { fileCommand(doc)["safe_stop"] = "yes" })
	_, err := ScaffoldTwin(baseLibrary(t), kinds, []byte(heaterChoicesJSON), "test.twin.json")
	invalid := asViolation(t, err, model.ContractTwinChoices)
	if !errors.Is(err, ErrKindNotReproducible) || len(invalid.Issues) != 1 || invalid.Issues[0].Keyword != KeywordDecode {
		t.Fatalf("ScaffoldTwin = %v (issues %+v), want one issue with the keyword %s", err, invalid.Issues, KeywordDecode)
	}
}

func instrumentLike(t *testing.T, id, unit string, low, high int) File {
	t.Helper()
	replaced := strings.NewReplacer(`"id": "acme/probe"`, `"id": "`+id+`"`, `"unit": "degC"`, `"unit": "`+unit+`"`,
		`"range": {"min": -50, "max": 400}`, `"range": {"min": `+strconv.Itoa(low)+`, "max": `+strconv.Itoa(high)+`}`).Replace(probeJSON)
	return file("instruments/"+strings.TrimPrefix(id, "acme/")+".instrument.json", replaced)
}

func summaryLines(summaries []Summary) []string {
	lines := []string{}
	for _, summary := range summaries {
		lines = append(lines, summary.ID+":"+summary.Unit+":"+summary.Match)
	}
	return lines
}

func TestInstrumentsForAUnitAreRankedExactFirstThenByIdentity(t *testing.T) {
	t.Parallel()
	library := libraryWith(t, instrumentLike(t, "acme/zzz-f", "degF", -58, 752), instrumentLike(t, "acme/alpha", "degC", 0, 100))
	cases := []struct {
		unit string
		want []string
	}{
		{"degF", []string{"acme/zzz-f:degF:exact", "acme/alpha:degC:convertible", "acme/probe:degC:convertible"}},
		{"degC", []string{"acme/alpha:degC:exact", "acme/probe:degC:exact", "acme/zzz-f:degF:convertible"}},
		{"K", []string{"acme/alpha:degC:convertible", "acme/probe:degC:convertible", "acme/zzz-f:degF:convertible"}},
		{"kW", []string{"acme/meter:kW:exact"}},
		{"furlong", []string{}},
		{"", nil},
	}
	for _, tc := range cases {
		got := summaryLines(library.InstrumentsForUnit(tc.unit))
		listed, err := library.ListForReading(tc.unit, nil, nil)
		if tc.unit == "" {
			if err != nil || len(listed) != len(library.List()) || len(got) != 0 {
				t.Errorf("an empty unit: ListForReading lists %d, want the %d of List; InstrumentsForUnit = %v", len(listed), len(library.List()), got)
			}
			continue
		}
		if !reflect.DeepEqual(got, tc.want) || err != nil || !reflect.DeepEqual(summaryLines(listed), tc.want) {
			t.Errorf("unit %q lists %v, want %v", tc.unit, got, tc.want)
		}
	}
}

func choicesOf(t *testing.T, library *Library, kinds map[string]map[string]any) map[string]any {
	t.Helper()
	choices, err := SuggestTwinChoices(library, kinds)
	if err != nil {
		t.Fatal(err)
	}
	return choices
}

func TestSuggestChoicesForTheHeaterKindAreTheBinarysDefaults(t *testing.T) {
	t.Parallel()
	got := choicesOf(t, baseLibrary(t), kindDocuments(t, heaterKindJSON))
	pick := func(instrument string) map[string]any { return map[string]any{"instrument": instrument} }
	want := map[string]any{"schema_version": "0.1", "id": "heater-twin", "version": "0.1.0", "title": "Twin of heater", "kinds": []any{
		map[string]any{"kind": "heater", "instance": "heater-01",
			"metrics": map[string]any{"temperature": pick("acme/probe@1"), "power": pick("acme/meter@1"), "setpoint": pick("acme/probe@1")},
			"commands": map[string]any{"set_setpoint": map[string]any{"response": "direct", "sets": map[string]any{"setpoint": "setpoint"}},
				"ping": map[string]any{"response": "none"}}}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("suggestion = %v, want %v", got, want)
	}
	raw, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ScaffoldTwin(baseLibrary(t), kindDocuments(t, heaterKindJSON), raw, "suggested.twin.json"); err != nil {
		t.Fatalf("the suggestion for the heater is refused by the scaffold: %v", err)
	}
}

func TestSuggestPrefersTheInstrumentNamedAfterTheReadingAndNoneWhenNothingFits(t *testing.T) {
	t.Parallel()
	library := libraryWith(t, instrumentLike(t, "acme/alpha", "degC", 0, 100))
	reading := func(name, unit, dataType string, eu string) map[string]map[string]any {
		raw := `{"kind":"k","metrics":[{"name":"` + name + `","data_type":"` + dataType + `",` + unit + eu + `"resolution":1,"writable":false}],"commands":[]}`
		return kindDocuments(t, raw)
	}
	cases := []struct {
		name  string
		kinds map[string]map[string]any
		want  string
	}{
		{"the first fitting by id", reading("temp", `"unit":"degC",`, "number", `"eu_range":{"min":0,"max":50},`), "acme/alpha@1"},
		{"the one named after the reading", reading("probe", `"unit":"degC",`, "number", `"eu_range":{"min":0,"max":50},`), "acme/probe@1"},
		{"only the wider instrument fits", reading("temp", `"unit":"degC",`, "number", `"eu_range":{"min":0,"max":300},`), "acme/probe@1"},
		{"a unit nothing measures", reading("temp", `"unit":"furlong",`, "number", `"eu_range":{"min":0,"max":50},`), "none"},
		{"a reading without a range", reading("temp", `"unit":"degC",`, "number", ``), "none"},
		{"a boolean reading", reading("on", ``, "boolean", ``), "none"},
	}
	for _, tc := range cases {
		metrics := choiceKind(choicesOf(t, library, tc.kinds))["metrics"].(map[string]any)
		for _, choice := range metrics {
			if got := choice.(map[string]any)["instrument"]; got != tc.want {
				t.Errorf("%s: instrument = %v, want %s", tc.name, got, tc.want)
			}
		}
	}
}

func choiceKind(choices map[string]any) map[string]any {
	return choices["kinds"].([]any)[0].(map[string]any)
}

func TestSuggestLeavesNoResponseDirectWhenTwoCommandsClaimTheSameReading(t *testing.T) {
	t.Parallel()
	const rivals = `{"kind":"valve","metrics":[{"name":"opening","data_type":"integer","unit":"permille","eu_range":{"min":0,"max":1000},"resolution":1,"writable":true}],
	 "commands":[
	  {"name":"open_to","args_schema":{"type":"object","properties":{"opening":{"type":"integer"}}},"risk_class":"low","safe_stop":false},
	  {"name":"trim_to","args_schema":{"type":"object","properties":{"opening":{"type":"integer"}}},"risk_class":"low","safe_stop":false},
	  {"name":"label","args_schema":{"type":"object","properties":{"opening":{"type":"string"}}},"risk_class":"low","safe_stop":false}]}`
	commands := choiceKind(choicesOf(t, baseLibrary(t), kindDocuments(t, rivals)))["commands"].(map[string]any)
	for name, response := range commands {
		if response.(map[string]any)["response"] != "none" || response.(map[string]any)["sets"] != nil {
			t.Errorf("command %s = %v, want response none", name, response)
		}
	}
	if len(commands) != 3 {
		t.Fatalf("commands = %v, want all three answered", commands)
	}
}

func TestSuggestNamesALongBoardAfterItsFirstKindAndTheCountOfTheOthers(t *testing.T) {
	t.Parallel()
	kinds := map[string]map[string]any{}
	for i := range 8 {
		name := "board_part_number_" + strconv.Itoa(i)
		kinds[name] = rawKind(t, `{"kind":"`+name+`","metrics":[{"name":"on","data_type":"boolean","resolution":1,"writable":false}],"commands":[]}`)
	}
	choices := choicesOf(t, baseLibrary(t), kinds)
	if choices["id"] != "board-part-number-0-and-7-more-twin" || choices["title"] != "Twin of board_part_number_0 and 7 more" {
		t.Fatalf("id %v and title %v, want the first kind and the count of the others", choices["id"], choices["title"])
	}
	if !idPattern.MatchString(choices["id"].(string)) {
		t.Fatalf("id %v does not fit the choices schema", choices["id"])
	}
}

func TestSuggestNamesATwinAfterTheKindsItIsGivenAndNoOther(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		kinds []string
		id    string
		title string
	}{
		{"one kind", []string{"zone"}, "zone-twin", "Twin of zone"},
		{"two kinds in alphabetical order", []string{"zone", "fan"}, "fan-zone-twin", "Twin of fan, zone"},
	}
	documents := kindDocuments(t, zoneKindJSON, fanKindJSON, heaterKindJSON)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			picked := map[string]map[string]any{}
			for _, kind := range tc.kinds {
				picked[kind] = documents[kind]
			}
			choices := choicesOf(t, baseLibrary(t), picked)
			if choices["id"] != tc.id || choices["title"] != tc.title || len(choices["kinds"].([]any)) != len(tc.kinds) {
				t.Fatalf("suggested id %v, title %v, %d kinds, want %s, %q and %d kinds", choices["id"], choices["title"], len(choices["kinds"].([]any)), tc.id, tc.title, len(tc.kinds))
			}
		})
	}
}

func TestSuggestRefusesNoKindFilesAndAKindATwinCannotReproduce(t *testing.T) {
	t.Parallel()
	if _, err := SuggestTwinChoices(baseLibrary(t), nil); !errors.Is(err, ErrNoKindFiles) {
		t.Fatalf("no kind file = %v, want ErrNoKindFiles", err)
	}
	thin := kindDocuments(t, `{"kind":"thin","metrics":[{"name":"m","data_type":"number"}],"commands":[]}`)
	_, err := SuggestTwinChoices(baseLibrary(t), thin)
	invalid := asViolation(t, err, model.ContractKindFile)
	if invalid.Source != "thin" || len(invalid.Issues) != 3 {
		t.Fatalf("a kind a twin cannot reproduce = %v, want its three issues named after the kind", err)
	}
}
