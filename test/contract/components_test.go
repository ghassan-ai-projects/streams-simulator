package contract

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func componentsDir() string { return filepath.Join(testsupport.RepositoryRoot(), "components") }

func contractSchema(t *testing.T, stem string) *jsonschema.Schema {
	t.Helper()
	raw, err := os.ReadFile(testsupport.Contract(stem))
	if err != nil {
		t.Fatal(err)
	}
	schema, err := jsonschema.CompileJSON(raw)
	if err != nil {
		t.Fatalf("%s: %v", stem, err)
	}
	return schema
}

func TestEveryLibraryFileFollowsItsContract(t *testing.T) {
	t.Parallel()
	files := map[string]string{"vocabulary.json": "component-vocabulary-v0.1", "equipment/*.component.json": "component-v0.1",
		"instruments/*.instrument.json": "instrument-v0.1", "compositions/*.composition.json": "composition-v0.1"}
	for pattern, stem := range files {
		schema := contractSchema(t, stem)
		paths, _ := filepath.Glob(filepath.Join(componentsDir(), pattern))
		if len(paths) == 0 {
			t.Fatalf("no %s files", pattern)
		}
		for _, path := range paths {
			if errs := schema.Validate(readJSON(t, path)); len(errs) > 0 {
				t.Errorf("%s: %v", filepath.Base(path), errs[0])
			}
		}
	}
}

func TestBuiltInLibraryFilesUseTheStreamsimNamespace(t *testing.T) {
	t.Parallel()
	paths, _ := filepath.Glob(filepath.Join(componentsDir(), "*", "*.json"))
	for _, path := range paths {
		if id, ok := member(readJSON(t, path), "id").(string); ok && strings.Contains(path, "/compositions/") == false && !strings.HasPrefix(id, "streamsim/") {
			t.Errorf("%s: %s is not in the streamsim namespace; a customer library lives outside this repository", filepath.Base(path), id)
		}
	}
}

func TestTheHandComposedPilotValidatesAndRunsDeterministically(t *testing.T) {
	t.Parallel()
	spec, err := domain.Load(filepath.Join(testsupport.RepositoryRoot(), "examples", "chiller-pilot.domain.json"))
	if err != nil {
		t.Fatal(err)
	}
	a, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	digests := make([]string, 2)
	for i := range digests {
		r, err := run.New(context.Background(), run.Config{Domain: spec, Adapter: a, Seed: 7, SinkName: model.SinkInproc,
			TimeMode: model.TimeStepped, StartTimeNS: model.DefaultStartTimeNS})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := r.Advance(context.Background(), model.DefaultStartTimeNS+3600e9, false); err != nil {
			t.Fatal(err)
		}
		art, err := r.End("")
		if err != nil || art.Counts.Emitted == 0 {
			t.Fatalf("end = %+v, %v", art, err)
		}
		digests[i] = art.ExpectedTraceDigest
	}
	if digests[0] != digests[1] {
		t.Fatalf("two runs of the pilot differ: %v", digests)
	}
}
