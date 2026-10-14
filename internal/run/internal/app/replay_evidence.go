package app

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/durable"
)

type evidenceFile struct {
	model.EvidenceDigest
	replayedDigest string
	replayedRaw    func() []byte
	againstDigest  bool
}

func (f evidenceFile) digestOf(raw []byte) string {
	if f.Replayable {
		return journal.Digest(raw)
	}
	return canonical.DigestBytes(raw)
}

func evidenceFiles(replay *Run, art *model.RunArtifact) []evidenceFile {
	replayed := replay.replayedDigests()
	var files []evidenceFile
	for _, digest := range art.EvidenceDigests() {
		file := evidenceFile{EvidenceDigest: digest, replayedDigest: replayed[digest.Name],
			againstDigest: digest.Name == "ledger" || digest.Name == "history"}
		if stream, ok := evidenceStreams[digest.Name]; ok {
			file.replayedRaw = func() []byte { return replay.journalBytes(stream) }
		}
		files = append(files, file)
	}
	return files
}

var evidenceStreams = map[string]string{"journal": journal.StreamObserver, "sealed_journal": journal.StreamDirector}

func (r *Run) replayedDigests() map[string]string {
	if r.replayedDigest == nil {
		ledger, history := r.evidenceDigests()
		r.replayedDigest = map[string]string{"trace": r.traceDigest, "ledger": ledger, "history": history,
			"journal": r.journalDigests.Observer, "sealed_journal": r.journalDigests.Director}
	}
	return r.replayedDigest
}

func (r *Run) journalBytes(stream string) []byte {
	raw, _ := r.journal.Bytes(stream)
	return raw
}

func noteMissingDigests(art *model.RunArtifact, result *ReplayResult) {
	if art.Incomplete {
		return
	}
	for _, digest := range art.EvidenceDigests() {
		if digest.Digest == "" {
			diverge(result, ReasonDigestMissing, digest.File, fmt.Sprintf("run.json records no %s for %s", digest.Field, digest.File))
		}
	}
}

func noteEvidenceDivergence(replay *Run, art *model.RunArtifact, result *ReplayResult) {
	if art.Incomplete {
		return
	}
	for _, file := range evidenceFiles(replay, art) {
		if file.againstDigest && file.Digest != "" && file.Digest != file.replayedDigest {
			diverge(result, ReasonEvidenceDivergence, file.File, fmt.Sprintf("the %s differs from the replay (recorded %s, replayed %s)",
				file.Name, file.Digest, file.replayedDigest))
		}
	}
}

func checkEvidenceFiles(dir string, replay *Run, art *model.RunArtifact, result *ReplayResult) error {
	if art.Incomplete {
		return nil
	}
	for _, group := range groupByFile(evidenceFiles(replay, art)) {
		if err := checkEvidenceGroup(dir, group, result); err != nil {
			return err
		}
	}
	return nil
}

func groupByFile(files []evidenceFile) [][]evidenceFile {
	var groups [][]evidenceFile
	index := map[string]int{}
	for _, file := range files {
		at, seen := index[file.File]
		if !seen {
			at = len(groups)
			index[file.File] = at
			groups = append(groups, nil)
		}
		groups[at] = append(groups[at], file)
	}
	return groups
}

func checkEvidenceGroup(dir string, group []evidenceFile, result *ReplayResult) error {
	name := group[0].File
	onDisk, present, err := durable.ReadJournal(dir, name)
	if err != nil {
		return err
	}
	if !present {
		diverge(result, ReasonFileMissing, name, fmt.Sprintf("%s is missing from the run folder", name))
		return nil
	}
	for _, file := range group {
		checkEvidenceBytes(file, onDisk, result)
	}
	return nil
}

func checkEvidenceBytes(file evidenceFile, onDisk []byte, result *ReplayResult) {
	computed := file.digestOf(onDisk)
	if file.Digest != "" && computed != file.Digest {
		diverge(result, ReasonFileDiffers, file.File, fmt.Sprintf("%s differs from the digest run.json recorded as %s", file.File, file.Field))
	}
	if file.replayedDigest != "" && computed != file.replayedDigest {
		diverge(result, ReasonFileDiffers, file.File, differsFromReplay(file, onDisk))
		noteDivergedStream(file, result)
	}
}

func noteDivergedStream(file evidenceFile, result *ReplayResult) {
	if stream, ok := evidenceStreams[file.Name]; ok {
		result.JournalDiverged = append(result.JournalDiverged, stream)
	}
}

func differsFromReplay(file evidenceFile, onDisk []byte) string {
	if file.replayedRaw != nil {
		return fmt.Sprintf("%s differs from the replay at record %d", file.File, journal.FirstDifference(onDisk, file.replayedRaw()))
	}
	return fmt.Sprintf("%s differs from the replay", file.File)
}
