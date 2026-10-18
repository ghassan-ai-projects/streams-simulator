package domain

import (
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonvalue"
)

func objects(value any) []map[string]any {
	list, _ := value.([]any)
	items := make([]map[string]any, 0, len(list))
	for _, entry := range list {
		if object, ok := entry.(map[string]any); ok {
			items = append(items, object)
		}
	}
	return items
}

func (l *Live) birth(r map[string]any) {
	for _, kind := range objects(r["kinds"]) {
		metrics := map[string]map[string]any{}
		for _, metric := range objects(kind["metrics"]) {
			metrics[text(metric, "name")] = metric
		}
		l.kinds[text(kind, "kind")] = metrics
	}
	for _, board := range objects(r["boards"]) {
		l.boards[text(board, "board")] = text(board, "environment")
	}
}

func (l *Live) deviceBirth(r map[string]any, at int64) {
	id, boot := text(r, "device"), text(r, "boot_id")
	d := l.devices[id]
	if d == nil {
		d = &LiveDevice{Device: id, Environment: "unknown", Readings: []LiveReading{}}
		l.devices[id] = d
	}
	if d.BootID != "" && d.BootID != boot {
		d.Reboots++
		l.mark(at, "connection", "device", id+" rebooted (boot "+boot+")", id)
	}
	l.adoptDevice(d, r)
	l.mark(at, "connection", "device", id+" online", id)
}

func (l *Live) adoptDevice(d *LiveDevice, r map[string]any) {
	d.Board, d.BootID, d.kind = text(r, "board"), text(r, "boot_id"), text(r, "kind")
	d.Environment = l.boardEnvironment(d.Board)
	d.Health = healthOf(d.Environment)
	d.Instance = l.instances[d.Device]
	if d.Environment == EnvironmentHardware {
		l.gateway = "online"
		l.bindDevice(d)
	}
	d.Readings, d.aliases = []LiveReading{}, map[uint64]string{}
	d.updated = map[string]int64{}
	l.readDevice(d, objects(r["metrics"]))
}

func (l *Live) boardEnvironment(board string) string {
	if environment, announced := l.boards[board]; announced && environment != "" {
		return environment
	}
	return "unknown"
}

func (l *Live) bindDevice(d *LiveDevice) {
	for _, b := range l.bound {
		if b.Device == d.Device && b.Board == d.Board && b.Kind == d.kind {
			d.Instance = b.Instance
		}
	}
}

func (l *Live) readings(r map[string]any) bool {
	d := l.devices[text(r, "device")]
	if d == nil || !d.up() || d.BootID != text(r, "boot_id") || r["historical"] == true {
		return false
	}
	key := "metrics"
	if text(r, "type") == "DSTATE" {
		key = "values"
	}
	l.readDevice(d, objects(r[key]))
	return true
}

func (l *Live) readDevice(d *LiveDevice, metrics []map[string]any) {
	for _, metric := range metrics {
		alias, name := Seq(metric["alias"]), text(metric, "name")
		if name == "" {
			name = d.aliases[alias]
		}
		if name == "" {
			continue
		}
		if l.kinds[d.kind][name] != nil {
			d.aliases[alias] = name
		}
		l.putReading(d, name, metric)
	}
}

func (l *Live) putReading(d *LiveDevice, name string, metric map[string]any) {
	template := l.kinds[d.kind][name]
	index := slices.IndexFunc(d.Readings, func(entry LiveReading) bool { return entry.Name == name })
	if template == nil || (index < 0 && len(d.Readings) >= maxLiveReadings) {
		l.rejectReading(d, name)
		return
	}
	d.updated[name] = l.elapsedMS
	r := LiveReading{Name: name, Unit: text(template, "unit"), Value: metric["value"], Quality: text(metric, "quality")}
	r.State = readingState(r, template)
	if r.State == "good" && d.Environment == EnvironmentSimulated {
		r.State = "simulated"
	}
	storeReading(d, index, r)
}

func storeReading(d *LiveDevice, index int, r LiveReading) {
	if index < 0 {
		d.Readings = append(d.Readings, r)
		return
	}
	d.Readings[index] = r
}

func (l *Live) rejectReading(d *LiveDevice, name string) {
	if !d.unannounced {
		d.unannounced = true
		l.mark(l.nowNS, "connection", "device", d.Device+" reading "+name+" not announced: ignored", d.Device)
	}
}

func readingState(r LiveReading, template map[string]any) string {
	switch r.Quality {
	case "stale", "disconnected":
		return r.Quality
	case "good":
		if !plausibleReading(r.Value, template) {
			return "invalid"
		}
		return "good"
	default:
		return "untrusted"
	}
}

func plausibleReading(value any, template map[string]any) bool {
	if template == nil || !plausibleType(value, template) {
		return false
	}
	if bounds, ok := template["eu_range"].(map[string]any); ok {
		v, numeric := jsonvalue.Number(value)
		low, lowOK := jsonvalue.Number(bounds["min"])
		high, highOK := jsonvalue.Number(bounds["max"])
		return numeric && lowOK && highOK && v >= low && v <= high
	}
	return true
}

func (l *Live) death(r map[string]any, at int64) bool {
	d := l.devices[text(r, "device")]
	if d == nil || d.BootID != text(r, "boot_id") {
		return false
	}
	l.disconnectDevice(d)
	l.mark(at, "connection", "device", d.Device+" offline", d.Device)
	return true
}

func healthOf(environment string) string {
	if environment == EnvironmentSimulated {
		return "simulated"
	}
	return "online"
}

func (d *LiveDevice) up() bool { return d.Health == "online" || d.Health == "simulated" }
