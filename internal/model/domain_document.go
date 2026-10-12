package model

type DomainDocument struct {
	ID          string        `json:"id"`
	Version     string        `json:"version"`
	Title       string        `json:"title"`
	Description string        `json:"description,omitempty"`
	Axes        Axes          `json:"axes"`
	Stresses    string        `json:"stresses"`
	Kinds       []Kind        `json:"kinds"`
	Instances   []InstanceSet `json:"instances"`
	Profiles    []Profile     `json:"profiles"`
	GroundTruth GroundTruth   `json:"ground_truth"`
}

type Kind struct {
	Name      string     `json:"name"`
	State     []State    `json:"state,omitempty"`
	Dynamics  []Dynamics `json:"dynamics,omitempty"`
	Channels  []Channel  `json:"channels"`
	Faults    []Fault    `json:"faults"`
	Effectors []Effector `json:"effectors,omitempty"`
}

type InstanceSet struct {
	Kind       string      `json:"kind"`
	IDTemplate string      `json:"id_template"`
	Count      EntityCount `json:"count"`
	Churn      *Churn      `json:"churn,omitempty"`
}
