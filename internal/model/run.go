package model

import (
	"encoding/json"
	"runtime"
)

// RunArtifact is the single file that reproduces a run, matching
// docs/contracts/run-artifact-v0.1.schema.json. If a failure cannot be
// reduced to one of these, it is not reproducible and the architecture is
// wrong.
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
}

// ArtifactRef identifies a versioned data artifact (domain spec or adapter)
// by digest.
type ArtifactRef struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Digest  string `json:"digest"`
}

// WorldConfig pins the parts of world creation that enter the determinism
// tuple.
type WorldConfig struct {
	StartTimeNS     int64            `json:"start_time_ns"`
	EntityIDs       []string         `json:"entity_ids"`
	ScenarioProfile string           `json:"scenario_profile,omitempty"`
	Sinks           []map[string]any `json:"sinks,omitempty"`
	ClockMultiplier float64          `json:"clock_multiplier,omitempty"`
}

// Command is one mutation in the total order that reproduces a run.
type Command struct {
	Seq          int64          `json:"seq"`
	AtNS         int64          `json:"at_ns"`
	Op           string         `json:"op"`
	Args         map[string]any `json:"args"`
	ResultDigest string         `json:"result_digest,omitempty"`
}

// Command operation names, matching the MCP director tools minus the "sim."
// prefix.
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

// Sink names.
const (
	SinkInproc   = "inproc"
	SinkFile     = "file"
	SinkHTTPPush = "http-push"
	SinkBroker   = "broker"
)

// Time modes.
const (
	TimeStepped = "stepped"
	TimeScaled  = "scaled"
	TimeWall    = "wall"
)

// Platform is recorded so cross-architecture digest divergence is
// diagnosable. Not part of the determinism tuple.
type Platform struct {
	GOOS      string `json:"goos,omitempty"`
	GOARCH    string `json:"goarch,omitempty"`
	GoVersion string `json:"go_version,omitempty"`
}

// CurrentPlatform reports the build platform.
func CurrentPlatform() Platform {
	return Platform{GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, GoVersion: runtime.Version()}
}

// Counts summarizes a run.
type Counts struct {
	Emitted         int64 `json:"emitted,omitempty"`
	Perturbed       int64 `json:"perturbed,omitempty"`
	DroppedByDesign int64 `json:"dropped_by_design,omitempty"`
	EffectorCalls   int64 `json:"effector_calls,omitempty"`
	FaultsInjected  int64 `json:"faults_injected,omitempty"`
}
