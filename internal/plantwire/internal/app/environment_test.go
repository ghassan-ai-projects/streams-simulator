package app

import (
	"errors"
	"strings"
	"testing"

	wire "github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/host"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestASimulatedNodeAnnouncesItselfAndEveryBoardAsSimulated(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		run  func(*testing.T) *run.Run
	}{
		{"a domain without declared boards", utilityRun},
		{"a domain with declared boards", func(t *testing.T) *run.Run { return runOf(t, testsupport.PumpBoardDomain()) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			peer := &scriptedHost{t: t}
			if err := NewSession(tc.run(t), peer, sliceOptions(1)).Run(t.Context()); err != nil {
				t.Fatal(err)
			}
			birth := birthsOf(peer)[0]
			boards, _ := birth["boards"].([]any)
			if birth["environment"] != "simulated" || len(boards) == 0 {
				t.Fatalf("NBIRTH environment %v with %d boards, want simulated and at least one board", birth["environment"], len(boards))
			}
			for _, board := range boards {
				if got := board.(map[string]any)["environment"]; got != "simulated" {
					t.Errorf("board %v announces environment %v, want simulated", board.(map[string]any)["board"], got)
				}
			}
		})
	}
}

func TestASimulatedNodeCannotBeMadeToAnnounceHardwareOrMixed(t *testing.T) {
	t.Parallel()
	for _, environment := range []string{"hardware", "mixed", "lab", ""} {
		t.Run("environment "+environment, func(t *testing.T) {
			t.Parallel()
			peer, options := &scriptedHost{t: t}, sliceOptions(1)
			options.Environment = environment
			err := NewSession(utilityRun(t), peer, options).Run(t.Context())
			if !errors.Is(err, ErrBadOptions) || !strings.Contains(err.Error(), `a simulated world announces only "simulated"`) || len(peer.sent) != 0 {
				t.Fatalf("Run = %v after %v, want ErrBadOptions saying a simulated world announces only simulated, and nothing sent", err, peer.types())
			}
		})
	}
}

func TestTheReferenceHostIsSetToSimulatedHardwareOrMixedAndNothingElse(t *testing.T) {
	t.Parallel()
	for _, environment := range []string{"lab", "", "Simulated"} {
		report, err := ServeHost(t.Context(), testsupport.SocketPath(t), 0, host.Options{Environment: environment})
		if !errors.Is(err, wire.ErrEnvironment) || !strings.Contains(err.Error(), "none of simulated, hardware, mixed") || report.Refused != "" {
			t.Errorf("ServeHost with environment %q = %+v, %v, want wire.ErrEnvironment naming the three environments", environment, report, err)
		}
	}
}
