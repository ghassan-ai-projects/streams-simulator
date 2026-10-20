package domain

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var ErrRunFolderTaken = errors.New("plantwire: the run folder already holds a run; every session needs a new folder")

var runFiles = []string{model.FileRunArtifact, model.FileTrace, model.FileLedger, model.FileHistory, model.FileFieldTrace,
	model.FileInterruptLog, model.FileInterruptLogRejected, model.FileObserverJournal, model.FileDirectorJournal,
	model.FileVerdict, model.FileLabel}

func CheckFreshRunFolder(dir string, present []string) error {
	for _, name := range slices.Sorted(slices.Values(present)) {
		if isRunFile(name) {
			return fmt.Errorf("%w: %s holds %s", ErrRunFolderTaken, dir, name)
		}
	}
	return nil
}

func isRunFile(name string) bool {
	rejected := strings.HasPrefix(name, "interrupts.rejected.") && strings.HasSuffix(name, ".jsonl")
	return rejected || slices.Contains(runFiles, name)
}
