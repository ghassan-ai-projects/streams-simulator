package app

import (
	"fmt"
	"strconv"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// onEmit is the world's emitter: perturb -> adapter -> sink -> ledger.
func (r *Run) onEmit(ev model.SimEvent) {
	atNS, err := r.acceptEmissionTime(ev)
	if err != nil {
		return
	}
	r.captureEmissionState(ev, atNS)
	// One delivery path serves live emissions and records released at a
	// boundary: every delivery gets a ledger row of its own event, and a
	// failing sink is recorded for each record rather than silently ending
	// the ledger.
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
	// World-state history: what was actually happening when the record was
	// emitted (director-only, for post-hoc analysis).
	states := map[string]float64{}
	for _, name := range r.Config.Domain.StateNames() {
		states[name] = r.World.StateValue(ev.EntityID, name, atNS)
	}
	r.history = append(r.history, model.StateSnapshot{Seq: ev.Seq, TimeNS: atNS, Entity: ev.EntityID, States: states})
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
