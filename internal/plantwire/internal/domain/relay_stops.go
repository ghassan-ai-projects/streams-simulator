package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
)

func (r *Relay) issueStops(m *Moves, orders []StopOrder) {
	for _, order := range orders {
		r.issueStop(m, order)
	}
}

func (r *Relay) issueStop(m *Moves, order StopOrder) {
	origin, boot := r.OriginOf(order.Device), r.stopBoot(order.Device)
	switch {
	case boot == "":
		m.Interrupt = append(m.Interrupt, r.stopFailure(order, stopNotBorn))
	case origin == OriginSimulated && r.sim.lost:
		m.Interrupt = append(m.Interrupt, r.stopFailure(order, stopSimulatorGone))
	default:
		r.sendStop(m, order, origin, boot)
	}
}

func (r *Relay) sendStop(m *Moves, order StopOrder, origin Origin, boot string) {
	record := r.stopRecord(order, origin, boot)
	m.Interrupt = append(m.Interrupt, EffectEntry(text(record.Fields, "command_id"), order.Device, "ordered", ClockWall, r.ints.wallNS))
	if origin == OriginGateway {
		r.transmitDown(m, record)
		return
	}
	r.sendSim(m, record)
}

func (r *Relay) bootOf(device string) string {
	if boot, born := r.gateway.born[device]; born && r.gatewayDevices[device] {
		return boot
	}
	if r.sim != nil && r.sim.devices[device] {
		return r.sim.born[device]
	}
	return ""
}

func (i *relayInterrupts) issue(device string) string {
	if len(i.order) >= maxIssuedStops {
		delete(i.issued, i.order[0])
		i.order = i.order[1:]
	}
	i.orders++
	id := fmt.Sprintf("%s%d", stopIDPrefix, i.orders)
	i.issued[id] = device
	i.order = append(i.order, id)
	return id
}

func (r *Relay) stopFailure(order StopOrder, why string) InterruptEntry {
	return EffectEntry("none", order.Device, why, ClockWall, r.ints.wallNS)
}

func (r *Relay) stopRecord(order StopOrder, origin Origin, boot string) Record {
	_, now := r.ints.clockOf(origin)
	id := r.ints.issue(order.Device)
	key, _ := canonical.Digest(map[string]any{"node_stop": id, "device": order.Device, "command": order.Command, "args": map[string]any{}, "expected_boot_id": boot})
	return Record{Type: "DCMD", Lane: LaneControl, Fields: map[string]any{"command_id": id, "idempotency_key": key,
		"device": order.Device, "command": order.Command, "args": map[string]any{}, "expected_boot_id": boot, "deadline_ns": now + stopDeadlineNS,
		"not_before_ns": int64(0), "authority_epoch": r.hostEpoch(), "policy_digest": stopPolicy,
		"fallback_on_lease_loss": "safe_off"}}
}
