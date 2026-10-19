package app

import (
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/host"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/standin"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func benchGatewayPlaying(t *testing.T, steps []standin.Step) *standin.Node {
	t.Helper()
	node, err := standin.New(standin.Config{Session: "s-1", Board: "mega-01",
		Kinds: []map[string]any{thermalKindFile(t, "thermal_sensor.kind.json"), thermalKindFile(t, "fan.kind.json")},
		Devices: []standin.Device{{Name: "zone-01", Kind: "thermal_sensor", BootID: "boot-1", Initial: map[string]any{"temperature": twinOpenTemp}},
			{Name: "fan-01", Kind: "fan", BootID: "boot-1"}},
		Script: standin.Script{Steps: steps, Replies: []standin.Reply{{Command: "set_pwm_lease", Receipt: "accepted", Result: "succeeded"}}}})
	if err != nil {
		t.Fatal(err)
	}
	return node
}

func benchTemperatures(value func(at int64) float64) []standin.Step {
	var steps []standin.Step
	for at := twinStartNS + 5*twinTickNS; at <= twinStartNS+100*twinTickNS; at += 5 * twinTickNS {
		sample := standin.Sample{Metric: "temperature", Value: value(at), Quality: "good"}
		steps = append(steps, standin.Step{AtNS: at, Device: "zone-01", Samples: []standin.Sample{sample}})
	}
	return steps
}

// recordBenchSession records a hardware relay session: the gateway reports the zone every 5 s,
// and the reference host sends one fan lease at the first barrier, after the first 55 s of readings.
func recordBenchSession(t *testing.T, value func(at int64) float64) string {
	t.Helper()
	hostPath, out := testsupport.SocketPath(t), t.TempDir()
	lease := host.Scripted{ID: "cmd-1", AtNS: 1, Device: "fan-01", Command: "set_pwm_lease", Args: map[string]any{"duty_permille": 300, "lease_ms": 10000}}
	started := startHost(t, hostPath, relayHostOptions("hardware", lease))
	relay := startRelay(t, hostPath, out, nil)
	node := benchGatewayPlaying(t, benchTemperatures(value))
	gateway := dialGateway(t, relay.gateway, node)
	gateway.handshake()
	gateway.send(node.Play(twinStartNS + 55*twinTickNS)...)
	gateway.sync()
	gateway.send(node.Play(twinStartNS + 100*twinTickNS)...)
	gateway.sync()
	if ended := relay.stop(); ended.err != nil {
		t.Fatalf("relay ended with %v", ended.err)
	}
	started.stop()
	return out
}

func restingZone(int64) float64 { return twinOpenTemp }

func TestARelayedSessionThatFollowsTheTwinAgrees(t *testing.T) {
	t.Parallel()
	out := recordBenchSession(t, restingZone)
	report, err := TwinReport(t.Context(), twinOptions(t, out))
	if err != nil {
		t.Fatalf("TwinReport: %v", err)
	}
	if report.Verdict != domain.VerdictAgrees || report.Summary.Compared != 10 || report.Summary.Within != 10 || report.Summary.NotCompared != 10 {
		t.Fatalf("verdict %s summary %+v, want 10 readings compared and within, the 10 before the lease not compared", report.Verdict, report.Summary)
	}
	want := domain.TwinCommand{ID: "cmd-1", Device: "fan-01", Name: "set_pwm_lease", AtNS: twinStartNS + 55*twinTickNS, Status: domain.CommandApplied}
	if len(report.Commands) != 1 || report.Commands[0].ID != want.ID || report.Commands[0].Device != want.Device ||
		report.Commands[0].Name != want.Name || report.Commands[0].AtNS != want.AtNS || report.Commands[0].Status != want.Status {
		t.Fatalf("commands = %+v, want %+v: applied at the last reading before the gateway's receipt", report.Commands, want)
	}
	if report.Twin.StartNS != twinStartNS+5*twinTickNS || report.Run.InterruptLogDigest != "" {
		t.Fatalf("twin %+v run %+v, want the twin started at the first reading and no interrupt log", report.Twin, report.Run)
	}
}

func TestARelayedSessionWhoseBoardLeavesTheTwinDiverges(t *testing.T) {
	t.Parallel()
	hot := func(at int64) float64 {
		if at >= twinStartNS+80*twinTickNS {
			return 41
		}
		return twinOpenTemp
	}
	report, err := TwinReport(t.Context(), twinOptions(t, recordBenchSession(t, hot)))
	if err != nil {
		t.Fatalf("TwinReport: %v", err)
	}
	if report.Verdict != domain.VerdictDiverges || len(report.Divergences) != 1 {
		t.Fatalf("verdict %s divergences %+v, want one divergence", report.Verdict, report.Divergences)
	}
	stretch := report.Divergences[0]
	if stretch.Name != "zone-01.temperature#1" || stretch.FirstNS != twinStartNS+80*twinTickNS || stretch.LastNS != twinStartNS+100*twinTickNS ||
		stretch.Readings != 5 || stretch.BoardValue != 41 || stretch.WorstError < 6.5 || stretch.WorstError > 7.5 {
		t.Fatalf("stretch = %+v, want five readings from 80 s to 100 s at 41 degC, about 7 degC from the twin", stretch)
	}
	if len(stretch.Commands) != 1 || stretch.Commands[0] != "cmd-1" {
		t.Fatalf("commands in force = %v, want the lease cmd-1", stretch.Commands)
	}
}

func TestTheRelayJournalHoldsTheHostCommandTheReportReads(t *testing.T) {
	t.Parallel()
	out := recordBenchSession(t, restingZone)
	commands := 0
	for _, line := range wireLines(t, out) {
		if record := recordOf(line); line["direction"] == domain.TraceHost && record["type"] == "DCMD" {
			commands++
		}
	}
	if commands != 1 {
		t.Fatalf("journal holds %d host commands, want the one lease", commands)
	}
}

func TestAReportOfARealMixedRunPassesItsDigestChecksAndListsOnlyTheBoardsCommands(t *testing.T) {
	t.Parallel()
	hostPath, out := testsupport.SocketPath(t), t.TempDir()
	wireHost := serveWireHost(t, hostPath, true)
	options := mixedOptions(t, out)
	relay := startMixedRelay(t, hostPath, options)
	gateway := dialGateway(t, relay.gateway, mixedGateway(t, "gw-1"))
	gateway.handshake()
	wireHost.waitFor(func(r []map[string]any) bool {
		return hasRecord(r, "DBIRTH", "device", "fan-02") && hasRecord(r, "NSYNC", "", nil)
	})
	now := lastWorldTime(wireHost.snapshot())
	wireHost.write(commandRecord("sim-cmd", "fan-01", "mega-01-boot-1", now), commandRecord("hw-cmd", "fan-02", "boot-1", now))
	for _, command := range ofType(recordsOf(gateway.await(ofKind("DCMD"))), "DCMD") {
		gateway.send(gateway.node.Handle(command)...)
	}
	wireHost.release()
	wireHost.waitFor(func(r []map[string]any) bool {
		return outcomeOf(r, "sim-cmd") != "" && outcomeOf(r, "hw-cmd") != "" && hasRecord(r, "DDEATH", "device", "zone-01")
	})
	relay.stop()
	report, err := TwinReport(t.Context(), TwinReportOptions{Run: out, DomainsDir: options.Simulated.DomainsDir, Domain: "arduino-thermal",
		KindsDirs: []string{testsupport.ThermalBoardKindsDir()}})
	if err != nil {
		t.Fatalf("TwinReport of a real mixed run: %v", err)
	}
	art, err := run.LoadArtifact(filepath.Join(out, model.FileRunArtifact))
	if err != nil || report.Run.JournalDigest != art.JournalFileDigest || report.Run.FieldTraceDigest != art.FieldTraceDigest {
		t.Fatalf("report evidence %+v (%v), artifact %s and %s, want the digests run.json recorded", report.Run, err, art.JournalFileDigest, art.FieldTraceDigest)
	}
	if !report.Run.EvidenceBound {
		t.Fatalf("report evidence %+v, want it bound to the run.json whose digests it matched", report.Run)
	}
	if len(report.Commands) != 1 || report.Commands[0].ID != "hw-cmd" || report.Commands[0].Status != domain.CommandTwinRefused {
		t.Fatalf("commands = %+v, want only the hardware board's hw-cmd, which the twin has no entity for", report.Commands)
	}
}
