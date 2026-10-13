package domain

func (w *World) latestStateFault(entity, state string, at int64) (int64, bool) {
	last := int64(-1)
	any := false
	for _, fault := range w.faultsOn(state) {
		if fault.entity != entity || !fault.activeAt(at) {
			continue
		}
		any = true
		if fault.onsetNS > last {
			last = fault.onsetNS
		}
	}
	return last, any
}

func (w *World) faultsOn(state string) []*activeFault {
	if len(w.faultsByState) == 0 {
		return nil
	}
	return w.faultsByState[state]
}

func (w *World) applyStateFaults(value float64, entity, state string, at int64) float64 {
	kind := w.entities[entity].kind
	for _, fault := range w.faultsByState[state] {
		if fault.entity != entity || !fault.activeAt(at) {
			continue
		}
		add, mult := fault.contributionAt(state, at, w.substream(faultSubstream(fault, state)), dtFor(kind, state))
		if mult != 0 {
			value = value*(1+mult) + add
		} else {
			value += add
		}
	}
	return value
}

type kickDrivers struct {
	last       int64
	any        bool
	assignment *kick
}

func latestKickDrivers(kicks []*kick, at int64) kickDrivers {
	drivers := kickDrivers{last: -1}
	for _, kick := range kicks {
		if at <= kick.startNS {
			continue
		}
		drivers.record(kick)
	}
	return drivers
}

func drivenValue(value float64, kicks []*kick, before, at int64) float64 {
	drivers := latestKickDrivers(kicks, before)
	if drivers.assignment == nil {
		return applyAdditiveKicks(value, kicks, before, at)
	}
	return applyAssignedKicks(value, kicks, drivers.assignment, before, at)
}

func applyAssignedKicks(value float64, kicks []*kick, assignment *kick, before, at int64) float64 {
	driven := assignment.assignedFrom(drivenValue(value, kicks, assignment.startNS, at), at)
	for _, kick := range kicks {
		if kick.assign || kick.startNS <= assignment.startNS || before <= kick.startNS {
			continue
		}
		driven += kick.valueAt(at)
	}
	return driven
}

func applyAdditiveKicks(value float64, kicks []*kick, before, at int64) float64 {
	for _, kick := range kicks {
		if before > kick.startNS {
			value += kick.valueAt(at)
		}
	}
	return value
}

func (drivers *kickDrivers) record(kick *kick) {
	drivers.any = true
	if kick.startNS > drivers.last {
		drivers.last = kick.startNS
	}
	if kick.assign && (drivers.assignment == nil || kick.startNS >= drivers.assignment.startNS) {
		drivers.assignment = kick
	}
}
