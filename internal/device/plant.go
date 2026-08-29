package device

// Plant is the physical process the device actuates. The emulator asks the
// plant to apply an accepted, in-bounds command and reports back whether the
// output is actually energized — the ground truth verification is scored
// against, independent of the command's acknowledgement.
//
// A real integration binds this to the Streams Simulator world (world.World's
// effector/plant), so temperature responds to the fan. The default in-memory
// plant is enough to prove the wire loop and the desired≠observed cases.
type Plant interface {
	// Apply drives target with operation and numeric parameters. It returns the
	// resulting output value and whether the output is energized. The emulator
	// calls it only for accepted commands; a stuck-actuator fault is modeled by
	// the emulator, not the plant.
	Apply(target, operation string, params map[string]float64) (value float64, energized bool)
}

// memPlant is a minimal deterministic plant: a PWM/lease output is energized
// while its duty is positive; an indicator is energized while its state is not
// "off". It keeps no history beyond the latest output.
type memPlant struct{}

// NewMemoryPlant returns the default in-memory plant.
func NewMemoryPlant() Plant { return memPlant{} }

func (memPlant) Apply(_, operation string, params map[string]float64) (float64, bool) {
	switch operation {
	case "set_pwm_lease":
		duty := params["duty_permille"]
		return duty, duty > 0
	case "set_indicator":
		// Indicator state is carried as a numeric level here (0 = off).
		level := params["level"]
		return level, level > 0
	default:
		return 0, false
	}
}
