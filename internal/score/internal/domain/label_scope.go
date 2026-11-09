package domain

import (
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// With other labels on its entity, a detection or call that names or expects only them is theirs; one naming none is shared.
type labelScope struct {
	gt     *model.GroundTruthRecord
	shared []*model.GroundTruthRecord
	acted  map[string]bool
}

func scopeOf(gt *model.GroundTruthRecord, gts []*model.GroundTruthRecord) labelScope {
	shared := selected(gts, func(o *model.GroundTruthRecord) bool { return o != gt && o.EntityID == gt.EntityID })
	return labelScope{gt: gt, shared: shared, acted: actedEntities(gt)}
}

func (s labelScope) judges(d model.Detection) bool {
	if d.EntityID != s.gt.EntityID {
		return false
	}
	return d.Label == s.gt.Label || !slices.ContainsFunc(s.shared, func(o *model.GroundTruthRecord) bool { return o.Label == d.Label })
}

func (s labelScope) handles(entity, effector string) bool {
	if !s.acted[entity] {
		return false
	}
	if entity != s.gt.EntityID || expects(s.gt, entity, effector) {
		return true
	}
	return !slices.ContainsFunc(s.shared, func(o *model.GroundTruthRecord) bool { return expects(o, entity, effector) })
}

func expects(gt *model.GroundTruthRecord, entity, effector string) bool {
	if entity == gt.EntityID && gt.ExpectedEffector != "" && effector == gt.ExpectedEffector {
		return true
	}
	return slices.Contains(gt.ExpectedActions, model.ExpectedAction{EntityID: entity, Effector: effector})
}

func (s labelScope) evidence(ev Evidence) Evidence {
	verdict := *ev.Verdict
	verdict.Detections = selected(ev.Verdict.Detections, s.judges)
	verdict.Actions = selected(ev.Verdict.Actions, func(a model.Action) bool { return s.handles(a.EntityID, a.Effector) })
	ev.Verdict = &verdict
	ev.Calls = selected(ev.Calls, func(c world.EffectorCall) bool { return s.handles(c.EntityID, c.Effector) })
	return ev
}
