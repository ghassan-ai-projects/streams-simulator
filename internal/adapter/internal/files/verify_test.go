package files

// Conformance of the shipped adapters: every adapter renders the fixture,
// validates against its declared schema, and byte-matches its committed
// golden. This is the S1 adapter-conformance gate.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestShippedAdaptersConform(t *testing.T) {
	t.Parallel()
	root := testsupport.AdaptersDir()
	cases := map[string]int{"native-jsonl": 12, "agentic-stream": 14} // events vs preamble+events+postamble
	for name, want := range cases {
		name := name
		want := want
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			res, err := Verify(filepath.Join(root, name+".adapter.json"), "", root)
			if err != nil {
				t.Fatalf("verify: %v", err)
			}
			if !res.SchemaOK {
				t.Fatalf("schema check failed: %s", res.FirstDivergence)
			}
			if !res.GoldenMatch {
				t.Fatalf("golden mismatch: %s (%s)", res.FirstDivergence, res.Detail)
			}
			if res.RecordCount != want {
				t.Fatalf("record count %d, want %d", res.RecordCount, want)
			}
		})
	}
}

func TestVerifyRejectsNonMonotonicFixture(t *testing.T) {
	t.Parallel()
	base := model.DefaultStartTimeNS
	events := []model.SimEvent{
		{ObservedTime: model.FormatTime(base + 2)},
		{ObservedTime: model.FormatTime(base + 1)},
	}
	var fixture strings.Builder
	for _, ev := range events {
		raw, err := json.Marshal(ev)
		if err != nil {
			t.Fatal(err)
		}
		fixture.Write(raw)
		fixture.WriteByte('\n')
	}
	fixturePath := filepath.Join(t.TempDir(), "fixture.jsonl")
	if err := os.WriteFile(fixturePath, []byte(fixture.String()), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := Verify(testsupport.Adapter("native-jsonl"), fixturePath, testsupport.AdaptersDir())
	if err == nil || !strings.Contains(err.Error(), "strict observed-time order") {
		t.Fatalf("adapter verify accepted non-monotonic fixture: %v", err)
	}
}

// copyShippedAdapter copies the native-jsonl adapter, its golden and its
// vendored output schema into a temp directory so a test can alter one.
func copyShippedAdapter(t *testing.T) (adapterPath, base string) {
	t.Helper()
	source := testsupport.AdaptersDir()
	base = t.TempDir()
	for _, rel := range []string{
		"native-jsonl.adapter.json",
		"golden/native-jsonl.jsonl",
		"vendor/streamsim/sim-event-v0.1.schema.json",
	} {
		raw, err := os.ReadFile(filepath.Join(source, rel))
		if err != nil {
			t.Fatal(err)
		}
		writeUnder(t, base, rel, raw)
	}
	return filepath.Join(base, "native-jsonl.adapter.json"), base
}

func TestVerifyReportsTheFirstDivergentGoldenByte(t *testing.T) {
	t.Parallel()
	adapterPath, base := copyShippedAdapter(t)
	golden := filepath.Join(base, "golden", "native-jsonl.jsonl")
	raw, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	raw[0] ^= 0x01
	writeUnder(t, base, "golden/native-jsonl.jsonl", raw)
	res, err := Verify(adapterPath, "", base)
	if err != nil {
		t.Fatal(err)
	}
	if res.GoldenMatch || !res.SchemaOK {
		t.Fatalf("result = %+v, want schema ok and golden mismatch", res)
	}
	if res.FirstDivergence != "first divergent byte at line 1, column 1" || res.Detail == "" {
		t.Fatalf("divergence = %q, detail = %q", res.FirstDivergence, res.Detail)
	}
}

func TestVerifyReportsATruncatedGolden(t *testing.T) {
	t.Parallel()
	adapterPath, base := copyShippedAdapter(t)
	golden := filepath.Join(base, "golden", "native-jsonl.jsonl")
	raw, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	writeUnder(t, base, "golden/native-jsonl.jsonl", raw[:len(raw)-1])
	res, err := Verify(adapterPath, "", base)
	if err != nil {
		t.Fatal(err)
	}
	if res.GoldenMatch || res.FirstDivergence != "lengths differ" {
		t.Fatalf("result = %+v, want a length difference", res)
	}
}

func TestVerifyReportsASchemaViolatingRecord(t *testing.T) {
	t.Parallel()
	adapterPath, base := copyShippedAdapter(t)
	writeUnder(t, base, "vendor/streamsim/sim-event-v0.1.schema.json",
		[]byte(`{"type":"object","required":["no_such_field"]}`))
	res, err := Verify(adapterPath, "", base)
	if err != nil {
		t.Fatal(err)
	}
	if res.SchemaOK || !strings.Contains(res.FirstDivergence, "record 0 fails sim-event-v0.1.schema.json") {
		t.Fatalf("result = %+v", res)
	}
}

func TestVerifyNamesAMissingGoldenFile(t *testing.T) {
	t.Parallel()
	adapterPath, base := copyShippedAdapter(t)
	if err := os.Remove(filepath.Join(base, "golden", "native-jsonl.jsonl")); err != nil {
		t.Fatal(err)
	}
	if _, err := Verify(adapterPath, "", base); err == nil || !strings.Contains(err.Error(), "adapter: read golden") {
		t.Fatalf("err = %v", err)
	}
}

// writeUnder writes a file below base, creating parent directories, without
// letting the relative path escape base.
func writeUnder(t *testing.T, base, rel string, content []byte) {
	t.Helper()
	root, err := os.OpenRoot(base)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := root.Close(); err != nil {
			t.Error(err)
		}
	}()
	if err := root.MkdirAll(filepath.Dir(rel), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := root.WriteFile(rel, content, 0o600); err != nil {
		t.Fatal(err)
	}
}
