package domain

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func twDCMD(t *testing.T, id, device string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(testsupport.RepositoryRoot(), "contracts", "plant-wire", "fixtures", "valid", "DCMD.json"))
	if err != nil {
		t.Fatal(err)
	}
	record := twDecode(t, raw)
	record["command_id"], record["device"] = id, device
	return record
}

func TestTheHostsCommandsAreReadFromTheJournal(t *testing.T) {
	t.Parallel()
	journal := twJournalLine(TraceHost, twDCMD(t, "c-1", "P1")) + twJournalLine(TraceHost, twDCMD(t, "c-2", "P2"))
	commands, err := ReadHostCommands([]byte(journal))
	if err != nil {
		t.Fatal(err)
	}
	if len(commands) != 2 || commands[0].ID != "c-1" || commands[0].Device != "P1" || commands[0].Name != "set_speed" ||
		commands[1].ID != "c-2" || commands[1].Device != "P2" {
		t.Fatalf("commands = %+v, want c-1 on P1 and c-2 on P2, both set_speed", commands)
	}
}

func TestNodeAndGatewayRecordsAreNotHostCommands(t *testing.T) {
	t.Parallel()
	journal := twJournalLine(TraceNode, twDCMD(t, "c-1", "P1")) + twJournalLine(TraceGateway, twDCMD(t, "c-2", "P1"))
	commands, err := ReadHostCommands([]byte(journal))
	if err != nil || len(commands) != 0 {
		t.Fatalf("ReadHostCommands = %+v, %v, want no commands from the node or the gateway", commands, err)
	}
}

func TestARepeatedCommandIdKeepsItsFirstRecord(t *testing.T) {
	t.Parallel()
	journal := twJournalLine(TraceHost, twDCMD(t, "c-1", "P1")) + twJournalLine(TraceHost, twDCMD(t, "c-1", "P2"))
	commands, err := ReadHostCommands([]byte(journal))
	if err != nil || len(commands) != 1 || commands[0].Device != "P1" {
		t.Fatalf("ReadHostCommands = %+v, %v, want the first c-1 on P1 only", commands, err)
	}
}

func TestATornLastLineIsIgnored(t *testing.T) {
	t.Parallel()
	whole := twJournalLine(TraceHost, twDCMD(t, "c-1", "P1"))
	torn := twJournalLine(TraceHost, twDCMD(t, "c-2", "P1"))
	commands, err := ReadHostCommands([]byte(whole + torn[:len(torn)-12]))
	if err != nil || len(commands) != 1 || commands[0].ID != "c-1" {
		t.Fatalf("ReadHostCommands = %+v, %v, want only the complete c-1 line", commands, err)
	}
	if empty, err := ReadHostCommands(nil); err != nil || len(empty) != 0 {
		t.Fatalf("ReadHostCommands(nil) = %+v, %v, want nothing", empty, err)
	}
}

func TestAMalformedCommandLineIsRefusedAtItsLineNumber(t *testing.T) {
	t.Parallel()
	journal := twJournalLine(TraceHost, twDCMD(t, "c-1", "P1")) + `{"type":"wire","direction":"host","record":{"type":"DCMD"` + "\n"
	_, err := ReadHostCommands([]byte(journal))
	if !errors.Is(err, ErrTwinJournal) || !strings.Contains(err.Error(), "line 2") {
		t.Fatalf("ReadHostCommands = %v, want ErrTwinJournal naming line 2", err)
	}
}

func TestLinesWithoutACommandAreSkippedWithoutError(t *testing.T) {
	t.Parallel()
	journal := "not json at all\n" + twJournalLine(TraceHost, map[string]any{"type": "DDATA", "device": "P1"})
	commands, err := ReadHostCommands([]byte(journal))
	if err != nil || len(commands) != 0 {
		t.Fatalf("ReadHostCommands = %+v, %v, want no commands and no error", commands, err)
	}
}

func twVerdict(id, verdict, reason string) map[string]any {
	return map[string]any{"event": EventVerdict, "command_id": id, "verdict": verdict, "reason": reason}
}

func twResolve(id, resolution, reason string, waiting ...string) map[string]any {
	list := []any{}
	for _, w := range waiting {
		list = append(list, w)
	}
	return map[string]any{"event": EventResolve, "command_id": id, "resolution": resolution, "reason": reason, "waiting": list}
}

func TestInterruptBlocksFollowTheVerdictAndResolutionRecords(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		log  []map[string]any
		want map[string]InterruptBlock
	}{
		{"pass blocks nothing", []map[string]any{twVerdict("c-1", VerdictPass, "")}, map[string]InterruptBlock{}},
		{"refuse blocks with its reason", []map[string]any{twVerdict("c-1", VerdictRefuse, "local_control")},
			map[string]InterruptBlock{"c-1": {Reason: "local_control"}}},
		{"queue blocks unresolved", []map[string]any{twVerdict("c-1", VerdictQueue, "")},
			map[string]InterruptBlock{"c-1": {Reason: "queued", Unresolved: true}}},
		{"queue then forward unblocks", []map[string]any{twVerdict("c-1", VerdictQueue, ""), twResolve("c-1", ResolutionForward, "")},
			map[string]InterruptBlock{}},
		{"queue then refuse blocks with the resolution's reason", []map[string]any{twVerdict("c-1", VerdictQueue, ""),
			twResolve("c-1", ResolutionRefuse, "operator_stop")}, map[string]InterruptBlock{"c-1": {Reason: "operator_stop"}}},
		{"a refused resolution repeats for the commands waiting behind it", []map[string]any{
			twVerdict("c-1", VerdictQueue, ""), twVerdict("c-2", VerdictQueue, ""), twVerdict("c-3", VerdictQueue, ""),
			twResolve("c-1", ResolutionRefuse, "operator_stop", "c-2", "c-3")},
			map[string]InterruptBlock{"c-1": {Reason: "operator_stop"}, "c-2": {Reason: "operator_stop"}, "c-3": {Reason: "operator_stop"}}},
		{"a later pass clears a refusal", []map[string]any{twVerdict("c-1", VerdictRefuse, "local_control"),
			twVerdict("c-1", VerdictPass, "")}, map[string]InterruptBlock{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := InterruptBlocks(tt.log)
			if len(got) != len(tt.want) {
				t.Fatalf("blocks = %+v, want %+v", got, tt.want)
			}
			for id, block := range tt.want {
				if got[id] != block {
					t.Errorf("block of %s = %+v, want %+v", id, got[id], block)
				}
			}
		})
	}
}
