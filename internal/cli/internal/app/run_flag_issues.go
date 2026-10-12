package app

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// flagScenarioFailure restates the violations of a scenario the run flags
// describe in terms of the flags: the user wrote flags, not the in-memory
// document the issue pointers locate. Each issue names the flag occurrence
// behind it.
func flagScenarioFailure(err error, origins []string) error {
	var invalid *model.ValidationError
	if !errors.As(err, &invalid) {
		return fmt.Errorf("streamsim: %w", err)
	}
	issues := make([]model.Issue, len(invalid.Issues))
	for i, issue := range invalid.Issues {
		issues[i] = model.UnlocatedIssue(flagOrigin(issue.Pointer, origins) + ": " + issue.Message)
	}
	cause := &model.ValidationError{Contract: invalid.Contract, Summary: "run: " + issues[0].Message, Issues: issues}
	return &usageError{message: cause.Summary, cause: cause}
}

// flagOrigin names the flag behind the scenario field a pointer locates.
func flagOrigin(pointer string, origins []string) string {
	if index, ok := commandIndex(pointer); ok && index < len(origins) {
		return origins[index]
	}
	if pointer == "/world/profile" {
		return "--profile"
	}
	return "the run flags"
}

// commandIndex is the command a pointer locates: /commands/<i> or anything
// below it.
func commandIndex(pointer string) (int, bool) {
	rest, found := strings.CutPrefix(pointer, "/commands/")
	if !found {
		return 0, false
	}
	digits, _, _ := strings.Cut(rest, "/")
	index, err := strconv.Atoi(digits)
	return index, err == nil && index >= 0
}
