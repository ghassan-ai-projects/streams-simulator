package host

import (
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

func (h *Host) deviceBirth(record map[string]any) {
	name := record["device"].(string)
	if h.quarantine(name, record["kind"].(string)) {
		return
	}
	d := &device{boot: record["boot_id"].(string), aliases: map[int64]string{}, values: map[string]float64{}, updated: map[string]int64{}}
	h.devices[name] = d
	for _, metric := range record["metrics"].([]any) {
		entry := metric.(map[string]any)
		d.aliases[domain.Int(entry["alias"])] = entry["name"].(string)
		h.sample(d, entry["name"].(string), entry["value"])
	}
}

func (h *Host) data(record map[string]any) {
	d := h.devices[record["device"].(string)]
	if d == nil || record["historical"] == true {
		return
	}
	for _, sample := range record["metrics"].([]any) {
		if entry := sample.(map[string]any); entry["quality"] == "good" {
			h.sample(d, d.aliases[domain.Int(entry["alias"])], entry["value"])
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
	if len(h.options.Inventory) == 0 || (listed && expected == kind) {
		return false
	}
	if !slices.Contains(h.report.Quarantined, name) {
		h.report.Quarantined = append(h.report.Quarantined, name)
	}
	return true
}
