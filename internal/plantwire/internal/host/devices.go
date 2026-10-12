package host

import (
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

const maxDevices = 4096

func (h *Host) deviceBirth(m domain.Message) {
	if h.quarantine(m.Device, m.Kind) {
		delete(h.devices, m.Device)
		return
	}
	d := &device{boot: m.BootID, aliases: map[int64]string{}, values: map[string]float64{}, updated: map[string]int64{}}
	h.devices[m.Device] = d
	for _, metric := range m.Metrics {
		d.aliases[metric.Alias] = metric.Name
		if metric.Quality == "good" {
			h.sample(d, metric.Name, metric.Value)
		}
	}
}

func (h *Host) data(m domain.Message) {
	d := h.devices[m.Device]
	if d == nil || m.Historical || m.BootID != d.boot {
		return
	}
	for _, metric := range m.Metrics {
		if metric.Quality == "good" {
			h.sample(d, d.aliases[metric.Alias], metric.Value)
		}
	}
}

func (h *Host) sample(d *device, metric string, raw any) {
	if value, numeric := number(raw); numeric {
		h.samples++
		d.values[metric], d.updated[metric] = value, h.samples
	}
}

func number(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case interface{ Float64() (float64, error) }:
		parsed, err := typed.Float64()
		return parsed, err == nil
	}
	return 0, false
}

func (h *Host) quarantine(name, kind string) bool {
	expected, listed := h.options.Inventory[name]
	commissioned := len(h.options.Inventory) == 0 || (listed && expected == kind)
	if commissioned && (len(h.devices) < maxDevices || h.devices[name] != nil) {
		return false
	}
	if len(h.report.Quarantined) < maxDevices && !slices.Contains(h.report.Quarantined, name) {
		h.report.Quarantined = append(h.report.Quarantined, name)
	}
	return true
}
