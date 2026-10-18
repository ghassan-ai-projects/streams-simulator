package domain

import "slices"

const routedWindow = 4096

type routeLedger struct {
	origin map[string]Origin
	order  []string
}

func (l *routeLedger) noteCommand(in Inbound, origin Origin) {
	if in.Type != "DCMD" || in.CommandID == "" || origin == OriginUnknown {
		return
	}
	if l.origin == nil {
		l.origin = map[string]Origin{}
	}
	if _, known := l.origin[in.CommandID]; !known {
		l.order = append(l.order, in.CommandID)
	}
	l.origin[in.CommandID] = origin
	l.trim()
}

func (l *routeLedger) trim() {
	for len(l.order) > routedWindow {
		delete(l.origin, l.order[0])
		l.order = l.order[1:]
	}
}

func (l *routeLedger) routedTo(commandID string, origin Origin) bool {
	return l.origin[commandID] == origin
}

func answersCommand(kind string) bool { return kind == "DRECEIPT" || kind == "DRESULT" }

func (r *Relay) forgedAnswer(in Inbound, side Origin) bool {
	return r.sim != nil && answersCommand(in.Type) && !r.routed.routedTo(in.CommandID, side)
}

func (r *Relay) forgedFromGateway(in Inbound) bool {
	return r.forgedAnswer(in, OriginGateway) || (in.Type == "DBIRTH" && !r.bornAsAnnounced(in))
}

func (r *Relay) bornAsAnnounced(in Inbound) bool {
	board := text(in.Record, "board")
	onBoard := slices.ContainsFunc(r.birth.Boards, func(b Board) bool { return b.Board == board && slices.Contains(b.Members, in.Device) })
	return onBoard && slices.ContainsFunc(r.birth.Kinds, func(entry any) bool { return kindNamed(entry) == in.Kind })
}

func (r *Relay) forged() {
	r.stats.Dropped++
	r.stats.Forged++
}
