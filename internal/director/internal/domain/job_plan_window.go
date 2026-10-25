package domain

import (
	"cmp"
	"maps"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func FaultIDs(faults []JobFault) []string {
	ids := make([]string, len(faults))
	for rank, index := range InjectionOrder(faults) {
		ids[index] = world.FaultID(rank)
	}
	return ids
}

func InjectionOrder(faults []JobFault) []int {
	order := make([]int, len(faults))
	for index := range order {
		order[index] = index
	}
	slices.SortStableFunc(order, func(a, b int) int { return cmp.Compare(faults[a].AtS, faults[b].AtS) })
	return order
}

func clearsAt(req JobRequest, ids []string, startNS, atS int64) []model.ScenarioCommand {
	var commands []model.ScenarioCommand
	for index, fault := range req.Faults {
		if fault.UntilS != nil && *fault.UntilS == atS {
			commands = append(commands, clearCommand(ids[index], atS, startNS))
		}
	}
	return commands
}

func clearCommand(faultID string, atS, startNS int64) model.ScenarioCommand {
	return model.ScenarioCommand{Op: model.OpFaultClear, Args: map[string]any{"fault_id": faultID, "at_ns": startNS + atS*nsPerSecond}}
}

func perturbationsAt(req JobRequest, startNS, atS int64) []model.ScenarioCommand {
	var commands []model.ScenarioCommand
	for _, perturbation := range req.Perturbations {
		if perturbation.AtS == atS {
			commands = append(commands, applyCommand(perturbation, startNS))
		}
	}
	return commands
}

func applyCommand(perturbation JobPerturbation, startNS int64) model.ScenarioCommand {
	params := map[string]any{}
	maps.Copy(params, perturbation.Params)
	until := int64(0)
	if perturbation.UntilS != nil {
		until = startNS + *perturbation.UntilS*nsPerSecond
	}
	return model.ScenarioCommand{Op: model.OpPerturbApply, Args: map[string]any{
		"perturbation": perturbation.Kind, "params": params, "from_ns": startNS + perturbation.AtS*nsPerSecond, "until_ns": until,
	}}
}
