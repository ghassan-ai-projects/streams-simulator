package model

import (
	"encoding/json"
	"fmt"
)

// MarshalJSON writes the object form. The coefficient is left out only when
// it was never given and is zero; an explicit zero must be written, or the
// value would reload as the default 1.
func (f F1Input) MarshalJSON() ([]byte, error) {
	type wire struct {
		State string   `json:"state"`
		Coef  *float64 `json:"coef,omitempty"`
	}
	out := wire{State: f.State}
	if f.CoefSet || f.Coef != 0 {
		out.Coef = &f.Coef
	}
	raw, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("model: marshal f1 input: %w", err)
	}
	return raw, nil
}

// UnmarshalJSON accepts both the bare-string and object forms.
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
	f.State, f.Coef = value.State, value.Coef
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
