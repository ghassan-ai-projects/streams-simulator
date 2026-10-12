package domain

import (
	"os"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

// This digest is shared with Agentic Stream's conformance fixture. A mismatch
// prevents the live device handshake even when each repository passes alone.
const thermalCapabilityCatalogDigest = "sha256:0d61225286c628cfba8cbf7aea514e1fdc95918b514b4b810516dbe0fc44fc76"

func TestLoadCanonicalCapabilityCatalogAndDigest(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(testsupport.DeviceCapabilities())
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
	t.Parallel()
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
	t.Parallel()
	const presets = `"presets": {"on": {"f": 1}}`
	const route = `"x": {"operation": "o", "target": "t", "selector_field": "s", "expires_after_ms": 1000, ` + presets
	cases := map[string]struct{ body, want string }{
		"bad-json":        {`{`, "unexpected EOF"},
		"wrong-version":   {`{"protocol_version": 2, "routes": {` + route + `}}}`, "protocol_version 2 != 1"},
		"no-routes":       {`{"protocol_version": 1, "routes": {}}`, "declare no routes"},
		"legacy-targets":  {`{"protocol_version": 1, "targets": {}}`, `unknown field "targets"`},
		"no-operation":    {`{"protocol_version": 1, "routes": {"x": {"target": "t", "selector_field": "s", "expires_after_ms": 1000, ` + presets + `}}}`, "must set operation"},
		"no-presets":      {`{"protocol_version": 1, "routes": {"x": {"operation": "o", "target": "t", "selector_field": "s", "expires_after_ms": 1000}}}`, "has no presets"},
		"max-below-min":   {`{"protocol_version": 1, "routes": {` + route + `, "bounds": {"f": {"min": 5, "max": 1}}}}}`, "max < min"},
		"unknown-field":   {`{"protocol_version": 1, "routes": {` + route + `, "typo": true}}}`, `unknown field "typo"`},
		"trailing-json":   {`{"protocol_version": 1, "routes": {` + route + `}}} {}`, "trailing JSON"},
		"duplicate-route": {`{"protocol_version": 1, "routes": {` + route + `}, "y": {"operation": "o", "target": "t", "selector_field": "s", "expires_after_ms": 1000, ` + presets + `}}}`, "multiple routes"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := LoadCapabilities([]byte(tc.body)); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("%s: err = %v, want %q", name, err, tc.want)
			}
		})
	}
}
