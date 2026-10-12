package domain

import (
	"errors"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestEveryAdapterRejectionIsAValidationErrorNamingItsSource(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		reject  func() error
		located bool
	}{
		{"not JSON", func() error { _, err := LoadBytes([]byte("{"), "src.json"); return err }, false},
		{"schema violation", func() error { _, err := DecodeFile([]byte(`{"id":"x"}`), "src.json"); return err }, true},
		{"cross-reference", func() error {
			return crossCheck(&model.Adapter{Record: model.RecordTemplate{Fields: []model.Field{{From: model.ValueExpr{Op: "const"}}}}}, "src.json")
		}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var invalid *model.ValidationError
			if err := tc.reject(); !errors.As(err, &invalid) {
				t.Fatalf("err = %v, want a *model.ValidationError", err)
			}
			if invalid.Contract != model.ContractOutputAdapter || invalid.Source != "src.json" || invalid.Issues[0].Located != tc.located {
				t.Fatalf("%+v, want an output-adapter error from src.json with located=%v issues", invalid, tc.located)
			}
		})
	}
}
