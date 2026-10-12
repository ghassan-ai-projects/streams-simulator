package domain

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type outcomeRecord struct {
	Refused string `json:"refused"`
	Result  struct {
		Accepted      *bool    `json:"accepted"`
		Mode          string   `json:"mode"`
		EffectApplied *bool    `json:"effect_applied"`
		AckLatencyMS  *float64 `json:"ack_latency_ms"`
	} `json:"result"`
}

func effectorCalls(commands []model.Command, director []journal.Record) []EffectorCall {
	pending := outcomesByCommand(director)
	return collect(commands, func(c model.Command) (EffectorCall, bool) {
		return callOf(c, pending)
	})
}

func callOf(c model.Command, pending map[string][]*Outcome) (EffectorCall, bool) {
	if c.Op != model.OpEffectorInvoke {
		return EffectorCall{}, false
	}
	id := text(c.Args, "command_id")
	call := EffectorCall{Seq: c.Seq, AtNS: callTime(c), CommandID: id, Effector: text(c.Args, "effector"),
		EntityID: text(c.Args, "entity_id"), Args: object(c.Args, "args")}
	call.Outcome, pending[id] = nextOutcome(pending[id])
	return call, true
}

func callTime(c model.Command) int64 {
	if at := optionalInt(c.Args, "at_ns"); at != nil {
		return *at
	}
	return c.AtNS
}

func outcomesByCommand(director []journal.Record) map[string][]*Outcome {
	byCommand := map[string][]*Outcome{}
	for _, record := range director {
		if record.Type != journal.TypeEffectorOutcome {
			continue
		}
		id, _ := record.Fields["command_id"].(string)
		byCommand[id] = append(byCommand[id], outcomeOf(record))
	}
	return byCommand
}

func nextOutcome(queue []*Outcome) (*Outcome, []*Outcome) {
	if len(queue) == 0 {
		return nil, nil
	}
	return queue[0], queue[1:]
}

func outcomeOf(record journal.Record) *Outcome {
	var decoded outcomeRecord
	_ = convert(record.Fields, &decoded)
	if decoded.Refused != "" {
		return &Outcome{Refusal: decoded.Refused}
	}
	result := decoded.Result
	applied := result.EffectApplied != nil && *result.EffectApplied
	return &Outcome{Accepted: result.Accepted, Mode: result.Mode, EffectApplied: &applied, AckLatencyMS: result.AckLatencyMS}
}
