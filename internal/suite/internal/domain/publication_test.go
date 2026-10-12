package domain

import "testing"

func TestFileNameNamesTheDomainAndProfile(t *testing.T) {
	t.Parallel()
	if got := FileName("aquaculture-pond", "nominal"); got != "aquaculture-pond-nominal-suite.json" {
		t.Fatalf("FileName = %q", got)
	}
}

func TestDefaultSampleIsTwoMinutes(t *testing.T) {
	t.Parallel()
	if DefaultSampleNS != 120_000_000_000 {
		t.Fatalf("DefaultSampleNS = %d, want 120 s in ns", DefaultSampleNS)
	}
}
