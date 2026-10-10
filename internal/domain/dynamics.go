package domain

import (
	"sort"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func rejectDynamicsCycles(spec *model.DomainSpec, c *Compiled, src string) error {
	traversal := dynamicsTraversal{dependencies: dynamicsDependencies(spec), state: make(map[string]uint8), source: src}
	for _, name := range c.StateNames() {
		if err := traversal.visit(name); err != nil {
			return err
		}
	}
	return nil
}

func checkF1(f1 *model.F1Dyn, c *Compiled, src string) error {
	checks := domainChecks{compiled: c, source: src}
	if err := checks.f1Inputs(f1); err != nil {
		return err
	}
	return checks.f1Parameters(f1)
}

func checkDetector(d *model.Detector, c *Compiled, src, who string) error {
	checks := domainChecks{compiled: c, source: src}
	switch d.Form {
	case model.DetectorSingleChannel, model.DetectorPeerResidual:
		return checks.detectorChannel(d, who)
	case model.DetectorDivergence:
		return checks.detectorDivergence(d, who)
	case model.DetectorConservation:
		return checks.detectorConservation(d, who)
	default:
		return checks.bad("%s: unknown detector form %q", who, d.Form)
	}
}

func dynamicsFor(c *Compiled, state string) *model.Dynamics {
	for i := range c.Spec.Dynamics {
		if c.Spec.Dynamics[i].Target == state {
			return &c.Spec.Dynamics[i]
		}
	}
	return nil
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
