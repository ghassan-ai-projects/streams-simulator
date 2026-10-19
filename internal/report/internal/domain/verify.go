package domain

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func verify(folder Folder) Verification {
	checks := []Check{}
	if folder.Artifact != nil {
		for _, entry := range folder.Artifact.EvidenceDigests() {
			if entry.Digest != "" || !folder.Artifact.Incomplete {
				checks = append(checks, check(entry, folder.document(entry.File)))
			}
		}
	}
	return Verification{Status: statusOf(checks), Checks: checks}
}

func (f Folder) document(file string) Document {
	documents := map[string]Document{model.FileTrace: f.Trace, model.FileObserverJournal: f.Observer,
		model.FileDirectorJournal: f.Director, model.FileLedger: f.Ledger, model.FileFieldTrace: f.FieldTrace,
		model.FileInterruptLog: f.InterruptLog}
	if doc, known := documents[file]; known {
		return doc
	}
	return f.History
}

func check(entry model.EvidenceDigest, doc Document) Check {
	result := Check{Name: entry.Name, Recorded: entry.Digest, Status: CheckMissing}
	if doc.Present {
		result.Computed = computeDigest(entry, doc.scan(kindOf(entry.File)))
		result.Status = digestStatus(result)
	}
	if entry.Digest == "" {
		result.Status = CheckNoDigest
	}
	return result
}

func digestStatus(result Check) string {
	if result.Computed == result.Recorded {
		return CheckMatch
	}
	return CheckDiffers
}

func computeDigest(entry model.EvidenceDigest, scan *Scan) string {
	if entry.Replayable {
		return scan.ReplayDigest
	}
	return scan.RawDigest
}

func kindOf(file string) Kind {
	switch file {
	case model.FileObserverJournal:
		return ObserverJournal
	case model.FileDirectorJournal:
		return DirectorJournal
	case model.FileLedger:
		return LedgerRows
	}
	return DigestOnly
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
