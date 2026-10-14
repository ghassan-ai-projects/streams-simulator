package model

import (
	"encoding/json"
	"testing"
)

func TestEvidenceDigestsNameEveryDigestFieldTheArtifactRecords(t *testing.T) {
	t.Parallel()
	art := &RunArtifact{
		ExpectedTraceDigest: "t", JournalDigest: "j", SealedJournalDigest: "s", JournalFileDigest: "jf",
		SealedFileDigest: "sf", LedgerDigest: "l", HistoryDigest: "h",
	}
	raw, err := json.Marshal(art)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"trace": FileTrace, "journal": FileObserverJournal, "sealed_journal": FileDirectorJournal,
		"journal_file": FileObserverJournal, "sealed_journal_file": FileDirectorJournal, "ledger": FileLedger, "history": FileHistory}
	digests := art.EvidenceDigests()
	if len(digests) != len(want) {
		t.Fatalf("digests = %+v, want %d entries", digests, len(want))
	}
	for _, digest := range digests {
		if digest.File != want[digest.Name] || digest.Digest == "" || doc[digest.Field] != digest.Digest {
			t.Errorf("%+v: want file %q and the digest recorded under its field in run.json", digest, want[digest.Name])
		}
	}
}
