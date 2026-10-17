package domain

import (
	"errors"
	"testing"
)

func TestARunJournalIsNamedByOneRunIdAndNeverReachesASealedFile(t *testing.T) {
	t.Parallel()
	path, err := RunJournal("w-3")
	if err != nil || path.Rel != "runs/w-3/journal.jsonl" || path.RunFile != "journal.jsonl" || path.RunID != "w-3" {
		t.Fatalf("RunJournal = %+v, %v", path, err)
	}
	for _, id := range []string{"w-3/sealed", "w-3/../w-3", "..", ".", "", "w 3", "w-3\x00", "w-3%2Fsealed", "/w-3"} {
		if _, err := RunJournal(id); !errors.Is(err, ErrBadPath) {
			t.Errorf("RunJournal(%q) = %v, want ErrBadPath", id, err)
		}
	}
}

func TestAJournalIsClosedOnlyWhenItsLastRecordClosesTheRun(t *testing.T) {
	t.Parallel()
	for tail, want := range map[string]bool{
		"": false, "\n": false, `{"type":"run.closed"}`: true, `{"type":"run.closed"}` + "\n": true,
		`{"type":"run.opened"}` + "\n" + `{"type":"run.closed"}` + "\n":  true,
		`{"type":"run.closed"}` + "\n" + `{"type":"observation"}` + "\n": false,
		`{"type":"observation"}` + "\n":                                  false, `partial {"type":"run.closed"}`: false, `{"type":"run.clo`: false,
	} {
		if got := SummarizeRun([]byte(`{"type":"run.opened"}`), []byte(tail)).Closed; got != want {
			t.Errorf("SummarizeRun closed for tail %q = %v, want %v", tail, got, want)
		}
	}
}

func TestWorkspacePathsAreScrubbedFromMessages(t *testing.T) {
	t.Parallel()
	root := "/work/space"
	for text, want := range map[string]string{
		"open /work/space/" + inDomains("x.domain.json") + ": denied": "open " + inDomains("x.domain.json") + ": denied",
		"in /work/space":  "in .",
		"nothing to hide": "nothing to hide",
	} {
		if got := Scrub(text, root); got != want {
			t.Errorf("Scrub(%q) = %q, want %q", text, got, want)
		}
	}
	if Scrub("/a/b", "") != "/a/b" {
		t.Error("an empty root scrubs nothing")
	}
}

func TestOnlyCompositionTwinAndScenarioExportGuardTheirOutputWithAVersion(t *testing.T) {
	t.Parallel()
	for _, action := range Actions() {
		if action.Etag != (action.Name == "domain.compose" || action.Name == "kind.twin" || action.Name == "scenario.export") {
			t.Errorf("%s: Etag = %v", action.Name, action.Etag)
		}
	}
	compose, _ := Lookup("domain.compose")
	args := map[string]any{"composition": "compositions/f.composition.json", "out": inDomains("f.domain.json"), "etag": "abc"}
	guard, guarded, err := compose.Guarded(args)
	if err != nil || !guarded || guard != (Guard{Path: inDomains("f.domain.json"), Etag: "abc"}) {
		t.Fatalf("Guarded = %+v, %v, %v", guard, guarded, err)
	}
	if parts, err := compose.Argv(args); err != nil || len(parts) != 5 {
		t.Fatalf("the etag must not reach argv: %+v, %v", parts, err)
	}
	delete(args, "etag")
	if guard, guarded, _ := compose.Guarded(args); !guarded || guard.Etag != "" {
		t.Fatalf("a missing etag names the file as absent: %+v", guard)
	}
	args["etag"] = 7.0
	if _, _, err := compose.Guarded(args); !errors.Is(err, ErrBadArguments) {
		t.Fatalf("a numeric etag: %v", err)
	}
	validate, _ := Lookup("domain.validate")
	if _, guarded, _ := validate.Guarded(map[string]any{"path": inDomains("f.domain.json"), "etag": "x"}); guarded {
		t.Fatal("validate writes nothing")
	}
	if _, err := validate.Argv(map[string]any{"path": inDomains("f.domain.json"), "etag": "x"}); !errors.Is(err, ErrBadArguments) {
		t.Fatalf("validate takes no etag: %v", err)
	}
}

func TestAConflictNamesItsFileAndIsAConflict(t *testing.T) {
	t.Parallel()
	err := error(&ConflictError{Path: inDomains("a.domain.json"), Etag: "e"})
	if !errors.Is(err, ErrConflict) || errors.Is(err, ErrAbsent) || err.Error() == "" {
		t.Fatalf("ConflictError = %v", err)
	}
}
