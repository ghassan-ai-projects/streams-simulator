package domain

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
)

func worldIdentity(spec *domain.Compiled, seed uint64, id string) (string, string) {
	if id != "" {
		return id, id
	}
	key := "w-" + strconv.FormatUint(randutil.Fnv1a64(spec.Spec.ID+":"+strconv.FormatUint(seed, 10))%0xffffff, 36)
	return "w-" + strconv.FormatUint(randutil.Fnv1a64(spec.Spec.ID)%0xffffff, 36), key
}

func newWorldState(spec *domain.Compiled, seed uint64, id, key string, startNS int64, opts Options) *World {
	return &World{
		ID: id, streamKey: key, Spec: spec, seed: seed,
		StartNS:          startNS,
		clockNS:          startNS,
		noiseless:        opts.Noiseless,
		emitDisabled:     opts.EmitDisabled || opts.Unobserved,
		unobserved:       opts.Unobserved,
		forceEffectorOK:  opts.ForceEffectorOK,
		forceFailureMode: opts.ForceFailureMode,
		kpis:             newKPITracker(opts.TrackKPIs),
	}
}

func (w *World) initializeRuntime(seed uint64) {
	w.subs = map[string]*randutil.SplitMix64{}
	w.entities = map[string]*Entity{}
	w.kicks = map[driverKey][]*kick{}
	w.leases = map[driverKey][]*kick{}
	w.shadow = map[driverKey][]*kick{}
	w.faultsByState = map[string][]*activeFault{}
	w.faultsByChannel = map[driverKey][]*activeFault{}
	w.faultsByID = map[string]*activeFault{}
	w.idempotent = map[string]*idempotentResult{}
	w.argSchemas = map[*model.Effector]*jsonschema.Schema{}
	w.coupling = newCoupling(w.Spec)
}

func initialEntityIDs(spec *domain.Compiled, ids []string) []string {
	if len(ids) > 0 {
		return ids
	}
	return PopulationIDs(spec, 0)
}

func PopulationIDs(spec *domain.Compiled, firstSetCount int) []string {
	var ids []string
	counted := firstTemplatedSet(spec)
	for i := range spec.Spec.Instances {
		set := &spec.Spec.Instances[i]
		count := set.Count.Default
		if i == counted && firstSetCount > 0 {
			count = firstSetCount
		}
		ids = append(ids, set.Names...)
		ids = append(ids, templatedIDs(set, count)...)
	}
	return ids
}

func firstTemplatedSet(spec *domain.Compiled) int {
	return slices.IndexFunc(spec.Spec.Instances, func(set model.InstanceSet) bool { return set.IDTemplate != "" })
}

func templatedIDs(set *model.InstanceSet, count int) []string {
	var ids []string
	for n := 1; set.IDTemplate != "" && n <= max(count, 1); n++ {
		ids = append(ids, renderID(set.IDTemplate, n, nil))
	}
	return ids
}

func (w *World) populateInitialEntities(ids []string, startNS int64) error {
	w.nextIndex = make([]int, len(w.Spec.Spec.Instances))
	for _, id := range ids {
		set := max(w.instanceSetOf(id), 0)
		if err := w.addEntity(id, set, startNS); err != nil {
			return fmt.Errorf("streamsim: %w", err)
		}
		if !w.Spec.IsInstanceName(id) {
			w.nextIndex[set]++
		}
	}
	return nil
}

func (w *World) instanceSetOf(id string) int {
	set := w.Spec.InstanceSetFor(id)
	if set < 0 && len(w.Spec.Spec.Instances) == 1 {
		return 0
	}
	return set
}

func (w *World) scheduleInitialChurn(startNS int64) {
	for set := range w.Spec.Spec.Instances {
		w.scheduleChurn(set, startNS)
	}
}
