package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
)

const maxSeverity = 1e6

type FaultWeight struct {
	Weight        float64   `json:"weight"`
	SeverityRange []float64 `json:"severity_range,omitempty"`
}

type faultWeightObject FaultWeight

func (w *FaultWeight) UnmarshalJSON(raw []byte) error {
	if json.Unmarshal(raw, &w.Weight) == nil {
		return nil
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var object faultWeightObject
	if err := decoder.Decode(&object); err != nil {
		return fmt.Errorf("a fault weight is a number or {weight, severity_range}: %w", err)
	}
	*w = FaultWeight(object)
	return nil
}

func (w FaultWeight) severityRangeHolds() bool {
	if w.SeverityRange == nil {
		return true
	}
	if len(w.SeverityRange) != 2 {
		return false
	}
	lo, hi := w.SeverityRange[0], w.SeverityRange[1]
	return lo >= 0 && lo <= hi && hi <= maxSeverity
}

func (p *CampaignPhase) weights() map[string]float64 {
	weights := make(map[string]float64, len(p.Faults))
	for _, id := range slices.Sorted(maps.Keys(p.Faults)) {
		weights[id] = p.Faults[id].Weight
	}
	return weights
}

func (p *CampaignPhase) checkSeverityRanges() error {
	for _, id := range slices.Sorted(maps.Keys(p.Faults)) {
		if !p.Faults[id].severityRangeHolds() {
			return specErrorf("phase %q: fault %q: severity_range is [lo, hi] with 0 <= lo <= hi <= %g", p.Name, id, maxSeverity)
		}
	}
	return nil
}

// drawSeverities draws from the run's own severity stream, in disturbance order, only for faults with a range, so a
// phase without ranges draws exactly what it drew before ranges existed.
func (p *CampaignPhase) drawSeverities(disturbances []model.Disturbance, rng *randutil.SplitMix64) {
	for i := range disturbances {
		if span := p.Faults[disturbances[i].Fault].SeverityRange; span != nil {
			severity := span[0] + rng.Float64()*(span[1]-span[0])
			disturbances[i].Severity = &severity
		}
	}
}
