package app

import (
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
}

func newDevices(r *run.Run, withheld []string) *devices {
	d := &devices{run: r, spec: r.Domain(), born: map[string]bool{}, current: map[string]map[string]any{},
		wireIDs: map[string]string{}, ids: map[string]string{}, withheld: map[string]bool{}}
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
		kinds[i] = &d.spec.Spec.Kinds[i]
	}
	return kinds
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
	effective := d.spec.InstanceKind(d.spec.InstanceSetFor(id))
	identity := wire.DeviceIdentity{Device: d.wireOf(id), Kind: effective.Name, Board: d.board(id), BootID: d.boot(id), BirthTimeNS: atNS}
	return wire.DeviceBirth(identity, d.spec.Kind(effective.Name), effective, d.current[id])
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
	kind := d.spec.InstanceKind(d.spec.InstanceSetFor(event.EntityID))
	alias := wire.Alias(kind, event.Channel)
	if alias == 0 || event.Value == nil || !wire.Finite(event.Value) || !wire.Carried(kind.ChannelDecl(event.Channel)) {
		return records
	}
	d.remember(event)
	sample := wire.Sample(alias, event.Value, event.Quality, mustTime(event.EventTime))
	return append(records, wire.DeviceData(d.wireOf(event.EntityID), d.boot(event.EntityID), false, []map[string]any{sample}))
}

func (d *devices) remember(event model.SimEvent) {
	if d.current[event.EntityID] == nil {
		d.current[event.EntityID] = map[string]any{}
	}
	d.current[event.EntityID][event.Channel] = event.Value
}

func mustTime(text string) int64 {
	ns, _ := model.ParseTime(text)
	return ns
}
