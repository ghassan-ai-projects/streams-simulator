// Package files is the workspace edge of the canvas bridge: every read and
// write of a workspace file goes through one os.Root, so no path leaves the
// workspace (no traversal, no symlink escape). Writes are atomic and
// conditional on the version the edit started from.
package files

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

type Workspace struct {
	root *os.Root
	name string
}

func Open(dir string) (*Workspace, error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("canvasbridge: open workspace %s: %w", dir, err)
	}
	return &Workspace{root: root, name: path.Base(strings.TrimRight(dir, "/"))}, nil
}

func (w *Workspace) Name() string { return w.name }

func (w *Workspace) Close() error {
	if err := w.root.Close(); err != nil {
		return fmt.Errorf("canvasbridge: close workspace: %w", err)
	}
	return nil
}

func (w *Workspace) Read(rel string, limit int64) ([]byte, error) {
	file, err := w.root.Open(rel)
	if err != nil {
		return nil, absent(rel, err)
	}
	defer func() { _ = file.Close() }()
	content, err := io.ReadAll(io.LimitReader(file, limit+1))
	switch {
	case err != nil:
		return nil, fmt.Errorf("canvasbridge: read %s: %w", rel, err)
	case int64(len(content)) > limit:
		return nil, fmt.Errorf("%w: %s", domain.ErrTooLarge, rel)
	}
	return content, nil
}

func (w *Workspace) ReadFrom(rel string, offset int64, limit int64) ([]byte, error) {
	file, err := w.root.Open(rel)
	if err != nil {
		return nil, absent(rel, err)
	}
	defer func() { _ = file.Close() }()
	content, err := io.ReadAll(io.NewSectionReader(file, offset, limit))
	if err != nil {
		return nil, fmt.Errorf("canvasbridge: read %s: %w", rel, err)
	}
	return content, nil
}

func (w *Workspace) Exists(rel string) bool {
	_, err := w.root.Stat(rel)
	return err == nil
}

func (w *Workspace) Modified(rel string) (time.Time, error) {
	info, err := w.root.Stat(rel)
	if err != nil {
		return time.Time{}, absent(rel, err)
	}
	return info.ModTime(), nil
}

func (w *Workspace) Names(dir string) ([]string, error) {
	entries, err := fs.ReadDir(w.root.FS(), dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("canvasbridge: list %s: %w", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	slices.Sort(names)
	return names, nil
}

func (w *Workspace) Tail(rel string, limit int64) ([]byte, error) {
	file, err := w.root.Open(rel)
	if err != nil {
		return nil, absent(rel, err)
	}
	defer func() { _ = file.Close() }()
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("canvasbridge: stat %s: %w", rel, err)
	}
	return readSection(file, rel, max(0, info.Size()-limit), limit)
}

func readSection(file *os.File, rel string, offset, limit int64) ([]byte, error) {
	content, err := io.ReadAll(io.NewSectionReader(file, offset, limit))
	if err != nil {
		return nil, fmt.Errorf("canvasbridge: read %s: %w", rel, err)
	}
	return content, nil
}

const escapeMessage = "path escapes from parent"

func absent(rel string, err error) error {
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("%w: %s", domain.ErrAbsent, rel)
	}
	if strings.Contains(err.Error(), escapeMessage) {
		return fmt.Errorf("%w: %s leaves the workspace", domain.ErrBadPath, rel)
	}
	return fmt.Errorf("canvasbridge: open %s: %w", rel, err)
}

func Absolute(dir string) (string, error) {
	absolute, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("canvasbridge: resolve %s: %w", dir, err)
	}
	return absolute, nil
}

const maxLinkDepth = 32

func Aliases(dir string) []string {
	resolved := resolveLinks(filepath.Clean(dir), 0)
	if resolved == dir {
		return nil
	}
	return []string{resolved}
}

func resolveLinks(path string, depth int) string {
	resolved := string(filepath.Separator)
	for _, name := range strings.Split(path, string(filepath.Separator)) {
		if name != "" {
			resolved = followLink(resolved, name, depth)
		}
	}
	return resolved
}

func followLink(parent, name string, depth int) string {
	next := filepath.Join(parent, name)
	target, err := os.Readlink(next)
	if err != nil || depth >= maxLinkDepth {
		return next
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(parent, target)
	}
	return resolveLinks(target, depth+1)
}
