package app

import (
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/perturb"
)

func TestEndClosesDurableLedgerWithAndWithoutPublication(t *testing.T) {
	for _, publish := range []bool{false, true} {
		t.Run(map[bool]string{false: "memory artifact", true: "published artifact"}[publish], func(t *testing.T) {
			spec, a := testBase(t)
			dir := t.TempDir()
			path := filepath.Join(dir, "ledger.jsonl")
			r, err := New(t.Context(), Config{Domain: spec, Adapter: a, SinkName: model.SinkInproc, LedgerPath: path})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := r.Advance(t.Context(), r.World.Clock()+60*1e9, false); err != nil {
				t.Fatal(err)
			}
			out := ""
			if publish {
				out = dir
			}
			art, err := r.End(out)
			if err != nil {
				t.Fatal(err)
			}
			if !r.durableLedger.Closed() {
				t.Fatal("ledger must be closed after End")
			}
			rows := loadLedgerFile(t, path)
			if int64(len(rows)) != art.Counts.Emitted {
				t.Fatalf("durable rows=%d emitted=%d", len(rows), art.Counts.Emitted)
			}
		})
	}
}

func TestDeliveryLedgerPreservesIdentityAndTerminalReason(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		malformed, delivered bool
		reason, want         string
	}{
		{"drop", false, false, model.DeliveryDroppedByPerturb, model.DeliveryDroppedByPerturb},
		{"malformed", true, true, model.DeliveryMangled, model.DeliveryMangled},
		{"normal", false, true, model.DeliveryOK, model.DeliveryOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			spec, a := testBase(t)
			r, err := New(t.Context(), Config{Domain: spec, Adapter: a, SinkName: model.SinkInproc})
			if err != nil {
				t.Fatal(err)
			}
			ev := model.SimEvent{Seq: 7, WorldID: r.World.ID, EntityID: "site-a/pond-1", Channel: "pond.do_mg_l", EventTime: model.FormatTime(r.World.Clock()), ObservedTime: model.FormatTime(r.World.Clock() + 1), Value: 3.0}
			r.deliver(perturb.Delivered{DeliveryID: 42, Event: ev, Malformed: tc.malformed, Delivered: tc.delivered, Reason: tc.reason})
			ledger := r.Ledger()
			if len(ledger) != 1 || ledger[0].DeliveryID != 42 || ledger[0].Seq != 7 || ledger[0].DeliveryReason != tc.want || ledger[0].Delivered != tc.delivered {
				t.Fatalf("ledger=%+v", ledger)
			}
			if _, err := r.End(""); err != nil {
				t.Fatal(err)
			}
		})
	}
}
