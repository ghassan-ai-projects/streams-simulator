package app

import (
	"errors"
	"slices"
	"sync"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

type immediateGate struct {
	mu     sync.Mutex
	closed bool
	seq    int
}

func (j *job) closeImmediate() {
	j.immediate.mu.Lock()
	defer j.immediate.mu.Unlock()
	j.immediate.closed = true
}

func (j *job) admitImmediate() error {
	j.mu.Lock()
	defer j.mu.Unlock()
	if !j.takesCommands() || j.immediate.closed {
		return notRunning(j.status)
	}
	return nil
}

func (j *job) injectNow(entityID, fault string) (rules.JobImmediate, error) {
	if err := rules.CheckImmediateFault(entityID, fault); err != nil {
		return rules.JobImmediate{}, err
	}
	j.immediate.mu.Lock()
	defer j.immediate.mu.Unlock()
	if err := j.admitImmediate(); err != nil {
		return rules.JobImmediate{}, err
	}
	faultID, atNS, err := j.run.InjectFaultNow(entityID, fault, map[string]any{})
	if err != nil {
		return rules.JobImmediate{}, rules.ToolErrorf(rules.CodeDomainInvalid, "%v", err)
	}
	return j.recordImmediateFault(rules.NewImmediateFault(entityID, fault, faultID, (atNS-j.status.StartNS)/nsPerSecond)), nil
}

func (j *job) recordImmediateFault(fault rules.JobFaultView) rules.JobImmediate {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.status.Faults = append(j.status.Faults, fault)
	return rules.JobImmediate{JobStatus: j.copyStatus(), FaultID: fault.FaultID}
}

func (j *job) invokeNow(entityID, effector string, args map[string]any) (rules.JobImmediate, error) {
	j.immediate.mu.Lock()
	defer j.immediate.mu.Unlock()
	if err := j.admitImmediate(); err != nil {
		return rules.JobImmediate{}, err
	}
	j.immediate.seq++
	ack, atNS, err := j.run.InvokeAcknowledgedNow(effector, entityID, rules.ImmediateCommandID(j.immediate.seq), args)
	if err != nil {
		return rules.JobImmediate{}, invocationError(err)
	}
	return j.recordImmediate(rules.NewImmediateCallOutcome((atNS-j.status.StartNS)/nsPerSecond, entityID, effector, ack)), nil
}

func (j *job) recordImmediate(outcome rules.JobCallOutcome) rules.JobImmediate {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.status.Calls = append(j.status.Calls, outcome)
	return rules.JobImmediate{JobStatus: j.copyStatus(), Call: &outcome}
}

func invocationError(err error) error {
	if errors.Is(err, world.ErrInvalidInvocation) {
		return rules.ToolErrorf(rules.CodeInvalidArgs, "%v", err)
	}
	return rules.ToolErrorf(rules.CodeDomainInvalid, "%v", err)
}

func declaresEffector(nameplate *rules.Nameplate, effector string) bool {
	return slices.ContainsFunc(nameplate.Effectors, func(info rules.EffectorInfo) bool { return info.Name == effector })
}
