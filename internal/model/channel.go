package model

type Channel struct {
	Name               string         `json:"name"`
	Description        string         `json:"description,omitempty"`
	ValueType          string         `json:"value_type"`
	Unit               string         `json:"unit,omitempty"`
	EnumValues         []string       `json:"enum_values,omitempty"`
	Resolution         float64        `json:"resolution"`
	Range              *Range         `json:"range,omitempty"`
	Observes           string         `json:"observes,omitempty"`
	ObservationGain    float64        `json:"observation_gain,omitempty"`
	ObservationOffset  float64        `json:"observation_offset,omitempty"`
	ObservationBias    *Bias          `json:"observation_bias,omitempty"`
	Fidelity           string         `json:"fidelity"`
	Tag                string         `json:"tag,omitempty"`
	ISO13374Role       string         `json:"iso13374_role,omitempty"`
	Cadence            Cadence        `json:"cadence"`
	Noise              Noise          `json:"noise"`
	Drift              *Drift         `json:"drift,omitempty"`
	Absence            string         `json:"absence"`
	Availability       *Availability  `json:"availability,omitempty"`
	LinkDelay          *LinkDelay     `json:"link_delay,omitempty"`
	AttackerControlled bool           `json:"attacker_controlled,omitempty"`
	States             []ChannelState `json:"states,omitempty"`
}

const (
	AttentionAlarm = "alarm"
	AttentionWarn  = "warn"
)

type ChannelState struct {
	Value     int     `json:"value"`
	Name      string  `json:"name"`
	Attention *string `json:"attention"`
}

type Range struct {
	Min float64 `json:"min,omitempty"`
	Max float64 `json:"max,omitempty"`
}

type Bias struct {
	State            string   `json:"state"`
	Coef             float64  `json:"coef"`
	NoiseScaleAtFull *float64 `json:"noise_scale_at_full,omitempty"`
}

func (b *Bias) NoiseScale() float64 {
	if b.NoiseScaleAtFull == nil {
		return 1
	}
	return *b.NoiseScaleAtFull
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
