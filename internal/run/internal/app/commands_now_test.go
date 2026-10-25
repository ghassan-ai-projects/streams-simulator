package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

const nowPond = "site-a/pond-1"

func nowRun(t *testing.T) *Run {
	t.Helper()
	spec, a := testBase(t)
	r, err := New(t.Context(), Config{Domain: spec, Adapter: a, Seed: 8, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: model.DefaultStartTimeNS, StartTimeSet: true})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestAFaultInjectedNowStartsAtTheRunsCurrentClockAndIsLoggedInOrder(t *testing.T) {
	t.Parallel()
	r := nowRun(t)
	at := model.DefaultStartTimeNS + 600e9
	if _, err := r.Advance(t.Context(), at, false); err != nil {
		t.Fatal(err)
	}
	faultID, atNS, err := r.InjectFaultNow(nowPond, "aerator_failure", map[string]any{})
	if err != nil || faultID == "" || atNS != at {
		t.Fatalf("InjectFaultNow = %q at %d, %v, want a fault at the clock %d", faultID, atNS, err, at)
	}
	art, err := r.End("")
	if err != nil || len(art.CommandLog) != 2 {
		t.Fatalf("End = %d logged commands, %v, want the advance and the fault", len(art.CommandLog), err)
	}
	logged := art.CommandLog[1]
	if logged.Op != model.OpFaultInject || logged.AtNS != at || logged.Args["onset_ns"] != at {
		t.Fatalf("logged %+v, want a fault.inject at the clock", logged)
	}
}

func TestAnInvalidFaultNowIsRefusedAndLeavesNoCommandBehind(t *testing.T) {
	t.Parallel()
	r := nowRun(t)
	if _, _, err := r.InjectFaultNow(nowPond, "no-such-fault", nil); err == nil || !strings.Contains(err.Error(), `unknown fault "no-such-fault"`) {
		t.Fatalf("an unknown fault: err = %v, want the world's refusal naming it", err)
	}
	if art, err := r.End(""); err != nil || len(art.CommandLog) != 0 {
		t.Fatalf("End = %d commands, %v, want none", len(art.CommandLog), err)
	}
}

func TestAnEffectorInvokedNowIsAcknowledgedAtTheClockAndReplaysToTheSameTrace(t *testing.T) {
	t.Parallel()
	spec, a := testBase(t)
	r := nowRun(t)
	at := model.DefaultStartTimeNS + 1200e9
	if _, err := r.Advance(t.Context(), at, false); err != nil {
		t.Fatal(err)
	}
	args := map[string]any{"pond_id": nowPond, "level": 1.0}
	ack, atNS, err := r.InvokeAcknowledgedNow("start_aerator", nowPond, "now-1", args)
	if err != nil || ack == nil || !ack.Accepted || atNS != at {
		t.Fatalf("InvokeAcknowledgedNow = %+v at %d, %v, want an accepted call at %d", ack, atNS, err, at)
	}
	if _, err := r.Advance(t.Context(), at+600e9, false); err != nil {
		t.Fatal(err)
	}
	art, err := r.End("")
	if err != nil || art.CommandLog[1].Op != model.OpEffectorInvoke || art.CommandLog[1].AtNS != at {
		t.Fatalf("End = %+v, %v, want the call logged between the two advances", art.CommandLog, err)
	}
	res, err := ReplayArtifact(context.Background(), art, spec, a, "")
	if err != nil || !res.Matches {
		t.Fatalf("replay = %+v, %v, want the recorded trace", res, err)
	}
}

func TestAnInvalidEffectorCallNowIsAnInvalidInvocationError(t *testing.T) {
	t.Parallel()
	r := nowRun(t)
	_, _, err := r.InvokeAcknowledgedNow("no-such-effector", nowPond, "now-1", nil)
	if !errors.Is(err, world.ErrInvalidInvocation) {
		t.Fatalf("an unknown effector: err = %v, want ErrInvalidInvocation", err)
	}
}
