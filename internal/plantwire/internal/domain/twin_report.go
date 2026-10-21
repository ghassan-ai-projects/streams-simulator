package domain

import "errors"

const (
	TwinReportVersion = "0.1"

	VerdictAgrees      = "agrees"
	VerdictDiverges    = "diverges"
	VerdictCannotJudge = "cannot_judge"

	ReasonNoCommands = "no_commands_applied"
	ReasonNoReadings = "no_readings_compared"

	ReasonNodeStopNotApplied  = "node_stop_not_applied"
	ReasonNodeStopUnconfirmed = "node_stop_unconfirmed"

	DefaultSpanFraction = 0.05
	MaxSeriesPoints     = 2000
	maxDetailBytes      = 200
	MaxTwinSpanNS       = int64(7 * 24 * 3600 * 1e9)
)

const (
	ToleranceDeclared    = "declared"
	ToleranceDefaultSpan = "default_span"
	ToleranceNone        = "none"
)

const (
	SkipQuality       = "quality_not_good"
	SkipHistorical    = "historical"
	SkipNotNumeric    = "not_numeric"
	SkipNoEntity      = "no_twin_entity"
	SkipNoChannel     = "no_twin_channel"
	SkipNoTolerance   = "no_tolerance_declared"
	SkipBeforeCommand = "before_first_command"
	SkipTwinNotFinite = "twin_not_finite"
)

var (
	ErrTwinJournal   = errors.New("plantwire: the run journal cannot be read for its commands")
	ErrTwinTolerance = errors.New("plantwire: invalid tolerance")
	ErrTwinKinds     = errors.New("plantwire: the twin domain implements none of the board's kinds")
	ErrTwinNoDevices = errors.New("plantwire: the field trace announces no device")
	ErrTwinSpan      = errors.New("plantwire: the field trace spans more time than a twin is run for")
)

type TwinReport struct {
	SchemaVersion string         `json:"schema_version"`
	Verdict       string         `json:"verdict"`
	Reason        string         `json:"reason,omitempty"`
	Summary       TwinSummary    `json:"summary"`
	Run           TwinEvidence   `json:"run"`
	Twin          TwinInfo       `json:"twin"`
	ToleranceRule ToleranceRule  `json:"tolerance_rule"`
	Commands      []TwinCommand  `json:"commands"`
	Metrics       []MetricReport `json:"metrics"`
	Divergences   []Divergence   `json:"divergences"`
}

type TwinSummary struct {
	Devices            int `json:"devices"`
	Metrics            int `json:"metrics"`
	Compared           int `json:"compared"`
	Within             int `json:"within"`
	Diverged           int `json:"diverged"`
	NotCompared        int `json:"not_compared"`
	Unattributed       int `json:"unattributed_readings"`
	MetricsCannotJudge int `json:"metrics_cannot_judge"`
	Divergences        int `json:"divergences"`
	CommandsApplied    int `json:"commands_applied"`
	CommandsNotApplied int `json:"commands_not_applied"`
}

type TwinEvidence struct {
	FieldTraceDigest   string `json:"field_trace_digest"`
	JournalDigest      string `json:"journal_digest"`
	InterruptLogDigest string `json:"interrupt_log_digest,omitempty"`
	EvidenceBound      bool   `json:"evidence_bound"`
}

type TwinKind struct {
	Kind        string `json:"kind"`
	KindDigest  string `json:"kind_digest"`
	BoardType   string `json:"board_type,omitempty"`
	BoardDigest string `json:"board_digest,omitempty"`
}

type TwinInfo struct {
	Domain       string     `json:"domain"`
	DomainDigest string     `json:"domain_digest"`
	Kinds        []TwinKind `json:"kinds"`
	StartNS      int64      `json:"start_ns"`
}

type ToleranceRule struct {
	Default      string  `json:"default"`
	SpanFraction float64 `json:"span_fraction"`
	Description  string  `json:"description"`
}

type Tolerance struct {
	Value  *float64 `json:"value"`
	Source string   `json:"source"`
}

type MetricSummary struct {
	Compared    int `json:"compared"`
	Within      int `json:"within"`
	Diverged    int `json:"diverged"`
	NotCompared int `json:"not_compared"`
}

type ReasonCount struct {
	Reason string `json:"reason"`
	Count  int    `json:"count"`
}

type SeriesPoint struct {
	AtNS  int64   `json:"at_ns"`
	Board float64 `json:"board"`
	Twin  float64 `json:"twin"`
}

type Series struct {
	Points    []SeriesPoint `json:"points"`
	Total     int           `json:"total"`
	Truncated bool          `json:"truncated"`
}

type MetricReport struct {
	Device      string        `json:"device"`
	Kind        string        `json:"kind"`
	Metric      string        `json:"metric"`
	Unit        string        `json:"unit,omitempty"`
	Tolerance   Tolerance     `json:"tolerance"`
	Verdict     string        `json:"verdict"`
	Summary     MetricSummary `json:"summary"`
	NotCompared []ReasonCount `json:"not_compared_reasons"`
	Series      Series        `json:"series"`
}

type Divergence struct {
	Name       string   `json:"name"`
	Device     string   `json:"device"`
	Metric     string   `json:"metric"`
	FirstNS    int64    `json:"first_ns"`
	LastNS     int64    `json:"last_ns"`
	Readings   int      `json:"readings"`
	Tolerance  float64  `json:"tolerance"`
	WorstError float64  `json:"worst_error"`
	WorstAtNS  int64    `json:"worst_at_ns"`
	BoardValue float64  `json:"board_value"`
	TwinValue  float64  `json:"twin_value"`
	Commands   []string `json:"commands"`
}
