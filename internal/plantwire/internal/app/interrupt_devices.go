package app

import wire "github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"

func (d *devices) interruptInventory() []wire.InterruptDevice {
	var inventory []wire.InterruptDevice
	for _, id := range d.announced() {
		inventory = append(inventory, wire.InterruptDevice{Device: d.wireOf(id), Board: d.board(id), Clock: wire.ClockWorld, SafeStop: d.safeStopOf(id)})
	}
	return inventory
}

func (d *devices) safeStopOf(id string) string {
	kind, _ := d.kindOf(id)
	if kind == nil {
		return ""
	}
	for i := range kind.Effectors {
		if kind.Effectors[i].SafeStop {
			return kind.Effectors[i].Name
		}
	}
	return ""
}
