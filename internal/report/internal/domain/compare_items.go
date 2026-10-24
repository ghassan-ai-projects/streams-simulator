package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

type FaultItem struct {
	EntityID  string  `json:"entity_id"`
	Fault     string  `json:"fault"`
	OnsetNS   int64   `json:"onset_ns"`
	ClearedNS *int64  `json:"cleared_ns"`
	ClearedBy *string `json:"cleared_by"`
}

type PerturbationItem struct {
	Action       string         `json:"action"`
	Perturbation string         `json:"perturbation,omitempty"`
	PerturbID    string         `json:"perturb_id,omitempty"`
	AtNS         int64          `json:"at_ns"`
	FromNS       *int64         `json:"from_ns,omitempty"`
	UntilNS      *int64         `json:"until_ns,omitempty"`
	Params       map[string]any `json:"params,omitempty"`
}

type EnvironmentItem struct {
	AtNS   int64          `json:"at_ns"`
	Target string         `json:"target"`
	Fault  string         `json:"fault"`
	Params map[string]any `json:"params,omitempty"`
}

type CallItem struct {
	AtNS     int64          `json:"at_ns"`
	EntityID string         `json:"entity_id,omitempty"`
	Effector string         `json:"effector"`
	Args     map[string]any `json:"args,omitempty"`
	Outcome  *Outcome       `json:"outcome,omitempty"`
}

func (f FaultItem) key() string { return f.EntityID + "|" + f.Fault }

func (p PerturbationItem) key() string { return p.Action + "|" + first(p.Perturbation, p.PerturbID) }

func (e EnvironmentItem) key() string { return e.Target + "|" + e.Fault }

func (c CallItem) key() string { return c.EntityID + "|" + c.Effector }

func planOf(a, b *Report) PlanDiff {
	return PlanDiff{
		Faults:        diffItems(faultItems(a), faultItems(b)),
		Perturbations: diffItems(perturbationItems(a), perturbationItems(b)),
		Environment:   diffItems(environmentItems(a), environmentItems(b)),
		EffectorCalls: diffItems(callItems(a), callItems(b)),
	}
}

func truthOfReport(report *Report) *Truth {
	if report.Truth == nil {
		return &Truth{}
	}
	return report.Truth
}

func faultItems(report *Report) []FaultItem {
	return mapItems(truthOfReport(report).Faults, func(f model.FaultInterval) FaultItem {
		return FaultItem{EntityID: f.EntityID, Fault: f.Fault, OnsetNS: f.OnsetNS, ClearedNS: f.ClearedNS, ClearedBy: f.ClearedBy}
	})
}

func perturbationItems(report *Report) []PerturbationItem {
	return mapItems(truthOfReport(report).Perturbations, func(p Perturbation) PerturbationItem {
		return PerturbationItem{Action: p.Action, Perturbation: p.Perturbation, PerturbID: p.PerturbID, AtNS: p.AtNS, FromNS: p.FromNS, UntilNS: p.UntilNS, Params: p.Params}
	})
}

func environmentItems(report *Report) []EnvironmentItem {
	return mapItems(truthOfReport(report).Environment, func(e Environment) EnvironmentItem {
		return EnvironmentItem{AtNS: e.AtNS, Target: e.Target, Fault: e.Fault, Params: e.Params}
	})
}

func callItems(report *Report) []CallItem {
	return mapItems(truthOfReport(report).EffectorCalls, func(c EffectorCall) CallItem {
		return CallItem{AtNS: c.AtNS, EntityID: c.EntityID, Effector: c.Effector, Args: c.Args, Outcome: c.Outcome}
	})
}

func mapItems[S, T any](source []S, convert func(S) T) []T {
	items := make([]T, len(source))
	for i, value := range source {
		items[i] = convert(value)
	}
	return items
}
