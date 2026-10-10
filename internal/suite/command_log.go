package suite

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type scenarioCommandLog struct{ commands []model.Command }

func (log *scenarioCommandLog) append(op string, args map[string]any, at int64) {
	log.commands = append(log.commands, model.Command{Seq: int64(len(log.commands)), AtNS: at, Op: op, Args: args})
}

func (log *scenarioCommandLog) appendSetup(setup []model.SetupCall) {
	for _, call := range setup {
		log.append(model.OpEffectorInvoke, map[string]any{"effector": call.Effector, "entity_id": call.EntityID,
			"command_id": call.CommandID, "args": call.Args, "at_ns": call.AtNS}, call.AtNS)
	}
}

func (log *scenarioCommandLog) appendPerturbations(perturbations []Perturbation) {
	for _, item := range perturbations {
		log.append(model.OpPerturbApply, map[string]any{"perturbation": item.Name, "params": item.Params,
			"from_ns": item.FromNS, "until_ns": item.UntilNS}, item.FromNS)
	}
}
