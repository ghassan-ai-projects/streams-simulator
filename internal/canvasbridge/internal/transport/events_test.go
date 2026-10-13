package transport

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

func record(index int64, line string) domain.Record {
	return domain.Record{Index: index, Type: "observation", Line: []byte(line)}
}

func TestARunIsStreamedAsServerSentEventsAndClosesAfterTheClosingRecord(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	r.service.records = []domain.Record{record(3, `{"type":"observation","seq":3}`), record(4, `{"type":"observation","seq":4}`),
		{Index: 5, Type: "run.closed", Line: []byte(`{"type":"run.closed"}`), Closes: true}}
	got := r.do(t, req{Path: "/api/runs/w-7/events", Paired: true})
	want := "id: 3\nevent: record\ndata: {\"type\":\"observation\",\"seq\":3}\n\n" +
		"id: 4\nevent: record\ndata: {\"type\":\"observation\",\"seq\":4}\n\n" +
		"id: 5\nevent: record\ndata: {\"type\":\"run.closed\"}\n\n" +
		"event: closed\ndata: {}\n\n"
	if got.Status != http.StatusOK || got.Body != want || got.Header.Get("Content-Type") != "text/event-stream" {
		t.Fatalf("%d %q\nwant %q", got.Status, got.Body, want)
	}
	r.service.mu.Lock()
	defer r.service.mu.Unlock()
	if r.service.world != "w-7" || r.service.after != -1 {
		t.Fatalf("followed %q after %d", r.service.world, r.service.after)
	}
}

func TestAFollowFailureIsAnEventNotAStatus(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	r.service.followErr = errors.New(`journal "broken"`)
	got := r.do(t, req{Path: "/api/runs/w-7/events", Paired: true})
	if got.Status != http.StatusOK || got.Body != "event: failure\ndata: \"journal \\\"broken\\\"\"\n\n" {
		t.Fatalf("%d %q", got.Status, got.Body)
	}
}

func TestLastEventIDResumesAfterThatRecord(t *testing.T) {
	t.Parallel()
	for header, want := range map[string]int64{"": -1, "7": 7, "0": 0, "garbage": -1, "-3": -3} {
		r := newRig(t)
		r.do(t, req{Path: "/api/runs/w-7/events", Paired: true, Header: map[string]string{"Last-Event-ID": header}})
		r.service.mu.Lock()
		if r.service.after != want {
			t.Errorf("Last-Event-ID %q resumes after %d, want %d", header, r.service.after, want)
		}
		r.service.mu.Unlock()
	}
}

func TestTheWorkspaceStreamAnnouncesEachChange(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	r.service.changes = 2
	got := r.do(t, req{Path: "/api/events", Paired: true})
	if want := strings.Repeat("event: workspace\ndata: {}\n\n", 2); got.Body != want || got.Header.Get("Content-Type") != "text/event-stream" {
		t.Fatalf("%q", got.Body)
	}
}

func TestStreamsNeedAdmissionLikeEveryOtherAPIRoute(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	for _, path := range []string{"/api/runs/w-7/events", "/api/events"} {
		got := r.do(t, req{Path: path, Paired: true, Header: map[string]string{"Sec-Fetch-Site": "cross-site"}})
		if got.Status != http.StatusForbidden || strings.Contains(got.Header.Get("Content-Type"), "event-stream") {
			t.Errorf("%s: %d", path, got.Status)
		}
	}
}

func TestAnEncodedSlashInTheRunIdIsRefusedBeforeAnyStreamOpens(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	r.service.openErr = fmt.Errorf("%w: run id", domain.ErrBadPath)
	r.service.records = []domain.Record{record(0, `{"type":"sealed.secret"}`)}
	got := r.do(t, req{Path: "/api/runs/w-2%2Fsealed/events", Paired: true})
	if got.Status != http.StatusBadRequest || got.errorCode(t) != "bad_path" || strings.Contains(got.Body, "sealed.secret") ||
		strings.Contains(got.Header.Get("Content-Type"), "event-stream") {
		t.Fatalf("%d %q", got.Status, got.Body)
	}
	if calls := r.service.called(); len(calls) != 1 || calls[0] != "OpenRun w-2/sealed" {
		t.Fatalf("the service saw %v, want the decoded id refused by OpenRun and no Follow", calls)
	}
}

func TestARunThatDoesNotExistIsNotFoundNotAnEndlessStream(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	r.service.openErr = fmt.Errorf("%w: run w-9", domain.ErrAbsent)
	got := r.do(t, req{Path: "/api/runs/w-9/events", Paired: true})
	if got.Status != http.StatusNotFound || got.errorCode(t) != "not_found" {
		t.Fatalf("%d %q", got.Status, got.Body)
	}
}

func TestAClosingRecordTheClientAlreadyHoldsOnlyTellsTheStreamItClosed(t *testing.T) {
	t.Parallel()
	r := newRig(t)
	r.service.records = []domain.Record{{Index: 5, Type: "run.closed", Line: []byte(`{"type":"run.closed"}`), Closes: true, Seen: true}}
	got := r.do(t, req{Path: "/api/runs/w-7/events", Paired: true, Header: map[string]string{"Last-Event-ID": "5"}})
	if got.Body != "event: closed\ndata: {}\n\n" {
		t.Fatalf("%q", got.Body)
	}
}
