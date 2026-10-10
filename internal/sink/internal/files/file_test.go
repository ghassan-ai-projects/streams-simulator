package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileWritesBufferedLinesAndReturnsThemAtClose(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "out", "trace.jsonl")
	f, err := New(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range []string{`{"a":1}`, `{"a":2}`} {
		if err := f.Write([]byte(line)); err != nil {
			t.Fatal(err)
		}
	}
	got, err := f.Close()
	if err != nil || string(got) != "{\"a\":1}\n{\"a\":2}\n" {
		t.Fatalf("close = %q (%v)", got, err)
	}
	if disk, err := os.ReadFile(path); err != nil || string(disk) != string(got) {
		t.Fatalf("disk = %q (%v)", disk, err)
	}
}

func TestNewNamesAPathItCannotCreate(t *testing.T) {
	t.Parallel()
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := New(filepath.Join(blocker, "child", "trace.jsonl")); err == nil || !strings.Contains(err.Error(), "sink: mkdir") {
		t.Fatalf("err = %v, want a mkdir failure", err)
	}
}

func TestFlushMakesBufferedLinesVisibleAndAClosedFileRefusesMore(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "trace.jsonl")
	f, err := New(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Write([]byte("a")); err != nil {
		t.Fatal(err)
	}
	if err := f.Flush(); err != nil {
		t.Fatal(err)
	}
	if disk, err := os.ReadFile(path); err != nil || string(disk) != "a\n" {
		t.Fatalf("after flush = %q (%v)", disk, err)
	}
	if _, err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Write([]byte("b")); err != nil {
		t.Fatalf("buffered write after close: %v", err)
	}
	if err := f.Flush(); err == nil || !strings.Contains(err.Error(), "sink: flush") {
		t.Fatalf("flush after close: %v", err)
	}
	if _, err := f.Close(); err == nil {
		t.Fatal("closing twice must fail")
	}
}
