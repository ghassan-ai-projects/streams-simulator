package app

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/device"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

var fixedNow = time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)

func worldOptions() OpenOptions {
	return OpenOptions{
		Capabilities:  testsupport.DeviceCapabilities(),
		WorldBindings: testsupport.DeviceWorldBindings(),
		WorldDomain:   testsupport.Domain("cold-chain-transit"),
		BootID:        "boot-A", DeviceID: "dev-01", Now: time.Now,
	}
}

func TestOpenDeviceBindsWorldPlant(t *testing.T) {
	t.Parallel()
	dev, worldState, err := OpenDevice(worldOptions())
	if err != nil {
		t.Fatalf("build world-backed device: %v", err)
	}
	if worldState == nil {
		t.Fatal("world binding must construct a world")
	}

	commandData, err := os.ReadFile(testsupport.DeviceConformance("valid", "command.json"))
	if err != nil {
		t.Fatal(err)
	}
	var command map[string]any
	if err := json.Unmarshal(commandData, &command); err != nil {
		t.Fatal(err)
	}
	command["not_before_mono_us"] = float64(0)
	out := dev.ApplyCommand(command)
	if out.Receipt["accepted"] != true {
		t.Fatalf("valid command must be accepted: %v", out.Receipt)
	}
	calls := worldState.EffectorCalls()
	if len(calls) != 1 {
		t.Fatalf("device command must invoke world effector once, got %d", len(calls))
	}
	output, ok := dev.State()["current_output"].(map[string]any)
	if !ok || output["energized"] != calls[0].EffectApplied {
		t.Fatalf("device state must report world truth: output=%v calls=%+v", output, calls)
	}
}

func TestOpenDeviceWithoutBindingsHasNoWorld(t *testing.T) {
	t.Parallel()
	options := worldOptions()
	options.WorldBindings = ""
	dev, worldState, err := OpenDevice(options)
	if err != nil || dev == nil || worldState != nil {
		t.Fatalf("device = %v, world = %v, err = %v; want a device without a world", dev, worldState, err)
	}
}

func TestOpenDeviceNamesWhatItCannotCompose(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "absent.json")
	for name, tc := range map[string]struct {
		change func(*OpenOptions)
		want   string
	}{
		"capabilities": {func(o *OpenOptions) { o.Capabilities = missing }, "read capabilities: "},
		"bindings":     {func(o *OpenOptions) { o.WorldBindings = missing }, "read world bindings: "},
		"world domain": {func(o *OpenOptions) { o.WorldDomain = missing }, "load device world domain: "},
		"entity":       {func(o *OpenOptions) { o.WorldEntity = "no-such" }, `device world entity "no-such" does not exist`},
		"fault schedule": {func(o *OpenOptions) {
			o.FaultSchedule = []device.FaultInjection{{Name: "no_such_fault", AcceptedCommand: 1}}
		}, "validate fault schedule: "},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			options := worldOptions()
			tc.change(&options)
			if _, _, err := OpenDevice(options); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want %q", err, tc.want)
			}
		})
	}
	options := worldOptions()
	options.Capabilities = missing
	if _, _, err := OpenDevice(options); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a missing catalog must stay a file-system failure: %v", err)
	}
}

func TestDeviceClockCountsMicrosecondsOfTheInjectedClock(t *testing.T) {
	t.Parallel()
	now := fixedNow
	clock := deviceClock(func() time.Time { return now })
	if got := clock(); got != 0 {
		t.Fatalf("clock at start = %d, want 0", got)
	}
	now = now.Add(1500 * time.Microsecond)
	if got := clock(); got != 1500 {
		t.Fatalf("clock after 1.5ms = %d, want 1500", got)
	}
}
