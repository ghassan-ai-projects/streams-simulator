package model

import "github.com/ghassan-ai-projects/streams-simulator/internal/schemas"

const (
	SessionPlanSchemaVersion = "0.1"
	PaceWall                 = "wall"
)

type SessionPlan struct {
	SchemaVersion string        `json:"schema_version"`
	Domain        string        `json:"domain"`
	Seed          *uint64       `json:"seed,omitempty"`
	HostSocket    string        `json:"host_socket"`
	GatewaySocket string        `json:"gateway_socket,omitempty"`
	Real          []RealBinding `json:"real,omitempty"`
	Policies      []string      `json:"policies,omitempty"`
	Disturbances  []Disturbance `json:"disturbances,omitempty"`
	StartS        *int64        `json:"start_s,omitempty"`
	Pace          string        `json:"pace,omitempty"`
}

type Disturbance struct {
	AtS       float64  `json:"at_s"`
	Instance  string   `json:"instance"`
	Fault     string   `json:"fault"`
	DurationS *float64 `json:"duration_s,omitempty"`
	Severity  *float64 `json:"severity,omitempty"`
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

func (p *SessionPlan) WorldStartNS() int64 {
	if p.StartS == nil {
		return DefaultStartTimeNS
	}
	return DefaultStartTimeNS + *p.StartS*1e9
}

func (p *SessionPlan) PacedByWall() bool {
	return p.Pace == PaceWall
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
