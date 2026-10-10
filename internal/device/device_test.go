package device_test

import (
	"errors"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/device"
)

func TestFacadeAppliesACommandReportsStateAndSetsFaults(t *testing.T) {
	t.Parallel()
	d := device.New(device.Config{Capabilities: testCaps(t)})
	if state := d.State(); state["message_type"] != "state" {
		t.Fatalf("state = %v", state)
	}
	accepted := d.ApplyCommand(validCommand(t, nil))
	if accepted.Receipt["accepted"] != true {
		t.Fatalf("receipt = %v", accepted.Receipt)
	}
	d.SetFaults(device.Faults{AckLost: true})
	again := d.ApplyCommand(validCommand(t, func(c map[string]any) { c["command_id"] = "cmd-lost"; c["idempotency_key"] = "key-lost" }))
	if !again.AckLost {
		t.Fatalf("an ack_lost fault must withhold the receipt: %+v", again)
	}
}

func TestLoadCapabilitiesAndFaultSpecsAreValidatedAtTheFacade(t *testing.T) {
	t.Parallel()
	if _, err := device.LoadCapabilities([]byte(`{"not":"a catalog"}`)); err == nil {
		t.Fatal("an invalid catalog must be refused")
	}
	entry, err := device.ParseFaultSpec("ack_lost@2")
	if err != nil || entry.Name != "ack_lost" || entry.AcceptedCommand != 2 {
		t.Fatalf("parsed = %+v (%v)", entry, err)
	}
	if err := device.ValidateFaultSchedule([]device.FaultInjection{entry}); err != nil {
		t.Fatal(err)
	}
	if err := device.ValidateFaultSchedule([]device.FaultInjection{{Name: "no_such_fault", AcceptedCommand: 1}}); err == nil {
		t.Fatal("an unknown fault name must be refused")
	}
	if _, err := device.ParseFaultSpec("garbage"); err == nil {
		t.Fatal("an unparsable fault spec must be refused")
	}
}

func TestPlantErrorsAreTheFacadeSentinels(t *testing.T) {
	t.Parallel()
	if errors.Is(device.ErrPlantUnavailable, device.ErrPlantInterlocked) {
		t.Fatal("the plant refusals must stay distinct")
	}
}
