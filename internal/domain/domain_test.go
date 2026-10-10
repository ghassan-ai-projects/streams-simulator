package domain_test

import (
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func shipped(t *testing.T, id string) *domain.Compiled {
	t.Helper()
	c, err := domain.Load(testsupport.Domain(id))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestCatalogListsDescribesAndReportsCoverage(t *testing.T) {
	t.Parallel()
	pond, pumps := shipped(t, "aquaculture-pond"), shipped(t, "rotating-machinery")
	catalog := domain.NewCatalog([]*domain.Compiled{pond, pumps})
	if all := catalog.List(""); len(all) != 2 {
		t.Fatalf("listed %d domains, want 2", len(all))
	}
	if got, err := catalog.Describe("rotating-machinery"); err != nil || got != pumps {
		t.Fatalf("describe = %v, %v", got, err)
	}
	if _, err := catalog.Describe("no-such-domain"); err == nil || !strings.Contains(err.Error(), "no-such-domain") {
		t.Fatalf("unknown domain: %v", err)
	}
	report := catalog.Coverage()
	if len(report.ByAxis) == 0 {
		t.Fatalf("coverage has no axes: %+v", report)
	}
	for axis, values := range report.ByAxis {
		for value, count := range values {
			if count < 0 || count > 2 {
				t.Fatalf("axis %s value %s carried by %d of 2 domains", axis, value, count)
			}
		}
	}
}

func TestParseCompilesAShippedDocumentToTheSameDigestAsLoad(t *testing.T) {
	t.Parallel()
	loaded := shipped(t, "cold-chain-transit")
	parsed, err := domain.Parse(loaded.Raw, "in-memory")
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Digest != loaded.Digest || !parsed.HasChannel(loaded.ChannelNames()[0]) {
		t.Fatalf("digest %s vs %s", parsed.Digest, loaded.Digest)
	}
}

func TestParseNamesTheSourceOfAnInvalidDocument(t *testing.T) {
	t.Parallel()
	if _, err := domain.Parse([]byte(`{"id":`), "broken.json"); err == nil || !strings.Contains(err.Error(), "broken.json") {
		t.Fatalf("err = %v", err)
	}
}
