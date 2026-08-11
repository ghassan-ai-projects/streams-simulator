package world

// World faults: applied inside the world core, they change hidden state and
// are what a consumer is supposed to diagnose. They are distinct from
// delivery perturbations (the observer is unreliable) and environment faults
// (the consumer is broken) — three independently seeded injection surfaces,
// never conflated.

import (
	"fmt"
	"math"
	"strconv"
)

// FaultInfo describes one active fault (director role only).
type FaultInfo struct {
	FaultID  string  `json:"fault_id"`
	EntityID string  `json:"entity_id"`
	Fault    string  `json:"fault"`
	OnsetNS  int64   `json:"onset_ns"`
	Severity float64 `json:"severity,omitempty"`
}

// InjectFault applies a declared fault to an entity. onsetNS defaults to the
// current clock. params.severity scales the fault's deltas.
func (w *World) InjectFault(entityID, faultID string, onsetNS int64, params map[string]any) (string, error) {
	fault := w.Spec.Fault(faultID)
	if fault == nil {
		return "", fmt.Errorf("world: unknown fault %q", faultID)
	}
	if _, ok := w.entities[entityID]; !ok {
		return "", fmt.Errorf("world: unknown entity %q", entityID)
	}
	if onsetNS <= 0 {
		onsetNS = w.ClockNS
	}
	severity := 1.0
	if fault.Onset.Magnitude != 0 {
		severity = math.Abs(fault.Onset.Magnitude)
	}
	if params != nil {
		if s, ok := params["severity"]; ok {
			switch x := s.(type) {
			case float64:
				severity = x
			case int64:
				severity = float64(x)
			default:
				return "", fmt.Errorf("world: severity must be numeric")
			}
		}
	}
	if math.IsNaN(severity) || math.IsInf(severity, 0) || severity < 0 {
		return "", fmt.Errorf("world: severity must be finite and non-negative")
	}
	af := &activeFault{
		fault:    fault,
		entity:   entityID,
		onsetNS:  onsetNS,
		severity: severity,
	}
	if fault.Onset.Shape == "stochastic" {
		// The envelope is defined relative to fault onset. Starting at epoch
		// zero would make a first read at a modern Unix timestamp replay
		// millions of random-walk steps.
		af.walkStep = onsetNS
	}
	// A fault already active on this entity is re-injected as a fresh
	// instance (clear first if the caller wants a single instance).
	fid := "f-" + strconv.FormatInt(int64(len(w.faultsByID)), 10)
	w.faultsByID[fid] = af
	w.faultOrder = append(w.faultOrder, fid)
	for _, a := range fault.Affects {
		w.faultsByState[a.State] = append(w.faultsByState[a.State], af)
	}
	return fid, nil
}

// ClearFault removes a fault's contribution from the given time onward.
func (w *World) ClearFault(faultID string, atNS int64) error {
	af, ok := w.faultsByID[faultID]
	if !ok {
		return fmt.Errorf("world: unknown fault id %q", faultID)
	}
	if atNS <= 0 {
		atNS = w.ClockNS
	}
	af.clearedNS = atNS
	return nil
}

// ListFaults returns the active faults (faults whose contribution is still
// in force at the current clock).
func (w *World) ListFaults() []FaultInfo {
	var out []FaultInfo
	for _, fid := range w.faultOrder {
		af := w.faultsByID[fid]
		if af == nil || (af.clearedNS > 0 && w.ClockNS >= af.clearedNS) {
			continue
		}
		out = append(out, FaultInfo{
			FaultID:  fid,
			EntityID: af.entity,
			Fault:    af.fault.ID,
			OnsetNS:  af.onsetNS,
		})
	}
	return out
}

// ActiveFaultsCount is the number of faults in force at the current clock.
func (w *World) ActiveFaultsCount() int { return len(w.ListFaults()) }
