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

func ReplayInterrupts(records []map[string]any) error {
	machine := NewInterrupts()
	for at := 0; at < len(records); {
		input := InterruptEntry(records[at])
		if !isInterruptInput(input.Event()) {
			return fmt.Errorf("%w: record %d is a %q with no input before it", ErrInterruptReplay, at, input.Event())
		}
		end := decisionsEnd(records, at+1)
		if err := sameDecisions(at, machine.Step(input), records[at+1:end]); err != nil {
			return err
		}
		at = end
	}
	return nil
}

func decisionsEnd(records []map[string]any, from int) int {
	end := from
	for end < len(records) && !isInterruptInput(InterruptEntry(records[end]).Event()) {
		end++
	}
	return end
}

func sameDecisions(inputAt int, want []InterruptEntry, logged []map[string]any) error {
	got := slices.DeleteFunc(slices.Clone(logged), func(record map[string]any) bool { return text(record, "event") == EventEffect })
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
