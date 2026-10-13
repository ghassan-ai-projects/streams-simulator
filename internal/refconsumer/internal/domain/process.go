package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonvalue"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

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

	s.lastSeq = ev.Seq
	r.evaluateReading(ev, s, t)
	return t
}

func (r *Runner) evaluateReading(ev model.SimEvent, s *series, t int64) {
	v, numeric := jsonvalue.Number(ev.Value)
	if !numeric {

		markPresent(s, t)
		return
	}

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
