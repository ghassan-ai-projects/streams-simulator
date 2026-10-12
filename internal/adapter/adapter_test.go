package adapter_test

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func shippedAdapter(t *testing.T, id string) *model.Adapter {
	t.Helper()
	a, err := adapter.Load(testsupport.Adapter(id))
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestNewEngineRefusesAMissingAdapter(t *testing.T) {
	t.Parallel()
	if _, err := adapter.NewEngine(nil, nil); !errors.Is(err, adapter.ErrNoAdapter) {
		t.Fatalf("err = %v, want ErrNoAdapter", err)
	}
}

func TestLoadBytesValidatesADocumentAndNamesItsSource(t *testing.T) {
	t.Parallel()
	loaded := shippedAdapter(t, "native-jsonl")
	again, err := adapter.LoadBytes(loaded.Raw, "in-memory")
	if err != nil || again.ID != loaded.ID {
		t.Fatalf("round trip: %v, %+v", err, again)
	}
	if _, err := adapter.LoadBytes([]byte(`{"id":"x"}`), "broken.json"); err == nil || !strings.Contains(err.Error(), "broken.json") {
		t.Fatalf("invalid document: %v", err)
	}
}

func TestEngineRendersAStreamingSessionTheSameAsAWholeRun(t *testing.T) {
	t.Parallel()
	events, err := adapter.FixtureEvents()
	if err != nil || len(events) == 0 {
		t.Fatalf("fixture: %d events (%v)", len(events), err)
	}
	meta := map[string]any{
		"run_id": "r", "sim_version": "0.1.0", "domain_id": "d", "domain_version": "0.0.0",
		"world_start_time": events[0].EventTime, "world_end_time": events[len(events)-1].EventTime, "seed": float64(0),
	}
	end := int64(1)
	whole, err := adapter.NewEngine(shippedAdapter(t, "native-jsonl"), meta)
	if err != nil {
		t.Fatal(err)
	}
	want, err := whole.RenderRun(events, end)
	if err != nil {
		t.Fatal(err)
	}
	stream, err := adapter.NewEngine(shippedAdapter(t, "native-jsonl"), meta)
	if err != nil {
		t.Fatal(err)
	}
	var got strings.Builder
	lines, err := stream.Begin()
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range lines {
		got.WriteString(line + "\n")
	}
	for i := range events {
		line, err := stream.RenderStreamRecord(&events[i])
		if err != nil {
			t.Fatal(err)
		}
		if line != "" {
			got.WriteString(line + "\n")
		}
	}
	tail, err := stream.End(end)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range tail {
		got.WriteString(line + "\n")
	}
	if got.String() != string(want) {
		t.Fatalf("streaming session differs from RenderRun:\n%q\n%q", got.String(), want)
	}
}

func TestVerifyProvesEveryShippedAdapterAgainstItsGolden(t *testing.T) {
	t.Parallel()
	root := testsupport.AdaptersDir()
	for _, name := range []string{"native-jsonl", "agentic-stream"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			res, err := adapter.Verify(filepath.Join(root, name+".adapter.json"), "", root)
			if err != nil {
				t.Fatalf("verify: %v", err)
			}
			if !res.SchemaOK || !res.GoldenMatch || res.RecordCount == 0 {
				t.Fatalf("result = %+v", res)
			}
		})
	}
}

func TestVerifyNamesAnAdapterFileItCannotLoad(t *testing.T) {
	t.Parallel()
	_, err := adapter.Verify(filepath.Join(t.TempDir(), "absent.adapter.json"), "", "")
	if err == nil || !strings.HasPrefix(err.Error(), "adapter: ") {
		t.Fatalf("err = %v", err)
	}
}

func TestLookupReturnsErrUnknownAdapterThroughTheFacade(t *testing.T) {
	t.Parallel()
	native, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	set := map[string]*model.Adapter{native.ID: native}
	if got, err := adapter.Lookup(set, native.ID); err != nil || got != native {
		t.Fatalf("Lookup(%s) = %v, %v", native.ID, got, err)
	}
	if _, err := adapter.Lookup(set, "no-such"); !errors.Is(err, adapter.ErrUnknownAdapter) {
		t.Fatalf("Lookup(no-such) = %v, want ErrUnknownAdapter", err)
	}
}

func TestLoadDirKeysEveryShippedAdapterByIDAndRefusesAnEmptyDirectory(t *testing.T) {
	t.Parallel()
	adapters, err := adapter.LoadDir(testsupport.AdaptersDir())
	if err != nil {
		t.Fatal(err)
	}
	if adapters["native-jsonl"] == nil || adapters["native-jsonl"].ID != "native-jsonl" {
		t.Fatalf("adapters = %v, want the native adapter keyed by its id", adapters)
	}
	empty := t.TempDir()
	if _, err := adapter.LoadDir(empty); err == nil || err.Error() != "no adapters found in "+empty {
		t.Fatalf("an empty directory: err = %v", err)
	}
	if _, err := adapter.LoadDir(empty + "/absent"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a missing directory: err = %v, want ErrNotExist", err)
	}
}

func TestVerifyFailureIsEmptyOnlyWhenADeclaredCheckPassed(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name         string
		res          adapter.VerifyResult
		want         string
		wantDiverged bool
	}{
		{"nothing declared", adapter.VerifyResult{Adapter: "a"}, "adapter verify: a declares no conformance schema or golden to check", false},
		{"schema only, ok", adapter.VerifyResult{Adapter: "a", SchemaChecked: true, SchemaOK: true}, "", false},
		{"golden only, ok", adapter.VerifyResult{Adapter: "a", GoldenChecked: true, GoldenMatch: true}, "", false},
		{"schema failed", adapter.VerifyResult{Adapter: "a", SchemaChecked: true, FirstDivergence: "record 0 fails"}, "adapter verify FAILED: record 0 fails", true},
		{"golden failed", adapter.VerifyResult{Adapter: "a", SchemaChecked: true, SchemaOK: true, GoldenChecked: true, Detail: "lengths differ"}, "adapter verify FAILED: lengths differ", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, diverged := adapter.VerifyFailure(&tc.res)
			if got != tc.want || diverged != tc.wantDiverged {
				t.Fatalf("VerifyFailure = %q, %v; want %q, %v", got, diverged, tc.want, tc.wantDiverged)
			}
		})
	}
}
