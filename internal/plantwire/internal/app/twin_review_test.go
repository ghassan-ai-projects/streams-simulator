package app

import (
	"errors"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

var twinFanDevice = []domain.InterruptDevice{{Device: "fan-01", Board: "mega-01", Clock: domain.ClockWall, SafeStop: "stop_fan"},
	{Device: "zone-01", Board: "mega-01", Clock: domain.ClockWall}}

func logRecords(entries ...domain.InterruptEntry) []map[string]any {
	records := make([]map[string]any, len(entries))
	for i, entry := range entries {
		records[i] = map[string]any(entry)
	}
	return records
}

func fanCommandIn(id string) domain.Inbound {
	message := domain.Message{Type: "DCMD", CommandID: id, Device: "fan-01", Command: "set_pwm_lease", ExpectedBootID: "boot-1",
		Args: map[string]any{"duty_permille": 300, "lease_ms": 10000}, DeadlineNS: twinEndNS}
	message.IdempotencyKey = domain.CommandOf(message).Content()
	return domain.Inbound{Message: message}
}

func heldFanLog(t *testing.T, commandID string) []map[string]any {
	t.Helper()
	g := domain.NewInterruptGate()
	hold, err := domain.DecodeControl([]byte(`{"op":"hold","scope":{"kind":"device","id":"fan-01"}}`))
	if err != nil {
		t.Fatal(err)
	}
	entries := g.SetInventory(twinFanDevice, domain.ClockWall, twinStartNS).Entries
	entries = append(entries, g.Control(hold, domain.ClockWall, twinStartNS).Entries...)
	return logRecords(append(entries, g.Judge(fanCommandIn(commandID), domain.ClockWall, twinStartNS).Entries...)...)
}

func estopLog(t *testing.T, declared string, answered ...string) []map[string]any {
	t.Helper()
	g := domain.NewInterruptGate()
	stop, err := domain.DecodeControl([]byte(`{"op":"estop","scope":{"kind":"device","id":"fan-01"}}`))
	if err != nil {
		t.Fatal(err)
	}
	devices := []domain.InterruptDevice{{Device: "fan-01", Board: "mega-01", Clock: domain.ClockWall, SafeStop: declared}}
	entries := g.SetInventory(devices, domain.ClockWall, twinStartNS).Entries
	entries = append(entries, g.Control(stop, domain.ClockWall, twinStartNS).Entries...)
	entries = append(entries, domain.EffectEntry("node-stop-1", "fan-01", "ordered", domain.ClockWall, twinStartNS))
	for _, outcome := range answered {
		entries = append(entries, domain.EffectEntry("node-stop-1", "fan-01", outcome, domain.ClockWall, twinStartNS))
	}
	return logRecords(entries...)
}

// stoppedFolder is a bench run in which the host renewed the fan lease until stoppedAt, the operator then
// emergency-stopped the fan and the board followed: it relaxes from the temperature it had at that moment.
func stoppedFolder(t *testing.T, status string, interrupts []map[string]any) string {
	t.Helper()
	stoppedAt := twinLeaseAt + 120*twinTickNS
	var leases []twinLease
	for _, lease := range twinRenewals("accepted") {
		if lease.AtNS < stoppedAt {
			leases = append(leases, lease)
		}
	}
	atStop := twinClosedForm(stoppedAt)
	relax := func(at int64) float64 {
		if at < stoppedAt {
			return 0
		}
		return 34 + (atStop-34)*math.Exp(-float64(at-stoppedAt)/1e9/180) - twinClosedForm(at)
	}
	field, host := twinRecords(t, relax, append(leases, twinLease{ID: "node-stop-1", AtNS: stoppedAt, Status: status}))
	return writeTwinFolder(t, field, host[:len(host)-1], interrupts...)
}

func nodeCommandOf(t *testing.T, report domain.TwinReport) domain.TwinCommand {
	t.Helper()
	for _, command := range report.Commands {
		if command.Origin == domain.OriginNode {
			return command
		}
	}
	t.Fatalf("commands = %+v, want a command of the node", report.Commands)
	return domain.TwinCommand{}
}

func TestAnEmergencyStopOnTheRealBoardIsAppliedToTheTwinSoTheyAgreeAboutIt(t *testing.T) {
	t.Parallel()
	report := twinJudge(t, stoppedFolder(t, "accepted", estopLog(t, "stop_fan", "accepted", "succeeded")))
	stop := nodeCommandOf(t, report)
	if stop.ID != "node-stop-1" || stop.Device != "fan-01" || stop.Name != "stop_fan" || stop.Status != domain.CommandApplied || stop.AtNS == 0 {
		t.Fatalf("stop = %+v, want the node's stop applied at the board's time of the gateway's receipt", stop)
	}
	if report.Verdict != domain.VerdictAgrees || report.Reason != "" {
		t.Fatalf("verdict %s reason %q, want a twin that stopped when the board did to agree with it", report.Verdict, report.Reason)
	}
}

func TestAStopTheGatewayNeverAnsweredMakesTheReportCannotJudgeWithThatReason(t *testing.T) {
	t.Parallel()
	report := twinJudge(t, stoppedFolder(t, "", estopLog(t, "stop_fan")))
	stop := nodeCommandOf(t, report)
	if stop.Status != domain.CommandNotApplied || stop.Reason != domain.NotAppliedNeverForwarded {
		t.Fatalf("stop = %+v, want it listed as never answered", stop)
	}
	if report.Verdict != domain.VerdictCannotJudge || report.Reason != domain.ReasonNodeStopUnconfirmed {
		t.Fatalf("verdict %s reason %q, want cannot_judge for node_stop_unconfirmed, not a divergence with the wrong cause", report.Verdict, report.Reason)
	}
}

func TestAStopTheTwinCannotApplyMakesTheReportCannotJudgeWithThatReason(t *testing.T) {
	t.Parallel()
	report := twinJudge(t, stoppedFolder(t, "accepted", estopLog(t, "stop_nothing", "accepted")))
	if stop := nodeCommandOf(t, report); stop.Status != domain.CommandTwinRefused {
		t.Fatalf("stop = %+v, want the twin to refuse a stop it has no effector for", stop)
	}
	if report.Verdict != domain.VerdictCannotJudge || report.Reason != domain.ReasonNodeStopNotApplied {
		t.Fatalf("verdict %s reason %q, want cannot_judge for node_stop_not_applied", report.Verdict, report.Reason)
	}
}

func TestAnInterruptLogThatDoesNotReplayIsRefusedBeforeTheReportUsesIt(t *testing.T) {
	t.Parallel()
	field, host := twinRecords(t, noOffset, twinRenewals("accepted"))
	forged := []map[string]any{{"event": domain.EventVerdict, "command_id": "c-1", "verdict": domain.VerdictRefuse, "reason": "local_control",
		"clock": domain.ClockWall, "node_ns": twinStartNS}}
	_, err := TwinReport(t.Context(), twinOptions(t, writeTwinFolder(t, field, host, forged...)))
	if !errors.Is(err, domain.ErrInterruptReplay) {
		t.Fatalf("TwinReport = %v, want domain.ErrInterruptReplay: a verdict no input explains is not evidence", err)
	}
}

func TestAnInterruptLogTheArtifactRecordsButTheFolderLacksIsAnErrorAndNotNoInterrupts(t *testing.T) {
	t.Parallel()
	field, _ := twinRecords(t, noOffset, twinRenewals("accepted"))
	dir := twinArtifactFolder(t, field, true, heldFanLog(t, "c-1")...)
	if err := os.Remove(filepath.Join(dir, model.FileInterruptLog)); err != nil {
		t.Fatal(err)
	}
	if _, err := TwinReport(t.Context(), twinOptions(t, dir)); !errors.Is(err, ErrTwinNoInterruptLog) {
		t.Fatalf("TwinReport = %v, want ErrTwinNoInterruptLog", err)
	}
}

func TestAFolderWhoseInterruptLogWasSetAsideAsUnreplayableIsRefusedBecauseTheStopsAreUnknown(t *testing.T) {
	t.Parallel()
	field, host := twinRecords(t, noOffset, twinRenewals("accepted"))
	dir := writeTwinFolder(t, field, host)
	testsupport.WriteFileIn(t, dir, model.FileInterruptLogRejected, []byte("{}\n"))
	if _, err := TwinReport(t.Context(), twinOptions(t, dir)); !errors.Is(err, ErrTwinInterruptLogRejected) {
		t.Fatalf("TwinReport = %v, want ErrTwinInterruptLogRejected rather than a report that ignores the stops", err)
	}
}

func TestTheReportSaysWhetherItsEvidenceIsBoundToARunArtifact(t *testing.T) {
	t.Parallel()
	field, _ := twinRecords(t, noOffset, twinRenewals("accepted"))
	if bound := twinJudge(t, twinArtifactFolder(t, field, true)).Run.EvidenceBound; !bound {
		t.Fatal("a folder with run.json recording the digests is reported unbound")
	}
	if bound := twinJudge(t, agreeingFolder(t)).Run.EvidenceBound; bound {
		t.Fatal("a folder with no run.json is reported bound: nothing it holds is covered by a digest")
	}
}

func TestAJournalThatIsASymbolicLinkIsRefusedInsteadOfBeingReadWhateverItPointsAt(t *testing.T) {
	t.Parallel()
	dir := agreeingFolder(t)
	journal := filepath.Join(dir, model.FileObserverJournal)
	if err := os.Remove(journal); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(dir, model.FileFieldTrace), journal); err != nil {
		t.Fatal(err)
	}
	if _, err := TwinReport(t.Context(), twinOptions(t, dir)); !errors.Is(err, kernel.ErrNotRegular) {
		t.Fatalf("TwinReport = %v, want kernel.ErrNotRegular", err)
	}
}
