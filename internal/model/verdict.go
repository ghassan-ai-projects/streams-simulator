package model

// Verdict is the typed form of docs/contracts/consumer-verdict-v0.1.schema.json:
// what a consumer submits back after processing a run, in the simulator's
// neutral vocabulary. The simulator never reads a consumer's database, never
// parses its logs, and never links its code.
type Verdict struct {
	SchemaVersion string           `json:"schema_version"`
	RunID         string           `json:"run_id"`
	Consumer      ConsumerInfo     `json:"consumer"`
	Detections    []Detection      `json:"detections,omitempty"`
	Admission     []Admission      `json:"admission,omitempty"`
	Actions       []Action         `json:"actions,omitempty"`
	Counters      map[string]int64 `json:"counters,omitempty"`
	Resource      *Resource        `json:"resource,omitempty"`
}

// ConsumerInfo identifies the submitting consumer.
type ConsumerInfo struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	ConfigDigest string `json:"config_digest,omitempty"`
	Arm          string `json:"arm,omitempty"`
}

// Detection is one consumer conclusion, scored against sealed labels.
type Detection struct {
	EntityID     string   `json:"entity_id"`
	DetectedAt   string   `json:"detected_at"`
	Label        string   `json:"label,omitempty"`
	Confidence   float64  `json:"confidence,omitempty"`
	EvidenceRefs []string `json:"evidence_refs,omitempty"`
	Narrative    string   `json:"narrative,omitempty"`
}

// The seven neutral admission outcomes.
const (
	AdmissionAccepted      = "accepted"
	AdmissionDuplicate     = "duplicate"
	AdmissionConflict      = "conflict"
	AdmissionLate          = "late"
	AdmissionOutOfContract = "out_of_contract"
	AdmissionMalformed     = "malformed"
	AdmissionRejected      = "rejected"
)

// Admission is one event's disposition, as declared by the consumer.
type Admission struct {
	Seq     int64  `json:"seq"`
	Outcome string `json:"outcome"`
	Note    string `json:"note,omitempty"`
}

// Action is a command the consumer believes it issued; cross-checked against
// the simulator's own effector-call log, which is the authority.
type Action struct {
	CommandID       string `json:"command_id"`
	Effector        string `json:"effector"`
	EntityID        string `json:"entity_id,omitempty"`
	IssuedAt        string `json:"issued_at"`
	OutcomeBelieved string `json:"outcome_believed,omitempty"`
}

// Outcome belief values.
const (
	BelievedSucceeded = "succeeded"
	BelievedFailed    = "failed"
	BelievedUnknown   = "unknown"
)

// Resource reports consumer cost, so cost-per-resolved-scenario is
// computable without inspecting a consumer's billing.
type Resource struct {
	ReasoningInvocations int64   `json:"reasoning_invocations,omitempty"`
	InputTokens          int64   `json:"input_tokens,omitempty"`
	OutputTokens         int64   `json:"output_tokens,omitempty"`
	WallSeconds          float64 `json:"wall_seconds,omitempty"`
}
