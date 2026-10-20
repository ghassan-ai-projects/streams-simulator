package app

import (
	"log/slog"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	wire "github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

type devices struct {
	run      *run.Run
	spec     *domain.Compiled
	born     map[string]bool
	current  map[string]map[string]any
	wireIDs  map[string]string
	ids      map[string]string
	withheld map[string]bool
	faces    wire.Interfaces
}

func newDevices(r *run.Run, withheld []string, faces wire.Interfaces) *devices {
	d := &devices{run: r, spec: r.Domain(), born: map[string]bool{}, current: map[string]map[string]any{},
		wireIDs: map[string]string{}, ids: map[string]string{}, withheld: map[string]bool{}, faces: faces}
	for _, id := range withheld {
		d.withheld[id] = true
	}
	return d
}

func boundInstances(bound []wire.Bound) []string {
	ids := make([]string, len(bound))
	for i, b := range bound {
		ids[i] = b.Instance
	}
	return ids
}

func (d *devices) announced() []string {
	var ids []string
	for _, id := range d.run.World.EntityIDs() {
		if !d.withheld[id] {
			ids = append(ids, id)
		}
	}
	return ids
}

func (d *devices) instances() map[string]string {
	instances := map[string]string{}
	for _, id := range d.announced() {
		instances[d.wireOf(id)] = id
	}
	return instances
}

func (d *devices) wireOf(id string) string {
	if known, ok := d.wireIDs[id]; ok {
		return known
	}
	chosen := wire.Distinct(wire.WireID(id), func(candidate string) bool { _, taken := d.ids[candidate]; return taken })
	d.wireIDs[id], d.ids[chosen] = chosen, id
	return chosen
}

func (d *devices) entity(device string) (string, bool) {
	id, ok := d.ids[device]
	return id, ok && !d.withheld[id] && d.run.World.Entity(id) != nil
}

func (d *devices) kinds() []*model.Kind {
	kinds := make([]*model.Kind, len(d.spec.Spec.Kinds))
	for i := range d.spec.Spec.Kinds {
		kinds[i], _ = d.faces.Of(&d.spec.Spec.Kinds[i])
	}
	return kinds
}

func (d *devices) kindOf(id string) (*model.Kind, *wire.MappedKind) {
	effective := d.spec.InstanceKind(d.spec.InstanceSetFor(id))
	if effective == nil {
		return nil, nil
	}
	if face, mapped := d.faces.Of(d.spec.Kind(effective.Name)); mapped != nil {
		return face, mapped
	}
	return effective, nil
}

func (d *devices) births(atNS int64) []wire.Record {
	var records []wire.Record
	for _, id := range d.announced() {
		records = append(records, d.birth(id, atNS))
	}
	return records
}

func (d *devices) birth(id string, atNS int64) wire.Record {
	d.born[id] = true
	effective, mapped := d.kindOf(id)
	base := d.spec.Kind(effective.Name)
	if mapped != nil {
		base = effective
	}
	identity := wire.DeviceIdentity{Device: d.wireOf(id), Kind: effective.Name, Board: d.board(id), BootID: d.boot(id), BirthTimeNS: atNS}
	return wire.DeviceBirth(identity, base, effective, d.current[id])
}

func (d *devices) observe(event model.SimEvent) []wire.Record {
	if d.withheld[event.EntityID] {
		return nil
	}
	return d.carried(event)
}

func (d *devices) carried(event model.SimEvent) []wire.Record {
	var records []wire.Record
	if !d.born[event.EntityID] {
		records = append(records, d.birth(event.EntityID, mustTime(event.EventTime)))
	}
	kind, channel, value, carried := d.reading(event)
	alias := wire.Alias(kind, channel)
	if !carried || alias == 0 || value == nil || !wire.Finite(value) || !wire.Carried(kind.ChannelDecl(channel)) {
		return records
	}
	d.remember(event.EntityID, channel, value)
	sample := wire.Sample(alias, value, event.Quality, mustTime(event.EventTime))
	return append(records, wire.DeviceData(d.wireOf(event.EntityID), d.boot(event.EntityID), false, []map[string]any{sample}))
}

func (d *devices) reading(event model.SimEvent) (*model.Kind, string, any, bool) {
	kind, mapped := d.kindOf(event.EntityID)
	if mapped == nil {
		return kind, event.Channel, event.Value, kind != nil
	}
	metric, value, carried := wire.SemanticReading(*mapped, event.Channel, event.Value)
	if metric != "" && !carried {
		slog.Warn("plantwire: a raw reading has no word in its board mapping and is not carried",
			"device", event.EntityID, "raw", event.Channel, "value", event.Value)
	}
	return kind, metric, value, carried
}

func (d *devices) remember(id, channel string, value any) {
	if d.current[id] == nil {
		d.current[id] = map[string]any{}
	}
	d.current[id][channel] = value
}

func mustTime(text string) int64 {
	ns, _ := model.ParseTime(text)
	return ns
}
