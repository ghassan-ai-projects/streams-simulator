package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/report/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func inputDirs() run.Directories {
	return run.Directories{Domains: testsupport.DomainsDir(), Adapters: testsupport.AdaptersDir()}
}

func questionCodes(report *domain.Report) []string {
	codes := []string{}
	for _, question := range report.OpenQuestions {
		codes = append(codes, question.Code)
	}
	return codes
}

func TestAReportWithoutAReplayAsksWhetherTheRunReproduces(t *testing.T) {
	t.Parallel()
	built, err := Build(t.Context(), testsupport.ReportRun())
	if err != nil {
		t.Fatal(err)
	}
	if built.Replay.Status != domain.NotChecked || !strings.Contains(strings.Join(questionCodes(built), " "), "replay_not_checked") {
		t.Fatalf("replay = %+v, questions = %v, want not_checked and the replay question", built.Replay, questionCodes(built))
	}
}

func TestAReplayOfARecordingThatReplaysToAnotherWorldDoesNotReproduceAndSaysWhy(t *testing.T) {
	t.Parallel()
	testsupport.SkipUnderRace(t, "replays one pinned run")
	dir := copyFolder(t, testsupport.ReportRun())
	built, err := BuildVerified(t.Context(), dir, inputDirs())
	if err != nil {
		t.Fatal(err)
	}
	if built.Replay.Status != domain.ReplayDiverges || built.Replay.Detail == "" || strings.Contains(built.Replay.Detail, dir) {
		t.Fatalf("replay = %+v, want does_not_reproduce with a detail free of the folder path", built.Replay)
	}
	if !strings.Contains(strings.Join(questionCodes(built), " "), "replay_does_not_reproduce") {
		t.Fatalf("questions = %v, want replay_does_not_reproduce", questionCodes(built))
	}
}

func TestARunThatHasNotEndedIsNotReplayed(t *testing.T) {
	t.Parallel()
	open := copyFolder(t, testsupport.ReportRun())
	if err := os.Remove(filepath.Join(open, "run.json")); err != nil {
		t.Fatal(err)
	}
	built, err := BuildVerified(t.Context(), open, inputDirs())
	if err != nil || built.Replay.Status != domain.NotChecked {
		t.Fatalf("replay = %+v, %v, want not_checked for a run that has not ended", built.Replay, err)
	}
}
