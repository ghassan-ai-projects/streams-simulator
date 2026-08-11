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
	"math"
	"strconv"
	"strings"

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

// RenderID renders one entity id from the domain's id template. Exported
// for hosts that must enumerate the domain's entity ids without building a
// world (the audit path).
func RenderID(tmpl string, n int) string {
	return renderID(tmpl, n, nil)
}

// renderID expands the id template.
func renderID(tmpl string, n int, params map[string]any) string {
	out := tmpl
	out = strings.ReplaceAll(out, "{n}", strconv.Itoa(n))
	for {
		start := strings.Index(out, "{")
		end := strings.Index(out, "}")
		if start < 0 || end < 0 || end < start {
			break
		}
		name := out[start+1 : end]
		val := "a"
		if params != nil {
			if v, ok := params[name]; ok {
				switch x := v.(type) {
				case string:
					val = x
				default:
					val = fmt.Sprint(x)
				}
			}
		}
		out = out[:start] + val + out[end+1:]
	}
	return out
}

// addEntity creates an entity with initial hidden state and schedules its
// first emissions.
func (w *World) addEntity(id string, atNS int64, params map[string]any) error {
	if _, exists := w.entities[id]; exists {
		return fmt.Errorf("world: entity %q already exists", id)
	}
	et := w.Spec.Spec.Entities.EntityType
	if et == "" {
		et = strings.Split(w.Spec.Spec.ID, "-")[0]
	}
	ent := &Entity{
		ID:       id,
		Type:     et,
		BornNS:   atNS,
		States:   map[string]*stateValue{},
		channels: map[string]*channelRunState{},
		alive:    true,
	}
	for i := range w.Spec.Spec.State {
		st := &w.Spec.Spec.State[i]
		ent.States[st.Name] = &stateValue{
			x:        st.Initial,
			lastStep: atNS,
		}
	}
	w.entities[id] = ent
	w.entityOrder = append(w.entityOrder, id)

	// Generate any template ids from params for this entity.
	if tmpl := w.Spec.Spec.Entities.IDTemplate; tmpl != "" {
		// re-render a fresh n if this is an autonomous birth
		_ = tmpl
	}
	_ = params

	for i := range w.Spec.Spec.Channels {
		ch := &w.Spec.Spec.Channels[i]
		cs := &channelRunState{lastTrigger: w.stateAt(id, ch.Cadence.TriggerState, atNS)}
		ent.channels[ch.Name] = cs
		next := w.nextEmission(id, ch, atNS)
		if next > 0 {
			w.schedule(kindEmission, id, ch.Name, next, nil)
		}
	}
	return nil
}

// scheduleChurn seeds autonomous entity births.
func (w *World) scheduleChurn(startNS int64) {
	ch := w.Spec.Spec.Entities.Churn
	if ch == nil || ch.BirthsPerHour <= 0 {
		return
	}
	rng := w.substream("churn/births")
	next := startNS + int64(rng.Exp(3600*secondsPerNS/ch.BirthsPerHour))
	w.schedule(kindBirth, "", "", next, nil)
}

// Schedule pushes an event onto the DES queue with a monotonic tiebreak.
func (w *World) schedule(kind eventKind, entity, channel string, atNS int64, payload any) {
	if atNS < 0 {
		atNS = 0
	}
	w.tiebreak++
	heap.Push(&w.queue, &item{
		timeNS:   atNS,
		tiebreak: w.tiebreak,
		kind:     kind,
		entity:   entity,
		channel:  channel,
		payload:  payload,
	})
}

// NextEventNS is the time of the next scheduled event.
func (w *World) NextEventNS() int64 {
	if w.queue.Len() == 0 {
		return 0
	}
	return w.queue[0].timeNS
}

// PendingEvents is the number of scheduled events.
func (w *World) PendingEvents() int { return w.queue.Len() }

// Advance processes every event scheduled at or before to, then sets the
// clock to to. Moving the clock backwards is refused.
func (w *World) Advance(to int64) (int, int, error) {
	if to < w.ClockNS {
		return 0, 0, fmt.Errorf("world: clock would move backwards (%d -> %d)", w.ClockNS, to)
	}
	w.emittedThisAdvance = 0
	w.effectsAppliedThis = 0
	for w.queue.Len() > 0 {
		it := w.queue[0]
		if it.timeNS > to {
			break
		}
		heap.Pop(&w.queue)
		w.ClockNS = it.timeNS
		switch it.kind {
		case kindEmission:
			w.processEmission(it.entity, it.channel, it.timeNS)
		case kindEffectStart:
			if k, ok := it.payload.(*kick); ok && !k.applied {
				k.applied = true
				w.effectsAppliedThis++
			}
		case kindBirth:
			w.birthAutonomous(it.timeNS)
		case kindDeath:
			w.retire(it.entity, "lifetime", it.timeNS)
		}
	}
	w.ClockNS = to
	return w.emittedThisAdvance, w.effectsAppliedThis, nil
}

// Clock is the current world time.
func (w *World) Clock() int64 { return w.ClockNS }

// SetEmitter installs the event sink. Events flow world -> emitter.
func (w *World) SetEmitter(emitter func(model.SimEvent)) {
	w.emitter = emitter
}

// EmittedCount is the total number of native events produced.
func (w *World) EmittedCount() int64 { return w.seq }

// substream returns the cached PRNG for a named substream under this world's
// seed. Names are "<world_id>/<entity>/<channel>/<purpose>"; the world id
// prefix keeps distinct worlds independent even under the same seed.
func (w *World) substream(name string) *randutil.SplitMix64 {
	full := w.ID + "/" + name
	if s, ok := w.subs[full]; ok {
		return s
	}
	s := randutil.Substream(w.Seed, full)
	w.subs[full] = s
	return s
}

// EntityIDs returns the live entity ids in creation order.
func (w *World) EntityIDs() []string {
	var out []string
	for _, id := range w.entityOrder {
		if w.entities[id] != nil && w.entities[id].alive {
			out = append(out, id)
		}
	}
	return out
}

// InitialEntityIDs returns the entities created at world construction, in
// order (before any churn or entity.add).
func (w *World) InitialEntityIDs() []string {
	var out []string
	for _, id := range w.entityOrder {
		if w.entities[id] != nil && w.entities[id].BornNS == w.StartNS {
			out = append(out, id)
		}
	}
	return out
}

// AddEntity creates an entity at the given time (entity.add; churn births
// go through birthAutonomous, which schedules the death).
func (w *World) AddEntity(id string, atNS int64, params map[string]any) error {
	return w.addEntity(id, atNS, params)
}

// Entity returns the entity by id, or nil.
func (w *World) Entity(id string) *Entity { return w.entities[id] }

// DynamicsFor returns the dynamics declaration for a state, or nil.
func (w *World) DynamicsFor(state string) *model.Dynamics {
	for i := range w.Spec.Spec.Dynamics {
		if w.Spec.Spec.Dynamics[i].Target == state {
			return &w.Spec.Spec.Dynamics[i]
		}
	}
	return nil
}

// StateValue exposes a hidden state to the director only (solver and truth).
func (w *World) StateValue(entity, state string, t int64) float64 {
	return w.stateAt(entity, state, t)
}

// quantize rounds v to the channel's declared resolution.
func quantize(v, resolution float64) float64 {
	if resolution <= 0 {
		return v
	}
	return math.Round(v/resolution) * resolution
}
