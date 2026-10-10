package world

// World faults: applied inside the world core, they change hidden state and
// are what a consumer is supposed to diagnose. They are distinct from
// delivery perturbations (the observer is unreliable) and environment faults
// (the consumer is broken) — three independently seeded injection surfaces,
// never conflated.

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
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
	return w.admitFaultInjection(fault, entityID, faultID, onsetNS, params)
}

func faultSeverity(fault *model.Fault, faultID string, params map[string]any) (float64, error) {
	severity := 1.0
	if fault.Onset.Magnitude != 0 {
		severity = math.Abs(fault.Onset.Magnitude)
	}
	if err := validateFaultParameters(faultID, params); err != nil {
		return 0, err
	}
	return declaredSeverity(severity, params)
}

// ClearFault removes a fault's contribution from the given time onward.
func (w *World) ClearFault(faultID string, atNS int64) error {
	af, ok := w.faultsByID[faultID]
	if !ok {
		return fmt.Errorf("world: unknown fault id %q", faultID)
	}
	if atNS <= 0 {
		atNS = w.clockNS
	}
	af.clearedNS = atNS
	return nil
}

// ListFaults returns the active faults (faults whose contribution is still
// in force at the current clock).
func (w *World) ListFaults() []FaultInfo {
	var out []FaultInfo
	for _, id := range w.faultOrder {
		fault := w.faultsByID[id]
		if fault == nil || (fault.clearedNS > 0 && w.clockNS >= fault.clearedNS) {
			continue
		}
		out = append(out, fault.info(id))
	}
	return out
}

// ActiveFaultsCount is the number of faults in force at the current clock.
func (w *World) ActiveFaultsCount() int { return len(w.ListFaults()) }

func (w *World) registerFault(fault *model.Fault, entity string, onset int64, severity float64) string {
	active := &activeFault{fault: fault, entity: entity, onsetNS: onset, severity: severity}
	// Stochastic envelopes start at onset, avoiding epoch-length replay.
	if fault.Onset.Shape == "stochastic" {
		active.walkStep = onset
	}
	id := "f-" + strconv.FormatInt(int64(len(w.faultsByID)), 10)
	w.faultsByID[id] = active
	w.faultOrder = append(w.faultOrder, id)
	for _, affected := range fault.Affects {
		w.faultsByState[affected.State] = append(w.faultsByState[affected.State], active)
	}
	return id
}

func validateFaultParameters(id string, params map[string]any) error {
	for name := range params {
		if name != "severity" {
			return fmt.Errorf("world: fault %q has no parameter %q (declared: severity)", id, name)
		}
	}
	return nil
}

func severityNumber(value any) (float64, error) {
	switch number := value.(type) {
	case float64:
		return number, nil
	case int64:
		return float64(number), nil
	case json.Number:
		parsed, err := number.Float64()
		if err == nil {
			return parsed, nil
		}
	}
	return 0, fmt.Errorf("world: severity must be numeric")
}

func finiteSeverity(severity float64) (float64, error) {
	if math.IsNaN(severity) || math.IsInf(severity, 0) || severity < 0 {
		return 0, fmt.Errorf("world: severity must be finite and non-negative")
	}
	return severity, nil
}

func (fault *activeFault) info(id string) FaultInfo {
	return FaultInfo{FaultID: id, EntityID: fault.entity, Fault: fault.fault.ID, OnsetNS: fault.onsetNS}
}

func (w *World) admitFaultInjection(fault *model.Fault, entityID, faultID string, onsetNS int64, params map[string]any) (string, error) {
	if onsetNS <= 0 {
		onsetNS = w.clockNS
	}
	severity, err := faultSeverity(fault, faultID, params)
	if err != nil {
		return "", err
	}
	return w.registerFault(fault, entityID, onsetNS, severity), nil
}

func declaredSeverity(severity float64, params map[string]any) (float64, error) {
	if value, present := params["severity"]; present {
		var err error
		severity, err = severityNumber(value)
		if err != nil {
			return 0, err
		}
	}
	return finiteSeverity(severity)
}
