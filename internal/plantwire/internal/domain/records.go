// Package domain holds the plant wire rules of a node: records, digests, schema checks and sequencing, without I/O.
package domain

const (
	LaneControl = "control"
	LaneData    = "data"
	Version     = 1
)

type Record struct {
	Type   string
	Lane   string
	Fields map[string]any
}

func newRecord(kind, lane string, fields map[string]any) Record {
	return Record{Type: kind, Lane: lane, Fields: fields}
}

func DeviceData(device, bootID string, historical bool, samples []map[string]any) Record {
	return newRecord("DDATA", LaneData, map[string]any{"device": device, "boot_id": bootID, "historical": historical, "metrics": toAny(samples)})
}

func Sample(alias int, value any, quality string, eventTimeNS int64) map[string]any {
	if quality == "" {
		quality = "good"
	}
	return map[string]any{"alias": alias, "value": value, "quality": quality, "event_time_ns": eventTimeNS, "time_source": "device"}
}

func NodeDeath(bdSeq uint64, reason string) Record {
	return newRecord("NDEATH", LaneControl, map[string]any{"bd_seq": bdSeq, "reason": reason})
}

func Receipt(commandID, status, reason, bootID string) Record {
	fields := map[string]any{"command_id": commandID, "status": status, "boot_id": bootID}
	if reason != "" {
		fields["reason"] = reason
	}
	return newRecord("DRECEIPT", LaneControl, fields)
}

func Sync(throughSeq uint64, worldTimeNS int64) Record {
	return newRecord("NSYNC", LaneControl, map[string]any{"through_seq": throughSeq, "world_time_ns": worldTimeNS})
}

func toAny[T any](items []T) []any {
	out := make([]any, len(items))
	for i, item := range items {
		out[i] = item
	}
	return out
}

func DeviceState(device, bootID string, values []any) Record {
	return newRecord("DSTATE", LaneControl, map[string]any{"device": device, "boot_id": bootID, "self_reported": true, "safe_state": false,
		"outputs": []any{}, "values": values})
}
