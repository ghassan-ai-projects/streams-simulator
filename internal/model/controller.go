package model

const (
	FormPID       = "pid"
	ActionReverse = "reverse"
	ActionDirect  = "direct"
)

type Controller struct {
	PV           string  `json:"pv"`
	SP           string  `json:"sp"`
	Kp           float64 `json:"kp,omitempty"`
	Ki           float64 `json:"ki,omitempty"`
	Kd           float64 `json:"kd,omitempty"`
	Action       string  `json:"action"`
	Manual       string  `json:"manual,omitempty"`
	ManualOutput string  `json:"manual_output,omitempty"`
}

func (c *Controller) Sign() float64 {
	if c.Action == ActionDirect {
		return -1
	}
	return 1
}
