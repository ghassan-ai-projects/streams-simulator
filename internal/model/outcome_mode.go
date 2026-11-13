package model

import "strings"

type OutcomeMode string

const (
	OutcomeAsRun     OutcomeMode = "as_run"
	OutcomeDoNothing OutcomeMode = "do_nothing"

	outcomePolicyPrefix = "policy:"
)

func PolicyOutcome(name string) OutcomeMode { return OutcomeMode(outcomePolicyPrefix + name) }

func (m OutcomeMode) Policy() (string, bool) {
	return strings.CutPrefix(string(m), outcomePolicyPrefix)
}
