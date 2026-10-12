package domain

import (
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestCheckEffectorArgsJudgesWithoutInvoking(t *testing.T) {
	t.Parallel()
	spec := testSpec(t, func(s *model.DomainSpec) {
		s.Effectors = []model.Effector{{
			Name:       "act",
			ArgsSchema: map[string]any{"type": "object", "required": []any{"level"}},
			Ack:        model.Ack{LatencyMS: model.Latency{Mean: 100}},
			Effect:     model.Effect{StateDeltas: []model.StateDelta{{State: "x", Delta: 1}}},
		}}
	})
	w := newTestWorld(t, spec, 1, 0)
	cases := []struct {
		name, effector string
		args           map[string]any
		want           string
	}{
		{"valid", "act", map[string]any{"level": 1}, ""},
		{"missing argument", "act", map[string]any{}, `missing required property "level"`},
		{"unknown effector", "nope", nil, `the domain declares no effector "nope"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assertViolation(t, w.CheckEffectorArgs(tc.effector, tc.args), tc.want)
		})
	}
	if calls := w.EffectorCalls(); len(calls) != 0 {
		t.Fatalf("checking invoked the effector: %+v", calls)
	}
}

// The fault check states the same violations InjectFault refuses, without
// the world prefix and without injecting anything.
func TestCheckFaultParamsJudgesWithoutInjecting(t *testing.T) {
	t.Parallel()
	w := newTestWorld(t, testSpec(t, nil), 23, model.DefaultStartTimeNS)
	cases := []struct {
		name, fault string
		params      map[string]any
		want        string
	}{
		{"no params", "f1", nil, ""},
		{"declared severity", "f1", map[string]any{"severity": 2.5}, ""},
		{"unknown fault", "nope", nil, `the domain declares no fault "nope"`},
		{"unknown parameter", "f1", map[string]any{"bogus": 1}, `fault "f1" has no parameter "bogus" (declared: severity)`},
		{"non-numeric severity", "f1", map[string]any{"severity": "high"}, "severity must be numeric"},
		{"negative severity", "f1", map[string]any{"severity": -1.0}, "severity must be finite and non-negative"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := w.CheckFaultParams(tc.fault, tc.params)
			assertViolation(t, err, tc.want)
			if err != nil && strings.HasPrefix(err.Error(), "world:") {
				t.Fatalf("violation %q carries the runtime prefix", err)
			}
		})
	}
	if faults := w.ListFaults(); len(faults) != 0 {
		t.Fatalf("checking injected a fault: %+v", faults)
	}
}

func TestNewRefusesAnUnknownForcedFailureMode(t *testing.T) {
	t.Parallel()
	_, err := New(testSpec(t, nil), 1, "", model.DefaultStartTimeNS, Options{ForceFailureMode: "bogus_mode"})
	assertViolation(t, err, `world: unknown failure mode "bogus_mode"`)
	for _, mode := range []FailureMode{ModeOK, ModeSlow, ModeAckLost, ModeReject, ModePartial, ModeConfirmedNoEffect, ModeSilentNoEffect} {
		if !KnownFailureMode(mode) {
			t.Fatalf("KnownFailureMode(%q) = false", mode)
		}
	}
}

func assertViolation(t *testing.T, err error, want string) {
	t.Helper()
	if (want == "") != (err == nil) || (err != nil && !strings.Contains(err.Error(), want)) {
		t.Fatalf("error = %v, want %q", err, want)
	}
}
