package domain

const NoteConnection = "connection"

const (
	gatewayIdleMS  int64 = 30_000
	rebirthGraceMS int64 = 5_000
)

func connectionNote(side, state string) Note {
	return Note{Type: NoteConnection, Fields: map[string]any{"side": side, "state": state}}
}

func (r *Relay) gatewayTick(elapsedMS int64) Moves {
	if !r.gateway.admitted || !r.host.shaken {
		return Moves{}
	}
	r.gateway.silent += elapsedMS
	r.host.present += elapsedMS
	if r.gateway.silent < gatewayIdleMS || r.host.present < rebirthGraceMS {
		return Moves{}
	}
	moves := r.GatewayLost()
	moves.CloseGateway = true
	return moves
}
