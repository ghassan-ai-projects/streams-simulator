package files

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

func (w *Workspace) WriteIfUnchanged(rel string, content []byte, expected string, version func([]byte) string, limit int64) (string, error) {
	current, err := w.versionOf(rel, version, limit)
	if err != nil {
		return "", err
	}
	if current != expected {
		return current, fmt.Errorf("%w: %s", domain.ErrConflict, rel)
	}
	if err := w.writeAtomic(rel, content); err != nil {
		return "", err
	}
	return version(content), nil
}

func (w *Workspace) versionOf(rel string, version func([]byte) string, limit int64) (string, error) {
	content, err := w.Read(rel, limit)
	if errors.Is(err, domain.ErrAbsent) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return version(content), nil
}

func (w *Workspace) writeAtomic(rel string, content []byte) error {
	if err := w.root.MkdirAll(path.Dir(rel), 0o750); err != nil {
		return fmt.Errorf("canvasbridge: create %s: %w", path.Dir(rel), err)
	}
	temporary := path.Join(path.Dir(rel), "."+path.Base(rel)+".tmp-"+strconv.Itoa(os.Getpid()))
	if err := w.writeFile(temporary, content); err != nil {
		return err
	}
	if err := w.root.Rename(temporary, rel); err != nil {
		_ = w.root.Remove(temporary)
		return fmt.Errorf("canvasbridge: replace %s: %w", rel, err)
	}
	return nil
}

func (w *Workspace) writeFile(rel string, content []byte) error {
	file, err := w.root.OpenFile(rel, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("canvasbridge: write %s: %w", rel, err)
	}
	_, writeErr := file.Write(content)
	syncErr := file.Sync()
	closeErr := file.Close()
	if err := errors.Join(writeErr, syncErr, closeErr); err != nil {
		_ = w.root.Remove(rel)
		return fmt.Errorf("canvasbridge: write %s: %w", rel, err)
	}
	return nil
}

func (w *Workspace) Fingerprint(dirs []string) string {
	var b strings.Builder
	for _, dir := range dirs {
		entries, _ := fs.ReadDir(w.root.FS(), dir)
		for _, entry := range entries {
			if info, err := entry.Info(); err == nil {
				fmt.Fprintf(&b, "%s/%s:%d:%d;", dir, entry.Name(), info.Size(), info.ModTime().UnixNano())
			}
		}
	}
	return b.String()
}

func (w *Workspace) Ensure(dirs []string) error {
	for _, dir := range dirs {
		if err := w.root.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("canvasbridge: create %s: %w", dir, err)
		}
	}
	return nil
}
