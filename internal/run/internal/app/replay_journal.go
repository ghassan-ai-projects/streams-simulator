package app

import (
	"fmt"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type journalStream struct {
	name, file, recorded, replayed string
}

func journalStreams(replay *Run, art *model.RunArtifact) []journalStream {
	return []journalStream{
		{journal.StreamObserver, journal.ObserverFile, art.JournalDigest, replay.journalDigests.Observer},
		{journal.StreamDirector, journal.DirectorFile, art.SealedJournalDigest, replay.journalDigests.Director},
	}
}

func noteJournalDivergence(replay *Run, art *model.RunArtifact, result *ReplayResult) {
	if art.Incomplete {
		return
	}
	for _, stream := range journalStreams(replay, art) {
		if stream.recorded != "" && stream.recorded != stream.replayed {
			note := fmt.Sprintf("the %s journal stream differs from the replay (recorded %s, replayed %s)", stream.name, stream.recorded, stream.replayed)
			divergeJournal(result, stream.name, note)
		}
	}
}

func divergeJournal(result *ReplayResult, stream, note string) {
	noteDivergedJournal(result, stream)
	diverge(result, ReasonJournalDivergence, "", note)
}

func noteDivergedJournal(result *ReplayResult, stream string) {
	if !slices.Contains(result.JournalDiverged, stream) {
		result.JournalDiverged = append(result.JournalDiverged, stream)
	}
}
