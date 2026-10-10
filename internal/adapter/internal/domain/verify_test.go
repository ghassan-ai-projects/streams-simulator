package domain

import (
	"errors"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestValidateStrictObservedOrder(t *testing.T) {
	t.Parallel()
	base := model.DefaultStartTimeNS
	cases := []struct {
		name    string
		events  []model.SimEvent
		wantErr bool
	}{
		{
			name: "strict",
			events: []model.SimEvent{
				{ObservedTime: model.FormatTime(base)},
				{ObservedTime: model.FormatTime(base + 1)},
			},
		},
		{
			name: "tie",
			events: []model.SimEvent{
				{ObservedTime: model.FormatTime(base)},
				{ObservedTime: model.FormatTime(base)},
			},
			wantErr: true,
		},
		{
			name: "out-of-order",
			events: []model.SimEvent{
				{ObservedTime: model.FormatTime(base + 1)},
				{ObservedTime: model.FormatTime(base)},
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validateStrictObservedOrder(tc.events)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateStrictObservedOrder() error = %v, wantErr %t", err, tc.wantErr)
			}
		})
	}
}

// memoryFiles serves the schema and golden bytes a test supplies.
type memoryFiles struct {
	schema, golden []byte
	schemaErr      error
	goldenErr      error
}

func (m memoryFiles) OutputSchema(string) ([]byte, error) { return m.schema, m.schemaErr }
func (m memoryFiles) Golden(string) ([]byte, error)       { return m.golden, m.goldenErr }

func shippedNative(t *testing.T) (*model.Adapter, []model.SimEvent, []byte) {
	t.Helper()
	a, err := Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := FixtureEvents()
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := renderVerification(a, fixture)
	if err != nil {
		t.Fatal(err)
	}
	return a, fixture, rendered
}

const anyObject = `{"type":"object"}`

func TestVerifyAcceptsAnAdapterWhoseOutputMatchesItsSchemaAndGolden(t *testing.T) {
	t.Parallel()
	a, fixture, rendered := shippedNative(t)
	res, err := Verify(a, fixture, "base", memoryFiles{schema: []byte(anyObject), golden: rendered})
	if err != nil || !res.SchemaOK || !res.GoldenMatch || res.RecordCount != len(fixture) {
		t.Fatalf("result = %+v (%v)", res, err)
	}
}

func TestVerifyReportsTheFirstGoldenDivergence(t *testing.T) {
	t.Parallel()
	a, fixture, rendered := shippedNative(t)
	altered := append([]byte{rendered[0] ^ 1}, rendered[1:]...)
	res, err := Verify(a, fixture, "", memoryFiles{schema: []byte(anyObject), golden: altered})
	if err != nil || res.GoldenMatch || res.FirstDivergence != "first divergent byte at line 1, column 1" {
		t.Fatalf("result = %+v (%v)", res, err)
	}
	short, err := Verify(a, fixture, "", memoryFiles{schema: []byte(anyObject), golden: rendered[:len(rendered)-1]})
	if err != nil || short.FirstDivergence != "lengths differ" || short.Detail == "" {
		t.Fatalf("truncated golden = %+v (%v)", short, err)
	}
}

func TestVerifyStopsAtTheFirstRecordThatViolatesTheSchema(t *testing.T) {
	t.Parallel()
	a, fixture, rendered := shippedNative(t)
	schema := []byte(`{"type":"object","required":["no_such_field"]}`)
	res, err := Verify(a, fixture, "", memoryFiles{schema: schema, golden: rendered})
	if err != nil || res.SchemaOK || res.GoldenMatch || !strings.Contains(res.FirstDivergence, "record 0 fails") {
		t.Fatalf("result = %+v (%v)", res, err)
	}
}

func TestVerifyNamesTheFileItCouldNotRead(t *testing.T) {
	t.Parallel()
	a, fixture, rendered := shippedNative(t)
	cases := []struct {
		name  string
		files memoryFiles
		want  string
	}{
		{"schema", memoryFiles{schemaErr: errors.New("gone")}, "adapter: read output schema"},
		{"golden", memoryFiles{schema: []byte(anyObject), goldenErr: errors.New("gone")}, "adapter: read golden"},
		{"schema text", memoryFiles{schema: []byte("not json"), golden: rendered}, "adapter: "},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, err := Verify(a, fixture, "", tc.files); err == nil || !strings.HasPrefix(err.Error(), tc.want) {
				t.Fatalf("err = %v, want prefix %q", err, tc.want)
			}
		})
	}
}

func TestVerifyWithoutDeclaredConformanceChecksNothingAndResolvesNoPath(t *testing.T) {
	t.Parallel()
	a, fixture, _ := shippedNative(t)
	a.Conformance = nil
	res, err := Verify(a, fixture, "", memoryFiles{})
	if err != nil || res.Adapter != a.ID || res.SchemaOK || res.GoldenMatch {
		t.Fatalf("result = %+v (%v)", res, err)
	}
	if got := resolveConformancePath("/base", "/abs/x.json"); got != "/abs/x.json" {
		t.Fatalf("absolute path = %q", got)
	}
	if got := resolveConformancePath("", "x.json"); got != "x.json" {
		t.Fatalf("empty base = %q", got)
	}
}

func TestDecodeFixtureRecordsReadsOneEventPerLine(t *testing.T) {
	t.Parallel()
	events, err := DecodeFixtureRecords([]byte("{\"observed_time\":\"2026-01-01T00:00:00Z\"}\n\n"))
	if err != nil || len(events) != 1 {
		t.Fatalf("events = %v (%v)", events, err)
	}
	if _, err := DecodeFixtureRecords([]byte("not json\n")); err == nil || !strings.HasPrefix(err.Error(), "adapter: ") {
		t.Fatalf("malformed fixture: %v", err)
	}
}

func TestVerifyRefusesAnEmptyFixtureInsteadOfPanicking(t *testing.T) {
	t.Parallel()
	a, _, _ := shippedNative(t)
	_, err := Verify(a, nil, "", memoryFiles{})
	if err == nil || !strings.Contains(err.Error(), "holds no events") {
		t.Fatalf("err = %v", err)
	}
}

func TestSplitRecordsReadsJSONArrayOutputElementwise(t *testing.T) {
	t.Parallel()
	got, err := splitRecords([]byte("[\n{\"a\":1}\n,{\"b\":2}\n]\n"), "json-array")
	if err != nil || len(got) != 2 || got[0] != `{"a":1}` || got[1] != `{"b":2}` {
		t.Fatalf("records = %q, err = %v", got, err)
	}
	if _, err := splitRecords([]byte("not an array"), "json-array"); err == nil || !strings.Contains(err.Error(), "not a JSON array") {
		t.Fatalf("err = %v", err)
	}
	lines, err := splitRecords([]byte("{\"a\":1}\n\n{\"b\":2}\n"), "jsonl")
	if err != nil || len(lines) != 2 {
		t.Fatalf("jsonl records = %q, err = %v", lines, err)
	}
}

func TestVerifyReportsWhichDeclaredChecksRan(t *testing.T) {
	t.Parallel()
	a, fixture, rendered := shippedNative(t)
	res, err := Verify(a, fixture, "", memoryFiles{schema: []byte(anyObject), golden: rendered})
	if err != nil || !res.SchemaChecked || !res.GoldenChecked {
		t.Fatalf("both declared checks must be reported as run: %+v (%v)", res, err)
	}
	a.Conformance = nil
	res, err = Verify(a, fixture, "", memoryFiles{})
	if err != nil || res.SchemaChecked || res.GoldenChecked {
		t.Fatalf("no declared checks, none reported: %+v (%v)", res, err)
	}
}
