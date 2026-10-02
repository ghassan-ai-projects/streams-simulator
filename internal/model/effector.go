package model

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
	State         string  `json:"state"`
	Delta         float64 `json:"delta"`
	FromArg       string  `json:"from_arg,omitempty"`
	AssignFromArg string  `json:"assign_from_arg,omitempty"`
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
