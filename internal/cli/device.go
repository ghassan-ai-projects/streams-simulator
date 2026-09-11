package cli

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/device"
	"github.com/ghassan-ai-projects/streams-simulator/internal/deviceworld"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
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
	fs := flag.NewFlagSet("device serve", flag.ExitOnError)
	socket := fs.String("socket", "", "Unix domain socket path to listen on (required)")
	capabilities := fs.String("capabilities", "", "device capability catalog JSON path (required)")
	worldBindings := fs.String("world", "", "deviceworld binding catalog JSON path (optional)")
	worldDomain := fs.String("world-domain", "domains/cold-chain-transit.domain.json", "world domain JSON path when --world is set")
	worldEntity := fs.String("world-entity", "", "world entity bound to device targets; defaults to the first entity")
	bootID := fs.String("boot-id", "boot-A", "initial device boot identity")
	deviceID := fs.String("device-id", "dev-01", "device identity")
	var faultSchedule faultSpecFlag
	fs.Var(&faultSchedule, "fault", "deterministic fault name[@accepted-command-ordinal]; repeatable")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	if *socket == "" {
		return fmt.Errorf("device serve requires --socket")
	}
	if *capabilities == "" {
		return fmt.Errorf("device serve requires --capabilities")
	}
	dev, _, err := newDeviceServeDevice(*capabilities, *worldBindings, *worldDomain, *worldEntity, *bootID, *deviceID, faultSchedule.entries)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	listener, err := device.Listen(*socket, dev)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	fmt.Fprintf(os.Stderr, "streamsim device: listening on %s (device_id=%s boot_id=%s)\n", *socket, *deviceID, *bootID)

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

func newDeviceServeDevice(capabilityPath, bindingPath, domainPath, entity, bootID, deviceID string, schedule []device.FaultInjection) (*device.Device, *world.World, error) {
	capabilityData, err := os.ReadFile(capabilityPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read capabilities: %w", err)
	}
	caps, err := device.LoadCapabilities(capabilityData)
	if err != nil {
		return nil, nil, fmt.Errorf("load capabilities: %w", err)
	}
	if err := device.ValidateFaultSchedule(schedule); err != nil {
		return nil, nil, fmt.Errorf("validate fault schedule: %w", err)
	}

	var plant device.Plant
	var worldState *world.World
	var clock func() int64
	if bindingPath != "" {
		worldData, readErr := os.ReadFile(bindingPath)
		if readErr != nil {
			return nil, nil, fmt.Errorf("read world bindings: %w", readErr)
		}
		domainSpec, loadErr := domain.Load(domainPath)
		if loadErr != nil {
			return nil, nil, fmt.Errorf("load device world domain: %w", loadErr)
		}
		worldState, err = world.New(domainSpec, 1, "device-world", model.DefaultStartTimeNS, world.Options{EmitDisabled: true})
		if err != nil {
			return nil, nil, fmt.Errorf("create device world: %w", err)
		}
		if entity == "" {
			ids := worldState.EntityIDs()
			if len(ids) == 0 {
				return nil, nil, fmt.Errorf("device world has no entities")
			}
			entity = ids[0]
		}
		if worldState.Entity(entity) == nil {
			return nil, nil, fmt.Errorf("device world entity %q does not exist", entity)
		}
		bindings, loadErr := deviceworld.LoadBindings(worldData, entity)
		if loadErr != nil {
			return nil, nil, fmt.Errorf("load device world bindings: %w", loadErr)
		}
		requiredTargets := make([]string, 0, len(bindings))
		for target := range bindings {
			requiredTargets = append(requiredTargets, target)
		}
		requiredSafeStops := make([]string, 0, len(bindings))
		for _, target := range caps.SafeStopNames() {
			if _, ok := bindings[target]; ok {
				requiredSafeStops = append(requiredSafeStops, target)
			}
		}
		if validateErr := deviceworld.ValidateBindings(worldState, bindings, requiredTargets, requiredSafeStops); validateErr != nil {
			return nil, nil, fmt.Errorf("validate device world bindings: %w", validateErr)
		}
		plant = deviceworld.New(worldState, bindings)
		started := time.Now()
		clock = func() int64 { return time.Since(started).Microseconds() }
	}
	return device.New(device.Config{
		BootID:        bootID,
		DeviceID:      deviceID,
		Capabilities:  caps,
		Plant:         plant,
		Clock:         clock,
		FaultSchedule: schedule,
	}), worldState, nil
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
