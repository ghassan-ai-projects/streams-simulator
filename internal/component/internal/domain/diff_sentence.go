package domain

import (
	"encoding/json"
	"fmt"
	"strings"
)

const shownRunes = 80

func newChange(code, subject, section, field string, before, after any) Change {
	change := Change{Code: code, Subject: subject, Section: section, Field: field, Before: before, After: after}
	change.Sentence = sentenceOf(change)
	return change
}

func sentenceOf(change Change) string {
	switch {
	case strings.HasSuffix(change.Code, "_added"):
		return fmt.Sprintf("%s: %s added.", change.Subject, change.Field)
	case strings.HasSuffix(change.Code, "_removed"):
		return fmt.Sprintf("%s: %s removed.", change.Subject, change.Field)
	}
	return fmt.Sprintf("%s: %s changed from %s to %s.", change.Subject, change.Field, shown(change.Before), shown(change.After))
}

func shown(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return "?"
	}
	text := []rune(string(raw))
	if len(text) > shownRunes {
		return string(text[:shownRunes]) + "..."
	}
	return string(text)
}
