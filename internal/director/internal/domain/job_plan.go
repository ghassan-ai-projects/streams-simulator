package domain

import (
	"fmt"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const (
	nsPerSecond    = int64(1_000_000_000)
	pacedChunks    = int64(30)
	unpacedChunks  = int64(20)
	jobCommandBase = "job-"
)

func PlanScenario(req JobRequest, adapterID string, startNS int64) model.Scenario {
	return model.Scenario{
		SchemaVersion: model.ScenarioSchemaVersion,
		ID:            fmt.Sprintf("job-seed-%d-%ds", req.Seed, req.DurationS),
		Description:   fmt.Sprintf("Planned run of %s: %d s, %d faults, %d effector calls", req.Domain, req.DurationS, len(req.Faults), len(req.Calls)),
		Domain:        req.Domain,
		Adapter:       adapterID,
		Seed:          uint64(max(0, req.Seed)),
		World:         model.ScenarioWorld{StartTimeNS: startNS},
		Commands:      planCommands(req, startNS),
	}
}

func planCommands(req JobRequest, startNS int64) []model.ScenarioCommand {
	commands := actionsAt(req, startNS, 0)
	for _, at := range boundaries(req) {
		advance := model.ScenarioCommand{Op: model.OpClockAdvance, Args: map[string]any{"to_ns": startNS + at*nsPerSecond}}
		commands = append(append(commands, advance), actionsAt(req, startNS, at)...)
	}
	return commands
}

func boundaries(req JobRequest) []int64 {
	marks := append(chunkMarks(req), actionMarks(req)...)
	marks = slices.DeleteFunc(marks, func(at int64) bool { return at <= 0 })
	slices.Sort(marks)
	return slices.Compact(marks)
}

func chunkMarks(req JobRequest) []int64 {
	step := max(1, ceilDiv(req.DurationS, chunkCount(req)))
	marks := []int64{req.DurationS}
	for at := step; at < req.DurationS; at += step {
		marks = append(marks, at)
	}
	return marks
}

func actionMarks(req JobRequest) []int64 {
	var marks []int64
	for _, fault := range req.Faults {
		marks = append(marks, fault.AtS)
	}
	for _, call := range req.Calls {
		marks = append(marks, call.AtS)
	}
	return marks
}

func chunkCount(req JobRequest) int64 {
	if req.Rate > 0 {
		return pacedChunks
	}
	return unpacedChunks
}

func ceilDiv(n, d int64) int64 {
	return (n + d - 1) / d
}

func actionsAt(req JobRequest, startNS, atS int64) []model.ScenarioCommand {
	var commands []model.ScenarioCommand
	for _, fault := range req.Faults {
		if fault.AtS == atS {
			commands = append(commands, injectCommand(fault, startNS))
		}
	}
	for index, call := range req.Calls {
		if call.AtS == atS {
			commands = append(commands, invokeCommand(call, index, startNS))
		}
	}
	return commands
}

func injectCommand(fault JobFault, startNS int64) model.ScenarioCommand {
	return model.ScenarioCommand{Op: model.OpFaultInject, Args: map[string]any{
		"entity_id": fault.EntityID, "fault": fault.Fault, "onset_ns": startNS + fault.AtS*nsPerSecond, "params": map[string]any{},
	}}
}

func invokeCommand(call JobCall, index int, startNS int64) model.ScenarioCommand {
	args := call.Args
	if args == nil {
		args = map[string]any{}
	}
	return model.ScenarioCommand{Op: model.OpEffectorInvoke, Args: map[string]any{
		"effector": call.Effector, "entity_id": call.EntityID, "command_id": fmt.Sprintf("%s%d", jobCommandBase, index+1),
		"args": args, "at_ns": startNS + call.AtS*nsPerSecond,
	}}
}
