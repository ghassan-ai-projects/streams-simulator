package model

import "slices"

func (f *Fault) IsPulse() bool { return f.Onset.Shape == ShapePulse }

func (k *Kind) FaultCommanded(f *Fault) bool {
	commanded := map[string]bool{}
	for _, effector := range k.Effectors {
		for _, delta := range effector.Effect.StateDeltas {
			commanded[delta.State] = true
		}
	}
	return slices.ContainsFunc(f.Affects, func(e FaultEffect) bool { return commanded[e.State] })
}
