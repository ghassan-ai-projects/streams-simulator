package model

type EvidenceDigest struct {
	Name       string
	Field      string
	File       string
	Digest     string
	Replayable bool
}

func (a *RunArtifact) EvidenceDigests() []EvidenceDigest {
	return []EvidenceDigest{
		{"trace", "expected_trace_digest", FileTrace, a.ExpectedTraceDigest, false},
		{"journal", "journal_digest", FileObserverJournal, a.JournalDigest, true},
		{"sealed_journal", "sealed_journal_digest", FileDirectorJournal, a.SealedJournalDigest, true},
		{"journal_file", "journal_file_digest", FileObserverJournal, a.JournalFileDigest, false},
		{"sealed_journal_file", "sealed_journal_file_digest", FileDirectorJournal, a.SealedFileDigest, false},
		{"ledger", "ledger_digest", FileLedger, a.LedgerDigest, false},
		{"history", "history_digest", FileHistory, a.HistoryDigest, false},
	}
}
