package app

import (
	"errors"
	"io/fs"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type failureDocument struct {
	Error failureBody `json:"error"`
}

type failureBody struct {
	Code    failureCode    `json:"code"`
	Message string         `json:"message"`
	Source  string         `json:"source,omitempty"`
	Issues  []model.Issue  `json:"issues,omitempty"`
	Result  map[string]any `json:"result,omitempty"`
}

func describeFailure(command string, err error) failureDocument {
	body := failureBody{Code: classifyFailure(command, err), Message: plainMessage(err), Source: failureSource(err)}
	if invalid, ok := asValidationError(err); ok {
		body.Issues = invalid.Issues
	}
	if outcome, ok := asOutcomeError(err); ok {
		body.Result = outcome.result
	}
	return failureDocument{Error: body}
}

func failureSource(err error) string {
	if invalid, ok := asValidationError(err); ok {
		return invalid.Source
	}
	var pathFailure *fs.PathError
	if errors.As(err, &pathFailure) {
		return pathFailure.Path
	}
	return ""
}

func plainMessage(err error) string {
	message := err.Error()
	for strings.HasPrefix(message, "streamsim: ") {
		message = strings.TrimPrefix(message, "streamsim: ")
	}
	if _, ok := asValidationError(err); ok {
		message, _, _ = strings.Cut(message, "\n")
		message = strings.TrimSuffix(message, ":")
	}
	return message
}

func asValidationError(err error) (*model.ValidationError, bool) {
	var invalid *model.ValidationError
	return invalid, errors.As(err, &invalid)
}
