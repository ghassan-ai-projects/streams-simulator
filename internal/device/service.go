package device

import layer "github.com/ghassan-ai-projects/streams-simulator/internal/device/internal/domain"

// Device is a wire-faithful serial device emulator. It is safe for use by one
// gateway connection at a time.
type Device struct {
	device *layer.Device
}

// New builds a device from cfg, applying deterministic defaults. A device
// with no capabilities declares no targets and rejects every command.
func New(cfg Config) *Device {
	return &Device{device: layer.New(cfg)}
}
