package component

import (
	"encoding/json"
	"reflect"
	"testing"
)

type found struct{ Pointer, Keyword, Member string }

func foundIn(issues []Issue) []found {
	out := make([]found, len(issues))
	for i, issue := range issues {
		out[i] = found{issue.Pointer, issue.Keyword, issue.Member}
	}
	return out
}

func kindOf(t *testing.T, raw string) map[string]any {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		t.Fatal(err)
	}
	return doc
}

const noArguments = `{"type":"object","properties":{},"required":[],"additionalProperties":false}`

func TestTwinRequirementsReportEveryViolationOfAKindFileInOnePass(t *testing.T) {
	t.Parallel()
	doc := kindOf(t, `{"kind":"probe","metrics":[
	  {"name":"temperature","data_type":"number","unit":"Cel","eu_range":{"min":0,"max":50},"writable":false},
	  {"name":"humidity","data_type":"number","unit":"%","writable":false},
	  {"name":"level","data_type":"number","resolution":1,"writable":false}],
	 "commands":[{"name":"halt","args_schema":`+noArguments+`,"risk_class":"low"}]}`)
	want := []found{
		{"/metrics/0/resolution", "twin_member_required", "resolution"},
		{"/metrics/1/resolution", "twin_member_required", "resolution"},
		{"/metrics/2/unit", "twin_unit_required", "unit"},
		{"/commands/0/safe_stop", "twin_member_required", "safe_stop"},
	}
	if got := foundIn(TwinRequirements(doc)); !reflect.DeepEqual(got, want) {
		t.Fatalf("TwinRequirements = %+v, want %+v", got, want)
	}
}

func TestTwinRequirementsNameTheKeywordAndMemberOfEachRule(t *testing.T) {
	t.Parallel()
	const metric = `{"name":"m","data_type":"number","unit":"Cel","eu_range":{"min":0,"max":9},"resolution":1,"writable":false}`
	command := func(name, risk string) string {
		return `{"name":"` + name + `","args_schema":` + noArguments + `,"risk_class":"` + risk + `","safe_stop":false}`
	}
	cases := []struct {
		name, raw string
		want      []found
	}{
		{"a kind name with capitals", `{"kind":"Probe","metrics":[` + metric + `],"commands":[]}`, []found{{"/kind", "twin_name_pattern", "kind"}}},
		{"a kind without metrics", `{"kind":"probe","metrics":[],"commands":[]}`, []found{{"/metrics", "twin_member_required", "metrics"}}},
		{"a member beyond the interface", `{"kind":"probe","description":"x","metrics":[` + metric + `],"commands":[]}`,
			[]found{{"/description", "twin_member_unsupported", "description"}}},
		{"a metric member beyond the interface", `{"kind":"probe","metrics":[{"name":"m","data_type":"number","unit":"Cel","resolution":1,"writable":false,"scan":"fast"}],"commands":[]}`,
			[]found{{"/metrics/0/scan", "twin_member_unsupported", "scan"}}},
		{"a metric name beyond the limit", `{"kind":"probe","metrics":[{"name":"temperature_of_the_heating_element","data_type":"number","unit":"Cel","resolution":1,"writable":false}],"commands":[]}`,
			[]found{{"/metrics/0/name", "twin_name_pattern", "name"}}},
		{"an enum metric", `{"kind":"probe","metrics":[{"name":"m","data_type":"enum","resolution":1,"writable":false}],"commands":[]}`,
			[]found{{"/metrics/0/data_type", "twin_type_unsupported", "data_type"}}},
		{"a boolean metric with a unit", `{"kind":"probe","metrics":[{"name":"m","data_type":"boolean","unit":"on","resolution":1,"writable":false}],"commands":[]}`,
			[]found{{"/metrics/0/unit", "twin_unit_unexpected", "unit"}}},
		{"a number metric without a unit", `{"kind":"probe","metrics":[{"name":"m","data_type":"integer","resolution":1,"writable":false}],"commands":[]}`,
			[]found{{"/metrics/0/unit", "twin_unit_required", "unit"}}},
		{"an inverted range", `{"kind":"probe","metrics":[{"name":"m","data_type":"number","unit":"Cel","eu_range":{"min":5,"max":5},"resolution":1,"writable":false}],"commands":[]}`,
			[]found{{"/metrics/0/eu_range", "twin_range", "eu_range"}}},
		{"a command without a name", `{"kind":"probe","metrics":[` + metric + `],"commands":[{"args_schema":` + noArguments + `,"risk_class":"low","safe_stop":false}]}`,
			[]found{{"/commands/0/name", "twin_member_required", "name"}}},
		{"a command without arguments or risk class", `{"kind":"probe","metrics":[` + metric + `],"commands":[{"name":"go","safe_stop":false}]}`,
			[]found{{"/commands/0/args_schema", "twin_member_required", "args_schema"}, {"/commands/0/risk_class", "twin_member_required", "risk_class"}}},
		{"a critical command", `{"kind":"probe","metrics":[` + metric + `],"commands":[` + command("go", "critical") + `]}`, []found{{"/commands/0/risk_class", "twin_risk_class", "risk_class"}}},
		{"a command name with capitals", `{"kind":"probe","metrics":[` + metric + `],"commands":[` + command("Go", "low") + `]}`, []found{{"/commands/0/name", "twin_name_pattern", "name"}}},
		{"a metric that is no object", `{"kind":"probe","metrics":[1],"commands":[]}`, []found{{"/metrics/0", "twin_not_object", ""}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := foundIn(TwinRequirements(kindOf(t, tc.raw))); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("TwinRequirements = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestTwinRequirementsAcceptTheThermalKindFilesAndGiveEveryMessageAsEnglish(t *testing.T) {
	t.Parallel()
	for name, doc := range thermalKinds(t) {
		if issues := TwinRequirements(doc); len(issues) != 0 {
			t.Errorf("kind %s: %+v, want a kind file a twin can reproduce", name, issues)
		}
	}
	for _, issue := range TwinRequirements(kindOf(t, `{"kind":"probe","metrics":[{"name":"m","data_type":"number"}],"commands":[]}`)) {
		if issue.Message == "" || issue.Keyword == "" || !issue.Located {
			t.Errorf("issue %+v, want a located issue with a keyword and a message", issue)
		}
	}
}
