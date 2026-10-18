package files

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestReadRegularReadsAFileWithinItsLimitAndRefusesTheRest(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	regular := filepath.Join(dir, "regular")
	if err := os.WriteFile(regular, []byte("1234"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link")
	pipe := filepath.Join(dir, "pipe")
	if err := errors.Join(os.Symlink(regular, link), syscall.Mkfifo(pipe, 0o600)); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name  string
		path  string
		limit int64
		want  error
		text  string
	}{
		{"a file at its limit", regular, 4, nil, "1234"},
		{"a file over its limit", regular, 3, ErrTooLarge, ""},
		{"a symbolic link", link, 100, ErrNotRegular, ""},
		{"a named pipe, which must not be opened", pipe, 100, ErrNotRegular, ""},
		{"a directory", dir, 100, ErrNotRegular, ""},
		{"a missing file", filepath.Join(dir, "none"), 100, fs.ErrNotExist, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			raw, err := ReadRegular(tc.path, tc.limit)
			if !errors.Is(err, tc.want) && tc.want != nil || tc.want == nil && err != nil {
				t.Fatalf("ReadRegular = %v, want %v", err, tc.want)
			}
			if string(raw) != tc.text {
				t.Fatalf("ReadRegular = %q, want %q", raw, tc.text)
			}
		})
	}
}
