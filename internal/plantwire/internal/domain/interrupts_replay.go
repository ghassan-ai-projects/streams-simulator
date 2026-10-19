package domain

import (
	"bytes"
	"errors"
	"fmt"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
)

var ErrInterruptReplay = errors.New("plantwire: the interrupt log does not replay")

func isInterruptInput(event string) bool {
	return slices.Contains([]string{EventInventory, EventControl, EventCommand, EventClock, EventFlush}, event)
}

func isEvidenceEvent(event string) bool { return event == EventEffect || event == EventPurge }

type evidenceLedger struct{ stopped, passed map[string]bool }

func checkEvidence(records []map[string]any) error {
	ledger := evidenceLedger{stopped: map[string]bool{}, passed: map[string]bool{}}
	for at, record := range records {
		if fault := ledger.note(InterruptEntry(record)); fault != "" {
			return fmt.Errorf("%w: record %d %s", ErrInterruptReplay, at, fault)
		}
	}
	return nil
}

func (l evidenceLedger) note(entry InterruptEntry) string {
	switch {
	case entry.Event() == EventStop:
		l.stopped[entry.Text("device")] = true
	case entry.Event() == EventVerdict && entry.Text("verdict") == VerdictPass:
		l.passed[entry.Text("command_id")] = true
	case entry.Event() == EventEffect && !l.stopped[entry.Text("device")]:
		return fmt.Sprintf("is an effect on %q, which no stop ordered", entry.Text("device"))
	case entry.Event() == EventPurge && !l.passed[entry.Text("command_id")]:
		return fmt.Sprintf("purges %q, which no verdict passed", entry.Text("command_id"))
	}
	return ""
}

func ReplayInterrupts(records []map[string]any) error {
	if err := checkEvidence(records); err != nil {
		return err
	}
	machine := NewInterrupts()
	for at := 0; at < len(records); {
		end, err := replayInput(machine, records, at)
		if err != nil {
			return err
		}
		at = end
	}
	return nil
}

func replayInput(machine *Interrupts, records []map[string]any, at int) (int, error) {
	input := InterruptEntry(records[at])
	if !isInterruptInput(input.Event()) {
		return 0, fmt.Errorf("%w: record %d is a %q with no input before it", ErrInterruptReplay, at, input.Event())
	}
	end := decisionsEnd(records, at+1)
	return end, sameDecisions(at, machine.Step(input), records[at+1:end])
}

func decisionsEnd(records []map[string]any, from int) int {
	end := from
	for end < len(records) && !isInterruptInput(InterruptEntry(records[end]).Event()) {
		end++
	}
	return end
}

func sameDecisions(inputAt int, want []InterruptEntry, logged []map[string]any) error {
	got := slices.DeleteFunc(slices.Clone(logged), func(record map[string]any) bool { return isEvidenceEvent(text(record, "event")) })
	if len(got) != len(want) {
		return fmt.Errorf("%w: after the input at record %d the rules decide %d records, the log holds %d", ErrInterruptReplay, inputAt, len(want), len(got))
	}
	for i := range want {
		if !sameEntry(want[i], got[i]) {
			return fmt.Errorf("%w: after the input at record %d the rules decide %s, the log holds %s", ErrInterruptReplay, inputAt,
				clip(canonicalText(want[i])), clip(canonicalText(got[i])))
		}
	}
	return nil
}

func sameEntry(a, b map[string]any) bool {
	left, errLeft := canonical.Marshal(a)
	right, errRight := canonical.Marshal(b)
	return errLeft == nil && errRight == nil && bytes.Equal(left, right)
}

func canonicalText(entry map[string]any) string {
	raw, err := canonical.Marshal(entry)
	if err != nil {
		return "(unreadable)"
	}
	return string(raw)
}
