package domain

import (
	"reflect"
	"testing"
)

func kindAnnouncement() map[string]any {
	return map[string]any{
		"type": "NBIRTH",
		"kinds": []any{
			map[string]any{"kind": "fan", "metrics": []any{map[string]any{
				"name": "level_percent", "data_type": "integer", "unit": "%", "eu_range": map[string]any{"min": 0.0, "max": 60.0}}}},
			map[string]any{"kind": "indicator", "metrics": []any{map[string]any{
				"name": "state", "data_type": "enum", "enum_values": []any{"off", "watch", "alert"}}}},
		},
		"boards": []any{map[string]any{"board": "board-01", "environment": EnvironmentHardware}},
	}
}

func birthOfKind(device, kind, metric string, value any) map[string]any {
	return map[string]any{"type": "DBIRTH", "device": device, "kind": kind, "board": "board-01", "boot_id": "boot-72",
		"metrics": []any{map[string]any{"name": metric, "alias": 1, "value": value, "quality": "warming_up"}}}
}

func aliasData(device string, value any) map[string]any {
	return map[string]any{"type": "DDATA", "device": device, "boot_id": "boot-72",
		"metrics": []any{map[string]any{"alias": 1, "value": value, "quality": "good"}}}
}

func readingsOf(l *Live, device string) []LiveReading { return l.devices[device].Readings }

func TestARealDeviceReportsTheMetricsOfItsKindWhateverTheDomainCallsItsChannels(t *testing.T) {
	t.Parallel()
	bound := []Bound{{Instance: "fan-01", Board: "board-01", Device: "fan-01", Kind: "fan"},
		{Instance: "led-01", Board: "board-01", Device: "led-01", Kind: "indicator"}}
	l := NewLive(true, nil, bound)
	l.Observe(TraceNode, kindAnnouncement(), 0)
	l.Observe(TraceNode, birthOfKind("fan-01", "fan", "level_percent", nil), 1)
	l.Observe(TraceNode, birthOfKind("led-01", "indicator", "state", nil), 1)
	l.Observe(TraceNode, aliasData("fan-01", 0.0), 2)
	l.Observe(TraceNode, aliasData("led-01", "off"), 2)
	if got, want := readingsOf(l, "fan-01"), []LiveReading{{Name: "level_percent", Unit: "%", Value: 0.0, Quality: "good", State: "good"}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("fan readings = %+v, want %+v", got, want)
	}
	if got := readingsOf(l, "led-01"); len(got) != 1 || got[0].Name != "state" || got[0].Value != "off" || got[0].State != "good" {
		t.Fatalf("indicator readings = %+v", got)
	}
	l.Observe(TraceNode, aliasData("fan-01", 45.0), 3)
	l.Observe(TraceNode, aliasData("led-01", "alert"), 3)
	if fan, led := readingsOf(l, "fan-01")[0], readingsOf(l, "led-01")[0]; fan.Value != 45.0 || led.Value != "alert" || l.devices["fan-01"].Instance != "fan-01" {
		t.Fatalf("after a change: fan %+v, indicator %+v", fan, led)
	}
}
