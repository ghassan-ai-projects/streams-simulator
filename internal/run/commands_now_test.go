package run_test

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestTheFacadeAppliesAFaultAndAnEffectorCallAtTheRunsClock(t *testing.T) {
	t.Parallel()
	spec, adapterSpec, err := run.LoadInputs(testsupport.DomainsDir(), testsupport.AdaptersDir(), "aquaculture-pond", "native-jsonl")
	if err != nil {
		t.Fatal(err)
	}
	r, err := run.New(t.Context(), run.Config{Domain: spec, Adapter: adapterSpec, Seed: 2, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: model.DefaultStartTimeNS, StartTimeSet: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(t.Context(), model.DefaultStartTimeNS+300e9, false); err != nil {
		t.Fatal(err)
	}
	if _, at, err := r.InjectFaultNow("site-a/pond-1", "do_probe_fouling", nil); err != nil || at != model.DefaultStartTimeNS+300e9 {
		t.Fatalf("InjectFaultNow at %d, %v, want the run's clock", at, err)
	}
	ack, at, err := r.InvokeAcknowledgedNow("start_aerator", "site-a/pond-1", "now-1", map[string]any{"pond_id": "site-a/pond-1"})
	if err != nil || ack == nil || !ack.Accepted || at != model.DefaultStartTimeNS+300e9 {
		t.Fatalf("InvokeAcknowledgedNow = %+v at %d, %v, want an accepted call at the run's clock", ack, at, err)
	}
}
