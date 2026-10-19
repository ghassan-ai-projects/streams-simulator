package domain

import (
	"bytes"
	"encoding/json"
	"strconv"
	"testing"
)

type twAlias struct {
	name  string
	alias int64
}

func twInt(n int64) json.Number { return json.Number(strconv.FormatInt(n, 10)) }

func twBirth(device, kind string, at int64, aliases ...twAlias) map[string]any {
	metrics := []any{}
	for _, a := range aliases {
		metrics = append(metrics, map[string]any{"name": a.name, "alias": twInt(a.alias)})
	}
	return map[string]any{"type": "DBIRTH", "device": device, "kind": kind, "board": "mega-01",
		"birth_time_ns": twInt(at), "metrics": metrics}
}

func twData(device string, historical bool, samples ...map[string]any) map[string]any {
	list := []any{}
	for _, sample := range samples {
		list = append(list, sample)
	}
	return map[string]any{"type": "DDATA", "device": device, "historical": historical, "metrics": list}
}

func twSample(alias int64, value any, quality string, at int64) map[string]any {
	return map[string]any{"alias": twInt(alias), "value": value, "quality": quality, "event_time_ns": twInt(at)}
}

func twReceipt(id, status, reason string) map[string]any {
	return map[string]any{"type": "DRECEIPT", "command_id": id, "status": status, "reason": reason}
}

func twResult(id, outcome string) map[string]any {
	return map[string]any{"type": "DRESULT", "command_id": id, "reported_outcome": outcome}
}

// twJournalLine is one line of journal.jsonl as the wire trace writes a host record.
func twJournalLine(direction string, record map[string]any) string {
	raw, err := json.Marshal(map[string]any{"type": "wire", "direction": direction, "record": record})
	if err != nil {
		panic(err)
	}
	return string(raw) + "\n"
}

// twDecode reads a JSON object the way run.ReadFieldTrace does: numbers stay json.Number.
func twDecode(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var record map[string]any
	if err := decoder.Decode(&record); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}
	return record
}

func twNumber(t *testing.T, text string) json.Number {
	t.Helper()
	number := json.Number(text)
	if _, err := number.Float64(); err != nil {
		t.Fatalf("twNumber(%q): %v", text, err)
	}
	return number
}
