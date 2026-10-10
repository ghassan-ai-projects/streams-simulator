package device_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/device"
)

// countingPlant is a plant that records how often it was driven.
type countingPlant struct {
	applyCalls    int
	safeStopCalls int
	safeStopErr   error
}

func (p *countingPlant) Apply(device.PlantCommand) (device.PlantEffect, error) {
	p.applyCalls++
	return device.PlantEffect{Value: 450, Energized: true}, nil
}

func (p *countingPlant) SafeStop(string, int64) (device.PlantEffect, error) {
	p.safeStopCalls++
	return device.PlantEffect{}, p.safeStopErr
}

// testCaps loads the thermal capability catalog from its data fixture, the
// same data-not-code path the emulator uses at runtime.
func testCaps(t *testing.T) *device.Capabilities {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "thermal_capability_catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	caps, err := device.LoadCapabilities(data)
	if err != nil {
		t.Fatal(err)
	}
	return caps
}

// validCommand loads the vendored golden command fixture (fan-01, boot-A, in
// range) so transport tests exercise the real contract, then applies optional
// mutations.
func validCommand(t *testing.T, mutate func(map[string]any)) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("contract", "conformance", "v1", "valid", "command.json"))
	if err != nil {
		t.Fatal(err)
	}
	var command map[string]any
	if err := json.Unmarshal(raw, &command); err != nil {
		t.Fatal(err)
	}
	command["not_before_mono_us"] = float64(0)
	if mutate != nil {
		mutate(command)
	}
	return command
}
