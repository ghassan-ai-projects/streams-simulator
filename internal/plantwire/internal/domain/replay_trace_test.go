package domain

import (
	"errors"
	"slices"
	"testing"
)

func traceOf(records ...Record) []map[string]any {
	gateway := NewTransmitter("gw-1")
	var trace []map[string]any
	for _, record := range records {
		gateway.Queue(record)
		trace = append(trace, gateway.Drain()[0])
	}
	return trace
}

func recordedSession(t *testing.T) []map[string]any {
	t.Helper()
	return traceOf(hardwareBirth(t, nil), deviceBirthOf("P1", "boot-1"), validData(), barrierSync(2),
		DeviceData("P1", "boot-1", false, []map[string]any{Sample(1, 2.75, "", 9)}), Receipt("c-1", "accepted", "", "boot-1"),
		Result("c-1", "succeeded", "boot-1"), NodeDeath(0, "shutdown"))
}

func TestAReplayAnnouncesEveryBoardAsSimulatedAndKeepsTheRest(t *testing.T) {
	t.Parallel()
	trace := recordedSession(t)
	plan, err := PlanReplay(trace)
	if err != nil || len(plan.Chunks) != 2 || plan.Births != 1 || plan.Skipped != 3 {
		t.Fatalf("plan = %+v, %v, want two chunks split at the recorded NSYNC, one birth and the receipt, the result and the death skipped", plan, err)
	}
	first := plan.Chunks[0].Records
	if types := []string{first[0].Type, first[1].Type, first[2].Type}; !slices.Equal(types, []string{"NBIRTH", "DBIRTH", "DDATA"}) || plan.Chunks[1].Records[0].Type != "DDATA" {
		t.Fatalf("chunks = %v, want the births and data in the order they were recorded", plan.Chunks)
	}
	birth := relayDecode(t, firstOf(first[0]))
	if birth.Invalid != nil || birth.Environment != EnvironmentSimulated || EnvironmentClaimFault(birth.Message) != nil {
		t.Fatalf("replayed NBIRTH %v (%v), want a valid simulated node", birth.Record, birth.Invalid)
	}
	for _, board := range birth.Boards {
		if board.Environment != EnvironmentSimulated {
			t.Errorf("board %s announced %s, a replay is not a board", board.Board, board.Environment)
		}
	}
	original := relayDecode(t, trace[0])
	if original.Environment != EnvironmentHardware || original.Boards[0].Environment != EnvironmentHardware {
		t.Fatal("the recording was rewritten by planning its replay")
	}
	announced := map[string]string{}
	for _, kind := range birth.Kinds {
		announced[kindNamed(kind)] = digestOfKind(kind)
	}
	for _, kind := range original.Kinds {
		if announced[kindNamed(kind)] != digestOfKind(kind) {
			t.Errorf("kind %s replays with digest %q, was recorded with %q", kindNamed(kind), announced[kindNamed(kind)], digestOfKind(kind))
		}
	}
	if birth.CapabilityDigest != original.CapabilityDigest || len(announced) != len(original.Kinds) {
		t.Fatalf("capability %q kinds %d, want the recorded capability digest and every kind", birth.CapabilityDigest, len(announced))
	}
	if plan.Boots["P1"] != "boot-1" || plan.Chunks[0].WorldTimeNS != 1 || plan.Chunks[1].WorldTimeNS != 9 {
		t.Fatalf("boots %v world times %d %d, want the recorded boot and the latest event time of each chunk", plan.Boots, plan.Chunks[0].WorldTimeNS, plan.Chunks[1].WorldTimeNS)
	}
}

func TestAReplayIsSimulatedWhateverTheRecordingClaimed(t *testing.T) {
	t.Parallel()
	for _, environment := range []string{EnvironmentHardware, EnvironmentMixed, EnvironmentSimulated} {
		trace := traceOf(hardwareBirth(t, func(fields map[string]any) { fields["environment"] = environment }))
		plan, err := PlanReplay(trace)
		if err != nil {
			t.Fatal(err)
		}
		if got := plan.Chunks[0].Records[0].Fields["environment"]; got != EnvironmentSimulated {
			t.Errorf("a recording that claimed %s replays as %v, want simulated", environment, got)
		}
	}
}

func TestAReplayPlaysARebirthAsARebirthAndANewSessionAsOne(t *testing.T) {
	t.Parallel()
	first := traceOf(hardwareBirth(t, nil), deviceBirthOf("P1", "boot-1"), hardwareBirth(t, nil))
	second := NewTransmitter("gw-2")
	second.Queue(hardwareBirth(t, nil))
	plan, err := PlanReplay(append(first, second.Drain()[0]))
	if err != nil || plan.Births != 3 {
		t.Fatalf("plan = %+v, %v, want three births: the first, a rebirth in its session and a new session's birth", plan, err)
	}
	var bdSeqs []uint64
	for _, record := range plan.Chunks[0].Records {
		if record.Type == "NBIRTH" {
			bdSeqs = append(bdSeqs, record.Fields["bd_seq"].(uint64))
		}
	}
	if !slices.Equal(bdSeqs, []uint64{0, 1, 2}) {
		t.Fatalf("bd_seq of the births = %v, want the replay's own count", bdSeqs)
	}
}

func TestAReplaySkipsWhatItCannotPlayAndCountsIt(t *testing.T) {
	t.Parallel()
	birth, device := traceOf(hardwareBirth(t, nil))[0], traceOf(hardwareBirth(t, nil), deviceBirthOf("P1", "boot-1"))[1]
	broken := copyOf(device)
	delete(broken, "metrics")
	repeated := copyOf(device)
	cases := map[string]struct {
		trace   []map[string]any
		skipped int
		err     error
	}{
		"records before the first birth": {[]map[string]any{device, birth}, 1, nil},
		"a record that fails its schema": {[]map[string]any{birth, broken}, 1, nil},
		"a record that repeats a seq":    {[]map[string]any{birth, device, repeated}, 1, nil},
		"a record that is no record":     {[]map[string]any{birth, nil}, 1, nil},
		"no birth at all":                {[]map[string]any{device}, 0, ErrNoReplayBirth},
		"an empty trace":                 {nil, 0, ErrNoReplayBirth},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			plan, err := PlanReplay(tc.trace)
			if !errors.Is(err, tc.err) || (tc.err == nil && plan.Skipped != tc.skipped) {
				t.Fatalf("PlanReplay = %+v, %v, want %d skipped and %v", plan, err, tc.skipped, tc.err)
			}
		})
	}
}

func copyOf(record map[string]any) map[string]any {
	copied := make(map[string]any, len(record))
	for member, value := range record {
		copied[member] = value
	}
	return copied
}

func firstOf(record Record) map[string]any {
	out := NewTransmitter("rp-1")
	out.Queue(record)
	return out.Drain()[0]
}
