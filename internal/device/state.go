package device

import (
	"fmt"
)

// Advance moves the default manual clock forward by deltaMicros. It is a no-op
// when a custom Clock was supplied.
func (d *Device) Advance(deltaMicros int64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.manualMono += deltaMicros
	d.expireLease(d.clock())
}

// SetFaults installs the fault state applied to subsequent accepted commands.
func (d *Device) SetFaults(f Faults) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.faults = f
}

// SetFaultSchedule installs a deterministic one-shot fault schedule. Entries
// are keyed by the one-based admission ordinal and are consumed when that
// ordinal is reached.
func (d *Device) SetFaultSchedule(schedule []FaultInjection) error {
	if err := ValidateFaultSchedule(schedule); err != nil {
		return err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.faultSchedule = faultNames(schedule)
	d.acceptedCommands = 0
	return nil
}

// AcceptedCommandCount returns the number of commands that reached the
// schedule's admission ordinal. It is useful for deterministic test evidence.
func (d *Device) AcceptedCommandCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.acceptedCommands
}

// Reboot simulates a power cycle: a new boot identity, safe outputs, and a
// cleared volatile dedup ledger. Commands bound to the old boot are then
// rejected wrong_boot, which is exactly what forces upstream reconciliation.
func (d *Device) Reboot(newBootID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.rebootLocked(newBootID)
}

func (d *Device) rebootLocked(newBootID string) {
	confirmed := d.invokeSafeStopLocked()
	d.bootID = newBootID
	d.recordBootTransition(confirmed)
	d.dedup = map[string]dedupEntry{}
	d.leaseUntilMicros = 0
}

// BootID returns the current device boot identity.
func (d *Device) BootID() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.bootID
}

// State returns a fresh device.state record.
func (d *Device) State() map[string]any {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.expireLease(d.clock())
	state := d.stateIdentity()
	if d.curTarget != "" {
		state["current_output"] = d.currentOutput()
	}
	return state
}

// HandleCommand decodes, validates, and applies one command frame, returning
// the encoded receipt and result frames. When ackLost is true the receipt frame
// must not be written to the wire (the effect still happened).
func (d *Device) HandleCommand(frame []byte) (receipt []byte, result []byte, ackLost bool, err error) {
	outcome, err := d.handleCommand(frame)
	if err != nil {
		return nil, nil, false, err
	}
	receiptFrame, err := EncodeRecord(outcome.Receipt)
	if err != nil {
		return nil, nil, false, fmt.Errorf("device: encode receipt: %w", err)
	}
	resultFrame, err := EncodeRecord(outcome.Result)
	if err != nil {
		return nil, nil, false, fmt.Errorf("device: encode result: %w", err)
	}
	return receiptFrame, resultFrame, outcome.AckLost, nil
}

func (d *Device) handleCommand(frame []byte) (Outcome, error) {
	command, decodeErr := DecodeRecord(frame)
	if decodeErr != nil {
		return Outcome{}, decodeErr
	}
	if command["message_type"] != "command" {
		return Outcome{}, fmt.Errorf("device: expected a command frame, got %v", command["message_type"])
	}
	return d.ApplyCommand(command), nil
}

func (d *Device) stateIdentity() map[string]any {
	return map[string]any{"message_type": "state", "protocol_version": float64(ProtocolVersion),
		"device_id": d.deviceID, "boot_id": d.bootID, "firmware_digest": d.firmwareDigest, "capability_digest": d.capabilityDigest,
		"safe_state": d.safeState, "dedup_ledger": map[string]any{"persistent": false, "size": float64(len(d.dedup))}}
}

func (d *Device) currentOutput() map[string]any {
	return map[string]any{"target": d.curTarget, "operation": d.curOp, "value": d.curValue, "energized": d.energized}
}

func (d *Device) recordBootTransition(confirmed bool) {
	if confirmed {
		d.energized = false
		d.curTarget, d.curOp, d.curValue = "", "", 0
		d.safeState = true
	} else {
		d.energized = true
		d.safeState = false
	}
}
