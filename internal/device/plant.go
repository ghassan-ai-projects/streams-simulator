package device

import "errors"

var (
	// ErrPlantInterlocked identifies a plant refusal caused by an independent
	// safety interlock.
	ErrPlantInterlocked = errors.New("device plant interlocked")
	// ErrPlantUnavailable identifies a plant that could not apply the command.
	ErrPlantUnavailable = errors.New("device plant unavailable")
)

// Plant is the physical process the device actuates. The emulator asks the
// plant to apply an accepted, in-bounds command and reports back the resulting
// output value and whether it is actually energized — the ground truth
// verification is scored against, independent of the command's acknowledgement.
//
// A real integration binds this to the Streams Simulator world (see the
// deviceworld package), so temperature responds to the fan and the world's
// effector failure modes (confirmed/silent no-effect, interlock) become the
// device's observed truth. The default in-memory plant is enough to prove the
// wire loop and the desired≠observed cases.
type Plant interface {
	Apply(PlantCommand) (PlantEffect, error)
}

// SafeStopper is an optional plant lifecycle seam. A device invokes it when a
// lease expires or the device reboots, allowing a world-backed plant to record
// the safe-stop effect instead of only clearing device-local state.
type SafeStopper interface {
	SafeStop(target string, atMicros int64) (PlantEffect, error)
}

// PlantCommand is one accepted, in-bounds actuation handed to the plant. It
// carries the device command identity and monotonic time so a world-backed
// plant can invoke its effector idempotently and at the right instant.
type PlantCommand struct {
	Target    string
	Operation string
	Params    map[string]float64
	CommandID string
	AtMicros  int64
}

// PlantEffect is what the plant did: the resulting output value and whether the
// output is energized.
type PlantEffect struct {
	Value     float64
	Energized bool
}
