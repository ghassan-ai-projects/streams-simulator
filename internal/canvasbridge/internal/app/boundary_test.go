package app

import (
	"context"
	"errors"
	"io/fs"
	"reflect"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

func TestTheListingBoundaryScrubsARunReasonUnderEveryNameOfTheWorkspace(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.bridge.settings.WorkspaceAliases = []string{h.dir + "-resolved"}
	reason := "lstat " + h.dir + "/sockets/a.sock; dial " + h.dir + "-resolved/sockets/b.sock"
	writeRun(t, h.dir, "c-1", openedLine, `{"seed":18446744073709551615,"error":`+quoted(t, reason)+`}`)
	listing, err := boundary{h.bridge}.Workspace()
	if err != nil || len(listing.Runs) != 1 {
		t.Fatalf("listing = %+v, %v", listing, err)
	}
	run := listing.Runs[0]
	if run.IncompleteReason != "lstat sockets/a.sock; dial sockets/b.sock" || run.Seed == nil || *run.Seed != 18446744073709551615 {
		t.Fatalf("run = %+v, want the scrubbed reason and the seed intact", run)
	}
}

func TestTheListingBoundaryScrubsASessionPlanOfARecordedRun(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	plan := `{"socket":"` + h.dir + `/sockets/a.sock"}`
	journal := openedLine + `{"type":"session","stream":"observer","mode":"connected","plan":` + plan + "}\n" + closedLine
	h.write(t, "runs/c-1/journal.jsonl", journal)
	listing, err := boundary{h.bridge}.Workspace()
	if err != nil || len(listing.Runs) != 1 || string(listing.Runs[0].Plan) != `{"socket":"sockets/a.sock"}` {
		t.Fatalf("listing = %+v, %v", listing, err)
	}
}

func TestTheFileBoundaryScrubsErrorsButServesFileContentVerbatim(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	content := `{"note":"` + h.dir + `/x"}`
	h.write(t, inDomains("a.domain.json"), content)
	document, err := boundary{h.bridge}.ReadFile(inDomains("a.domain.json"))
	if err != nil || document.Content != content || document.Etag != domain.Version([]byte(content)) {
		t.Fatalf("ReadFile = %+v, %v, want the content as it is on disk", document, err)
	}
	current, err := boundary{h.bridge}.WriteFile(inDomains("a.domain.json"), "{}", "stale")
	if !errors.Is(err, domain.ErrConflict) || current.Etag != document.Etag {
		t.Fatalf("WriteFile with a stale etag = %+v, %v, want the conflict and the current etag kept", current, err)
	}
}

func TestTheStreamBoundaryScrubsTheFailureOfAFollow(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.write(t, "runs/c-1/journal.jsonl", openedLine)
	failing := errors.New("write " + h.dir + "/runs/c-1: broken pipe")
	err := boundary{h.bridge}.Follow(context.Background(), "c-1", -1, func(domain.Record) error { return failing })
	if err == nil || strings.Contains(err.Error(), h.dir) || !errors.Is(err, failing) {
		t.Fatalf("Follow = %v, want the workspace hidden and the cause kept", err)
	}
	if err := (boundary{h.bridge}).OpenRun("c-9"); !errors.Is(err, domain.ErrAbsent) {
		t.Fatalf("OpenRun of an absent run = %v, want ErrAbsent kept", err)
	}
}

func TestTheListingBoundaryHidesTheWorkspaceInAFailureAndKeepsItsCause(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.blockFolder(t, "runs")
	listing, err := boundary{h.bridge}.Workspace()
	var pathError *fs.PathError
	if !errors.As(err, &pathError) || strings.Contains(err.Error(), h.dir) || !strings.Contains(err.Error(), "canvasbridge: list runs") {
		t.Fatalf("Workspace = %v, want the list failure of runs with the workspace path removed", err)
	}
	if !reflect.DeepEqual(listing, domain.Listing{}) {
		t.Fatalf("listing = %+v, want none with a failure", listing)
	}
}
