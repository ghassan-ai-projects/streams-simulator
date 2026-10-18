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
)

const (
	recordBirth    = "NBIRTH"
	recordWire     = "wire"
	streamDirector = "director"
	nodeDirection  = "node"
)

var birthMarker = []byte(recordBirth)

type envelope struct {
	Type      string          `json:"type"`
	Stream    string          `json:"stream"`
	Direction string          `json:"direction"`
	Record    json.RawMessage `json:"record"`
}

func FindBirth(raw []byte) ([]byte, error) {
	birth, found, err := birthOf(raw)
	if found || errors.Is(err, ErrSealedStream) {
		return birth, err
	}
	return birthInLines(raw)
}

func birthInLines(raw []byte) ([]byte, error) {
	var skipped error
	for rest := raw; len(rest) > 0; {
		var line []byte
		line, rest, _ = bytes.Cut(rest, []byte{'\n'})
		birth, found, err := birthOf(line)
		if found || errors.Is(err, ErrSealedStream) {
			return birth, err
		}
		skipped = firstError(skipped, err)
	}
	return nil, noBirth(skipped)
}

func firstError(first, next error) error {
	if first != nil {
		return first
	}
	return next
}

func birthOf(document []byte) ([]byte, bool, error) {
	if !bytes.Contains(document, birthMarker) {
		return nil, false, nil
	}
	var wrapper envelope
	if err := json.Unmarshal(document, &wrapper); err != nil {
		return nil, false, fmt.Errorf("plantwire: not a JSON record: %w", err)
	}
	return wrapper.birth(document)
}

func (e envelope) birth(document []byte) ([]byte, bool, error) {
	switch {
	case e.Type == recordBirth:
		return document, true, nil
	case e.Type != recordWire || e.Direction != nodeDirection || typeOf(e.Record) != recordBirth:
		return nil, false, nil
	case e.Stream == streamDirector:
		return nil, false, birthViolation("the NBIRTH is a record of the sealed director stream, which is never taken", "sealed_stream", ErrSealedStream)
	}
	return e.Record, true, nil
}

func typeOf(record json.RawMessage) string {
	var head struct {
		Type string `json:"type"`
	}
	_ = json.Unmarshal(record, &head)
	return head.Type
}

func noBirth(skipped error) error {
	message := "no NBIRTH record found"
	if skipped != nil {
		message += ": the lines that name one are not JSON records (" + skipped.Error() + ")"
	}
	return birthViolation(message, "no_birth", ErrNoBirth)
}

func birthViolation(message, keyword string, cause error) *model.ValidationError {
	return &model.ValidationError{Contract: model.ContractPlantWireBirth, Summary: "plantwire: " + message,
		Issues: []model.Issue{model.UnlocatedRuleIssue(keyword, message)}, Cause: cause}
}
