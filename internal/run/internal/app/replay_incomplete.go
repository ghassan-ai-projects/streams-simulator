package app

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func concludeIncomplete(art *model.RunArtifact, result *ReplayResult) {
	if art.Incomplete && art.Reproducible {
		diverge(result, ReasonIncompleteStamp, "", "run.json says the run ended incomplete and reproducible; a run that ended incomplete is never reproducible")
	}
	if art.Incomplete {
		diverge(result, ReasonIncomplete, "", "the run ended incomplete, so it cannot be verified as complete")
	}
}

func noteClosingStamp(file string, raw []byte, art *model.RunArtifact, result *ReplayResult) {
	if file != model.FileObserverJournal && file != model.FileDirectorJournal {
		return
	}
	incomplete, closed := journal.ClosedIncomplete(raw)
	if closed && incomplete != art.Incomplete {
		diverge(result, ReasonIncompleteStamp, file,
			fmt.Sprintf("%s closed with incomplete=%v, run.json says incomplete=%v", file, incomplete, art.Incomplete))
	}
}
