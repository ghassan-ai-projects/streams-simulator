package run

import (
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/perturb"
	"strconv"
)

// onEmit is the world's emitter: perturb -> adapter -> sink -> ledger.
func (r *Run) onEmit(ev model.SimEvent) {
	atNS, err := r.acceptEmissionTime(ev)
	if err != nil {
		return
	}
	r.captureEmissionState(ev, atNS)
	for _, d := range r.Perturb.Process(ev, atNS) {
		if !r.deliverEmission(ev, d, atNS) {
			return
		}
	}
}

func (r *Run) acceptEmissionTime(ev model.SimEvent) (int64, error) {
	atNS, err := model.ParseTime(ev.EventTime)
	if err != nil {
		r.fail(fmt.Errorf("strict observed-time guard: event %d has invalid event_time: %w", ev.Seq, err))
		return 0, r.runErr
	}
	observedNS, err := model.ParseTime(ev.ObservedTime)
	if err != nil {
		r.fail(fmt.Errorf("strict observed-time guard: event %d has invalid observed_time: %w", ev.Seq, err))
		return 0, r.runErr
	}
	if r.hasObservedTime && observedNS <= r.lastObservedNS {
		r.fail(fmt.Errorf("strict observed-time guard: event %d observed_time %s is not after %s", ev.Seq, ev.ObservedTime, model.FormatTime(r.lastObservedNS)))
		return 0, r.runErr
	}
	r.lastObservedNS = observedNS
	r.hasObservedTime = true
	return atNS, nil
}

func (r *Run) captureEmissionState(ev model.SimEvent, atNS int64) {
	// World-state history: what was actually happening when the record was
	// emitted (director-only, for post-hoc analysis).
	states := map[string]float64{}
	for _, name := range r.Config.Domain.StateNames() {
		states[name] = r.World.StateValue(ev.EntityID, name, atNS)
	}
	r.history = append(r.history, stateSnapshot{Seq: ev.Seq, TimeNS: atNS, Entity: ev.EntityID, States: states})
}

func (r *Run) deliverEmission(ev model.SimEvent, d perturb.Delivered, atNS int64) bool {
	if d.Malformed {
		// One bad record must not poison a file: render a broken line.
		if err := r.writeMalformed(ev); err != nil {
			r.recordDelivery(ev, d.DeliveryID, atNS, atNS, false, model.DeliverySinkError)
			r.fail(err)
			return false
		}
		r.recordDelivery(ev, d.DeliveryID, atNS, atNS, true, model.DeliveryMangled)
		return true
	}
	if !d.Delivered {
		r.recordDelivery(ev, d.DeliveryID, atNS, atNS, false, d.Reason)
		return true
	}
	if r.evidenceRec != nil {
		r.evidenceRec(d.Event)
	}
	line, err := r.Engine.RenderStreamRecord(&d.Event)
	if err != nil {
		r.recordDelivery(d.Event, d.DeliveryID, atNS, atNS, false, model.DeliverySinkError)
		r.fail(err)
		return false
	}
	if line == "" {
		r.recordDelivery(d.Event, d.DeliveryID, atNS, atNS, false, model.DeliveryOmitted)
		return true
	}
	if err := r.Sink.Write([]byte(line)); err != nil {
		r.recordDelivery(d.Event, d.DeliveryID, atNS, atNS, false, model.DeliverySinkError)
		r.fail(err)
		return false
	}
	r.noteTraceArrival(d.Event)
	otNS, _ := model.ParseTime(d.Event.ObservedTime)
	r.recordDelivery(d.Event, d.DeliveryID, atNS, otNS, true, d.Reason)
	return true
}

func (r *Run) writeMalformed(ev model.SimEvent) error {
	// A broken line in the adapter's encoding: unterminated JSON.
	if err := r.Sink.Write([]byte(`{"seq":` + strconv.FormatInt(ev.Seq, 10) + `,"broken":`)); err != nil {
		return fmt.Errorf("write malformed event %d: %w", ev.Seq, err)
	}
	return nil
}

// fail aborts the run with an error; the run is marked incomplete.
func (r *Run) fail(err error) {
	if err == nil || r.runErr != nil {
		return
	}
	r.runErr = fmt.Errorf("run %s aborted: %w", r.ID, err)
	r.incomplete = true
	r.reproducible = false
}

// RenderRecord exposes the adapter's per-record rendering (used by the MCP
// trace export path and tests).
func (r *Run) RenderRecord(ev *model.SimEvent) (string, error) {
	line, err := r.Engine.RenderRecord(ev)
	if err != nil {
		return "", fmt.Errorf("run: render: %w", err)
	}
	return line, nil
}
