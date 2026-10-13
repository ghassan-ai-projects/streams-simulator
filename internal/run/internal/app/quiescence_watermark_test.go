package app

import (
	"context"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestTheQuiescenceWatermarkNeverMovesBack(t *testing.T) {
	t.Parallel()
	spec, a := testBase(t)
	r, err := New(context.Background(), Config{Domain: spec, Adapter: a, Seed: 77, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: model.DefaultStartTimeNS, quiescenceClock: &fakeQuiescenceClock{}})
	if err != nil {
		t.Fatal(err)
	}
	hour := int64(3600 * 1e9)
	start := model.DefaultStartTimeNS
	if _, err := r.Advance(context.Background(), start+2*hour, false); err != nil {
		t.Fatal(err)
	}
	steps := []struct {
		name        string
		reportedNS  int64
		wantThrough int64
	}{
		{"first report", start + hour, start + hour},
		{"earlier report is ignored", start + hour/2, start + hour},
		{"equal report is ignored", start + hour, start + hour},
		{"later report advances", start + 2*hour, start + 2*hour},
		{"zero report is ignored", 0, start + 2*hour},
	}
	for _, s := range steps {
		r.ReportQuiesced(s.reportedNS)
		if got := r.quiescedThrough(); got != s.wantThrough {
			t.Fatalf("%s: watermark = %d, want %d", s.name, got, s.wantThrough)
		}
	}
}
