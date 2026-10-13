package device

import layer "github.com/ghassan-ai-projects/streams-simulator/internal/device/internal/domain"

type Device struct {
	device *layer.Device
}

func New(cfg Config) *Device {
	return &Device{device: layer.New(cfg)}
}
