package run

import (
	"encoding/json"
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/perturb"
)

// appendLedger records one delivery row in memory and, when durable
// persistence is configured, appends it to the ledger file immediately. The
// file is flushed at command boundaries, so a crash between boundaries loses
// nothing that was acknowledged at a boundary.
func (r *Run) appendLedger(rec model.LedgerRecord) {
	r.ledger = append(r.ledger, rec)
	if r.ledgerWriter != nil {
		if raw, err := json.Marshal(rec); err == nil {
			_, _ = r.ledgerWriter.Write(raw)
			_ = r.ledgerWriter.WriteByte('\n')
		}
	}
}

// flushDurable pushes the ledger writer and the file sink (when present) to
// their file descriptors. Called at every command boundary; End adds fsync.
func (r *Run) flushDurable() error {
	if r.ledgerWriter != nil {
		if err := r.ledgerWriter.Flush(); err != nil {
			return fmt.Errorf("run: flush ledger: %w", err)
		}
	}
	if f, ok := r.Sink.(interface{ Flush() error }); ok {
		if err := f.Flush(); err != nil {
			return fmt.Errorf("run: flush sink: %w", err)
		}
	}
	return nil
}

func (r *Run) deliver(d perturb.Delivered) {
	if d.Malformed {
		if err := r.writeMalformed(d.Event); err != nil {
			atNS, _ := model.ParseTime(d.Event.EventTime)
			r.recordDelivery(d.Event, d.DeliveryID, atNS, atNS, false, model.DeliverySinkError)
			r.fail(err)
			return
		}
		atNS, _ := model.ParseTime(d.Event.EventTime)
		r.recordDelivery(d.Event, d.DeliveryID, atNS, atNS, true, model.DeliveryMangled)
		return
	}
	if !d.Delivered {
		atNS, _ := model.ParseTime(d.Event.EventTime)
		r.recordDelivery(d.Event, d.DeliveryID, atNS, atNS, false, d.Reason)
		return
	}
	line, err := r.Engine.RenderStreamRecord(&d.Event)
	if err != nil {
		atNS, _ := model.ParseTime(d.Event.EventTime)
		r.recordDelivery(d.Event, d.DeliveryID, atNS, atNS, false, model.DeliverySinkError)
		r.fail(err)
		return
	}
	if line == "" {
		atNS, _ := model.ParseTime(d.Event.EventTime)
		r.recordDelivery(d.Event, d.DeliveryID, atNS, atNS, false, model.DeliveryOmitted)
		return
	}
	if r.evidenceRec != nil {
		r.evidenceRec(d.Event)
	}
	if err := r.Sink.Write([]byte(line)); err != nil {
		atNS, _ := model.ParseTime(d.Event.EventTime)
		r.recordDelivery(d.Event, d.DeliveryID, atNS, atNS, false, model.DeliverySinkError)
		r.fail(err)
		return
	}
	r.noteTraceArrival(d.Event)
	atNS, _ := model.ParseTime(d.Event.EventTime)
	otNS, _ := model.ParseTime(d.Event.ObservedTime)
	r.recordDelivery(d.Event, d.DeliveryID, atNS, otNS, true, d.Reason)
}

// noteTraceArrival records the latest arrival timestamp that was rendered
// into the trace. Perturbations can rewrite an event's observed_time after
// the world has emitted it, so the trailer must use delivered records rather
// than the world clock or the native emission watermark.
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

// traceEndTime returns a deterministic trailer horizon strictly after every
// rendered event arrival while preserving the scenario horizon when it is
// already later.
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
