package device

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const (
	// FaultAckLost withholds a receipt after an otherwise accepted command.
	FaultAckLost = "ack_lost"
	// FaultStuck acknowledges a command without applying it to the plant.
	FaultStuck = "stuck"
	// FaultReboot changes boot identity after the selected command.
	FaultReboot = "reboot"
	// FaultDuplicate makes the device receive the selected command twice.
	FaultDuplicate = "duplicate"
	// FaultStale forces a command that passed normal admission to be not_ready.
	FaultStale = "stale"
	// FaultExpired forces a command that passed normal admission to be expired.
	FaultExpired = "expired"
	// FaultDisconnect closes the current gateway connection after the command.
	FaultDisconnect = "disconnect"
)

// FaultInjection schedules one deterministic protocol fault at an admission
// ordinal. The ordinal counts commands that pass the normal boot, freshness,
// target, operation, and bounds checks, including a command deliberately
// changed to stale or expired by this schedule.
type FaultInjection struct {
	Name            string
	AcceptedCommand int
}

// ValidateFaultSchedule rejects unknown fault names and invalid command
// ordinals before a device starts listening.
func ValidateFaultSchedule(schedule []FaultInjection) error {
	for _, injection := range schedule {
		if !knownFault(injection.Name) {
			return fmt.Errorf("device: unknown fault %q", injection.Name)
		}
		if injection.AcceptedCommand < 1 {
			return fmt.Errorf("device: fault %q requires a positive accepted command ordinal", injection.Name)
		}
	}
	return nil
}

func knownFault(name string) bool {
	switch name {
	case FaultAckLost, FaultStuck, FaultReboot, FaultDuplicate, FaultStale, FaultExpired, FaultDisconnect:
		return true
	default:
		return false
	}
}

// ParseFaultSpec parses the CLI form name[@n], where n is the one-based
// accepted-command ordinal. Without @n the fault applies to the first
// admission-valid command.
func ParseFaultSpec(spec string) (FaultInjection, error) {
	name, ordinalText, hasOrdinal := strings.Cut(spec, "@")
	if name == "" || strings.Contains(ordinalText, "@") {
		return FaultInjection{}, fmt.Errorf("device: fault must be name[@accepted-command], got %q", spec)
	}
	ordinal := 1
	if hasOrdinal {
		parsed, err := strconv.Atoi(ordinalText)
		if err != nil {
			return FaultInjection{}, fmt.Errorf("device: fault %q has invalid accepted command ordinal: %w", spec, err)
		}
		ordinal = parsed
	}
	injection := FaultInjection{Name: name, AcceptedCommand: ordinal}
	if err := ValidateFaultSchedule([]FaultInjection{injection}); err != nil {
		return FaultInjection{}, err
	}
	return injection, nil
}

func faultNames(schedule []FaultInjection) map[int][]string {
	byOrdinal := make(map[int][]string)
	for _, injection := range schedule {
		byOrdinal[injection.AcceptedCommand] = append(byOrdinal[injection.AcceptedCommand], injection.Name)
	}
	for ordinal := range byOrdinal {
		sort.Strings(byOrdinal[ordinal])
	}
	return byOrdinal
}
