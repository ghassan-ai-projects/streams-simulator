package domain

import (
	"encoding/json"
	"errors"
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
	if _, err := runner.Process(raw, 2e9); err == nil {
		t.Fatal("quiescence failure must propagate")
	}
	if failed.got != nil {
		t.Fatal("verdict submitted before quiescence")
	}
}
