package model

type DomainSpec struct {
	ID          string        `json:"id"`
	Version     string        `json:"version"`
	Title       string        `json:"title"`
	Description string        `json:"description,omitempty"`
	Axes        Axes          `json:"axes"`
	Stresses    string        `json:"stresses"`
	Kinds       []Kind        `json:"kinds"`
	Instances   []InstanceSet `json:"instances"`
	LinkTypes   []LinkType    `json:"link_types,omitempty"`
	Links       []Link        `json:"links,omitempty"`
	Boards      []Board       `json:"boards,omitempty"`
	Profiles    []Profile     `json:"profiles"`
	GroundTruth GroundTruth   `json:"ground_truth"`
}

type Kind struct {
	Name       string     `json:"name"`
	System     bool       `json:"system,omitempty"`
	State      []State    `json:"state,omitempty"`
	Dynamics   []Dynamics `json:"dynamics,omitempty"`
	Flows      []Flow     `json:"flows,omitempty"`
	Channels   []Channel  `json:"channels"`
	Faults     []Fault    `json:"faults"`
	Effectors  []Effector `json:"effectors,omitempty"`
	Boundaries []Boundary `json:"boundaries,omitempty"`
}

type InstanceSet struct {
	Kind       string      `json:"kind"`
	Names      []string    `json:"names,omitempty"`
	IDTemplate string      `json:"id_template,omitempty"`
	Count      EntityCount `json:"count"`
	Churn      *Churn      `json:"churn,omitempty"`
	Override   *Override   `json:"override,omitempty"`
}

type Axes struct {
	Rate        string   `json:"rate"`
	Cardinality string   `json:"cardinality"`
	ValueShape  []string `json:"value_shape"`
	Cadence     []string `json:"cadence"`
	Lateness    string   `json:"lateness"`
	Absence     string   `json:"absence"`
	TimeRef     string   `json:"time_reference"`
	Correlation []string `json:"correlation"`
	Seasonality []string `json:"seasonality"`
	Actuation   string   `json:"actuation"`
	Consequence []string `json:"consequence"`
	Fidelity    []string `json:"fidelity"`
}

type EntityCount struct {
	Default int `json:"default"`
	Min     int `json:"min,omitempty"`
	Max     int `json:"max,omitempty"`
}

type Churn struct {
	BirthsPerHour float64 `json:"births_per_hour,omitempty"`
	MeanLifetimeS float64 `json:"mean_lifetime_s,omitempty"`
}

type State struct {
	Name        string  `json:"name"`
	Unit        string  `json:"unit,omitempty"`
	Initial     float64 `json:"initial"`
	Min         float64 `json:"min,omitempty"`
	Max         float64 `json:"max,omitempty"`
	Description string  `json:"description,omitempty"`
}

type Dynamics struct {
	Target string `json:"target"`
	Tier   string `json:"tier"`
	DTMs   int64  `json:"dt_ms,omitempty"`
	F0     *F0Dyn `json:"f0,omitempty"`
	F1     *F1Dyn `json:"f1,omitempty"`
	F2     *F2Dyn `json:"f2,omitempty"`
	F3     *F3Dyn `json:"f3,omitempty"`
}

type F0Dyn struct {
	Baseline     float64       `json:"baseline,omitempty"`
	TrendPerHour float64       `json:"trend_per_hour,omitempty"`
	Seasonality  []Seasonality `json:"seasonality,omitempty"`
	DriftPerHour float64       `json:"drift_per_hour,omitempty"`
	Schedule     *Schedule     `json:"schedule,omitempty"`
}

type Seasonality struct {
	PeriodS   float64 `json:"period_s"`
	Amplitude float64 `json:"amplitude"`
	PhaseRad  float64 `json:"phase_rad,omitempty"`
}

type F1Dyn struct {
	Form           string     `json:"form"`
	TimeConstantS  float64    `json:"time_constant_s,omitempty"`
	TimeConstant2S float64    `json:"time_constant_2_s,omitempty"`
	DeadTimeS      float64    `json:"dead_time_s,omitempty"`
	Gain           float64    `json:"gain,omitempty"`
	Inputs         []F1Input  `json:"inputs,omitempty"`
	Threshold      float64    `json:"threshold,omitempty"`
	Direction      string     `json:"direction,omitempty"`
	ThresholdLow   float64    `json:"threshold_low,omitempty"`
	ThresholdHigh  float64    `json:"threshold_high,omitempty"`
	Capacity       float64    `json:"capacity,omitempty"`
	Selector       string     `json:"selector,omitempty"`
	Below          *Regime    `json:"below,omitempty"`
	Above          *Regime    `json:"above,omitempty"`
	BlockS         float64    `json:"block_s,omitempty"`
	PeriodS        float64    `json:"period_s,omitempty"`
	Ratchet        float64    `json:"ratchet,omitempty"`
	PreviousPeak   float64    `json:"previous_peak,omitempty"`
	When           *Condition `json:"when,omitempty"`
	ForS           float64    `json:"for_s,omitempty"`
	Latch          bool       `json:"latch,omitempty"`
	ResetOn        string     `json:"reset_on,omitempty"`
	TrippedValue   float64    `json:"tripped_value,omitempty"`
	Watch          string     `json:"watch,omitempty"`
	Edge           string     `json:"edge,omitempty"`
	HoldS          float64    `json:"hold_s,omitempty"`
	StageUp        *Condition `json:"stage_up,omitempty"`
	StageDown      *Condition `json:"stage_down,omitempty"`
	Stages         int        `json:"stages,omitempty"`
	Clamp          *Clamp     `json:"clamp,omitempty"`
}

type F1Input struct {
	State   string  `json:"state"`
	Coef    float64 `json:"coef,omitempty"`
	CoefSet bool    `json:"-"`
}

type Clamp struct {
	Min *float64 `json:"min,omitempty"`
	Max *float64 `json:"max,omitempty"`
}

type F2Dyn struct {
	Model  string            `json:"model"`
	Params map[string]any    `json:"params,omitempty"`
	Inputs map[string]string `json:"inputs,omitempty"`
}

type F3Dyn struct {
	FMU      string            `json:"fmu"`
	Instance string            `json:"instance,omitempty"`
	Inputs   map[string]string `json:"inputs,omitempty"`
	Outputs  map[string]string `json:"outputs,omitempty"`
}

type Channel struct {
	Name               string        `json:"name"`
	Description        string        `json:"description,omitempty"`
	ValueType          string        `json:"value_type"`
	Unit               string        `json:"unit,omitempty"`
	EnumValues         []string      `json:"enum_values,omitempty"`
	Resolution         float64       `json:"resolution"`
	Range              *Range        `json:"range,omitempty"`
	Observes           string        `json:"observes,omitempty"`
	ObservationGain    float64       `json:"observation_gain,omitempty"`
	ObservationOffset  float64       `json:"observation_offset,omitempty"`
	ObservationBias    *Bias         `json:"observation_bias,omitempty"`
	Fidelity           string        `json:"fidelity"`
	ISO13374Role       string        `json:"iso13374_role,omitempty"`
	Cadence            Cadence       `json:"cadence"`
	Noise              Noise         `json:"noise"`
	Drift              *Drift        `json:"drift,omitempty"`
	Absence            string        `json:"absence"`
	Availability       *Availability `json:"availability,omitempty"`
	LinkDelay          *LinkDelay    `json:"link_delay,omitempty"`
	AttackerControlled bool          `json:"attacker_controlled,omitempty"`
}

type Range struct {
	Min float64 `json:"min,omitempty"`
	Max float64 `json:"max,omitempty"`
}

type Bias struct {
	State            string  `json:"state"`
	Coef             float64 `json:"coef"`
	NoiseScaleAtFull float64 `json:"noise_scale_at_full,omitempty"`
}

type Cadence struct {
	Mode         string  `json:"mode"`
	PeriodS      float64 `json:"period_s,omitempty"`
	JitterS      float64 `json:"jitter_s,omitempty"`
	Deadband     float64 `json:"deadband,omitempty"`
	BatchSize    int     `json:"batch_size,omitempty"`
	TriggerState string  `json:"trigger_state,omitempty"`
}

type Noise struct {
	Model string  `json:"model"`
	Sigma float64 `json:"sigma"`
}

type Drift struct {
	RatePerHour float64 `json:"rate_per_hour,omitempty"`
	Model       string  `json:"model,omitempty"`
}

type Availability struct {
	Uptime float64 `json:"uptime,omitempty"`
	MTTRS  float64 `json:"mttr_s,omitempty"`
}

type LinkDelay struct {
	Model  string  `json:"model"`
	MeanS  float64 `json:"mean_s,omitempty"`
	SigmaS float64 `json:"sigma_s,omitempty"`
	MaxS   float64 `json:"max_s,omitempty"`
}

type Fault struct {
	ID               string          `json:"id"`
	Description      string          `json:"description,omitempty"`
	IsNegativeClass  bool            `json:"is_negative_class,omitempty"`
	Onset            FaultOnset      `json:"onset"`
	Affects          []FaultEffect   `json:"affects"`
	Observability    Observability   `json:"observability"`
	ExpectedEffector string          `json:"expected_effector,omitempty"`
	DeadlineS        float64         `json:"deadline_s,omitempty"`
	Counterfactual   *Counterfactual `json:"counterfactual,omitempty"`
}

type FaultOnset struct {
	Shape       string  `json:"shape"`
	RatePerHour float64 `json:"rate_per_hour,omitempty"`
	Magnitude   float64 `json:"magnitude,omitempty"`
	DutyCycle   float64 `json:"duty_cycle,omitempty"`
	DurationS   float64 `json:"duration_s,omitempty"`
}

type FaultEffect struct {
	State          string  `json:"state"`
	Delta          float64 `json:"delta"`
	Multiplicative bool    `json:"multiplicative,omitempty"`
}

type Observability struct {
	Detector           Detector `json:"detector"`
	FirstObservableSNR float64  `json:"first_observable_snr,omitempty"`
	UnavoidableSNR     float64  `json:"unavoidable_snr,omitempty"`
}

type Detector struct {
	Form     string   `json:"form"`
	Channel  string   `json:"channel,omitempty"`
	ChannelA string   `json:"channel_a,omitempty"`
	ChannelB string   `json:"channel_b,omitempty"`
	Inputs   []string `json:"inputs,omitempty"`
	Outputs  []string `json:"outputs,omitempty"`
}

type Counterfactual struct {
	IfNoAction         string `json:"if_no_action,omitempty"`
	IfActionByDeadline string `json:"if_action_by_deadline,omitempty"`
}
