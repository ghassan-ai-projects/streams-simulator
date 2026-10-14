package model

type Override struct {
	State     []State    `json:"state,omitempty"`
	Dynamics  []Dynamics `json:"dynamics,omitempty"`
	Channels  []Channel  `json:"channels,omitempty"`
	Faults    []Fault    `json:"faults,omitempty"`
	Effectors []Effector `json:"effectors,omitempty"`
	Narrow    *Narrowing `json:"narrow,omitempty"`
}

type Narrowing struct {
	State     []StateNarrowing    `json:"state,omitempty"`
	Channels  []ChannelNarrowing  `json:"channels,omitempty"`
	Effectors []EffectorNarrowing `json:"effectors,omitempty"`
}

type StateNarrowing struct {
	Name string  `json:"name"`
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
}

type ChannelNarrowing struct {
	Name  string `json:"name"`
	Range Range  `json:"range"`
}

type EffectorNarrowing struct {
	Name string               `json:"name"`
	Args map[string]ArgBounds `json:"args"`
}

type ArgBounds struct {
	Minimum *float64 `json:"minimum,omitempty"`
	Maximum *float64 `json:"maximum,omitempty"`
}

type LinkType struct {
	Name        string `json:"name"`
	From        string `json:"from"`
	To          string `json:"to"`
	Description string `json:"description,omitempty"`
}

type Link struct {
	Type     string `json:"type"`
	From     string `json:"from"`
	To       string `json:"to"`
	FromPort string `json:"from_port,omitempty"`
	ToPort   string `json:"to_port,omitempty"`
}

type Board struct {
	Name        string   `json:"name"`
	Members     []string `json:"members"`
	Description string   `json:"description,omitempty"`
}
