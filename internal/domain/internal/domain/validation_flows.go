package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (checks domainChecks) checkFlows() error {
	kind := checks.subject
	seen := make(map[string]bool, len(kind.Flows))
	return checks.each("flows", len(kind.Flows), func(at domainChecks, i int) error {
		flow := &kind.Flows[i]
		if seen[flow.Name] {
			return at.bad("/name", "flow %q is declared twice", flow.Name)
		}
		seen[flow.Name] = true
		return at.checkFlow(flow)
	})
}

func (checks domainChecks) checkFlow(flow *model.Flow) error {
	if !checks.subject.IsStore(flow.From) {
		return checks.bad("/from", "flow %q: from %q is not a store state", flow.Name, flow.From)
	}
	if flow.To != "" && (flow.To == flow.From || !checks.subject.IsStore(flow.To)) {
		return checks.bad("/to", "flow %q: to %q is not another store state", flow.Name, flow.To)
	}
	return checks.flowScale(flow)
}

func (checks domainChecks) flowScale(flow *model.Flow) error {
	for i, name := range flow.Scale {
		if !checks.hasState(name) {
			return checks.bad(fmt.Sprintf("/scale/%d", i), "flow %q: scale %q is not a declared state", flow.Name, name)
		}
	}
	return nil
}

func (checks domainChecks) checkStoreSteps() error {
	step := firstStoreStep(checks.kind)
	return checks.each("dynamics", len(checks.subject.Dynamics), func(at domainChecks, i int) error {
		dyn := &checks.subject.Dynamics[i]
		if dyn.F1 != nil && dyn.F1.Form == model.FormStore && dyn.StepMs() != step {
			return at.bad("/dt_ms", "store %q steps every %d ms; every store of a kind shares one step (%d ms)", dyn.Target, dyn.StepMs(), step)
		}
		return nil
	})
}

func firstStoreStep(kind *model.Kind) int64 {
	for i := range kind.Dynamics {
		if kind.IsStore(kind.Dynamics[i].Target) {
			return kind.Dynamics[i].StepMs()
		}
	}
	return 0
}
