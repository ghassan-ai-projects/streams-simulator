package app

import (
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestLoadDirectorServesEveryShippedDomainAndAdapter(t *testing.T) {
	t.Parallel()
	out := t.TempDir()
	d, err := LoadDirector(t.Context(), testsupport.DomainsDir(), testsupport.AdaptersDir(), out, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if d.library == nil || d.OutDir != out || len(d.Adapters) < 2 || len(d.domains.List("")) < 2 {
		t.Fatalf("director = out %q, %d adapters, %d domains", d.OutDir, len(d.Adapters), len(d.domains.List("")))
	}
}

func TestLoadDirectorNamesTheInputItCannotLoad(t *testing.T) {
	t.Parallel()
	empty := t.TempDir()
	_, err := LoadDirector(t.Context(), empty, testsupport.AdaptersDir(), "", nil, nil)
	if err == nil || !strings.HasSuffix(err.Error(), "no domains found in "+empty) {
		t.Fatalf("an empty domains directory: err = %v", err)
	}
	_, err = LoadDirector(t.Context(), testsupport.DomainsDir(), empty, "", nil, nil)
	if err == nil || !strings.HasSuffix(err.Error(), "no adapters found in "+empty) {
		t.Fatalf("an empty adapters directory: err = %v", err)
	}
}
