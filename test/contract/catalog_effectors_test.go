package contract

import (
	"fmt"
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

const (
	catalogPreludeNS = 60e9
	catalogTestNS    = 2 * 3600e9
	catalogWindowNS  = 3 * 3600e9
	catalogSampleNS  = 120e9
)

func checkCatalogEffectors(t *testing.T, p catalogPlant) {
	for _, effector := range p.kind().Effectors {
		if len(effector.Effect.StateDeltas) == 0 {
			t.Errorf("%s names no state in its effect", effector.Name)
		} else if !effectorMoves(t, p, effector) {
			t.Errorf("%s moves none of %v, whatever the arguments or the state it starts from", effector.Name, effectStates(effector))
		}
	}
}

func effectStates(effector model.Effector) []string {
	var states []string
	for _, delta := range effector.Effect.StateDeltas {
		states = append(states, delta.State)
	}
	return states
}

func effectorMoves(t *testing.T, p catalogPlant, effector model.Effector) bool {
	t.Helper()
	preludes := []*model.Effector{nil}
	for i := range p.kind().Effectors {
		if other := &p.kind().Effectors[i]; other.Name != effector.Name {
			preludes = append(preludes, other)
		}
	}
	for _, prelude := range preludes {
		for attempt := 0; attempt < catalogArgAttempts; attempt++ {
			if moved, err := tryEffector(p, effector, prelude, attempt); err != nil {
				t.Fatal(err)
			} else if moved {
				return true
			}
		}
	}
	return false
}

func tryEffector(p catalogPlant, effector model.Effector, prelude *model.Effector, attempt int) (bool, error) {
	with, err := newCatalogWorld(p)
	if err != nil {
		return false, err
	}
	without, err := newCatalogWorld(p)
	if err != nil {
		return false, err
	}
	entity, start := p.entities[0], model.DefaultStartTimeNS
	if prelude != nil {
		for _, w := range []*world.World{with, without} {
			args := catalogArgs(prelude.ArgsSchema, entity, 0)
			_, _ = w.InvokeEffector(prelude.Name, entity, "prelude", args, start+catalogPreludeNS)
		}
	}
	_, _ = with.InvokeEffector(effector.Name, entity, "test", catalogArgs(effector.ArgsSchema, entity, attempt), start+catalogTestNS)
	return statesDiverge(with, without, entity, effectStates(effector), start)
}

func newCatalogWorld(p catalogPlant) (*world.World, error) {
	w, err := world.New(p.spec, catalogSeed, "catalog", model.DefaultStartTimeNS,
		world.Options{Noiseless: true, EmitDisabled: true, ForceEffectorOK: true})
	if err != nil {
		return nil, fmt.Errorf("new world: %w", err)
	}
	return w, nil
}

func statesDiverge(with, without *world.World, entity string, states []string, start int64) (bool, error) {
	for at := start + catalogTestNS; at <= start+catalogTestNS+catalogWindowNS; at += catalogSampleNS {
		for _, w := range []*world.World{with, without} {
			if _, _, err := w.Advance(at); err != nil {
				return false, fmt.Errorf("advance: %w", err)
			}
		}
		for _, state := range states {
			a, b := with.StateValue(entity, state, at), without.StateValue(entity, state, at)
			if math.Abs(a-b) > 1e-9+1e-6*math.Abs(b) {
				return true, nil
			}
		}
	}
	return false, nil
}
