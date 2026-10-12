package device

import (
	"io"

	layer "github.com/ghassan-ai-projects/streams-simulator/internal/device/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/device/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/device/internal/uds"
)

// LoadCapabilities parses and validates a capability catalog document; the
// catalog is data, never hard-coded.
func LoadCapabilities(data []byte) (*Capabilities, error) {
	return layer.LoadCapabilities(data)
}

// LoadCapabilitiesFile reads and validates the capability catalog at path.
func LoadCapabilitiesFile(path string) (*Capabilities, error) {
	return files.LoadCapabilities(path)
}

// ValidateFaultSchedule rejects a schedule the device could not carry out.
func ValidateFaultSchedule(schedule []FaultInjection) error {
	return layer.ValidateFaultSchedule(schedule)
}

// ParseFaultSpec parses one `ordinal=fault` schedule entry.
func ParseFaultSpec(spec string) (FaultInjection, error) {
	return layer.ParseFaultSpec(spec)
}

// Serve serves d on the unix socket at path, announces the banner on stderr,
// and returns after the process is interrupted and the listener is closed.
func Serve(stderr io.Writer, socket string, d *Device, banner string) error {
	return uds.Serve(stderr, socket, d.device, banner)
}

// ApplyCommand applies one decoded command record and returns its outcome.
func (d *Device) ApplyCommand(command map[string]any) Outcome {
	return d.device.ApplyCommand(command)
}

// State returns the current device.state record.
func (d *Device) State() map[string]any {
	return d.device.State()
}
