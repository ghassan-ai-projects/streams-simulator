package model

import "testing"

func TestEveryRequestOfTheNodeControlContractIsAccepted(t *testing.T) {
	t.Parallel()
	for _, request := range []string{
		`{"op":"hold","scope":{"kind":"all"}}`,
		`{"op":"release","scope":{"kind":"board","id":"uno-07"}}`,
		`{"op":"estop","scope":{"kind":"device","id":"fan-01"}}`,
		`{"op":"approve_mode","scope":{"kind":"all"},"enabled":false}`,
		`{"op":"approve","command_id":"cmd-1"}`,
		`{"op":"deny","command_id":"cmd-1"}`,
		`{"op":"state"}`,
	} {
		if err := ValidateNodeControl([]byte(request)); err != nil {
			t.Errorf("%s: %v, want it accepted", request, err)
		}
	}
}

func TestARequestThatBreaksTheNodeControlContractIsRefusedAsAViolationOfIt(t *testing.T) {
	t.Parallel()
	for name, request := range map[string]string{
		"an operation the node does not have": `{"op":"reboot"}`,
		"a member nothing defines":            `{"op":"state","verbose":true}`,
		"a hold with no scope":                `{"op":"hold"}`,
		"a scope of all that names an id":     `{"op":"hold","scope":{"kind":"all","id":"b"}}`,
		"a device scope with no id":           `{"op":"hold","scope":{"kind":"device"}}`,
		"approval mode with no switch":        `{"op":"approve_mode","scope":{"kind":"all"}}`,
		"an approval with no command":         `{"op":"approve"}`,
		"a state request with a command":      `{"op":"state","command_id":"c-1"}`,
	} {
		err := ValidateNodeControl([]byte(request))
		invalid, ok := AsValidationError(err)
		if !ok || invalid.Contract != ContractNodeControl || len(invalid.Issues) == 0 {
			t.Errorf("%s: %v, want a violation of %s", name, err, ContractNodeControl)
		}
	}
}

func TestAnInterruptLogRecordIsValidatedAgainstItsContract(t *testing.T) {
	t.Parallel()
	valid := `{"event":"verdict","command_id":"c-1","verdict":"refuse","reason":"local_control","clock":"world","node_ns":5}`
	if err := ValidateInterruptRecord([]byte(valid)); err != nil {
		t.Fatalf("a verdict = %v, want it accepted", err)
	}
	for name, record := range map[string]string{
		"an event nobody logs":     `{"event":"explode","clock":"world","node_ns":5}`,
		"no clock":                 `{"event":"clock","node_ns":5}`,
		"a verdict with no id":     `{"event":"verdict","verdict":"pass","clock":"world","node_ns":5}`,
		"a stop with no command":   `{"event":"stop","device":"fan-01","clock":"world","node_ns":5}`,
		"a reason nobody gives":    `{"event":"verdict","command_id":"c-1","verdict":"refuse","reason":"because","clock":"world","node_ns":5}`,
		"a member nothing defines": `{"event":"clock","clock":"world","node_ns":5,"extra":1}`,
	} {
		err := ValidateInterruptRecord([]byte(record))
		invalid, ok := AsValidationError(err)
		if !ok || invalid.Contract != ContractInterruptLog {
			t.Errorf("%s: %v, want a violation of %s", name, err, ContractInterruptLog)
		}
	}
}
