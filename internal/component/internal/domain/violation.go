package domain

import (
	"fmt"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type issueList struct{ items []model.Issue }

func (l *issueList) addf(pointer, format string, args ...any) {
	l.items = append(l.items, model.LocatedIssue(pointer, fmt.Sprintf(format, args...)))
}

func (l *issueList) empty() bool { return len(l.items) == 0 }

func (l *issueList) violation(contract model.Contract, source string) error {
	if l.empty() {
		return nil
	}
	return violation(contract, source, l.items)
}

func violation(contract model.Contract, source string, issues []model.Issue) *model.ValidationError {
	lines := make([]string, len(issues))
	for i, issue := range issues {
		lines[i] = issue.Pointer + ": " + issue.Message
	}
	summary := fmt.Sprintf("component: %s fails %s validation:\n  %s", source, contract, strings.Join(lines, "\n  "))
	return &model.ValidationError{Contract: contract, Source: source, Summary: summary, Issues: issues}
}

func refuse(contract model.Contract, source, pointer, format string, args ...any) error {
	issue := model.LocatedIssue(pointer, fmt.Sprintf(format, args...))
	return violation(contract, source, []model.Issue{issue})
}

func pointer(parts ...string) string {
	var out strings.Builder
	for _, part := range parts {
		part = strings.ReplaceAll(strings.ReplaceAll(part, "~", "~0"), "/", "~1")
		out.WriteString("/" + part)
	}
	return out.String()
}

func relative(parts ...string) string { return pointer(parts...)[1:] }

func indexPart(i int) string { return fmt.Sprint(i) }
