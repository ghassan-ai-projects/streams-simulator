// Package domain holds the run report vocabulary, its assembly from a parsed run folder and its renderings.
package domain

import (
	"encoding/json"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const SchemaVersion = "0.1"

type Report struct {
	SchemaVersion    string       `json:"schema_version"`
	Run              Run          `json:"run"`
	Digests          *Digests     `json:"digests,omitempty"`
	Verification     Verification `json:"verification"`
	Counts           Counts       `json:"counts"`
	Truth            *Truth       `json:"truth,omitempty"`
	TruthUnavailable string       `json:"truth_unavailable,omitempty"`
	TornTail         []string     `json:"torn_tail,omitempty"`
	Consumer         *Consumer    `json:"consumer,omitempty"`
	Scoring          *Scoring     `json:"scoring,omitempty"`
}

type Run struct {
	RunID            string             `json:"run_id,omitempty"`
	Mode             string             `json:"mode"`
	Complete         bool               `json:"complete"`
	IncompleteReason string             `json:"incomplete_reason,omitempty"`
	Ended            bool               `json:"ended"`
	Unblinded        bool               `json:"unblinded"`
	SimVersion       string             `json:"sim_version,omitempty"`
	Domain           *model.ArtifactRef `json:"domain,omitempty"`
	Adapter          *model.ArtifactRef `json:"adapter,omitempty"`
	Seed             *uint64            `json:"seed,omitempty"`
	Sink             string             `json:"sink,omitempty"`
	TimeMode         string             `json:"time_mode,omitempty"`
	StartTimeNS      int64              `json:"start_time_ns,omitempty"`
	EndTimeNS        int64              `json:"end_time_ns,omitempty"`
	EntityCount      int                `json:"entity_count,omitempty"`
}

type Digests struct {
	World         string `json:"world,omitempty"`
	Trace         string `json:"trace,omitempty"`
	Journal       string `json:"journal,omitempty"`
	SealedJournal string `json:"sealed_journal,omitempty"`
	Ledger        string `json:"ledger,omitempty"`
	History       string `json:"history,omitempty"`
}

type Verification struct {
	Status string  `json:"status"`
	Checks []Check `json:"checks"`
}

type Check struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Recorded string `json:"recorded"`
	Computed string `json:"computed,omitempty"`
}

type Counts struct {
	Emitted           *int64           `json:"emitted,omitempty"`
	Observed          *int64           `json:"observed,omitempty"`
	LedgerRows        *int64           `json:"ledger_rows,omitempty"`
	Delivered         *int64           `json:"delivered,omitempty"`
	Dropped           *int64           `json:"dropped,omitempty"`
	DeliveredByReason map[string]int64 `json:"delivered_by_reason,omitempty"`
	DroppedByReason   map[string]int64 `json:"dropped_by_reason,omitempty"`
}

type Truth struct {
	Faults        []Fault         `json:"faults"`
	Perturbations []Perturbation  `json:"perturbations"`
	Environment   []Environment   `json:"environment"`
	EffectorCalls []EffectorCall  `json:"effector_calls"`
	KPIs          *model.KPITruth `json:"kpis,omitempty"`
}

type Fault struct {
	Seq      int64          `json:"seq"`
	AtNS     int64          `json:"at_ns"`
	Action   string         `json:"action"`
	EntityID string         `json:"entity_id,omitempty"`
	Fault    string         `json:"fault,omitempty"`
	FaultID  string         `json:"fault_id,omitempty"`
	OnsetNS  *int64         `json:"onset_ns,omitempty"`
	Params   map[string]any `json:"params,omitempty"`
}

type Perturbation struct {
	Seq          int64          `json:"seq"`
	AtNS         int64          `json:"at_ns"`
	Action       string         `json:"action"`
	Perturbation string         `json:"perturbation,omitempty"`
	PerturbID    string         `json:"perturb_id,omitempty"`
	FromNS       *int64         `json:"from_ns,omitempty"`
	UntilNS      *int64         `json:"until_ns,omitempty"`
	Params       map[string]any `json:"params,omitempty"`
}

type Environment struct {
	Seq    int64          `json:"seq"`
	AtNS   int64          `json:"at_ns"`
	Target string         `json:"target"`
	Fault  string         `json:"fault"`
	Params map[string]any `json:"params,omitempty"`
}

type EffectorCall struct {
	Seq       int64          `json:"seq"`
	AtNS      int64          `json:"at_ns"`
	CommandID string         `json:"command_id"`
	Effector  string         `json:"effector"`
	EntityID  string         `json:"entity_id,omitempty"`
	Args      map[string]any `json:"args,omitempty"`
	Outcome   *Outcome       `json:"outcome,omitempty"`
}

type Outcome struct {
	Accepted      *bool    `json:"accepted,omitempty"`
	Mode          string   `json:"mode,omitempty"`
	EffectApplied *bool    `json:"effect_applied,omitempty"`
	AckLatencyMS  *float64 `json:"ack_latency_ms,omitempty"`
	Refusal       string   `json:"refusal,omitempty"`
}

type Consumer struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	Arm          string `json:"arm,omitempty"`
	ConfigDigest string `json:"config_digest,omitempty"`
}

type Scoring struct {
	Status    string          `json:"status"`
	Reason    string          `json:"reason,omitempty"`
	Scorecard json.RawMessage `json:"scorecard,omitempty"`
}

const (
	FilesMatch   = "files_match"
	FilesDiffer  = "files_differ"
	NotChecked   = "not_checked"
	CheckMatch   = "match"
	CheckDiffers = "mismatch"
	CheckMissing = "missing_file"
	Scored       = "scored"
	NotScored    = "not_scored"
	Failed       = "failed"
)

type Document struct {
	Raw     []byte
	Present bool
}

type Folder struct {
	Mismatch string
	Artifact *model.RunArtifact
	Verdict  *model.Verdict
	Trace    Document
	Observer Document
	Director Document
	Ledger   Document
}
