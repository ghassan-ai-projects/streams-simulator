package domain

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const (
	rk4StableRatio   = 2.78
	eulerStableRatio = 2.0
)

func (checks domainChecks) checkStability() error {
	kind := checks.subject
	return checks.each("dynamics", len(kind.Dynamics), func(at domainChecks, i int) error {
		dyn := &kind.Dynamics[i]
		if dyn.F1 == nil {
			return nil
		}
		if dyn.F1.Form == model.FormStore {
			return at.storeStiffness(dyn)
		}
		return at.lagStability(dyn)
	})
}

func (checks domainChecks) lagStability(dyn *model.Dynamics) error {
	limit, taus := stabilityLimit(dyn.F1)
	for _, tau := range taus {
		if tau > 0 && stepSeconds(dyn)/tau >= limit {
			return checks.bad("/dt_ms", "state %q steps every %v s against a time constant of %v s; the %s form is stable only below dt/τ = %v", dyn.Target, stepSeconds(dyn), tau, dyn.F1.Form, limit)
		}
	}
	return nil
}

func stabilityLimit(f1 *model.F1Dyn) (float64, []float64) {
	switch f1.Form {
	case "first_order_lag", "rc_network":
		return rk4StableRatio, []float64{f1.TimeConstantS, f1.TimeConstant2S}
	case "dead_time":
		return eulerStableRatio, []float64{f1.TimeConstantS}
	case model.FormRegime:
		return rk4StableRatio, []float64{regimeTau(f1.Below), regimeTau(f1.Above)}
	}
	return 0, nil
}

func regimeTau(r *model.Regime) float64 {
	if r == nil {
		return 0
	}
	return r.TimeConstantS
}

func stepSeconds(dyn *model.Dynamics) float64 {
	return float64(dyn.StepMs()) / 1000
}

func (checks domainChecks) storeStiffness(dyn *model.Dynamics) error {
	if dyn.F1.Capacity <= 0 {
		return nil
	}
	if rate := checks.storeExchangeRate(dyn.Target) / dyn.F1.Capacity; rate*stepSeconds(dyn) >= rk4StableRatio {
		return checks.bad("/dt_ms", "store %q exchanges at %v per second; with a %v s step the store is stable only below rate·dt = %v", dyn.Target, rate, stepSeconds(dyn), rk4StableRatio)
	}
	return nil
}

func (checks domainChecks) storeExchangeRate(store string) float64 {
	var rate float64
	for _, flow := range checks.subject.Flows {
		if flow.From != store && flow.To != store {
			continue
		}
		reach := 1.0
		if flow.To != "" {
			reach = 2
		}
		rate += reach * flow.Conductance * checks.scaleBound(flow.Scale)
	}
	return rate
}

func (checks domainChecks) scaleBound(scale []string) float64 {
	bound := 1.0
	for _, name := range scale {
		if decl := checks.kind.StateDecl(name); decl != nil && decl.Max != nil {
			bound *= max(*decl.Max, 0)
		}
	}
	return bound
}
