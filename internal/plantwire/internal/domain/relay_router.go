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
	switch r.OriginOf(in.Device) {
	case OriginGateway:
		r.forwardDown(m, in)
	case OriginSimulated:
		r.sendSim(m, Record{Type: in.Type, Lane: text(in.Record, "lane"), Fields: content(in.Record)})
	default:
		r.refuseUnknown(m, in)
	}
}

func (r *Relay) refuseUnknown(m *Moves, in Inbound) {
	if in.Type != "DCMD" || in.CommandID == "" || WireID(in.CommandID) != in.CommandID {
		r.stats.Dropped++
		return
	}
	r.stats.UnknownDevice++
	r.transmitUp(m, Receipt(in.CommandID, "refused", "unknown_device", "none"))
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
	fields["through_seq"] = through
	r.sendSim(m, Record{Type: "HSYNC", Lane: text(in.Record, "lane"), Fields: fields})
}
