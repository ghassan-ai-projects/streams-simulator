package model

const (
	FormRegime        = "regime"
	FormPeak          = "peak"
	InterpolateStep   = "step"
	InterpolateLinear = "linear"
)

type Schedule struct {
	PeriodS       float64         `json:"period_s,omitempty"`
	Interpolation string          `json:"interpolation"`
	Points        []SchedulePoint `json:"points"`
}

type SchedulePoint struct {
	AtS   float64 `json:"at_s"`
	Value float64 `json:"value"`
}

type Regime struct {
	Gain          float64 `json:"gain"`
	TimeConstantS float64 `json:"time_constant_s"`
}

const (
	FormTrip        = "trip"
	FormAntiRecycle = "anti_recycle"
	EdgeFall        = "fall"
	EdgeRise        = "rise"
)

type Condition struct {
	State     string  `json:"state"`
	Operator  string  `json:"operator"`
	Threshold float64 `json:"threshold"`
}

type Boundary struct {
	Name     string `json:"name"`
	State    string `json:"state"`
	Severity string `json:"severity"`
}
