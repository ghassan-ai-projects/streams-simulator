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
	SampleS   float64  `json:"sample_s,omitempty"`
}

type Follow struct {
	Link      string  `json:"link"`
	State     string  `json:"state"`
	Aggregate string  `json:"aggregate"`
	SampleS   float64 `json:"sample_s,omitempty"`
}

func (a InputLink) NeedsDefault() bool {
	return a.Aggregate == AggregateMean || a.Aggregate == AggregateMin || a.Aggregate == AggregateMax
}

type LinkRead struct {
	Type    string
	State   string
	SampleS float64
}

func (k *Kind) LinkReads() []string {
	var types []string
	for _, read := range k.LinkReadsDetailed() {
		types = appendNew(types, read.Type)
	}
	return types
}

func (k *Kind) LinkReadsDetailed() []LinkRead {
	var reads []LinkRead
	for _, d := range k.Dynamics {
		if d.Follow != nil {
			reads = append(reads, LinkRead{Type: d.Follow.Link, State: d.Follow.State, SampleS: d.Follow.SampleS})
		}
		for _, in := range d.F1.linkedInputs() {
			reads = append(reads, LinkRead{Type: in.Link.Type, State: in.State, SampleS: in.Link.SampleS})
		}
	}
	return reads
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
