package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// Process consumes one batch of native-format JSONL sim events (delivery
// order) and returns the consumer's cumulative verdict. endNS is the world
// instant the batch covers, used for the quiescence watermark and absence
// detection. Sequences already processed by an earlier call are skipped, so
// a harness may re-feed the full trace without double-counting.
func (r *Runner) Process(trace []byte, endNS int64) (*model.Verdict, error) {
	now := max(r.processTrace(trace), endNS)
	r.detectSilence(now)
	return r.reportVerdict(now)
}

func (r *Runner) processTrace(trace []byte) int64 {
	var now int64
	for _, line := range splitLines(trace) {
		var ev model.SimEvent
		if err := model.DecodeBytes([]byte(line), &ev); err != nil {
			// A malformed record is itself evidence: report it and move on.
			continue
		}
		now = max(now, r.consumeEvent(ev))
	}
	return now
}

func (r *Runner) consumeEvent(ev model.SimEvent) int64 {
	if !r.admitDelivery(ev) {
		return 0
	}
	t, _ := model.ParseTime(ev.ObservedTime)
	s := r.seriesFor(ev.EntityID, ev.Channel)
	recordArrivalGap(s, t)
	r.evaluateReading(ev, s, t)
	return t
}

func (r *Runner) evaluateReading(ev model.SimEvent, s *series, t int64) {
	v, numeric := asFloat(ev.Value)
	if !numeric {
		// Heartbeats and strings refresh the presence marker.
		markPresent(s, t)
		return
	}
	// Compare against the baseline before adding the anomaly.
	suspicious := r.suspicious(s, v)
	r.appendReading(s, v, t)
	r.gradeReading(ev, s, t, suspicious)
}

func (r *Runner) detectSilence(now int64) {
	for _, key := range r.sortedSeriesKeys() {
		r.detectSeriesSilence(key, r.stats[key], now)
	}
}

func (r *Runner) reportVerdict(now int64) (*model.Verdict, error) {
	if err := r.reportProcessedWatermark(now); err != nil {
		return nil, err
	}
	v := r.consumerVerdict()
	if err := r.report.SubmitVerdict(v); err != nil {
		return nil, fmt.Errorf("Process: %w", err)
	}
	return v, nil
}
