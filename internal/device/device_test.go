package device_test

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/device"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestFacadeAppliesACommandReportsStateAndSetsFaults(t *testing.T) {
	t.Parallel()
	d := device.New(device.Config{Capabilities: testCaps(t), FaultSchedule: []device.FaultInjection{{Name: device.FaultAckLost, AcceptedCommand: 2}}})
	if state := d.State(); state["message_type"] != "state" {
		t.Fatalf("state = %v", state)
	}
	accepted := d.ApplyCommand(validCommand(t, nil))
	if accepted.Receipt["accepted"] != true {
		t.Fatalf("receipt = %v", accepted.Receipt)
	}
	again := d.ApplyCommand(validCommand(t, func(c map[string]any) { c["command_id"] = "cmd-lost"; c["idempotency_key"] = "key-lost" }))
	if !again.AckLost {
		t.Fatalf("an ack_lost fault must withhold the receipt: %+v", again)
	}
}

func TestLoadCapabilitiesAndFaultSpecsAreValidatedAtTheFacade(t *testing.T) {
	t.Parallel()
	if _, err := device.LoadCapabilities([]byte(`{"not":"a catalog"}`)); err == nil || !strings.Contains(err.Error(), "device: decode capabilities: json: unknown field \"not\"") {
		t.Fatal("an invalid catalog must be refused")
	}
	entry, err := device.ParseFaultSpec("ack_lost@2")
	if err != nil || entry.Name != "ack_lost" || entry.AcceptedCommand != 2 {
		t.Fatalf("parsed = %+v (%v)", entry, err)
	}
	if err := device.ValidateFaultSchedule([]device.FaultInjection{entry}); err != nil {
		t.Fatal(err)
	}
	if err := device.ValidateFaultSchedule([]device.FaultInjection{{Name: "no_such_fault", AcceptedCommand: 1}}); err == nil || !strings.Contains(err.Error(), "device: unknown fault \"no_such_fault\"") {
		t.Fatal("an unknown fault name must be refused")
	}
	if _, err := device.ParseFaultSpec("garbage"); err == nil || !strings.Contains(err.Error(), "device: unknown fault \"garbage\"") {
		t.Fatal("an unparsable fault spec must be refused")
	}
}

func TestPlantErrorsAreTheFacadeSentinels(t *testing.T) {
	t.Parallel()
	if errors.Is(device.ErrPlantUnavailable, device.ErrPlantInterlocked) {
		t.Fatal("the plant refusals must stay distinct")
	}
}

func TestFacadeLoadsCapabilitiesFromAFileAndRefusesAMissingOne(t *testing.T) {
	t.Parallel()
	caps, err := device.LoadCapabilitiesFile(testsupport.DeviceCapabilities())
	if err != nil || caps.Digest() == "" {
		t.Fatalf("capabilities = %v, %v", caps, err)
	}
	if _, err := device.LoadCapabilitiesFile(filepath.Join(t.TempDir(), "absent.json")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a missing catalog: err = %v, want ErrNotExist", err)
	}
}

func TestServeRefusesASocketPathItCannotBindAndNamesIt(t *testing.T) {
	t.Parallel()
	socket := filepath.Join(t.TempDir(), "absent", "d.sock")
	err := device.Serve(io.Discard, socket, device.New(device.Config{}), "banner")
	var pathFailure *os.PathError
	if !errors.As(err, &pathFailure) || pathFailure.Path != socket {
		t.Fatalf("Serve = %v, want a path failure naming %s", err, socket)
	}
}
