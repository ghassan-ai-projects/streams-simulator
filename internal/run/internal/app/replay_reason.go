package app

const (
	ReasonMatch              = "match"
	ReasonRecordedSpec       = "recorded_spec"
	ReasonWorldDigest        = "world_digest"
	ReasonTraceDivergence    = "trace_divergence"
	ReasonJournalDivergence  = "journal_divergence"
	ReasonEvidenceDivergence = "evidence_divergence"
	ReasonUnblindedStamp     = "unblinded_stamp"
	ReasonDigestMissing      = "digest_missing"
	ReasonFileDiffers        = "file_differs"
	ReasonFileMissing        = "file_missing"
)

func diverge(result *ReplayResult, reason, file, note string) {
	if result.Matches {
		result.Reason, result.File = reason, file
	}
	result.Matches = false
	result.Detail = divergenceDetail(result.Detail, note)
}

func divergenceDetail(existing, note string) string {
	if existing == "" {
		return note
	}
	return existing + "; " + note
}

func brokenSpecReason(subject string) string {
	if subject == "world" {
		return ReasonWorldDigest
	}
	return ReasonRecordedSpec
}
