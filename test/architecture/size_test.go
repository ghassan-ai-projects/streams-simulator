package architecture

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"testing"
)

const maxGoFileLines = 300

func TestGoFileSize(t *testing.T) {
	t.Parallel()
	root, err := os.OpenRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := root.Close(); err != nil {
			t.Error(err)
		}
	}()
	err = fs.WalkDir(root.FS(), ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("inspect Go file %s: %w", path, err)
		}
		if entry.IsDir() && entry.Name() == ".git" {
			return fs.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		source, err := root.ReadFile(path)
		if err != nil {
			return fmt.Errorf("inspect Go file %s: %w", path, err)
		}
		if lines := sourceLines(source); lines > maxGoFileLines {
			t.Errorf("%s has %d lines; maximum is %d: split by responsibility", path, lines, maxGoFileLines)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func sourceLines(source []byte) int {
	lines := bytes.Count(source, []byte{'\n'})
	if len(source) > 0 && source[len(source)-1] != '\n' {
		lines++
	}
	return lines
}

func TestSourceLinesCountsCommentsBlanksAndFinalLine(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, source string
		want         int
	}{
		{"empty", "", 0},
		{"newline", "\n", 1},
		{"unterminated", "package example", 1},
		{"comments and blanks", "// comment\n\npackage example\n", 3},
		{"final line", "// comment\npackage example", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := sourceLines([]byte(tc.source)); got != tc.want {
				t.Fatalf("got %d lines, want %d", got, tc.want)
			}
		})
	}
}
