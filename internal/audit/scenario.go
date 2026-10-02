package audit

import (
	"encoding/json"
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/perturb"
	"github.com/ghassan-ai-projects/streams-simulator/internal/truth"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// emissionLog records the delivered events of one world, keyed by channel.
type emissionLog map[string][]int64

// gridValue holds the last delivered value at or before the sample.
func gridValue(series []float64, i int) float64 {
	if i >= 0 && i < len(series) {
		return series[i]
	}
	return 0
}

// build runs one world through the real delivery pipeline — perturbation
// layer then capture of the delivered records in monotonic order — and
// returns the per-channel series on the audit grid plus the delivered
// emission log. The series are immutable evidence, never live world reads.
func (p *Panel) build(entityID string, startNS int64, entityIDs []string, faults map[string]int64, setup []truth.SetupCall, perturbations []Perturbation, durationNS int64) (map[string][]float64, emissionLog, error) {
	w, err := world.New(p.spec, p.seed, "w-audit", startNS, world.Options{
		InitialEntities: entityIDs,
		ForceEffectorOK: true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("build: %w", err)
	}
	layer := perturb.New(w.ID, p.seed, p.spec)
	for _, pert := range perturbations {
		if _, err := layer.Apply(pert.Name, pert.Params, pert.FromNS, pert.UntilNS); err != nil {
			return nil, nil, fmt.Errorf("build: perturb %s: %w", pert.Name, err)
		}
	}
	for _, call := range setup {
		if _, err := w.InvokeEffector(call.Effector, call.EntityID, call.CommandID, call.Args, call.AtNS); err != nil {
			return nil, nil, fmt.Errorf("build: %w", err)
		}
	}
	log := emissionLog{}
	series := map[string][]float64{}
	times := map[string][]int64{}
	deliver := func(d perturb.Delivered) {
		// The audit is per-entity: only the audited entity's delivered
		// records form the series, exactly as a per-entity Reading did.
		if d.Malformed || !d.Delivered || d.Event.EntityID != entityID {
			return
		}
		t, _ := model.ParseTime(d.Event.EventTime)
		log[d.Event.Channel] = append(log[d.Event.Channel], t)
		if v, ok := asFloat(d.Event.Value); ok {
			series[d.Event.Channel] = append(series[d.Event.Channel], v)
			times[d.Event.Channel] = append(times[d.Event.Channel], t)
		}
	}
	w.SetEmitter(func(ev model.SimEvent) {
		t, _ := model.ParseTime(ev.EventTime)
		for _, d := range layer.Process(ev, t) {
			deliver(d)
		}
	})
	for fid, onset := range faults {
		if fid == "" {
			continue
		}
		if _, err := w.InjectFault(entityID, fid, onset, nil); err != nil {
			return nil, nil, fmt.Errorf("build: %w", err)
		}
	}
	// Emit everything up to the horizon so the log is complete. The audit
	// only needs the scenario window, so horizon is bounded by 24h.
	horizon := startNS + 24*3600*1e9
	if durationNS > 0 && startNS+durationNS < horizon {
		horizon = startNS + durationNS
	}
	if _, _, err := w.Advance(horizon); err != nil {
		return nil, nil, fmt.Errorf("build: %w", err)
	}
	for _, d := range layer.Flush(horizon) {
		deliver(d)
	}
	// Grid: value at sample i = the last delivered value at or before the
	// sample instant; 0 before the first delivery.
	n := int((horizon-startNS)/p.sampleNS) + 1
	grid := map[string][]float64{}
	for _, ch := range p.spec.ChannelNames() {
		out := make([]float64, n)
		si := 0
		last := 0.0
		for i := 0; i < n; i++ {
			t := startNS + int64(i)*p.sampleNS
			for si < len(times[ch]) && times[ch][si] <= t {
				last = series[ch][si]
				si++
			}
			out[i] = last
		}
		grid[ch] = out
	}
	return grid, log, nil
}

func asFloat(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case int64:
		return float64(x), true
	case json.Number:
		f, err := x.Float64()
		return f, err == nil
	}
	return 0, false
}
