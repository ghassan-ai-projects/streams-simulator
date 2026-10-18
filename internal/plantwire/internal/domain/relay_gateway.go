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
	moves := Moves{Trace: []Traced{{Direction: TraceGateway, Record: in.Record}}}
	switch arrivalOf(&r.gateway.in, in) {
	case Duplicate:
		return r.dropped(moves)
	case Gap:
		r.gatewayGap(&moves)
	}
	r.gatewayRecord(&moves, in)
	return moves
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
	refusal := r.admission.Refusal(in.Message, false)
	switch {
	case in.Invalid != nil || in.Type != "NBIRTH":
		r.stats.Dropped++
	case refusal != "":
		r.refuseGateway(m, in, refusal)
	default:
		r.takeBirth(in)
		r.afterBirth(m)
	}
}

func (r *Relay) refuseGateway(m *Moves, in Inbound, refusal string) {
	r.stats.Refused, r.stats.Reason = refusal, r.admission.Reason(refusal, in.Message)
	out := NewTransmitter(in.SessionID)
	out.Queue(Record{Type: "STATE", Lane: LaneControl, Fields: relayState(false, refusal)})
	sent := out.Drain()[0]
	m.Down = append(m.Down, sent)
	m.saw(TraceRelay, sent)
	m.CloseGateway = true
}

func (r *Relay) takeBirth(in Inbound) {
	r.birth = in
	r.host.table = seqTable{}
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
	case r.host.shaken:
		r.forwardUp(m, in)
	default:
		r.stats.Dropped++
	}
}

func (r *Relay) rebirth(m *Moves, in Inbound) {
	if r.admission.Refusal(in.Message, false) != "" {
		r.stats.Dropped++
		return
	}
	r.birth, r.gateway.asked = in, false
	if r.host.shaken {
		r.birthUp(m)
	}
}

func (r *Relay) birthUp(m *Moves) {
	fields := content(r.birth.Record)
	fields["bd_seq"], fields["prior_last_seq"] = r.bdSeq, r.lastUp
	fields["protocol_max"] = min(r.birth.ProtocolMax, relayProtocol)
	r.birthBD = r.bdSeq
	r.bdSeq++
	seq := r.transmitUp(m, Record{Type: "NBIRTH", Lane: LaneControl, Fields: fields})
	r.host.table.add(r.birth.Seq, seq)
	clear(r.gateway.born)
}

func (r *Relay) forwardUp(m *Moves, in Inbound) {
	fields, mapped := r.upward(in)
	if !mapped {
		r.stats.Dropped++
		return
	}
	seq := r.transmitUp(m, Record{Type: in.Type, Lane: text(in.Record, "lane"), Fields: fields})
	r.host.table.add(in.Seq, seq)
	r.track(in)
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

func (r *Relay) track(in Inbound) {
	switch in.Type {
	case "DBIRTH":
		r.gateway.born[in.Device] = in.BootID
	case "DDEATH":
		delete(r.gateway.born, in.Device)
	}
}

func (r *Relay) gatewayEnded(m *Moves) {
	r.diedUp(m)
	r.gateway.admitted = false
}

func (r *Relay) GatewayLost() Moves {
	moves := Moves{}
	r.diedUp(&moves)
	r.gateway.admitted = false
	if r.host.connected && !r.host.shaken {
		r.loseHost(&moves)
	}
	r.gateway = gatewaySide{born: map[string]string{}}
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
