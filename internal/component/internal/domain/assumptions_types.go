package domain

const (
	OriginTwin        = "twin"
	OriginComposition = "composition"
	OriginDomain      = "domain"

	StatusInSync    = "in_sync"
	StatusOutOfSync = "out_of_sync"

	BasisLibraryInstrument = "library_instrument"
	BasisPlainChannel      = "plain_channel"
	BasisDomainDeclared    = "domain_declared"

	RecordedComposition = "composition"
	RecordedTwinChoices = "twin_choices"
	RecordedDomainFile  = "domain_file"

	ResponseDirect        = responseDirect
	ResponseFirstOrder    = responseFirstOrder
	ResponseNone          = responseNone
	ResponseFixedSafeStop = "fixed_safe_stop"
	ResponseDeclared      = "declared"

	assumptionsVersion = "0.1"
)

type Assumptions struct {
	SchemaVersion string            `json:"schema_version"`
	Domain        DomainIdentity    `json:"domain"`
	Origin        AssumptionsOrigin `json:"origin"`
	Kinds         []KindAssumptions `json:"kinds"`
	NotModeled    []NotModeled      `json:"not_modeled"`
}

type DomainIdentity struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Digest  string `json:"digest"`
}

type AssumptionsOrigin struct {
	Kind string      `json:"kind"`
	Twin *TwinOrigin `json:"twin,omitempty"`
}

type TwinOrigin struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Status  string `json:"status"`
}

type NotModeled struct {
	Code     string `json:"code"`
	Scope    string `json:"scope"`
	Sentence string `json:"sentence"`
}

type KindAssumptions struct {
	Kind       string               `json:"kind"`
	Implements string               `json:"implements,omitempty"`
	Instances  []InstanceAssumption `json:"instances"`
	Readings   []ReadingAssumption  `json:"readings"`
	Dynamics   []DynamicsAssumption `json:"dynamics"`
	Commands   []CommandAssumption  `json:"commands"`
	Faults     []FaultAssumption    `json:"faults"`
}

type InstanceAssumption struct {
	Names        []string `json:"names,omitempty"`
	IDTemplate   string   `json:"id_template,omitempty"`
	CountDefault int      `json:"count_default,omitempty"`
	Area         string   `json:"area,omitempty"`
}

type RangeValue struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type NoiseValue struct {
	Model string  `json:"model"`
	Sigma float64 `json:"sigma"`
}

type CadenceValue struct {
	Mode     string   `json:"mode"`
	PeriodS  *float64 `json:"period_s,omitempty"`
	Deadband *float64 `json:"deadband,omitempty"`
}

type AvailabilityValue struct {
	Uptime float64 `json:"uptime"`
	MTTRS  float64 `json:"mttr_s"`
}

type InstrumentValue struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
	Ref     string `json:"ref"`
	Digest  string `json:"digest,omitempty"`
}

type ReadingAssumption struct {
	Channel                string             `json:"channel"`
	Observes               string             `json:"observes,omitempty"`
	Unit                   string             `json:"unit,omitempty"`
	ValueType              string             `json:"value_type"`
	Range                  *RangeValue        `json:"range,omitempty"`
	Resolution             float64            `json:"resolution"`
	Initial                *float64           `json:"initial,omitempty"`
	Noise                  NoiseValue         `json:"noise"`
	Cadence                CadenceValue       `json:"cadence"`
	Availability           *AvailabilityValue `json:"availability,omitempty"`
	Absence                string             `json:"absence"`
	Tag                    string             `json:"tag,omitempty"`
	Basis                  string             `json:"basis"`
	Instrument             *InstrumentValue   `json:"instrument,omitempty"`
	RecordedIn             string             `json:"recorded_in"`
	EqualsBinarySuggestion *bool              `json:"equals_binary_suggestion,omitempty"`
}

type InputValue struct {
	State string  `json:"state"`
	Coef  float64 `json:"coef"`
	Link  string  `json:"link,omitempty"`
}

type ClampValue struct {
	Min *float64 `json:"min,omitempty"`
	Max *float64 `json:"max,omitempty"`
}

type DynamicsAssumption struct {
	Target        string       `json:"target"`
	Tier          string       `json:"tier"`
	Form          string       `json:"form,omitempty"`
	Gain          *float64     `json:"gain,omitempty"`
	TimeConstantS *float64     `json:"time_constant_s,omitempty"`
	DeadTimeS     *float64     `json:"dead_time_s,omitempty"`
	Inputs        []InputValue `json:"inputs"`
	Clamp         *ClampValue  `json:"clamp,omitempty"`
}

type ArgumentValue struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Minimum *float64 `json:"minimum,omitempty"`
	Maximum *float64 `json:"maximum,omitempty"`
}

type SetValue struct {
	Argument string `json:"argument"`
	State    string `json:"state"`
}

type DriveValue struct {
	From          string  `json:"from"`
	Kind          string  `json:"kind"`
	Metric        string  `json:"metric"`
	Gain          float64 `json:"gain"`
	TimeConstantS float64 `json:"time_constant_s"`
}

type EffectValue struct {
	DeadTimeS     float64 `json:"dead_time_s"`
	TimeConstantS float64 `json:"time_constant_s"`
	StateChanges  int     `json:"state_changes"`
}

type FailureModeValue struct {
	Mode        string  `json:"mode"`
	Probability float64 `json:"probability"`
}

type AcknowledgementValue struct {
	LatencyMean  float64            `json:"latency_ms_mean"`
	LatencySigma float64            `json:"latency_ms_sigma"`
	FailureModes []FailureModeValue `json:"failure_modes"`
}

type LeaseValue struct {
	Argument string `json:"argument"`
	Unit     string `json:"unit"`
	Fallback string `json:"fallback,omitempty"`
}

type CommandAssumption struct {
	Name                   string               `json:"name"`
	RiskClass              string               `json:"risk_class,omitempty"`
	SafeStop               bool                 `json:"safe_stop"`
	Arguments              []ArgumentValue      `json:"arguments"`
	Response               string               `json:"response"`
	RecordedIn             string               `json:"recorded_in"`
	Sets                   []SetValue           `json:"sets"`
	Drives                 []DriveValue         `json:"drives"`
	Effect                 EffectValue          `json:"effect"`
	Acknowledgement        AcknowledgementValue `json:"acknowledgement"`
	Lease                  *LeaseValue          `json:"lease,omitempty"`
	EqualsBinarySuggestion *bool                `json:"equals_binary_suggestion,omitempty"`
	Statement              string               `json:"statement"`
}

type FaultAssumption struct {
	ID            string   `json:"id"`
	Description   string   `json:"description,omitempty"`
	NegativeClass bool     `json:"negative_class"`
	OnsetShape    string   `json:"onset_shape"`
	Affects       []string `json:"affects"`
	OnReading     string   `json:"on_reading,omitempty"`
}

type AssumptionsInput struct {
	Domain []byte
	Source string
	Twin   *TwinFiles
}

type TwinFiles struct {
	Choices []byte
	Source  string
	Kinds   map[string]map[string]any
}
