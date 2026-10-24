package app

import (
	"context"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

const nsPerSecond = int64(1_000_000_000)

func (j *job) hooks() run.ScenarioHooks {
	return run.ScenarioHooks{Before: j.before, After: j.after}
}

func (j *job) before(_ context.Context, _ int, command model.ScenarioCommand) (bool, error) {
	if j.mustStop() {
		return true, nil
	}
	if command.Op != model.OpClockAdvance || j.pacer == nil {
		return false, nil
	}
	return j.pace(command)
}

func (j *job) pace(command model.ScenarioCommand) (bool, error) {
	err := j.pacer.WaitUntil(j.waitCtx, timeArgument(command, "to_ns")-j.status.StartNS)
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
