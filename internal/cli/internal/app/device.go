package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/device"
	"github.com/ghassan-ai-projects/streams-simulator/internal/deviceworld"
)

func cmdDevice(s *session, args []string) (any, error) {
	if len(args) == 0 {
		return nil, usagef("device requires a subcommand: serve")
	}
	switch args[0] {
	case "serve":
		return nil, cmdDeviceServe(s, args[1:])
	case "-h", "--help":
		_, _ = fmt.Fprintln(s.stderr, "Usage: streamsim device serve --socket <path> --capabilities <path> [flags]\nRun `streamsim device serve -h` for the flags.")
		return nil, usageExit{code: 0}
	default:
		return nil, usagef("device: unknown subcommand %q (want: serve)", args[0])
	}
}

func cmdDeviceServe(s *session, args []string) error {
	options, err := parseDeviceServeOptions(args, s.stderr)
	if err != nil {
		return err
	}
	dev, _, err := deviceworld.OpenDevice(deviceOpenOptions(options, s.now))
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return serveDevice(s, options, dev)
}

func serveDevice(s *session, options deviceServeOptions, dev *device.Device) error {
	banner := fmt.Sprintf("streamsim device: listening on %s (device_id=%s boot_id=%s)", options.socket, options.deviceID, options.bootID)
	if err := device.Serve(s.stderr, options.socket, dev, banner); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

func deviceOpenOptions(options deviceServeOptions, now func() time.Time) deviceworld.OpenOptions {
	return deviceworld.OpenOptions{Capabilities: options.capabilities, WorldBindings: options.worldBindings,
		WorldDomain: options.worldDomain, WorldEntity: options.worldEntity, BootID: options.bootID,
		DeviceID: options.deviceID, FaultSchedule: options.faultSchedule.entries, Now: now}
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
