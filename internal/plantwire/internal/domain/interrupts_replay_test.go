package domain

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func scriptedLog(t *testing.T) []map[string]any {
	t.Helper()
	g := NewInterruptGate()
	var log []InterruptEntry
	take := func(o GateOutcome) { log = append(log, o.Entries...) }
	take(g.SetInventory(plantDevices, ClockWorld, nowNS))
	take(apply(t, g, approveModeLine(ScopeBoard, "b-1", true)))
	take(judge(g, "c-1", "pump-1"))
	take(judge(g, "c-2", "pump-2"))
	take(apply(t, g, `{"op":"approve","command_id":"c-1"}`))
	take(apply(t, g, scopeLine(OpEStop, ScopeDevice, "pump-2")))
	take(judge(g, "c-3", "valve-1"))
	take(g.Advance(ClockWorld, deadlineNS+1))
	take(g.Control(controlOf(t, scopeLine(OpRelease, ScopeDevice, "pump-2")), ClockWorld, nowNS+5))
	records := make([]map[string]any, len(log))
	for i, entry := range log {
		records[i] = parsedRecord(t, entry)
	}
	return records
}

func parsedRecord(t *testing.T, entry InterruptEntry) map[string]any {
	t.Helper()
	raw := []byte(canonicalText(entry))
	var record map[string]any
	if err := decode(raw, &record); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}
	return record
}

func TestTheRulesDecideTheSameWayEveryTimeTheyAreGivenTheSameInputs(t *testing.T) {
	t.Parallel()
	first, second := scriptedLog(t), scriptedLog(t)
	if len(first) < 10 || len(first) != len(second) {
		t.Fatalf("logs of %d and %d records, want two equal logs of at least ten", len(first), len(second))
	}
	for i := range first {
		if !sameEntry(first[i], second[i]) {
			t.Fatalf("record %d differs: %s and %s", i, canonicalText(first[i]), canonicalText(second[i]))
		}
	}
}

func TestReplayingTheInputsOfALogReproducesItsDecisionsAndRefusesAnyOtherLog(t *testing.T) {
	t.Parallel()
	log := scriptedLog(t)
	if err := ReplayInterrupts(log); err != nil {
		t.Fatalf("ReplayInterrupts = %v, want the log to replay", err)
	}
	tampers := map[string]func([]map[string]any) []map[string]any{
		"a verdict changed": func(l []map[string]any) []map[string]any {
			return editFirst(l, EventVerdict, func(r map[string]any) { r["verdict"] = VerdictPass })
		},
		"a decision removed": func(l []map[string]any) []map[string]any {
			index := slices.IndexFunc(l, func(r map[string]any) bool { return r["event"] == EventReply })
			return slices.Delete(slices.Clone(l), index, index+1)
		},
		"a decision invented": func(l []map[string]any) []map[string]any {
			invented := map[string]any{"event": EventStop, "device": "valve-1", "command": "halt", "clock": ClockWorld, "node_ns": nowNS}
			return slices.Insert(slices.Clone(l), len(l), invented)
		},
		"an input changed": func(l []map[string]any) []map[string]any {
			return editFirst(l, EventCommand, func(r map[string]any) { r["device"] = "valve-1" })
		},
	}
	for name, tamper := range tampers {
		err := ReplayInterrupts(tamper(cloneRecords(log)))
		if !errors.Is(err, ErrInterruptReplay) || !strings.Contains(err.Error(), "record") {
			t.Errorf("%s: ReplayInterrupts = %v, want ErrInterruptReplay naming the record", name, err)
		}
	}
}

func TestAnEffectRecordIsEvidenceAndIsNotReplayedButADecisionWithNoInputIsRefused(t *testing.T) {
	t.Parallel()
	log := scriptedLog(t)
	effect := EffectEntry("node-stop-1", "pump-2", "accepted", ClockWorld, nowNS)
	at := slices.IndexFunc(log, func(r map[string]any) bool { return r["event"] == EventStop }) + 1
	withEffect := slices.Insert(slices.Clone(log), at, parsedRecord(t, effect))
	if err := ReplayInterrupts(withEffect); err != nil {
		t.Fatalf("a log with an effect = %v, want it to replay", err)
	}
	if err := ReplayInterrupts(nil); err != nil {
		t.Fatalf("an empty log = %v", err)
	}
	orphan := []map[string]any{parsedRecord(t, newEntry(EventVerdict, "command_id", "c-1", "verdict", VerdictPass).at(ClockWorld, nowNS))}
	if err := ReplayInterrupts(orphan); !errors.Is(err, ErrInterruptReplay) || !strings.Contains(err.Error(), "no input before it") {
		t.Fatalf("a decision first = %v, want it refused for having no input", err)
	}
}

func editFirst(records []map[string]any, event string, edit func(map[string]any)) []map[string]any {
	for _, record := range records {
		if record["event"] == event {
			edit(record)
			break
		}
	}
	return records
}

func cloneRecords(records []map[string]any) []map[string]any {
	copied := make([]map[string]any, len(records))
	for i, record := range records {
		copied[i] = map[string]any{}
		for key, value := range record {
			copied[i][key] = value
		}
	}
	return copied
}

func approvedThenRestrictedLog(t *testing.T, restrict string) []map[string]any {
	t.Helper()
	g := NewInterruptGate()
	var log []InterruptEntry
	take := func(o GateOutcome) { log = append(log, o.Entries...) }
	take(g.SetInventory(plantDevices, ClockWorld, nowNS))
	take(apply(t, g, approveModeLine(ScopeDevice, "pump-1", true)))
	take(judge(g, "c-1", "pump-1"))
	take(apply(t, g, `{"op":"approve","command_id":"c-1"}`))
	if restrict != "" {
		take(apply(t, g, restrict))
	}
	log = append(log, PurgeEntry("c-1", "pump-1", ReasonLocalControl, ClockWorld, nowNS))
	records := make([]map[string]any, len(log))
	for i, entry := range log {
		records[i] = parsedRecord(t, entry)
	}
	return records
}

func TestTheRefusalOfAnApprovedCommandThatWaitedForItsStartTimeIsEvidenceOnceALaterHoldCoversIt(t *testing.T) {
	t.Parallel()
	if err := ReplayInterrupts(approvedThenRestrictedLog(t, scopeLine(OpHold, ScopeDevice, "pump-1"))); err != nil {
		t.Fatalf("ReplayInterrupts = %v, want the purge of a command the approval forwarded to replay", err)
	}
}

func TestThePurgeOfAnApprovedCommandWithNoRestrictionAfterItsApprovalIsRefused(t *testing.T) {
	t.Parallel()
	err := ReplayInterrupts(approvedThenRestrictedLog(t, ""))
	if !errors.Is(err, ErrInterruptReplay) || !strings.Contains(err.Error(), "covers") {
		t.Fatalf("ReplayInterrupts = %v, want a purge that nothing since the approval explains refused", err)
	}
}
