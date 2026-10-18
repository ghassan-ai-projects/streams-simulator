package domain

const gatewayIdleMS int64 = 30_000

func connectionNote(side, state string) Note {
	return Note{Type:"connection", Fields:map[string]any{"side":side,"state":state}}
}

func (r *Relay) gatewayTick() Moves {
	if !r.gateway.admitted {
		return Moves{}
	}
	r.gateway.silent += r.tickMS
	if r.gateway.silent < gatewayIdleMS {
		return Moves{}
	}
	moves := r.GatewayLost()
	moves.CloseGateway = true
	return moves
}
