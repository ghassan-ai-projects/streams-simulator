package domain

import (
	"testing"
)

func TestScheduledFaultsAreDeterministic(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		fault       string
		wantCode    string
		wantAckLost bool
		wantApply   int
		wantBoot    string
	}{
		{name: FaultStuck, fault: FaultStuck, wantApply: 0},
		{name: FaultAckLost, fault: FaultAckLost, wantAckLost: true, wantApply: 1},
		{name: FaultExpired, fault: FaultExpired, wantCode: FaultExpired, wantApply: 0},
		{name: FaultReboot, fault: FaultReboot, wantAckLost: true, wantApply: 1, wantBoot: "boot-reboot-1"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			plant := &countingPlant{}
			d := New(Config{
				Capabilities:  testCaps(t),
				Plant:         plant,
				FaultSchedule: []FaultInjection{{Name: tc.fault, AcceptedCommand: 1}},
			})
			out := d.ApplyCommand(validCommand(t, nil))
			if out.AckLost != tc.wantAckLost {
				t.Fatalf("ack_lost = %v, want %v: %+v", out.AckLost, tc.wantAckLost, out)
			}
			if tc.wantCode == "" && out.Receipt["accepted"] != true {
				t.Fatalf("fault %q should preserve acceptance: %v", tc.fault, out.Receipt)
			}
			if tc.wantCode != "" && out.Receipt["reject_code"] != tc.wantCode {
				t.Fatalf("fault %q reject_code = %v, want %s", tc.fault, out.Receipt["reject_code"], tc.wantCode)
			}
			if plant.applyCalls != tc.wantApply {
				t.Fatalf("fault %q plant calls = %d, want %d", tc.fault, plant.applyCalls, tc.wantApply)
			}
			if tc.wantBoot != "" && d.BootID() != tc.wantBoot {
				t.Fatalf("fault %q boot id = %q, want %q", tc.fault, d.BootID(), tc.wantBoot)
			}
			if tc.fault == FaultStuck && d.State()["current_output"].(map[string]any)["energized"] != false {
				t.Fatal("scheduled stuck fault must leave output de-energized")
			}
		})
	}
}

func TestParseFaultSpec(t *testing.T) {
	t.Parallel()
	cases := map[string]FaultInjection{
		"stuck":      {Name: FaultStuck, AcceptedCommand: 1},
		"ack_lost@3": {Name: FaultAckLost, AcceptedCommand: 3},
	}
	for input, want := range cases {
		got, err := ParseFaultSpec(input)
		if err != nil {
			t.Fatalf("parse %q: %v", input, err)
		}
		if got != want {
			t.Fatalf("parse %q = %+v, want %+v", input, got, want)
		}
	}
	for _, input := range []string{"unknown", "stuck@0", "stuck@x", "stuck@1@2"} {
		if _, err := ParseFaultSpec(input); err == nil {
			t.Fatalf("parse %q should fail", input)
		}
	}
}

// A wire fault belongs to one delivery. The retry of an identified command
// replays its evidence without repeating the fault, or a disconnecting link
// could never complete a command.
func TestWireFaultsAreNotReplayedWithTheEvidenceOfARetry(t *testing.T) {
	t.Parallel()
	cases := map[string]func(Outcome) bool{
		FaultDisconnect: func(o Outcome) bool { return o.Disconnect },
		FaultDuplicate:  func(o Outcome) bool { return o.Duplicate },
		FaultAckLost:    func(o Outcome) bool { return o.AckLost },
	}
	for fault, injected := range cases {
		t.Run(fault, func(t *testing.T) {
			t.Parallel()
			plant := &countingPlant{}
			d := New(Config{Capabilities: testCaps(t), Plant: plant,
				FaultSchedule: []FaultInjection{{Name: fault, AcceptedCommand: 1}}})
			command := validCommand(t, nil)
			first := d.ApplyCommand(command)
			if !injected(first) {
				t.Fatalf("the first delivery must carry the %s fault: %+v", fault, first)
			}
			retry := d.ApplyCommand(command)
			if injected(retry) {
				t.Fatalf("the retry must not repeat the %s fault: %+v", fault, retry)
			}
			if retry.Receipt["accepted"] != true || plant.applyCalls != 1 {
				t.Fatalf("the retry replays the accepted receipt without a second plant effect: %+v calls=%d", retry.Receipt, plant.applyCalls)
			}
		})
	}
}
