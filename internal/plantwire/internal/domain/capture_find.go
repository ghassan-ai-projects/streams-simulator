package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var (
	ErrNoBirth      = errors.New("plantwire: no NBIRTH record found")
	ErrSealedStream = errors.New("plantwire: the NBIRTH is a record of the sealed director stream")
	ErrNotObserver  = errors.New("plantwire: the NBIRTH is not a record of the observer stream")
)

const (
	recordBirth    = "NBIRTH"
	recordWire     = "wire"
	streamDirector = "director"
	streamObserver = "observer"
	nodeDirection  = "node"
)

var birthMarker = []byte(recordBirth)

func classify(line []byte, tooLong bool) ([]byte, bool, error) {
	if tooLong {
		return nil, false, fmt.Errorf("%w: a line of more than %d bytes", ErrTooLarge, MaxLineBytes)
	}
	if !bytes.Contains(line, birthMarker) {
		return nil, false, nil
	}
	fields, err := strictFields(line)
	if err != nil {
		return nil, false, err
	}
	return birthIn(fields, line)
}

func strictFields(document []byte) (map[string]json.RawMessage, error) {
	if err := refuseDuplicateMembers(document); err != nil {
		return nil, fmt.Errorf("plantwire: not a distinct-member JSON record: %w", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(document, &fields); err != nil {
		return nil, fmt.Errorf("plantwire: not a JSON record: %w", err)
	}
	return fields, nil
}

func birthIn(fields map[string]json.RawMessage, line []byte) ([]byte, bool, error) {
	switch exactString(fields["type"]) {
	case recordBirth:
		return acceptedBirth(fields, line, false)
	case recordWire:
		if exactString(fields["direction"]) == nodeDirection && recordType(fields["record"]) == recordBirth {
			return acceptedBirth(fields, fields["record"], true)
		}
	}
	return nil, false, nil
}

func acceptedBirth(fields map[string]json.RawMessage, birth []byte, wrapped bool) ([]byte, bool, error) {
	if err := requireObserver(fields, wrapped); err != nil {
		return nil, false, err
	}
	return birth, true, nil
}

func requireObserver(fields map[string]json.RawMessage, wrapped bool) error {
	raw, present := fields["stream"]
	stream := exactString(raw)
	switch {
	case !present && !wrapped, stream == streamObserver:
		return nil
	case stream == streamDirector:
		return birthViolation("the NBIRTH is a record of the sealed director stream, which is never taken", "sealed_stream", ErrSealedStream)
	}
	return birthViolation(fmt.Sprintf("the NBIRTH is of stream %q: only the observer stream is taken", clip(stream)), "stream_not_observer", ErrNotObserver)
}

func isStreamRefusal(err error) bool {
	return errors.Is(err, ErrSealedStream) || errors.Is(err, ErrNotObserver)
}

func exactString(raw json.RawMessage) string {
	var text string
	_ = json.Unmarshal(raw, &text)
	return text
}

func recordType(raw json.RawMessage) string {
	var fields map[string]json.RawMessage
	_ = json.Unmarshal(raw, &fields)
	return exactString(fields["type"])
}

func noBirth(scan Scan) error {
	message := "no NBIRTH record found"
	if scan.Truncated {
		message += fmt.Sprintf(" in the first %d bytes", MaxScanBytes)
	}
	cause := ErrNoBirth
	if len(scan.Unreadable) > 0 {
		first := scan.Unreadable[0]
		message += fmt.Sprintf(": line %d names one but cannot be read (%v)", first.Line, first.Cause)
		cause = errors.Join(ErrNoBirth, first.Cause)
	}
	return birthViolation(message, "no_birth", cause)
}

func birthViolation(message, keyword string, cause error) *model.ValidationError {
	return &model.ValidationError{Contract: model.ContractPlantWireBirth, Summary: "plantwire: " + message,
		Issues: []model.Issue{model.UnlocatedRuleIssue(keyword, message)}, Cause: cause}
}
