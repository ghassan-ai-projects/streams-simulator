package domain

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestEveryDomainRejectionIsAValidationErrorNamingItsSource(t *testing.T) {
	t.Parallel()
	example, err := Load(testsupport.Example())
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, raw string
		located   bool
	}{
		{"not JSON", "{", false},
		{"schema violation", `{"id":"x"}`, true},
		{"cross-reference", string(breakChannelReference(t, example)), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := Parse([]byte(tc.raw), "src.json")
			var invalid *model.ValidationError
			if !errors.As(err, &invalid) {
				t.Fatalf("err = %v, want a *model.ValidationError", err)
			}
			if invalid.Contract != model.ContractDomainSpec || invalid.Source != "src.json" || len(invalid.Issues) == 0 ||
				invalid.Issues[0].Located != tc.located || invalid.Error() != invalid.Summary {
				t.Fatalf("%+v, want a domain-spec error from src.json with located=%v issues", invalid, tc.located)
			}
		})
	}
}

func TestDescribeOfAnUnknownIDIsErrUnknownDomain(t *testing.T) {
	t.Parallel()
	_, err := NewCatalog(nil).Describe("no-such")
	if !errors.Is(err, ErrUnknownDomain) || err.Error() != `catalog: unknown domain "no-such"` {
		t.Fatalf("Describe = %v, want ErrUnknownDomain with the id", err)
	}
}

// breakChannelReference returns the compiled spec as JSON with its first
// channel observing a state that does not exist.
func breakChannelReference(t *testing.T, c *Compiled) []byte {
	t.Helper()
	raw, err := json.Marshal(c.Spec)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	channels, _ := doc["channels"].([]any)
	first, _ := channels[0].(map[string]any)
	first["observes"] = "no_such_state"
	broken, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return broken
}
