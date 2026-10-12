package app

import (
	"encoding/json"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// pondOperatorResponse is the operator's raw response to one start_aerator
// call on a pond world whose effectors all fail in the given mode.
func pondOperatorResponse(t *testing.T, mode world.FailureMode) []byte {
	t.Helper()
	spec, err := domain.Load(testsupport.Domain("aquaculture-pond"))
	if err != nil {
		t.Fatal(err)
	}
	w, err := world.New(spec, 5, "w-1", model.DefaultStartTimeNS, world.Options{ForceFailureMode: mode})
	if err != nil {
		t.Fatal(err)
	}
	view := NewOperatorView("w-1", "token", &Nameplate{Effectors: []EffectorInfo{{Name: "start_aerator"}}}, w, nil)
	pond := w.EntityIDs()[0]
	ack, err := view.Invoke("token", "start_aerator", pond, "cmd-1", map[string]any{"pond_id": pond, "level": 1.0}, w.Clock())
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(ack)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

// The silent_no_effect non-negotiable at the operator boundary: every mode
// that acknowledges the call reads exactly like one that worked, so the
// consumer cannot learn the effect failed from the response.
func TestEveryAcknowledgingModeReadsLikeSuccess(t *testing.T) {
	t.Parallel()
	ok := pondOperatorResponse(t, world.ModeOK)
	for _, mode := range []world.FailureMode{world.ModeSlow, world.ModePartial, world.ModeSilentNoEffect, world.ModeConfirmedNoEffect} {
		if got := pondOperatorResponse(t, mode); string(got) != string(ok) {
			t.Errorf("%s response %s, want the ok response %s", mode, got, ok)
		}
	}
}

// Whatever the mode, the response carries only the fields MCP_SURFACE §4.3
// allows.
func TestTheOperatorResponseCarriesOnlyItsContractFields(t *testing.T) {
	t.Parallel()
	allowed := map[string]bool{"accepted": true, "simulated": true, "world_id": true, "command_id": true, "effect_eta_ns": true, "reason": true}
	for _, mode := range []world.FailureMode{world.ModeOK, world.ModeSlow, world.ModeAckLost, world.ModeReject,
		world.ModePartial, world.ModeConfirmedNoEffect, world.ModeSilentNoEffect} {
		var fields map[string]any
		if err := json.Unmarshal(pondOperatorResponse(t, mode), &fields); err != nil {
			t.Fatal(err)
		}
		for name := range fields {
			if !allowed[name] {
				t.Errorf("%s: the response carries %q, which the operator may not see", mode, name)
			}
		}
	}
}

// A call that is not acknowledged reads as a refusal or a timeout, never as
// a failure mode, and a lost acknowledgement does not betray, through an
// effect time, that the effect applied.
func TestAnUnacknowledgedCallRevealsNoOutcome(t *testing.T) {
	t.Parallel()
	cases := map[world.FailureMode]string{
		world.ModeAckLost: `{"accepted":false,"simulated":true,"world_id":"w-1","command_id":"cmd-1","reason":"acknowledgement_timeout"}`,
		world.ModeReject:  `{"accepted":false,"simulated":true,"world_id":"w-1","command_id":"cmd-1","reason":"effector_refused"}`,
	}
	for mode, want := range cases {
		if got := pondOperatorResponse(t, mode); string(got) != want {
			t.Errorf("%s response %s, want %s", mode, got, want)
		}
	}
}
