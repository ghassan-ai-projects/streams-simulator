package model

import (
	"encoding/json"
	"fmt"
)

// DomainSpec is the typed form of docs/contracts/domain-spec-v0.1.schema.json.
// A simulated world, as data: the simulator binary contains no domain
// behavior.
type DomainSpec struct {
	ID          string      `json:"id"`
	Version     string      `json:"version"`
	Title       string      `json:"title"`
	Description string      `json:"description,omitempty"`
	Axes        Axes        `json:"axes"`
	Stresses    string      `json:"stresses"`
	Entities    Entities    `json:"entities"`
	State       []State     `json:"state,omitempty"`
	Dynamics    []Dynamics  `json:"dynamics,omitempty"`
	Channels    []Channel   `json:"channels"`
	Faults      []Fault     `json:"faults"`
	Effectors   []Effector  `json:"effectors,omitempty"`
	Profiles    []Profile   `json:"profiles"`
	GroundTruth GroundTruth `json:"ground_truth"`
}

// Axes is the twelve-axis property vector.
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

// Entities describes the entity population.
type Entities struct {
	IDTemplate string `json:"id_template"`
	EntityType string `json:"entity_type,omitempty"`
	Count      struct {
		Default int `json:"default"`
		Min     int `json:"min,omitempty"`
		Max     int `json:"max,omitempty"`
	} `json:"count"`
	Churn *Churn `json:"churn,omitempty"`
}

// Churn models entity birth and death during a run.
type Churn struct {
	BirthsPerHour float64 `json:"births_per_hour,omitempty"`
	MeanLifetimeS float64 `json:"mean_lifetime_s,omitempty"`
}

// State is one hidden world variable. Nothing outside the director role may
// ever read these.
type State struct {
	Name        string  `json:"name"`
	Unit        string  `json:"unit,omitempty"`
	Initial     float64 `json:"initial"`
	Min         float64 `json:"min,omitempty"`
	Max         float64 `json:"max,omitempty"`
	Description string  `json:"description,omitempty"`
}

// Dynamics describes how one hidden state evolves.
type Dynamics struct {
	Target string `json:"target"`
	Tier   string `json:"tier"`
	DTMs   int64  `json:"dt_ms,omitempty"`
	F0     *F0Dyn `json:"f0,omitempty"`
	F1     *F1Dyn `json:"f1,omitempty"`
	F2     *F2Dyn `json:"f2,omitempty"`
	F3     *F3Dyn `json:"f3,omitempty"`
}

// F0Dyn is a statistical generator: baseline + trend + seasonality + drift.
type F0Dyn struct {
	Baseline     float64       `json:"baseline,omitempty"`
	TrendPerHour float64       `json:"trend_per_hour,omitempty"`
	Seasonality  []Seasonality `json:"seasonality,omitempty"`
	DriftPerHour float64       `json:"drift_per_hour,omitempty"`
}

// Seasonality is one sinusoidal component.
type Seasonality struct {
	PeriodS   float64 `json:"period_s"`
	Amplitude float64 `json:"amplitude"`
	PhaseRad  float64 `json:"phase_rad,omitempty"`
}

// F1Dyn is a first-order dynamics form integrated with fixed-step RK4.
type F1Dyn struct {
	Form           string    `json:"form"`
	TimeConstantS  float64   `json:"time_constant_s,omitempty"`
	TimeConstant2S float64   `json:"time_constant_2_s,omitempty"`
	DeadTimeS      float64   `json:"dead_time_s,omitempty"`
	Gain           float64   `json:"gain,omitempty"`
	Inputs         []F1Input `json:"inputs,omitempty"`
	Threshold      float64   `json:"threshold,omitempty"`
	Direction      string    `json:"direction,omitempty"`
	ThresholdLow   float64   `json:"threshold_low,omitempty"`
	ThresholdHigh  float64   `json:"threshold_high,omitempty"`
	Clamp          *Clamp    `json:"clamp,omitempty"`
}

// F1Input is a driving state, with an optional coefficient. The schema
// allows a bare string (coefficient 1.0) or an object.
type F1Input struct {
	State   string  `json:"state"`
	Coef    float64 `json:"coef,omitempty"`
	CoefSet bool    `json:"-"`
}

// UnmarshalJSON accepts both the bare-string and object forms.
func (f *F1Input) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		f.State = s
		f.Coef = 1
		f.CoefSet = true
		return nil
	}
	type alias F1Input
	var a alias
	if err := json.Unmarshal(b, &a); err != nil {
		return fmt.Errorf("UnmarshalJSON: %w", err)
	}
	f.State = a.State
	f.Coef = a.Coef
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(b, &fields); err != nil {
		return fmt.Errorf("UnmarshalJSON: %w", err)
	}
	f.CoefSet = false
	if _, ok := fields["coef"]; ok {
		f.CoefSet = true
	}
	if !f.CoefSet {
		f.Coef = 1
	}
	return nil
}

// Clamp is the value interval a state is kept within.
type Clamp struct {
	Min *float64 `json:"min,omitempty"`
	Max *float64 `json:"max,omitempty"`
}

// F2Dyn is a named reference model from the standard library.
type F2Dyn struct {
	Model  string            `json:"model"`
	Params map[string]any    `json:"params,omitempty"`
	Inputs map[string]string `json:"inputs,omitempty"`
}

// F3Dyn is the FMI co-simulation seam: specified, not implemented.
type F3Dyn struct {
	FMU      string            `json:"fmu"`
	Instance string            `json:"instance,omitempty"`
	Inputs   map[string]string `json:"inputs,omitempty"`
	Outputs  map[string]string `json:"outputs,omitempty"`
}

// Channel is one observation channel; it becomes the native event's
// `channel` field.
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

// Range is the declared nameplate range.
type Range struct {
	Min float64 `json:"min,omitempty"`
	Max float64 `json:"max,omitempty"`
}

// Bias lets a hidden state corrupt a channel's reading (a fouled probe).
type Bias struct {
	State            string  `json:"state"`
	Coef             float64 `json:"coef"`
	NoiseScaleAtFull float64 `json:"noise_scale_at_full,omitempty"`
}

// Cadence is the emission schedule of a channel.
type Cadence struct {
	Mode         string  `json:"mode"`
	PeriodS      float64 `json:"period_s,omitempty"`
	JitterS      float64 `json:"jitter_s,omitempty"`
	Deadband     float64 `json:"deadband,omitempty"`
	BatchSize    int     `json:"batch_size,omitempty"`
	TriggerState string  `json:"trigger_state,omitempty"`
}

// Noise is the observation noise model.
type Noise struct {
	Model string  `json:"model"`
	Sigma float64 `json:"sigma"`
}

// Drift is a slow calibration drift on the reading.
type Drift struct {
	RatePerHour float64 `json:"rate_per_hour,omitempty"`
	Model       string  `json:"model,omitempty"`
}

// Availability is a producer up/down renewal model.
type Availability struct {
	Uptime float64 `json:"uptime,omitempty"`
	MTTRS  float64 `json:"mttr_s,omitempty"`
}

// LinkDelay models transport delay: observed_time = event_time + delay.
type LinkDelay struct {
	Model  string  `json:"model"`
	MeanS  float64 `json:"mean_s,omitempty"`
	SigmaS float64 `json:"sigma_s,omitempty"`
	MaxS   float64 `json:"max_s,omitempty"`
}

// Fault is a world fault: it changes hidden state and is what the reasoner
// must diagnose.
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

// FaultOnset is the onset shape of a fault's deviation.
type FaultOnset struct {
	Shape       string  `json:"shape"`
	RatePerHour float64 `json:"rate_per_hour,omitempty"`
	Magnitude   float64 `json:"magnitude,omitempty"`
	DutyCycle   float64 `json:"duty_cycle,omitempty"`
}

// FaultEffect is one state delta a fault applies.
type FaultEffect struct {
	State          string  `json:"state"`
	Delta          float64 `json:"delta"`
	Multiplicative bool    `json:"multiplicative,omitempty"`
}

// Observability declares how detectability is computed: a detector
// expression, not a channel name.
type Observability struct {
	Detector           Detector `json:"detector"`
	FirstObservableSNR float64  `json:"first_observable_snr,omitempty"`
	UnavoidableSNR     float64  `json:"unavoidable_snr,omitempty"`
}

// Detector is one of the four detector expressions.
type Detector struct {
	Form     string   `json:"form"`
	Channel  string   `json:"channel,omitempty"`
	ChannelA string   `json:"channel_a,omitempty"`
	ChannelB string   `json:"channel_b,omitempty"`
	Inputs   []string `json:"inputs,omitempty"`
	Outputs  []string `json:"outputs,omitempty"`
}

// Counterfactual describes what happens with and without action.
type Counterfactual struct {
	IfNoAction         string `json:"if_no_action,omitempty"`
	IfActionByDeadline string `json:"if_action_by_deadline,omitempty"`
}

// Effector is an actuation endpoint, invoked through the single generic
// sim.effector.invoke tool. The binary contains no effector name.
type Effector struct {
	Name                 string         `json:"name"`
	Description          string         `json:"description,omitempty"`
	ArgsSchema           map[string]any `json:"args_schema"`
	RiskClass            string         `json:"risk_class,omitempty"`
	Ack                  Ack            `json:"ack"`
	ConfirmationChannels []string       `json:"confirmation_channels,omitempty"`
	Interlock            *Interlock     `json:"interlock,omitempty"`
	Effect               Effect         `json:"effect"`
	IdempotencyWindowS   float64        `json:"idempotency_window_s,omitempty"`
}

// Ack describes ack latency and probabilistic failure modes.
type Ack struct {
	LatencyMS    Latency       `json:"latency_ms"`
	FailureModes []FailureMode `json:"failure_modes,omitempty"`
}

// Latency is a mean/sigma pair for ack latency.
type Latency struct {
	Mean  float64 `json:"mean"`
	Sigma float64 `json:"sigma,omitempty"`
}

// FailureMode is one probabilistic effector failure mode.
type FailureMode struct {
	Mode        string  `json:"mode"`
	Probability float64 `json:"probability"`
}

// Interlock is an independent safety system that may refuse.
type Interlock struct {
	State            string  `json:"state"`
	Operator         string  `json:"operator"`
	Threshold        float64 `json:"threshold"`
	AutonomousAction *struct {
		State string  `json:"state"`
		Delta float64 `json:"delta"`
	} `json:"autonomous_action,omitempty"`
}

// Effect is the physical consequence of an effector call.
type Effect struct {
	StateDeltas   []StateDelta `json:"state_deltas"`
	TimeConstantS float64      `json:"time_constant_s,omitempty"`
	DeadTimeS     float64      `json:"dead_time_s,omitempty"`
}

// StateDelta is one state change; magnitude may come from an argument.
type StateDelta struct {
	State   string  `json:"state"`
	Delta   float64 `json:"delta"`
	FromArg string  `json:"from_arg,omitempty"`
}

// Profile is a named scenario family.
type Profile struct {
	Name                 string             `json:"name"`
	Description          string             `json:"description,omitempty"`
	NotApplicable        string             `json:"not_applicable,omitempty"`
	FaultWeights         map[string]float64 `json:"fault_weights,omitempty"`
	EntityCount          int                `json:"entity_count,omitempty"`
	DurationS            float64            `json:"duration_s,omitempty"`
	SimultaneousEntities int                `json:"simultaneous_entities,omitempty"`
	Setup                []ProfileSetup     `json:"setup,omitempty"`
}

// ProfileSetup declares context calls made before a scenario fault. The
// scenario entity is substituted for the {entity_id} placeholder in string
// arguments, keeping setup data-defined rather than domain-coded.
type ProfileSetup struct {
	Effector  string         `json:"effector"`
	CommandID string         `json:"command_id,omitempty"`
	Args      map[string]any `json:"args,omitempty"`
}

// GroundTruth declares suite composition targets.
type GroundTruth struct {
	NegativeClassFraction float64 `json:"negative_class_fraction"`
	PreDegradedFraction   float64 `json:"pre_degraded_fraction,omitempty"`
	MinScenarios          int     `json:"min_scenarios,omitempty"`
	MinPerLabel           int     `json:"min_per_label,omitempty"`
}
