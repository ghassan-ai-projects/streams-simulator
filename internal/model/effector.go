package model

type Effector struct {
	Name                 string         `json:"name"`
	Description          string         `json:"description,omitempty"`
	ArgsSchema           map[string]any `json:"args_schema"`
	RiskClass            string         `json:"risk_class,omitempty"`
	SafeStop             bool           `json:"safe_stop,omitempty"`
	Ack                  Ack            `json:"ack"`
	ConfirmationChannels []string       `json:"confirmation_channels,omitempty"`
	Interlock            *Interlock     `json:"interlock,omitempty"`
	Requires             []Condition    `json:"requires,omitempty"`
	Responder            *Responder     `json:"responder,omitempty"`
	Lease                *Lease         `json:"lease,omitempty"`
	Effect               Effect         `json:"effect"`
	IdempotencyWindowS   float64        `json:"idempotency_window_s,omitempty"`
}

type Ack struct {
	LatencyMS    Latency       `json:"latency_ms"`
	FailureModes []FailureMode `json:"failure_modes,omitempty"`
}

type Latency struct {
	Mean  float64 `json:"mean"`
	Sigma float64 `json:"sigma,omitempty"`
}

type FailureMode struct {
	Mode        string  `json:"mode"`
	Probability float64 `json:"probability"`
}

type Interlock struct {
	State            string  `json:"state"`
	Operator         string  `json:"operator"`
	Threshold        float64 `json:"threshold"`
	AutonomousAction *struct {
		State string  `json:"state"`
		Delta float64 `json:"delta"`
	} `json:"autonomous_action,omitempty"`
}

type Effect struct {
	StateDeltas   []StateDelta `json:"state_deltas"`
	TimeConstantS float64      `json:"time_constant_s,omitempty"`
	DeadTimeS     float64      `json:"dead_time_s,omitempty"`
}

type StateDelta struct {
	State         string  `json:"state"`
	Delta         float64 `json:"delta"`
	FromArg       string  `json:"from_arg,omitempty"`
	AssignFromArg string  `json:"assign_from_arg,omitempty"`
}

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

type ProfileSetup struct {
	Effector  string         `json:"effector"`
	CommandID string         `json:"command_id,omitempty"`
	Args      map[string]any `json:"args,omitempty"`
}

type GroundTruth struct {
	NegativeClassFraction float64 `json:"negative_class_fraction"`
	PreDegradedFraction   float64 `json:"pre_degraded_fraction,omitempty"`
	MinScenarios          int     `json:"min_scenarios,omitempty"`
	MinPerLabel           int     `json:"min_per_label,omitempty"`
}

type EffectorAcknowledgement struct {
	Accepted    bool   `json:"accepted"`
	Simulated   bool   `json:"simulated"`
	WorldID     string `json:"world_id"`
	CommandID   string `json:"command_id"`
	EffectETANS int64  `json:"effect_eta_ns,omitempty"`
	Reason      string `json:"reason,omitempty"`
}
