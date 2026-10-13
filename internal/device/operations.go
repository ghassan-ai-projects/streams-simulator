package device

import (
	"io"

	layer "github.com/ghassan-ai-projects/streams-simulator/internal/device/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/device/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/device/internal/uds"
)

func LoadCapabilities(data []byte) (*Capabilities, error) {
	return layer.LoadCapabilities(data)
}

func LoadCapabilitiesFile(path string) (*Capabilities, error) {
	return files.LoadCapabilities(path)
}

func ValidateFaultSchedule(schedule []FaultInjection) error {
	return layer.ValidateFaultSchedule(schedule)
}

func ParseFaultSpec(spec string) (FaultInjection, error) {
	return layer.ParseFaultSpec(spec)
}

func Serve(stderr io.Writer, socket string, d *Device, banner string) error {
	return uds.Serve(stderr, socket, d.device, banner)
}

func (d *Device) ApplyCommand(command map[string]any) Outcome {
	return d.device.ApplyCommand(command)
}

func (d *Device) State() map[string]any {
	return d.device.State()
}
