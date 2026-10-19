package app

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

func twinMetric(t *testing.T, report domain.TwinReport, device, metric string) domain.MetricReport {
	t.Helper()
	for _, candidate := range report.Metrics {
		if candidate.Device == device && candidate.Metric == metric {
			return candidate
		}
	}
	t.Fatalf("report has no %s.%s metric: %+v", device, metric, report.Metrics)
	return domain.MetricReport{}
}

func twinJudge(t *testing.T, dir string) domain.TwinReport {
	t.Helper()
	report, err := buildTwinReport(twinOptions(t, dir), nil)
	if err != nil {
		t.Fatalf("buildTwinReport: %v", err)
	}
	return report
}

func agreeingFolder(t *testing.T) string {
	t.Helper()
	field, host := twinRecords(t, noOffset, twinRenewals("accepted"))
	return writeTwinFolder(t, field, host)
}

func TestAFolderThatFollowsItsTwinIsJudgedAgreeing(t *testing.T) {
	t.Parallel()
	report := twinJudge(t, agreeingFolder(t))
	renewals := len(twinRenewals("accepted"))
	if report.Verdict != domain.VerdictAgrees || report.Reason != "" {
		t.Fatalf("verdict %s reason %q, want agrees", report.Verdict, report.Reason)
	}
	if report.Summary.CommandsApplied != renewals || report.Summary.CommandsNotApplied != 0 || report.Summary.Diverged != 0 {
		t.Fatalf("summary = %+v, want %d commands applied and nothing diverged", report.Summary, renewals)
	}
	if report.Twin.Domain != "arduino-thermal" || report.Twin.StartNS != twinStartNS || len(report.Twin.Kinds) != 3 {
		t.Fatalf("twin = %+v, want the arduino-thermal twin starting at the birth with its three kinds", report.Twin)
	}
}

func TestTheTwinsZoneFollowsTheBenchsClosedFormWithinQuantisation(t *testing.T) {
	t.Parallel()
	series := twinMetric(t, twinJudge(t, agreeingFolder(t)), "zone-01", "temperature").Series
	if len(series.Points) < 500 || series.Truncated {
		t.Fatalf("series = %d points, truncated %v, want the ten minutes of comparisons", len(series.Points), series.Truncated)
	}
	worst := 0.0
	for _, point := range series.Points {
		if want := twinClosedForm(point.AtNS); math.Abs(point.Board-want) > 1e-9 {
			t.Fatalf("board at %d = %v, the closed form says %v", point.AtNS, point.Board, want)
		}
		worst = math.Max(worst, math.Abs(point.Twin-twinClosedForm(point.AtNS)))
	}
	t.Logf("worst twin gap from the closed form: %.4f", worst)
	if worst > 0.5+0.05 {
		t.Fatalf("worst twin gap %.4f from the closed form, want within the 1 degC resolution (0.5) plus the 1 s integrator step", worst)
	}
}

func TestAStretchFarFromTheTwinDivergesWithItsIntervalAndWorstPoint(t *testing.T) {
	t.Parallel()
	offset := func(at int64) float64 {
		if at >= twinLeaseAt+200*twinTickNS && at <= twinLeaseAt+300*twinTickNS {
			return 6
		}
		return 0
	}
	field, host := twinRecords(t, offset, twinRenewals("accepted"))
	report := twinJudge(t, writeTwinFolder(t, field, host))
	if report.Verdict != domain.VerdictDiverges || len(report.Divergences) != 1 {
		t.Fatalf("verdict %s divergences %+v, want diverges with one stretch", report.Verdict, report.Divergences)
	}
	stretch := report.Divergences[0]
	if stretch.Name != "zone-01.temperature#1" || stretch.FirstNS != twinLeaseAt+200*twinTickNS || stretch.LastNS != twinLeaseAt+300*twinTickNS {
		t.Fatalf("stretch = %+v, want zone-01.temperature#1 from 200 s to 300 s after the lease", stretch)
	}
	if stretch.Readings != 101 || stretch.WorstError < 5.5 || stretch.WorstError > 6.5 {
		t.Fatalf("stretch = %+v, want 101 readings with a worst error near the 6 degC offset", stretch)
	}
	if stretch.WorstAtNS < stretch.FirstNS || stretch.WorstAtNS > stretch.LastNS || len(stretch.Commands) != 1 {
		t.Fatalf("worst point %d and commands %v, want it inside the stretch with one lease in force", stretch.WorstAtNS, stretch.Commands)
	}
}

func TestWithoutAnyCommandTheReportCannotJudge(t *testing.T) {
	t.Parallel()
	field, _ := twinRecords(t, noOffset, nil)
	report := twinJudge(t, writeTwinFolder(t, field, nil))
	if report.Verdict != domain.VerdictCannotJudge || report.Reason != domain.ReasonNoCommands || report.Summary.CommandsApplied != 0 {
		t.Fatalf("verdict %s reason %q commands %d, want cannot_judge for no_commands_applied", report.Verdict, report.Reason, report.Summary.CommandsApplied)
	}
}

func TestARefusedReceiptAndAnUnforwardedCommandAreNotApplied(t *testing.T) {
	t.Parallel()
	leases := []twinLease{
		{ID: "c-1", AtNS: twinLeaseAt, Status: "refused", Reason: "permissive_not_met"},
		{ID: "c-2", AtNS: twinLeaseAt + 5*twinTickNS},
	}
	field, host := twinRecords(t, noOffset, leases)
	report := twinJudge(t, writeTwinFolder(t, field, host))
	first, second := report.Commands[0], report.Commands[1]
	if first.ID != "c-1" || first.Status != domain.CommandNotApplied || first.Reason != domain.NotAppliedRefusedByBoard || first.Detail != "permissive_not_met" || first.AtNS != twinLeaseAt {
		t.Errorf("c-1 = %+v, want refused by the board for permissive_not_met at the lease", first)
	}
	if second.ID != "c-2" || second.Status != domain.CommandNotApplied || second.Reason != domain.NotAppliedNeverForwarded || second.AtNS != twinStartNS {
		t.Errorf("c-2 = %+v, want never forwarded, clamped to the start of the trace", second)
	}
	if report.Verdict != domain.VerdictCannotJudge || report.Reason != domain.ReasonNoCommands {
		t.Fatalf("verdict %s reason %q, want cannot_judge for no_commands_applied", report.Verdict, report.Reason)
	}
}

func TestAnInterruptRefusalExcludesACommandTheBoardAccepted(t *testing.T) {
	t.Parallel()
	field, host := twinRecords(t, noOffset, []twinLease{{ID: "c-1", AtNS: twinLeaseAt, Status: "accepted"}})
	refusal := map[string]any{"event": domain.EventVerdict, "command_id": "c-1", "verdict": domain.VerdictRefuse, "reason": "local_control"}
	report := twinJudge(t, writeTwinFolder(t, field, host, refusal))
	if len(report.Commands) != 1 || report.Commands[0].Reason != domain.NotAppliedInterruptRefused || report.Commands[0].Detail != "local_control" {
		t.Fatalf("commands = %+v, want c-1 refused by the interrupt for local_control", report.Commands)
	}
	if report.Run.InterruptLogDigest == "" {
		t.Fatal("run evidence has no interrupt log digest, want the interrupt log named")
	}
}

func TestTwoReportsOfOneFolderMarshalIdentically(t *testing.T) {
	t.Parallel()
	dir := agreeingFolder(t)
	first, err := json.Marshal(twinJudge(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	second, err := json.Marshal(twinJudge(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatal("two reports of one folder differ")
	}
}

func TestAReportWithADivergenceIsAcceptedByItsContract(t *testing.T) {
	t.Parallel()
	offset := func(at int64) float64 {
		if at >= twinLeaseAt+200*twinTickNS && at <= twinLeaseAt+300*twinTickNS {
			return 6
		}
		return 0
	}
	field, host := twinRecords(t, offset, twinRenewals("accepted"))
	report, err := TwinReport(t.Context(), twinOptions(t, writeTwinFolder(t, field, host)))
	if err != nil {
		t.Fatalf("TwinReport: %v", err)
	}
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	if err := model.ValidateTwinReport(raw); err != nil || report.Verdict != domain.VerdictDiverges {
		t.Fatalf("ValidateTwinReport = %v, verdict %s, want the diverging report accepted", err, report.Verdict)
	}
}

func TestAnAgreeingFolderPassesTwinReportsContract(t *testing.T) {
	t.Parallel()
	report, err := TwinReport(t.Context(), twinOptions(t, agreeingFolder(t)))
	if err != nil || report.Verdict != domain.VerdictAgrees {
		t.Fatalf("TwinReport = %s, %v, want an agreeing report that passes its contract", report.Verdict, err)
	}
}
