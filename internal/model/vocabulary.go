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
	FormSequencer   = "sequencer"
	FallbackSafeOff = "safe_off"
	FallbackHold    = "hold_last"
	FallbackRevert  = "revert"
	EdgeFall        = "fall"
	EdgeRise        = "rise"
)

type Condition struct {
	State          string  `json:"state"`
	Operator       string  `json:"operator"`
	Threshold      float64 `json:"threshold"`
	ThresholdState string  `json:"threshold_state,omitempty"`
}

type Boundary struct {
	Name     string `json:"name"`
	State    string `json:"state"`
	Severity string `json:"severity"`
}

type Responder struct {
	MatchArg   string       `json:"match_arg"`
	DelayS     float64      `json:"delay_s"`
	OnMatch    []StateDelta `json:"on_match,omitempty"`
	OnMiss     []StateDelta `json:"on_miss,omitempty"`
	ClearFault bool         `json:"clear_fault,omitempty"`
}

type Lease struct {
	Arg        string       `json:"arg"`
	Fallback   string       `json:"fallback"`
	SafeDeltas []StateDelta `json:"safe_deltas,omitempty"`
}
