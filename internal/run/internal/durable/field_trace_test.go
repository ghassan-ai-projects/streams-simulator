package durable

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestAFieldTraceFileKeepsWhatItWasGivenAndReadsItBack(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	file, err := OpenEvidenceLog(dir, model.FileFieldTrace, NounFieldTrace)
	if err != nil {
		t.Fatal(err)
	}
	if err := errors.Join(file.Append([]byte("one\n")), file.Append([]byte("two\n")), file.Flush()); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	raw, present, err := ReadEvidenceLog(dir, model.FileFieldTrace, NounFieldTrace, 64)
	if err != nil || !present || string(raw) != "one\ntwo\n" {
		t.Fatalf("ReadFieldTrace = %q, %v, %v, want both lines present", raw, present, err)
	}
}

func TestAFolderWithoutAFieldTraceReadsAsAbsentNotAsAFailure(t *testing.T) {
	t.Parallel()
	raw, present, err := ReadEvidenceLog(t.TempDir(), model.FileFieldTrace, NounFieldTrace, 64)
	if err != nil || present || raw != nil {
		t.Fatalf("ReadFieldTrace = %q, %v, %v, want absent and no error", raw, present, err)
	}
}

func TestAFieldTraceThatCannotBeReadIsNamedAsSuch(t *testing.T) {
	t.Parallel()
	tooLarge := t.TempDir()
	if err := os.WriteFile(filepath.Join(tooLarge, model.FileFieldTrace), []byte("12345"), 0o600); err != nil {
		t.Fatal(err)
	}
	notAFile := t.TempDir()
	if err := os.Mkdir(filepath.Join(notAFile, model.FileFieldTrace), 0o700); err != nil {
		t.Fatal(err)
	}
	for name, dir := range map[string]string{"over the limit": tooLarge, "not a regular file": notAFile} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, _, err := ReadEvidenceLog(dir, model.FileFieldTrace, NounFieldTrace, 4); err == nil || !strings.HasPrefix(err.Error(), "run: read field trace: ") {
				t.Fatalf("ReadFieldTrace = %v, want a refusal that names the field trace", err)
			}
		})
	}
}

func TestAFieldTraceFileNamesAFolderItCannotCreateAndAFileItCannotWriteAfterClosing(t *testing.T) {
	t.Parallel()
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenEvidenceLog(blocker, model.FileFieldTrace, NounFieldTrace); err == nil || !strings.HasPrefix(err.Error(), "run: open field trace in ") {
		t.Fatalf("OpenFieldTrace under a file = %v, want a refusal that names the folder", err)
	}
	file, err := OpenEvidenceLog(t.TempDir(), model.FileFieldTrace, NounFieldTrace)
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); !errors.Is(err, os.ErrClosed) {
		t.Fatalf("a second Close = %v, want %v", err, os.ErrClosed)
	}
	if err := file.Append(make([]byte, 1<<16)); !errors.Is(err, os.ErrClosed) {
		t.Fatalf("Append after Close = %v, want %v", err, os.ErrClosed)
	}
	if err := file.Flush(); !errors.Is(err, os.ErrClosed) {
		t.Fatalf("Flush of what a closed file could not take = %v, want %v", err, os.ErrClosed)
	}
}
