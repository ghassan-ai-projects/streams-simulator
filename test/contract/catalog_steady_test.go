package contract

import (
	"math"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

const catalogSteadyNS = 600e9

// checkCatalogSteadyStart proves a nominal day starts at rest: no state moves by more than 3 % in its first ten minutes.
func checkCatalogSteadyStart(t *testing.T, p catalogPlant) {
	w, err := world.New(p.spec, catalogSeed, "steady", model.DefaultStartTimeNS, world.Options{Noiseless: true, EmitDisabled: true})
	if err != nil {
		t.Fatal(err)
	}
	start := model.DefaultStartTimeNS
	entity := p.entities[0]
	before := map[string]float64{}
	for _, state := range p.kind().State {
		before[state.Name] = w.StateValue(entity, state.Name, start)
	}
	if _, _, err := w.Advance(start + catalogSteadyNS); err != nil {
		t.Fatal(err)
	}
	for _, state := range p.kind().State {
		if strings.HasSuffix(state.Name, "__bias") {
			continue
		}
		a, b := before[state.Name], w.StateValue(entity, state.Name, start+catalogSteadyNS)
		if math.Abs(b-a) > 0.03*math.Max(math.Abs(a), math.Abs(b))+0.05 {
			t.Errorf("state %s moves from %.4g to %.4g in its first ten minutes: its initial value is not the nominal steady state", state.Name, a, b)
		}
	}
}
