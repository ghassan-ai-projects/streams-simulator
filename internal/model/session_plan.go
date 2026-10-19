package model

import "github.com/ghassan-ai-projects/streams-simulator/internal/schemas"

const SessionPlanSchemaVersion = "0.1"

type SessionPlan struct {
	SchemaVersion string        `json:"schema_version"`
	Domain        string        `json:"domain"`
	Seed          *uint64       `json:"seed,omitempty"`
	HostSocket    string        `json:"host_socket"`
	GatewaySocket string        `json:"gateway_socket,omitempty"`
	Real          []RealBinding `json:"real,omitempty"`
	Policies      []string      `json:"policies,omitempty"`
}

type RealBinding struct {
	Instance string `json:"instance"`
	Board    string `json:"board"`
	Device   string `json:"device"`
}

func (p *SessionPlan) WorldSeed() uint64 {
	if p.Seed == nil {
		return 1
	}
	return *p.Seed
}

func DecodeSessionPlan(raw []byte) (*SessionPlan, error) {
	if err := validateAgainst(raw, "session plan", ContractSessionPlan, schemas.SessionPlan()); err != nil {
		return nil, err
	}
	var plan SessionPlan
	if err := DecodeBytes(raw, &plan); err != nil {
		return nil, NotJSONError(ContractSessionPlan, "", "model: session plan cannot be decoded", err)
	}
	return &plan, nil
}
