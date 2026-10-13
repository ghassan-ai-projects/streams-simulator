package domain

import (
	"fmt"
	"sort"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func rejectDynamicsCycles(kind *model.Kind, at string, src string) error {
	targets := dynamicsTargets(kind)
	traversal := cycleTraversal{dependencies: dynamicsDependencies(kind), state: make(map[string]uint8), cycle: func(name string, path []string) error {
		return invalidDomain(src, at+itemPointer("dynamics", targets[name])+"/target", fmt.Sprintf("dynamics dependency cycle at state %q (path %v)", name, path))
	}}
	for _, name := range sortedStateNames(kind) {
		if err := traversal.visit(name); err != nil {
			return err
		}
	}
	return nil
}

func sortedStateNames(kind *model.Kind) []string {
	names := make([]string, len(kind.State))
	for i := range kind.State {
		names[i] = kind.State[i].Name
	}
	sort.Strings(names)
	return names
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
