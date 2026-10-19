package domain

import (
	"encoding/json"
	"fmt"
)

type Lifecycle string

const (
	LastingStart   Lifecycle = "start"
	LastingStatus  Lifecycle = "status"
	LastingStop    Lifecycle = "stop"
	LastingControl Lifecycle = "control"
)

type SessionState string

const (
	SessionNone     SessionState = "none"
	SessionStarting SessionState = "starting"
	SessionRunning  SessionState = "running"
	SessionStopping SessionState = "stopping"
	SessionEnded    SessionState = "ended"
	SessionFailed   SessionState = "failed"
)

const (
	SessionBarrierCapMS = "10000"
	MaxSessionResult    = 64 << 10
	RunIDAttempts       = 5
)

var (
	cliSessionCheck = Action{Name: "session.check", Kind: KindCLI, Command: []string{"plantwire", "connect", "--check", "--stdin"},
		Catalog: true, Kinds: true, Input: inputKey}
	cliSessionStart = Action{Name: "session.start", Kind: KindCLI,
		Command: []string{"plantwire", "connect", "--stdin", "--barrier-cap-ms", SessionBarrierCapMS},
		Catalog: true, Kinds: true, Input: inputKey, Mutating: true, Lasting: LastingStart}
	cliSessionStatus  = Action{Name: "session.status", Kind: KindCLI, Command: []string{"plantwire", "connect"}, Lasting: LastingStatus}
	cliSessionStop    = Action{Name: "session.stop", Kind: KindCLI, Command: []string{"plantwire", "connect"}, Mutating: true, Lasting: LastingStop}
	cliSessionControl = Action{Name: "session.control", Kind: KindCLI, Command: []string{"plantwire", "connect"}, Mutating: true, Lasting: LastingControl}
)

func SessionPrecheck() Action {
	return cliSessionCheck
}

type SessionStatus struct {
	Run      string          `json:"run,omitempty"`
	State    SessionState    `json:"state"`
	Domain   string          `json:"domain,omitempty"`
	Plan     json.RawMessage `json:"plan,omitempty"`
	Result   json.RawMessage `json:"result,omitempty"`
	Failure  *Failure        `json:"failure,omitempty"`
	Controls bool            `json:"controls,omitempty"`
}

type SessionView struct {
	Run    string
	Domain string
	Plan   json.RawMessage
}

func Controlling(state SessionState) bool {
	return state == SessionStarting || state == SessionRunning || state == SessionStopping
}

func PlanDomain(check json.RawMessage) string {
	var head struct {
		Domain string `json:"domain"`
	}
	_ = json.Unmarshal(check, &head)
	return head.Domain
}

func (v SessionView) Status(state SessionState) SessionStatus {
	return SessionStatus{Run: v.Run, State: state, Domain: v.Domain, Plan: v.Plan}
}

func (v SessionView) Outcome(stdout []byte, exitCode int, stderr string) SessionStatus {
	if exitCode == 0 && json.Valid(stdout) {
		status := v.Status(SessionEnded)
		if len(stdout) <= MaxSessionResult {
			status.Result = json.RawMessage(stdout)
		}
		return status
	}
	status := v.Status(SessionFailed)
	status.Failure = CommandFailure(stdout, exitCode, stderr)
	return status
}

func (v SessionView) Unlaunched(err error) SessionStatus {
	status := v.Status(SessionFailed)
	status.Failure = &Failure{Code: "session_failed", Message: err.Error()}
	return status
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
