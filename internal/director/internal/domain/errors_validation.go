package domain

import (
	"fmt"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func ValidationToolError(err error, fallback Code) *ToolError {
	invalid, ok := model.AsValidationError(err)
	if !ok {
		return errTool(fallback, "%v", err)
	}
	code := fallback
	if known, found := invalid.Contract.FailureCode(); found {
		code = Code(known)
	}
	return errTool(code, "%s", withIssues(err.Error(), invalid.Issues))
}

func withIssues(message string, issues []model.Issue) string {
	var out strings.Builder
	out.WriteString(message)
	for _, issue := range issues {
		line := fmt.Sprintf("%s: %s", issue.Pointer, issue.Message)
		if issue.Located && !strings.Contains(message, line) {
			out.WriteString("\n  " + line)
		}
	}
	return out.String()
}
