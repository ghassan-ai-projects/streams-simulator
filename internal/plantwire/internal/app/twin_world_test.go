package app

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	simdomain "github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func thermalTwin(t *testing.T) *twinWorld {
	t.Helper()
	spec, err := simdomain.Load(testsupport.ThermalBoardDomain())
	if err != nil {
		t.Fatal(err)
	}
	devices := []domain.BoardDevice{{Name: "zone-01", Kind: "thermal_sensor"}, {Name: "fan-01", Kind: "fan"},
		{Name: "pump-1", Kind: "pump"}, {Name: "led-01", Kind: "fan"}}
	twin, err := newTwinWorld(spec, twinStartNS, devices)
	if err != nil {
		t.Fatal(err)
	}
	return twin
}

func fanLease(id string, at int64, duty string) domain.TwinCommand {
	return domain.TwinCommand{ID: id, Device: "fan-01", Name: "set_pwm_lease", AtNS: at,
		Args: map[string]any{"duty_permille": json.Number(duty), "lease_ms": json.Number("10000")}}
}

func TestATwinBindsADeviceToTheEntityOfItsIdAndKind(t *testing.T) {
	t.Parallel()
	twin := thermalTwin(t)
	tests := []struct{ device, metric, want string }{
		{"zone-01", "temperature", ""},
		{"zone-01", "humidity", ""},
		{"fan-01", "duty_permille", ""},
		{"zone-01", "pressure", domain.SkipNoChannel},
		{"pump-1", "flow", domain.SkipNoEntity},
		{"led-01", "duty_permille", domain.SkipNoEntity},
		{"ghost-9", "temperature", domain.SkipNoEntity},
	}
	for _, tt := range tests {
		if got := twin.Support(tt.device, tt.metric); got != tt.want {
			t.Errorf("Support(%s, %s) = %q, want %q", tt.device, tt.metric, got, tt.want)
		}
	}
}

func TestACommandMovesTheTwinsPredictionOnlyFromItsOwnTime(t *testing.T) {
	t.Parallel()
	twin := thermalTwin(t)
	at := twinStartNS + 60*twinTickNS
	if err := twin.Apply(fanLease("c-1", at, "600")); err != nil {
		t.Fatalf("Apply = %v, want the lease accepted", err)
	}
	if before := twin.Predict("zone-01", "temperature", at-twinTickNS); before != twinOpenTemp {
		t.Errorf("prediction before the command = %v, want the rest temperature %v", before, twinOpenTemp)
	}
	if after := twin.Predict("zone-01", "temperature", at+9*twinTickNS); after > twinOpenTemp-1 {
		t.Errorf("prediction 9 s into a full-duty lease = %v, want the zone cooled to at most %v", after, twinOpenTemp-1)
	}
	if duty := twin.Predict("fan-01", "duty_permille", at+twinTickNS); duty != 600 {
		t.Errorf("fan duty reading = %v, want the commanded 600", duty)
	}
}

func TestACommandTheTwinCannotTakeIsRefusedWithErrTwinRefused(t *testing.T) {
	t.Parallel()
	twin := thermalTwin(t)
	at := twinStartNS + twinTickNS
	if err := twin.Apply(fanLease("c-1", at, "300")); err != nil {
		t.Fatal(err)
	}
	unknownDevice := fanLease("c-2", at, "300")
	unknownDevice.Device = "ghost-9"
	unknownEffector := fanLease("c-3", at, "300")
	unknownEffector.Name = "no_such_effector"
	tests := []struct {
		name    string
		command domain.TwinCommand
		also    error
	}{
		{"a device with no twin entity", unknownDevice, nil},
		{"an effector the entity does not have", unknownEffector, nil},
		{"arguments outside the effector's schema", fanLease("c-4", at, "9999"), nil},
		{"a command id reused with other arguments", fanLease("c-1", at, "500"), world.ErrCommandIDReused},
	}
	for _, tt := range tests {
		err := twin.Apply(tt.command)
		if !errors.Is(err, ErrTwinRefused) || (tt.also != nil && !errors.Is(err, tt.also)) {
			t.Errorf("%s: Apply = %v, want ErrTwinRefused (and %v)", tt.name, err, tt.also)
		}
	}
}

func TestATwinCannotStartBeforeTheEpoch(t *testing.T) {
	t.Parallel()
	spec, err := simdomain.Load(testsupport.ThermalBoardDomain())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := newTwinWorld(spec, -1, nil); err == nil || !strings.Contains(err.Error(), "outside") {
		t.Fatalf("newTwinWorld(-1) = %v, want the world's refusal of a start outside its range", err)
	}
}
