package app

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/perturb"
)

func (r *Run) appendLedger(rec model.LedgerRecord) {
	r.ledger = append(r.ledger, rec)
	r.journal.Delivery(r.journalNow(), rec)
	if r.durableLedger != nil {
		r.durableLedger.Append(rec)
	}
}

func (r *Run) flushDurable() error {
	if r.durableLedger != nil {
		if err := r.durableLedger.Flush(); err != nil {
			return err
		}
	}
	if err := r.journal.Flush(); err != nil {
		return fmt.Errorf("run: %w", err)
	}
	return r.flushSink()
}

func (r *Run) flushSink() error {
	if f, ok := r.Sink.(interface{ Flush() error }); ok {
		if err := f.Flush(); err != nil {
			return fmt.Errorf("run: flush sink: %w", err)
		}
	}
	return nil
}

func (r *Run) deliver(d perturb.Delivered) {
	if d.Malformed {
		r.deliverMalformedEvent(d.Event, d.DeliveryID)
		return
	}
	if !d.Delivered {
		r.recordTerminalDelivery(d.Event, d.DeliveryID, false, d.Reason)
		return
	}
	r.deliverRendered(d)
}

func (r *Run) noteTraceArrival(ev model.SimEvent) {
	arrivalNS, err := model.ParseTime(ev.ObservedTime)
	if err != nil {
		return
	}
	if !r.hasTraceArrivalTime || arrivalNS > r.lastTraceArrivalNS {
		r.lastTraceArrivalNS = arrivalNS
		r.hasTraceArrivalTime = true
	}
}

func (r *Run) traceEndTime() int64 {
	end := r.worldEndTimeNS
	if r.hasTraceArrivalTime && r.lastTraceArrivalNS >= end {
		end = r.lastTraceArrivalNS + 1
	}
	return end
}

func (r *Run) recordDelivery(ev model.SimEvent, deliveryID uint64, atNS, observedNS int64, delivered bool, reason string) {
	r.appendLedger(model.LedgerRecord{
		DeliveryID: deliveryID, Seq: ev.Seq, WorldID: ev.WorldID, EntityID: ev.EntityID,
		Channel: ev.Channel, EventTimeNS: atNS, ObservedTimeNS: observedNS,
		Delivered: delivered, DeliveryReason: reason, WrittenAtNS: r.World.Clock(),
	})
}

func (r *Run) deliverRendered(d perturb.Delivered) {
	r.observeDelivery(d.Event)
	line, err := r.Engine.RenderStreamRecord(&d.Event)
	if err != nil {
		r.failPendingDelivery(d, err)
		return
	}
	if line == "" {
		r.recordTerminalDelivery(d.Event, d.DeliveryID, false, model.DeliveryOmitted)
		return
	}
	r.recordEvidence(d.Event)
	r.writePendingDelivery(d, line)
}

func (r *Run) recordEvidence(event model.SimEvent) {
	if r.evidenceRec != nil {
		r.evidenceRec(event)
	}
}

func (r *Run) writePendingDelivery(d perturb.Delivered, line string) {
	if err := r.Sink.Write([]byte(line)); err != nil {
		r.failPendingDelivery(d, err)
		return
	}
	r.noteTraceArrival(d.Event)
	r.journal.Observation(r.journalNow(), line)
	atNS, _ := model.ParseTime(d.Event.EventTime)
	observedNS, _ := model.ParseTime(d.Event.ObservedTime)
	r.recordDelivery(d.Event, d.DeliveryID, atNS, observedNS, true, d.Reason)
}

func (r *Run) failPendingDelivery(d perturb.Delivered, err error) {
	r.recordTerminalDelivery(d.Event, d.DeliveryID, false, model.DeliverySinkError)
	r.fail(err)
}

func (r *Run) recordTerminalDelivery(ev model.SimEvent, id uint64, delivered bool, reason string) {
	atNS, _ := model.ParseTime(ev.EventTime)
	r.recordDelivery(ev, id, atNS, atNS, delivered, reason)
}

func (r *Run) deliverMalformedEvent(ev model.SimEvent, id uint64) bool {

	if err := r.writeMalformed(ev); err != nil {
		r.recordTerminalDelivery(ev, id, false, model.DeliverySinkError)
		r.fail(err)
		return false
	}
	r.journal.Malformed(r.journalNow(), malformedLine(ev))
	r.recordTerminalDelivery(ev, id, true, model.DeliveryMangled)
	return true
}

func (r *Run) observeDelivery(event model.SimEvent) {
	if r.deliveryObserver != nil {
		r.deliveryObserver(event)
	}
}
