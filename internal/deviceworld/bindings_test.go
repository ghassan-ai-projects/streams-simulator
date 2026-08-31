package deviceworld

import (
	"os"
	"testing"
)

func TestLoadBindingsFromData(t *testing.T) {
	data, err := os.ReadFile("testdata/thermal.bindings.json")
	if err != nil {
		t.Fatal(err)
	}
	bindings, err := LoadBindings(data, "entity-01")
	if err != nil {
		t.Fatal(err)
	}
	if len(bindings) != 2 {
		t.Fatalf("loaded %d bindings, want fan-01 and led-01", len(bindings))
	}

	fan, ok := bindings["fan-01"]
	if !ok {
		t.Fatal("fan-01 binding is required")
	}
	if fan.effector != "set_fan_duty" || fan.valueState != "fan_duty_true" || fan.safeStopEffector != "stop_fan" {
		t.Fatalf("fan-01 binding = %+v", fan)
	}
	fanArgs, err := fan.args(map[string]float64{"duty_permille": 450})
	if err != nil {
		t.Fatalf("build fan binding arguments: %v", err)
	}
	if fanArgs["reefer_id"] != "entity-01" || fanArgs["duty_permille"] != float64(450) {
		t.Fatalf("fan binding arguments = %#v, want entity and duty parameter", fanArgs)
	}
	fanStopArgs, err := fan.safeStopArgsForEntity()
	if err != nil || fanStopArgs["reefer_id"] != "entity-01" {
		t.Fatalf("fan safe-stop binding args=%#v err=%v", fanStopArgs, err)
	}

	led, ok := bindings["led-01"]
	if !ok {
		t.Fatal("led-01 binding is required for the set_indicator route")
	}
	if led.effector != "set_led" || led.valueState != "led_brightness_true" || led.safeStopEffector != "stop_led" {
		t.Fatalf("led-01 binding = %+v", led)
	}
	ledArgs, err := led.args(map[string]float64{"brightness_permille": 1000})
	if err != nil {
		t.Fatalf("build LED binding arguments: %v", err)
	}
	if ledArgs["reefer_id"] != "entity-01" || ledArgs["brightness_permille"] != float64(1000) || ledArgs["pattern"] != "solid" {
		t.Fatalf("LED binding arguments = %#v, want entity, brightness, and solid pattern", ledArgs)
	}
	ledStopArgs, err := led.safeStopArgsForEntity()
	if err != nil || ledStopArgs["reefer_id"] != "entity-01" {
		t.Fatalf("LED safe-stop binding args=%#v err=%v", ledStopArgs, err)
	}
}

func TestLoadBindingsFailsClosed(t *testing.T) {
	cases := map[string]string{
		"unknown field":     `{"bindings":{"target":{"effector":"effect","unexpected":true}}}`,
		"trailing json":     `{"bindings":{"target":{"effector":"effect"}}} {}`,
		"unknown source":    `{"bindings":{"target":{"effector":"effect","arguments":{"arg":{"source":"unknown"}}}}}`,
		"missing parameter": `{"bindings":{"target":{"effector":"effect","arguments":{"arg":{"source":"parameter"}}}}}`,
		"entity with value": `{"bindings":{"target":{"effector":"effect","arguments":{"arg":{"source":"entity","value":"x"}}}}}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := LoadBindings([]byte(body), "entity-01"); err == nil {
				t.Fatal("invalid binding catalog was accepted")
			}
		})
	}
}

func TestLoadBindingsRequiresEntitySourceValue(t *testing.T) {
	data, err := os.ReadFile("testdata/thermal.bindings.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := LoadBindings(data, ""); err == nil {
		t.Fatal("entity-sourced binding must require a runtime entity")
	}
}

func TestValidateBindingsChecksWorldCompositionBeforeListen(t *testing.T) {
	w := coldChainWorld(t, 1)
	entity := w.EntityIDs()[0]
	data, err := os.ReadFile("testdata/thermal.bindings.json")
	if err != nil {
		t.Fatal(err)
	}
	bindings, err := LoadBindings(data, entity)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateBindings(w, bindings, []string{"fan-01"}, []string{"fan-01"}); err != nil {
		t.Fatalf("valid world binding rejected: %v", err)
	}

	broken := bindings["fan-01"]
	broken.effector = "missing-effector"
	bindings["fan-01"] = broken
	if err := ValidateBindings(w, bindings, []string{"fan-01"}, []string{"fan-01"}); err == nil {
		t.Fatal("unknown world effector accepted")
	}
}
