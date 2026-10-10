package app

import (
	"errors"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type recordingSink struct {
	verdictErr error
	quiesced   []int64
	verdicts   int
}

func (s *recordingSink) ReportQuiesced(through int64) { s.quiesced = append(s.quiesced, through) }
func (s *recordingSink) SubmitVerdict(*model.Verdict) error {
	s.verdicts++
	return s.verdictErr
}

func reportView(sink *recordingSink) *OperatorView {
	v := NewOperatorView("w-1", "t-secret", &Nameplate{}, nil, sink)
	v.RunID = "r-1"
	return v
}

func TestReportChangesNothingWhenItsVerdictIsRefused(t *testing.T) {
	t.Parallel()
	sink := &recordingSink{verdictErr: errors.New("schema mismatch")}
	err := reportView(sink).Report("t-secret", "r-1", 500, &model.Verdict{})
	var tool *ToolError
	if !errors.As(err, &tool) || tool.Code != CodeInvalidArgs || !strings.Contains(tool.Msg, "schema mismatch") {
		t.Fatalf("err = %v", err)
	}
	if len(sink.quiesced) != 0 {
		t.Fatalf("a refused report must not advance quiescence: %v", sink.quiesced)
	}
}

func TestReportAppliesBothPartsOfAnAcceptedReport(t *testing.T) {
	t.Parallel()
	sink := &recordingSink{}
	if err := reportView(sink).Report("t-secret", "r-1", 500, &model.Verdict{}); err != nil {
		t.Fatal(err)
	}
	if sink.verdicts != 1 || len(sink.quiesced) != 1 || sink.quiesced[0] != 500 {
		t.Fatalf("verdicts=%d quiesced=%v", sink.verdicts, sink.quiesced)
	}
}

func TestReportRefusesAnotherWorldsRunAndAWrongToken(t *testing.T) {
	t.Parallel()
	sink := &recordingSink{}
	view := reportView(sink)
	var tool *ToolError
	if err := view.Report("t-secret", "r-other", 500, nil); !errors.As(err, &tool) || tool.Code != CodeInvalidArgs {
		t.Fatalf("a report naming another run: %v", err)
	}
	if err := view.Report("t-wrong", "r-1", 500, nil); !errors.As(err, &tool) || tool.Code != CodeCapabilityDenied {
		t.Fatalf("a report with the wrong token: %v", err)
	}
	if len(sink.quiesced) != 0 || sink.verdicts != 0 {
		t.Fatalf("refused reports must change nothing: %+v", sink)
	}
}
