package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestEffectorValidatesArgumentsBeforeIdempotentReplay(t *testing.T) {
	spec := testSpec(t, func(s *model.DomainSpec) {
		s.Effectors = []model.Effector{{Name: "act", ArgsSchema: map[string]any{"type": "object", "required": []any{"level"}, "properties": map[string]any{"level": map[string]any{"type": "number"}}}, Effect: model.Effect{StateDeltas: []model.StateDelta{{State: "x", Delta: 3}}}}}
	})
	w := newTestWorld(t, spec, 17, model.DefaultStartTimeNS)
	w.SetFailureMode(ModeOK)
	if _, err := w.InvokeEffector("act", "e-1", "command", map[string]any{"level": 1.0}, w.Clock()); err != nil {
		t.Fatal(err)
	}
	if _, err := w.InvokeEffector("act", "e-1", "command", map[string]any{"level": "invalid"}, w.Clock()); err == nil {
		t.Fatal("invalid retry must fail argument validation before cached replay")
	}
	if len(w.EffectorCalls()) != 1 {
		t.Fatal("invalid retry recorded a second call")
	}
}
