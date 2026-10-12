package domain

import (
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type protectionState struct {
	holding   bool
	heldSince int64
	latched   bool
	seen      bool
	lastSeen  float64
	until     int64
}

func (w *World) conditionHolds(entityID string, c *model.Condition, atNS int64) bool {
	v := w.stateAt(entityID, c.State, atNS)
	switch c.Operator {
	case "lt":
		return v < c.Threshold
	case "lte":
		return v <= c.Threshold
	case "gt":
		return v > c.Threshold
	case "gte":
		return v >= c.Threshold
	}
	return false
}

func (w *World) stepTrip(ent *Entity, dyn *model.Dynamics, s *stateValue, at int64) {
	f1, p := dyn.F1, s.protectionOf()
	if f1.ResetOn != "" && p.changed(w.stateAt(ent.ID, f1.ResetOn, at)) {
		p.latched = false
	}
	p.track(w.conditionHolds(ent.ID, f1.When, at), at)
	tripped := p.latched || (p.holding && at-p.heldSince >= int64(f1.ForS*secondsPerNS))
	p.latched = f1.Latch && tripped
	s.x = w.initial(ent, dyn.Target)
	if tripped {
		s.x = f1.TrippedValue
	}
}

func (w *World) stepAntiRecycle(ent *Entity, dyn *model.Dynamics, s *stateValue, at, dt int64) {
	f1, p := dyn.F1, s.protectionOf()
	previous, seen := p.lastSeen, p.seen
	value := w.stateAt(ent.ID, f1.Watch, at)
	p.lastSeen, p.seen = value, true
	if seen && edgeOf(f1.Edge, previous, value) {
		p.until = at + int64(f1.HoldS*secondsPerNS)
	}
	s.x = float64(max(0, p.until-(at+dt))) / secondsPerNS
}

func edgeOf(edge string, previous, value float64) bool {
	if edge == model.EdgeFall {
		return value < previous
	}
	return value > previous
}

func (s *stateValue) protectionOf() *protectionState {
	if s.protection == nil {
		s.protection = &protectionState{}
	}
	return s.protection
}

func (p *protectionState) track(holds bool, at int64) {
	if holds && !p.holding {
		p.heldSince = at
	}
	p.holding = holds
}

func (p *protectionState) changed(value float64) bool {
	changed := p.seen && value != p.lastSeen
	p.lastSeen, p.seen = value, true
	return changed
}

const ReasonPreconditionRefused = "precondition_refused"

func (w *World) refusePrecondition(eff *model.Effector, inv invocation) *InvokeResult {
	for i := range eff.Requires {
		if !w.conditionHolds(inv.entityID, &eff.Requires[i], inv.atNS) {
			w.recordCall(inv, callOutcome{mode: ModeReject, reason: ReasonPreconditionRefused})
			return &InvokeResult{Simulated: true, WorldID: w.ID, CommandID: inv.commandID, Mode: ModeReject, Reason: "effector_refused"}
		}
	}
	return nil
}

func (w *World) refuseInvocation(eff *model.Effector, inv invocation) (*InvokeResult, error) {
	if eff.Interlock != nil || len(eff.Requires) > 0 {
		defer w.lookAhead(inv.entityID, inv.atNS)()
	}
	if w.refuseInterlock(eff, inv) {
		return nil, ErrInterlockRefused
	}
	return w.refusePrecondition(eff, inv), nil
}

func (w *World) ViolatedBoundaries(entityID string, t int64) []string {
	ent := w.entities[entityID]
	if ent == nil {
		return nil
	}
	defer w.lookAhead(entityID, t)()
	var violated []string
	for _, boundary := range ent.kind.Boundaries {
		if dyn := ent.kind.DynamicsOf(boundary.State); w.stateAt(entityID, boundary.State, t) == dyn.F1.TrippedValue {
			violated = append(violated, boundary.Name)
		}
	}
	return violated
}

func (w *World) lookAhead(entityID string, atNS int64) func() {
	ent := w.entities[entityID]
	if ent == nil || atNS <= w.clockNS {
		return func() {}
	}
	committed := ent.states
	ent.states = make(map[string]*stateValue, len(committed))
	// determinism-safe: copies a map; no order reaches output.
	for name, value := range committed {
		ent.states[name] = value.withHistory()
	}
	return func() { ent.states = committed }
}

func (s *stateValue) withHistory() *stateValue {
	c := s.clone()
	c.history = slices.Clone(s.history)
	return c
}
