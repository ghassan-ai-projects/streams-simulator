package device

import (
	"strconv"
	"strings"
)

func (d *Device) applyIdentifiedCommand(command map[string]any, id, key, digest string, now int64) Outcome {
	// A matching key replays execution evidence without another plant effect.
	if prior, ok := d.dedup[key]; ok {
		if prior.digest != digest {
			return d.rejection(id, now, "duplicate")
		}
		return prior.outcome
	}
	if reject := d.admit(command, now); reject != "" {
		return d.rejection(id, now, reject)
	}
	return d.executeScheduledCommand(command, id, key, digest, now)
}

type commandFaults struct {
	ordinal    int
	injections []string
	name       string
	original   Faults
}

func (d *Device) takeCommandFaults() commandFaults {
	d.acceptedCommands++
	ordinal := d.acceptedCommands
	injections := d.faultSchedule[ordinal]
	delete(d.faultSchedule, ordinal)
	faults := commandFaults{ordinal: ordinal, injections: injections, name: strings.Join(injections, ","), original: d.faults}
	d.injectExecutionFaults(injections)
	return faults
}

func (d *Device) executeScheduledCommand(command map[string]any, id, key, digest string, now int64) Outcome {
	faults := d.takeCommandFaults()
	if outcome, rejected := d.scheduledFreshnessRejection(id, now, faults.injections, faults.name, faults.ordinal); rejected {
		d.faults = faults.original
		return outcome
	}
	outcome, accepted := d.applyAcceptedCommand(command, id, now)
	return d.completeScheduledCommand(outcome, accepted, key, digest, faults)
}

func (d *Device) completeScheduledCommand(outcome Outcome, accepted bool, key, digest string, faults commandFaults) Outcome {
	d.faults = faults.original
	outcome.Fault = faults.name
	outcome.AcceptedCommand = faults.ordinal
	if accepted {
		outcome = d.applyResponseFaults(outcome, faults.injections, faults.ordinal)
		if !containsFault(faults.injections, FaultReboot) {
			d.rememberExecution(key, digest, outcome)
		}
	}
	return outcome
}

func (d *Device) rememberExecution(key, digest string, outcome Outcome) {
	// Ack loss belongs to this wire delivery; retries recover the receipt.
	outcome.AckLost = false
	d.dedup[key] = dedupEntry{digest: digest, outcome: outcome}
}

func scheduledFreshnessCode(injection string) string {
	switch injection {
	case FaultStale:
		return "not_ready"
	case FaultExpired:
		return "expired"
	default:
		return ""
	}
}

func (d *Device) applyResponseFault(outcome *Outcome, injection string, ordinal int) {
	switch injection {
	case FaultDuplicate:
		outcome.Duplicate = true
	case FaultDisconnect:
		outcome.Disconnect = true
	case FaultReboot:
		d.rebootLocked("boot-reboot-" + strconv.Itoa(ordinal))
		// The old-boot response must not escape after reboot.
		outcome.AckLost = true
	}
}

func (d *Device) executeSafeStop(target, operation, id string, now int64) (Outcome, bool) {
	if !d.applySafeStop(target, now) {
		return d.rejection(id, now, "not_ready"), false
	}
	d.curTarget, d.curOp, d.curValue, d.energized = target, operation, 0, false
	d.safeState = true
	d.leaseUntilMicros = 0
	return d.acceptedOutcome(id, now, "safe_state"), true
}

func (d *Device) executePlantCommand(command PlantCommand) (Outcome, bool) {
	value, energized, err := d.applyPlant(command)
	if err != nil {
		return d.rejection(command.CommandID, command.AtMicros, plantRejectCode(err)), false
	}
	d.curTarget, d.curOp, d.curValue, d.energized = command.Target, command.Operation, value, energized
	d.safeState = !energized
	d.leaseUntilMicros = leaseDeadline(command.AtMicros, command.Params)
	outcome := d.acceptedOutcome(command.CommandID, command.AtMicros, "executed")
	outcome.Result["detail"] = resultDetail(energized)
	return outcome, true
}
