package contract

import (
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func TestTheStuckOffLedGoesDarkWhileTheBoardStillReportsItsPattern(t *testing.T) {
	t.Parallel()
	w := faultedArduinoWorld(t, arduinoLED, "led_stuck_off")
	arduinoCommand(t, w, "set_led", arduinoLED, "led-on", 30, map[string]any{"brightness_permille": 800.0, "pattern": 2.0})
	for _, tc := range []struct {
		state         string
		seconds, want float64
	}{
		{"led_brightness", 59, 800}, {"led_brightness", 61, 0}, {"led_brightness", 600, 0}, {"led_pattern", 600, 2},
	} {
		if got := arduinoState(w, arduinoLED, tc.state, tc.seconds); got != tc.want {
			t.Errorf("%s at %v s = %v, want %v", tc.state, tc.seconds, got, tc.want)
		}
	}
}

func stalledFanCurve(seconds float64) float64 {
	const commandS, decay = 55.0, 5.0 / arduinoTauS
	switch {
	case seconds <= commandS:
		return arduinoRestC
	case seconds <= arduinoFaultOnsetS:
		return arduinoCooledC + (arduinoRestC-arduinoCooledC)*math.Exp(-(seconds-commandS)/arduinoTauS)
	}
	cooled := arduinoCooledC + (arduinoRestC-arduinoCooledC)*math.Exp(-decay)
	return arduinoRestC + (cooled-arduinoRestC)*math.Exp(-(seconds-arduinoFaultOnsetS)/arduinoTauS)
}

func TestTheStalledFanDropsItsDutyAndTheZoneWarmsBackToRest(t *testing.T) {
	t.Parallel()
	stalled := faultedArduinoWorld(t, arduinoFan, "fan_stall")
	running := arduinoWorld(t, 11, world.Options{Noiseless: true, EmitDisabled: true})
	arduinoFanAt(t, stalled, "cool", 55, 600)
	arduinoFanAt(t, running, "cool", 55, 600)
	for _, tc := range []struct{ seconds, want float64 }{{59, 600}, {61, 0}, {64, 0}} {
		if got := arduinoState(stalled, arduinoFan, "fan_duty", tc.seconds); got != tc.want {
			t.Errorf("fan duty at %v s = %v, want %v", tc.seconds, got, tc.want)
		}
	}
	for _, seconds := range []float64{58, 62, 65, 120, 600} {
		if want, got := stalledFanCurve(seconds), arduinoState(stalled, arduinoZone, "zone_temperature", seconds); math.Abs(got-want) > 0.1 {
			t.Errorf("zone temperature at %v s = %.4f, closed form %.4f; the held-input step allows 0.1", seconds, got, want)
		}
	}
	if warm, cool := arduinoState(stalled, arduinoZone, "zone_temperature", 65), arduinoState(running, arduinoZone, "zone_temperature", 65); warm-cool < 0.2 {
		t.Errorf("zone at 65 s = %.4f stalled, %.4f running; want the stalled fan to leave it at least 0.2 degC warmer", warm, cool)
	}
}
