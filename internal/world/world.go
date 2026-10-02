// Package world is the seeded discrete-event world core. It holds hidden
// state, dynamics, faults and effectors, and produces the native
// sim-event-v0.1 stream. It is strictly single-goroutine; concurrency lives
// only in the sinks, behind a bounded channel with ordered writes.
//
// A run is a pure function of (sim_version, domain_digest, adapter_digest,
// seed, command_log, sink). The world honors its half: nothing here reads a
// wall clock, iterates a map to produce output, or draws from a shared
// generator.
package world

import (
	"container/heap"
	"fmt"
	"strconv"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
)

// Options tune world construction. Everything here is part of the
// determinism tuple when it changes output.
type Options struct {
	// Noiseless disables channel observation noise: used by the observability
	// solver, which diffs a faulted world against a clean one.
	Noiseless bool
	// EmitDisabled suppresses emission: used by the solver when only hidden
	// state is needed.
	EmitDisabled bool
	// InitialEntities overrides the spec's entity count with explicit ids.
	InitialEntities []string
	// ForceEffectorOK makes every effector invocation apply its effect with
	// mode ok, bypassing the declared failure-mode distribution: used by the
	// solver's scenario setup, which must apply deterministically.
	ForceEffectorOK bool
	// ForceFailureMode forces every effector invocation into the given
	// failure mode ("" = declared distribution): used by tests that must
	// exercise a specific mode deterministically.
	ForceFailureMode string
}

// World is one seeded simulated world.
type World struct {
	ID               string
	Spec             *domain.Compiled
	Seed             uint64
	StartNS          int64
	ClockNS          int64
	Noiseless        bool
	EmitDisabled     bool
	forceEffectorOK  bool
	forceFailureMode string

	queue    priorityQueue
	tiebreak uint64
	seq      int64
	// lastObservedNS serializes native delivery timestamps in emission order.
	// The link-delay model supplies a timestamp candidate; this watermark adds
	// the deterministic total-order tiebreak required by consumers.
	lastObservedNS    int64
	hasObservedTimeNS bool

	subs        map[string]*randutil.SplitMix64
	entities    map[string]*Entity
	entityOrder []string
	nextIndex   int

	kicks  map[driverKey][]*kick
	shadow map[driverKey][]*kick

	faultsByState map[string][]*activeFault
	faultsByID    map[string]*activeFault
	faultOrder    []string

	effectorCalls  []EffectorCall
	idempotent     map[string]*idempotentResult
	interlockActed map[string]bool
	argSchemas     map[string]*jsonschema.Schema

	emitter func(model.SimEvent)

	// deterministic entity-id source for autonomous churn
	churnSeed uint64

	// AdvanceCounters (per advance call, reset each call)
	emittedThisAdvance int
	effectsAppliedThis int
}

// Entity is one simulated producer.
type Entity struct {
	ID        string
	Type      string
	BornNS    int64
	RetiredNS int64
	States    map[string]*stateValue
	channels  map[string]*channelRunState
	alive     bool
}

// channelRunState is the per-channel scheduling state of one entity.
type channelRunState struct {
	availDown   bool
	availUntil  int64
	availInit   bool
	lastSent    float64
	hasSent     bool
	lastTrigger float64
	walk        float64
	walkLastNS  int64
}

// EffectorCall is the director-side record of one effector invocation: the
// authority when scoring actions, not the consumer's claim.
type EffectorCall struct {
	CommandID        string         `json:"command_id"`
	Effector         string         `json:"effector"`
	EntityID         string         `json:"entity_id"`
	WorldID          string         `json:"world_id"`
	Args             map[string]any `json:"args"`
	AtNS             int64          `json:"at_ns"`
	Mode             string         `json:"mode"`
	Accepted         bool           `json:"accepted"`
	InterlockRefused bool           `json:"interlock_refused,omitempty"`
	Reason           string         `json:"reason,omitempty"`
	AckLatencyMS     float64        `json:"ack_latency_ms"`
	EffectApplied    bool           `json:"effect_applied"`
	ResultDigest     string         `json:"result_digest,omitempty"`
}

type idempotentResult struct {
	call      EffectorCall // value copy of the original invocation
	expiresNS int64
}

// New creates a world. The world id is derived deterministically from the
// seed and domain, so replay reconstructs the same id.
func New(spec *domain.Compiled, seed uint64, id string, startNS int64, opts Options) (*World, error) {
	if id == "" {
		id = "w-" + strconv.FormatUint(randutil.Fnv1a64(spec.Spec.ID+":"+strconv.FormatUint(seed, 10))%0xffffff, 36)
	}
	w := &World{
		ID:               id,
		Spec:             spec,
		Seed:             seed,
		StartNS:          startNS,
		ClockNS:          startNS,
		Noiseless:        opts.Noiseless,
		EmitDisabled:     opts.EmitDisabled,
		forceEffectorOK:  opts.ForceEffectorOK,
		forceFailureMode: opts.ForceFailureMode,
		subs:             map[string]*randutil.SplitMix64{},
		entities:         map[string]*Entity{},
		kicks:            map[driverKey][]*kick{},
		shadow:           map[driverKey][]*kick{},
		faultsByState:    map[string][]*activeFault{},
		faultsByID:       map[string]*activeFault{},
		idempotent:       map[string]*idempotentResult{},
		interlockActed:   map[string]bool{},
		argSchemas:       map[string]*jsonschema.Schema{},
		churnSeed:        seed,
	}
	heap.Init(&w.queue)

	ids := opts.InitialEntities
	if len(ids) == 0 {
		n := spec.Spec.Entities.Count.Default
		if n < 1 {
			n = 1
		}
		for i := 1; i <= n; i++ {
			ids = append(ids, renderID(spec.Spec.Entities.IDTemplate, i, nil))
		}
	}
	for _, id := range ids {
		if err := w.addEntity(id, startNS, nil); err != nil {
			return nil, fmt.Errorf("streamsim: %w", err)
		}
	}
	// Start autonomous IDs after the generated initial set. Births still
	// probe for a free ID because callers may supply custom initial IDs.
	w.nextIndex = len(ids)
	if ch := spec.Spec.Entities.Churn; ch != nil && ch.BirthsPerHour > 0 {
		w.scheduleChurn(startNS)
	}
	return w, nil
}
