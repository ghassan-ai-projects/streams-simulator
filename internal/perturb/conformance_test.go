package perturb

// Slice D: conformance tests for the twelve perturbations that previously
// had none. Each asserts the exact transformation a record must undergo at
// rate/window settings that force the perturbation to fire.

import (
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestDelayTailConformance(t *testing.T) {
	l := New("w", 11, testSpec(t))
	if _, err := l.Apply(DelayTail, map[string]any{"mean_s": 30, "sigma_s": 60}, 0, 0); err != nil {
		t.Fatal(err)
	}
	recs := l.Process(ev(0, "num", 1.0), 1000000000)
	if len(recs) != 1 || !recs[0].Delivered {
		t.Fatalf("delay_tail must deliver: %+v", recs)
	}
	if recs[0].Reason != model.DeliveryDelayed {
		t.Fatalf("wrong reason: %s", recs[0].Reason)
	}
	ot, _ := model.ParseTime(recs[0].Event.ObservedTime)
	et, _ := model.ParseTime(recs[0].Event.EventTime)
	if ot <= et {
		t.Fatalf("delay_tail must push observed time forward: ot=%v et=%v", ot, et)
	}
}

func TestIDReuseConformance(t *testing.T) {
	l := New("w", 12, testSpec(t))
	if _, err := l.Apply(IDReuse, map[string]any{"rate": 1.0}, 0, 0); err != nil {
		t.Fatal(err)
	}
	recs := l.Process(ev(3, "num", 2.5), 1000000000)
	if len(recs) != 2 {
		t.Fatalf("id_reuse must double records: %+v", recs)
	}
	if recs[1].Reason != model.DeliveryDuplicated {
		t.Fatalf("reused identity must be marked duplicated: %s", recs[1].Reason)
	}
	if v, ok := recs[1].Event.Value.(float64); !ok || v != 102.5 {
		t.Fatalf("reused identity must shift the payload: %+v", recs[1].Event.Value)
	}
	if recs[0].DeliveryID == recs[1].DeliveryID {
		t.Fatal("reused identity must still carry a distinct delivery id")
	}
}

func TestGrossBackfillConformance(t *testing.T) {
	l := New("w", 13, testSpec(t))
	from, until := int64(1000000000), int64(4000000000)
	if _, err := l.Apply(GrossBackfill, nil, from, until); err != nil {
		t.Fatal(err)
	}
	// Inside the gap: withheld.
	for i := int64(0); i < 3; i++ {
		if recs := l.Process(ev(i, "num", float64(i)), from+i*1e9); len(recs) != 0 {
			t.Fatalf("gross_backfill must withhold inside the gap: %+v", recs)
		}
	}
	// After the gap ends: delivery resumes.
	if recs := l.Process(ev(3, "num", 3.0), until+1e9); len(recs) != 1 {
		t.Fatalf("gross_backfill must resume after the gap: %+v", recs)
	}
	// The boundary flush republishes the withheld records at the gap end.
	flushed := l.Flush(until)
	if len(flushed) != 3 {
		t.Fatalf("expected 3 backfilled records, got %d", len(flushed))
	}
	for i, r := range flushed {
		if !r.Delivered || r.Event.Seq != int64(i) {
			t.Fatalf("backfill must preserve identity: %+v", r)
		}
	}
}

func TestClockSkewConformance(t *testing.T) {
	l := New("w", 14, testSpec(t))
	if _, err := l.Apply(ClockSkew, map[string]any{"offset_s": 300, "sign": "positive"}, 0, 0); err != nil {
		t.Fatal(err)
	}
	recs := l.Process(ev(0, "num", 1.0), 1000000000)
	if recs[0].Reason != model.DeliveryRewritten {
		t.Fatalf("clock_skew must rewrite: %s", recs[0].Reason)
	}
	ot, _ := model.ParseTime(recs[0].Event.ObservedTime)
	et, _ := model.ParseTime(recs[0].Event.EventTime)
	if ot-et != 300*1e9 {
		t.Fatalf("positive skew must add 300s: ot=%v et=%v", ot, et)
	}
	// Negative skew.
	l2 := New("w", 15, testSpec(t))
	if _, err := l2.Apply(ClockSkew, map[string]any{"offset_s": 300, "sign": "negative"}, 0, 0); err != nil {
		t.Fatal(err)
	}
	recs = l2.Process(ev(0, "num", 1.0), 1000000000)
	ot, _ = model.ParseTime(recs[0].Event.ObservedTime)
	if et-ot != 300*1e9 {
		t.Fatalf("negative skew must subtract 300s: ot=%v et=%v", ot, et)
	}
}

func TestNonMonotonicConformance(t *testing.T) {
	l := New("w", 16, testSpec(t))
	if _, err := l.Apply(NonMonotonic, nil, 0, 0); err != nil {
		t.Fatal(err)
	}
	// A late-observed record (observed > event) is rewritten to precede it.
	rec := ev(0, "num", 1.0)
	et, _ := model.ParseTime(rec.EventTime)
	rec.ObservedTime = model.FormatTime(et + 1e9)
	recs := l.Process(rec, 1000000000)
	if recs[0].Reason != model.DeliveryRewritten {
		t.Fatalf("non_monotonic must rewrite: %s", recs[0].Reason)
	}
	ot, _ := model.ParseTime(recs[0].Event.ObservedTime)
	if ot >= et {
		t.Fatalf("observed time must precede event time: ot=%v et=%v", ot, et)
	}
}

func TestUnitMismatchConformance(t *testing.T) {
	l := New("w", 17, testSpec(t))
	if _, err := l.Apply(UnitMismatch, nil, 0, 0); err != nil {
		t.Fatal(err)
	}
	rec := ev(0, "num", 1.0)
	rec.Unit = "u"
	recs := l.Process(rec, 1000000000)
	if recs[0].Event.Unit != "err" {
		t.Fatalf("unit must be mangled to err: %+v", recs[0].Event)
	}
	if recs[0].Reason != model.DeliveryMangled {
		t.Fatalf("unit mismatch must mangle: %s", recs[0].Reason)
	}
}

func TestOversizeConformance(t *testing.T) {
	l := New("w", 18, testSpec(t))
	if _, err := l.Apply(Oversize, map[string]any{"bytes": 8}, 0, 0); err != nil {
		t.Fatal(err)
	}
	recs := l.Process(ev(0, "num", 1.0), 1000000000)
	if s, ok := recs[0].Event.Value.(string); !ok || len(s) != 8 {
		t.Fatalf("oversize must inflate the value to the byte limit: %+v", recs[0].Event.Value)
	}
	if recs[0].Reason != model.DeliveryMangled {
		t.Fatalf("oversize must mangle: %s", recs[0].Reason)
	}
}

func TestMalformedConformance(t *testing.T) {
	l := New("w", 19, testSpec(t))
	if _, err := l.Apply(Malformed, nil, 0, 0); err != nil {
		t.Fatal(err)
	}
	recs := l.Process(ev(0, "num", 1.0), 1000000000)
	if !recs[0].Malformed {
		t.Fatalf("malformed must mark the record: %+v", recs[0])
	}
	if recs[0].Reason != model.DeliveryMangled {
		t.Fatalf("malformed must mangle: %s", recs[0].Reason)
	}
}

func TestNaNInfConformance(t *testing.T) {
	l := New("w", 20, testSpec(t))
	if _, err := l.Apply(NaNInf, nil, 0, 0); err != nil {
		t.Fatal(err)
	}
	recs := l.Process(ev(0, "num", 1.0), 1000000000)
	s, ok := recs[0].Event.Value.(string)
	if !ok || (s != "NaN" && s != "Infinity" && s != "-Infinity") {
		t.Fatalf("nan_inf must replace the value: %+v", recs[0].Event.Value)
	}
	if recs[0].Reason != model.DeliveryMangled {
		t.Fatalf("nan_inf must mangle: %s", recs[0].Reason)
	}
}

func TestStormConformance(t *testing.T) {
	l := New("w", 21, testSpec(t))
	if _, err := l.Apply(Storm, map[string]any{"multiplier": 3}, 0, 0); err != nil {
		t.Fatal(err)
	}
	recs := l.Process(ev(0, "num", 1.0), 1000000000)
	if len(recs) != 3 {
		t.Fatalf("storm must multiply records: %d", len(recs))
	}
	if recs[1].Reason != model.DeliveryDuplicated || recs[2].Reason != model.DeliveryDuplicated {
		t.Fatalf("storm copies must be marked duplicated: %+v", recs)
	}
	if recs[0].DeliveryID == recs[1].DeliveryID || recs[1].DeliveryID == recs[2].DeliveryID {
		t.Fatal("storm copies need distinct delivery ids")
	}
}

func TestTimeEncodingConformance(t *testing.T) {
	l := New("w", 22, testSpec(t))
	if _, err := l.Apply(TimeEncoding, nil, 0, 0); err != nil {
		t.Fatal(err)
	}
	recs := l.Process(ev(0, "num", 1.0), 1000000000)
	if recs[0].Reason != model.DeliveryRewritten {
		t.Fatalf("time_encoding must rewrite: %s", recs[0].Reason)
	}
	if recs[0].Event.ObservedTime == recs[0].Event.EventTime {
		t.Fatal("time_encoding must change the observed time representation")
	}
	if _, err := model.ParseTime(recs[0].Event.ObservedTime); err != nil {
		t.Fatalf("alternate encoding must still parse: %v", err)
	}
}

func TestPrecisionEdgeConformance(t *testing.T) {
	l := New("w", 23, testSpec(t))
	if _, err := l.Apply(PrecisionEdge, nil, 0, 0); err != nil {
		t.Fatal(err)
	}
	// A timestamp with sub-second precision: the truncation must shorten it.
	ev := ev(0, "num", 1.0)
	ev.EventTime = model.FormatTime(1000000000 + 123456789)
	ev.ObservedTime = ev.EventTime
	recs := l.Process(ev, 1000000000)
	if recs[0].Reason != model.DeliveryRewritten {
		t.Fatalf("precision_edge must rewrite: %s", recs[0].Reason)
	}
	if len(recs[0].Event.ObservedTime) >= len(ev.EventTime) {
		t.Fatalf("precision_edge must shorten the timestamp: %q vs %q", recs[0].Event.ObservedTime, ev.EventTime)
	}
}

func TestUnitMismatchSkipsUnitlessChannels(t *testing.T) {
	l := New("w", 24, testSpec(t))
	if _, err := l.Apply(UnitMismatch, nil, 0, 0); err != nil {
		t.Fatal(err)
	}
	// The heartbeat channel has no unit: unit mismatch must pass through
	// undamaged.
	if recs := l.Process(ev(0, "heartbeat", 0.0), 1000000000); recs[0].Reason == model.DeliveryMangled {
		t.Fatalf("unitless records must pass through untouched: %+v", recs[0])
	}
}

func TestNonMonotonicPassesThroughOrderedRecords(t *testing.T) {
	l := New("w", 25, testSpec(t))
	if _, err := l.Apply(NonMonotonic, nil, 0, 0); err != nil {
		t.Fatal(err)
	}
	recs := l.Process(ev(0, "num", 1.0), 1000000000)
	if strings.Contains(recs[0].Reason, "rewritten") {
		t.Fatalf("non_monotonic must not touch ordered records: %+v", recs[0])
	}
}
