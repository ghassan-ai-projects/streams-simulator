package model

import (
	"encoding/json"
	"fmt"
)

type f1InputWire struct {
	State string     `json:"state"`
	Coef  *float64   `json:"coef,omitempty"`
	Link  *InputLink `json:"link,omitempty"`
}

func (f F1Input) MarshalJSON() ([]byte, error) {
	raw, err := json.Marshal(f.wire())
	if err != nil {
		return nil, fmt.Errorf("model: marshal f1 input: %w", err)
	}
	return raw, nil
}

func (f F1Input) wire() f1InputWire {
	out := f1InputWire{State: f.State, Link: f.Link}
	if f.CoefSet || f.Coef != 0 {
		out.Coef = &f.Coef
	}
	return out
}

func (f *F1Input) UnmarshalJSON(b []byte) error {
	var state string
	if err := json.Unmarshal(b, &state); err == nil {
		f.State, f.Coef, f.CoefSet = state, 1, true
		return nil
	}
	return f.unmarshalObject(b)
}

func (f *F1Input) unmarshalObject(b []byte) error {
	type alias F1Input
	var value alias
	if err := json.Unmarshal(b, &value); err != nil {
		return fmt.Errorf("UnmarshalJSON: %w", err)
	}
	f.State, f.Coef, f.Link = value.State, value.Coef, value.Link
	return f.resolveCoefficient(b)
}

func (f *F1Input) resolveCoefficient(b []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(b, &fields); err != nil {
		return fmt.Errorf("UnmarshalJSON: %w", err)
	}
	_, f.CoefSet = fields["coef"]
	if !f.CoefSet {
		f.Coef = 1
	}
	return nil
}
