package model

import "regexp"

type EvidenceDigest struct {
	Name       string
	Field      string
	File       string
	Digest     string
	Replayable bool
}

var wellFormedDigest = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

type evidenceSlot struct {
	EvidenceDigest
	value *string
}

const fieldTraceEvidence = "field_trace"

func (a *RunArtifact) evidenceSlots() []evidenceSlot {
	slots := a.recordedSlots()
	if a.FieldTraceDigest != "" {
		slots = append(slots, evidenceSlot{EvidenceDigest{fieldTraceEvidence, "field_trace_digest", FileFieldTrace, "", false}, &a.FieldTraceDigest})
	}
	return slots
}

func (a *RunArtifact) recordedSlots() []evidenceSlot {
	return []evidenceSlot{
		{EvidenceDigest{"trace", "expected_trace_digest", FileTrace, "", false}, &a.ExpectedTraceDigest},
		{EvidenceDigest{"journal", "journal_digest", FileObserverJournal, "", true}, &a.JournalDigest},
		{EvidenceDigest{"sealed_journal", "sealed_journal_digest", FileDirectorJournal, "", true}, &a.SealedJournalDigest},
		{EvidenceDigest{"journal_file", "journal_file_digest", FileObserverJournal, "", false}, &a.JournalFileDigest},
		{EvidenceDigest{"sealed_journal_file", "sealed_journal_file_digest", FileDirectorJournal, "", false}, &a.SealedFileDigest},
		{EvidenceDigest{"ledger", "ledger_digest", FileLedger, "", false}, &a.LedgerDigest},
		{EvidenceDigest{"history", "history_digest", FileHistory, "", false}, &a.HistoryDigest},
	}
}

func (a *RunArtifact) EvidenceDigests() []EvidenceDigest {
	var digests []EvidenceDigest
	for _, slot := range a.evidenceSlots() {
		slot.Digest = *slot.value
		digests = append(digests, slot.EvidenceDigest)
	}
	return digests
}

func (a *RunArtifact) DropMalformedEvidenceDigests() {
	for _, slot := range a.evidenceSlots() {
		if slot.Name != fieldTraceEvidence && !wellFormedDigest.MatchString(*slot.value) {
			*slot.value = ""
		}
	}
}
