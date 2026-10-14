package domain

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func verify(folder Folder) Verification {
	checks := []Check{}
	if folder.Artifact != nil {
		for _, entry := range folder.Artifact.EvidenceDigests() {
			if entry.Digest != "" {
				checks = append(checks, check(entry, folder.document(entry.File)))
			}
		}
	}
	return Verification{Status: statusOf(checks), Checks: checks}
}

func (f Folder) document(file string) Document {
	switch file {
	case model.FileTrace:
		return f.Trace
	case model.FileObserverJournal:
		return f.Observer
	case model.FileDirectorJournal:
		return f.Director
	case model.FileLedger:
		return f.Ledger
	}
	return f.History
}

func check(entry model.EvidenceDigest, doc Document) Check {
	result := Check{Name: entry.Name, Recorded: entry.Digest, Status: CheckMissing}
	if !doc.Present {
		return result
	}
	result.Computed = computeDigest(entry, doc.Raw)
	result.Status = CheckMatch
	if result.Computed != entry.Digest {
		result.Status = CheckDiffers
	}
	return result
}

func computeDigest(entry model.EvidenceDigest, raw []byte) string {
	if entry.Replayable {
		return journal.Digest(raw)
	}
	return canonical.DigestBytes(raw)
}

func statusOf(checks []Check) string {
	if len(checks) == 0 {
		return NotChecked
	}
	for _, c := range checks {
		if c.Status != CheckMatch {
			return FilesDiffer
		}
	}
	return FilesMatch
}
