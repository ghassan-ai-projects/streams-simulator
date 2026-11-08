package model

const (
	TermBoundaryViolations = "boundary_violations"
	TermBoundaryHours      = "boundary_hours"
	TermTrips              = "trips"
	TermKPIs               = "kpis"
	TermCalls              = "calls"

	VerdictBeats = "beats"
	VerdictTies  = "ties"
	VerdictLoses = "loses"
)

type OutcomeReport struct {
	RunID  string `json:"run_id"`
	Domain string `json:"domain"`
	OutcomeScore
}

type OutcomeScore struct {
	Valid    bool             `json:"valid"`
	Tie      *float64         `json:"tie,omitempty"`
	Runs     []OutcomeRun     `json:"runs"`
	Verdicts []OutcomeVerdict `json:"verdicts,omitempty"`
}

type OutcomeRun struct {
	Mode  OutcomeMode   `json:"mode"`
	Facts *OutcomeFacts `json:"facts"`
	Cost  *OutcomeCost  `json:"cost,omitempty"`
}

type OutcomeCost struct {
	J     float64    `json:"j"`
	Terms []CostTerm `json:"terms"`
}

type CostTerm struct {
	Term     string  `json:"term"`
	Kind     string  `json:"kind,omitempty"`
	Name     string  `json:"name"`
	EntityID string  `json:"entity_id"`
	Quantity float64 `json:"quantity"`
	Cost     float64 `json:"cost"`
}

type OutcomeVerdict struct {
	Baseline OutcomeMode `json:"baseline"`
	Delta    float64     `json:"delta"`
	Verdict  string      `json:"verdict"`
}
