package domain

import (
	"fmt"
	"testing"
)

func TestCommandWindowReportsTruncationAndRetainsNewestCommands(t *testing.T) {
	t.Parallel()
	l := liveBoard(t)
	for i := range 129 {
		command := liveRecord(hostCommand())
		command["command_id"] = fmt.Sprintf("c-%d", i)
		l.Observe(TraceHost, command, int64(i))
	}
	if len(l.commands) != 128 || l.command("c-0") != nil || l.command("c-128") == nil || l.Fields()["commands_truncated"] != true {
		t.Fatalf("bounded command history = %d, truncated %v", len(l.commands), l.truncated)
	}
}

func TestHostEvidenceCannotContradictItsOwnExpectation(t *testing.T) {
	t.Parallel()
	for _, field := range []string{"expected", "tolerance", "observed_value"} {
		t.Run(field, func(t *testing.T) {
			t.Parallel()
			l := liveBoard(t)
			acceptedCommand(t, l)
			entry := hostEvidence()
			entry[field] = 999.0
			if field == "tolerance" {
				entry[field] = -1.0
			}
			if l.Observe(TraceHost, map[string]any{"type": "HSYNC", "command_evidence": []any{entry}}, 5) {
				t.Fatalf("contradictory %s accepted as effect evidence", field)
			}
		})
	}
}

func TestNoObservedEffectIsInventedForHostTimeout(t *testing.T) {
	t.Parallel()
	l := liveBoard(t)
	acceptedCommand(t, l)
	entry := hostEvidence()
	entry["outcome"] = "not_observed"
	delete(entry, "observed_value")
	l.Observe(TraceHost, map[string]any{"type": "HSYNC", "command_evidence": []any{entry}}, 5)
	c := l.command("c-1")
	if c.Effect != "unknown" || c.EffectEvidence != "Host checked flow: not_observed" {
		t.Fatalf("absence of feedback = %+v", c)
	}
}

func TestDeclaredTypesAndEnumsCannotBeDisplayedAsGoodValues(t *testing.T) {
	t.Parallel()
	for name, data := range map[string]struct {
		value    any
		template map[string]any
	}{
		"boolean string":     {"true", map[string]any{"data_type": "boolean"}},
		"fractional integer": {1.5, map[string]any{"data_type": "integer"}},
		"number string":      {"1.0", map[string]any{"data_type": "number"}},
		"unknown enum":       {"bad", map[string]any{"data_type": "enum", "enum_values": []any{"ready"}}},
		"unannounced metric": {20.0, nil},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if state := readingState(LiveReading{Value: data.value, Quality: "good"}, data.template); state != "invalid" {
				t.Fatalf("invalid reading state = %s", state)
			}
		})
	}
}

func TestADeviceOnAnUnannouncedBoardIsUnknownAndARepeatedLinkStateMarksOnce(t *testing.T) {
	t.Parallel()
	l := NewLive(true, nil, nil)
	l.Observe(TraceNode, liveRecord(deviceBirthOf("P1", "boot-1")), 1)
	if d := l.devices["P1"]; d.Environment != "unknown" {
		t.Fatalf("device on an unannounced board = %+v, want environment unknown", d)
	}
	l.Fields()
	l.Connection("host", "online", 2)
	l.Connection("host", "online", 3)
	marks := l.Fields()["marks"].([]LiveMark)
	if len(marks) != 1 || marks[0].Label != "Host link online" || marks[0].AtNS != 2 {
		t.Fatalf("repeated host state marks = %+v, want one change", marks)
	}
}
