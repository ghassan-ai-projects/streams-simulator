package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

type storeGroup struct {
	names    []string
	values   []*stateValue
	dynamics []*model.Dynamics
	stepNS   int64
}

func (w *World) commitStores(ent *Entity, keep retention) {
	group := storeGroupOf(ent, -1)
	from := group.values[0].lastStep
	levels := w.stepStores(ent, group, from, group.stepNS)
	for i, value := range group.values {
		value.remember(keep)
		value.x = w.boundedLevel(ent, group, i, levels[i])
		value.lastStep = from + group.stepNS
	}
}

func (w *World) storeProbe(ent *Entity, target string, from, t int64) float64 {
	group := storeGroupOf(ent, from)
	levels := w.stepStores(ent, group, from, t-from)
	index := group.index(target)
	return w.boundedLevel(ent, group, index, levels[index])
}

func (w *World) boundedLevel(ent *Entity, group *storeGroup, index int, level float64) float64 {
	return w.clampState(ent, group.names[index], clamp(level, group.dynamics[index].F1.Clamp))
}

func storeGroupOf(ent *Entity, at int64) *storeGroup {
	group := &storeGroup{}
	for i := range ent.kind.Dynamics {
		dyn := &ent.kind.Dynamics[i]
		if isStore(dyn) {
			group.names = append(group.names, dyn.Target)
			group.values = append(group.values, storeState(ent.states[dyn.Target], at))
			group.dynamics = append(group.dynamics, dyn)
		}
	}
	group.stepNS = stepNS(group.dynamics[0])
	return group
}

func storeState(s *stateValue, at int64) *stateValue {
	if at < 0 {
		return s
	}
	return s.at(at)
}

func stepNS(dyn *model.Dynamics) int64 {
	return dyn.StepMs() * 1e6
}

func (group *storeGroup) index(name string) int {
	for i, candidate := range group.names {
		if candidate == name {
			return i
		}
	}
	return -1
}

func (group *storeGroup) levels() []float64 {
	out := make([]float64, len(group.values))
	for i, value := range group.values {
		out[i] = value.x
	}
	return out
}
