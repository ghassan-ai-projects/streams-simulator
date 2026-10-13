package domain

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
)

type recorded struct {
	name    string
	digest  string
	doc     Document
	compute func([]byte) string
}

func verify(folder Folder) Verification {
	checks := []Check{}
	for _, entry := range recordedDigests(folder) {
		if entry.digest != "" {
			checks = append(checks, check(entry))
		}
	}
	return Verification{Status: statusOf(checks), Checks: checks}
}

func recordedDigests(folder Folder) []recorded {
	art := folder.Artifact
	if art == nil {
		return nil
	}
	return []recorded{
		{"trace", art.ExpectedTraceDigest, folder.Trace, canonical.DigestBytes},
		{"journal", art.JournalDigest, folder.Observer, journal.Digest},
		{"sealed_journal", art.SealedJournalDigest, folder.Director, journal.Digest},
		{"journal_file", art.JournalFileDigest, folder.Observer, canonical.DigestBytes},
		{"sealed_journal_file", art.SealedFileDigest, folder.Director, canonical.DigestBytes},
		{"ledger", art.LedgerDigest, folder.Ledger, canonical.DigestBytes},
	}
}

func check(entry recorded) Check {
	result := Check{Name: entry.name, Recorded: entry.digest, Status: CheckMissing}
	if !entry.doc.Present {
		return result
	}
	result.Computed = entry.compute(entry.doc.Raw)
	result.Status = CheckMatch
	if result.Computed != entry.digest {
		result.Status = CheckDiffers
	}
	return result
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
