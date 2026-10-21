package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestARunFolderIsFreshUntilItHoldsAnyRunFile(t *testing.T) {
	t.Parallel()
	fresh := [][]string{nil, {}, {"notes.txt", ".DS_Store", "plan.json", "interrupts.rejected.txt"}}
	for _, present := range fresh {
		if err := CheckFreshRunFolder("runs/a", present); err != nil {
			t.Errorf("CheckFreshRunFolder(%v) = %v, want a folder without run files accepted", present, err)
		}
	}
	taken := []string{"run.json", "trace.jsonl", "ledger.jsonl", "world_state_history.jsonl", "field-trace.jsonl", "interrupts.jsonl",
		"interrupts.rejected.jsonl", "interrupts.rejected.2.jsonl", "journal.jsonl", "sealed/journal.jsonl", "verdict.json", "label.json"}
	for _, name := range taken {
		err := CheckFreshRunFolder("runs/a", []string{"notes.txt", name})
		if !errors.Is(err, ErrRunFolderTaken) || !strings.Contains(err.Error(), "runs/a holds "+name) {
			t.Errorf("CheckFreshRunFolder with %s = %v, want ErrRunFolderTaken naming the folder and the file", name, err)
		}
	}
}

func TestATakenRunFolderNamesTheSameFileWhateverTheListingOrder(t *testing.T) {
	t.Parallel()
	first := CheckFreshRunFolder("out", []string{"run.json", "journal.jsonl"})
	second := CheckFreshRunFolder("out", []string{"journal.jsonl", "run.json"})
	if first == nil || second == nil || first.Error() != second.Error() {
		t.Fatalf("refusals = %v and %v, want one refusal for both orders", first, second)
	}
}
