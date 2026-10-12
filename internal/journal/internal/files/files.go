// Package files is the file-system edge of the journal module: the two
// append-only JSON Lines files of a run, written through a buffer, flushed at
// command boundaries and synced when the run ends.
package files

import (
	"fmt"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/journal/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
)

func Open(dir string, retain bool) (*domain.Journal, error) {
	observer, err := openSink(filepath.Join(dir, domain.ObserverFile))
	if err != nil {
		return nil, err
	}
	director, err := openSink(filepath.Join(dir, domain.DirectorFile))
	if err != nil {
		_ = observer.Close()
		return nil, err
	}
	return domain.New(domain.NewStream(domain.StreamObserver, observer, retain),
		domain.NewStream(domain.StreamDirector, director, retain)), nil
}

func openSink(path string) (*kernel.Log, error) {
	log, err := kernel.NewLog(path, false)
	if err != nil {
		return nil, fmt.Errorf("journal: open %s: %w", path, err)
	}
	return log, nil
}
