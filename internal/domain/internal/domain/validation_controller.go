package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

func (checks domainChecks) f1Controller(f1 *model.F1Dyn) error {
	c := f1.PID
	if err := checks.controllerStates(c); err != nil {
		return err
	}
	if c.Kp == 0 && c.Ki == 0 {
		return checks.bad("/pid/kp", "a pid controller needs a proportional or an integral gain")
	}
	if f1.Clamp == nil || f1.Clamp.Min == nil || f1.Clamp.Max == nil || *f1.Clamp.Min >= *f1.Clamp.Max {
		return checks.bad("/clamp", "a pid output needs a clamp with min below max")
	}
	return nil
}

func (checks domainChecks) controllerStates(c *model.Controller) error {
	for _, read := range [][2]string{{"pv", c.PV}, {"sp", c.SP}} {
		if !checks.hasState(read[1]) {
			return checks.bad("/pid/"+read[0], "pid %s %q is not a declared state", read[0], read[1])
		}
	}
	if (c.Manual == "") != (c.ManualOutput == "") {
		return checks.bad("/pid/manual", "a pid manual mode names both manual and manual_output")
	}
	if c.Manual != "" && (!checks.hasState(c.Manual) || !checks.hasState(c.ManualOutput)) {
		return checks.bad("/pid/manual", "pid manual %q and manual_output %q must be declared states", c.Manual, c.ManualOutput)
	}
	return nil
}

func (checks domainChecks) f1Alarm(f1 *model.F1Dyn) error {
	a := f1.Alarm
	if (a.Limit == nil) == (a.LimitState == "") {
		return checks.bad("/alarm/limit", "an alarm names exactly one of limit and limit_state")
	}
	for _, read := range [][2]string{{"watch", a.Watch}, {"limit_state", a.LimitState}, {"acknowledge", a.Acknowledge}, {"shelve", a.Shelve}} {
		if read[1] != "" && !checks.hasState(read[1]) {
			return checks.bad("/alarm/"+read[0], "alarm %s %q is not a declared state", read[0], read[1])
		}
	}
	return nil
}
