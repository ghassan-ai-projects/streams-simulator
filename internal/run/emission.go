package run

import (
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"strconv"
)

// onEmit is the world's emitter: perturb -> adapter -> sink -> ledger.
func (r *Run) onEmit(ev model.SimEvent) {
	atNS, err := model.ParseTime(ev.EventTime)
	if err != nil {
		r.fail(fmt.Errorf("strict observed-time guard: event %d has invalid event_time: %w", ev.Seq, err))
		return
	}
	observedNS, err := model.ParseTime(ev.ObservedTime)
	if err != nil {
		r.fail(fmt.Errorf("strict observed-time guard: event %d has invalid observed_time: %w", ev.Seq, err))
		return
	}
	if r.hasObservedTime && observedNS <= r.lastObservedNS {
		r.fail(fmt.Errorf("strict observed-time guard: event %d observed_time %s is not after %s", ev.Seq, ev.ObservedTime, model.FormatTime(r.lastObservedNS)))
		return
	}
	r.lastObservedNS = observedNS
	r.hasObservedTime = true
	// World-state history: what was actually happening when the record was
	// emitted (director-only, for post-hoc analysis).
	states := map[string]float64{}
	for _, name := range r.Config.Domain.StateNames() {
		states[name] = r.World.StateValue(ev.EntityID, name, atNS)
	}
	r.history = append(r.history, stateSnapshot{Seq: ev.Seq, TimeNS: atNS, Entity: ev.EntityID, States: states})
	recs := r.Perturb.Process(ev, atNS)
	for _, d := range recs {
		if d.Malformed {
			// One bad record must not poison a file: render a broken line.
			if err := r.writeMalformed(ev); err != nil {
				r.appendLedger(model.LedgerRecord{
					DeliveryID: d.DeliveryID, Seq: ev.Seq, WorldID: ev.WorldID, EntityID: ev.EntityID,
					Channel: ev.Channel, EventTimeNS: atNS, ObservedTimeNS: atNS,
					Delivered: false, DeliveryReason: model.DeliverySinkError, WrittenAtNS: r.World.Clock(),
				})
				r.fail(err)
				return
			}
			r.appendLedger(model.LedgerRecord{
				DeliveryID: d.DeliveryID, Seq: ev.Seq, WorldID: ev.WorldID, EntityID: ev.EntityID,
				Channel: ev.Channel, EventTimeNS: atNS, ObservedTimeNS: atNS,
				Delivered: true, DeliveryReason: model.DeliveryMangled,
				WrittenAtNS: r.World.Clock(),
			})
			continue
		}
		if !d.Delivered {
			r.appendLedger(model.LedgerRecord{
				DeliveryID: d.DeliveryID, Seq: ev.Seq, WorldID: ev.WorldID, EntityID: ev.EntityID,
				Channel: ev.Channel, EventTimeNS: atNS, ObservedTimeNS: atNS,
				Delivered: false, DeliveryReason: d.Reason,
				WrittenAtNS: r.World.Clock(),
			})
			continue
		}
		if r.evidenceRec != nil {
			r.evidenceRec(d.Event)
		}
		line, err := r.Engine.RenderStreamRecord(&d.Event)
		if err != nil {
			r.appendLedger(model.LedgerRecord{
				DeliveryID: d.DeliveryID, Seq: d.Event.Seq, WorldID: d.Event.WorldID, EntityID: d.Event.EntityID,
				Channel: d.Event.Channel, EventTimeNS: atNS, ObservedTimeNS: atNS,
				Delivered: false, DeliveryReason: model.DeliverySinkError, WrittenAtNS: r.World.Clock(),
			})
			r.fail(err)
			return
		}
		if line == "" {
			r.appendLedger(model.LedgerRecord{
				DeliveryID: d.DeliveryID, Seq: d.Event.Seq, WorldID: d.Event.WorldID, EntityID: d.Event.EntityID,
				Channel: d.Event.Channel, EventTimeNS: atNS, ObservedTimeNS: atNS,
				Delivered: false, DeliveryReason: model.DeliveryOmitted, WrittenAtNS: r.World.Clock(),
			})
			continue
		}
		if err := r.Sink.Write([]byte(line)); err != nil {
			r.appendLedger(model.LedgerRecord{
				DeliveryID: d.DeliveryID, Seq: d.Event.Seq, WorldID: d.Event.WorldID, EntityID: d.Event.EntityID,
				Channel: d.Event.Channel, EventTimeNS: atNS, ObservedTimeNS: atNS,
				Delivered: false, DeliveryReason: model.DeliverySinkError, WrittenAtNS: r.World.Clock(),
			})
			r.fail(err)
			return
		}
		r.noteTraceArrival(d.Event)
		otNS, _ := model.ParseTime(d.Event.ObservedTime)
		r.appendLedger(model.LedgerRecord{
			DeliveryID: d.DeliveryID, Seq: d.Event.Seq, WorldID: d.Event.WorldID, EntityID: d.Event.EntityID,
			Channel: d.Event.Channel, EventTimeNS: atNS, ObservedTimeNS: otNS,
			Delivered: true, DeliveryReason: d.Reason, WrittenAtNS: r.World.Clock(),
		})
	}
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
