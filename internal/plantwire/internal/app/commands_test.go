package app

import (
	"errors"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	wire "github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

const policy = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

func commandFields(id, device, command string, args map[string]any, now int64) map[string]any {
	fields := map[string]any{"command_id": id, "device": device, "command": command, "args": args, "expected_boot_id": device + "-boot-1",
		"deadline_ns": now + 120e9, "not_before_ns": 0, "authority_epoch": 1, "policy_digest": policy, "fallback_on_lease_loss": "safe_off"}
	return withKey(fields)
}

func withKey(fields map[string]any) map[string]any {
	fields["idempotency_key"] = wire.Command{Device: fields["device"].(string), Name: fields["command"].(string),
		Args: fields["args"].(map[string]any), Boot: fields["expected_boot_id"].(string)}.Content()
	return fields
}

func sliceEnd(syncs int) int64 { return model.DefaultStartTimeNS + int64(syncs)*60e9 }

func commandSession(t *testing.T, slices int64, script func(h *scriptedHost, syncs int), adjust func(*Options)) (*scriptedHost, map[string][]string) {
	t.Helper()
	host := &scriptedHost{t: t, afterSync: script}
	options := sliceOptions(slices)
	options.Policies, options.SafeStop = []string{policy}, []string{"stop_chiller"}
	if adjust != nil {
		adjust(&options)
	}
	r := utilityRun(t)
	if err := NewSession(r, host, options).Run(t.Context()); err != nil {
		t.Fatal(err)
	}
	answers := map[string][]string{}
	for _, record := range host.sent {
		if id, ok := record["command_id"].(string); ok {
			answer := record["type"].(string) + ":" + str(record["status"]) + str(record["reason"]) + str(record["reported_outcome"])
			answers[id] = append(answers[id], answer)
		}
	}
	return host, answers
}

func str(value any) string {
	text, _ := value.(string)
	return text
}

func TestAnAcceptedCommandActsOnTheWorldAndReportsItsResult(t *testing.T) {
	t.Parallel()
	_, answers := commandSession(t, 2, func(h *scriptedHost, syncs int) {
		if syncs == 1 {
			h.reply("DCMD", commandFields("c-1", "meter-1", "shed_load", map[string]any{"meter_id": "meter-1"}, sliceEnd(1)))
		}
	}, nil)
	if got := answers["c-1"]; len(got) != 2 || got[0] != "DRECEIPT:accepted" || got[1] != "DRESULT:succeeded" {
		t.Fatalf("answers %v", got)
	}
}

func TestEveryRefusalClassIsReportedByItsCheck(t *testing.T) {
	t.Parallel()
	now := sliceEnd(1)
	cases := map[string]map[string]any{
		"unknown_device":     commandFields("c-device", "pump-77", "shed_load", map[string]any{"meter_id": "x"}, now),
		"unknown_command":    commandFields("c-command", "meter-1", "fly", map[string]any{}, now),
		"out_of_range":       commandFields("c-args", "meter-1", "shed_load", map[string]any{}, now),
		"wrong_boot":         withKey(set(commandFields("c-boot", "meter-1", "shed_load", map[string]any{"meter_id": "m"}, now), "expected_boot_id", "meter-1-boot-9")),
		"expired":            set(commandFields("c-late", "meter-1", "shed_load", map[string]any{"meter_id": "m"}, now), "deadline_ns", now-1),
		"stale_epoch":        set(commandFields("c-epoch", "meter-1", "shed_load", map[string]any{"meter_id": "m"}, now), "authority_epoch", 0),
		"policy_not_allowed": set(commandFields("c-policy", "meter-1", "shed_load", map[string]any{"meter_id": "m"}, now), "policy_digest", "sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"),
	}
	_, answers := commandSession(t, 2, func(h *scriptedHost, syncs int) {
		if syncs == 1 {
			for _, fields := range cases {
				h.reply("DCMD", fields)
			}
		}
	}, nil)
	for reason, fields := range cases {
		if got := answers[fields["command_id"].(string)]; len(got) != 1 || got[0] != "DRECEIPT:refused"+reason {
			t.Errorf("%s: answers %v", reason, got)
		}
	}
}

func set(fields map[string]any, key string, value any) map[string]any {
	fields[key] = value
	return fields
}

func TestARepeatedKeyReplaysItsAnswerAndOtherContentIsADuplicate(t *testing.T) {
	t.Parallel()
	first := commandFields("c-1", "meter-1", "shed_load", map[string]any{"meter_id": "meter-1"}, sliceEnd(1))
	again := set(commandFields("c-2", "meter-1", "shed_load", map[string]any{"meter_id": "meter-1"}, sliceEnd(1)), "idempotency_key", first["idempotency_key"])
	other := set(commandFields("c-3", "meter-1", "shed_load", map[string]any{"meter_id": "meter-2"}, sliceEnd(1)), "idempotency_key", first["idempotency_key"])
	_, answers := commandSession(t, 2, func(h *scriptedHost, syncs int) {
		if syncs == 1 {
			h.reply("DCMD", first)
			h.reply("DCMD", again)
			h.reply("DCMD", other)
		}
	}, nil)
	if answers["c-2"][0] != "DRECEIPT:accepted" || answers["c-3"][0] != "DRECEIPT:refusedduplicate" {
		t.Fatalf("answers %v", answers)
	}
}

func TestAQueuedCommandRunsWhenDueAndIsCheckedAgain(t *testing.T) {
	t.Parallel()
	_, answers := commandSession(t, 4, func(h *scriptedHost, syncs int) {
		if syncs == 1 {
			h.reply("DCMD", set(commandFields("c-later", "meter-1", "shed_load", map[string]any{"meter_id": "m"}, sliceEnd(1)), "not_before_ns", sliceEnd(2)))
			h.reply("DCMD", set(commandFields("c-epoch", "meter-1", "shed_load", map[string]any{"meter_id": "n"}, sliceEnd(1)), "not_before_ns", sliceEnd(3)))
		}
		if syncs == 3 {
			h.reply("STATE", map[string]any{"host_id": "host-a", "online": true, "heartbeat": 1, "heartbeat_interval_ms": 1000, "authority_epoch": 2, "protocol_min": 1, "protocol_max": 1})
		}
	}, nil)
	if got := answers["c-later"]; len(got) != 3 || got[0] != "DRECEIPT:queued" || got[1] != "DRECEIPT:accepted" || got[2] != "DRESULT:succeeded" {
		t.Fatalf("a due queued command did not run: %v", got)
	}
	if got := answers["c-epoch"]; len(got) != 2 || got[0] != "DRECEIPT:queued" || got[1] != "DRECEIPT:refusedstale_epoch" {
		t.Fatalf("a queued command was not re-checked at execution: %v", got)
	}
}

func TestASafeStopSkipsEpochAndPolicyAndAPermissiveRefusalIsReported(t *testing.T) {
	t.Parallel()
	_, answers := commandSession(t, 3, func(h *scriptedHost, syncs int) {
		if syncs == 1 {
			stop := set(set(commandFields("c-stop", "CH1", "stop_chiller", map[string]any{"chiller_id": "CH1"}, sliceEnd(1)), "authority_epoch", 0), "policy_digest", policy[:len(policy)-1]+"c")
			h.reply("DCMD", stop)
		}
		if syncs == 2 {
			h.reply("DCMD", commandFields("c-start", "CH1", "start_chiller", map[string]any{"chiller_id": "CH1"}, sliceEnd(2)))
		}
	}, nil)
	if answers["c-stop"][0] != "DRECEIPT:accepted" || answers["c-start"][0] != "DRECEIPT:refusedpermissive_not_met" {
		t.Fatalf("answers %v", answers)
	}
}

func TestALostAcknowledgementIsSilentAndASilentNoEffectStillReportsSuccess(t *testing.T) {
	t.Parallel()
	for mode, want := range map[world.FailureMode]int{world.ModeAckLost: 0, world.ModeSilentNoEffect: 2} {
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
		if got := countOf(host.types(), "DRECEIPT") + countOf(host.types(), "DRESULT"); got != want {
			t.Fatalf("%s: %d answers, want %d", mode, got, want)
		}
	}
}

func TestAQueuedCommandNeverRunsAfterTheHostIsLost(t *testing.T) {
	t.Parallel()
	host := &scriptedHost{t: t, afterSync: func(h *scriptedHost, syncs int) {
		if syncs == 1 {
			h.reply("DCMD", set(commandFields("c-later", "meter-1", "shed_load", map[string]any{"meter_id": "m"}, sliceEnd(1)), "not_before_ns", sliceEnd(2)))
			h.reply("HSYNC", map[string]any{"through_seq": uint64(0)})
			h.silent = true
		}
	}}
	options := sliceOptions(6)
	options.Policies = []string{policy}
	r := utilityRun(t)
	if err := NewSession(r, host, options).Run(t.Context()); !errors.Is(err, ErrHostLost) {
		t.Fatalf("err = %v, want host loss", err)
	}
	if calls := r.World.EffectorCalls(); len(calls) != 0 || countOf(host.types(), "DRECEIPT") != 1 {
		t.Fatalf("calls %v, records %v", calls, host.types())
	}
}

func TestARepeatOfAQueuedCommandWaitsForItsOneExecution(t *testing.T) {
	t.Parallel()
	queued := set(commandFields("c-1", "meter-1", "shed_load", map[string]any{"meter_id": "m"}, sliceEnd(1)), "not_before_ns", sliceEnd(2))
	again := set(commandFields("c-2", "meter-1", "shed_load", map[string]any{"meter_id": "m"}, sliceEnd(1)), "not_before_ns", sliceEnd(2))
	_, answers := commandSession(t, 3, func(h *scriptedHost, syncs int) {
		if syncs == 1 {
			h.reply("DCMD", queued)
			h.reply("DCMD", again)
		}
	}, nil)
	for _, id := range []string{"c-1", "c-2"} {
		if got := answers[id]; len(got) != 3 || got[0] != "DRECEIPT:queued" || got[1] != "DRECEIPT:accepted" {
			t.Fatalf("%s: %v", id, got)
		}
	}
}

func TestARetryLearnsTheAnswerWhoseAcknowledgementWasLost(t *testing.T) {
	t.Parallel()
	first := commandFields("c-1", "meter-1", "shed_load", map[string]any{"meter_id": "m"}, sliceEnd(1))
	host := &scriptedHost{t: t, afterSync: func(h *scriptedHost, syncs int) {
		if syncs == 1 {
			h.reply("DCMD", first)
		}
		if syncs == 2 {
			h.reply("DCMD", set(commandFields("c-retry", "meter-1", "shed_load", map[string]any{"meter_id": "m"}, sliceEnd(2)), "idempotency_key", first["idempotency_key"]))
			h.replyInvalid(map[string]any{"type": "DCMD", "command_id": "bad id with spaces"})
		}
	}}
	r := utilityRun(t)
	r.SetFailureMode(world.ModeAckLost)
	options := sliceOptions(3)
	options.Policies = []string{policy}
	if err := NewSession(r, host, options).Run(t.Context()); err != nil {
		t.Fatal(err)
	}
	var ids []any
	for _, record := range host.sent {
		if record["type"] == "DRECEIPT" {
			ids = append(ids, record["command_id"])
		}
	}
	if len(ids) != 1 || ids[0] != "c-retry" || len(r.World.EffectorCalls()) != 1 {
		t.Fatalf("receipts for %v, calls %d", ids, len(r.World.EffectorCalls()))
	}
}

func TestASafeStopWithPreconditionsIsRefusedAtStart(t *testing.T) {
	t.Parallel()
	options := sliceOptions(1)
	options.SafeStop = []string{"start_chiller"}
	if err := NewSession(utilityRun(t), &scriptedHost{t: t}, options).Run(t.Context()); !errors.Is(err, ErrBadOptions) {
		t.Fatalf("err = %v, want a refusal of a guarded safe stop", err)
	}
}
