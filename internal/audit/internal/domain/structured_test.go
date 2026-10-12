package domain

import (
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestTheAuditChannelsComeFromTheInjectedEntitysEffectiveKind(t *testing.T) {
	t.Parallel()
	spec, err := domain.Load(testsupport.StructuredDomain())
	if err != nil {
		t.Fatal(err)
	}
	p := NewPanel(spec, 1, 60*1e9)
	if got := p.kindChannels("tank-1", "tank_leak"); !slices.Equal(got, []string{"tank.level"}) {
		t.Fatalf("plain tank channels = %v, want the kind's only", got)
	}
	if got := p.kindChannels("T-main", "tank_overheat"); !slices.Equal(got, []string{"tank.level", "tank.temp"}) {
		t.Fatalf("named tank channels = %v, want the kind's plus the added channel", got)
	}
}
