package mcp

import (
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestAuditScenarioReportsTheTrivialBaselineVerdict(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	start := model.DefaultStartTimeNS + 4*3600*1e9
	out, err := d.AuditScenario("aquaculture-pond", "site-a/pond-1", "aerator_failure", start+2*3600*1e9, start, 6*3600*1e9)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"trivial", "scores", "best"} {
		if _, ok := out[key]; !ok {
			t.Fatalf("result lacks %q: %v", key, out)
		}
	}
}

func TestAuditScenarioNamesAnUnknownDomainAndAnUnknownFault(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	if _, err := d.AuditScenario("no-such-domain", "e", "f", 0, 0, 3600*1e9); err == nil || !strings.Contains(err.Error(), "no-such-domain") {
		t.Fatalf("unknown domain: %v", err)
	}
	_, err := d.AuditScenario("aquaculture-pond", "site-a/pond-1", "no_such_fault", 0, 0, 3600*1e9)
	if err == nil || !strings.Contains(err.Error(), "no_such_fault") {
		t.Fatalf("unknown fault: %v", err)
	}
}
