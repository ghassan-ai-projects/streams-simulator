package refconsumer

import (
	"fmt"
	"math"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (r *Runner) suspicious(s *series, v float64) bool {
	if len(s.window) < 5 {
		return false
	}
	mean, std := meanStd(s.window)
	if std < 1e-12 {
		return false
	}
	return math.Abs(v-mean) > r.cfg.Threshold*std
}

func (r *Runner) detect(entity, channel string, t int64, s *series) model.Detection {
	// Evidence refs: the scoring side validates citations against delivered
	// records; the consumer cites the channel it detected on.
	return model.Detection{
		EntityID:     entity,
		DetectedAt:   model.FormatTime(t),
		Confidence:   0.8,
		Narrative:    fmt.Sprintf("%s deviated beyond %.1f sigma of its %d-reading window", channel, r.cfg.Threshold, len(s.window)),
		EvidenceRefs: []string{fmt.Sprintf("seq:%d", s.lastSeq)},
	}
}

func (r *Runner) issue(entity string, t int64) model.Action {
	r.commandSeq++
	cmdID := fmt.Sprintf("rc-%05d", r.commandSeq)
	if r.invoker != nil {
		if _, err := r.invoker.InvokeEffector(r.cfg.OnDetectionEffector, entity, cmdID, r.argsFor(entity), t); err != nil {
			return r.issuedAction(cmdID, entity, t, model.BelievedFailed)
		}
	}
	return r.issuedAction(cmdID, entity, t, model.BelievedUnknown)
}

// argsFor constructs a minimal argument set for the detection effector:
// required string properties named *_id are bound to the entity id. This is
// consumer configuration made from the nameplate, not simulator knowledge.
func (r *Runner) argsFor(entity string) map[string]any {
	out := map[string]any{}
	schema := r.detectionEffectorSchema()
	required, _ := schema["required"].([]any)
	for _, item := range required {
		if name, ok := entityArgument(schema, item); ok {
			out[name] = entity
		}
	}
	return out
}

func (r *Runner) configDigest() string {
	d, _ := canonical.Digest(map[string]any{
		"threshold": r.cfg.Threshold, "window": r.cfg.Window,
		"min_consecutive": r.cfg.MinConsecutive, "effector": r.cfg.OnDetectionEffector,
		"absence_factor": r.cfg.AbsenceFactor,
	})
	return d
}

func (r *Runner) issuedAction(command, entity string, at int64, outcome string) model.Action {
	return model.Action{CommandID: command, Effector: r.cfg.OnDetectionEffector,
		EntityID: entity, IssuedAt: model.FormatTime(at), OutcomeBelieved: outcome}
}

func (r *Runner) detectionEffectorSchema() map[string]any {
	var schema map[string]any
	for _, effector := range r.np.Effectors {
		if effector.Name == r.cfg.OnDetectionEffector && effector.Schema != nil {
			schema = effector.Schema
		}
	}
	return schema
}

func entityArgument(schema map[string]any, item any) (string, bool) {
	name, _ := item.(string)
	if name == "" {
		return "", false
	}
	properties, _ := schema["properties"].(map[string]any)
	property, _ := properties[name].(map[string]any)
	return name, property["type"] == "string" && len(name) >= 3 && name[len(name)-3:] == "_id"
}
