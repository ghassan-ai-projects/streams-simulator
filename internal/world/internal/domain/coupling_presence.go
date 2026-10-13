package domain

func (e *Entity) lastAliveAt(t int64) int64 {
	if e.alive {
		return t
	}
	return min(t, e.retiredNS)
}

func (e *Entity) finishedFor(source *Entity) bool {
	return !e.alive && source.readersAt >= e.retiredNS
}

func (e *Entity) presentAt(t int64) bool {
	return e.BornNS <= t && (e.alive || t <= e.retiredNS)
}
