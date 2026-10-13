package domain

import (
	"fmt"
)

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

func (d *Device) BootID() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.bootID
}

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

func (d *Device) ApplyFrame(frame []byte) (Outcome, error) {
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
