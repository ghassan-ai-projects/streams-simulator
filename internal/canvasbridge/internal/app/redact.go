package app

import (
	"errors"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

type redacted struct {
	err  error
	text string
}

func (r *redacted) Error() string { return r.text }

func (r *redacted) Unwrap() error { return r.err }

func redact(err error, root string) error {
	var failure *domain.Failure
	if errors.As(err, &failure) {
		return &domain.Failure{Code: failure.Code, Message: domain.Scrub(failure.Message, root),
			Issues: []byte(domain.Scrub(string(failure.Issues), root)), Result: []byte(domain.Scrub(string(failure.Result), root))}
	}
	if text := domain.Scrub(err.Error(), root); text != err.Error() {
		return &redacted{err: err, text: text}
	}
	return err
}
