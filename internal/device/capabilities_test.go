package device

import "testing"

func TestLoadCapabilitiesFromData(t *testing.T) {
	caps := testCaps(t)
	fan, ok := caps.target("fan-01")
	if !ok {
		t.Fatal("fan-01 must load from the data fixture")
	}
	if fan.Operation != "set_pwm_lease" || fan.EnergizeField != "duty_permille" {
		t.Fatalf("fan-01 loaded wrong: %+v", fan)
	}
	if b := fan.Bounds["duty_permille"]; b != [2]float64{0, 600} {
		t.Fatalf("fan-01 duty bounds wrong: %v", b)
	}
	if _, ok := caps.target("nope"); ok {
		t.Fatal("undeclared target must not resolve")
	}
}

func TestLoadCapabilitiesFailsClosed(t *testing.T) {
	cases := map[string]string{
		"bad-json":             `{`,
		"wrong-version":        `{"protocol_version": 2, "targets": {"x": {"operation": "o", "energize_field": "f", "bounds": {"f": {"min": 0, "max": 1}}}}}`,
		"no-targets":           `{"protocol_version": 1, "targets": {}}`,
		"no-operation":         `{"protocol_version": 1, "targets": {"x": {"energize_field": "f", "bounds": {"f": {"min": 0, "max": 1}}}}}`,
		"energize-not-bounded": `{"protocol_version": 1, "targets": {"x": {"operation": "o", "energize_field": "f", "bounds": {"g": {"min": 0, "max": 1}}}}}`,
		"max-below-min":        `{"protocol_version": 1, "targets": {"x": {"operation": "o", "energize_field": "f", "bounds": {"f": {"min": 5, "max": 1}}}}}`,
	}
	for name, body := range cases {
		body := body
		t.Run(name, func(t *testing.T) {
			if _, err := LoadCapabilities([]byte(body)); err == nil {
				t.Fatalf("%s must fail closed, but loaded", name)
			}
		})
	}
}
