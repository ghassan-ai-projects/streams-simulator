package domain

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"
)

func observeMoves(l *Live, moves Moves) {
	for _, line := range moves.Trace {
		if line.Direction == TraceNode || (line.Direction == TraceHost && line.Admitted) {
			l.Observe(line.Direction, line.Record, 10)
		}
	}
}

func evidenceSync(through uint64) Record {
	return Record{Type: "HSYNC", Lane: LaneControl, Fields: map[string]any{"through_seq": through, "command_evidence": []any{hostEvidence()}}}
}

func simulatedLive(t *testing.T) *Live {
	t.Helper()
	l := NewLive(false, nil, nil)
	birth := hardwareBirth(t, func(fields map[string]any) {
		for _, board := range fields["boards"].([]any) {
			board.(map[string]any)["environment"] = EnvironmentSimulated
		}
	})
	l.Observe(TraceNode, liveRecord(birth), 0)
	l.Observe(TraceNode, liveRecord(deviceBirthOf("P1", "boot-1")), 1)
	return l
}

func marksOf(l *Live) []LiveMark { return l.Fields()["marks"].([]LiveMark) }

func TestARefusalBeforeAnyBootReachesTheLiveCommandButAnAcceptanceDoesNot(t *testing.T) {
	t.Parallel()
	l := liveBoard(t)
	l.Observe(TraceHost, liveRecord(hostCommand()), 2)
	if l.Observe(TraceNode, liveAnswer("DRECEIPT", "c-1", "none", "accepted"), 3) || l.command("c-1").Receipt != "pending" {
		t.Fatal("an acceptance without a boot was taken")
	}
	refused := liveAnswer("DRECEIPT", "c-1", "none", "refused")
	refused["reason"] = "not_ready"
	if !l.Observe(TraceNode, refused, 4) || l.command("c-1").Receipt != "refused" || l.command("c-1").Reason != "not_ready" {
		t.Fatalf("relay refusal = %+v", l.command("c-1"))
	}
}

func TestTheRelaysRefusalOfACommandWhileTheGatewayIsOfflineReachesTheLiveCommand(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.relay.HostAttached()
	g.hostSends(hostState(7, 0))
	l := NewLive(true, nil, nil)
	observeMoves(l, g.hostSends(hostCommand()))
	if c := l.command("c-1"); c == nil || c.Receipt != "refused" || c.Reason != "not_ready" {
		t.Fatalf("command through an offline gateway = %+v, want refused not_ready", c)
	}
}

func TestAnAnswerNamingAnotherDeviceChangesNothing(t *testing.T) {
	t.Parallel()
	l := liveBoard(t)
	l.Observe(TraceHost, liveRecord(hostCommand()), 2)
	answer := liveAnswer("DRECEIPT", "c-1", "boot-1", "accepted")
	answer["device"] = "V1"
	if l.Observe(TraceNode, answer, 3) || l.command("c-1").Receipt != "pending" {
		t.Fatal("an answer for another device was taken")
	}
}

func TestEvidenceCountsOnlyFromAnHsyncTheRelayRoutedAndIsNotForwardedToTheGateway(t *testing.T) {
	t.Parallel()
	g := barrierRig(t)
	g.gw(barrierSync(4))
	l := liveBoard(t)
	acceptedCommand(t, l)
	moves := g.hostSends(evidenceSync(4))
	if _, forwarded := moves.Down[0]["command_evidence"]; forwarded {
		t.Fatalf("firmware received host evidence: %v", moves.Down[0])
	}
	observeMoves(l, moves)
	if l.command("c-1").Effect != "observed" {
		t.Fatalf("routed evidence = %+v", l.command("c-1"))
	}
}

func TestEvidenceOnADroppedHsyncNeverConfirmsAnEffect(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	g.relay.GatewayLost()
	g.relay.GatewayAttached()
	g.gateway = NewTransmitter("gw-2")
	g.gw(hardwareBirth(t, nil))
	l := liveBoard(t)
	acceptedCommand(t, l)
	moves := g.hostSends(evidenceSync(2))
	observeMoves(l, moves)
	if moves.Trace[0].Admitted || l.command("c-1").Effect != "unknown" {
		t.Fatalf("dropped HSYNC admitted %v, effect %s", moves.Trace[0].Admitted, l.command("c-1").Effect)
	}
}

func TestASimulatedDeviceIsNeverStaleOrOnALinkAndItsCommandSaysSo(t *testing.T) {
	t.Parallel()
	l := simulatedLive(t)
	d := l.devices["P1"]
	l.Tick(10 * 60 * 1000)
	l.Connection("host", "online", 2)
	l.Connection("host", "offline", 3)
	l.End(4)
	l.Observe(TraceHost, liveRecord(hostCommand()), 5)
	if d.Health != "simulated" || d.Readings[0].State != "simulated" || l.command("c-1").Environment != "simulated" {
		t.Fatalf("simulated device = %+v, command %+v", d, l.command("c-1"))
	}
}

func TestOnlyBoardsAndUnknownDevicesAgeWithWallTicksAndLoseTheirGatewayLink(t *testing.T) {
	t.Parallel()
	l := NewLive(true, nil, nil)
	l.Observe(TraceNode, liveRecord(deviceBirthOf("P1", "boot-1")), 1)
	d := l.devices["P1"]
	l.Connection("gateway", "online", 2)
	l.Connection("gateway", "offline", 3)
	if d.Environment != "unknown" || d.Health != "offline" {
		t.Fatalf("device on an unannounced board = %+v, want disconnected with the gateway", d)
	}
}

func TestReadingsAreBoundedAndOnlyAnnouncedMetricsGrowTheState(t *testing.T) {
	t.Parallel()
	l := liveBoard(t)
	d := l.devices["P1"]
	l.Observe(TraceNode, map[string]any{"type": "DDATA", "device": "P1", "boot_id": "boot-1", "metrics": []any{
		map[string]any{"name": "bogus", "alias": 9, "value": 1.0, "quality": "good"}}}, 5)
	if len(d.Readings) != 1 || len(d.aliases) != 1 || !slices.ContainsFunc(marksOf(l), func(m LiveMark) bool { return strings.Contains(m.Label, "bogus") }) {
		t.Fatalf("unannounced metric grew the state or left no mark: %+v %v", d.Readings, d.aliases)
	}
	metrics := []any{}
	for i := range 600 {
		name := fmt.Sprintf("m-%d", i)
		l.kinds["pump"][name] = map[string]any{"name": name, "data_type": "number"}
		metrics = append(metrics, map[string]any{"name": name, "alias": i + 20, "value": 1.0, "quality": "good"})
	}
	l.Observe(TraceNode, map[string]any{"type": "DDATA", "device": "P1", "boot_id": "boot-1", "metrics": metrics}, 6)
	if len(d.Readings) != maxLiveReadings {
		t.Fatalf("readings = %d, want the schema bound %d", len(d.Readings), maxLiveReadings)
	}
}

func TestEvidenceAfterTheBoardRebootedLeavesAMarkInsteadOfVanishing(t *testing.T) {
	t.Parallel()
	l := liveBoard(t)
	acceptedCommand(t, l)
	l.Observe(TraceNode, liveRecord(deviceBirthOf("P1", "boot-2")), 6)
	l.Fields()
	if !l.Observe(TraceHost, map[string]any{"type": "HSYNC", "command_evidence": []any{hostEvidence()}}, 7) {
		t.Fatal("dropped evidence left no trace")
	}
	marks := marksOf(l)
	if len(marks) != 1 || !strings.Contains(marks[0].Label, "evidence not applied") || l.command("c-1").Effect != "unknown" {
		t.Fatalf("marks = %+v", marks)
	}
}

func TestForgedEvidenceNeverBecomesAnObservedEffect(t *testing.T) {
	t.Parallel()
	edits := map[string]func(map[string]any){
		"huge tolerance":     func(e map[string]any) { e["tolerance"] = 1e9 },
		"infinite tolerance": func(e map[string]any) { e["tolerance"] = math.Inf(1) },
		"unknown metric":     func(e map[string]any) { e["metric"] = "ghost" },
	}
	for name, edit := range edits {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			l := liveBoard(t)
			acceptedCommand(t, l)
			e := hostEvidence()
			edit(e)
			if l.Observe(TraceHost, map[string]any{"type": "HSYNC", "command_evidence": []any{e}}, 5) || l.command("c-1").Effect != "unknown" {
				t.Fatalf("%s confirmed an effect", name)
			}
		})
	}
}

func TestEvidenceDisagreeingWithTheReadingTheRelayHoldsIsUnverifiedNotObserved(t *testing.T) {
	t.Parallel()
	l := liveBoard(t)
	acceptedCommand(t, l)
	e := hostEvidence()
	e["expected"], e["observed_value"], e["tolerance"] = 80.0, 80.0, 0.5
	l.Observe(TraceHost, map[string]any{"type": "HSYNC", "command_evidence": []any{e}}, 5)
	if c := l.command("c-1"); c.Effect != "unknown" || c.EffectEvidence != "Host checked flow: unverified" {
		t.Fatalf("evidence the held reading contradicts = %+v", c)
	}
}

func TestOnlyTheFirstHundredAndTwentyEightEvidenceEntriesAreRead(t *testing.T) {
	t.Parallel()
	l := liveBoard(t)
	acceptedCommand(t, l)
	entries := make([]any, 0, 200)
	for range 199 {
		entries = append(entries, map[string]any{"command_id": "nope"})
	}
	entries = append(entries, hostEvidence())
	if l.Observe(TraceHost, map[string]any{"type": "HSYNC", "command_evidence": entries}, 5) || l.command("c-1").Effect != "unknown" {
		t.Fatal("an entry beyond the bound was read")
	}
}

func TestACommandIdResentAfterTheWindowIsNotListedAgain(t *testing.T) {
	t.Parallel()
	l := liveBoard(t)
	for i := range 130 {
		command := liveRecord(hostCommand())
		command["command_id"] = fmt.Sprintf("c-%d", i)
		l.Observe(TraceHost, command, int64(i))
	}
	again := liveRecord(hostCommand())
	again["command_id"] = "c-0"
	if l.Observe(TraceHost, again, 200) || l.command("c-0") != nil {
		t.Fatal("an old command id came back as a new command")
	}
}

func TestFieldsReturnCopiesAndBoundTheMarks(t *testing.T) {
	t.Parallel()
	l := liveBoard(t)
	for i := range maxLiveMarks + 20 {
		l.mark(int64(i), "connection", "host", "Host link online", "host")
	}
	fields := l.Fields()
	fields["devices"].([]*LiveDevice)[0].Health = "changed"
	fields["devices"].([]*LiveDevice)[0].Readings[0].State = "changed"
	if len(fields["marks"].([]LiveMark)) != maxLiveMarks || l.devices["P1"].Health == "changed" || l.devices["P1"].Readings[0].State == "changed" {
		t.Fatal("Fields exposed live state or an unbounded mark list")
	}
}

func TestHostLinkMarksNeedAnEarlierOnlineAndARefusedStateIsNeverOnline(t *testing.T) {
	t.Parallel()
	l := NewLive(false, nil, nil)
	l.Observe(TraceHost, map[string]any{"type": "STATE", "online": false}, 1)
	l.Observe(TraceHost, map[string]any{"type": "STATE", "online": true, "protocol_version": 1.0, "protocol_min": 7.0, "protocol_max": 9.0}, 2)
	if l.host != "unknown" || len(l.marks) != 0 {
		t.Fatalf("host %s, marks %+v, want no link marks before the host was ever online", l.host, l.marks)
	}
	state := liveRecord(hostState(7, 0))
	state["protocol_version"] = 1
	l.Observe(TraceHost, state, 3)
	if l.host != "online" {
		t.Fatalf("host = %s", l.host)
	}
}

func TestMarksCarryTheEventWordsAndASubject(t *testing.T) {
	t.Parallel()
	l := liveBoard(t)
	l.Fields()
	acceptedCommand(t, l)
	l.Observe(TraceNode, liveRecord(deviceBirthOf("P1", "boot-2")), 6)
	l.Connection("gateway", "offline", 7)
	want := []LiveMark{{2, "command", "command", "P1 command start sent", "P1"}, {3, "command", "command", "P1 command start accepted", "P1"},
		{4, "command", "command", "P1 reported command start succeeded", "P1"}, {6, "connection", "device", "P1 rebooted (boot boot-2)", "P1"},
		{6, "connection", "device", "P1 online", "P1"}, {7, "connection", "gateway", "Gateway link offline", "gateway"}}
	got := marksOf(l)
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("marks = %+v, want %+v", got, want)
	}
}

func TestASnapshotWithNoMarksWritesAnEmptyListNotNull(t *testing.T) {
	t.Parallel()
	l := NewLive(true, nil, nil)
	l.Fields()
	encoded, err := json.Marshal(l.Fields()["marks"])
	if err != nil || string(encoded) != "[]" {
		t.Fatalf("marks = %s, %v, want [] (the journal schema requires an array)", encoded, err)
	}
}
