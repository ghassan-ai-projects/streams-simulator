// Package domain holds the seeded discrete-event world core: hidden state,
// dynamics, faults and effectors, and the production of the native
// sim-event-v0.1 stream. It is strictly single-goroutine and performs no I/O.
//
// A run is a pure function of (sim_version, domain_digest, adapter_digest,
// seed, command_log, sink). The world honors its half: nothing here reads a
// wall clock, iterates a map to produce output, or draws from a shared
// generator.
package domain

import (
	"container/heap"

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
	seed             uint64
	StartNS          int64
	clockNS          int64
	noiseless        bool
	emitDisabled     bool
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
	ID       string
	Type     string
	BornNS   int64
	states   map[string]*stateValue
	channels map[string]*channelRunState
	alive    bool
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
	request   string       // effector, entity and arguments the command_id first carried
	result    InvokeResult // value copy of the answer it first received
	expiresNS int64
}

// New creates a world. The world id is derived deterministically from the
// seed and domain, so replay reconstructs the same id.
func New(spec *domain.Compiled, seed uint64, id string, startNS int64, opts Options) (*World, error) {
	id = worldIdentity(spec, seed, id)
	w := newWorldState(spec, seed, id, startNS, opts)
	w.initializeRuntime(seed)
	heap.Init(&w.queue)
	ids := initialEntityIDs(spec, opts.InitialEntities)
	if err := w.populateInitialEntities(ids, startNS); err != nil {
		return nil, err
	}
	w.scheduleInitialChurn(startNS)
	return w, nil
}
