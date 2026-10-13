package model

type KPITruth struct {
	FromNS      int64           `json:"from_ns"`
	ToNS        int64           `json:"to_ns"`
	Entities    []EntityKPIs    `json:"entities"`
	AlarmSystem AlarmSystemKPIs `json:"alarm_system"`
}

type EntityKPIs struct {
	EntityID string     `json:"entity_id"`
	KPIs     []KPIValue `json:"kpis,omitempty"`
	Alarms   []AlarmKPI `json:"alarms,omitempty"`
}

type KPIValue struct {
	State   string  `json:"state"`
	Mean    float64 `json:"mean"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Samples int     `json:"samples"`
}

type AlarmKPI struct {
	State          string  `json:"state"`
	Priority       string  `json:"priority"`
	Activations    int     `json:"activations"`
	Chattering     bool    `json:"chattering"`
	ActiveS        float64 `json:"active_s"`
	LongestActiveS float64 `json:"longest_active_s"`
	ActiveAtEnd    bool    `json:"active_at_end"`
}

type AlarmSystemKPIs struct {
	Activations    int `json:"activations"`
	PeakPer10Min   int `json:"peak_per_10_min"`
	FloodIntervals int `json:"flood_intervals"`
	Chattering     int `json:"chattering"`
	Stale          int `json:"stale"`
	StandingAtEnd  int `json:"standing_at_end"`
}
