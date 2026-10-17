package domain

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestParseKindFilesIndexesEachKindAndSummarisesIt(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(testsupport.PumpKindFile())
	if err != nil {
		t.Fatal(err)
	}
	other := `{"kind": "valve", "metrics": [], "commands": []}`
	files, summaries, err := ParseKindFiles([]KindSource{{Path: "pump.kind.json", Raw: raw}, {Path: "valve.kind.json", Raw: []byte(other)}})
	if err != nil || len(files) != 2 || len(summaries) != 2 {
		t.Fatalf("ParseKindFiles = %v, %v, %v, want two kinds", files, summaries, err)
	}
	pump := summaries[0]
	if pump.Path != "pump.kind.json" || pump.Kind != "pump" || pump.Metrics != 2 || pump.Commands != 3 || pump.KindDigest != KindDigest(files["pump"]) {
		t.Fatalf("summary = %+v, want the pump kind file with its digest", pump)
	}
}

func TestParseKindFilesRefusesWhatIsNotAKindFile(t *testing.T) {
	t.Parallel()
	valid := `{"kind": "valve", "metrics": [], "commands": []}`
	cases := []struct {
		name    string
		sources []KindSource
		want    error
		message string
	}{
		{"not JSON", []KindSource{{Path: "bad.kind.json", Raw: []byte("{")}}, nil, "kind file bad.kind.json"},
		{"a kind file without commands", []KindSource{{Path: "thin.kind.json", Raw: []byte(`{"kind": "k", "metrics": []}`)}}, nil,
			"thin.kind.json: plantwire: kind file fails its schema"},
		{"one kind defined twice", []KindSource{{Path: "a.kind.json", Raw: []byte(valid)}, {Path: "b.kind.json", Raw: []byte(valid)}},
			ErrKindFileDuplicate, `"valve" is defined by a.kind.json and b.kind.json`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			files, summaries, err := ParseKindFiles(tc.sources)
			wrongError := err == nil || !strings.Contains(err.Error(), tc.message) || (tc.want != nil && !errors.Is(err, tc.want))
			if wrongError || files != nil || summaries != nil {
				t.Fatalf("ParseKindFiles = %v, %v, %v, want an error containing %q", files, summaries, err, tc.message)
			}
		})
	}
}

func TestFirstDifferenceNamesThePathOfTheFirstDifferingMember(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		want, got any
		path      string
		differs   bool
	}{
		{"equal trees", map[string]any{"a": []any{"x", map[string]any{"b": true}}}, map[string]any{"a": []any{"x", map[string]any{"b": true}}}, "", false},
		{"a scalar", map[string]any{"a": "x"}, map[string]any{"a": "y"}, "/a", true},
		{"a member only the kind file has", map[string]any{"a": "x", "b": "y"}, map[string]any{"a": "x"}, "/b", true},
		{"a member only the kind has", map[string]any{"a": "x"}, map[string]any{"a": "x", "c": nil}, "/c", true},
		{"a longer list", map[string]any{"l": []any{"x"}}, map[string]any{"l": []any{"x", "y"}}, "/l/1", true},
		{"a shorter list", map[string]any{"l": []any{"x", "y"}}, map[string]any{"l": []any{"x"}}, "/l/1", true},
		{"an object where a list was", map[string]any{"l": []any{}}, map[string]any{"l": map[string]any{}}, "/l", true},
		{"a list where an object was", map[string]any{"o": map[string]any{}}, map[string]any{"o": []any{}}, "/o", true},
		{"a number against a string", map[string]any{"n": "1"}, map[string]any{"n": 1}, "/n", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if path, differs := firstDifference(tc.want, tc.got, ""); path != tc.path || differs != tc.differs {
				t.Fatalf("firstDifference = %q, %v, want %q, %v", path, differs, tc.path, tc.differs)
			}
		})
	}
}
