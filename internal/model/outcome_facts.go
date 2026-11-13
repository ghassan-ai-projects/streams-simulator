package model

type OutcomeFacts struct {
	HorizonNS int64           `json:"horizon_ns"`
	Entities  []EntityOutcome `json:"entities"`
	Calls     []OutcomeCall   `json:"calls"`
}

type EntityOutcome struct {
	EntityID   string            `json:"entity_id"`
	Kind       string            `json:"kind"`
	Trips      []TripOutcome     `json:"trips,omitempty"`
	Boundaries []BoundaryOutcome `json:"boundaries,omitempty"`
	KPIs       []KPIOutcome      `json:"kpis,omitempty"`
}

type TripOutcome struct {
	State   string `json:"state"`
	FirstNS int64  `json:"first_ns"`
	Count   int    `json:"count"`
}

type BoundaryOutcome struct {
	Name       string  `json:"name"`
	Severity   string  `json:"severity"`
	Violations int     `json:"violations"`
	FirstNS    int64   `json:"first_ns"`
	DurationS  float64 `json:"duration_s"`
}

type KPIOutcome struct {
	State string  `json:"state"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

type OutcomeCall struct {
	Effector string `json:"effector"`
	EntityID string `json:"entity_id"`
	AtNS     int64  `json:"at_ns"`
	Accepted bool   `json:"accepted"`
}
