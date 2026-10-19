package domain

import (
	"maps"
	"slices"
)

func (i *relayInterrupts) remembered(current []InterruptDevice) []InterruptDevice {
	all, present := maps.Clone(i.known), map[string]bool{}
	for _, d := range current {
		all[d.Device], present[d.Device] = d, true
		if _, kept := i.known[d.Device]; kept || len(i.known) < maxInterruptDevices {
			i.known[d.Device] = d
		}
	}
	return listedDevices(all, present)
}

func listedDevices(all map[string]InterruptDevice, present map[string]bool) []InterruptDevice {
	devices := make([]InterruptDevice, 0, len(all))
	for _, device := range slices.Sorted(maps.Keys(all)) {
		d := all[device]
		d.Unreachable = d.Unreachable || !present[device]
		devices = append(devices, d)
	}
	return devices
}
