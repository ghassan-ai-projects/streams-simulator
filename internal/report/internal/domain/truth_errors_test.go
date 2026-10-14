package domain

import (
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestAssembleNamesACommandRecordThatDoesNotDecode(t *testing.T) {
	t.Parallel()
	bad := `{"stream":"director","seq":1,"type":"command","at_ns":1000,"command":"inject"}`
	_, err := Assemble(Folder{Observer: doc(openedExplore), Director: doc(bad)}, nil)
	if err == nil || !strings.Contains(err.Error(), "sealed/journal.jsonl record 1") {
		t.Fatalf("err = %v, want the record named", err)
	}
}

func TestAssembleNamesAStreamThatDoesNotParse(t *testing.T) {
	t.Parallel()
	_, err := Assemble(Folder{Observer: doc(openedExplore), Director: doc("not json")}, nil)
	if err == nil || !strings.Contains(err.Error(), "sealed/journal.jsonl") {
		t.Fatalf("err = %v, want the stream named", err)
	}
}

func TestAnEffectorCallWithoutATimeTakesTheTimeOfItsCommand(t *testing.T) {
	t.Parallel()
	invoke := command(t, 0, model.OpEffectorInvoke, map[string]any{"effector": "e", "entity_id": "a", "command_id": "c1"})
	calls := effectorCalls([]model.Command{invoke}, nil)
	if len(calls) != 1 || calls[0].AtNS != invoke.AtNS {
		t.Fatalf("calls = %+v, want one at the command time %d", calls, invoke.AtNS)
	}
}

func TestIntegerOfReadsOnlyJSONNumbers(t *testing.T) {
	t.Parallel()
	if got := integerOf("7"); got != 0 {
		t.Fatalf("integerOf(string) = %d, want 0", got)
	}
}
