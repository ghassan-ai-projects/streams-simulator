package app

import (
	"context"
	"errors"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

const nsPerSecond = int64(1_000_000_000)

func (j *job) hooks() run.ScenarioHooks {
	return run.ScenarioHooks{Before: j.before, After: j.after}
}

func (j *job) before(_ context.Context, _ int, command model.ScenarioCommand) (bool, error) {
	for {
		j.awaitRelease()
		if j.mustStop() {
			return true, nil
		}
		if command.Op != model.OpClockAdvance {
			return false, nil
		}
		if err := j.pace(command); !errors.Is(err, errPaused) {
			return j.settlePace(err)
		}
	}
}

func (j *job) awaitRelease() {
	for {
		wake, hold := j.holdState()
		if !hold {
			return
		}
		select {
		case <-wake:
		case <-j.ctx.Done():
			return
		}
	}
}

func (j *job) holdState() (<-chan struct{}, bool) {
	j.mu.Lock()
	defer j.mu.Unlock()
	hold := j.gate.paused && !j.gate.stepping && !j.stopRequested
	if hold && j.gate.parked != nil {
		close(j.gate.parked)
		j.gate.parked = nil
	}
	return j.gate.wake, hold
}

func (j *job) pace(command model.ScenarioCommand) error {
	ctx, skip := j.paceContext()
	if skip {
		return nil
	}
	err := j.pacer.WaitUntil(ctx, timeArgument(command, "to_ns")-j.status.StartNS)
	if err != nil && errors.Is(context.Cause(ctx), errPaused) {
		return errPaused
	}
	return err
}

func (j *job) paceContext() (context.Context, bool) {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.gate.paceCtx, j.gate.stepping
}

func (j *job) settlePace(err error) (bool, error) {
	if err == nil || j.stopWasRequested() {
		return j.mustStop(), nil
	}
	return false, err
}

func (j *job) mustStop() bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.stopRequested && j.advances > 0
}

func (j *job) stopWasRequested() bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.stopRequested
}

func (j *job) after(_ context.Context, _ int, command model.ScenarioCommand, result run.StepResult) {
	j.mu.Lock()
	defer j.mu.Unlock()
	switch command.Op {
	case model.OpClockAdvance:
		j.advances++
		j.gate.stepping = false
		j.status.ReachedNS = timeArgument(command, "to_ns")
	case model.OpEffectorInvoke:
		j.status.Calls = append(j.status.Calls, j.callOutcome(command, result.Acknowledgement))
	}
}

func (j *job) callOutcome(command model.ScenarioCommand, ack *model.EffectorAcknowledgement) rules.JobCallOutcome {
	atS := (timeArgument(command, "at_ns") - j.status.StartNS) / nsPerSecond
	entityID, _ := command.Args["entity_id"].(string)
	effector, _ := command.Args["effector"].(string)
	return rules.NewJobCallOutcome(atS, entityID, effector, ack)
}

func timeArgument(command model.ScenarioCommand, key string) int64 {
	value, _ := command.Args[key].(int64)
	return value
}
