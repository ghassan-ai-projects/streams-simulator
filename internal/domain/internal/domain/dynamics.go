package domain

import (
	"sort"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func rejectDynamicsCycles(kind *model.Kind, at string, c *Compiled, src string) error {
	traversal := dynamicsTraversal{dependencies: dynamicsDependencies(kind), targets: dynamicsTargets(kind), state: make(map[string]uint8), source: src, base: at}
	for _, name := range c.StateNames(kind.Name) {
		if err := traversal.visit(name); err != nil {
			return err
		}
	}
	return nil
}

func (checks domainChecks) checkF1(f1 *model.F1Dyn) error {
	if err := checks.f1Inputs(f1); err != nil {
		return err
	}
	return checks.f1Parameters(f1)
}

func (checks domainChecks) checkDetector(d *model.Detector, who string) error {
	switch d.Form {
	case model.DetectorSingleChannel, model.DetectorPeerResidual:
		return checks.detectorChannel(d, who)
	case model.DetectorDivergence:
		return checks.detectorDivergence(d, who)
	case model.DetectorConservation:
		return checks.detectorConservation(d, who)
	default:
		return checks.bad("/form", "%s: unknown detector form %q", who, d.Form)
	}
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	// determinism-safe: collected keys are sorted below.
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
