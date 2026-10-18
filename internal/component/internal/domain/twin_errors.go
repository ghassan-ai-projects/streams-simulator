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

type twinRule struct{ keyword, member string }

var twinRules = []struct {
	cause error
	rule  twinRule
}{
	{ErrNoKindFile, twinRule{"twin_kind_missing", "kind"}},
	{ErrNameCollision, twinRule{"twin_name_collision", ""}},
	{ErrUnknownChoice, twinRule{"twin_unknown_choice", ""}},
	{ErrIncompleteChoices, twinRule{"twin_choice_missing", ""}},
	{ErrInstrumentMismatch, twinRule{"twin_instrument_mismatch", "instrument"}},
	{ErrResponseMismatch, twinRule{"twin_response_mismatch", ""}},
	{ErrValueOutOfRange, twinRule{"twin_initial_range", "initial"}},
}

func ruleOf(cause error) twinRule {
	for _, known := range twinRules {
		if errors.Is(cause, known.cause) {
			return known.rule
		}
	}
	return twinRule{}
}

func twinRefusal(cause error, rule twinRule, source, at, format string, args ...any) error {
	issue := model.RuleIssue(at, rule.keyword, rule.member, fmt.Sprintf(format, args...))
	invalid := violation(model.ContractTwinChoices, source, []model.Issue{issue})
	invalid.Cause = cause
	return invalid
}

func (t *twinBuild) refuse(cause error, at, format string, args ...any) error {
	return twinRefusal(cause, ruleOf(cause), t.source, at, format, args...)
}

func (t *twinBuild) refuseRule(cause error, rule twinRule, at, format string, args ...any) error {
	return twinRefusal(cause, rule, t.source, at, format, args...)
}

func (t *twinBuild) refuseKindFile(kind, at string, found []model.Issue) error {
	issues := make([]model.Issue, len(found))
	for i, issue := range found {
		issues[i] = model.RuleIssue(at, issue.Keyword, issue.Member, fmt.Sprintf("kind file %q at %s: %s", kind, issue.Pointer, issue.Message))
	}
	invalid := violation(model.ContractTwinChoices, t.source, issues)
	invalid.Cause = fmt.Errorf("%w: %s", ErrKindNotReproducible, found[0].Message)
	return invalid
}
