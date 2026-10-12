package domain

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func doc(lines ...string) Document {
	return Document{Raw: []byte(strings.Join(lines, "\n") + "\n"), Present: true}
}

func command(t *testing.T, seq int64, op string, args map[string]any) model.Command {
	t.Helper()
	return model.Command{Seq: seq, AtNS: 1000, Op: op, Args: args}
}

func directorCommand(t *testing.T, seq int, c model.Command) string {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"stream": "director", "seq": seq, "type": "command", "at_ns": 1000, "command": c})
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

const (
	openedEvaluate = `{"stream":"observer","seq":0,"type":"run.opened","at_ns":0,"run_id":"r-1","sim_version":"9","mode":"evaluate","start_time_ns":1000,"entity_ids":["a","b"],"domain":{"id":"d","version":"1","digest":"sha256:aa"},"adapter":{"id":"n","version":"1","digest":"sha256:bb"}}`
	openedExplore  = `{"stream":"observer","seq":0,"type":"run.opened","at_ns":0,"run_id":"r-2","mode":"explore","start_time_ns":1000,"entity_ids":["a"]}`
	advanced       = `{"stream":"observer","seq":1,"type":"clock.advanced","at_ns":5000,"to_ns":5000}`
)

func TestAssembleFromAnObserverStreamAloneIsOpenAndHasNoTruth(t *testing.T) {
	t.Parallel()
	report, err := Assemble(Folder{Observer: doc(openedEvaluate, advanced)}, nil)
	if err != nil {
		t.Fatal(err)
	}
	run := report.Run
	if run.Complete || run.Ended || run.RunID != "r-1" || run.EndTimeNS != 5000 || run.EntityCount != 2 || report.Truth != nil {
		t.Fatalf("report = %+v truth %+v, want an open run r-1 ending at 5000 with 2 entities and no truth", run, report.Truth)
	}
	if report.Verification.Status != NotChecked || *report.Counts.Observed != 0 {
		t.Fatalf("verification %+v counts %+v, want not_checked and no observations", report.Verification, report.Counts)
	}
}

func TestAssembleReadsAnExploreRunsCommandsFromTheDirectorStream(t *testing.T) {
	t.Parallel()
	inject := command(t, 0, model.OpFaultInject, map[string]any{"entity_id": "a", "fault": "f", "onset_ns": json.Number("2000")})
	report, err := Assemble(Folder{Observer: doc(openedExplore), Director: doc(directorCommand(t, 1, inject))}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !report.Run.Unblinded || report.Run.Mode != "explore" || report.Truth == nil || len(report.Truth.Faults) != 1 || *report.Truth.Faults[0].OnsetNS != 2000 {
		t.Fatalf("report = %+v truth %+v, want an unblinded explore run with its one fault", report.Run, report.Truth)
	}
}

func TestAssembleKeepsTruthOfAnEvaluateRunOutOfTheDirectorStreamUnlessEnded(t *testing.T) {
	t.Parallel()
	inject := command(t, 0, model.OpFaultInject, map[string]any{"entity_id": "a", "fault": "f"})
	report, err := Assemble(Folder{Observer: doc(openedEvaluate), Director: doc(directorCommand(t, 1, inject))}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if report.Truth != nil {
		t.Fatalf("truth = %+v, want none for an open evaluate run", report.Truth)
	}
}

func TestAssembleNamesALedgerLineThatDoesNotParse(t *testing.T) {
	t.Parallel()
	_, err := Assemble(Folder{Observer: doc(openedExplore), Ledger: doc(`{"delivery_id":1}`, "{")}, nil)
	if err == nil || !strings.Contains(err.Error(), "ledger.jsonl line 2") {
		t.Fatalf("err = %v, want one naming ledger.jsonl line 2", err)
	}
}

func TestAssembleNamesADirectorRecordWhoseCommandIsNotACommand(t *testing.T) {
	t.Parallel()
	bad := `{"stream":"director","seq":4,"type":"command","at_ns":1,"command":"nope"}`
	_, err := Assemble(Folder{Observer: doc(openedExplore), Director: doc(bad)}, nil)
	if err == nil || !strings.Contains(err.Error(), "record 4") {
		t.Fatalf("err = %v, want one naming record 4", err)
	}
}

func TestConsumerDigestAppearsOnlyWhenTheVerdictRecordedOne(t *testing.T) {
	t.Parallel()
	with := &model.Verdict{Consumer: model.ConsumerInfo{Name: "c", Version: "1", ConfigDigest: "sha256:cc"}}
	without := &model.Verdict{Consumer: model.ConsumerInfo{Name: "c", Version: "1"}}
	for verdict, want := range map[*model.Verdict]bool{with: true, without: false} {
		report, err := Assemble(Folder{Observer: doc(openedEvaluate), Verdict: verdict}, nil)
		if err != nil {
			t.Fatal(err)
		}
		raw, _ := JSON(report)
		if got := strings.Contains(string(raw), "config_digest"); got != want {
			t.Errorf("config_digest present = %v, want %v in %s", got, want, raw)
		}
	}
}

func TestAssembleIgnoresTheLedgerAndDirectorStreamOfAnOpenEvaluateRun(t *testing.T) {
	t.Parallel()
	ledger := doc(`{"delivery_id":1,"delivered":false,"delivery_reason":"dropped_by_perturbation"}`)
	director := doc(`{"stream":"director","seq":0,"type":"run.opened","at_ns":0,"seed":7}`)
	report, err := Assemble(Folder{Observer: doc(openedEvaluate), Ledger: ledger, Director: director}, nil)
	if err != nil {
		t.Fatal(err)
	}
	counts := report.Counts
	if counts.LedgerRows != nil || counts.Dropped != nil || counts.DroppedByReason != nil || report.Run.Seed != nil || report.Truth != nil {
		t.Fatalf("report = %+v counts %+v, want none of the director-side evidence", report.Run, counts)
	}
}

func TestAssembleTakesTheSeedOfAnOpenExploreRunFromTheDirectorStream(t *testing.T) {
	t.Parallel()
	director := doc(`{"stream":"director","seq":0,"type":"run.opened","at_ns":0,"seed":7}`)
	report, err := Assemble(Folder{Observer: doc(openedExplore), Director: director}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if report.Run.Seed == nil || *report.Run.Seed != 7 {
		t.Fatalf("seed = %v, want 7", report.Run.Seed)
	}
}
