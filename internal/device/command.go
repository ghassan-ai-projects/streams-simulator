package device

import (
	"strconv"
	"strings"
)

// ApplyCommand runs the device's admission logic on an already-decoded command
// and returns the receipt and result records. It is the deterministic core of
// the emulator; HandleCommand is the wire wrapper.
func (d *Device) ApplyCommand(command map[string]any) Outcome {
	d.mu.Lock()
	defer d.mu.Unlock()

	commandID, _ := command["command_id"].(string)
	idempotencyKey, _ := command["idempotency_key"].(string)
	now := d.clock()
	d.expireLease(now)
	digest, digestErr := semanticCommandDigest(command)
	if digestErr != nil {
		return d.rejection(commandID, now, "malformed")
	}

	// Idempotency: a repeated key replays the prior outcome and applies no new
	// effect, regardless of faults.
	if prior, ok := d.dedup[idempotencyKey]; ok {
		if prior.digest != digest {
			return d.rejection(commandID, now, "duplicate")
		}
		return prior.outcome
	}

	if reject := d.admit(command, now); reject != "" {
		return d.rejection(commandID, now, reject)
	}
	d.acceptedCommands++
	ordinal := d.acceptedCommands
	injections := d.faultSchedule[ordinal]
	delete(d.faultSchedule, ordinal)
	fault := strings.Join(injections, ",")
	originalFaults := d.faults
	d.injectExecutionFaults(injections)

	if outcome, rejected := d.scheduledFreshnessRejection(commandID, now, injections, fault, ordinal); rejected {
		d.faults = originalFaults
		return outcome
	}

	outcome, accepted := d.applyAcceptedCommand(command, commandID, now)
	d.faults = originalFaults
	outcome.Fault = fault
	outcome.AcceptedCommand = ordinal
	if accepted {
		outcome = d.applyResponseFaults(outcome, injections, ordinal)
		if !containsFault(injections, FaultReboot) {
			replay := outcome
			// Ack loss is a property of this wire delivery, not of the
			// idempotent execution. A later retry must be able to recover the
			// receipt without applying the plant a second time.
			replay.AckLost = false
			d.dedup[idempotencyKey] = dedupEntry{digest: digest, outcome: replay}
		}
	}
	return outcome
}

func (d *Device) scheduledFreshnessRejection(commandID string, now int64, injections []string, fault string, ordinal int) (Outcome, bool) {
	for _, injection := range injections {
		if injection != FaultStale && injection != FaultExpired {
			continue
		}
		code := "not_ready"
		if injection == FaultExpired {
			code = "expired"
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
		switch injection {
		case FaultDuplicate:
			outcome.Duplicate = true
		case FaultDisconnect:
			outcome.Disconnect = true
		case FaultReboot:
			d.rebootLocked("boot-reboot-" + strconv.Itoa(ordinal))
			// The response was created under the old boot identity. Do not
			// emit it after reboot; require the caller to observe new state.
			outcome.AckLost = true
		}
	}
	return outcome
}

func (d *Device) applyAcceptedCommand(command map[string]any, commandID string, now int64) (Outcome, bool) {
	target, _ := command["target"].(string)
	operation, _ := command["operation"].(string)
	params, _ := numericParams(command["parameters"])
	if operation == "safe_stop" {
		if !d.applySafeStop(target, now) {
			return d.rejection(commandID, now, "not_ready"), false
		}
		d.curTarget, d.curOp, d.curValue, d.energized = target, operation, 0, false
		d.safeState = true
		d.leaseUntilMicros = 0
		return Outcome{
			Receipt: map[string]any{
				"message_type": "receipt", "protocol_version": float64(ProtocolVersion),
				"command_id": commandID, "boot_id": d.bootID, "accepted": true,
				"received_mono_us": float64(now),
			},
			Result: map[string]any{
				"message_type": "result", "protocol_version": float64(ProtocolVersion),
				"command_id": commandID, "boot_id": d.bootID, "status": "safe_state",
				"completed_mono_us": float64(now),
			},
			AckLost: d.faults.AckLost,
		}, true
	}
	plantCommand := PlantCommand{
		Target: target, Operation: operation, Params: params,
		CommandID: commandID, AtMicros: now,
	}
	value, energized, err := d.applyPlant(plantCommand)
	if err != nil {
		return d.rejection(commandID, now, plantRejectCode(err)), false
	}
	d.curTarget, d.curOp, d.curValue, d.energized = target, operation, value, energized
	d.safeState = !energized
	d.leaseUntilMicros = leaseDeadline(now, params)

	return Outcome{
		Receipt: map[string]any{
			"message_type":     "receipt",
			"protocol_version": float64(ProtocolVersion),
			"command_id":       commandID,
			"boot_id":          d.bootID,
			"accepted":         true,
			"received_mono_us": float64(now),
		},
		Result: map[string]any{
			"message_type":      "result",
			"protocol_version":  float64(ProtocolVersion),
			"command_id":        commandID,
			"boot_id":           d.bootID,
			"status":            "executed",
			"detail":            resultDetail(energized),
			"completed_mono_us": float64(now),
		},
		AckLost: d.faults.AckLost,
	}, true
}
