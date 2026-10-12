package contract

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
)

func TestEveryValidPlantWireFixtureMatchesItsSchema(t *testing.T) {
	t.Parallel()
	schemas := compiledSchemas(t)
	if len(fixtures(t, "valid", "*.json")) != len(schemas) {
		t.Fatalf("one valid fixture per record type: %d fixtures, %d schemas", len(fixtures(t, "valid", "*.json")), len(schemas))
	}
	for _, file := range fixtures(t, "valid", "*.json") {
		name := strings.TrimSuffix(filepath.Base(file), ".json")
		if errs := schemas[name].Validate(readJSON(t, file)); len(errs) > 0 {
			t.Errorf("%s: %v", name, errs)
		}
	}
}

func TestEveryInvalidPlantWireFixtureFailsAtTheMemberItNames(t *testing.T) {
	t.Parallel()
	schemas := compiledSchemas(t)
	for _, file := range fixtures(t, "invalid", "*.json") {
		fixture := readJSON(t, file)
		schema, want := member(fixture, "schema").(string), member(fixture, "fails_at").(string)
		errs := schemas[schema].Validate(member(fixture, "record"))
		if !slices.ContainsFunc(errs, func(e jsonschema.Error) bool { return strings.HasPrefix(e.Path, want) }) {
			t.Errorf("%s: want a failure at %s, got %v", filepath.Base(file), want, errs)
		}
	}
}

func TestMustIgnoreAndInjectionFixturesAreValidRecords(t *testing.T) {
	t.Parallel()
	schemas := compiledSchemas(t)
	for _, file := range append(fixtures(t, "must-ignore", "*.json"), fixtures(t, "injection", "*.json")...) {
		fixture := readJSON(t, file)
		if errs := schemas[member(fixture, "schema").(string)].Validate(member(fixture, "record")); len(errs) > 0 {
			t.Errorf("%s: an unknown member or a hostile string must still validate: %v", filepath.Base(file), errs)
		}
	}
}

func TestExchangesHoldValidRecordsInContiguousSequencePerDirection(t *testing.T) {
	t.Parallel()
	schemas := compiledSchemas(t)
	for _, file := range fixtures(t, "exchanges", "*.jsonl") {
		header, steps := exchangeSteps(t, file)
		next := map[string]int64{}
		gaps := 0
		for _, step := range steps {
			record, from := member(step, "record"), member(step, "from")
			if record == nil {
				continue
			}
			kind := member(record, "type").(string)
			if errs := schemas[kind].Validate(record); len(errs) > 0 {
				t.Errorf("%s %s: %v", filepath.Base(file), kind, errs)
			}
			seq, _ := member(record, "seq").(json.Number).Int64()
			if seq != next[from.(string)] {
				gaps++
			}
			next[from.(string)] = seq + 1
		}
		if declared := member(header, "gap") == true; (gaps > 0) != declared {
			t.Errorf("%s: %d sequence gaps, gap declared %v", filepath.Base(file), gaps, declared)
		}
	}
}

func exchangeSteps(t *testing.T, file string) (any, []any) {
	t.Helper()
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var steps []any
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	for scanner.Scan() {
		decoder := json.NewDecoder(bytes.NewReader(scanner.Bytes()))
		decoder.UseNumber()
		var step any
		if err := decoder.Decode(&step); err != nil {
			t.Fatalf("%s: %v", file, err)
		}
		steps = append(steps, step)
	}
	return steps[0], steps[1:]
}
