package app

import (
	"errors"
	"io/fs"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// failureDocument is what a failed command prints on stdout (CLI_SURFACE §2).
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

// describeFailure gathers what a tool needs to act on a failure: the stable
// code, a one-line message, the path at fault, every contract issue and, for
// a failed check, the command's result document.
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

// failureSource is the path at fault: the document a contract violation
// names, or the path of a failed file or socket operation.
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

// plainMessage is the error text without the repeated "streamsim: " prefixes
// that wrapping adds. For a contract violation it is the summary's first line:
// the issues list carries the details.
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
