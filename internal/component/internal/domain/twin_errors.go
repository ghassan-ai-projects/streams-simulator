package domain

import (
	"errors"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var (
	ErrKindNotReproducible = errors.New("component: the kind file holds what a twin cannot reproduce")
	ErrNoKindFile          = errors.New("component: no kind file defines a kind the choices name")
	ErrNameCollision       = errors.New("component: the choices use one name twice")
	ErrUnknownChoice       = errors.New("component: the choices name what the kind files do not hold")
	ErrIncompleteChoices   = errors.New("component: the choices leave a metric or command undecided")
	ErrInstrumentMismatch  = errors.New("component: the instrument does not fit the metric")
	ErrResponseMismatch    = errors.New("component: the response does not fit the command")
	ErrValueOutOfRange     = errors.New("component: the initial value lies outside the metric's range")
)

func twinRefusal(cause error, source, at, format string, args ...any) error {
	invalid := violation(model.ContractTwinChoices, source, []model.Issue{model.LocatedIssue(at, fmt.Sprintf(format, args...))})
	invalid.Cause = cause
	return invalid
}

func (t *twinBuild) refuse(cause error, at, format string, args ...any) error {
	return twinRefusal(cause, t.source, at, format, args...)
}
