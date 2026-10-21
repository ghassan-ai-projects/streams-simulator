package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func nodeStop(status, reason string) TwinCommand {
	return TwinCommand{ID: "node-stop-1", Device: "fan-01", Name: "stop_fan", AtNS: 150, Origin: OriginNode, Status: status, Reason: reason}
}

func divergingTrace() BoardTrace { return twTrace(twTemperature(300, 41), twTemperature(400, 42)) }

func TestAStopTheNodeOrderedThatTheTwinCannotApplyMakesTheReportCannotJudgeInsteadOfDiverge(t *testing.T) {
	t.Parallel()
	twin := newFakeTwin()
	twin.refuse["node-stop-1"] = errors.New("no such effector")
	report := twEvaluate(t, twInput(divergingTrace(), twApplied("c-1", 100), nodeStop(CommandApplied, "")), twin)
	if report.Verdict != VerdictCannotJudge || report.Reason != ReasonNodeStopNotApplied {
		t.Fatalf("verdict %s reason %q, want cannot_judge for node_stop_not_applied even though the readings differ", report.Verdict, report.Reason)
	}
	if got := report.Commands[1]; got.Origin != OriginNode || got.Status != CommandTwinRefused || !strings.Contains(got.Detail, "no such effector") {
		t.Fatalf("command = %+v, want the node's stop listed as refused by the twin with its cause", got)
	}
}

func TestAStopTheBoardNeverConfirmedMakesTheReportCannotJudgeBecauseWhetherItStoppedIsUnknown(t *testing.T) {
	t.Parallel()
	for _, reason := range []string{NotAppliedNeverForwarded, NotAppliedNotAccepted} {
		report := twEvaluate(t, twInput(divergingTrace(), twApplied("c-1", 100), nodeStop(CommandNotApplied, reason)), newFakeTwin())
		if report.Verdict != VerdictCannotJudge || report.Reason != ReasonNodeStopUnconfirmed {
			t.Errorf("%s: verdict %s reason %q, want cannot_judge for node_stop_unconfirmed", reason, report.Verdict, report.Reason)
		}
	}
}

func TestAStopTheBoardRefusedLeavesTheVerdictToTheReadings(t *testing.T) {
	t.Parallel()
	report := twEvaluate(t, twInput(divergingTrace(), twApplied("c-1", 100), nodeStop(CommandNotApplied, NotAppliedRefusedByBoard)), newFakeTwin())
	if report.Verdict != VerdictDiverges || report.Reason != "" {
		t.Fatalf("verdict %s reason %q, want a board that refused the stop judged by its readings", report.Verdict, report.Reason)
	}
}

func TestAnAppliedNodeStopIsAppliedToTheTwinLikeAHostCommandAndMarkedAsTheNodes(t *testing.T) {
	t.Parallel()
	twin := newFakeTwin()
	report := twEvaluate(t, twInput(twTrace(twTemperature(200, 31)), twApplied("c-1", 100), nodeStop(CommandApplied, "")), twin)
	if twin.position("apply node-stop-1") < 0 || report.Verdict != VerdictAgrees || report.Summary.CommandsApplied != 2 {
		t.Fatalf("log %v verdict %s summary %+v, want the stop applied to the twin and counted", twin.log, report.Verdict, report.Summary)
	}
	raw, err := json.Marshal(report.Commands)
	if err != nil || !strings.Contains(string(raw), `"origin":"node"`) || strings.Count(string(raw), `"origin"`) != 1 {
		t.Fatalf("commands %s (%v), want origin node on the stop alone and none on the host command", raw, err)
	}
}

func TestANodeStopIsTimedFromTheGatewaysReceiptAndAHostCommandPurgedByTheNodeIsNotApplied(t *testing.T) {
	t.Parallel()
	board := twTrace(twTemperature(100, 30), twTemperature(200, 31))
	board.Receipts = map[string]BoardReceipt{"node-stop-1": {Status: "accepted", AtNS: 200}, "c-1": {Status: "accepted", AtNS: 150}}
	sent := []HostCommand{{ID: "c-1", Device: "fan-01", Name: "set_cooling"},
		{ID: "node-stop-1", Device: "fan-01", Name: "stop_fan", Origin: OriginNode, Args: map[string]any{}}}
	blocks := InterruptBlocks([]map[string]any{{"event": EventPurge, "command_id": "c-1", "reason": ReasonLocalControl}})
	board.Devices = append(board.Devices, BoardDevice{Name: "fan-01", Kind: "fan", Board: "mega-01"})
	commands := ClassifyCommands(sent, board, blocks)
	if len(commands) != 2 || commands[0].Reason != NotAppliedInterruptRefused || commands[0].Detail != ReasonLocalControl {
		t.Fatalf("commands = %+v, want the purged host command refused by the interrupt for local_control", commands)
	}
	if stop := commands[1]; stop.Origin != OriginNode || stop.Status != CommandApplied || stop.AtNS != 200 {
		t.Fatalf("stop = %+v, want it applied at the board's time of the gateway's receipt", stop)
	}
}

func TestTheEvidenceStatesWhetherItIsBoundToARunArtifact(t *testing.T) {
	t.Parallel()
	raw, err := json.Marshal(TwinEvidence{FieldTraceDigest: "sha256:aa", JournalDigest: "sha256:bb"})
	if err != nil || !strings.Contains(string(raw), `"evidence_bound":false`) {
		t.Fatalf("evidence %s (%v), want evidence_bound written out even when false", raw, err)
	}
}
