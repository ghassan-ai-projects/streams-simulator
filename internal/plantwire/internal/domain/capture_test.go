package domain

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func birthKinds() []any {
	return []any{
		map[string]any{"kind": "pump", "metrics": []any{map[string]any{"name": "flow", "data_type": "number", "unit": "m3/h", "writable": false}},
			"commands": []any{map[string]any{"name": "start", "args_schema": map[string]any{"type": "object"}, "risk_class": "low", "safe_stop": false}}},
		map[string]any{"kind": "valve", "metrics": []any{}, "commands": []any{}},
	}
}

func stampedBirth(t *testing.T, environment string, kinds []any) map[string]any {
	t.Helper()
	record, err := NodeBirthOf(NodeIdentity{Environment: environment, ProtocolMax: 1, BDSeq: 3, PriorLastSeq: -1}, kinds,
		[]model.Board{{Name: "b-1", Members: []string{"P1", "V1"}, Environment: environment}})
	if err != nil {
		t.Fatal(err)
	}
	return stamped(t, record)[0]
}

func birthLine(t *testing.T, edit func(birth map[string]any)) string {
	t.Helper()
	birth := stampedBirth(t, "hardware", birthKinds())
	if edit != nil {
		edit(birth)
	}
	raw, err := json.Marshal(birth)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func birthTextOf(t *testing.T, kinds []any) string {
	t.Helper()
	raw, err := json.Marshal(stampedBirth(t, "hardware", kinds))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func wrapped(stream, record string) string {
	return `{"stream":"` + stream + `","seq":2,"type":"wire","at_ns":5,"direction":"node","record":` + record + `}`
}

func firstKind(birth map[string]any) map[string]any {
	return birth["kinds"].([]any)[0].(map[string]any)
}

func parse(input string) (Capture, error) {
	return ParseCapture(strings.NewReader(input))
}

func birthViolationOf(t *testing.T, err error) *model.ValidationError {
	t.Helper()
	invalid, ok := model.AsValidationError(err)
	if !ok || invalid.Contract != model.ContractPlantWireBirth || len(invalid.Issues) == 0 {
		t.Fatalf("error = %v, want a plant wire birth violation with issues", err)
	}
	return invalid
}

func violation(t *testing.T, err error, cause error) *model.ValidationError {
	t.Helper()
	invalid := birthViolationOf(t, err)
	if !errors.Is(err, cause) {
		t.Fatalf("error = %v, want it caused by %v", err, cause)
	}
	return invalid
}

func keywords(issues []model.Issue) []string {
	found := []string{}
	for _, issue := range issues {
		found = append(found, issue.Keyword)
	}
	return found
}

func TestABirthIsCapturedAsOneDraftKindFilePerAnnouncedKind(t *testing.T) {
	t.Parallel()
	captured, err := parse(birthLine(t, nil))
	if err != nil {
		t.Fatal(err)
	}
	if captured.Environment != "hardware" || captured.BDSeq != 3 || len(captured.Kinds) != 2 || captured.BirthsSeen != 1 ||
		captured.Taken != (Provenance{Line: 1, BDSeq: 3}) || captured.LaterBirths != nil || captured.Issues != nil {
		t.Fatalf("capture = %+v, want one hardware birth of boot sequence 3 taken from line 1, with its two kinds and no later birth", captured)
	}
	for i, name := range []string{"pump", "valve"} {
		kind := captured.Kinds[i]
		digest, err := KindDigest(kind.Draft)
		if _, announced := kind.Draft["kind_digest"]; announced || kind.Kind != name || err != nil || digest != kind.KindDigest || len(kind.Issues) != 0 || kind.Comparison != nil {
			t.Errorf("kind %d = %+v, want the draft of %s without a kind_digest whose content digests to %s (%v)", i, kind, name, kind.KindDigest, err)
		}
	}
	if err := CheckKind(captured.Kinds[0].Draft); err != nil {
		t.Fatalf("the draft of pump is no kind file: %v", err)
	}
}

func TestEveryKindNamesTheMembersTheBoardClaimedAndNobodyVerified(t *testing.T) {
	t.Parallel()
	kinds := birthKinds()
	command := kinds[0].(map[string]any)["commands"].([]any)[0].(map[string]any)
	command["args_schema"] = map[string]any{"type": "object", "required": []any{"minimum"}, "properties": map[string]any{
		"maximum":  map[string]any{"type": "integer", "minimum": 0, "maximum": 600},
		"lease/ms": map[string]any{"type": "integer", "exclusiveMinimum": 1, "enum": []any{map[string]any{"minimum": 9}}},
		"nested":   map[string]any{"anyOf": []any{map[string]any{"type": "string", "maxLength": 4}}},
	}}
	raw, err := json.Marshal(stampedBirth(t, "hardware", kinds))
	if err != nil {
		t.Fatal(err)
	}
	captured, err := parse(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"/commands/0/risk_class", "/commands/0/safe_stop", "/commands/0/args_schema/properties/lease~1ms/exclusiveMinimum",
		"/commands/0/args_schema/properties/maximum/maximum", "/commands/0/args_schema/properties/maximum/minimum",
		"/commands/0/args_schema/properties/nested/anyOf/0/maxLength"}
	if got := captured.Kinds[0].Claims; !reflect.DeepEqual(got, want) {
		t.Fatalf("claims of pump = %v, want %v", got, want)
	}
	if got := captured.Kinds[1].Claims; got == nil || len(got) != 0 {
		t.Fatalf("claims of valve = %#v, want an empty list that is present", got)
	}
}

func TestTheEnvironmentOfTheBirthIsEchoedWhetherSimulatedOrHardware(t *testing.T) {
	t.Parallel()
	for _, environment := range []string{"simulated", "hardware"} {
		raw, err := json.Marshal(stampedBirth(t, environment, birthKinds()))
		if err != nil {
			t.Fatal(err)
		}
		if captured, err := parse(string(raw)); err != nil || captured.Environment != environment {
			t.Errorf("ParseCapture = %+v, %v, want environment %s echoed", captured, err, environment)
		}
	}
}

func TestADraftWithAnEmptyRangeIsKeptWithTheIssuesAKindFileWouldReport(t *testing.T) {
	t.Parallel()
	kinds := birthKinds()
	metric := kinds[0].(map[string]any)["metrics"].([]any)[0].(map[string]any)
	metric["eu_range"] = map[string]any{"min": 5, "max": 5}
	raw, err := json.Marshal(stampedBirth(t, "hardware", kinds))
	if err != nil {
		t.Fatal(err)
	}
	captured, err := parse(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	issues := captured.Kinds[0].Issues
	if len(issues) != 1 || issues[0].Pointer != "/metrics/0/eu_range" || issues[0].Keyword != KeywordRangeEmpty || issues[0].Member != "eu_range" || len(captured.Kinds[1].Issues) != 0 {
		t.Fatalf("issues = %+v, want the one empty range of pump and none for valve", issues)
	}
}

func TestAKindDraftThatFailsForAnotherReasonIsReportedAsOneUnlocatedIssue(t *testing.T) {
	t.Parallel()
	issues := draftIssues(map[string]any{"kind": "k", "metrics": []any{}, "commands": []any{}, "bad": make(chan int)})
	if len(issues) != 1 || issues[0].Keyword != "canonical" {
		t.Fatalf("issues = %+v, want the missing canonical form reported", issues)
	}
}

func TestACapabilityDigestThatDoesNotHoldIsReportedBesideTheCaptureNotInsteadOfIt(t *testing.T) {
	t.Parallel()
	captured, err := parse(birthLine(t, func(b map[string]any) { b["capability_digest"] = "sha256:" + strings.Repeat("0", 64) }))
	if err != nil || len(captured.Kinds) != 2 {
		t.Fatalf("ParseCapture = %+v, %v, want the kinds captured all the same", captured, err)
	}
	if len(captured.Issues) != 1 || captured.Issues[0].Keyword != "capability_digest_mismatch" || captured.Issues[0].Pointer != "/capability_digest" || captured.Issues[0].Member != "capability_digest" {
		t.Fatalf("issues = %+v, want one capability_digest_mismatch at /capability_digest", captured.Issues)
	}
	if honest, err := parse(birthLine(t, nil)); err != nil || len(honest.Issues) != 0 {
		t.Fatalf("a birth with its own capability digest = %+v, %v, want no issue", honest.Issues, err)
	}
}

func TestTheBirthIsFoundWhereAJournalOrAStreamKeepsIt(t *testing.T) {
	t.Parallel()
	bare := birthLine(t, nil)
	pretty, err := json.MarshalIndent(json.RawMessage(bare), "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	observation := `{"stream":"observer","seq":1,"type":"observation","at_ns":1,"line":"NBIRTH is only a word here"}`
	hostSide := `{"stream":"observer","type":"wire","direction":"host","record":{"type":"STATE"}}`
	cases := []struct {
		name, input string
		line        int
		unreadable  []string
	}{
		{"one bare record", bare, 1, nil},
		{"one bare record with a trailing newline", bare + "\n", 1, nil},
		{"one record over several lines", string(pretty), 1, nil},
		{"one record of the observer stream", `{"stream":"observer",` + bare[1:], 1, nil},
		{"one wrapped journal record", wrapped("observer", bare), 1, nil},
		{"the first wire record of a journal", observation + "\n" + "{\"torn\": NBIRTH\n" + wrapped("observer", bare) + "\n" + hostSide + "\n", 3, []string{"line 2"}},
		{"a stream of records with blank lines", `{"type":"DDATA"}` + "\n\n" + bare + "\n" + `{"type":"NDEATH"}`, 3, nil},
		{"a journal with windows line ends", observation + "\r\n" + wrapped("observer", bare) + "\r\n", 2, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			captured, err := parse(tc.input)
			if err != nil || len(captured.Kinds) != 2 || captured.BDSeq != 3 || captured.Taken.Line != tc.line || captured.BirthsSeen != 1 {
				t.Fatalf("ParseCapture = %+v, %v, want the two kinds of the one birth taken from line %d", captured, err, tc.line)
			}
			if len(captured.Issues) != len(tc.unreadable) || (len(tc.unreadable) == 1 && (captured.Issues[0].Keyword != "birth_line_unreadable" || !strings.Contains(captured.Issues[0].Message, tc.unreadable[0]))) {
				t.Fatalf("issues = %+v, want the unreadable %v", captured.Issues, tc.unreadable)
			}
		})
	}
}

func TestInputWithoutABirthIsRefusedAndSaysWhatItSkipped(t *testing.T) {
	t.Parallel()
	hostSide := `{"stream":"observer","type":"wire","direction":"host","record":` + birthLine(t, nil) + `}`
	cases := []struct{ name, input, message string }{
		{"nothing", "", "no NBIRTH record found"},
		{"records of other types", `{"type":"DDATA"}` + "\n" + `{"type":"NDEATH"}`, "no NBIRTH record found"},
		{"a birth the host would have sent", hostSide, "no NBIRTH record found"},
		{"a line that names a birth but is cut", `{"type":"NBIRTH","seq":`, "line 1 names one but cannot be read"},
		{"a list instead of a record", `["NBIRTH"]`, "line 1 names one but cannot be read"},
		{"a record with another case of its keys", birthLine(t, func(b map[string]any) { delete(b, "type"); b["Type"] = "NBIRTH" }), "no NBIRTH record found"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := parse(tc.input)
			invalid := violation(t, err, ErrNoBirth)
			if len(invalid.Issues) != 1 || invalid.Issues[0].Keyword != "no_birth" || !strings.Contains(invalid.Issues[0].Message, tc.message) {
				t.Fatalf("issues = %+v, want one no_birth issue saying %q", invalid.Issues, tc.message)
			}
		})
	}
}
