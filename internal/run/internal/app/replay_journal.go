package app

import (
	"fmt"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/durable"
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
	result.JournalDiverged = append(result.JournalDiverged, stream)
	result.Detail = divergenceDetail(result.Detail, note)
	result.Matches = false
}

func checkJournalFiles(dir string, replay *Run, art *model.RunArtifact, result *ReplayResult) error {
	if art.Incomplete {
		return nil
	}
	for _, stream := range journalStreams(replay, art) {
		if err := checkJournalFile(dir, stream, replay, result); err != nil {
			return err
		}
	}
	return nil
}

func checkJournalFile(dir string, stream journalStream, replay *Run, result *ReplayResult) error {
	onDisk, present, err := durable.ReadJournal(dir, stream.file)
	if err != nil || !present {
		return err
	}
	rebuilt, err := replay.journal.Bytes(stream.name)
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}
	if journal.Digest(onDisk) != journal.Digest(rebuilt) {
		at := journal.FirstDifference(onDisk, rebuilt)
		divergeJournal(result, stream.name, fmt.Sprintf("%s differs from the replay at record %d", filepath.ToSlash(stream.file), at))
	}
	return nil
}
