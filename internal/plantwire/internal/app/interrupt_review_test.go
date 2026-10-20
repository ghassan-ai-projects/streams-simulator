package app

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func laterFan(id string, notBefore int64) map[string]any {
	return set(set(fanCommand(id, sliceEnd(1)), "not_before_ns", notBefore), "deadline_ns", sliceEnd(1)+600e9)
}

func effectorsOf(c *controlled) []string {
	var names []string
	for _, call := range c.run.World.EffectorCalls() {
		names = append(names, call.Effector)
	}
	return names
}

func TestACommandWaitingForItsStartTimeIsRefusedWhenAnEmergencyStopCoversItAndNeverRunsLater(t *testing.T) {
	t.Parallel()
	c := newControlled(t, nil)
	c.host.afterSync = func(h *scriptedHost, syncs int) {
		switch syncs {
		case 1:
			h.reply("DCMD", laterFan("c-later", sliceEnd(3)))
		case 2:
			c.submit(t, estopBoard)
		}
	}
	c.play(t, 4)
	if got := c.answers()["c-later"]; !slices.Equal(got, []string{"DRECEIPT:queued", "DRECEIPT:refusedlocal_control"}) {
		t.Fatalf("answers = %v, want the queued command refused as local_control by the stop and nothing after", got)
	}
	if got := effectorsOf(c); !slices.Equal(got, []string{"stop_fan", "stop_led"}) {
		t.Fatalf("effector calls = %v, want the two safe stops alone: the stopped fan was not started again by a command accepted before the stop", got)
	}
	if !slices.Contains(c.log.events(), "purge") {
		t.Fatalf("log = %v, want the refusal of the waiting command recorded", c.log.events())
	}
}

func TestACommandWaitingForItsStartTimeIsRefusedWhenAHoldOrApprovalModeIsSetOverIt(t *testing.T) {
	t.Parallel()
	for name, line := range map[string]string{"a hold": holdBoard, "approval mode": approveFan} {
		c := newControlled(t, nil)
		c.host.afterSync = func(h *scriptedHost, syncs int) {
			switch syncs {
			case 1:
				h.reply("DCMD", laterFan("c-later", sliceEnd(3)))
			case 2:
				c.submit(t, line)
			}
		}
		c.play(t, 4)
		if got := c.answers()["c-later"]; !slices.Equal(got, []string{"DRECEIPT:queued", "DRECEIPT:refusedlocal_control"}) || len(effectorsOf(c)) != 0 {
			t.Errorf("%s: answers %v effectors %v, want the command refused and never run unapproved", name, got, effectorsOf(c))
		}
	}
}

func TestACommandWaitingForItsStartTimeOutsideTheScopeStillRuns(t *testing.T) {
	t.Parallel()
	c := newControlled(t, nil)
	c.host.afterSync = func(h *scriptedHost, syncs int) {
		switch syncs {
		case 1:
			h.reply("DCMD", laterFan("c-later", sliceEnd(3)))
		case 2:
			c.submit(t, `{"op":"hold","scope":{"kind":"device","id":"led-01"}}`)
		}
	}
	c.play(t, 4)
	if got := c.answers()["c-later"]; !slices.Equal(got, []string{"DRECEIPT:queued", "DRECEIPT:accepted", "DRESULT:succeeded"}) {
		t.Fatalf("answers = %v, want a command the hold does not cover to run when it is due", got)
	}
}

func TestTwoCommandsUnderOneIdNeverLetTheOperatorApproveOneWhileTheOtherRuns(t *testing.T) {
	t.Parallel()
	c := newControlled(t, nil)
	c.host.afterSync = func(h *scriptedHost, syncs int) {
		switch syncs {
		case 1:
			h.reply("DCMD", fanCommand("c-1", sliceEnd(1)))
			h.reply("DCMD", thermalCommand("c-1", "led-01", "show_state", map[string]any{"state": "alert"}, sliceEnd(1)))
		case 2:
			c.submit(t, `{"op":"approve","command_id":"c-1"}`)
		}
	}
	c.submit(t, `{"op":"approve_mode","scope":{"kind":"board","id":"mega-01"},"enabled":true}`)
	c.play(t, 3)
	want := []string{"DRECEIPT:queued", "DRECEIPT:refusedduplicate", "DRECEIPT:accepted", "DRESULT:succeeded"}
	if got := c.answers()["c-1"]; !slices.Equal(got, want) {
		t.Fatalf("answers = %v, want %v", got, want)
	}
	if got := effectorsOf(c); !slices.Equal(got, []string{"set_pwm_lease"}) {
		t.Fatalf("effector calls = %v, want the fan command that was shown and approved, not the led command", got)
	}
}

var errHostGone = errors.New("the host is gone")

type receiptFailingHost struct {
	*scriptedHost
	fail bool
}

func (h *receiptFailingHost) Send(records []map[string]any) error {
	for _, record := range records {
		if h.fail && record["type"] == "DRECEIPT" {
			return errHostGone
		}
	}
	return h.scriptedHost.Send(records)
}

func TestTheSafeStopsOfAnEmergencyStopAreOrderedEvenWhenAnsweringTheHostFails(t *testing.T) {
	t.Parallel()
	c := newControlled(t, nil)
	link := &receiptFailingHost{scriptedHost: c.host}
	c.host.afterSync = func(h *scriptedHost, syncs int) {
		switch syncs {
		case 1:
			h.reply("DCMD", fanCommand("c-1", sliceEnd(1)))
		case 2:
			link.fail = true
			c.submit(t, estopBoard)
		}
	}
	c.submit(t, approveFan)
	options := sliceOptions(3)
	options.Policies = []string{policy}
	options.Control, options.InterruptLog, options.Interfaces = c.port, c.log, c.faces
	err := NewSession(c.run, link, options).Run(t.Context())
	if !errors.Is(err, errHostGone) {
		t.Fatalf("Run = %v, want the failed answer reported", err)
	}
	if got := effectorsOf(c); !slices.Equal(got, []string{"stop_fan", "stop_led"}) {
		t.Fatalf("effector calls = %v, want both safe stops performed although the receipt to the host failed", got)
	}
}

func TestAnApprovalStillWaitingWhenTheSessionEndsIsRefusedInTheLogAsTheSessionEnding(t *testing.T) {
	t.Parallel()
	c := newControlled(t, func(c *controlled, syncs int) {
		if syncs == 1 {
			c.host.reply("DCMD", fanCommand("c-1", sliceEnd(1)))
		}
	})
	c.submit(t, approveFan)
	c.play(t, 1)
	events := c.log.events()
	if tail := events[len(events)-2:]; !slices.Equal(tail, []string{"flush", "resolve"}) {
		t.Fatalf("log = %v, want the waiting command resolved when the session ends", events)
	}
	if err := domain.ReplayInterrupts(c.log.records); err != nil {
		t.Fatalf("ReplayInterrupts = %v, want a log that ends with its flush to replay", err)
	}
}

func TestAnEmergencyStopRepliesWithTheDevicesItCouldNotOrderEvenWhenThereAreNone(t *testing.T) {
	t.Parallel()
	c := newControlled(t, nil)
	stop := c.submit(t, estopBoard)
	c.play(t, 1)
	reply := c.reply(t, stop)
	unreachable, present := reply["unreachable"].([]any)
	if !present || len(unreachable) != 0 || len(reply["stopped"].([]any)) != 2 {
		t.Fatalf("reply = %v, want the unreachable list present and empty beside two stopped devices", reply)
	}
}

func TestALogThatDoesNotReplayIsSetAsideAndTheSessionSaysSo(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	file, err := openInterruptFile(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Add(map[string]any{"event": domain.EventVerdict, "command_id": "c-1", "verdict": domain.VerdictPass, "clock": domain.ClockWorld, "node_ns": 1}); err != nil {
		t.Fatal(err)
	}
	err = file.finish(dir)
	if !errors.Is(err, ErrInterruptLogRejected) || !errors.Is(err, domain.ErrInterruptReplay) {
		t.Fatalf("finish = %v, want ErrInterruptLogRejected naming the replay failure", err)
	}
	if file.digest() != "" {
		t.Fatal("a rejected log still has a digest to record")
	}
	if _, err := os.Stat(filepath.Join(dir, model.FileInterruptLog)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Stat of the log = %v, want it gone from the evidence names", err)
	}
	if _, err := os.Stat(filepath.Join(dir, model.FileInterruptLogRejected)); err != nil {
		t.Fatalf("Stat of the set-aside log = %v, want it kept for a human to read", err)
	}
	if _, _, err := run.ReadInterruptLog(dir); !errors.Is(err, run.ErrInterruptLogMissing) {
		t.Fatalf("ReadInterruptLog = %v, want the folder to hold no interrupt log", err)
	}
}

type stateFailingHost struct{ *scriptedHost }

func (h *stateFailingHost) Send(records []map[string]any) error {
	for _, record := range records {
		if record["type"] == "DSTATE" {
			return errHostGone
		}
	}
	return h.scriptedHost.Send(records)
}

func TestEveryDeviceOfAnEmergencyStopIsStoppedEvenWhenTellingTheHostTheStateOfTheFirstFails(t *testing.T) {
	t.Parallel()
	c := newControlled(t, nil)
	link := &stateFailingHost{scriptedHost: c.host}
	c.submit(t, estopBoard)
	options := sliceOptions(2)
	options.Policies = []string{policy}
	options.Control, options.InterruptLog, options.Interfaces = c.port, c.log, c.faces
	if err := NewSession(c.run, link, options).Run(t.Context()); !errors.Is(err, errHostGone) {
		t.Fatalf("Run = %v, want the failed state record reported", err)
	}
	if got := effectorsOf(c); !slices.Equal(got, []string{"stop_fan", "stop_led"}) {
		t.Fatalf("effector calls = %v, want both devices stopped: a host that cannot be told must not leave a second device running", got)
	}
}

func TestAnApprovedCommandWaitingForItsStartTimeAndPurgedByAHoldLeavesALogThatStillReplays(t *testing.T) {
	t.Parallel()
	c := newControlled(t, nil)
	c.host.afterSync = func(h *scriptedHost, syncs int) {
		switch syncs {
		case 1:
			h.reply("DCMD", laterFan("c-later", sliceEnd(5)))
		case 2:
			c.submit(t, `{"op":"approve","command_id":"c-later"}`)
		case 3:
			c.submit(t, holdBoard)
		}
	}
	c.submit(t, approveFan)
	c.play(t, 6)
	if !slices.Contains(c.log.events(), "purge") || len(effectorsOf(c)) != 0 {
		t.Fatalf("log %v effectors %v, want the approved command purged by the hold and never run", c.log.events(), effectorsOf(c))
	}
	if err := domain.ReplayInterrupts(c.log.records); err != nil {
		t.Fatalf("ReplayInterrupts = %v, want a purge of an approved command to be evidence the rules explain", err)
	}
}
