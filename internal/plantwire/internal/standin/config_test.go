package standin

import (
	"errors"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

func TestAConfigurationThatCouldNotSpeakTheWireIsRefusedWithTheReason(t *testing.T) {
	t.Parallel()
	sample := func(metric string) []Sample { return []Sample{{Metric: metric, Value: 1}} }
	cases := []struct {
		name   string
		change func(*Config)
		cause  error
		reason string
	}{
		{"a session that is no wire id", func(c *Config) { c.Session = "" }, ErrConfig, `"" is no wire id`},
		{"a board that is no wire id", func(c *Config) { c.Board = "mega 01" }, ErrConfig, `"mega 01" is no wire id`},
		{"an environment that is neither", func(c *Config) { c.Environment = "lab" }, ErrConfig, `environment "lab"`},
		{"a kind file that carries its digest", func(c *Config) { c.Kinds[0]["kind_digest"] = "sha256:" + strings.Repeat("0", 64) }, domain.ErrKindDigestInFile, "kind file 0"},
		{"a kind given twice", func(c *Config) { c.Kinds = append(c.Kinds, c.Kinds[1]) }, ErrConfig, `kind "fan" is given twice`},
		{"a device of an undefined kind", func(c *Config) { c.Devices[0].Kind = "pump" }, ErrConfig, `kind "pump", which no kind file defines`},
		{"a device given twice", func(c *Config) { c.Devices[1].Name = "zone-01" }, ErrConfig, `device "zone-01" is given twice`},
		{"a device without a boot", func(c *Config) { c.Devices[1].BootID = "" }, ErrConfig, `needs wire ids`},
		{"a step for a device off the board", func(c *Config) { c.Script.Steps = []Step{{Device: "pump-01", Samples: sample("x")}} }, ErrConfig, `device "pump-01" is not on the board`},
		{"a step that does nothing", func(c *Config) { c.Script.Steps = []Step{{Device: "fan-01"}} }, ErrConfig, "not both and not neither"},
		{"a step that does two things", func(c *Config) {
			c.Script.Steps = []Step{{Device: "fan-01", Link: LinkDown, Samples: sample("level_percent")}}
		}, ErrConfig, "not both and not neither"},
		{"a link that is neither down nor up", func(c *Config) { c.Script.Steps = []Step{{Device: "fan-01", Link: "sideways"}} }, ErrConfig, `link "sideways"`},
		{"a boot that is no wire id", func(c *Config) { c.Script.Steps = []Step{{Device: "fan-01", Link: LinkUp, BootID: "boot 2"}} }, ErrConfig, `boot "boot 2"`},
		{"a reading the kind does not have", func(c *Config) { c.Script.Steps = []Step{{Device: "fan-01", Samples: sample("rpm")}} }, ErrConfig, `has no metric "rpm"`},
		{"a reply that names no command", func(c *Config) { c.Script.Replies = []Reply{{Receipt: "accepted"}} }, ErrConfig, "reply 0"},
		{"a reply with an unknown receipt", func(c *Config) { c.Script.Replies = []Reply{{Command: "stop_fan", Receipt: "done"}} }, ErrConfig, "reply 0"},
		{"a reply with an unknown result", func(c *Config) {
			c.Script.Replies = []Reply{{Command: "stop_fan", Receipt: "accepted", Result: "great"}}
		}, ErrConfig, "reply 0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			config := boardConfig(t)
			tc.change(&config)
			_, err := New(config)
			if !errors.Is(err, tc.cause) || !strings.Contains(err.Error(), tc.reason) {
				t.Fatalf("New = %v, want %v mentioning %q", err, tc.cause, tc.reason)
			}
		})
	}
}

func TestAKindFileOfAnotherRecordShapeIsRefusedByTheKindSchema(t *testing.T) {
	t.Parallel()
	config := boardConfig(t)
	config.Kinds = []map[string]any{{"kind": "fan"}}
	if _, err := New(config); !errors.Is(err, ErrConfig) || !strings.Contains(err.Error(), "kind file 0") {
		t.Fatalf("New = %v, want ErrConfig naming the kind file that fails the kind schema", err)
	}
}
