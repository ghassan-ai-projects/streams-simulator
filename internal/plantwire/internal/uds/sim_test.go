package uds

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

func inboundOf(kind string) domain.Inbound {
	return domain.Inbound{Message: domain.Message{Type: kind}}
}

func isKind(kind string) func(domain.Message) bool {
	return func(m domain.Message) bool { return m.Type == kind }
}

func TestAMailboxKeepsItsItemsInOrderAndRefusesWhatIsOverItsLimitOrAfterItIsShut(t *testing.T) {
	t.Parallel()
	box := NewMailbox[int]()
	if !box.Put(1, 2) || !box.Put(3) {
		t.Fatal("a mailbox refused items it had room for")
	}
	if first, ok := box.Pop(); !ok || first != 1 {
		t.Fatalf("Pop = %d, %v, want the first item put", first, ok)
	}
	if rest := box.Take(); len(rest) != 2 || rest[0] != 2 || rest[1] != 3 {
		t.Fatalf("Take = %v, want the remaining items in order", rest)
	}
	if _, ok := box.Pop(); ok || len(box.Take()) != 0 {
		t.Fatal("an emptied mailbox still held items")
	}
	if box.Put(make([]int, mailboxLimit+1)...) {
		t.Fatal("a mailbox took more items than its limit")
	}
	box.Shut()
	if !box.IsShut() || box.Put(4) {
		t.Fatal("a shut mailbox took an item")
	}
}

func TestAMailboxSignalsWhenItIsWokenPutInOrShut(t *testing.T) {
	t.Parallel()
	cases := map[string]func(*Mailbox[int]){
		"put":  func(m *Mailbox[int]) { m.Put(1) },
		"wake": func(m *Mailbox[int]) { m.Wake() },
		"shut": func(m *Mailbox[int]) { m.Shut() },
	}
	for name, act := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			box := NewMailbox[int]()
			act(box)
			select {
			case <-box.Ready():
			case <-t.Context().Done():
				t.Fatal("the mailbox did not signal")
			}
		})
	}
}

func TestTheSimulatedNodesRecordsPassUpAndRefuseAfterTheLinkCloses(t *testing.T) {
	t.Parallel()
	link := NewSimLink(time.Second)
	if err := link.Send([]map[string]any{{"type": "NBIRTH"}}); err != nil {
		t.Fatal(err)
	}
	<-link.UpReady()
	if taken := link.TakeUp(); len(taken) != 1 || taken[0]["type"] != "NBIRTH" {
		t.Fatalf("TakeUp = %v, want the record sent", taken)
	}
	link.Close()
	if err := link.Send([]map[string]any{{"type": "NDATA"}}); !errors.Is(err, ErrSimClosed) {
		t.Fatalf("Send after Close = %v, want %v", err, ErrSimClosed)
	}
	if link.Deliver(inboundOf("STATE")) {
		t.Fatal("a closed link took a record for the simulated node")
	}
}

func TestTheSimulatedNodeAwaitsWhatTheRelayDeliversUntilTheConditionHolds(t *testing.T) {
	t.Parallel()
	link := NewSimLink(time.Second)
	link.Deliver(inboundOf("DCMD"))
	link.Deliver(inboundOf("HSYNC"))
	got, met, err := link.Await(t.Context(), isKind("HSYNC"))
	if err != nil || !met || len(got) != 2 || got[1].Type != "HSYNC" {
		t.Fatalf("Await = %v, %v, %v, want both records and the condition met", got, met, err)
	}
	link.Deliver(inboundOf("STATE"))
	if got, met, err := link.AwaitHandshake(t.Context(), isKind("STATE")); err != nil || !met || len(got) != 1 {
		t.Fatalf("AwaitHandshake = %v, %v, %v, want the STATE", got, met, err)
	}
	if link.WaitedMS() < 0 {
		t.Fatal("the link reported a negative wait")
	}
}

func TestTheSimulatedNodesWaitEndsOnExpiryCancellationOrClosing(t *testing.T) {
	t.Parallel()
	t.Run("expiry", func(t *testing.T) {
		t.Parallel()
		got, met, err := NewSimLink(10*time.Millisecond).Await(t.Context(), isKind("HSYNC"))
		if err != nil || met || len(got) != 0 {
			t.Fatalf("Await = %v, %v, %v, want an unmet wait without error", got, met, err)
		}
	})
	t.Run("cancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		if _, _, err := NewSimLink(time.Second).Await(ctx, isKind("HSYNC")); !errors.Is(err, context.Canceled) {
			t.Fatalf("Await = %v, want %v", err, context.Canceled)
		}
	})
	t.Run("closing", func(t *testing.T) {
		t.Parallel()
		link := NewSimLink(time.Second)
		link.Close()
		if _, _, err := link.Await(t.Context(), isKind("HSYNC")); !errors.Is(err, ErrSimClosed) {
			t.Fatalf("Await = %v, want %v", err, ErrSimClosed)
		}
	})
}

func TestSpawnReportsTheResultThenCallsAfter(t *testing.T) {
	t.Parallel()
	failure := errors.New("the node ended")
	called := make(chan struct{})
	done := Spawn(func() error { return failure }, func() { close(called) })
	if err := <-done; !errors.Is(err, failure) {
		t.Fatalf("Spawn result = %v, want %v", err, failure)
	}
	<-called
}
