package app

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/durable"
)

type evidenceFile struct {
	name, file, recorded, replayed string
}

func evidenceFiles(replay *Run, art *model.RunArtifact) []evidenceFile {
	ledger, history := replay.evidenceDigests()
	return []evidenceFile{
		{"ledger", model.FileLedger, art.LedgerDigest, ledger},
		{"history", model.FileHistory, art.HistoryDigest, history},
	}
}

func checkEvidenceFiles(dir string, replay *Run, art *model.RunArtifact, result *ReplayResult) error {
	if art.Incomplete {
		return nil
	}
	for _, file := range evidenceFiles(replay, art) {
		if err := checkEvidenceFile(dir, file, result); err != nil {
			return err
		}
	}
	return nil
}

func checkEvidenceFile(dir string, file evidenceFile, result *ReplayResult) error {
	if file.recorded == "" {
		return nil
	}
	if file.recorded != file.replayed {
		diverge(result, fmt.Sprintf("the %s differs from the replay (recorded %s, replayed %s)", file.name, file.recorded, file.replayed))
	}
	return checkFileDigest(dir, file, result)
}

func checkFileDigest(dir string, file evidenceFile, result *ReplayResult) error {
	onDisk, present, err := durable.ReadJournal(dir, file.file)
	if err != nil || !present {
		return err
	}
	if canonical.DigestBytes(onDisk) != file.recorded {
		diverge(result, fmt.Sprintf("%s differs from the digest run.json recorded", file.file))
	}
	return nil
}

func diverge(result *ReplayResult, note string) {
	result.Detail = divergenceDetail(result.Detail, note)
	result.Matches = false
}
