package domain

import (
	"cmp"
	"maps"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (b *entryBuild) chooseInstruments() error {
	points, err := b.sizedPoints()
	if err != nil {
		return err
	}
	attached, err := b.attachedPoints(points)
	if err != nil {
		return err
	}
	if err := b.pickInstruments(attached); err != nil {
		return err
	}
	return b.requireReferencedPoints()
}

func (b *entryBuild) sizedPoints() ([]MeasurementPoint, error) {
	raw, err := canonical.Marshal(b.ws.points)
	if err != nil {
		return nil, b.refuse("use", "measurement points of %s: %v", b.equipment.ID, err)
	}
	var points []MeasurementPoint
	if err := model.DecodeBytes(raw, &points); err != nil {
		return nil, b.refuse("use", "measurement points of %s: %v", b.equipment.ID, err)
	}
	return points, nil
}

func (b *entryBuild) attachedPoints(points []MeasurementPoint) ([]MeasurementPoint, error) {
	declared := make(map[string]MeasurementPoint, len(points))
	for _, point := range points {
		declared[point.Name] = point
	}
	if err := b.admitAdditions(declared); err != nil {
		return nil, err
	}
	if err := b.admitRemovals(declared); err != nil {
		return nil, err
	}
	return slices.DeleteFunc(slices.Clone(points), b.detached), nil
}

func (b *entryBuild) detached(point MeasurementPoint) bool {
	changes := b.entry.Instruments
	optionalOff := point.Optional && !slices.Contains(changes.Add, point.Name)
	return optionalOff || slices.Contains(changes.Remove, point.Name)
}

func (b *entryBuild) admitAdditions(declared map[string]MeasurementPoint) error {
	for index, name := range b.entry.Instruments.Add {
		point, found := declared[name]
		switch {
		case !found:
			return b.refuse(relative("instruments", "add", indexPart(index)), "%s has no measurement point %q", b.equipment.ID, name)
		case !point.Optional:
			return b.refuse(relative("instruments", "add", indexPart(index)), "measurement point %q is not optional: it is attached already", name)
		}
	}
	return nil
}

func (b *entryBuild) admitRemovals(declared map[string]MeasurementPoint) error {
	for index, name := range b.entry.Instruments.Remove {
		point, found := declared[name]
		if !found || (point.Optional && !slices.Contains(b.entry.Instruments.Add, name)) {
			return b.refuse(relative("instruments", "remove", indexPart(index)), "measurement point %q is not attached", name)
		}
	}
	return nil
}

func (b *entryBuild) pickInstruments(attached []MeasurementPoint) error {
	b.points = attached
	if err := b.admitReplacements(attached); err != nil {
		return err
	}
	for _, point := range attached {
		if err := b.pick(point); err != nil {
			return err
		}
	}
	return nil
}

func (b *entryBuild) admitReplacements(attached []MeasurementPoint) error {
	for _, name := range slices.Sorted(maps.Keys(b.entry.Instruments.Replace)) {
		if !slices.ContainsFunc(attached, func(point MeasurementPoint) bool { return point.Name == name }) {
			return b.refuse(relative("instruments", "replace", name), "measurement point %q is not attached", name)
		}
	}
	return nil
}

func (b *entryBuild) pick(point MeasurementPoint) error {
	ref := cmp.Or(b.entry.Instruments.Replace[point.Name], point.DefaultInstrument)
	instrument, found := b.c.lib.instruments[ref]
	at := relative("instruments", "replace", point.Name)
	switch {
	case !found:
		return b.refuse(at, "%s is not an instrument of the library", ref)
	case instrument.Tag != point.Tag || instrument.UnitFamily != point.UnitFamily:
		return b.refuse(at, "%s measures %s/%s, the point %q needs %s/%s", ref, instrument.Tag, instrument.UnitFamily, point.Name, point.Tag, point.UnitFamily)
	}
	b.instruments[point.Name] = instrument
	return nil
}

func (b *entryBuild) requireReferencedPoints() error {
	for _, reference := range pointReferences(b.ws.kind) {
		if _, attached := b.instruments[reference.name]; !attached {
			return b.refuse("instruments", "%s needs measurement point %q, which this entry does not attach", reference.what, reference.name)
		}
	}
	return nil
}
