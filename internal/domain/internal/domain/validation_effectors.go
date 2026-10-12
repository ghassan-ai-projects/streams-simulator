package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (checks domainChecks) checkEffector(e *model.Effector) error {
	for _, check := range []func(*model.Effector) error{checks.effectorStates, checks.effectorInterlock, checks.effectorAutonomousAction, checks.effectorConfirmation, checks.effectorDeadTime, checks.effectorRequires, checks.effectorResponder, checks.effectorLease} {
		if err := check(e); err != nil {
			return err
		}
	}
	return nil
}

func (checks domainChecks) effectorStates(e *model.Effector) error {
	for i, d := range e.Effect.StateDeltas {
		if !checks.hasState(d.State) {
			return checks.bad(fmt.Sprintf("/effect/state_deltas/%d/state", i), "effector %q effect targets undeclared state %q", e.Name, d.State)
		}
	}
	return nil
}

func (checks domainChecks) effectorInterlock(e *model.Effector) error {
	if e.Interlock != nil && !checks.hasState(e.Interlock.State) {
		return checks.bad("/interlock/state", "effector %q interlock references undeclared state %q", e.Name, e.Interlock.State)
	}
	return nil
}

func (checks domainChecks) effectorAutonomousAction(e *model.Effector) error {
	if e.Interlock != nil && e.Interlock.AutonomousAction != nil {
		if !checks.hasState(e.Interlock.AutonomousAction.State) {
			return checks.bad("/interlock/autonomous_action/state", "effector %q autonomous action targets undeclared state %q", e.Name, e.Interlock.AutonomousAction.State)
		}
	}
	return nil
}

func (checks domainChecks) effectorConfirmation(e *model.Effector) error {
	for i, ch := range e.ConfirmationChannels {
		if !checks.hasChannel(ch) {
			return checks.bad(fmt.Sprintf("/confirmation_channels/%d", i), "effector %q confirmation channel %q is not declared", e.Name, ch)
		}
	}
	return nil
}

func (checks domainChecks) effectorDeadTime(e *model.Effector) error {
	// Model dead time once; targeting a dead_time state would apply the delay twice.
	if e.Effect.DeadTimeS > 0 {
		for _, d := range e.Effect.StateDeltas {
			if dyn := checks.kind.DynamicsOf(d.State); dyn != nil && dyn.F1 != nil && dyn.F1.Form == "dead_time" {
				return checks.bad("/effect/dead_time_s", "effector %q has dead_time_s and drives dead_time state %q: delay would apply twice", e.Name, d.State)
			}
		}
	}
	return nil
}
