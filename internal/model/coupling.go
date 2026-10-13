package model

import "slices"

const (
	AggregateSum          = "sum"
	AggregateMean         = "mean"
	AggregateMin          = "min"
	AggregateMax          = "max"
	AggregateCountRunning = "count_running"
)

type InputLink struct {
	Type      string   `json:"type"`
	Aggregate string   `json:"aggregate"`
	Default   *float64 `json:"default,omitempty"`
}

type Follow struct {
	Link      string `json:"link"`
	State     string `json:"state"`
	Aggregate string `json:"aggregate"`
}

func (a InputLink) NeedsDefault() bool {
	return a.Aggregate == AggregateMean || a.Aggregate == AggregateMin || a.Aggregate == AggregateMax
}

func (k *Kind) LinkReads() []string {
	var types []string
	for _, d := range k.Dynamics {
		if d.Follow != nil {
			types = appendNew(types, d.Follow.Link)
		}
		for _, in := range d.F1.linkedInputs() {
			types = appendNew(types, in.Link.Type)
		}
	}
	return types
}

func (f *F1Dyn) linkedInputs() []F1Input {
	if f == nil {
		return nil
	}
	var linked []F1Input
	for _, in := range f.Inputs {
		if in.Link != nil {
			linked = append(linked, in)
		}
	}
	return linked
}

func appendNew(types []string, name string) []string {
	if slices.Contains(types, name) {
		return types
	}
	return append(types, name)
}
