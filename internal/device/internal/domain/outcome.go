package domain

func (d *Device) acceptedOutcome(id string, now int64, status string) Outcome {
	return Outcome{Receipt: d.commandReceipt(id, now, true), Result: d.commandCompletion(id, now, status), AckLost: d.faults.AckLost}
}

func (d *Device) commandReceipt(id string, now int64, accepted bool) map[string]any {
	return map[string]any{"message_type": "receipt", "protocol_version": float64(ProtocolVersion),
		"command_id": id, "boot_id": d.bootID, "accepted": accepted, "received_mono_us": float64(now)}
}

func (d *Device) commandCompletion(id string, now int64, status string) map[string]any {
	return map[string]any{"message_type": "result", "protocol_version": float64(ProtocolVersion),
		"command_id": id, "boot_id": d.bootID, "status": status, "completed_mono_us": float64(now)}
}

func (d *Device) rejectedResult(id, code string) map[string]any {
	return map[string]any{"message_type": "result", "protocol_version": float64(ProtocolVersion),
		"command_id": id, "boot_id": d.bootID, "status": "rejected", "error_code": code}
}
