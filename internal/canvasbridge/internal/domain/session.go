package domain

import (
	"encoding/json"
	"fmt"
)

type Lifecycle string

const (
	LastingStart  Lifecycle = "start"
	LastingStatus Lifecycle = "status"
	LastingStop   Lifecycle = "stop"
)

type SessionState string

const (
	SessionNone    SessionState = "none"
	SessionRunning SessionState = "running"
	SessionEnded   SessionState = "ended"
	SessionFailed  SessionState = "failed"
)

var (
	cliSessionCheck = Action{Name: "session.check", Kind: KindCLI, Command: []string{"plantwire", "connect", "--check", "--stdin"},
		Catalog: true, Kinds: true, Input: inputKey}
	cliSessionStart = Action{Name: "session.start", Kind: KindCLI, Command: []string{"plantwire", "connect", "--stdin"},
		Catalog: true, Kinds: true, Input: inputKey, Mutating: true, Lasting: LastingStart}
	cliSessionStatus = Action{Name: "session.status", Kind: KindCLI, Command: []string{"plantwire", "connect"}, Lasting: LastingStatus}
	cliSessionStop   = Action{Name: "session.stop", Kind: KindCLI, Command: []string{"plantwire", "connect"}, Mutating: true, Lasting: LastingStop}
)

type SessionStatus struct {
	Run     string          `json:"run,omitempty"`
	State   SessionState    `json:"state"`
	Result  json.RawMessage `json:"result,omitempty"`
	Failure *Failure        `json:"failure,omitempty"`
}

func SessionOutcome(run string, stdout []byte, exitCode int, stderr string) SessionStatus {
	if exitCode == 0 && json.Valid(stdout) {
		return SessionStatus{Run: run, State: SessionEnded, Result: json.RawMessage(stdout)}
	}
	return SessionStatus{Run: run, State: SessionFailed, Failure: CommandFailure(stdout, exitCode, stderr)}
}

func SessionStillRunning(run string) *Failure {
	return &Failure{Code: "session_running", Message: fmt.Sprintf("the connected session %s is still running: stop it before starting another", run)}
}

func SessionFolder(run string) (ArgvPart, error) {
	if badSegment(run) {
		return ArgvPart{}, fmt.Errorf("%w: run id %q", ErrBadPath, run)
	}
	return ArgvPart{Path: "runs/" + run}, nil
}
