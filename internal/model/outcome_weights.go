package model

type OutcomeWeights struct {
	Boundaries []BoundaryWeight `json:"boundaries,omitempty"`
	Trips      []TripWeight     `json:"trips,omitempty"`
	KPIs       []KPIWeight      `json:"kpis,omitempty"`
	Calls      []CallWeight     `json:"calls,omitempty"`
	Tie        float64          `json:"tie,omitempty"`
}

type BoundaryWeight struct {
	Kind         string  `json:"kind"`
	Name         string  `json:"name"`
	PerViolation float64 `json:"per_violation,omitempty"`
	PerHour      float64 `json:"per_hour,omitempty"`
}

type TripWeight struct {
	Kind    string  `json:"kind"`
	State   string  `json:"state"`
	PerTrip float64 `json:"per_trip"`
}

type KPIWeight struct {
	Kind    string  `json:"kind"`
	State   string  `json:"state"`
	PerUnit float64 `json:"per_unit"`
}

type CallWeight struct {
	Effector string  `json:"effector"`
	PerCall  float64 `json:"per_call"`
}
