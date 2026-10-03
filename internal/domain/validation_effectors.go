package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

func (checks domainChecks) checkEffector(e *model.Effector) error {
	for _, check := range []func(*model.Effector) error{checks.effectorStates, checks.effectorInterlock, checks.effectorAutonomousAction, checks.effectorConfirmation, checks.effectorDeadTime} {
		if err := check(e); err != nil {
			return err
		}
	}
	return nil
}

func (checks domainChecks) effectorStates(e *model.Effector) error {
	for _, d := range e.Effect.StateDeltas {
		if !checks.compiled.HasState(d.State) {
			return checks.bad("effector %q effect targets undeclared state %q", e.Name, d.State)
		}
	}
	return nil
}

func (checks domainChecks) effectorInterlock(e *model.Effector) error {
	if e.Interlock != nil && !checks.compiled.HasState(e.Interlock.State) {
		return checks.bad("effector %q interlock references undeclared state %q", e.Name, e.Interlock.State)
	}
	return nil
}

func (checks domainChecks) effectorAutonomousAction(e *model.Effector) error {
	if e.Interlock != nil && e.Interlock.AutonomousAction != nil {
		if !checks.compiled.HasState(e.Interlock.AutonomousAction.State) {
			return checks.bad("effector %q autonomous action targets undeclared state %q", e.Name, e.Interlock.AutonomousAction.State)
		}
	}
	return nil
}

func (checks domainChecks) effectorConfirmation(e *model.Effector) error {
	for _, ch := range e.ConfirmationChannels {
		if !checks.compiled.HasChannel(ch) {
			return checks.bad("effector %q confirmation channel %q is not declared", e.Name, ch)
		}
	}
	return nil
}

func (checks domainChecks) effectorDeadTime(e *model.Effector) error {
	// Model dead time once; targeting a dead_time state would apply the delay twice.
	if e.Effect.DeadTimeS > 0 {
		for _, d := range e.Effect.StateDeltas {
			if dyn := dynamicsFor(checks.compiled, d.State); dyn != nil && dyn.F1 != nil && dyn.F1.Form == "dead_time" {
				return checks.bad("effector %q has dead_time_s and drives dead_time state %q: delay would apply twice", e.Name, d.State)
			}
		}
	}
	return nil
}
