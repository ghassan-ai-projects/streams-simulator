package domain

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func commandsOf(art *model.RunArtifact, director []journal.Record) ([]model.Command, error) {
	if art != nil {
		return art.CommandLog, nil
	}
	return journalCommands(director)
}

func journalCommands(director []journal.Record) ([]model.Command, error) {
	var commands []model.Command
	for _, record := range director {
		if record.Type != journal.TypeCommand {
			continue
		}
		command, err := commandOf(record)
		if err != nil {
			return nil, err
		}
		commands = append(commands, command)
	}
	return commands, nil
}

func commandOf(record journal.Record) (model.Command, error) {
	var command model.Command
	if err := convert(record.Fields["command"], &command); err != nil {
		return model.Command{}, fmt.Errorf("report: sealed/journal.jsonl record %d: %w", record.Seq, err)
	}
	return command, nil
}

func convert(from, to any) error {
	raw, err := json.Marshal(from)
	if err != nil {
		return fmt.Errorf("encode: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(to); err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	return nil
}

func truthOf(commands []model.Command, director []journal.Record) *Truth {
	return &Truth{
		Faults:        collect(commands, faultOf),
		Perturbations: collect(commands, perturbationOf),
		Environment:   collect(commands, environmentOf),
		EffectorCalls: effectorCalls(commands, director),
	}
}

func collect[T any](commands []model.Command, build func(model.Command) (T, bool)) []T {
	items := []T{}
	for _, command := range commands {
		if item, ok := build(command); ok {
			items = append(items, item)
		}
	}
	return items
}

func faultOf(c model.Command) (Fault, bool) {
	return pick(c, model.OpFaultInject, injectedFault, model.OpFaultClear, clearedFault)
}

func injectedFault(c model.Command) Fault {
	return Fault{Seq: c.Seq, AtNS: c.AtNS, Action: "inject", EntityID: text(c.Args, "entity_id"),
		Fault: text(c.Args, "fault"), OnsetNS: optionalInt(c.Args, "onset_ns"), Params: object(c.Args, "params")}
}

func clearedFault(c model.Command) Fault {
	return Fault{Seq: c.Seq, AtNS: c.AtNS, Action: "clear", FaultID: text(c.Args, "fault_id")}
}

func perturbationOf(c model.Command) (Perturbation, bool) {
	return pick(c, model.OpPerturbApply, appliedPerturbation, model.OpPerturbClear, clearedPerturbation)
}

func appliedPerturbation(c model.Command) Perturbation {
	return Perturbation{Seq: c.Seq, AtNS: c.AtNS, Action: "apply", Perturbation: text(c.Args, "perturbation"),
		FromNS: optionalInt(c.Args, "from_ns"), UntilNS: optionalInt(c.Args, "until_ns"), Params: object(c.Args, "params")}
}

func clearedPerturbation(c model.Command) Perturbation {
	return Perturbation{Seq: c.Seq, AtNS: c.AtNS, Action: "clear", PerturbID: text(c.Args, "perturb_id")}
}

func pick[T any](c model.Command, opA string, a func(model.Command) T, opB string, b func(model.Command) T) (T, bool) {
	switch c.Op {
	case opA:
		return a(c), true
	case opB:
		return b(c), true
	}
	var none T
	return none, false
}

func environmentOf(c model.Command) (Environment, bool) {
	if c.Op != model.OpEnvInject {
		return Environment{}, false
	}
	return Environment{Seq: c.Seq, AtNS: c.AtNS, Target: text(c.Args, "target"), Fault: text(c.Args, "fault"),
		Params: object(c.Args, "params")}, true
}

func text(args map[string]any, key string) string {
	value, _ := args[key].(string)
	return value
}

func optionalInt(args map[string]any, key string) *int64 {
	if _, present := args[key]; !present {
		return nil
	}
	value := integerOf(args[key])
	return &value
}

func object(args map[string]any, key string) map[string]any {
	value, _ := args[key].(map[string]any)
	if len(value) == 0 {
		return nil
	}
	return value
}
