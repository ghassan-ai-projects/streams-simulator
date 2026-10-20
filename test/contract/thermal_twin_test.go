package contract

// The thermal board's twin scaffolded from its three kind files and a small
// choices document, held to what the hand-written twin is held to: it
// implements the kind files, it is lint clean, the fan lowers the zone
// temperature along the closed form of the first-order lag, and a replay is
// byte for byte the same.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/component"
	"github.com/ghassan-ai-projects/streams-simulator/internal/lint"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func scaffoldedThermalTwin(t *testing.T) *component.Composed {
	t.Helper()
	library, err := component.NewLibrary(nil)
	if err != nil {
		t.Fatal(err)
	}
	choices, err := os.ReadFile(testsupport.TwinThermalChoices())
	if err != nil {
		t.Fatal(err)
	}
	scaffold := func(kinds map[string]map[string]any) (*component.Composed, *model.DomainSpec, error) {
		twin, err := component.Twin(library, kinds, choices, "thermal-board.twin.json")
		if err != nil {
			return nil, nil, fmt.Errorf("scaffold the thermal twin: %w", err)
		}
		return twin, twin.Domain.Spec, nil
	}
	twin, err := plantwire.Scaffold(nil, []string{testsupport.TwinThermalKindsDir()}, "thermal-board.twin.json", scaffold)
	if err != nil {
		t.Fatal(err)
	}
	return twin
}

func twinWorld(t *testing.T, twin *component.Composed, seed uint64, opts world.Options) *world.World {
	t.Helper()
	w, err := world.New(twin.Domain, seed, "twin", model.DefaultStartTimeNS, opts)
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func checkTwinCooling(t *testing.T, w *world.World, times ...float64) {
	t.Helper()
	for _, seconds := range times {
		want, got := coolingCurve(seconds), arduinoState(w, arduinoZone, "temperature_value", seconds)
		if math.Abs(got-want) > 0.1 {
			t.Errorf("zone temperature at %v s = %.4f, closed form %.4f; the held-input step allows 0.1", seconds, got, want)
		}
	}
}

func TestTheScaffoldedThermalTwinImplementsItsKindFilesAndPassesTheLint(t *testing.T) {
	t.Parallel()
	twin := scaffoldedThermalTwin(t)
	if want := []string{"thermal_sensor", "indicator", "fan"}; !slices.Equal(twin.Implements, want) {
		t.Fatalf("implements = %v, want %v", twin.Implements, want)
	}
	if checked, err := plantwire.CheckImplements(twin.Domain.Spec, []string{testsupport.TwinThermalKindsDir()}); err != nil || len(checked) != 3 {
		t.Fatalf("CheckImplements = %v, %v, want the scaffolded twin to equal its three kind files", checked, err)
	}
	report, err := lint.Check(twin.Domain, lint.Options{Seed: lint.DefaultSeed, Hours: lint.DefaultHours})
	if err != nil || len(report.Findings) != 0 {
		t.Fatalf("lint = %+v, %v, want a nominal day with no finding", report, err)
	}
}

func TestTheFanOfTheScaffoldedTwinLowersTheZoneTemperatureAlongTheClosedFormOfItsLag(t *testing.T) {
	t.Parallel()
	twin := scaffoldedThermalTwin(t)
	quiet := world.Options{Noiseless: true, EmitDisabled: true}
	w, idle := twinWorld(t, twin, 11, quiet), twinWorld(t, twin, 11, quiet)
	arduinoFanAt(t, w, "cool", 0.5, 600)
	checkTwinCooling(t, w, 0, 6, 10)
	if low, rest := arduinoState(w, arduinoZone, "temperature_value", 10), arduinoState(idle, arduinoZone, "temperature_value", 10); rest-low < 0.5 {
		t.Errorf("zone temperature at 10 s, lease active = %.4f against %.4f with no command, want it at least 0.5 degC lower", low, rest)
	}
	checkTwinCooling(t, w, 10.5, 30, 120, 600)
}

func TestTheLeaseAndTheSafeStopsOfTheScaffoldedTwinTurnTheFanAndTheLedOff(t *testing.T) {
	t.Parallel()
	w := twinWorld(t, scaffoldedThermalTwin(t), 11, world.Options{Noiseless: true, EmitDisabled: true})
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
		{arduinoFan, "duty_permille_value", 2, 450}, {arduinoFan, "duty_permille_value", 4, 0},
		{arduinoFan, "duty_permille_value", 25, 450}, {arduinoFan, "duty_permille_value", 31, 0},
		{arduinoLED, "brightness_permille_value", 2, 700}, {arduinoLED, "pattern_value", 2, 1},
		{arduinoLED, "brightness_permille_value", 401, 0}, {arduinoLED, "pattern_value", 401, 0},
		{arduinoLED, "brightness_permille_value", 501, 200}, {arduinoLED, "pattern_value", 501, 2},
	} {
		if got := arduinoState(w, tc.entity, tc.state, tc.seconds); got != tc.want {
			t.Errorf("%s %s at %v s = %v, want %v", tc.entity, tc.state, tc.seconds, got, tc.want)
		}
	}
}

func twinTrace(t *testing.T, twin *component.Composed, seed uint64) ([]byte, int) {
	t.Helper()
	w := twinWorld(t, twin, seed, world.Options{})
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

func TestTheScaffoldedTwinIsTheSameBytesEveryTimeAndReplaysTheSameTraceForOneSeed(t *testing.T) {
	t.Parallel()
	first, second := scaffoldedThermalTwin(t), scaffoldedThermalTwin(t)
	if !bytes.Equal(first.Bytes, second.Bytes) {
		t.Fatal("two scaffolds of one choices document differ")
	}
	a, count := twinTrace(t, first, 11)
	b, _ := twinTrace(t, second, 11)
	if count < 50 || !bytes.Equal(a, b) {
		t.Fatalf("%d events; two runs with one seed equal = %v, want a non-trivial trace that replays byte for byte", count, bytes.Equal(a, b))
	}
}

func TestADifferentSeedGivesADifferentTraceOfTheSameScaffoldedTwin(t *testing.T) {
	t.Parallel()
	twin := scaffoldedThermalTwin(t)
	a, _ := twinTrace(t, twin, 11)
	b, _ := twinTrace(t, twin, 12)
	if bytes.Equal(a, b) {
		t.Fatal("seeds 11 and 12 gave one trace: the twin ignores its seed, so the equality of two runs with one seed proves nothing")
	}
}

const scaffoldedThermalTwinDigest = "sha256:65b3325c7ad2e26b3eb26b412d3ee41cc3c2506298b654812c34b77f44dbb4c7"

func TestTheScaffoldedThermalTwinIsTheKnownDocument(t *testing.T) {
	t.Parallel()
	if got := canonical.DigestBytes(scaffoldedThermalTwin(t).Bytes); got != scaffoldedThermalTwinDigest {
		t.Fatalf("digest of the scaffolded thermal twin = %s, want %s: if the choices, the kind files or the scaffold changed on purpose, set the new digest", got, scaffoldedThermalTwinDigest)
	}
}
