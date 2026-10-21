package app

import (
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/host"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestTheReferenceHostRenewsALeaseWithRepeatedIdenticalCommandsThatEachExecute(t *testing.T) {
	t.Parallel()
	r, _ := mappedThermalRun(t)
	hostPath := testsupport.SocketPath(t)
	var script []host.Scripted
	for i, id := range []string{"r0", "r1", "r2", "r3"} {
		script = append(script, host.Scripted{ID: id, AtNS: model.DefaultStartTimeNS + int64(2*i+2)*1e9, Device: "fan-01",
			Command: "set_cooling", Args: map[string]any{"level_percent": 30}})
	}
	started := startHost(t, hostPath, host.Options{Environment: "simulated", ProtocolMax: 1, Epoch: 1, Policy: policy, Script: script})
	options := Options{Environment: "simulated", ProtocolMax: 1, SessionID: "s-1", SliceNS: 1e9, UntilNS: model.DefaultStartTimeNS + 12e9,
		Policies: []string{policy}, KindsDirs: []string{semanticKindsDir()}}
	if err := Serve(t.Context(), r, hostPath, 5*time.Second, options); err != nil {
		t.Fatal(err)
	}
	commandsSucceeded(t, started.stop(), "r0", "r1", "r2", "r3")
	executed := 0
	for _, call := range r.World.EffectorCalls() {
		if call.Effector == "set_pwm_lease" {
			executed++
		}
	}
	if executed != 4 {
		t.Fatalf("the twin ran set_pwm_lease %d times, want each of the four renewals executed, none replayed", executed)
	}
}
