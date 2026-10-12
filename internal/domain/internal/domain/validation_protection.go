package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (checks domainChecks) f1Trip(f1 *model.F1Dyn) error {
	if err := checks.condition("/when", f1.When); err != nil {
		return err
	}
	if f1.Latch && !checks.hasState(f1.ResetOn) {
		return checks.bad("/reset_on", "a latched trip needs reset_on naming a declared state, not %q", f1.ResetOn)
	}
	return nil
}

func (checks domainChecks) f1Sequencer(f1 *model.F1Dyn) error {
	if err := checks.condition("/stage_up", f1.StageUp); err != nil {
		return err
	}
	return checks.condition("/stage_down", f1.StageDown)
}

func (checks domainChecks) condition(member string, c *model.Condition) error {
	if !checks.hasState(c.State) {
		return checks.bad(member+"/state", "condition on %q, which is not a declared state", c.State)
	}
	if c.ThresholdState != "" && !checks.hasState(c.ThresholdState) {
		return checks.bad(member+"/threshold_state", "condition threshold %q is not a declared state", c.ThresholdState)
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
	for i := range e.Requires {
		if err := checks.condition(fmt.Sprintf("/requires/%d", i), &e.Requires[i]); err != nil {
			return err
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

func (checks domainChecks) faultDuration(f *model.Fault) error {
	if f.Onset.DurationS != 0 && f.Onset.Shape != model.ShapePulse {
		return checks.bad("/onset/duration_s", "fault %q declares duration_s, which only a pulse uses", f.ID)
	}
	return nil
}
