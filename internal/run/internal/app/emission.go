package app

import (
	"fmt"
	"strconv"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (r *Run) onEmit(ev model.SimEvent) {
	atNS, err := r.acceptEmissionTime(ev)
	if err != nil {
		return
	}
	r.captureEmissionState(ev, atNS)

	for _, d := range r.Perturb.Process(ev, atNS) {
		r.deliver(d)
	}
}

func (r *Run) acceptEmissionTime(ev model.SimEvent) (int64, error) {
	atNS, err := model.ParseTime(ev.EventTime)
	if err != nil {
		r.fail(fmt.Errorf("strict observed-time guard: event %d has invalid event_time: %w", ev.Seq, err))
		return 0, r.runErr
	}
	if err := r.acceptObservedOrder(ev); err != nil {
		return 0, err
	}
	return atNS, nil
}

func (r *Run) captureEmissionState(ev model.SimEvent, atNS int64) {

	states := map[string]float64{}
	for _, name := range r.World.StateNamesOf(ev.EntityID) {
		states[name] = r.World.StateValue(ev.EntityID, name, atNS)
	}
	snapshot := model.StateSnapshot{Seq: ev.Seq, TimeNS: atNS, Entity: ev.EntityID, States: states}
	r.history = append(r.history, snapshot)
	r.journal.State(r.journalNow(), snapshot)
}

func malformedLine(ev model.SimEvent) string {
	return `{"seq":` + strconv.FormatInt(ev.Seq, 10) + `,"broken":`
}

func (r *Run) writeMalformed(ev model.SimEvent) error {

	if err := r.Sink.Write([]byte(malformedLine(ev))); err != nil {
		return fmt.Errorf("write malformed event %d: %w", ev.Seq, err)
	}
	return nil
}

func (r *Run) fail(err error) {
	if err == nil || r.runErr != nil {
		return
	}
	r.runErr = fmt.Errorf("run %s aborted: %w", r.ID, err)
	r.incomplete = true
	r.reproducible = false
}

func (r *Run) RenderRecord(ev *model.SimEvent) (string, error) {
	line, err := r.Engine.RenderRecord(ev)
	if err != nil {
		return "", fmt.Errorf("run: render: %w", err)
	}
	return line, nil
}

func (r *Run) acceptObservedOrder(ev model.SimEvent) error {
	observedNS, err := model.ParseTime(ev.ObservedTime)
	if err != nil {
		r.fail(fmt.Errorf("strict observed-time guard: event %d has invalid observed_time: %w", ev.Seq, err))
		return r.runErr
	}
	if r.hasObservedTime && observedNS <= r.lastObservedNS {
		r.fail(fmt.Errorf("strict observed-time guard: event %d observed_time %s is not after %s", ev.Seq, ev.ObservedTime, model.FormatTime(r.lastObservedNS)))
		return r.runErr
	}
	r.lastObservedNS, r.hasObservedTime = observedNS, true
	return nil
}
