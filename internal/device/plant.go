package device

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
	Apply(PlantCommand) PlantEffect
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

// memPlant is a minimal deterministic plant: a PWM/lease output is energized
// while its duty is positive; an indicator is energized while its level is
// above zero. It keeps no history beyond the latest output.
type memPlant struct{}

// NewMemoryPlant returns the default in-memory plant.
func NewMemoryPlant() Plant { return memPlant{} }

func (memPlant) Apply(cmd PlantCommand) PlantEffect {
	switch cmd.Operation {
	case "set_pwm_lease":
		duty := cmd.Params["duty_permille"]
		return PlantEffect{Value: duty, Energized: duty > 0}
	case "set_indicator":
		level := cmd.Params["level"]
		return PlantEffect{Value: level, Energized: level > 0}
	default:
		return PlantEffect{}
	}
}
