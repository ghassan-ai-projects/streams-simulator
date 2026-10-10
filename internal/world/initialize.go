package world

import (
	"fmt"
	"strconv"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
)

func worldIdentity(spec *domain.Compiled, seed uint64, id string) string {
	if id == "" {
		id = "w-" + strconv.FormatUint(randutil.Fnv1a64(spec.Spec.ID+":"+strconv.FormatUint(seed, 10))%0xffffff, 36)
	}
	return id
}

func newWorldState(spec *domain.Compiled, seed uint64, id string, startNS int64, opts Options) *World {
	return &World{
		ID:               id,
		Spec:             spec,
		seed:             seed,
		StartNS:          startNS,
		clockNS:          startNS,
		noiseless:        opts.Noiseless,
		emitDisabled:     opts.EmitDisabled,
		forceEffectorOK:  opts.ForceEffectorOK,
		forceFailureMode: opts.ForceFailureMode,
	}
}

func (w *World) initializeRuntime(seed uint64) {
	w.subs = map[string]*randutil.SplitMix64{}
	w.entities = map[string]*Entity{}
	w.kicks = map[driverKey][]*kick{}
	w.shadow = map[driverKey][]*kick{}
	w.faultsByState = map[string][]*activeFault{}
	w.faultsByID = map[string]*activeFault{}
	w.idempotent = map[string]*idempotentResult{}
	w.interlockActed = map[string]bool{}
	w.argSchemas = map[string]*jsonschema.Schema{}
	w.churnSeed = seed
}

func initialEntityIDs(spec *domain.Compiled, ids []string) []string {
	if len(ids) > 0 {
		return ids
	}
	count := spec.Spec.Entities.Count.Default
	if count < 1 {
		count = 1
	}
	for i := 1; i <= count; i++ {
		ids = append(ids, renderID(spec.Spec.Entities.IDTemplate, i, nil))
	}
	return ids
}

func (w *World) populateInitialEntities(ids []string, startNS int64) error {
	for _, id := range ids {
		if err := w.addEntity(id, startNS); err != nil {
			return fmt.Errorf("streamsim: %w", err)
		}
	}
	// Births also probe for free IDs when initial IDs were supplied by callers.
	w.nextIndex = len(ids)
	return nil
}

func (w *World) scheduleInitialChurn(startNS int64) {
	if churn := w.Spec.Spec.Entities.Churn; churn != nil && churn.BirthsPerHour > 0 {
		w.scheduleChurn(startNS)
	}
}
