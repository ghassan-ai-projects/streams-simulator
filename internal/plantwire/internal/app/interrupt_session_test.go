package app

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

type memoryLog struct{ records []map[string]any }

func (m *memoryLog) Add(record map[string]any) error {
	m.records = append(m.records, record)
	return nil
}

func (m *memoryLog) Flush() error { return nil }

func (m *memoryLog) events() []string {
	events := make([]string, len(m.records))
	for i, record := range m.records {
		events[i], _ = record["event"].(string)
	}
	return events
}

type controlled struct {
	t       *testing.T
	host    *scriptedHost
	run     *run.Run
	log     *memoryLog
	port    *ControlPort
	calls   []ControlCall
	sink    InterruptLog
	origins bool
	faces   domain.Interfaces
}

func newControlled(t *testing.T, script func(c *controlled, syncs int)) *controlled {
	t.Helper()
	c := &controlled{t: t, port: NewControlPort(), log: &memoryLog{}, run: runOf(t, testsupport.ThermalBoardDomain())}
	_, faces, err := loadInterfaces(c.run.Domain().Spec, []string{testsupport.ThermalBoardKindsDir()})
	if err != nil {
		t.Fatal(err)
	}
	c.faces = faces
	c.host = &scriptedHost{t: t, afterSync: func(h *scriptedHost, syncs int) {
		if script != nil {
			script(c, syncs)
		}
	}}
	return c
}

func (c *controlled) submit(t *testing.T, line string) int {
	t.Helper()
	request, err := domain.DecodeControl([]byte(line))
	if err != nil {
		t.Fatalf("DecodeControl(%s) = %v", line, err)
	}
	call, queued := c.port.enqueue(t.Context(), request)
	if !queued {
		t.Fatal("the control port took no request")
	}
	c.calls = append(c.calls, call)
	return len(c.calls) - 1
}

func (c *controlled) reply(t *testing.T, index int) map[string]any {
	t.Helper()
	select {
	case raw := <-c.calls[index].Reply:
		var reply map[string]any
		if err := json.Unmarshal(raw, &reply); err != nil {
			t.Fatal(err)
		}
		return reply
	default:
		t.Fatalf("control %d was not answered", index)
		return nil
	}
}

func (c *controlled) play(t *testing.T, barriers int64) {
	t.Helper()
	options := sliceOptions(barriers)
	options.Policies = []string{policy}
	options.Control, options.InterruptLog, options.JournalOrigins, options.Interfaces = c.port, c.log, c.origins, c.faces
	if c.sink != nil {
		options.InterruptLog = c.sink
	}
	if err := NewSession(c.run, c.host, options).Run(t.Context()); err != nil {
		t.Fatal(err)
	}
}

func (c *controlled) answers() map[string][]string {
	answers := map[string][]string{}
	for _, record := range c.host.sent {
		if id, ok := record["command_id"].(string); ok {
			answers[id] = append(answers[id], str(record["type"])+":"+str(record["status"])+str(record["reason"])+str(record["reported_outcome"]))
		}
	}
	return answers
}

func thermalCommand(id, device, command string, args map[string]any, now int64) map[string]any {
	return withKey(set(commandFields(id, device, command, args, now), "expected_boot_id", "mega-01-boot-1"))
}

func fanCommand(id string, now int64) map[string]any {
	return thermalCommand(id, "fan-01", "set_cooling", map[string]any{"level_percent": 30}, now)
}

const (
	holdBoard    = `{"op":"hold","scope":{"kind":"board","id":"mega-01"}}`
	releaseBoard = `{"op":"release","scope":{"kind":"board","id":"mega-01"}}`
	estopBoard   = `{"op":"estop","scope":{"kind":"board","id":"mega-01"}}`
	approveFan   = `{"op":"approve_mode","scope":{"kind":"device","id":"fan-01"},"enabled":true}`
)

func TestAHeldBoardRefusesTheHostsCommandAsLocalControlAndTheWorldNeverSeesIt(t *testing.T) {
	t.Parallel()
	c := newControlled(t, func(c *controlled, syncs int) {
		if syncs == 1 {
			c.host.reply("DCMD", fanCommand("c-held", sliceEnd(1)))
		}
	})
	hold := c.submit(t, holdBoard)
	c.play(t, 2)
	if got := c.answers()["c-held"]; !slices.Equal(got, []string{"DRECEIPT:refusedlocal_control"}) {
		t.Fatalf("answers = %v, want one refusal as local_control", got)
	}
	if calls := c.run.World.EffectorCalls(); len(calls) != 0 {
		t.Fatalf("the world saw %d effector calls, want none: a held command never reaches it", len(calls))
	}
	if reply := c.reply(t, hold); reply["result"] != "applied" {
		t.Fatalf("reply = %v, want the hold applied", reply)
	}
	if want := []string{"inventory", "control", "reply", "command", "verdict"}; !slices.Equal(c.log.events(), want) {
		t.Fatalf("log = %v, want %v", c.log.events(), want)
	}
	for _, record := range c.host.sent {
		if record["command_id"] == "c-held" && record["boot_id"] != "none" {
			t.Fatalf("receipt = %v, want a refusal before any boot", record)
		}
	}
}

func TestACommandWaitsForApprovalAndRunsExactlyAsIfItHadJustArrivedWhenTheOperatorApprovesIt(t *testing.T) {
	t.Parallel()
	c := newControlled(t, nil)
	c.host.afterSync = func(h *scriptedHost, syncs int) {
		switch syncs {
		case 1:
			h.reply("DCMD", fanCommand("c-1", sliceEnd(1)))
		case 2:
			c.submit(t, `{"op":"approve","command_id":"c-1"}`)
		}
	}
	c.submit(t, approveFan)
	c.play(t, 3)
	want := []string{"DRECEIPT:queued", "DRECEIPT:accepted", "DRESULT:succeeded"}
	if got := c.answers()["c-1"]; !slices.Equal(got, want) {
		t.Fatalf("answers = %v, want %v", got, want)
	}
	calls := c.run.World.EffectorCalls()
	if len(calls) != 1 || calls[0].Effector != "set_pwm_lease" || calls[0].AtNS < sliceEnd(2) {
		t.Fatalf("effector calls = %+v, want the command performed once, at the barrier that approved it", calls)
	}
	if reply := c.reply(t, 1); reply["result"] != "applied" || reply["command_id"] != "c-1" {
		t.Fatalf("approval reply = %v", reply)
	}
}

func TestADeniedCommandIsRefusedAndAWaitingOneExpiresAtItsDeadline(t *testing.T) {
	t.Parallel()
	c := newControlled(t, nil)
	c.host.afterSync = func(h *scriptedHost, syncs int) {
		switch syncs {
		case 1:
			h.reply("DCMD", fanCommand("c-deny", sliceEnd(1)))
			h.reply("DCMD", set(thermalCommand("c-late", "led-01", "show_state", map[string]any{"state": "alert"}, sliceEnd(1)),
				"deadline_ns", sliceEnd(1)+30e9))
		case 2:
			c.submit(t, `{"op":"deny","command_id":"c-deny"}`)
		}
	}
	c.submit(t, `{"op":"approve_mode","scope":{"kind":"board","id":"mega-01"},"enabled":true}`)
	c.play(t, 3)
	answers := c.answers()
	if !slices.Equal(answers["c-deny"], []string{"DRECEIPT:queued", "DRECEIPT:refusedlocal_control"}) {
		t.Fatalf("denied = %v", answers["c-deny"])
	}
	if !slices.Equal(answers["c-late"], []string{"DRECEIPT:queued", "DRECEIPT:refusedexpired"}) {
		t.Fatalf("late = %v, want it queued and then refused as expired at the barrier after its deadline", answers["c-late"])
	}
	if calls := c.run.World.EffectorCalls(); len(calls) != 0 {
		t.Fatalf("effector calls = %d, want none", len(calls))
	}
}

func TestAnEmergencyStopPerformsTheSafeStopOfEveryDeviceInScopeLatchesAHoldAndIsReleasedByTheOperator(t *testing.T) {
	t.Parallel()
	c := newControlled(t, nil)
	c.host.afterSync = func(h *scriptedHost, syncs int) {
		switch syncs {
		case 1:
			h.reply("DCMD", fanCommand("c-held", sliceEnd(1)))
		case 2:
			c.submit(t, releaseBoard)
		case 3:
			h.reply("DCMD", fanCommand("c-after", sliceEnd(3)))
		}
	}
	stop := c.submit(t, estopBoard)
	c.play(t, 4)
	reply := c.reply(t, stop)
	if stopped, _ := reply["stopped"].([]any); !slices.Equal(stopped, []any{"fan-01", "led-01"}) || reply["skipped"].([]any)[0] != "zone-01" {
		t.Fatalf("reply = %v, want fan-01 and led-01 stopped and zone-01, with no safe stop, skipped", reply)
	}
	answers := c.answers()
	if !slices.Equal(answers["c-held"], []string{"DRECEIPT:refusedlocal_control"}) || !slices.Equal(answers["c-after"], []string{"DRECEIPT:accepted", "DRESULT:succeeded"}) {
		t.Fatalf("answers = %v, want the latch to refuse c-held and the release to let c-after through", answers)
	}
	var effectors []string
	for _, call := range c.run.World.EffectorCalls() {
		effectors = append(effectors, call.Effector)
	}
	if !slices.Equal(effectors, []string{"stop_fan", "stop_led", "set_pwm_lease"}) {
		t.Fatalf("effector calls = %v, want the two safe stops and then the released command", effectors)
	}
	states := 0
	for _, record := range c.host.sent {
		if record["type"] == "DSTATE" {
			states++
		}
	}
	if states != 2 || !slices.Contains(c.log.events(), "stop") || !slices.Contains(c.log.events(), "effect") {
		t.Fatalf("states %d events %v, want a state per stopped device and the stops and their effects logged", states, c.log.events())
	}
}

func TestAHostsOwnSafeStopIsNeverHeldNorQueuedByTheNode(t *testing.T) {
	t.Parallel()
	c := newControlled(t, func(c *controlled, syncs int) {
		if syncs == 1 {
			c.host.reply("DCMD", thermalCommand("c-stop", "led-01", "stop_led", map[string]any{}, sliceEnd(1)))
		}
	})
	c.submit(t, `{"op":"hold","scope":{"kind":"all"}}`)
	c.submit(t, `{"op":"approve_mode","scope":{"kind":"all"},"enabled":true}`)
	c.play(t, 2)
	if got := c.answers()["c-stop"]; !slices.Equal(got, []string{"DRECEIPT:accepted", "DRESULT:succeeded"}) {
		t.Fatalf("answers = %v, want the safe stop performed through a hold and an approval mode", got)
	}
}

func TestAControlThatNamesWhatTheNodeDoesNotHaveIsRefusedWithItsReasonAndChangesNothing(t *testing.T) {
	t.Parallel()
	c := newControlled(t, nil)
	board := c.submit(t, `{"op":"hold","scope":{"kind":"board","id":"mega-99"}}`)
	device := c.submit(t, `{"op":"estop","scope":{"kind":"device","id":"ghost"}}`)
	command := c.submit(t, `{"op":"approve","command_id":"c-404"}`)
	c.play(t, 1)
	for index, reason := range map[int]string{board: "unknown_board", device: "unknown_device", command: "unknown_command"} {
		if reply := c.reply(t, index); reply["result"] != "refused" || reply["reason"] != reason {
			t.Errorf("control %d = %v, want refused as %s", index, reply, reason)
		}
	}
	if calls := c.run.World.EffectorCalls(); len(calls) != 0 {
		t.Fatalf("effector calls = %d, want a refused emergency stop to stop nothing", len(calls))
	}
}

func TestAControlNeverReachesTheNodeOnceTheCallerGaveUpAndTheStateIsAnsweredWithoutTheNode(t *testing.T) {
	t.Parallel()
	port := NewControlPort()
	given, cancel := context.WithCancel(t.Context())
	cancel()
	var busy map[string]any
	if err := json.Unmarshal(port.Handle(given, []byte(holdBoard)), &busy); err != nil || busy["reason"] != "node_busy" || busy["result"] != "refused" {
		t.Fatalf("reply = %v (%v), want refused as node_busy when nobody serves", busy, err)
	}
	if call, taken := port.take(); taken {
		t.Fatalf("the node took %+v of a caller that had given up", call)
	}
	var state, malformed map[string]any
	if err := json.Unmarshal(port.Handle(t.Context(), []byte(`{"op":"state"}`)), &state); err != nil || state["result"] != "applied" || state["op"] != "state" {
		t.Fatalf("state = %v (%v), want it answered from the published view", state, err)
	}
	if err := json.Unmarshal(port.Handle(t.Context(), []byte(`hold`)), &malformed); err != nil || malformed["reason"] != "malformed" {
		t.Fatalf("malformed = %v (%v), want it refused without the node", malformed, err)
	}
}
