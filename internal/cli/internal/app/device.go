package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/cli/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/cli/internal/serve"
	"github.com/ghassan-ai-projects/streams-simulator/internal/device"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// cmdDevice serves the wire-faithful serial device emulator as a gateway link.
// It is the device end of the Real-World Sensor HIL-0 loop: the Agentic Stream
// serial effector connects to this socket and speaks the device gateway link
// (state/command/receipt/result; execution truth is queried as state).
func cmdDevice(s *session, args []string) (any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("device requires a subcommand: serve")
	}
	switch args[0] {
	case "serve":
		return nil, cmdDeviceServe(s, args[1:])
	default:
		return nil, fmt.Errorf("device: unknown subcommand %q (want: serve)", args[0])
	}
}

func cmdDeviceServe(s *session, args []string) error {
	options, err := parseDeviceServeOptions(args, s.stderr)
	if err != nil {
		return err
	}
	dev, _, err := newDeviceServeDevice(options, s.now)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	banner := fmt.Sprintf("streamsim device: listening on %s (device_id=%s boot_id=%s)", options.socket, options.deviceID, options.bootID)
	return serve.Device(s.stderr, options.socket, dev, banner)
}

func newDeviceServeDevice(options deviceServeOptions, now func() time.Time) (*device.Device, *world.World, error) {
	caps, err := loadDeviceCapabilities(options.capabilities)
	if err != nil {
		return nil, nil, err
	}
	if err := device.ValidateFaultSchedule(options.faultSchedule.entries); err != nil {
		return nil, nil, fmt.Errorf("validate fault schedule: %w", err)
	}
	plant, worldState, clock, err := prepareDeviceWorld(options, caps, now)
	if err != nil {
		return nil, nil, err
	}
	return device.New(deviceConfig(options, caps, plant, clock)), worldState, nil
}

func deviceConfig(options deviceServeOptions, caps *device.Capabilities, plant device.Plant, clock func() int64) device.Config {
	return device.Config{BootID: options.bootID, DeviceID: options.deviceID, Capabilities: caps,
		Plant: plant, Clock: clock, FaultSchedule: options.faultSchedule.entries}
}

func prepareDeviceWorld(options deviceServeOptions, caps *device.Capabilities, now func() time.Time) (device.Plant, *world.World, func() int64, error) {
	if options.worldBindings == "" {
		return nil, nil, nil, nil
	}
	data, err := files.Read(options.worldBindings)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read world bindings: %w", err)
	}
	w, err := createDeviceWorld(options.worldDomain)
	if err != nil {
		return nil, nil, nil, err
	}
	return bindDeviceWorld(w, data, options.worldEntity, caps, now)
}

type faultSpecFlag struct {
	entries []device.FaultInjection
}

func (f *faultSpecFlag) String() string {
	if f == nil {
		return ""
	}
	values := make([]string, 0, len(f.entries))
	for _, entry := range f.entries {
		values = append(values, entry.Name)
	}
	return strings.Join(values, ",")
}

func (f *faultSpecFlag) Set(value string) error {
	for _, spec := range strings.Split(value, ",") {
		entry, err := device.ParseFaultSpec(spec)
		if err != nil {
			return fmt.Errorf("parse fault %q: %w", spec, err)
		}
		f.entries = append(f.entries, entry)
	}
	return nil
}
