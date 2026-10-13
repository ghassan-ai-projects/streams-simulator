package model

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

type ConsumerInfo struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	ConfigDigest string `json:"config_digest,omitempty"`
	Arm          string `json:"arm,omitempty"`
}

type Detection struct {
	EntityID     string   `json:"entity_id"`
	DetectedAt   string   `json:"detected_at"`
	Label        string   `json:"label,omitempty"`
	Confidence   float64  `json:"confidence,omitempty"`
	EvidenceRefs []string `json:"evidence_refs,omitempty"`
	Narrative    string   `json:"narrative,omitempty"`
}

const (
	AdmissionAccepted      = "accepted"
	AdmissionDuplicate     = "duplicate"
	AdmissionConflict      = "conflict"
	AdmissionLate          = "late"
	AdmissionOutOfContract = "out_of_contract"
	AdmissionMalformed     = "malformed"
	AdmissionRejected      = "rejected"
)

type Admission struct {
	Seq     int64  `json:"seq"`
	Outcome string `json:"outcome"`
	Note    string `json:"note,omitempty"`
}

type Action struct {
	CommandID       string `json:"command_id"`
	Effector        string `json:"effector"`
	EntityID        string `json:"entity_id,omitempty"`
	IssuedAt        string `json:"issued_at"`
	OutcomeBelieved string `json:"outcome_believed,omitempty"`
}

const (
	BelievedSucceeded = "succeeded"
	BelievedFailed    = "failed"
	BelievedUnknown   = "unknown"
)

type Resource struct {
	ReasoningInvocations int64   `json:"reasoning_invocations,omitempty"`
	InputTokens          int64   `json:"input_tokens,omitempty"`
	OutputTokens         int64   `json:"output_tokens,omitempty"`
	WallSeconds          float64 `json:"wall_seconds,omitempty"`
}
