package domain

import (
	"maps"
	"slices"
)

type SimMoves struct {
	Up    []map[string]any
	Trace []Traced
	Notes []Note
}

func (r *Relay) FromSim(in Inbound) SimMoves {
	moves := Moves{}
	r.simRecord(&moves, in)
	return SimMoves{Up: moves.Up, Trace: moves.Trace, Notes: moves.Notes}
}

func (r *Relay) simRecord(m *Moves, in Inbound) {
	switch {
	case r.sim == nil || in.Invalid != nil || !nodeRecords[in.Type] || arrivalOf(&r.sim.in, in) == Duplicate:
		r.stats.Dropped++
	case in.Type == "NBIRTH":
		r.simBirth(in)
	case in.Type == "NDEATH":
		r.simEnded(m)
	case !r.host.shaken || !r.host.birthed || (in.Device != "" && !r.sim.devices[in.Device]):
		r.stats.Dropped++
	default:
		r.simCarry(m, in)
	}
}

func (r *Relay) simCarry(m *Moves, in Inbound) {
	if r.forgedAnswer(in, OriginSimulated) {
		r.forged()
		return
	}
	r.simUp(m, in)
}

func (r *Relay) SimLost() SimMoves {
	moves := Moves{}
	if r.sim != nil && !r.sim.lost {
		r.sim.lost = true
		r.stats.SimulatedLost++
		r.simEnded(&moves)
	}
	return SimMoves{Up: moves.Up, Trace: moves.Trace, Notes: moves.Notes}
}

func (r *Relay) simBirth(in Inbound) {
	if r.host.birthed {
		r.host.simTable.add(in.Seq, r.host.birthSeq)
	}
}

func (r *Relay) simUp(m *Moves, in Inbound) {
	fields, mapped := r.simUpward(in)
	if !mapped {
		r.stats.Dropped++
		return
	}
	seq := r.transmitUp(m, Record{Type: in.Type, Lane: text(in.Record, "lane"), Fields: fields})
	r.host.simTable.add(in.Seq, seq)
	r.stats.FromSimulated++
	r.trackSim(in, seq)
	r.noteOrigin(m, in, EnvironmentSimulated)
}

func (r *Relay) simUpward(in Inbound) (map[string]any, bool) {
	fields := content(in.Record)
	if in.Type != "NSYNC" {
		return fields, true
	}
	through, mapped := r.host.simTable.upFor(in.ThroughSeq)
	fields["through_seq"] = through
	return fields, mapped
}

func (r *Relay) trackSim(in Inbound, seq uint64) {
	switch in.Type {
	case "DBIRTH":
		r.sim.born[in.Device] = in.BootID
	case "DDEATH":
		delete(r.sim.born, in.Device)
	case "NSYNC":
		r.host.simSyncs.add(seq)
	}
}

func (r *Relay) simEnded(m *Moves) {
	if r.host.shaken && r.host.birthed {
		for _, device := range slices.Sorted(maps.Keys(r.sim.born)) {
			r.transmitUp(m, simDeath(device, r.sim.born[device], r.host.out.LastSeq()))
		}
	}
	clear(r.sim.born)
}

func simDeath(device, boot string, lastSeq uint64) Record {
	return Record{Type: "DDEATH", Lane: LaneControl,
		Fields: map[string]any{"device": device, "boot_id": boot, "reason": "offline", "last_seq": lastSeq}}
}

func (r *Relay) stateToSim(m *Moves, first bool) {
	if r.sim == nil || !r.host.birthed {
		return
	}
	r.sendSim(m, stateRecord(r.host.state))
	if first && r.sim.started {
		r.sendSim(m, simRebirth())
	}
	r.sim.started = true
}

func (r *Relay) askSimRebirth(m *Moves) {
	if r.sim != nil && r.sim.started && r.host.shaken {
		r.sendSim(m, simRebirth())
	}
}

func simRebirth() Record {
	return Record{Type: "NCMD", Lane: LaneControl, Fields: map[string]any{"command": "rebirth"}}
}

func (r *Relay) sendSim(m *Moves, record Record) {
	r.sim.out.Queue(record)
	m.ToSim = append(m.ToSim, r.sim.out.Drain()[0])
	r.stats.ToSimulated++
}
