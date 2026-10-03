package audit

import (
	"encoding/json"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/perturb"
	"github.com/ghassan-ai-projects/streams-simulator/internal/truth"
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
	w, layer, err := p.prepareWorld(startNS, entityIDs, setup, perturbations)
	if err != nil {
		return nil, nil, err
	}
	capture := auditCapture{entityID: entityID, log: emissionLog{}, series: map[string][]float64{}, times: map[string][]int64{}}
	capture.bindEmitter(w, layer)
	if err := injectAuditFaults(w, entityID, faults); err != nil {
		return nil, nil, err
	}
	horizon := auditHorizon(startNS, durationNS)
	if err := capture.finishWorld(w, layer, horizon); err != nil {
		return nil, nil, err
	}
	return capture.sampleGrid(p, startNS, horizon), capture.log, nil
}

type auditCapture struct {
	entityID string
	log      emissionLog
	series   map[string][]float64
	times    map[string][]int64
}

func (capture *auditCapture) deliver(d perturb.Delivered) {
	// The audit is per-entity: only the audited entity's delivered
	// records form the series, exactly as a per-entity Reading did.
	if d.Malformed || !d.Delivered || d.Event.EntityID != capture.entityID {
		return
	}
	t, _ := model.ParseTime(d.Event.EventTime)
	capture.log[d.Event.Channel] = append(capture.log[d.Event.Channel], t)
	if v, ok := asFloat(d.Event.Value); ok {
		capture.series[d.Event.Channel] = append(capture.series[d.Event.Channel], v)
		capture.times[d.Event.Channel] = append(capture.times[d.Event.Channel], t)
	}
}

func (capture *auditCapture) sampleGrid(p *Panel, startNS, horizon int64) map[string][]float64 {
	// Hold the last delivered value at each sample; zero before delivery.
	n := int((horizon-startNS)/p.sampleNS) + 1
	grid := map[string][]float64{}
	for _, channel := range p.spec.ChannelNames() {
		grid[channel] = sampleChannel(capture.series[channel], capture.times[channel], n, startNS, p.sampleNS)
	}
	return grid
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

func sampleChannel(values []float64, times []int64, n int, start, step int64) []float64 {
	out := make([]float64, n)
	index := 0
	last := 0.0
	for i := range out {
		at := start + int64(i)*step
		for index < len(times) && times[index] <= at {
			last = values[index]
			index++
		}
		out[i] = last
	}
	return out
}
