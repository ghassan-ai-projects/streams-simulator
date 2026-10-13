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
	"fmt"
	"math"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
)

const MaxStartNS = math.MaxInt64 / 2

func admitWorld(startNS int64, opts Options) error {
	if startNS < 0 || startNS > MaxStartNS {
		return fmt.Errorf("world: start %d is outside [0, %d]", startNS, int64(MaxStartNS))
	}
	if opts.ForceFailureMode != "" && !KnownFailureMode(opts.ForceFailureMode) {
		return fmt.Errorf("world: unknown failure mode %q", opts.ForceFailureMode)
	}
	return nil
}

type Options struct {
	Noiseless bool

	EmitDisabled bool

	InitialEntities []string

	ForceEffectorOK bool

	ForceFailureMode FailureMode
}

type World struct {
	ID               string
	Spec             *domain.Compiled
	seed             uint64
	StartNS          int64
	clockNS          int64
	noiseless        bool
	emitDisabled     bool
	forceEffectorOK  bool
	forceFailureMode FailureMode

	queue    priorityQueue
	tiebreak uint64
	seq      int64

	lastObservedNS    int64
	hasObservedTimeNS bool

	subs        map[string]*randutil.SplitMix64
	streamKey   string
	entities    map[string]*Entity
	entityOrder []string
	nextIndex   []int

	kicks  map[driverKey][]*kick
	leases map[driverKey][]*kick
	shadow map[driverKey][]*kick

	faultsByState   map[string][]*activeFault
	faultsByChannel map[driverKey][]*activeFault
	faultsByID      map[string]*activeFault
	faultOrder      []string

	effectorCalls []EffectorCall
	idempotent    map[string]*idempotentResult
	argSchemas    map[*model.Effector]*jsonschema.Schema

	emitter  func(model.SimEvent)
	coupling coupling

	emittedThisAdvance int
	effectsAppliedThis int
}

type Entity struct {
	ID        string
	Type      string
	BornNS    int64
	kind      *model.Kind
	set       int
	states    map[string]*stateValue
	channels  map[string]*channelRunState
	alive     bool
	retiredNS int64
	readersAt int64
}

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

type EffectorCall struct {
	CommandID        string         `json:"command_id"`
	Effector         string         `json:"effector"`
	EntityID         string         `json:"entity_id"`
	WorldID          string         `json:"world_id"`
	Args             map[string]any `json:"args"`
	AtNS             int64          `json:"at_ns"`
	Mode             FailureMode    `json:"mode"`
	Accepted         bool           `json:"accepted"`
	InterlockRefused bool           `json:"interlock_refused,omitempty"`
	Reason           string         `json:"reason,omitempty"`
	AckLatencyMS     float64        `json:"ack_latency_ms"`
	EffectApplied    bool           `json:"effect_applied"`
	ResultDigest     string         `json:"result_digest,omitempty"`
}

type idempotentResult struct {
	request   string
	result    InvokeResult
	expiresNS int64
}

func New(spec *domain.Compiled, seed uint64, id string, startNS int64, opts Options) (*World, error) {
	if err := admitWorld(startNS, opts); err != nil {
		return nil, err
	}
	id, key := worldIdentity(spec, seed, id)
	w := newWorldState(spec, seed, id, key, startNS, opts)
	w.initializeRuntime(seed)
	heap.Init(&w.queue)
	ids := initialEntityIDs(spec, opts.InitialEntities)
	if err := w.populateInitialEntities(ids, startNS); err != nil {
		return nil, err
	}
	w.scheduleInitialChurn(startNS)
	return w, nil
}
