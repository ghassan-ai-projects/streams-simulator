package domain

import (
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestTheAuditControlAndChannelsBelongToTheKindOfTheFault(t *testing.T) {
	t.Parallel()
	spec, err := domain.Load(testsupport.TwoKindDomain())
	if err != nil {
		t.Fatal(err)
	}
	p := NewPanel(spec, 1, 60*1e9)
	if p.controlFault("pump_wear") != "pump_nuisance" || p.controlFault("tank_leak") != "tank_nuisance" || p.controlFault("none") != "" {
		t.Fatal("the control fault is not the negative class of the fault's own kind")
	}
	if got := p.kindChannels("pump-1", "pump_wear"); !slices.Equal(got, []string{"pump.flow"}) {
		t.Fatalf("pump channels = %v", got)
	}
	if got := p.kindChannels("tank-1", "none"); len(got) != 2 {
		t.Fatalf("an unknown fault falls back to every channel, got %v", got)
	}
}
