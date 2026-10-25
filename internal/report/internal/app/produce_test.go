package app

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	reportdomain "github.com/ghassan-ai-projects/streams-simulator/internal/report/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func realRun(t *testing.T, abandon string) string {
	t.Helper()
	spec, err := domain.Load(testsupport.Domain("aquaculture-pond"))
	if err != nil {
		t.Fatal(err)
	}
	a, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	r, err := run.New(t.Context(), run.Config{Domain: spec, Adapter: a, Seed: 4, SinkName: model.SinkInproc, TimeMode: model.TimeStepped,
		StartTimeNS: model.DefaultStartTimeNS, JournalDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(t.Context(), model.DefaultStartTimeNS+60e9, false); err != nil {
		t.Fatal(err)
	}
	if abandon != "" {
		r.Abandon(abandon)
	}
	if _, err := r.End(dir); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestProduceChoosesTheReportTheReplayedReportOrTheComparisonTheRequestNames(t *testing.T) {
	t.Parallel()
	testsupport.SkipUnderRace(t, "replays one run")
	dir := realRun(t, "")
	for name, tc := range map[string]struct {
		request Request
		want    string
	}{
		"a plain report":    {Request{Dir: dir}, "*domain.Report"},
		"a replayed report": {Request{Dir: dir, Replay: true, Directories: inputDirs()}, "*domain.Report"},
		"a comparison":      {Request{A: dir, B: dir}, "*domain.Comparison"},
	} {
		got, err := Produce(t.Context(), tc.request)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		switch document := got.(type) {
		case *reportdomain.Report:
			if tc.want != "*domain.Report" || (tc.request.Replay && document.Replay.Status != reportdomain.ReplayReproduces) {
				t.Errorf("%s: report %+v", name, document.Replay)
			}
		case *reportdomain.Comparison:
			if tc.want != "*domain.Comparison" {
				t.Errorf("%s: got a comparison", name)
			}
		}
	}
	if _, err := Produce(t.Context(), Request{A: dir, B: filepath.Join(dir, "absent")}); !errors.Is(err, fs.ErrNotExist) || !strings.Contains(err.Error(), "run B") {
		t.Fatalf("err = %v, want fs.ErrNotExist naming run B", err)
	}
}

func TestAReplayOfARunThatEndedIncompleteSaysItIsIncomplete(t *testing.T) {
	t.Parallel()
	testsupport.SkipUnderRace(t, "replays one run")
	built, err := BuildVerified(t.Context(), realRun(t, "the consumer stopped"), inputDirs())
	if err != nil {
		t.Fatal(err)
	}
	if built.Replay.Status != reportdomain.ReplayIncomplete || !strings.Contains(strings.Join(questionCodes(built), " "), "run_incomplete") {
		t.Fatalf("replay = %+v questions = %v, want incomplete and run_incomplete", built.Replay, questionCodes(built))
	}
}

func TestAReplayThatCannotFindTheDomainIsNotCheckedAndSaysWhy(t *testing.T) {
	t.Parallel()
	dir := realRun(t, "")
	editFile(t, filepath.Join(dir, "run.json"), func(raw string) string {
		var artifact map[string]any
		if err := model.DecodeBytes([]byte(raw), &artifact); err != nil {
			t.Fatal(err)
		}
		delete(artifact, "domain_spec")
		return jsonText(t, artifact)
	})
	built, err := BuildVerified(t.Context(), dir, run.Directories{Domains: t.TempDir(), Adapters: testsupport.AdaptersDir()})
	if err != nil {
		t.Fatal(err)
	}
	if built.Replay.Status != reportdomain.NotChecked || !strings.HasPrefix(built.Replay.Detail, "The replay could not run: ") || strings.Contains(built.Replay.Detail, dir) {
		t.Fatalf("replay = %+v, want not_checked naming the failure without the folder path", built.Replay)
	}
}

func TestRenderRefusesADocumentItDoesNotKnow(t *testing.T) {
	t.Parallel()
	if _, err := Render("a string", JSONFormat); !errors.Is(err, ErrUnknownDocument) {
		t.Fatalf("err = %v, want ErrUnknownDocument", err)
	}
}
