package model

const (
	FormAlarm       = "alarm"
	AlarmHigh       = "high"
	AlarmNormal     = 0
	AlarmUnacked    = 1
	AlarmAcked      = 2
	AlarmReturned   = 3
	AlarmShelved    = 4
	ShelveThreshold = 0.5
)

type Alarm struct {
	Watch       string   `json:"watch"`
	Limit       *float64 `json:"limit,omitempty"`
	LimitState  string   `json:"limit_state,omitempty"`
	Direction   string   `json:"direction"`
	Deadband    float64  `json:"deadband,omitempty"`
	OnDelayS    float64  `json:"on_delay_s,omitempty"`
	OffDelayS   float64  `json:"off_delay_s,omitempty"`
	Priority    string   `json:"priority"`
	Acknowledge string   `json:"acknowledge,omitempty"`
	Shelve      string   `json:"shelve,omitempty"`
}
