package domain

type Origin int

const (
	OriginUnknown Origin = iota
	OriginGateway
	OriginSimulated
)

var routedDown = map[string]bool{"DCMD": true, "DQUERY": true, "HSYNC": true}

func (r *Relay) OriginOf(device string) Origin {
	switch {
	case r.gatewayDevices[device]:
		return OriginGateway
	case r.sim != nil && r.sim.devices[device]:
		return OriginSimulated
	}
	return OriginUnknown
}

func (r *Relay) routeDown(m *Moves, in Inbound) {
	if in.Type == "HSYNC" {
		r.routeSync(m, in)
		return
	}
	r.routeToDevice(m, in, r.OriginOf(in.Device))
}

func (r *Relay) routeToDevice(m *Moves, in Inbound, origin Origin) {
	r.routed.noteCommand(in, origin)
	switch {
	case origin == OriginGateway:
		r.forwardDown(m, in)
	case origin == OriginSimulated && r.sim.lost:
		r.refuseCommand(m, in, "not_ready")
	case origin == OriginSimulated:
		r.sendSim(m, Record{Type: in.Type, Lane: text(in.Record, "lane"), Fields: content(in.Record)})
	default:
		r.refuseUnknown(m, in)
	}
}

func (r *Relay) refuseUnknown(m *Moves, in Inbound) {
	if r.gatewayLostWith(in.Device) {
		r.refuseCommand(m, in, "not_ready")
		return
	}
	r.refuseCommand(m, in, "unknown_device")
}

func (r *Relay) gatewayLostWith(device string) bool {
	return !r.gateway.admitted && devicesOf(r.birth.Boards)[device]
}

func (r *Relay) refuseCommand(m *Moves, in Inbound, reason string) {
	if in.Type != "DCMD" || in.CommandID == "" || WireID(in.CommandID) != in.CommandID {
		r.stats.Dropped++
		return
	}
	if reason == "unknown_device" {
		r.stats.UnknownDevice++
	}
	r.transmitUp(m, Receipt(in.CommandID, "refused", reason, "none"))
}

func (r *Relay) routeSync(m *Moves, in Inbound) {
	simulated := r.host.simSyncs.answer(in.ThroughSeq)
	hardware := r.host.gatewaySyncs.answer(in.ThroughSeq)
	if simulated {
		r.syncToSim(m, in)
	}
	if hardware {
		r.forwardDown(m, in)
	}
	if !simulated && !hardware {
		r.stats.Dropped++
	}
}

func (r *Relay) syncToSim(m *Moves, in Inbound) {
	through, mapped := r.host.simTable.gatewayFor(in.ThroughSeq)
	if !mapped {
		r.stats.Dropped++
		return
	}
	fields := content(in.Record)
	delete(fields, "command_evidence")
	fields["through_seq"] = through
	r.sendSim(m, Record{Type: "HSYNC", Lane: text(in.Record, "lane"), Fields: fields})
	m.syncRouted = true
}
