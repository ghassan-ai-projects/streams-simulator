package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
)

const (
	MaxControlLine  = 4096
	MaxControlReply = 2 << 20
)

func ControlLine(args map[string]any) ([]byte, error) {
	line, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("%w: a control request: %w", ErrBadArguments, err)
	}
	if len(line) > MaxControlLine {
		return nil, fmt.Errorf("%w: a control request holds more than %d bytes", ErrTooLarge, MaxControlLine)
	}
	return line, nil
}

func ControlAnswer(run string, reply []byte) (json.RawMessage, error) {
	one := bytes.TrimSpace(reply)
	var document map[string]json.RawMessage
	if json.Unmarshal(one, &document) != nil || len(document) == 0 {
		return nil, controlFailure("control_unreadable", run, "the node answered with something that is not a JSON document; read the state before sending the control again")
	}
	return json.RawMessage(one), nil
}

func NoLiveSession() *Failure {
	return &Failure{Code: "no_session", Message: "no connected session is running: start one before sending a control"}
}

func ControlNotOffered(run string) *Failure {
	return controlFailure("control_unavailable", run, "the session has no control socket, so the node cannot be controlled")
}

func ControlGone(run string) *Failure {
	return controlFailure("control_gone", run, "the node's control socket is not there or refused the connection; the node may have stopped or not started listening yet")
}

func ControlSilent(run string) *Failure {
	return controlFailure("control_timeout", run, "the node did not answer in time; the control may still apply afterwards, so read the state before sending it again")
}

func controlFailure(code, run, message string) *Failure {
	return &Failure{Code: code, Message: fmt.Sprintf("connected session %s: %s", run, message)}
}
