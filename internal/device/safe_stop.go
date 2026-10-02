package device

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
	// A stuck device accepts the command but does not hand it to the plant; this
	// keeps a world-backed oracle aligned with the observed no-effect.
	if d.faults.Stuck {
		return 0, false, nil
	}
	if d.plant != nil {
		e, err := d.plant.Apply(cmd)
		if err != nil {
			return 0, false, fmt.Errorf("device: apply plant command: %w", err)
		}
		return e.Value, e.Energized, nil
	}
	capa, ok := d.capabilities.target(cmd.Target)
	if !ok {
		return 0, false, nil
	}
	value = cmd.Params[capa.EnergizeField]
	return value, value > 0, nil
}

func (d *Device) expireLease(now int64) {
	if d.leaseUntilMicros == 0 || now < d.leaseUntilMicros {
		return
	}
	safeStopSucceeded := d.invokeSafeStopLocked()
	if safeStopSucceeded {
		d.energized = false
		d.safeState = true
	} else {
		// A failed stop leaves the physical output unknown; do not claim the
		// device is safe or erase the last energized observation.
		d.energized = true
		d.safeState = false
	}
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
	return Outcome{
		Receipt: map[string]any{
			"message_type":     "receipt",
			"protocol_version": float64(ProtocolVersion),
			"command_id":       commandID,
			"boot_id":          d.bootID,
			"accepted":         false,
			"reject_code":      code,
			"received_mono_us": float64(now),
		},
		Result: map[string]any{
			"message_type":     "result",
			"protocol_version": float64(ProtocolVersion),
			"command_id":       commandID,
			"boot_id":          d.bootID,
			"status":           "rejected",
			"error_code":       code,
		},
	}
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
