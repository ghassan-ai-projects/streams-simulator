package refconsumer

import (
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"math"
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
			return model.Action{
				CommandID: cmdID, Effector: r.cfg.OnDetectionEffector, EntityID: entity,
				IssuedAt: model.FormatTime(t), OutcomeBelieved: model.BelievedFailed,
			}
		}
	}
	return model.Action{
		CommandID: cmdID, Effector: r.cfg.OnDetectionEffector, EntityID: entity,
		IssuedAt: model.FormatTime(t), OutcomeBelieved: model.BelievedUnknown,
	}
}

// argsFor constructs a minimal argument set for the detection effector:
// required string properties named *_id are bound to the entity id. This is
// consumer configuration made from the nameplate, not simulator knowledge.
func (r *Runner) argsFor(entity string) map[string]any {
	out := map[string]any{}
	var schema map[string]any
	for _, e := range r.np.Effectors {
		if e.Name == r.cfg.OnDetectionEffector {
			if e.Schema != nil {
				schema = e.Schema
			}
		}
	}
	if schema == nil {
		return out
	}
	if req, ok := schema["required"].([]any); ok {
		for _, rq := range req {
			name, _ := rq.(string)
			if name == "" {
				continue
			}
			if props, ok := schema["properties"].(map[string]any); ok {
				if p, ok := props[name].(map[string]any); ok {
					if p["type"] == "string" && len(name) >= 3 && name[len(name)-3:] == "_id" {
						out[name] = entity
					}
				}
			}
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
