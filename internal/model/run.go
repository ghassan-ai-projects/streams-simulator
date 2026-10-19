package model

import (
	"encoding/json"
	"runtime"
)

type RunArtifact struct {
	SchemaVersion        string          `json:"schema_version"`
	SimVersion           string          `json:"sim_version"`
	RunID                string          `json:"run_id,omitempty"`
	Label                string          `json:"label,omitempty"`
	CreatedAt            string          `json:"created_at,omitempty"`
	Domain               ArtifactRef     `json:"domain"`
	DomainSpec           json.RawMessage `json:"domain_spec,omitempty"`
	Seed                 uint64          `json:"seed"`
	Sink                 string          `json:"sink"`
	Adapter              ArtifactRef     `json:"adapter"`
	AdapterSpec          json.RawMessage `json:"adapter_spec,omitempty"`
	TimeMode             string          `json:"time_mode"`
	WorldConfig          WorldConfig     `json:"world_config"`
	WorldDigest          string          `json:"world_digest,omitempty"`
	CommandLog           []Command       `json:"command_log"`
	ExpectedTraceDigest  string          `json:"expected_trace_digest,omitempty"`
	Reproducible         bool            `json:"reproducible"`
	Incomplete           bool            `json:"incomplete,omitempty"`
	Error                string          `json:"error,omitempty"`
	Unblinded            bool            `json:"unblinded,omitempty"`
	UnblindedAt          string          `json:"unblinded_at,omitempty"`
	Platform             Platform        `json:"platform,omitempty"`
	Counts               Counts          `json:"counts,omitempty"`
	AppliedPerturbations []string        `json:"applied_perturbations,omitempty"`

	Mode string `json:"mode,omitempty"`

	JournalDigest       string `json:"journal_digest,omitempty"`
	SealedJournalDigest string `json:"sealed_journal_digest,omitempty"`
	JournalFileDigest   string `json:"journal_file_digest,omitempty"`
	SealedFileDigest    string `json:"sealed_journal_file_digest,omitempty"`
	LedgerDigest        string `json:"ledger_digest,omitempty"`
	HistoryDigest       string `json:"history_digest,omitempty"`

	FieldTraceDigest string `json:"field_trace_digest,omitempty"`

	InterruptLogDigest string `json:"interrupt_log_digest,omitempty"`
}

type ArtifactRef struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Digest  string `json:"digest"`
}

type WorldConfig struct {
	StartTimeNS     int64            `json:"start_time_ns"`
	EntityIDs       []string         `json:"entity_ids"`
	ScenarioProfile string           `json:"scenario_profile,omitempty"`
	Sinks           []map[string]any `json:"sinks,omitempty"`
	ClockMultiplier float64          `json:"clock_multiplier,omitempty"`

	Noiseless        bool   `json:"noiseless,omitempty"`
	ForceFailureMode string `json:"force_failure_mode,omitempty"`
}

type Command struct {
	Seq          int64          `json:"seq"`
	AtNS         int64          `json:"at_ns"`
	Op           string         `json:"op"`
	Args         map[string]any `json:"args"`
	ResultDigest string         `json:"result_digest,omitempty"`
}

const (
	OpWorldCreate    = "world.create"
	OpEntityAdd      = "entity.add"
	OpEntityRetire   = "entity.retire"
	OpClockAdvance   = "clock.advance"
	OpClockRun       = "clock.run"
	OpFaultInject    = "fault.inject"
	OpFaultClear     = "fault.clear"
	OpPerturbApply   = "perturb.apply"
	OpPerturbClear   = "perturb.clear"
	OpEnvInject      = "env.inject"
	OpEffectorInvoke = "effector.invoke"
	OpRunBegin       = "run.begin"
	OpRunEnd         = "run.end"
)

const (
	SinkInproc   = "inproc"
	SinkFile     = "file"
	SinkHTTPPush = "http-push"
)

const (
	TimeStepped = "stepped"
	TimeWall    = "wall"
)

const (
	ProfileNominal           = "nominal"
	ProfileCorrelatedCascade = "correlated_cascade"
	ProfileSensorPathology   = "sensor_pathology"
)

type Platform struct {
	GOOS      string `json:"goos,omitempty"`
	GOARCH    string `json:"goarch,omitempty"`
	GoVersion string `json:"go_version,omitempty"`
}

func CurrentPlatform() Platform {
	return Platform{GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, GoVersion: runtime.Version()}
}

type Counts struct {
	Emitted         int64 `json:"emitted,omitempty"`
	Perturbed       int64 `json:"perturbed,omitempty"`
	DroppedByDesign int64 `json:"dropped_by_design,omitempty"`
	EffectorCalls   int64 `json:"effector_calls,omitempty"`
	FaultsInjected  int64 `json:"faults_injected,omitempty"`
}

type StateSnapshot struct {
	Seq    int64              `json:"seq"`
	TimeNS int64              `json:"time_ns"`
	Entity string             `json:"entity_id"`
	States map[string]float64 `json:"states"`
}
