package cli

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ghassan-ai-projects/streams-simulator/internal/device"
)

// cmdDevice serves the wire-faithful serial device emulator as a gateway link.
// It is the device end of the Real-World Sensor HIL-0 loop: the Agentic Stream
// serial effector connects to this socket and speaks the device gateway link
// (state/command/receipt; execution truth is queried as state).
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
	bootID := fs.String("boot-id", "boot-A", "initial device boot identity")
	deviceID := fs.String("device-id", "dev-01", "device identity")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	if *socket == "" {
		return fmt.Errorf("device serve requires --socket")
	}
	if *capabilities == "" {
		return fmt.Errorf("device serve requires --capabilities")
	}
	capabilityData, err := os.ReadFile(*capabilities)
	if err != nil {
		return fmt.Errorf("streamsim: read capabilities: %w", err)
	}
	caps, err := device.LoadCapabilities(capabilityData)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}

	dev := device.New(device.Config{BootID: *bootID, DeviceID: *deviceID, Capabilities: caps})
	listener, err := device.Listen(*socket, dev)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	fmt.Fprintf(os.Stderr, "streamsim device: listening on %s (device_id=%s boot_id=%s)\n", *socket, *deviceID, *bootID)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	fmt.Fprintln(os.Stderr, "streamsim device: shutting down")
	return listener.Close()
}
