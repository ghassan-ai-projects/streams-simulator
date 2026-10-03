package cli

import (
	"flag"
	"fmt"
)

type deviceServeOptions struct {
	socket, capabilities, worldBindings, worldDomain, worldEntity, bootID, deviceID string
	faultSchedule                                                                   faultSpecFlag
}

func parseDeviceServeOptions(args []string) (deviceServeOptions, error) {
	var options deviceServeOptions
	fs := flag.NewFlagSet("device serve", flag.ExitOnError)
	registerDeviceServeFlags(fs, &options)
	if err := fs.Parse(args); err != nil {
		return deviceServeOptions{}, fmt.Errorf("streamsim: %w", err)
	}
	if options.socket == "" {
		return deviceServeOptions{}, fmt.Errorf("device serve requires --socket")
	}
	if options.capabilities == "" {
		return deviceServeOptions{}, fmt.Errorf("device serve requires --capabilities")
	}
	return options, nil
}

func registerDeviceServeFlags(fs *flag.FlagSet, options *deviceServeOptions) {
	fs.StringVar(&options.socket, "socket", "", "Unix domain socket path to listen on (required)")
	fs.StringVar(&options.capabilities, "capabilities", "", "device capability catalog JSON path (required)")
	fs.StringVar(&options.worldBindings, "world", "", "deviceworld binding catalog JSON path (optional)")
	fs.StringVar(&options.worldDomain, "world-domain", "domains/cold-chain-transit.domain.json", "world domain JSON path when --world is set")
	fs.StringVar(&options.worldEntity, "world-entity", "", "world entity bound to device targets; defaults to the first entity")
	fs.StringVar(&options.bootID, "boot-id", "boot-A", "initial device boot identity")
	fs.StringVar(&options.deviceID, "device-id", "dev-01", "device identity")
	fs.Var(&options.faultSchedule, "fault", "deterministic fault name[@accepted-command-ordinal]; repeatable")
}
