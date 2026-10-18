package files

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func readAll(t *testing.T, file io.ReadCloser) string {
	t.Helper()
	raw, err := io.ReadAll(file)
	if closed := file.Close(); err != nil || closed != nil {
		t.Fatalf("read %v, close %v", err, closed)
	}
	return string(raw)
}

func TestOpenRegularInRefusesASymbolicLinkAtAnyComponentOfTheRelativePath(t *testing.T) {
	t.Parallel()
	base, elsewhere := t.TempDir(), t.TempDir()
	journal := filepath.Join(base, "runs", "w-1", "journal.jsonl")
	if err := errors.Join(os.MkdirAll(filepath.Dir(journal), 0o750), os.WriteFile(journal, []byte("ok"), 0o600),
		os.WriteFile(filepath.Join(elsewhere, "journal.jsonl"), []byte("stolen"), 0o600), os.Symlink(elsewhere, filepath.Join(base, "runs", "w-2")),
		os.Symlink(journal, filepath.Join(base, "runs", "w-1", "link.jsonl")), syscall.Mkfifo(filepath.Join(base, "runs", "w-1", "pipe"), 0o600)); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, rel string
		want      error
	}{
		{"a folder that is a link", "runs/w-2/journal.jsonl", ErrSymlink},
		{"a file that is a link", "runs/w-1/link.jsonl", ErrSymlink},
		{"a named pipe, which must not be opened", "runs/w-1/pipe", ErrNotRegular},
		{"a directory", "runs/w-1", ErrNotRegular},
		{"a missing file", "runs/w-1/none", fs.ErrNotExist},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if file, err := OpenRegularIn(base, tc.rel); file != nil || !errors.Is(err, tc.want) {
				t.Fatalf("OpenRegularIn(%s) = %v, %v, want %v", tc.rel, file, err, tc.want)
			}
		})
	}
	file, err := OpenRegularIn(base, "runs/w-1/journal.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	if got := readAll(t, file); got != "ok" {
		t.Fatalf("OpenRegularIn read %q, want the file", got)
	}
	if _, err := OpenRegularIn(base, "../journal.jsonl"); err == nil || !strings.Contains(err.Error(), "escapes") {
		t.Fatalf("OpenRegularIn of a path that climbs out = %v, want a refusal saying it escapes", err)
	}
}

func TestOpenRegularChecksOnlyTheLastComponentOfAnAbsolutePath(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	regular, link := filepath.Join(dir, "journal.jsonl"), filepath.Join(dir, "link.jsonl")
	if err := errors.Join(os.WriteFile(regular, []byte("ok"), 0o600), os.Symlink(regular, link)); err != nil {
		t.Fatal(err)
	}
	file, err := OpenRegular(regular)
	if err != nil {
		t.Fatal(err)
	}
	if got := readAll(t, file); got != "ok" {
		t.Fatalf("OpenRegular read %q, want the file", got)
	}
	own, err := OpenRegular("open_regular_test.go")
	if err != nil {
		t.Fatalf("OpenRegular of a relative path = %v, want the file beside the test", err)
	}
	if got := readAll(t, own); !strings.HasPrefix(got, "package files") {
		t.Fatalf("OpenRegular of a relative path read %q, want this source file", got[:min(len(got), 20)])
	}
	for path, want := range map[string]error{link: ErrNotRegular, dir: ErrNotRegular, filepath.Join(dir, "none"): fs.ErrNotExist} {
		if _, err := OpenRegular(path); !errors.Is(err, want) {
			t.Errorf("OpenRegular(%s) = %v, want %v", path, err, want)
		}
	}
}

func TestOpenRegularInRefusesABaseThatDoesNotExist(t *testing.T) {
	t.Parallel()
	if _, err := OpenRegularIn(filepath.Join(t.TempDir(), "absent"), "x"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("OpenRegularIn of a missing base = %v, want fs.ErrNotExist", err)
	}
}
