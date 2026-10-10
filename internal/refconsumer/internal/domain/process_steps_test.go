package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type failedQuiescence struct{ verdictCapture }

func (*failedQuiescence) ReportQuiesced(int64) error { return errors.New("consumer not ready") }

func TestProcessPreservesCumulativeRefeedAndQuiescenceBoundary(t *testing.T) {
	t.Parallel()
	ev := model.SimEvent{Seq: 7, EntityID: "entity", Channel: "value", EventTime: model.FormatTime(1e9), ObservedTime: model.FormatTime(1e9), Value: 1.0}
	raw, err := json.Marshal(ev)
	if err != nil {
		t.Fatal(err)
	}
	capture := &verdictCapture{}
	runner := New(DefaultConfig(), &Nameplate{}, nil, capture, "run")
	for _, trace := range [][]byte{append(raw, '\n'), append(raw, '\n'), []byte("malformed\n")} {
		verdict, err := runner.Process(trace, 2e9)
		if err != nil {
			t.Fatal(err)
		}
		if verdict.Counters["records_seen"] != 1 {
			t.Fatalf("refeed changed cumulative count: %+v", verdict.Counters)
		}
	}
	failed := &failedQuiescence{}
	runner = New(DefaultConfig(), &Nameplate{}, nil, failed, "run")
	runner.quiescence = failed
	if _, err := runner.Process(raw, 2e9); err == nil || !strings.Contains(err.Error(), "Process: report quiescence: consumer not ready") {
		t.Fatal("quiescence failure must propagate")
	}
	if failed.got != nil {
		t.Fatal("verdict submitted before quiescence")
	}
}

// A silence is grounded in the last record the silent series delivered.
func TestSilenceDetectionCitesTheLastDeliveredRecordOfTheSeries(t *testing.T) {
	t.Parallel()
	start := model.DefaultStartTimeNS
	var lines []byte
	for i, at := range []int64{start, start + 60*1e9, start + 120*1e9} {
		ev := model.SimEvent{Seq: int64(40 + i), EntityID: "e-1", Channel: "temperature",
			EventTime: model.FormatTime(at), ObservedTime: model.FormatTime(at), Value: 10.0}
		raw, err := json.Marshal(ev)
		if err != nil {
			t.Fatal(err)
		}
		lines = append(append(lines, raw...), '\n')
	}
	cfg := DefaultConfig()
	cfg.MinConsecutive = 99
	cfg.AbsenceFactor = 2
	verdict, err := New(cfg, &Nameplate{}, nil, &verdictCapture{}, "r").Process(lines, start+3600*1e9)
	if err != nil {
		t.Fatal(err)
	}
	if len(verdict.Detections) != 1 || verdict.Detections[0].Narrative != "channel silence" {
		t.Fatalf("detections = %+v", verdict.Detections)
	}
	if refs := verdict.Detections[0].EvidenceRefs; len(refs) != 1 || refs[0] != "seq:42" {
		t.Fatalf("silence evidence = %v, want seq:42", refs)
	}
}
