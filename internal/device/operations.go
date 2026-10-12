package device

import (
	"io"
	"net"

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

// Listen serves the device over a unix-domain socket at path, one gateway
// connection at a time, and returns the listener to close.
func Listen(path string, d *Device) (*net.UnixListener, error) {
	return uds.Listen(path, d.device)
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

// SetFaults sets the injectable protocol fault state.
func (d *Device) SetFaults(faults Faults) {
	d.device.SetFaults(faults)
}
