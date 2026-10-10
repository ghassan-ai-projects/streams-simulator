package domain

// ApplyCommand runs the device's admission logic on an already-decoded command
// and returns the receipt and result records. It is the deterministic core of
// the emulator; ApplyFrame is the wire wrapper.
func (d *Device) ApplyCommand(command map[string]any) Outcome {
	d.mu.Lock()
	defer d.mu.Unlock()
	commandID, _ := command["command_id"].(string)
	key, _ := command["idempotency_key"].(string)
	now := d.clock()
	d.expireLease(now)
	digest, err := semanticCommandDigest(command)
	if err != nil {
		return d.rejection(commandID, now, "malformed")
	}
	return d.applyIdentifiedCommand(command, commandID, key, digest, now)
}

func (d *Device) scheduledFreshnessRejection(commandID string, now int64, injections []string, fault string, ordinal int) (Outcome, bool) {
	for _, injection := range injections {
		code := scheduledFreshnessCode(injection)
		if code == "" {
			continue
		}
		outcome := d.rejection(commandID, now, code)
		outcome.Fault = fault
		outcome.AcceptedCommand = ordinal
		return outcome, true
	}
	return Outcome{}, false
}

func (d *Device) injectExecutionFaults(injections []string) {
	for _, injection := range injections {
		switch injection {
		case FaultAckLost:
			d.faults.AckLost = true
		case FaultStuck:
			d.faults.Stuck = true
		}
	}
}

func (d *Device) applyResponseFaults(outcome Outcome, injections []string, ordinal int) Outcome {
	for _, injection := range injections {
		d.applyResponseFault(&outcome, injection, ordinal)
	}
	return outcome
}

func (d *Device) applyAcceptedCommand(command map[string]any, commandID string, now int64) (Outcome, bool) {
	target, _ := command["target"].(string)
	operation, _ := command["operation"].(string)
	params, _ := numericParams(command["parameters"])
	if operation == "safe_stop" {
		return d.executeSafeStop(target, operation, commandID, now)
	}
	return d.executePlantCommand(PlantCommand{Target: target, Operation: operation, Params: params, CommandID: commandID, AtMicros: now})
}
