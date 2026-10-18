package domain

import (
	"encoding/json"
	"errors"
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
		[]model.Board{{Name: "b-1", Members: []string{"P1", "V1"}}})
	if err != nil {
		t.Fatal(err)
	}
	return stamped(t, record)[0]
}

func birthLine(t *testing.T, edit func(birth map[string]any)) []byte {
	t.Helper()
	birth := stampedBirth(t, "hardware", birthKinds())
	if edit != nil {
		edit(birth)
	}
	raw, err := json.Marshal(birth)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func firstKind(birth map[string]any) map[string]any {
	return birth["kinds"].([]any)[0].(map[string]any)
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

func TestABirthIsCapturedAsOneDraftKindFilePerAnnouncedKind(t *testing.T) {
	t.Parallel()
	captured, err := ParseCapture(birthLine(t, nil))
	if err != nil {
		t.Fatal(err)
	}
	if captured.Environment != "hardware" || captured.BDSeq != 3 || len(captured.Kinds) != 2 {
		t.Fatalf("capture = %+v, want a hardware birth of boot sequence 3 with its two kinds", captured)
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

func TestTheEnvironmentOfTheBirthIsEchoedWhetherSimulatedOrHardware(t *testing.T) {
	t.Parallel()
	for _, environment := range []string{"simulated", "hardware"} {
		raw, err := json.Marshal(stampedBirth(t, environment, birthKinds()))
		if err != nil {
			t.Fatal(err)
		}
		if captured, err := ParseCapture(raw); err != nil || captured.Environment != environment {
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
	captured, err := ParseCapture(raw)
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

func TestTheBirthIsFoundWhereAJournalOrAStreamKeepsIt(t *testing.T) {
	t.Parallel()
	bare := string(birthLine(t, nil))
	wrapped := `{"stream":"observer","seq":4,"type":"wire","at_ns":5,"direction":"node","record":` + bare + `}`
	pretty, err := json.MarshalIndent(json.RawMessage(bare), "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	observation := `{"stream":"observer","seq":1,"type":"observation","at_ns":1,"line":"NBIRTH is only a word here"}`
	cases := []struct{ name, input string }{
		{"one bare record", bare},
		{"one bare record with a trailing newline", bare + "\n"},
		{"one record over several lines", string(pretty)},
		{"one wrapped journal record", wrapped},
		{"the first wire record of a journal", observation + "\n" + "{\"torn\": NBIRTH\n" + wrapped + "\n" + `{"stream":"observer","type":"wire","direction":"host","record":{"type":"STATE"}}` + "\n"},
		{"a stream of records", `{"type":"DDATA"}` + "\n\n" + bare + "\n" + `{"type":"NDEATH"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			captured, err := ParseCapture([]byte(tc.input))
			if err != nil || len(captured.Kinds) != 2 || captured.BDSeq != 3 {
				t.Fatalf("ParseCapture = %+v, %v, want the two kinds of the birth", captured, err)
			}
		})
	}
}

func TestInputWithoutABirthIsRefusedAndSaysWhatItSkipped(t *testing.T) {
	t.Parallel()
	hostSide := `{"stream":"observer","type":"wire","direction":"host","record":` + string(birthLine(t, nil)) + `}`
	cases := []struct{ name, input, message string }{
		{"nothing", "", "no NBIRTH record found"},
		{"records of other types", `{"type":"DDATA"}` + "\n" + `{"type":"NDEATH"}`, "no NBIRTH record found"},
		{"a birth the host would have sent", hostSide, "no NBIRTH record found"},
		{"a line that names a birth but is cut", `{"type":"NBIRTH","seq":`, "are not JSON records"},
		{"a list instead of a record", `["NBIRTH"]`, "are not JSON records"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseCapture([]byte(tc.input))
			invalid := violation(t, err, ErrNoBirth)
			if len(invalid.Issues) != 1 || invalid.Issues[0].Keyword != "no_birth" || !strings.Contains(invalid.Issues[0].Message, tc.message) {
				t.Fatalf("issues = %+v, want one no_birth issue saying %q", invalid.Issues, tc.message)
			}
		})
	}
}

func TestABirthOfTheSealedDirectorStreamIsNeverTaken(t *testing.T) {
	t.Parallel()
	sealed := `{"stream":"director","seq":2,"type":"wire","at_ns":5,"direction":"node","record":` + string(birthLine(t, nil)) + `}`
	observer := strings.Replace(sealed, `"director"`, `"observer"`, 1)
	for name, input := range map[string]string{"alone": sealed, "before an observer birth": sealed + "\n" + observer} {
		if _, err := ParseCapture([]byte(input)); !errors.Is(err, ErrSealedStream) {
			t.Errorf("%s: ParseCapture = %v, want ErrSealedStream", name, err)
		}
	}
	if _, err := ParseCapture([]byte(observer)); err != nil {
		t.Fatalf("the same record of the observer stream = %v, want it taken", err)
	}
}
