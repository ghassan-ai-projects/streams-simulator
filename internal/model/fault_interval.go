package model

const (
	ClearedByCommand   = "command"
	ClearedByPulseEnd  = "pulse_end"
	ClearedByResponder = "responder"
)

type FaultInterval struct {
	FaultID      string  `json:"fault_id"`
	EntityID     string  `json:"entity_id"`
	Fault        string  `json:"fault"`
	InjectedAtNS int64   `json:"injected_at_ns"`
	OnsetNS      int64   `json:"onset_ns"`
	ClearedNS    *int64  `json:"cleared_ns"`
	ClearedBy    *string `json:"cleared_by"`
}
