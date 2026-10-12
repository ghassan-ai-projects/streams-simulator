package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (checks domainChecks) f1Trip(f1 *model.F1Dyn) error {
	switch {
	case !checks.hasState(f1.When.State):
		return checks.bad("/when/state", "trip condition state %q is not a declared state", f1.When.State)
	case f1.Latch && !checks.hasState(f1.ResetOn):
		return checks.bad("/reset_on", "a latched trip needs reset_on naming a declared state, not %q", f1.ResetOn)
	}
	return nil
}

func (checks domainChecks) f1AntiRecycle(f1 *model.F1Dyn) error {
	if !checks.hasState(f1.Watch) {
		return checks.bad("/watch", "anti_recycle watches %q, which is not a declared state", f1.Watch)
	}
	return nil
}

func (checks domainChecks) checkBoundaries() error {
	kind := checks.subject
	seen := map[string]bool{}
	return checks.each("boundaries", len(kind.Boundaries), func(at domainChecks, i int) error {
		boundary := &kind.Boundaries[i]
		if seen[boundary.Name] {
			return at.bad("/name", "boundary %q is declared twice", boundary.Name)
		}
		seen[boundary.Name] = true
		if dyn := kind.DynamicsOf(boundary.State); dyn == nil || dyn.F1 == nil || dyn.F1.Form != model.FormTrip {
			return at.bad("/state", "boundary %q watches %q, which is not a trip state", boundary.Name, boundary.State)
		}
		return nil
	})
}

func (checks domainChecks) effectorRequires(e *model.Effector) error {
	for i, condition := range e.Requires {
		if !checks.hasState(condition.State) {
			return checks.bad(fmt.Sprintf("/requires/%d/state", i), "effector %q requires undeclared state %q", e.Name, condition.State)
		}
	}
	return nil
}

func (checks domainChecks) dynamicTripValue(d *model.Dynamics) error {
	if d.F1 == nil || d.F1.Form != model.FormTrip {
		return nil
	}
	if state := checks.kind.StateDecl(d.Target); state != nil && state.Initial == d.F1.TrippedValue {
		return checks.bad("/f1/tripped_value", "trip %q has a tripped_value equal to its initial value, so tripped cannot be told apart", d.Target)
	}
	return nil
}
