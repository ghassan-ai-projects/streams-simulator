package domain

import (
	"fmt"
	"maps"
	"math"
	"strconv"
)

func (r *Relay) HostAttached() Moves {
	r.stats.HostSessions++
	session := relayName + "-h" + strconv.Itoa(r.stats.HostSessions)
	r.host = hostSide{connected: true, session: session, out: NewTransmitter(session)}
	moves := Moves{}
	if r.gateway.admitted {
		r.birthUp(&moves)
	}
	return moves
}

func (r *Relay) HostUnreachable() Moves {
	moves := Moves{}
	if r.gateway.admitted && !r.gateway.answered {
		r.offline(&moves, nil)
	}
	return moves
}

func (r *Relay) HostLost() Moves {
	moves := Moves{}
	r.loseHost(&moves)
	return moves
}

func (r *Relay) loseHost(m *Moves) {
	if !r.host.connected {
		return
	}
	m.CloseHost = true
	r.lastUp = saturatingInt64(r.host.out.LastSeq())
	last := r.host.state
	r.host = hostSide{}
	m.Notes = append(m.Notes, connectionNote("host", "offline"))
	r.gateway.asked = false
	clear(r.gateway.born)
	if r.gateway.admitted {
		r.offline(m, last)
	}
}

func (r *Relay) offline(m *Moves, last map[string]any) {
	fields := relayState(false, "")
	if last != nil {
		fields = maps.Clone(last)
		fields["online"] = false
		delete(fields, "refusal")
	}
	r.gateway.answered = true
	r.transmitDown(m, Record{Type: "STATE", Lane: LaneControl, Fields: fields})
}

func (r *Relay) FromHost(in Inbound) Moves {
	moves := Moves{Trace: []Traced{{Direction: TraceHost, Record: in.Record}}}
	switch arrivalOf(&r.host.in, in) {
	case Duplicate:
		return r.dropped(moves)
	case Gap:
		r.loseHost(&moves)
		return moves
	}
	r.hostRecord(&moves, in)
	return moves
}

func (r *Relay) hostRecord(m *Moves, in Inbound) {
	in.Within(r.host.session, Version)
	if in.Invalid == nil && hostRecords[in.Type] && (r.host.shaken || in.Type == "STATE") {
		m.Trace[0].Admitted = true
	}
	switch {
	case !r.host.shaken:
		r.hostHandshake(m, in)
	case in.Invalid != nil || !hostRecords[in.Type]:
		r.stats.Dropped++
	default:
		r.deliverDown(m, in)
		r.settleEvidence(m, in)
	}
}

func (r *Relay) settleEvidence(m *Moves, in Inbound) {
	if in.Type == "HSYNC" {
		m.Trace[0].Admitted = m.syncRouted
	}
}

func (r *Relay) hostHandshake(m *Moves, in Inbound) {
	switch {
	case in.Invalid != nil || in.Type != "STATE":
		r.stats.Dropped++
	case !in.Online:
		r.hostRefused(m, in)
	default:
		r.hostAccepted(m, in)
	}
}

func (r *Relay) hostRefused(m *Moves, in Inbound) {
	r.stats.HostRefused = in.Refusal
	m.Err = fmt.Errorf("%w: %s", ErrRefused, in.Refusal)
	m.CloseHost, m.CloseGateway = true, true
	if r.gateway.admitted {
		r.transmitDown(m, stateRecord(content(in.Record)))
	}
	r.host = hostSide{}
}

func (r *Relay) hostAccepted(m *Moves, in Inbound) {
	if _, err := Negotiate(in.Message, relayProtocol); err != nil {
		r.loseHost(m)
		return
	}
	r.host.shaken = true
	r.stateDown(m, in)
}

func (r *Relay) stateDown(m *Moves, in Inbound) {
	first := r.host.state == nil
	r.noteBeat(in)
	r.stateToSim(m, first)
	if !r.gateway.admitted {
		return
	}
	rebirth := first && r.gateway.answered
	r.gateway.answered = true
	r.transmitDown(m, stateRecord(r.host.state))
	if rebirth {
		r.askRebirth(m)
	}
}

func (r *Relay) noteBeat(in Inbound) {
	r.host.state, r.host.silent = content(in.Record), 0
	r.host.interval = saturatingInt64(Seq(in.Record["heartbeat_interval_ms"]))
	r.host.beating = r.host.beating || Seq(in.Record["heartbeat"]) > 0
}

func (r *Relay) deliverDown(m *Moves, in Inbound) {
	switch {
	case in.Type == "STATE":
		r.stateDown(m, in)
	case r.sim != nil && routedDown[in.Type]:
		r.routeDown(m, in)
	case r.sim != nil && in.Type == "NCMD" && in.Command == "rebirth":
		r.forwardRebirth(m, in)
	default:
		r.forwardDown(m, in)
	}
}

func (r *Relay) forwardDown(m *Moves, in Inbound) {
	fields, mapped := r.downward(in)
	if !r.gateway.admitted {
		r.refuseCommand(m, in, "not_ready")
		return
	}
	if !mapped {
		r.stats.Dropped++
		return
	}
	r.transmitDown(m, Record{Type: in.Type, Lane: text(in.Record, "lane"), Fields: fields})
	m.syncRouted = m.syncRouted || in.Type == "HSYNC"
}

func (r *Relay) forwardRebirth(m *Moves, in Inbound) {
	r.forwardDown(m, in)
	r.askSimRebirth(m)
}

func (r *Relay) downward(in Inbound) (map[string]any, bool) {
	fields := content(in.Record)
	if in.Type != "HSYNC" {
		return fields, true
	}
	through, mapped := r.host.table.gatewayFor(in.ThroughSeq)
	delete(fields, "command_evidence")
	fields["through_seq"] = through
	return fields, mapped
}

func (r *Relay) TickBy(elapsedMS int64) Moves {
	moves := r.gatewayTick(elapsedMS)
	if !r.host.shaken || !r.host.beating {
		return moves
	}
	r.host.silent += elapsedMS
	if r.host.silent >= heartbeatsMissed*r.host.interval {
		r.loseHost(&moves)
	}
	return moves
}

func (r *Relay) Shutdown() Moves {
	moves := Moves{}
	if r.host.shaken {
		r.transmitUp(&moves, NodeDeath(r.birthBD, "shutdown"))
	}
	return moves
}

func saturatingInt64(v uint64) int64 {
	if v > math.MaxInt64 {
		return math.MaxInt64
	}
	return int64(v)
}
