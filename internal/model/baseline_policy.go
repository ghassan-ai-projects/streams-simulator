package model

const PolicyTargetMatched = "matched"

type BaselinePolicy struct {
	Name string          `json:"name"`
	When PolicyCondition `json:"when"`
	ForS float64         `json:"for_s,omitempty"`
	Then []PolicyCall    `json:"then"`
}

type PolicyCondition struct {
	Kind      string  `json:"kind"`
	State     string  `json:"state"`
	Operator  string  `json:"operator"`
	Threshold float64 `json:"threshold"`
}

type PolicyCall struct {
	Effector string         `json:"effector"`
	Target   string         `json:"target"`
	Args     map[string]any `json:"args,omitempty"`
	DelayS   float64        `json:"delay_s,omitempty"`
}

func (c PolicyCall) TargetsMatched() bool { return c.Target == PolicyTargetMatched }
