package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"math"
	"slices"
	"strconv"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type FaultInfo struct {
	FaultID  string  `json:"fault_id"`
	EntityID string  `json:"entity_id"`
	Fault    string  `json:"fault"`
	OnsetNS  int64   `json:"onset_ns"`
	Severity float64 `json:"severity,omitempty"`
}

var ErrFaultOnsetInPast = errors.New("world: fault onset is before the clock of a world that has already run")

var ErrFaultClearInPast = errors.New("world: fault clear is before the clock of a world that has already run")

func (w *World) InjectFault(entityID, faultID string, onsetNS int64, params map[string]any) (string, error) {
	fault := w.Spec.Fault(faultID)
	if fault == nil {
		return "", fmt.Errorf("world: unknown fault %q", faultID)
	}
	ent, ok := w.entities[entityID]
	if !ok {
		return "", fmt.Errorf("world: unknown entity %q", entityID)
	}
	if ent.kind.FaultDecl(faultID) == nil {
		return "", fmt.Errorf("world: fault %q is not declared by the kind %q of entity %q", faultID, ent.Type, entityID)
	}
	return w.admitFaultInjection(fault, entityID, faultID, onsetNS, params)
}

func faultSeverity(fault *model.Fault, faultID string, params map[string]any) (float64, error) {
	severity, err := parameterSeverity(fault, faultID, params)
	if err != nil {
		return 0, fmt.Errorf("world: %w", err)
	}
	return severity, nil
}

func (w *World) CheckFaultParams(faultID string, params map[string]any) error {
	fault := w.Spec.Fault(faultID)
	if fault == nil {
		return fmt.Errorf("the domain declares no fault %q", faultID)
	}
	_, err := parameterSeverity(fault, faultID, params)
	return err
}

func parameterSeverity(fault *model.Fault, faultID string, params map[string]any) (float64, error) {
	severity := 1.0
	if fault.Onset.Magnitude != 0 {
		severity = math.Abs(fault.Onset.Magnitude)
	}
	if err := validateFaultParameters(faultID, params); err != nil {
		return 0, err
	}
	return declaredSeverity(severity, params)
}

func (w *World) ClearFault(faultID string, atNS int64) error {
	return w.clearFaultBy(faultID, atNS, model.ClearedByCommand)
}

func (w *World) clearFaultBy(faultID string, atNS int64, by string) error {
	af, ok := w.faultsByID[faultID]
	if !ok {
		return fmt.Errorf("world: unknown fault id %q", faultID)
	}
	atNS, err := w.admitClear(atNS)
	if err != nil {
		return err
	}
	if af.clearedNS == 0 || atNS < af.clearedNS {
		af.clearedNS, af.clearedBy = atNS, by
		w.checkPulseEdge(af, atNS)
	}
	return nil
}

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

func (w *World) ActiveFaultsCount() int { return len(w.ListFaults()) }

func (w *World) registerFault(fault *model.Fault, entity string, onset int64, severity float64) string {
	active := &activeFault{fault: fault, entity: entity, onsetNS: onset, injectedNS: w.clockNS, severity: severity}
	id := "f-" + strconv.FormatInt(int64(len(w.faultsByID)), 10)
	w.faultsByID[id] = active
	w.faultOrder = append(w.faultOrder, id)
	for _, affected := range fault.Affects {
		w.faultsByState[affected.State] = append(w.faultsByState[affected.State], active)
	}
	w.registerObservationFault(active)
	return id
}

func validateFaultParameters(id string, params map[string]any) error {
	for _, name := range slices.Sorted(maps.Keys(params)) {
		if name != "severity" {
			return fmt.Errorf("fault %q has no parameter %q (declared: severity)", id, name)
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
	return 0, fmt.Errorf("severity must be numeric")
}

func finiteSeverity(severity float64) (float64, error) {
	if math.IsNaN(severity) || math.IsInf(severity, 0) || severity < 0 {
		return 0, fmt.Errorf("severity must be finite and non-negative")
	}
	return severity, nil
}

func (fault *activeFault) info(id string) FaultInfo {
	return FaultInfo{FaultID: id, EntityID: fault.entity, Fault: fault.fault.ID, OnsetNS: fault.onsetNS}
}

func (w *World) admitFaultInjection(fault *model.Fault, entityID, faultID string, onsetNS int64, params map[string]any) (string, error) {
	onsetNS, err := w.admitOnset(onsetNS)
	if err != nil {
		return "", err
	}
	severity, err := faultSeverity(fault, faultID, params)
	if err != nil {
		return "", err
	}
	id := w.registerFault(fault, entityID, onsetNS, severity)
	if fault.Onset.Shape == model.ShapePulse {
		w.schedulePulse(w.faultsByID[id], int64(fault.Onset.DurationS*secondsPerNS))
	}
	return id, nil
}

func (w *World) admitClear(atNS int64) (int64, error) {
	if atNS <= 0 {
		return w.clockNS, nil
	}
	if atNS < w.clockNS && w.clockNS > w.StartNS {
		return 0, fmt.Errorf("%w (clear %d, clock %d)", ErrFaultClearInPast, atNS, w.clockNS)
	}
	return atNS, nil
}

func (w *World) admitOnset(onsetNS int64) (int64, error) {
	if onsetNS <= 0 {
		return w.clockNS, nil
	}
	if onsetNS < w.clockNS && w.clockNS > w.StartNS {
		return 0, fmt.Errorf("%w (onset %d, clock %d)", ErrFaultOnsetInPast, onsetNS, w.clockNS)
	}
	return onsetNS, nil
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
