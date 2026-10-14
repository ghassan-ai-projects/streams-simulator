package files

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/report/internal/domain"
)

func TestReadReturnsAPresentFileAndAnAbsentOneWithoutError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "ledger.jsonl")
	if err := os.WriteFile(path, []byte("row\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	present, err := Read(path)
	if err != nil || !present.Present || string(present.Raw) != "row\n" {
		t.Fatalf("present = %+v (%v), want the file's bytes", present, err)
	}
	absent, err := Read(filepath.Join(dir, "absent.jsonl"))
	if err != nil || absent.Present {
		t.Fatalf("absent = %+v (%v), want not present and no error", absent, err)
	}
}

func TestReadNamesAPathItCannotRead(t *testing.T) {
	t.Parallel()
	if _, err := Read(t.TempDir()); err == nil || !strings.HasPrefix(err.Error(), "report: ") {
		t.Fatalf("err = %v, want a report-prefixed read failure", err)
	}
}

func TestScanMatchesTheWholeBufferScanOfTheSameBytes(t *testing.T) {
	t.Parallel()
	const opened = `{"stream":"observer","seq":0,"type":"run.opened","at_ns":0,"run_id":"r","mode":"evaluate","entity_ids":[]}`
	const observation = `{"stream":"observer","seq":1,"type":"observation","at_ns":1,"line":"x"}`
	for name, raw := range map[string]string{
		"committed":  opened + "\n" + observation + "\n",
		"torn tail":  opened + "\n" + `{"stream"`,
		"empty file": "",
	} {
		path := filepath.Join(t.TempDir(), "journal.jsonl")
		if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
			t.Fatal(err)
		}
		got, err := Scan(t.Context(), path, domain.ObserverJournal)
		if err != nil || !got.Present {
			t.Fatalf("%s: %+v (%v)", name, got, err)
		}
		scanner := domain.NewScanner(domain.ObserverJournal)
		for _, line := range strings.SplitAfter(raw, "\n") {
			if strings.HasSuffix(line, "\n") {
				scanner.Line([]byte(strings.TrimSuffix(line, "\n")))
			}
		}
		want := scanner.Finish(canonical.DigestBytes([]byte(raw)), raw != "" && !strings.HasSuffix(raw, "\n"))
		if !reflect.DeepEqual(want, got.Scan) {
			t.Errorf("%s: streamed scan %+v differs from the whole-buffer scan", name, got.Scan)
		}
	}
}

func TestScanOfAnAbsentFileIsNotPresent(t *testing.T) {
	t.Parallel()
	doc, err := Scan(t.Context(), filepath.Join(t.TempDir(), "absent.jsonl"), domain.LedgerRows)
	if err != nil || doc.Present {
		t.Fatalf("doc = %+v (%v), want not present and no error", doc, err)
	}
}

func TestScanRefusesALineOverTheBoundAndHonoursCancellation(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "ledger.jsonl")
	if err := os.WriteFile(path, []byte(strings.Repeat("x", maxLine+10)+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Scan(t.Context(), path, domain.LedgerRows); err == nil || !strings.Contains(err.Error(), "longer than") {
		t.Fatalf("err = %v, want the over-long line refused", err)
	}
	small := filepath.Join(t.TempDir(), "trace.jsonl")
	if err := os.WriteFile(small, []byte("a\nb\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	canceled, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := Scan(canceled, small, domain.DigestOnly); !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want the cancellation", err)
	}
}

func TestScanNamesAPathItCannotOpen(t *testing.T) {
	t.Parallel()
	if _, err := Scan(t.Context(), t.TempDir(), domain.DigestOnly); err == nil || !strings.HasPrefix(err.Error(), "report: ") {
		t.Fatalf("err = %v, want a report-prefixed failure scanning a directory", err)
	}
}

func TestScanNamesAFileItCannotOpen(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "ledger.jsonl")
	if err := os.WriteFile(path, []byte("row\n"), 0o000); err != nil {
		t.Fatal(err)
	}
	if file, err := os.Open(path); err == nil {
		_ = file.Close()
		t.Skip("the file opens despite mode 000 (running as root)")
	}
	if _, err := Scan(t.Context(), path, domain.LedgerRows); err == nil || !strings.HasPrefix(err.Error(), "report: ") {
		t.Fatalf("err = %v, want a report-prefixed open failure", err)
	}
}
