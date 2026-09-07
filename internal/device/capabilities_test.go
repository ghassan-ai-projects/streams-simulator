package device

import (
	"os"
	"testing"
)

// This digest is shared with Agentic Stream's conformance fixture. A mismatch
// prevents the live device handshake even when each repository passes alone.
const thermalCapabilityCatalogDigest = "sha256:177552ccdaa8d71ac3eb1e27a1a60e3433e1e2792bad2dc16ea95acb2d735eaf"
const legacyCapabilityCatalogDigest = "sha256:ed9ebf9685f9933c15578f06c65e059f906cfd1edcd8f75b778b6a4f794cc494"

func TestLoadCanonicalCapabilityCatalogAndDigest(t *testing.T) {
	data, err := os.ReadFile("testdata/thermal_capability_catalog.json")
	if err != nil {
		t.Fatal(err)
	}
	caps, err := LoadCapabilities(data)
	if err != nil {
		t.Fatalf("load canonical catalog: %v", err)
	}
	if got := caps.Digest(); got != thermalCapabilityCatalogDigest {
		t.Fatalf("canonical catalog digest = %q, want %q", got, thermalCapabilityCatalogDigest)
	}
	fan, ok := caps.target("fan-01")
	if !ok {
		t.Fatal("canonical route must derive fan-01")
	}
	if fan.Operation != "set_pwm_lease" || fan.EnergizeField != "duty_permille" {
		t.Fatalf("fan-01 capability = %+v", fan)
	}
	if fan.Bounds["duty_permille"] != [2]float64{0, 600} {
		t.Fatalf("fan duty bounds = %v, want preset min plus catalog max", fan.Bounds["duty_permille"])
	}
	if fan.Bounds["lease_ms"] != [2]float64{5000, 10000} {
		t.Fatalf("fan lease bounds = %v, want preset min plus catalog max", fan.Bounds["lease_ms"])
	}
	if !caps.hasSafeStop("fan-01") || !caps.hasSafeStop("led-01") {
		t.Fatal("canonical safe stops must be retained")
	}
}

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

func TestLegacyCapabilityCatalogDigestRemainsUnscoped(t *testing.T) {
	data, err := os.ReadFile("testdata/thermal.capabilities.json")
	if err != nil {
		t.Fatal(err)
	}
	caps, err := LoadCapabilities(data)
	if err != nil {
		t.Fatalf("load legacy catalog: %v", err)
	}
	if got := caps.Digest(); got != legacyCapabilityCatalogDigest {
		t.Fatalf("legacy catalog digest = %q, want historical digest %q", got, legacyCapabilityCatalogDigest)
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
		"unknown-field":        `{"protocol_version": 1, "targets": {"x": {"operation": "o", "energize_field": "f", "bounds": {"f": {"min": 0, "max": 1}}, "typo": true}}}`,
		"trailing-json":        `{"protocol_version": 1, "targets": {"x": {"operation": "o", "energize_field": "f", "bounds": {"f": {"min": 0, "max": 1}}}}} {}`,
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
