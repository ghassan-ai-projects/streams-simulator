package domain

import (
	"maps"
	"slices"
)

func (r *Relay) GatewayAttached() {
	r.stats.GatewaySessions++
	r.gateway = gatewaySide{born: map[string]string{}}
}

func (r *Relay) FromGateway(in Inbound) Moves {
	admitted, session := r.gateway.admitted, r.gateway.session
	moves := Moves{Trace: []Traced{{Direction: TraceGateway, Record: in.Record}}}
	switch arrivalOf(&r.gateway.in, in) {
	case Duplicate:
		return r.dropped(moves)
	case Gap:
		r.gatewayGap(&moves)
	}
	r.gatewayRecord(&moves, in)
	moves.Trace[0].Field = r.belongsToSession(in, admitted, session)
	return moves
}

func (r *Relay) belongsToSession(in Inbound, admitted bool, session string) bool {
	if in.Type == "NBIRTH" {
		return in.Invalid == nil && r.gateway.admitted && r.birth.SessionID == in.SessionID && r.birth.Seq == in.Seq
	}
	in.Within(session, Version)
	return admitted && in.Invalid == nil && nodeRecords[in.Type]
}

func (r *Relay) gatewayGap(m *Moves) {
	if r.gateway.admitted {
		r.askRebirth(m)
	}
}

func (r *Relay) askRebirth(m *Moves) {
	if r.gateway.asked {
		return
	}
	r.gateway.asked = true
	r.transmitDown(m, Record{Type: "NCMD", Lane: LaneControl, Fields: map[string]any{"command": "rebirth"}})
}

func (r *Relay) gatewayRecord(m *Moves, in Inbound) {
	if r.gateway.admitted {
		r.admittedRecord(m, in)
		return
	}
	r.admit(m, in)
}

func (r *Relay) admit(m *Moves, in Inbound) {
	refusal, reason := r.gatewayRefusal(in)
	switch {
	case in.Invalid != nil || in.Type != "NBIRTH":
		r.stats.Dropped++
	case refusal != "":
		r.refuseGateway(m, in, refusal, reason)
		r.noteBinding(m, in, refusal, reason)
	default:
		r.takeBirth(in)
		r.noteBinding(m, in, "", "")
		r.afterBirth(m)
	}
}

func (r *Relay) refuseGateway(m *Moves, in Inbound, refusal, reason string) {
	r.stats.Refused, r.stats.Reason = refusal, reason
	out := NewTransmitter(in.SessionID)
	out.Queue(Record{Type: "STATE", Lane: LaneControl, Fields: relayState(false, refusal)})
	sent := out.Drain()[0]
	m.Down = append(m.Down, sent)
	m.saw(TraceRelay, sent)
	m.CloseGateway = true
}

func (r *Relay) takeBirth(in Inbound) {
	r.birth, r.gatewayDevices = in, devicesOf(in.Boards)
	r.host.table, r.host.gatewaySyncs = seqTable{}, syncLedger{}
	r.gateway.admitted, r.gateway.answered, r.gateway.asked = true, false, false
	r.gateway.session, r.gateway.out = in.SessionID, NewTransmitter(in.SessionID)
	clear(r.gateway.born)
}

func (r *Relay) afterBirth(m *Moves) {
	if r.host.shaken {
		r.answerGateway(m)
		r.birthUp(m)
	}
}

func (r *Relay) answerGateway(m *Moves) {
	r.gateway.answered = true
	r.transmitDown(m, stateRecord(r.host.state))
}

func (r *Relay) admittedRecord(m *Moves, in Inbound) {
	in.Within(r.gateway.session, Version)
	switch {
	case in.Invalid != nil || !nodeRecords[in.Type]:
		r.stats.Dropped++
	case in.Type == "NBIRTH":
		r.rebirth(m, in)
	default:
		r.forwardOrEnd(m, in)
	}
}

func (r *Relay) forwardOrEnd(m *Moves, in Inbound) {
	switch {
	case in.Type == "NDEATH":
		r.gatewayEnded(m)
	case r.strayDevice(in):
		r.stats.Dropped++
	case r.forgedFromGateway(in):
		r.forged()
	case r.host.shaken:
		r.forwardUp(m, in)
	default:
		r.stats.Dropped++
	}
}

func (r *Relay) rebirth(m *Moves, in Inbound) {
	if refusal, reason := r.gatewayRefusal(in); refusal != "" {
		r.stats.Dropped++
		r.noteBinding(m, in, refusal, reason)
		return
	}
	r.birth, r.gateway.asked, r.gatewayDevices = in, false, devicesOf(in.Boards)
	if r.host.shaken {
		r.birthUp(m)
	}
}

func (r *Relay) birthUp(m *Moves) {
	fields := r.birthContent()
	fields["bd_seq"], fields["prior_last_seq"] = r.bdSeq, r.lastUp
	fields["protocol_max"] = min(r.birth.ProtocolMax, relayProtocol)
	r.birthBD = r.bdSeq
	r.bdSeq++
	seq := r.transmitUp(m, Record{Type: "NBIRTH", Lane: LaneControl, Fields: fields})
	r.host.table.add(r.birth.Seq, seq)
	r.markBirthed(seq)
	r.askSimRebirth(m)
}

func (r *Relay) forwardUp(m *Moves, in Inbound) {
	fields, mapped := r.upward(in)
	if !mapped {
		r.stats.Dropped++
		return
	}
	seq := r.transmitUp(m, Record{Type: in.Type, Lane: text(in.Record, "lane"), Fields: fields})
	r.host.table.add(in.Seq, seq)
	r.track(in, seq)
	r.noteMisbound(m, in)
	r.noteOrigin(m, in, EnvironmentHardware)
}

func (r *Relay) upward(in Inbound) (map[string]any, bool) {
	fields := content(in.Record)
	if in.Type != "NSYNC" {
		return fields, true
	}
	through, mapped := r.host.table.upFor(in.ThroughSeq)
	fields["through_seq"] = through
	return fields, mapped
}

func (r *Relay) track(in Inbound, seq uint64) {
	switch in.Type {
	case "DBIRTH":
		r.gateway.born[in.Device] = in.BootID
	case "DDEATH":
		delete(r.gateway.born, in.Device)
	case "NSYNC":
		r.host.gatewaySyncs.add(seq)
	}
}

func (r *Relay) gatewayEnded(m *Moves) {
	r.diedUp(m)
	r.clearGateway()
}

func (r *Relay) clearGateway() {
	r.gateway = gatewaySide{born: map[string]string{}}
	r.gatewayDevices = nil
	r.host.gatewaySyncs = syncLedger{}
}

func (r *Relay) GatewayLost() Moves {
	moves := Moves{}
	r.diedUp(&moves)
	r.gateway.admitted = false
	if r.host.connected && !r.host.shaken {
		r.loseHost(&moves)
	}
	r.clearGateway()
	return moves
}

func (r *Relay) diedUp(m *Moves) {
	if r.host.shaken {
		for _, device := range slices.Sorted(maps.Keys(r.gateway.born)) {
			r.transmitUp(m, deviceDeath(device, r.gateway.born[device], r.host.out.LastSeq()))
		}
	}
	clear(r.gateway.born)
}
