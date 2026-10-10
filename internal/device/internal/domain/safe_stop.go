package domain

import (
	"errors"
	"fmt"
	"log/slog"
)

// applyPlant resolves the output value and energized state for an accepted command.
// A bound Plant (e.g. the world oracle) is authoritative; otherwise the value is
// the command's energize_field and energized means that field is positive — both
// read from the capability catalog, never switched on an operation name.
func (d *Device) applyPlant(cmd PlantCommand) (value float64, energized bool, err error) {
	// Stuck devices accept without handing effects to the plant.
	if d.faults.Stuck {
		return 0, false, nil
	}
	if d.plant != nil {
		return d.applyBoundPlant(cmd)
	}
	capability, ok := d.capabilities.target(cmd.Target)
	if !ok {
		return 0, false, nil
	}
	value = cmd.Params[capability.EnergizeField]
	return value, value > 0, nil
}

func (d *Device) expireLease(now int64) {
	if d.leaseUntilMicros == 0 || now < d.leaseUntilMicros {
		return
	}
	d.recordSafeTransition(d.invokeSafeStopLocked())
	d.leaseUntilMicros = 0
}

func (d *Device) invokeSafeStopLocked() bool {
	if d.curTarget == "" || d.capabilities == nil || !d.capabilities.hasSafeStop(d.curTarget) {
		return true
	}
	return d.applySafeStop(d.curTarget, d.clock())
}

func (d *Device) applySafeStop(target string, atMicros int64) bool {
	stopper, ok := d.plant.(SafeStopper)
	if !ok {
		return true
	}
	effect, err := stopper.SafeStop(target, atMicros)
	return safeStopConfirmed(target, effect, err)
}

func leaseDeadline(now int64, params map[string]float64) int64 {
	leaseMS, ok := params["lease_ms"]
	if !ok || leaseMS <= 0 {
		return 0
	}
	return now + int64(leaseMS*1000)
}

func plantRejectCode(err error) string {
	if errors.Is(err, ErrPlantInterlocked) {
		return "interlocked"
	}
	return "not_ready"
}

func (d *Device) rejection(commandID string, now int64, code string) Outcome {
	receipt := d.commandReceipt(commandID, now, false)
	receipt["reject_code"] = code
	return Outcome{Receipt: receipt, Result: d.rejectedResult(commandID, code)}
}

func containsFault(faults []string, want string) bool {
	for _, fault := range faults {
		if fault == want {
			return true
		}
	}
	return false
}

func resultDetail(energized bool) string {
	if energized {
		return "output energized"
	}
	return "accepted, output not energized"
}

func (d *Device) applyBoundPlant(command PlantCommand) (float64, bool, error) {
	effect, err := d.plant.Apply(command)
	if err != nil {
		return 0, false, fmt.Errorf("device: apply plant command: %w", err)
	}
	return effect.Value, effect.Energized, nil
}

func (d *Device) recordSafeTransition(confirmed bool) {
	if confirmed {
		d.energized = false
		d.safeState = true
	} else {
		// Preserve conservative evidence when a physical stop is unconfirmed.
		d.energized = true
		d.safeState = false
	}
}

func safeStopConfirmed(target string, effect PlantEffect, err error) bool {
	if err != nil {
		slog.Error("device safe stop failed", "target", target, "error", err)
		return false
	}
	if effect.Energized {
		slog.Error("device safe stop did not de-energize target", "target", target)
		return false
	}
	return true
}
