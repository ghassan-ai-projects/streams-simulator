package domain

import (
	"fmt"
	"maps"
	"slices"
)

type Unit struct {
	Scale  float64 `json:"scale"`
	Offset float64 `json:"offset"`
}

type Family struct {
	Base  string          `json:"base"`
	Units map[string]Unit `json:"units"`
}

type PortType struct {
	UnitFamily string `json:"unit_family"`
}

type TagLetter struct {
	Measures []string `json:"measures"`
}

type Vocabulary struct {
	UnitFamilies map[string]Family    `json:"unit_families"`
	PortTypes    map[string]PortType  `json:"port_types"`
	TagLetters   map[string]TagLetter `json:"tag_letters"`
}

func newVocabulary() *Vocabulary {
	return &Vocabulary{UnitFamilies: map[string]Family{}, PortTypes: map[string]PortType{}, TagLetters: map[string]TagLetter{}}
}

func (v *Vocabulary) validateBases() error {
	for _, name := range slices.Sorted(maps.Keys(v.UnitFamilies)) {
		family := v.UnitFamilies[name]
		if base, ok := family.Units[family.Base]; !ok || base.Scale != 1 || base.Offset != 0 {
			return fmt.Errorf("unit family %q: the base unit %q must be a unit with scale 1 and no offset", name, family.Base)
		}
	}
	return nil
}

func (v *Vocabulary) merge(add *Vocabulary) error {
	for _, step := range []func(*Vocabulary) error{v.mergeFamilies, v.mergePortTypes, v.mergeTags} {
		if err := step(add); err != nil {
			return err
		}
	}
	if err := v.validateBases(); err != nil {
		return err
	}
	return v.checkReferences()
}

func (v *Vocabulary) mergeFamilies(add *Vocabulary) error {
	for _, name := range slices.Sorted(maps.Keys(add.UnitFamilies)) {
		family, known := v.UnitFamilies[name]
		if known && family.Base != add.UnitFamilies[name].Base {
			return fmt.Errorf("unit family %q: base unit %q differs from the existing %q", name, add.UnitFamilies[name].Base, family.Base)
		}
		merged, err := mergeUnits(name, family, add.UnitFamilies[name])
		if err != nil {
			return err
		}
		v.UnitFamilies[name] = merged
	}
	return nil
}

func mergeUnits(name string, existing, add Family) (Family, error) {
	units := maps.Clone(existing.Units)
	if units == nil {
		units = map[string]Unit{}
	}
	for _, unit := range slices.Sorted(maps.Keys(add.Units)) {
		if old, known := units[unit]; known && old != add.Units[unit] {
			return Family{}, fmt.Errorf("unit %q of family %q is defined differently from the existing definition", unit, name)
		}
		units[unit] = add.Units[unit]
	}
	return Family{Base: add.Base, Units: units}, nil
}

func (v *Vocabulary) mergePortTypes(add *Vocabulary) error {
	for _, name := range slices.Sorted(maps.Keys(add.PortTypes)) {
		if old, known := v.PortTypes[name]; known && old != add.PortTypes[name] {
			return fmt.Errorf("port type %q is defined differently: unit family %q, existing %q", name, add.PortTypes[name].UnitFamily, old.UnitFamily)
		}
		v.PortTypes[name] = add.PortTypes[name]
	}
	return nil
}

func (v *Vocabulary) mergeTags(add *Vocabulary) error {
	for _, tag := range slices.Sorted(maps.Keys(add.TagLetters)) {
		measures := slices.Concat(v.TagLetters[tag].Measures, add.TagLetters[tag].Measures)
		slices.Sort(measures)
		v.TagLetters[tag] = TagLetter{Measures: slices.Compact(measures)}
	}
	return nil
}

func (v *Vocabulary) checkReferences() error {
	for _, name := range slices.Sorted(maps.Keys(v.PortTypes)) {
		if _, ok := v.UnitFamilies[v.PortTypes[name].UnitFamily]; !ok {
			return fmt.Errorf("port type %q names unit family %q, which no vocabulary defines", name, v.PortTypes[name].UnitFamily)
		}
	}
	for _, tag := range slices.Sorted(maps.Keys(v.TagLetters)) {
		for _, family := range v.TagLetters[tag].Measures {
			if _, ok := v.UnitFamilies[family]; !ok {
				return fmt.Errorf("tag %q measures unit family %q, which no vocabulary defines", tag, family)
			}
		}
	}
	return nil
}

type conversion struct{ gain, offset float64 }

func (c conversion) identity() bool { return c.gain == 1 && c.offset == 0 }

func (c conversion) apply(value float64) float64 { return value*c.gain + c.offset }

func (v *Vocabulary) convert(family, from, to string) (conversion, error) {
	units := v.UnitFamilies[family].Units
	source, okFrom := units[from]
	target, okTo := units[to]
	if !okFrom || !okTo {
		return conversion{}, fmt.Errorf("unit %q or %q is not a unit of family %q", from, to, family)
	}
	return conversion{gain: source.Scale / target.Scale, offset: (source.Offset - target.Offset) / target.Scale}, nil
}

func (v *Vocabulary) hasUnit(family, unit string) bool {
	_, ok := v.UnitFamilies[family].Units[unit]
	return ok
}

func (v *Vocabulary) tagMeasures(tag, family string) bool {
	return slices.Contains(v.TagLetters[tag].Measures, family)
}
