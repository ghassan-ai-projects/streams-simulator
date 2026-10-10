package run

import (
	"fmt"
	"strconv"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/perturb"
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
	if err := r.acceptObservedOrder(ev); err != nil {
		return 0, err
	}
	return atNS, nil
}

func (r *Run) captureEmissionState(ev model.SimEvent, atNS int64) {
	// World-state history: what was actually happening when the record was
	// emitted (director-only, for post-hoc analysis).
	states := map[string]float64{}
	for _, name := range r.Config.Domain.StateNames() {
		states[name] = r.World.StateValue(ev.EntityID, name, atNS)
	}
	r.history = append(r.history, model.StateSnapshot{Seq: ev.Seq, TimeNS: atNS, Entity: ev.EntityID, States: states})
}

func (r *Run) deliverEmission(ev model.SimEvent, d perturb.Delivered, atNS int64) bool {
	if d.Malformed {
		return r.deliverMalformedEvent(ev, d.DeliveryID)
	}
	if !d.Delivered {
		r.recordDelivery(ev, d.DeliveryID, atNS, atNS, false, d.Reason)
		return true
	}
	if r.evidenceRec != nil {
		r.evidenceRec(d.Event)
	}
	return r.renderEmission(d, atNS)
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

func (r *Run) renderEmission(d perturb.Delivered, atNS int64) bool {
	line, err := r.Engine.RenderStreamRecord(&d.Event)
	if err != nil {
		return r.failEmission(d, atNS, err)
	}
	if line == "" {
		r.recordDelivery(d.Event, d.DeliveryID, atNS, atNS, false, model.DeliveryOmitted)
		return true
	}
	return r.writeEmission(d, line, atNS)
}

func (r *Run) writeEmission(d perturb.Delivered, line string, atNS int64) bool {
	if err := r.Sink.Write([]byte(line)); err != nil {
		return r.failEmission(d, atNS, err)
	}
	r.noteTraceArrival(d.Event)
	observedNS, _ := model.ParseTime(d.Event.ObservedTime)
	r.recordDelivery(d.Event, d.DeliveryID, atNS, observedNS, true, d.Reason)
	return true
}

func (r *Run) failEmission(d perturb.Delivered, atNS int64, err error) bool {
	r.recordDelivery(d.Event, d.DeliveryID, atNS, atNS, false, model.DeliverySinkError)
	r.fail(err)
	return false
}
