package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestAControlRequestIsForwardedAsOneLineOfItsOwnMembers(t *testing.T) {
	t.Parallel()
	line, err := ControlLine(map[string]any{"op": "hold", "scope": map[string]any{"kind": "board", "id": "mega-01"}})
	if err != nil || string(line) != `{"op":"hold","scope":{"id":"mega-01","kind":"board"}}` {
		t.Fatalf("ControlLine = %s, %v, want the members as given on one line", line, err)
	}
	if odd, err := ControlLine(map[string]any{"anything": "at all"}); err != nil || string(odd) != `{"anything":"at all"}` {
		t.Fatalf("ControlLine of an unknown member = %s, %v, want it forwarded: the binary decides what is a control", odd, err)
	}
}

func TestAControlRequestOverTheLineLimitOrThatCannotBeEncodedIsRefused(t *testing.T) {
	t.Parallel()
	if _, err := ControlLine(map[string]any{"op": strings.Repeat("a", MaxControlLine)}); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("a request over the limit = %v, want ErrTooLarge", err)
	}
	if _, err := ControlLine(map[string]any{"op": make(chan int)}); !errors.Is(err, ErrBadArguments) {
		t.Fatalf("a request that is not JSON = %v, want ErrBadArguments", err)
	}
}

func TestTheNodesAnswerIsReturnedUnchangedAndAnythingElseIsUnreadable(t *testing.T) {
	t.Parallel()
	reply := `{"result":"refused","reason":"node_busy","state":{"holds":[]}}`
	if got, err := ControlAnswer("c-1", []byte(reply+"\n")); err != nil || string(got) != reply {
		t.Fatalf("ControlAnswer = %s, %v, want the document unchanged", got, err)
	}
	for _, bad := range []string{"", "not json", "[]", "{}", "42"} {
		var failure *Failure
		if _, err := ControlAnswer("c-1", []byte(bad)); !errors.As(err, &failure) || failure.Code != "control_unreadable" {
			t.Errorf("ControlAnswer(%q) = %v, want control_unreadable", bad, err)
		}
	}
}

func TestEachWayAControlCanFailHasItsOwnCodeAndNamesTheRun(t *testing.T) {
	t.Parallel()
	cases := map[string]*Failure{
		"no_session":          NoLiveSession(),
		"control_unavailable": ControlNotOffered("c-1"),
		"control_gone":        ControlGone("c-1"),
		"control_timeout":     ControlSilent("c-1"),
	}
	for code, failure := range cases {
		if failure.Code != code || failure.Message == "" {
			t.Errorf("%s: got %+v", code, failure)
		}
	}
	if !strings.Contains(ControlSilent("c-1").Message, "c-1") || !strings.Contains(ControlSilent("c-1").Message, "read the state") {
		t.Fatalf("the timeout words = %q, want the run and the advice to read the state", ControlSilent("c-1").Message)
	}
}

func TestOnlyALivingSessionIsControlled(t *testing.T) {
	t.Parallel()
	for state, want := range map[SessionState]bool{SessionNone: false, SessionStarting: true, SessionRunning: true, SessionStopping: true,
		SessionEnded: false, SessionFailed: false} {
		if Controlling(state) != want {
			t.Errorf("Controlling(%s) = %v, want %v", state, !want, want)
		}
	}
	control, err := Lookup("session.control")
	if err != nil || control.Lasting != LastingControl || !control.Mutating || control.Input != "" {
		t.Fatalf("session.control = %+v, %v, want a lasting, mutating action that reads no standard input", control, err)
	}
}
