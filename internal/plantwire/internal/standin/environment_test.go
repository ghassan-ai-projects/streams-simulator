package standin

import (
	"strings"
	"testing"
)

func environmentNode(t *testing.T, environment string) *Node {
	t.Helper()
	config := boardConfig(t)
	config.Environment = environment
	node, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	return node
}

func TestAStandInAnnouncesAHardwareBoardUnlessToldToAnnounceASimulatedOne(t *testing.T) {
	t.Parallel()
	cases := []struct{ name, environment, want string }{
		{"by default", "", "hardware"},
		{"when told hardware", "hardware", "hardware"},
		{"when told to lie about being simulated", "simulated", "simulated"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			birth, err := environmentNode(t, tc.environment).Birth()
			if err != nil {
				t.Fatal(err)
			}
			boards, _ := birth["boards"].([]any)
			if err := wireErr(birth); err != nil || len(boards) != 1 || birth["environment"] != tc.want || boards[0].(map[string]any)["environment"] != tc.want {
				t.Fatalf("NBIRTH environment %v boards %v (%v), want the node and its one board %s", birth["environment"], boards, err, tc.want)
			}
		})
	}
}

func TestAHostRefusesAStandInThatAnnouncesTheWrongEnvironmentAndAMixedHostAcceptsEither(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, standIn, host, refusal string
	}{
		{"a simulated board before a hardware host", "simulated", "hardware", "the host accepts hardware nodes only, the node announces simulated"},
		{"a hardware board before a simulated host", "hardware", "simulated", "the host accepts simulated nodes only, the node announces hardware"},
		{"a simulated board before a mixed host", "simulated", "mixed", ""},
		{"a hardware board before a mixed host", "hardware", "mixed", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			b := newBench(t, environmentNode(t, tc.standIn), hostOptions(tc.host))
			birth, err := b.node.Birth()
			if err != nil {
				t.Fatal(err)
			}
			replies := b.deliver(birth)
			report := b.host.Report()
			accepted := tc.refusal == ""
			if len(replies) != 1 || replies[0]["online"] != accepted || b.host.Accepted() != accepted || !strings.Contains(report.Reason, tc.refusal) || (!accepted && report.Refused != "environment_mismatch") {
				t.Fatalf("host answered %v with report %+v, want accepted=%v and the reason %q", replies, report, accepted, tc.refusal)
			}
		})
	}
}
