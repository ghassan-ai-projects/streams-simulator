package app

import (
	"context"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

func requestOf(t *testing.T, line string) domain.InterruptEntry {
	t.Helper()
	request, err := domain.DecodeControl([]byte(line))
	if err != nil {
		t.Fatalf("DecodeControl(%s) = %v", line, err)
	}
	return request
}

func TestAnEmergencyStopIsTakenBeforeEveryOrdinaryRequestThatWasWaiting(t *testing.T) {
	t.Parallel()
	port := NewControlPort()
	for _, line := range []string{holdBoard, approveFan, estopBoard} {
		if _, queued := port.enqueue(t.Context(), requestOf(t, line)); !queued {
			t.Fatalf("%s was not queued", line)
		}
	}
	first, _ := port.take()
	if first.Entry.Op() != domain.OpEStop {
		t.Fatalf("first request taken = %s, want the emergency stop ahead of the hold and the approval mode that arrived before it", first.Entry.Op())
	}
	second, _ := port.take()
	if second.Entry.Op() != domain.OpHold {
		t.Fatalf("second request taken = %s, want the ordinary requests kept in arrival order", second.Entry.Op())
	}
}

func TestAnEmergencyStopIsQueuedEvenWhenTheOrdinaryRequestsFilledTheBacklog(t *testing.T) {
	t.Parallel()
	port := NewControlPort()
	for range controlBacklog {
		port.enqueue(t.Context(), requestOf(t, holdBoard))
	}
	if _, queued := port.enqueue(t.Context(), requestOf(t, holdBoard)); queued {
		t.Fatal("a hold beyond the backlog was queued, want the ordinary list bounded")
	}
	if _, queued := port.enqueue(t.Context(), requestOf(t, estopBoard)); !queued {
		t.Fatal("the emergency stop was refused as busy, want it to have a place of its own however many ordinary requests wait")
	}
}

func TestRequestsTheirCallersGaveUpOnNeitherFillTheBacklogNorAreTaken(t *testing.T) {
	t.Parallel()
	port := NewControlPort()
	gone, cancel := context.WithCancel(t.Context())
	for range controlBacklog {
		port.enqueue(gone, requestOf(t, holdBoard))
	}
	cancel()
	if _, queued := port.enqueue(t.Context(), requestOf(t, approveFan)); !queued {
		t.Fatal("a live request was refused as busy because abandoned ones still filled the backlog")
	}
	call, taken := port.take()
	if !taken || call.Entry.Op() != domain.OpApproveMode {
		t.Fatalf("take = %+v, %t, want the one live request and none of the abandoned holds", call, taken)
	}
	if again, more := port.take(); more {
		t.Fatalf("take returned %+v after the live request, want nothing left", again)
	}
}
