package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestVerifyComparesRecordedDigestsWithTheFiles(t *testing.T) {
	t.Parallel()
	trace := Document{Raw: []byte("trace\n"), Present: true}
	observer := doc(openedEvaluate)
	artifact := func(trace, observer, sealed string) *model.RunArtifact {
		return &model.RunArtifact{ExpectedTraceDigest: trace, JournalDigest: observer, SealedJournalDigest: sealed}
	}
	cases := []struct {
		name   string
		folder Folder
		status string
		checks []string
	}{
		{"no artifact", Folder{}, NotChecked, nil},
		{"nothing recorded", Folder{Artifact: artifact("", "", "")}, NotChecked, nil},
		{"all match", Folder{Artifact: artifact(canonical.DigestBytes(trace.Raw), journal.Digest(observer.Raw), ""), Trace: trace, Observer: observer},
			FilesMatch, []string{CheckMatch, CheckMatch}},
		{"edited trace", Folder{Artifact: artifact("sha256:00", "", ""), Trace: trace}, FilesDiffer, []string{CheckDiffers}},
		{"missing sealed journal", Folder{Artifact: artifact("", "", "sha256:00")}, FilesDiffer, []string{CheckMissing}},
		{"edited ledger", Folder{Artifact: &model.RunArtifact{LedgerDigest: canonical.DigestBytes([]byte("row\n"))}, Ledger: Document{Raw: []byte("other\n"), Present: true}},
			FilesDiffer, []string{CheckDiffers}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := verify(tc.folder)
			if got.Status != tc.status || len(got.Checks) != len(tc.checks) {
				t.Fatalf("verification = %+v, want %s with %v", got, tc.status, tc.checks)
			}
			for i, status := range tc.checks {
				if got.Checks[i].Status != status {
					t.Errorf("check %d = %+v, want %s", i, got.Checks[i], status)
				}
			}
		})
	}
}
