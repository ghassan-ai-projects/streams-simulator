package cli

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ghassan-ai-projects/streams-simulator/internal/device"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// cmdDevice serves the wire-faithful serial device emulator as a gateway link.
// It is the device end of the Real-World Sensor HIL-0 loop: the Agentic Stream
// serial effector connects to this socket and speaks the device gateway link
// (state/command/receipt/result; execution truth is queried as state).
func cmdDevice(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("device requires a subcommand: serve")
	}
	switch args[0] {
	case "serve":
		return cmdDeviceServe(args[1:])
	default:
		return fmt.Errorf("device: unknown subcommand %q (want: serve)", args[0])
	}
}

func cmdDeviceServe(args []string) error {
	options, err := parseDeviceServeOptions(args)
	if err != nil {
		return err
	}
	dev, _, err := newDeviceServeDevice(options.capabilities, options.worldBindings, options.worldDomain, options.worldEntity, options.bootID, options.deviceID, options.faultSchedule.entries)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return serveDevice(dev, options)
}

func newDeviceServeDevice(capabilityPath, bindingPath, domainPath, entity, bootID, deviceID string, schedule []device.FaultInjection) (*device.Device, *world.World, error) {
	caps, err := loadDeviceCapabilities(capabilityPath)
	if err != nil {
		return nil, nil, err
	}
	if err := device.ValidateFaultSchedule(schedule); err != nil {
		return nil, nil, fmt.Errorf("validate fault schedule: %w", err)
	}
	plant, worldState, clock, err := prepareDeviceWorld(bindingPath, domainPath, entity, caps)
	if err != nil {
		return nil, nil, err
	}
	return device.New(device.Config{BootID: bootID, DeviceID: deviceID, Capabilities: caps,
		Plant: plant, Clock: clock, FaultSchedule: schedule}), worldState, nil
}

func prepareDeviceWorld(bindingPath, domainPath, entity string, caps *device.Capabilities) (device.Plant, *world.World, func() int64, error) {
	if bindingPath == "" {
		return nil, nil, nil, nil
	}
	data, err := os.ReadFile(bindingPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read world bindings: %w", err)
	}
	w, err := createDeviceWorld(domainPath)
	if err != nil {
		return nil, nil, nil, err
	}
	return bindDeviceWorld(w, data, entity, caps)
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

func awaitDeviceShutdown(listener *net.UnixListener) error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(stop)
	<-stop
	fmt.Fprintln(os.Stderr, "streamsim device: shutting down")
	if err := listener.Close(); err != nil {
		return fmt.Errorf("streamsim: close device listener: %w", err)
	}
	return nil
}

func serveDevice(dev *device.Device, options deviceServeOptions) error {
	listener, err := device.Listen(options.socket, dev)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	fmt.Fprintf(os.Stderr, "streamsim device: listening on %s (device_id=%s boot_id=%s)\n", options.socket, options.deviceID, options.bootID)
	return awaitDeviceShutdown(listener)
}
