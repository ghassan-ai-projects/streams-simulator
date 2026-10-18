package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func TestASafeStopToAnUnknownDeviceOrCommandIsRefusedForOneReason(t *testing.T) {
	t.Parallel()
	_, answers := commandSession(t, 2, func(h *scriptedHost, syncs int) {
		if syncs == 1 {
			h.reply("DCMD", commandFields("c-nowhere", "nope", "stop_chiller", map[string]any{"chiller_id": "CH1"}, sliceEnd(1)))
			h.reply("DCMD", set(commandFields("c-wrong", "meter-1", "stop_chiller", map[string]any{}, sliceEnd(1)), "expected_boot_id", "old"))
		}
	}, nil)
	if got := answers["c-nowhere"]; !slices.Equal(got, []string{"DRECEIPT:refusedunknown_device"}) {
		t.Fatalf("a safe stop to an unknown device: %v", got)
	}
	if got := answers["c-wrong"]; !slices.Equal(got, []string{"DRECEIPT:refusedunknown_command"}) {
		t.Fatalf("a safe stop the device does not have: %v", got)
	}
}

func TestEveryCheckRunsBeforeTheIdempotencyLedger(t *testing.T) {
	t.Parallel()
	first := commandFields("c-1", "meter-1", "shed_load", map[string]any{"meter_id": "a"}, sliceEnd(1))
	_, answers := commandSession(t, 3, func(h *scriptedHost, syncs int) {
		switch syncs {
		case 1:
			h.reply("DCMD", first)
		case 2:
			other := commandFields("c-2", "meter-1", "shed_load", map[string]any{"meter_id": "b"}, sliceEnd(2))
			h.reply("DCMD", set(set(other, "idempotency_key", first["idempotency_key"]), "authority_epoch", 0))
			h.reply("DCMD", set(commandFields("c-3", "meter-1", "shed_load", map[string]any{"meter_id": "a"}, sliceEnd(2)), "deadline_ns", sliceEnd(2)-1))
			h.reply("DCMD", set(commandFields("c-4", "meter-1", "shed_load", map[string]any{"meter_id": "c"}, sliceEnd(2)), "not_before_ns", sliceEnd(2)+25*3600e9))
		}
	}, nil)
	want := map[string]string{"c-2": "DRECEIPT:refusedstale_epoch", "c-3": "DRECEIPT:refusedexpired", "c-4": "DRECEIPT:refusedout_of_range"}
	for id, answer := range want {
		if got := answers[id]; !slices.Equal(got, []string{answer}) {
			t.Errorf("%s: %v, want %s", id, got, answer)
		}
	}
}

func TestTheQueueAndAPendingKeyAreBounded(t *testing.T) {
	t.Parallel()
	_, answers := commandSession(t, 1, func(h *scriptedHost, syncs int) {
		later := func(id string, meter any) map[string]any {
			return set(commandFields(id, "meter-1", "shed_load", map[string]any{"meter_id": meter}, sliceEnd(1)), "not_before_ns", sliceEnd(5))
		}
		for i := range maxQueued + 1 {
			h.reply("DCMD", later(fmt.Sprintf("q-%d", i), fmt.Sprint(i)))
		}
		for i := range maxWaiting {
			h.reply("DCMD", later(fmt.Sprintf("r-%d", i), "0"))
		}
	}, nil)
	if got := answers[fmt.Sprintf("q-%d", maxQueued)]; !slices.Equal(got, []string{"DRECEIPT:refusednot_ready"}) {
		t.Fatalf("the command past the queue's bound: %v", got)
	}
	if last, before := answers[fmt.Sprintf("r-%d", maxWaiting-1)], answers[fmt.Sprintf("r-%d", maxWaiting-2)]; !slices.Equal(last, []string{"DRECEIPT:refusednot_ready"}) ||
		!slices.Equal(before, []string{"DRECEIPT:queued"}) {
		t.Fatalf("repeats of a pending key: %v then %v", before, last)
	}
}

func TestEverySafeStopActsEvenWhenItRepeats(t *testing.T) {
	t.Parallel()
	stop := func(id string) map[string]any {
		return commandFields(id, "CH1", "stop_chiller", map[string]any{"chiller_id": "CH1"}, sliceEnd(1))
	}
	host := &scriptedHost{t: t, afterSync: func(h *scriptedHost, syncs int) {
		if syncs == 1 {
			h.reply("DCMD", stop("c-stop-1"))
			h.reply("DCMD", stop("c-stop-2"))
			h.reply("DCMD", stop("c-stop-1"))
		}
	}}
	options := sliceOptions(2)
	options.Policies = []string{policy}
	r := utilityRun(t)
	if err := NewSession(r, host, options).Run(t.Context()); err != nil {
		t.Fatal(err)
	}
	if calls := r.World.EffectorCalls(); len(calls) != 3 || countOf(host.types(), "DRESULT") != 3 {
		t.Fatalf("calls %d, records %v", len(calls), host.types())
	}
}

func TestARecordFromAnotherSessionIsMalformed(t *testing.T) {
	t.Parallel()
	_, answers := commandSession(t, 2, func(h *scriptedHost, syncs int) {
		if syncs == 1 {
			h.reply("DCMD", set(commandFields("c-foreign", "meter-1", "shed_load", map[string]any{"meter_id": "a"}, sliceEnd(1)), "session_id", "s-other"))
		}
	}, nil)
	if got := answers["c-foreign"]; !slices.Equal(got, []string{"DRECEIPT:refusedmalformed"}) {
		t.Fatalf("a command from another session: %v", got)
	}
}

func TestCommandsReadBeforeTheHostHangsUpAreNotExecuted(t *testing.T) {
	t.Parallel()
	host := &scriptedHost{t: t, afterSync: func(h *scriptedHost, syncs int) {
		h.reply("DCMD", commandFields("c-last", "meter-1", "shed_load", map[string]any{"meter_id": "a"}, sliceEnd(1)))
		h.silent, h.hangUp = true, io.EOF
	}}
	options := sliceOptions(3)
	options.Policies = []string{policy}
	r := utilityRun(t)
	if err := NewSession(r, host, options).Run(t.Context()); !errors.Is(err, ErrHostLost) {
		t.Fatalf("err = %v, want host loss", err)
	}
	if calls := r.World.EffectorCalls(); len(calls) != 0 {
		t.Fatalf("a command from a lost host ran: %v", calls)
	}
}

func TestAnAcceptedAPartialAndASilentNoEffectCommandLookTheSame(t *testing.T) {
	t.Parallel()
	var answers []string
	for _, mode := range []world.FailureMode{world.ModeOK, world.ModePartial, world.ModeSilentNoEffect} {
		host := &scriptedHost{t: t, afterSync: func(h *scriptedHost, syncs int) {
			if syncs == 1 {
				h.reply("DCMD", commandFields("c-1", "meter-1", "shed_load", map[string]any{"meter_id": "meter-1"}, sliceEnd(1)))
			}
		}}
		r := utilityRun(t)
		r.SetFailureMode(mode)
		options := sliceOptions(2)
		options.Policies = []string{policy}
		if err := NewSession(r, host, options).Run(t.Context()); err != nil {
			t.Fatal(err)
		}
		var seen []byte
		for _, record := range host.sent {
			if record["type"] == "DRECEIPT" || record["type"] == "DRESULT" {
				line, _ := json.Marshal(record)
				seen = append(seen, line...)
			}
		}
		answers = append(answers, string(seen))
	}
	if answers[0] == "" || answers[1] != answers[0] || answers[2] != answers[0] {
		t.Fatalf("the device's answers differ by hidden outcome:\n%s", answers)
	}
}
