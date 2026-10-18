package domain

import (
	"fmt"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestDeviceDeathStopsReadingsAndDataAfterItIsIgnored(t *testing.T) {
	t.Parallel()
	l := liveBoard(t)
	death := map[string]any{"type": "DDEATH", "device": "P1", "boot_id": "boot-1"}
	if !l.Observe(TraceNode, death, 5) || l.devices["P1"].Health != "offline" || l.devices["P1"].Readings[0].State != "disconnected" {
		t.Fatalf("dead device = %+v", l.devices["P1"])
	}
	data := map[string]any{"type": "DDATA", "device": "P1", "boot_id": "boot-1", "metrics": []any{map[string]any{"alias": 1, "value": 5.0, "quality": "good"}}}
	if l.Observe(TraceNode, data, 6) || l.devices["P1"].Readings[0].State != "disconnected" {
		t.Fatal("data after death revived the reading")
	}
	if l.Observe(TraceNode, map[string]any{"type": "DDEATH", "device": "P1", "boot_id": "other"}, 7) || l.Observe(TraceNode, map[string]any{"type": "DDEATH", "device": "ghost"}, 7) {
		t.Fatal("a death of another boot or device was taken")
	}
}

func TestBootIdReuseIsNotARebootAndChurnIsCounted(t *testing.T) {
	t.Parallel()
	l := liveBoard(t)
	l.Observe(TraceNode, liveRecord(deviceBirthOf("P1", "boot-1")), 5)
	if l.devices["P1"].Reboots != 0 {
		t.Fatalf("a repeated boot id counted as a reboot: %d", l.devices["P1"].Reboots)
	}
	for i, boot := range []string{"boot-2", "boot-1", "boot-2"} {
		l.Observe(TraceNode, liveRecord(deviceBirthOf("P1", boot)), int64(6+i))
	}
	if l.devices["P1"].Reboots != 3 {
		t.Fatalf("reboots = %d, want 3", l.devices["P1"].Reboots)
	}
}

func TestTheHostsNodeDeathAndOtherRecordsAreProjectedOrIgnored(t *testing.T) {
	t.Parallel()
	l := liveBoard(t)
	l.Observe(TraceHost, map[string]any{"type": "STATE", "online": true, "protocol_version": 1, "protocol_min": 1, "protocol_max": 1}, 2)
	if !l.Observe(TraceNode, map[string]any{"type": "NDEATH"}, 3) || l.host != "offline" {
		t.Fatalf("host after node death = %s", l.host)
	}
	if l.Observe(TraceHost, map[string]any{"type": "DQUERY", "device": "P1"}, 4) || l.Observe(TraceNode, map[string]any{"type": "NSYNC"}, 4) || l.Observe("relay", map[string]any{}, 4) {
		t.Fatal("an unprojected record changed the live state")
	}
}

func TestAMetricWithoutAScanClassGoesStaleAfterTheGatewayIdleLimit(t *testing.T) {
	t.Parallel()
	l := liveBoard(t)
	l.kinds["pump"]["aux"] = map[string]any{"name": "aux", "data_type": "number"}
	l.Observe(TraceNode, map[string]any{"type": "DDATA", "device": "P1", "boot_id": "boot-1", "metrics": []any{
		map[string]any{"name": "aux", "alias": 5, "value": 1.0, "quality": "good"}}}, 5)
	l.Tick(gatewayIdleMS - 1)
	if state := l.devices["P1"].Readings[1].State; state != "good" {
		t.Fatalf("aux before the idle limit = %s", state)
	}
	if !l.Tick(1) || l.devices["P1"].Readings[1].State != "stale" {
		t.Fatalf("aux at the idle limit = %+v", l.devices["P1"].Readings)
	}
}

func TestReadingStatesFollowTheDeclaredTypeRangeAndQuality(t *testing.T) {
	t.Parallel()
	enum := map[string]any{"data_type": "enum", "enum_values": []any{"ready"}}
	ranged := map[string]any{"data_type": "number", "eu_range": map[string]any{"min": 0.0, "max": 10.0}}
	for name, c := range map[string]struct {
		value    any
		quality  string
		template map[string]any
		want     string
	}{
		"boolean":       {true, "good", map[string]any{"data_type": "boolean"}, "good"},
		"string":        {"x", "good", map[string]any{"data_type": "string"}, "good"},
		"enum":          {"ready", "good", enum, "good"},
		"integer":       {3.0, "good", map[string]any{"data_type": "integer"}, "good"},
		"in range":      {5.0, "good", ranged, "good"},
		"out of range":  {11.0, "good", ranged, "invalid"},
		"not a number":  {"5", "good", ranged, "invalid"},
		"suspect":       {5.0, "suspect", ranged, "untrusted"},
		"stale quality": {5.0, "stale", ranged, "stale"},
	} {
		if got := readingState(LiveReading{Value: c.value, Quality: c.quality}, c.template); got != c.want {
			t.Errorf("%s = %s, want %s", name, got, c.want)
		}
	}
}

func TestTheRoutedCommandWindowAndSequenceTablesAreBounded(t *testing.T) {
	t.Parallel()
	var ledger routeLedger
	for i := range routedWindow + 5 {
		ledger.noteCommand(Inbound{Message: Message{Type: "DCMD", CommandID: fmt.Sprintf("c-%d", i)}}, OriginGateway)
	}
	if ledger.routedTo("c-0", OriginGateway) || !ledger.routedTo(fmt.Sprintf("c-%d", routedWindow+4), OriginGateway) || len(ledger.order) != routedWindow {
		t.Fatalf("route window = %d", len(ledger.order))
	}
	var table seqTable
	var syncs syncLedger
	for i := range uint64(syncWindow + 10) {
		table.add(i, i+100)
		syncs.add(i)
	}
	if up, ok := table.upFor(syncWindow + 9); len(table.pairs) != syncWindow || !ok || up != syncWindow+109 || len(syncs.ups) != pendingSyncs {
		t.Fatalf("tables = %d pairs, %d syncs", len(table.pairs), len(syncs.ups))
	}
}

func TestKindFactsAliasesAndTransmitterVersionAreDerivedFromTheKind(t *testing.T) {
	t.Parallel()
	kind := &model.Kind{Name: "k", Channels: []model.Channel{{Name: "a", ValueType: "number"}, {Name: "b", ValueType: "other"}}}
	if Alias(kind, "b") != 2 || Alias(kind, "zz") != 0 || !Carried(&kind.Channels[0]) || Carried(&kind.Channels[1]) || Carried(nil) {
		t.Fatal("alias or carried decision wrong")
	}
	out := NewTransmitter("s")
	out.UseVersion(2)
	out.Queue(Receipt("c", "accepted", "", "b"))
	if sent := out.Drain()[0]; sent["protocol_version"] != 2 {
		t.Fatalf("stamped version = %v", sent["protocol_version"])
	}
}

func TestEvidenceToleranceIsOneSharedFiniteCheck(t *testing.T) {
	t.Parallel()
	if !WithinTolerance(10.4, 10, 0.5) || WithinTolerance(11, 10, 0.5) || WithinTolerance(10, 10, -1) {
		t.Fatal("tolerance bounds wrong")
	}
}
