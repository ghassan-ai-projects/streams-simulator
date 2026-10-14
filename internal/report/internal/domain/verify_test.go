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
		return &model.RunArtifact{Incomplete: true, ExpectedTraceDigest: trace, JournalDigest: observer, SealedJournalDigest: sealed}
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
		{"edited history", Folder{Artifact: &model.RunArtifact{Incomplete: true, HistoryDigest: canonical.DigestBytes([]byte("row\n"))}, History: Document{Raw: []byte("other\n"), Present: true}},
			FilesDiffer, []string{CheckDiffers}},
		{"missing history", Folder{Artifact: &model.RunArtifact{Incomplete: true, HistoryDigest: canonical.DigestBytes([]byte("row\n"))}}, FilesDiffer, []string{CheckMissing}},
		{"edited ledger", Folder{Artifact: &model.RunArtifact{Incomplete: true, LedgerDigest: canonical.DigestBytes([]byte("row\n"))}, Ledger: Document{Raw: []byte("other\n"), Present: true}},
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

func TestAChangedSessionRecordShowsInTheJournalFileCheckButNotInTheReplayableDigest(t *testing.T) {
	t.Parallel()
	opened := `{"stream":"observer","seq":0,"type":"run.opened","at_ns":0,"run_id":"r-1"}`
	written := []byte(opened + "\n" + `{"stream":"observer","seq":1,"type":"pacing","at_ns":0,"through_seq":0,"waited_ms":5,"missed":false}` + "\n")
	edited := []byte(opened + "\n" + `{"stream":"observer","seq":1,"type":"pacing","at_ns":0,"through_seq":0,"waited_ms":0,"missed":false}` + "\n")
	art := &model.RunArtifact{Incomplete: true, JournalDigest: journal.Digest(written), JournalFileDigest: canonical.DigestBytes(written)}
	for name, raw := range map[string][]byte{"as written": written, "a pacing record edited": edited} {
		got := verify(Folder{Artifact: art, Observer: Document{Raw: raw, Present: true}})
		want := []string{CheckMatch, CheckMatch}
		if name != "as written" {
			want = []string{CheckMatch, CheckDiffers}
		}
		if len(got.Checks) != 2 || got.Checks[0].Status != want[0] || got.Checks[1].Status != want[1] || got.Checks[1].Name != "journal_file" {
			t.Errorf("%s: checks %+v, want %v", name, got.Checks, want)
		}
	}
}

func TestACompleteRunMustRecordADigestForEveryEvidenceFile(t *testing.T) {
	t.Parallel()
	trace := Document{Raw: []byte("trace\n"), Present: true}
	recorded := &model.RunArtifact{ExpectedTraceDigest: canonical.DigestBytes(trace.Raw)}
	got := verify(Folder{Artifact: recorded, Trace: trace})
	if got.Status != FilesDiffer || len(got.Checks) != 7 || got.Checks[0].Status != CheckMatch {
		t.Fatalf("verification = %+v, want files_differ, the trace matching", got)
	}
	for _, c := range got.Checks[1:] {
		if c.Status != CheckNoDigest || c.Recorded != "" {
			t.Errorf("check %+v, want digest_missing with nothing recorded", c)
		}
	}
	recorded.Incomplete = true
	if got := verify(Folder{Artifact: recorded, Trace: trace}); got.Status != FilesMatch || len(got.Checks) != 1 {
		t.Fatalf("an incomplete run is not held to the closed-run rule: %+v", got)
	}
}
