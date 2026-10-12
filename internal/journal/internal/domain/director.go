package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

// Command records a command-log entry.
func (j *Journal) Command(atNS int64, command model.Command) {
	j.director.Append(TypeCommand, atNS, map[string]any{"command": command})
}

// Delivery records a delivery-ledger row.
func (j *Journal) Delivery(atNS int64, row model.LedgerRecord) {
	j.director.Append(TypeDelivery, atNS, map[string]any{"delivery": row})
}

// State records the world-state snapshot taken at an emission.
func (j *Journal) State(atNS int64, snapshot model.StateSnapshot) {
	j.director.Append(TypeState, atNS, map[string]any{"state": snapshot})
}

// Outcome records the world's record of an effector call: its result, or the
// refusal when the call was declined.
func (j *Journal) Outcome(atNS int64, call Call, result any, refusal string) {
	fields := map[string]any{"command_id": call.CommandID, "effector": call.Effector, "entity_id": call.EntityID}
	if refusal != "" {
		fields["refused"] = refusal
	} else {
		fields["result"] = result
	}
	j.director.Append(TypeEffectorOutcome, atNS, fields)
}

// Unblinded records that the run was unblinded and why. The wall-clock moment
// is in the run artifact, not here: the stream stays a pure function of the run.
func (j *Journal) Unblinded(atNS int64, reason string) {
	j.director.Append(TypeRunUnblinded, atNS, map[string]any{"reason": reason})
}
