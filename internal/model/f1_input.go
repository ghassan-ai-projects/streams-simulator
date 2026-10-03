package model

import (
	"encoding/json"
	"fmt"
)

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
