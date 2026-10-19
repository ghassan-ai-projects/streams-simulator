package domain

import "testing"

func twBoardWith(receipts map[string]BoardReceipt) BoardTrace {
	return BoardTrace{Devices: []BoardDevice{{Name: "fan-01", Kind: "fan", Board: "mega-01"}}, Receipts: receipts, StartNS: 100}
}

func twFanCommand(id string) HostCommand {
	return HostCommand{ID: id, Device: "fan-01", Name: "set_pwm_lease"}
}

func TestEveryCommandIsClassifiedByWhatTheBoardAndInterruptsSaidOfIt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		receipts map[string]BoardReceipt
		blocks   map[string]InterruptBlock
		status   string
		reason   string
		detail   string
		atNS     int64
	}{
		{"accepted and not failed is applied", map[string]BoardReceipt{"c-1": {Status: "accepted", AtNS: 150}}, nil,
			CommandApplied, "", "", 150},
		{"no receipt was never forwarded", nil, nil, CommandNotApplied, NotAppliedNeverForwarded, "", 100},
		{"a board refusal carries the board's reason", map[string]BoardReceipt{"c-1": {Status: "refused", Reason: "permissive_not_met", AtNS: 150}},
			nil, CommandNotApplied, NotAppliedRefusedByBoard, "permissive_not_met", 150},
		{"a receipt that is not accepted is not accepted", map[string]BoardReceipt{"c-1": {Status: "queued", AtNS: 150}}, nil,
			CommandNotApplied, NotAppliedNotAccepted, "queued", 150},
		{"an accepted command the board reported as failed", map[string]BoardReceipt{"c-1": {Status: "accepted", Outcome: "failed", AtNS: 150}},
			nil, CommandNotApplied, NotAppliedBoardFailed, "", 150},
		{"a refused interrupt blocks an accepted command", map[string]BoardReceipt{"c-1": {Status: "accepted", AtNS: 150}},
			map[string]InterruptBlock{"c-1": {Reason: "local_control"}}, CommandNotApplied, NotAppliedInterruptRefused, "local_control", 150},
		{"an unresolved interrupt blocks as unresolved", map[string]BoardReceipt{"c-1": {Status: "accepted", AtNS: 150}},
			map[string]InterruptBlock{"c-1": {Reason: "queued", Unresolved: true}}, CommandNotApplied, NotAppliedInterruptUnresolved, "queued", 150},
		{"an interrupt is read before the board's refusal", map[string]BoardReceipt{"c-1": {Status: "refused", Reason: "x", AtNS: 150}},
			map[string]InterruptBlock{"c-1": {Reason: "local_control"}}, CommandNotApplied, NotAppliedInterruptRefused, "local_control", 150},
		{"a receipt before the first board record is stamped at the start", map[string]BoardReceipt{"c-1": {Status: "accepted", AtNS: 10}},
			nil, CommandApplied, "", "", 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ClassifyCommands([]HostCommand{twFanCommand("c-1")}, twBoardWith(tt.receipts), tt.blocks)
			if len(got) != 1 {
				t.Fatalf("ClassifyCommands = %+v, want one command", got)
			}
			view := got[0]
			if view.Status != tt.status || view.Reason != tt.reason || view.Detail != tt.detail || view.AtNS != tt.atNS {
				t.Fatalf("view = %+v, want status %s reason %q detail %q at %d", view, tt.status, tt.reason, tt.detail, tt.atNS)
			}
		})
	}
}

func TestACommandToAnUnknownDeviceIsDropped(t *testing.T) {
	t.Parallel()
	got := ClassifyCommands([]HostCommand{{ID: "c-9", Device: "ghost-9", Name: "set_pwm_lease"}}, twBoardWith(nil), nil)
	if len(got) != 0 {
		t.Fatalf("ClassifyCommands = %+v, want the command to an unknown device dropped", got)
	}
}

func TestACommandKeepsItsIdentityAndArguments(t *testing.T) {
	t.Parallel()
	command := HostCommand{ID: "c-1", Device: "fan-01", Name: "set_pwm_lease", Args: map[string]any{"duty_permille": 300}}
	got := ClassifyCommands([]HostCommand{command}, twBoardWith(nil), nil)
	if len(got) != 1 || got[0].ID != "c-1" || got[0].Device != "fan-01" || got[0].Name != "set_pwm_lease" || got[0].Args["duty_permille"] != 300 {
		t.Fatalf("ClassifyCommands = %+v, want the identity and arguments of the host command", got)
	}
}
