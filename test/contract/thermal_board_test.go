package contract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

const (
	arduinoZone    = "zone-01"
	arduinoFan     = "fan-01"
	arduinoLED     = "led-01"
	arduinoRestC   = 34.0
	arduinoCooledC = 22.0
	arduinoTauS    = 180.0
)

func arduinoSpec(t *testing.T) *domain.Compiled {
	t.Helper()
	spec, err := domain.Load(testsupport.ThermalBoardDomain())
	if err != nil {
		t.Fatal(err)
	}
	return spec
}

func arduinoWorld(t *testing.T, seed uint64, opts world.Options) *world.World {
	t.Helper()
	w, err := world.New(arduinoSpec(t), seed, "arduino", model.DefaultStartTimeNS, opts)
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func arduinoAt(seconds float64) int64 { return model.DefaultStartTimeNS + int64(seconds*1e9) }

func arduinoCommand(t *testing.T, w *world.World, effector, entity, id string, seconds float64, args map[string]any) {
	t.Helper()
	result, err := w.InvokeEffector(effector, entity, id, args, arduinoAt(seconds))
	if err != nil {
		t.Fatal(err)
	}
	if !result.Accepted {
		t.Fatalf("%s %s at %v s: refused (%s), want the board to accept it", effector, id, seconds, result.Reason)
	}
}

func arduinoFanAt(t *testing.T, w *world.World, id string, seconds, duty float64) {
	t.Helper()
	arduinoCommand(t, w, "set_pwm_lease", arduinoFan, id, seconds, map[string]any{"duty_permille": duty, "lease_ms": 10000.0})
}

func arduinoState(w *world.World, entity, state string, seconds float64) float64 {
	return w.StateValue(entity, state, arduinoAt(seconds))
}

func coolingCurve(seconds float64) float64 {
	const start, lapse = 0.5, 10.5
	switch {
	case seconds <= start:
		return arduinoRestC
	case seconds <= lapse:
		return arduinoCooledC + (arduinoRestC-arduinoCooledC)*math.Exp(-(seconds-start)/arduinoTauS)
	}
	return arduinoRestC + (coolingCurve(lapse)-arduinoRestC)*math.Exp(-(seconds-lapse)/arduinoTauS)
}

func checkCoolingCurve(t *testing.T, w *world.World, times ...float64) {
	t.Helper()
	for _, seconds := range times {
		want, got := coolingCurve(seconds), arduinoState(w, arduinoZone, "zone_temperature", seconds)
		if math.Abs(got-want) > 0.1 {
			t.Errorf("zone temperature at %v s = %.4f, closed form %.4f; the held-input step allows 0.1", seconds, got, want)
		}
	}
}

func TestTheFanLowersTheZoneTemperatureAlongTheClosedFormOfItsFirstOrderLag(t *testing.T) {
	t.Parallel()
	w := arduinoWorld(t, 11, world.Options{Noiseless: true, EmitDisabled: true})
	idle := arduinoWorld(t, 11, world.Options{Noiseless: true, EmitDisabled: true})
	arduinoFanAt(t, w, "cool", 0.5, 600)
	checkCoolingCurve(t, w, 0, 6, 10)
	if low, rest := arduinoState(w, arduinoZone, "zone_temperature", 10), arduinoState(idle, arduinoZone, "zone_temperature", 10); rest-low < 0.5 {
		t.Errorf("zone temperature at 10 s, lease active = %.4f against %.4f with no command, want it at least 0.5 degC lower", low, rest)
	}
	checkCoolingCurve(t, w, 10.5, 30, 120, 600)
}

func TestARenewedFanLeaseHoldsTheZoneNearAmbientWhileAnIdleFanLeavesItWarm(t *testing.T) {
	t.Parallel()
	idle := arduinoWorld(t, 11, world.Options{Noiseless: true, EmitDisabled: true})
	fanned := arduinoWorld(t, 11, world.Options{Noiseless: true, EmitDisabled: true})
	for i := range 120 {
		arduinoFanAt(t, fanned, fmt.Sprintf("renew-%d", i), 0.5+5*float64(i), 600)
	}
	idleC, fannedC := arduinoState(idle, arduinoZone, "zone_temperature", 600), arduinoState(fanned, arduinoZone, "zone_temperature", 600)
	if math.Abs(idleC-arduinoRestC) > 1e-6 || arduinoRestC-fannedC < 10 || fannedC > arduinoCooledC+1 {
		t.Fatalf("zone at 600 s: %.4f idle, %.4f fanned; want the idle fan at rest %v and the renewed fan within 1 degC of ambient %v", idleC, fannedC, arduinoRestC, arduinoCooledC)
	}
	idleRH, fannedRH := arduinoState(idle, arduinoZone, "zone_humidity", 600), arduinoState(fanned, arduinoZone, "zone_humidity", 600)
	if fannedRH-idleRH < 15 {
		t.Fatalf("humidity at 600 s: %.2f idle, %.2f fanned; want the cooler air to read at least 15 points wetter", idleRH, fannedRH)
	}
}

func TestTheLeaseAndTheSafeStopsTurnTheFanAndTheLedOff(t *testing.T) {
	t.Parallel()
	w := arduinoWorld(t, 11, world.Options{Noiseless: true, EmitDisabled: true})
	arduinoFanAt(t, w, "on-1", 1, 450)
	arduinoCommand(t, w, "set_led", arduinoLED, "led-1", 1, map[string]any{"brightness_permille": 700.0, "pattern": 1.0})
	arduinoCommand(t, w, "stop_fan", arduinoFan, "stop-1", 3, map[string]any{})
	arduinoFanAt(t, w, "on-2", 20, 450)
	arduinoCommand(t, w, "stop_led", arduinoLED, "led-stop", 400, map[string]any{})
	arduinoCommand(t, w, "set_led", arduinoLED, "led-2", 500, map[string]any{"brightness_permille": 200.0, "pattern": 2.0})
	for _, tc := range []struct {
		entity, state string
		seconds, want float64
	}{
		{arduinoFan, "fan_duty", 2, 450}, {arduinoFan, "fan_duty", 4, 0}, {arduinoFan, "fan_duty", 25, 450}, {arduinoFan, "fan_duty", 31, 0},
		{arduinoLED, "led_brightness", 2, 700}, {arduinoLED, "led_pattern", 2, 1}, {arduinoLED, "led_brightness", 300, 700},
		{arduinoLED, "led_brightness", 401, 0}, {arduinoLED, "led_pattern", 401, 0},
		{arduinoLED, "led_brightness", 501, 200}, {arduinoLED, "led_pattern", 501, 2},
	} {
		if got := arduinoState(w, tc.entity, tc.state, tc.seconds); got != tc.want {
			t.Errorf("%s %s at %v s = %v, want %v", tc.entity, tc.state, tc.seconds, got, tc.want)
		}
	}
}

func arduinoTrace(t *testing.T, seed uint64) ([]byte, int) {
	t.Helper()
	w := arduinoWorld(t, seed, world.Options{})
	var events []model.SimEvent
	w.SetEmitter(func(event model.SimEvent) { events = append(events, event) })
	arduinoFanAt(t, w, "fan-on", 20, 450)
	arduinoCommand(t, w, "set_led", arduinoLED, "led-on", 25, map[string]any{"brightness_permille": 1000.0, "pattern": 2.0})
	if _, _, err := w.Advance(arduinoAt(120)); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(events)
	if err != nil {
		t.Fatal(err)
	}
	return raw, len(events)
}

func TestTheArduinoTwinReplaysTheSameBytesForOneSeedAndOtherBytesForAnother(t *testing.T) {
	t.Parallel()
	first, count := arduinoTrace(t, 11)
	second, _ := arduinoTrace(t, 11)
	other, _ := arduinoTrace(t, 12)
	if count < 100 || !bytes.Equal(first, second) {
		t.Fatalf("%d events; two runs with one seed equal = %v, want a non-trivial trace that replays byte for byte", count, bytes.Equal(first, second))
	}
	if bytes.Equal(first, other) {
		t.Fatal("two seeds produced one trace: the instruments draw no noise")
	}
}
