package domain

import (
	"fmt"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

type RunScore struct {
	RunID      string       `json:"run_id"`
	Scorecards []*Scorecard `json:"scorecards"`
	Unexpected Unexpected   `json:"unexpected"`
}

type Unexpected struct {
	Detections  []model.Detection `json:"detections"`
	Actions     []model.Action    `json:"actions"`
	EffectCalls int               `json:"effect_calls"`
}

func ScoreRun(ev Evidence, gts []*model.GroundTruthRecord) (*RunScore, error) {
	if ev.Verdict == nil {
		return nil, fmt.Errorf("score: no verdict submitted for run %s", ev.RunID)
	}
	cards, err := scorecardsOf(ev, gts)
	if err != nil {
		return nil, err
	}
	return &RunScore{RunID: ev.RunID, Scorecards: cards, Unexpected: unexpectedOf(ev, disturbedEntities(gts))}, nil
}

func (r *RunScore) Document() any {
	if len(r.Scorecards) == 1 {
		return r.Scorecards[0]
	}
	return r
}

func scorecardsOf(ev Evidence, gts []*model.GroundTruthRecord) ([]*Scorecard, error) {
	cards := []*Scorecard{}
	for _, gt := range gts {
		card, err := scoreOne(ev, gt, len(gts) > 1)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}
	return cards, nil
}

func scoreOne(ev Evidence, gt *model.GroundTruthRecord, among bool) (*Scorecard, error) {
	card, err := Score(ev, gt)
	if err != nil || !among {
		return card, err
	}
	scoped := onlyEntity(ev, gt.EntityID)
	card.Judgment, card.Loop = judgment(scoped, gt), loop(scoped, gt)
	return card, nil
}

func onlyEntity(ev Evidence, entity string) Evidence {
	return selectEntities(ev, func(id string) bool { return id == entity })
}

func unexpectedOf(ev Evidence, disturbed map[string]bool) Unexpected {
	rest := selectEntities(ev, func(id string) bool { return !disturbed[id] })
	return Unexpected{Detections: rest.Verdict.Detections, Actions: rest.Verdict.Actions, EffectCalls: len(rest.Calls)}
}

func selectEntities(ev Evidence, keep func(entity string) bool) Evidence {
	verdict := *ev.Verdict
	verdict.Detections = selected(ev.Verdict.Detections, func(d model.Detection) bool { return keep(d.EntityID) })
	verdict.Actions = selected(ev.Verdict.Actions, func(a model.Action) bool { return keep(a.EntityID) })
	ev.Verdict = &verdict
	ev.Calls = selected(ev.Calls, func(c world.EffectorCall) bool { return keep(c.EntityID) })
	return ev
}

func selected[T any](items []T, keep func(T) bool) []T {
	return slices.AppendSeq([]T{}, func(yield func(T) bool) {
		for _, item := range items {
			if keep(item) && !yield(item) {
				return
			}
		}
	})
}

func disturbedEntities(gts []*model.GroundTruthRecord) map[string]bool {
	entities := map[string]bool{}
	for _, gt := range gts {
		entities[gt.EntityID] = true
	}
	return entities
}
