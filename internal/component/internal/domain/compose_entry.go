package domain

import (
	"cmp"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type built struct {
	kindName  string
	kind      map[string]any
	instances map[string]any
	ports     map[string]Port
	entry     Entry
	index     int
}

type entryBuild struct {
	c           *composer
	index       int
	entry       Entry
	equipment   *Equipment
	name        string
	ws          *workspace
	preset      string
	parameters  map[string]any
	points      []MeasurementPoint
	instruments map[string]*Instrument
	channels    []any
	faults      []any
	biases      []any
}

func (c *composer) buildEntry(index int, entry Entry) (*built, error) {
	b := &entryBuild{c: c, index: index, entry: entry, parameters: map[string]any{}, instruments: map[string]*Instrument{}}
	steps := []func() error{b.resolveEquipment, b.checkArea, b.nameKind, b.applySizes, b.chooseInstruments, b.attachInstruments, b.qualifyKind}
	for _, step := range steps {
		if err := step(); err != nil {
			return nil, err
		}
	}
	return b.finish(), nil
}

func (b *entryBuild) at(member string) string {
	return pointer("equipment", indexPart(b.index)) + "/" + member
}

func (b *entryBuild) refuse(member, format string, args ...any) error {
	return b.c.refuse(b.at(member), format, args...)
}

func (b *entryBuild) resolveEquipment() error {
	equipment, found := b.c.lib.equipment[b.entry.Use]
	if !found {
		return b.refuse("use", "%s is not an equipment of the library", b.entry.Use)
	}
	b.equipment = equipment
	b.ws = newWorkspace(equipment)
	return nil
}

func (b *entryBuild) checkArea() error {
	if b.entry.Area != "" && !b.c.areas[b.entry.Area] {
		return b.refuse("area", "area %q is not declared in areas", b.entry.Area)
	}
	return nil
}

func (b *entryBuild) nameKind() error {
	b.name = cmp.Or(b.entry.As, defaultKindName(b.equipment.ID))
	switch {
	case !qualifiable.MatchString(b.name) || len(b.name) > maxKindName:
		return b.refuse("as", "kind name %q must match %s and be at most %d characters", b.name, qualifiable, maxKindName)
	case b.c.kinds[b.name]:
		return b.refuse("as", "kind name %q is used by another entry", b.name)
	}
	b.c.kinds[b.name] = true
	return nil
}

func (b *entryBuild) ports() map[string]Port {
	ports := make(map[string]Port, len(b.equipment.Ports))
	for _, port := range b.equipment.Ports {
		ports[port.Name] = port
	}
	return ports
}

func (b *entryBuild) finish() *built {
	return &built{kindName: b.name, kind: b.kindDocument(), instances: b.instanceSet(), ports: b.ports(), entry: b.entry, index: b.index}
}

func (b *entryBuild) instanceSet() map[string]any {
	set := map[string]any{"kind": b.name, "provenance": b.provenance()}
	if len(b.entry.Names) > 0 {
		set["names"] = stringList(b.entry.Names)
	}
	if b.entry.IDTemplate != "" {
		set["id_template"] = b.entry.IDTemplate
		set["count"] = countDocument(b.entry.Count)
	}
	if b.entry.Area != "" {
		set["area"] = b.entry.Area
	}
	return set
}

func countDocument(count model.EntityCount) map[string]any {
	document := map[string]any{"default": count.Default}
	if count.Min > 0 {
		document["min"] = count.Min
	}
	if count.Max > 0 {
		document["max"] = count.Max
	}
	return document
}

func stringList(values []string) []any {
	out := make([]any, len(values))
	for i, value := range values {
		out[i] = value
	}
	return out
}

func (b *entryBuild) provenance() map[string]any {
	refs, digests := b.instrumentRecords()
	provenance := map[string]any{"component": b.equipment.ID, "version": b.equipment.Version, "digest": b.equipment.Digest,
		"instruments": refs, "instrument_digests": digests}
	if b.preset != "" {
		provenance["preset"] = b.preset
	}
	if len(b.parameters) > 0 {
		provenance["parameters"] = b.parameters
	}
	return provenance
}

func (b *entryBuild) instrumentRecords() (map[string]any, map[string]any) {
	refs, digests := map[string]any{}, map[string]any{}
	for _, point := range b.points {
		instrument := b.instruments[point.Name]
		refs[point.Name] = reference(instrument.ID, instrument.Version)
		digests[reference(instrument.ID, instrument.Version)] = instrument.Digest
	}
	return refs, digests
}

func (b *entryBuild) kindDocument() map[string]any {
	kind := map[string]any{"name": b.name, "channels": b.channels, "faults": b.allFaults()}
	for _, section := range []string{"system", "dynamics", "flows", "effectors", "boundaries", "kpis"} {
		if value, present := b.ws.kind[section]; present {
			kind[section] = value
		}
	}
	kind["state"] = slices.Concat(asList(b.ws.kind["state"]), b.biases)
	return kind
}

func (b *entryBuild) allFaults() []any {
	return slices.Concat(asList(b.ws.kind["faults"]), b.faults)
}

func asList(value any) []any {
	list, _ := value.([]any)
	return list
}
