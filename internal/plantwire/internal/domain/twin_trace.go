package domain

import (
	"cmp"
	"encoding/json"
	"maps"
	"math"
	"slices"
)

type BoardDevice struct{ Name, Kind, Board string }

type BoardMetric struct{ Device, Kind, Metric string }

type BoardReading struct {
	Device, Metric, Quality string
	AtNS                    int64
	Value                   float64
	Numeric, Historical     bool
}

type BoardReceipt struct {
	Status, Reason, Outcome string
	AtNS                    int64
}

type BoardTrace struct {
	Devices      []BoardDevice
	Metrics      []BoardMetric
	Readings     []BoardReading
	Receipts     map[string]BoardReceipt
	StartNS      int64
	Unattributed int
}

type boardReader struct {
	out    BoardTrace
	names  map[string]map[int64]string
	latest int64
	timed  bool
}

func ReadBoardTrace(trace []map[string]any) BoardTrace {
	reader := &boardReader{out: BoardTrace{Receipts: map[string]BoardReceipt{}}, names: map[string]map[int64]string{}}
	for _, record := range trace {
		reader.take(record)
	}
	slices.SortFunc(reader.out.Metrics, func(a, b BoardMetric) int {
		return cmp.Or(cmp.Compare(a.Device, b.Device), cmp.Compare(a.Metric, b.Metric))
	})
	slices.SortFunc(reader.out.Devices, func(a, b BoardDevice) int { return cmp.Compare(a.Name, b.Name) })
	return reader.out
}

func (b BoardTrace) Has(device string) bool {
	return slices.ContainsFunc(b.Devices, func(d BoardDevice) bool { return d.Name == device })
}

func (r *boardReader) take(record map[string]any) {
	switch text(record, "type") {
	case "DBIRTH":
		r.birth(record)
	case "DDATA":
		r.data(record)
	case "DRECEIPT":
		r.receipt(record)
	case "DRESULT":
		r.result(record)
	}
}

func (r *boardReader) birth(record map[string]any) {
	device, kind := text(record, "device"), text(record, "kind")
	if device == "" {
		return
	}
	r.names[device] = aliasNames(record["metrics"])
	if at := slices.IndexFunc(r.out.Devices, func(d BoardDevice) bool { return d.Name == device }); at >= 0 {
		r.out.Devices = slices.Delete(r.out.Devices, at, at+1)
	}
	r.out.Devices = append(r.out.Devices, BoardDevice{Name: device, Kind: kind, Board: text(record, "board")})
	r.announce(device, kind)
}

func (r *boardReader) announce(device, kind string) {
	for _, alias := range slices.Sorted(maps.Keys(r.names[device])) {
		metric := BoardMetric{Device: device, Kind: kind, Metric: r.names[device][alias]}
		if !slices.Contains(r.out.Metrics, metric) {
			r.out.Metrics = append(r.out.Metrics, metric)
		}
	}
}

func aliasNames(metrics any) map[int64]string {
	names := map[int64]string{}
	for _, entry := range listOf(metrics) {
		fields, _ := entry.(map[string]any)
		if alias, ok := nsOf(fields["alias"]); ok {
			names[alias] = text(fields, "name")
		}
	}
	return names
}

func (r *boardReader) data(record map[string]any) {
	device := text(record, "device")
	historical, _ := record["historical"].(bool)
	for _, entry := range listOf(record["metrics"]) {
		sample, _ := entry.(map[string]any)
		r.reading(device, historical, sample)
	}
}

func (r *boardReader) reading(device string, historical bool, sample map[string]any) {
	at, _ := nsOf(sample["event_time_ns"])
	r.note(sample["event_time_ns"])
	alias, _ := nsOf(sample["alias"])
	metric := r.names[device][alias]
	if metric == "" {
		r.out.Unattributed++
		return
	}
	value, numeric := numberOf(sample["value"])
	r.out.Readings = append(r.out.Readings, BoardReading{Device: device, Metric: metric, Quality: text(sample, "quality"),
		AtNS: at, Value: value, Numeric: numeric, Historical: historical})
}

func (r *boardReader) note(value any) {
	at, ok := nsOf(value)
	switch {
	case !ok:
		return
	case !r.timed:
		r.out.StartNS, r.latest, r.timed = at, at, true
	default:
		r.out.StartNS, r.latest = min(r.out.StartNS, at), max(r.latest, at)
	}
}

func (r *boardReader) receipt(record map[string]any) {
	id := text(record, "command_id")
	kept := r.out.Receipts[id]
	r.out.Receipts[id] = BoardReceipt{Status: text(record, "status"), Reason: text(record, "reason"), Outcome: kept.Outcome, AtNS: r.latest}
}

func (r *boardReader) result(record map[string]any) {
	id := text(record, "command_id")
	kept := r.out.Receipts[id]
	kept.Outcome = text(record, "reported_outcome")
	r.out.Receipts[id] = kept
}

func nsOf(value any) (int64, bool) {
	switch number := value.(type) {
	case json.Number:
		parsed, err := number.Int64()
		return parsed, err == nil
	case float64:
		return int64(number), number == math.Trunc(number)
	}
	return 0, false
}
