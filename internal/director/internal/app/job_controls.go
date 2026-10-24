package app

import (
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
)

func (d *Director) jobOf(worldID string) (*job, error) {
	j := d.jobs.find(worldID)
	if j == nil {
		return nil, rules.ToolErrorf(rules.CodeWorldNotFound, "no job for world %q", worldID)
	}
	return j, nil
}

func (d *Director) PauseJob(worldID string) (rules.JobStatus, error) {
	j, err := d.jobOf(worldID)
	if err != nil {
		return rules.JobStatus{}, err
	}
	return j.pause()
}

func (d *Director) ResumeJob(worldID string) (rules.JobStatus, error) {
	j, err := d.jobOf(worldID)
	if err != nil {
		return rules.JobStatus{}, err
	}
	return j.resume()
}

func (d *Director) StepJob(worldID string) (rules.JobStatus, error) {
	j, err := d.jobOf(worldID)
	if err != nil {
		return rules.JobStatus{}, err
	}
	return j.step()
}

func (d *Director) SetJobRate(worldID string, rate float64) (rules.JobStatus, error) {
	j, err := d.jobOf(worldID)
	if err != nil {
		return rules.JobStatus{}, err
	}
	return j.setRate(rate)
}

func (d *Director) JobInjectFault(worldID, entityID, fault string) (rules.JobImmediate, error) {
	j, err := d.jobOf(worldID)
	if err != nil {
		return rules.JobImmediate{}, err
	}
	if err := rules.CheckInjectionAllowed(worldID, j.status.Explore); err != nil {
		return rules.JobImmediate{}, err
	}
	return j.injectNow(entityID, fault)
}

func (d *Director) JobInvokeEffector(worldID, entityID, effector string, args map[string]any) (rules.JobImmediate, error) {
	j, err := d.jobOf(worldID)
	if err != nil {
		return rules.JobImmediate{}, err
	}
	if err := d.admitJobCall(worldID, entityID, effector); err != nil {
		return rules.JobImmediate{}, err
	}
	return j.invokeNow(entityID, effector, args)
}

func (d *Director) admitJobCall(worldID, entityID, effector string) error {
	if err := rules.CheckImmediateCall(entityID, effector); err != nil {
		return err
	}
	if w := d.World(worldID); w != nil && !declaresEffector(w.Nameplate, effector) {
		return rules.ToolErrorf(rules.CodeUnknownEffector, "effector %q is not declared by this world's domain", effector)
	}
	return nil
}
