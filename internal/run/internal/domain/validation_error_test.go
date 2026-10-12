package domain

import (
	"errors"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestArtifactRejectionsAreValidationErrors(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, source string
		reject       func() error
	}{
		{"not JSON", "run.json", func() error { return ValidateArtifactDocument([]byte("{"), "run.json") }},
		{"schema violation", "run.json", func() error { return ValidateArtifactDocument([]byte("[]"), "run.json") }},
		{"trailing document", "", func() error { _, err := DecodeArtifact([]byte("{} {}")); return err }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var invalid *model.ValidationError
			if err := tc.reject(); !errors.As(err, &invalid) {
				t.Fatalf("err = %v, want a *model.ValidationError", err)
			}
			if invalid.Contract != model.ContractRunArtifact || invalid.Source != tc.source {
				t.Fatalf("%+v, want a run-artifact error with source %q", invalid, tc.source)
			}
		})
	}
}
