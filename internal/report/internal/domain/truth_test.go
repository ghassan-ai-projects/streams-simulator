package domain

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func outcomeLine(id, body string) string {
	return `{"stream":"director","seq":1,"type":"effector.outcome","at_ns":1,"command_id":"` + id + `",` + body + `}`
}

func TestTruthListsEachKindOfCommandInOrder(t *testing.T) {
	t.Parallel()
	commands := []model.Command{
		command(t, 0, model.OpFaultInject, map[string]any{"entity_id": "a", "fault": "f"}),
		command(t, 1, model.OpFaultClear, map[string]any{"fault_id": "f-1"}),
		command(t, 2, model.OpPerturbApply, map[string]any{"perturbation": "drop", "params": map[string]any{"rate": 0.5}}),
		command(t, 3, model.OpPerturbClear, map[string]any{"perturb_id": "p-1"}),
		command(t, 4, model.OpEnvInject, map[string]any{"target": "t", "fault": "kill"}),
		command(t, 5, model.OpClockAdvance, map[string]any{"to_ns": 9}),
	}
	truth, err := truthOf(commands, nil)
	if err != nil || truth.KPIs != nil || len(truth.Faults) != 0 ||
		len(truth.Perturbations) != 2 || truth.Perturbations[0].Params["rate"] != 0.5 ||
		len(truth.Environment) != 1 || len(truth.EffectorCalls) != 0 {
		t.Fatalf("truth = %+v, want no fault intervals without a fault.truth record, 2 perturbations, 1 environment and no calls", truth)
	}
}

func TestEffectorCallsPairWithTheirOutcomesByCommandAndKeepRefusals(t *testing.T) {
	t.Parallel()
	invoke := func(seq int64, id string, at int64) model.Command {
		return command(t, seq, model.OpEffectorInvoke, map[string]any{"effector": "e", "entity_id": "a", "command_id": id, "at_ns": json.Number(strconv.FormatInt(at, 10))})
	}
	director := doc(
		outcomeLine("c1", `"result":{"accepted":true,"mode":"silent_no_effect","ack_latency_ms":12.5}`),
		outcomeLine("c2", `"refused":"interlock"`),
	)
	records, err := readStream("sealed/journal.jsonl", director)
	if err != nil {
		t.Fatal(err)
	}
	calls := effectorCalls([]model.Command{invoke(0, "c1", 7), invoke(1, "c2", 8), invoke(2, "c3", 9)}, records)
	silent, refused, unrecorded := calls[0], calls[1], calls[2]
	if silent.AtNS != 7 || silent.Outcome.Mode != "silent_no_effect" || *silent.Outcome.EffectApplied || *silent.Outcome.AckLatencyMS != 12.5 {
		t.Errorf("silent call = %+v outcome %+v, want silent_no_effect with effect_applied false", silent, silent.Outcome)
	}
	if refused.Outcome.Refusal != "interlock" || unrecorded.Outcome != nil {
		t.Errorf("refused = %+v, unrecorded = %+v, want a refusal and no outcome", refused.Outcome, unrecorded.Outcome)
	}
}
