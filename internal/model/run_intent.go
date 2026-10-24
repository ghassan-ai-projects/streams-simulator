package model

import "slices"

const (
	TerminationCompleted         = "completed"
	TerminationStoppedByOperator = "stopped_by_operator"
)

type RunPlan struct {
	RequestedEndNS int64   `json:"requested_end_ns"`
	Rate           float64 `json:"rate,omitempty"`
}

type Termination struct {
	Reason string `json:"reason"`
	AtNS   int64  `json:"at_ns"`
}

func TerminationReasons() []string {
	return []string{TerminationCompleted, TerminationStoppedByOperator}
}

func IsTerminationReason(reason string) bool {
	return slices.Contains(TerminationReasons(), reason)
}
