package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
