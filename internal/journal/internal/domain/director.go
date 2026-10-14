package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

func (j *Journal) Command(atNS int64, command model.Command) {
	j.director.Append(TypeCommand, atNS, map[string]any{"command": command})
}

func (j *Journal) Delivery(atNS int64, row model.LedgerRecord) {
	j.director.Append(TypeDelivery, atNS, map[string]any{"delivery": row})
}

func (j *Journal) State(atNS int64, snapshot model.StateSnapshot) {
	j.director.Append(TypeState, atNS, map[string]any{"state": snapshot})
}

func (j *Journal) Outcome(atNS int64, call Call, result any, refusal string) {
	fields := map[string]any{"command_id": call.CommandID, "effector": call.Effector, "entity_id": call.EntityID}
	if refusal != "" {
		fields["refused"] = refusal
	} else {
		fields["result"] = result
	}
	j.director.Append(TypeEffectorOutcome, atNS, fields)
}

func (j *Journal) KPITruth(atNS int64, truth *model.KPITruth) {
	j.director.Append(TypeKPITruth, atNS, map[string]any{"kpis": truth})
}

func (j *Journal) FaultTruth(atNS int64, faults []model.FaultInterval) {
	j.director.Append(TypeFaultTruth, atNS, map[string]any{"faults": faults})
}

func (j *Journal) Unblinded(atNS int64, reason string) {
	j.director.Append(TypeRunUnblinded, atNS, map[string]any{"reason": reason})
}
