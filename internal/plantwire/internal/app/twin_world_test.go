package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	simdomain "github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func thermalTwin(t *testing.T) *twinWorld {
	t.Helper()
	devices := []domain.BoardDevice{{Name: "zone-01", Kind: "thermal_sensor"}, {Name: "fan-01", Kind: "fan"},
		{Name: "pump-1", Kind: "pump"}, {Name: "led-01", Kind: "fan"}}
	twin, err := newTwinWorld(thermalModel(t), twinStartNS, devices)
	if err != nil {
		t.Fatal(err)
	}
	return twin
}

func thermalModel(t *testing.T) twinModel {
	t.Helper()
	spec, err := simdomain.Load(testsupport.ThermalBoardDomain())
	if err != nil {
		t.Fatal(err)
	}
	_, faces, err := loadInterfaces(spec.Spec, []string{testsupport.ThermalBoardKindsDir()})
	if err != nil {
		t.Fatal(err)
	}
	return twinModel{spec: spec, faces: faces}
}

func fanLease(id string, at int64, level string) domain.TwinCommand {
	return domain.TwinCommand{ID: id, Device: "fan-01", Name: "set_cooling", AtNS: at, Args: map[string]any{"level_percent": json.Number(level)}}
}

func TestATwinBindsADeviceToTheEntityOfItsIdAndKind(t *testing.T) {
	t.Parallel()
	twin := thermalTwin(t)
	tests := []struct{ device, metric, want string }{
		{"zone-01", "temperature", ""},
		{"zone-01", "humidity", ""},
		{"fan-01", "level_percent", ""},
		{"fan-01", "duty_permille", domain.SkipNoChannel},
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
	for i, renewal := range []int64{at, at + 5*twinTickNS} {
		if err := twin.Apply(fanLease(fmt.Sprintf("c-%d", i+1), renewal, "60")); err != nil {
			t.Fatalf("Apply = %v, want the lease accepted", err)
		}
	}
	if before := twin.Predict("zone-01", "temperature", at-twinTickNS); before != twinOpenTemp {
		t.Errorf("prediction before the command = %v, want the rest temperature %v", before, twinOpenTemp)
	}
	if after := twin.Predict("zone-01", "temperature", at+9*twinTickNS); after > twinOpenTemp-1 {
		t.Errorf("prediction 9 s into a renewed full-duty lease = %v, want the zone cooled to at most %v", after, twinOpenTemp-1)
	}
	if level := twin.Predict("fan-01", "level_percent", at+twinTickNS); level != 60 {
		t.Errorf("fan level reading = %v, want the commanded 60", level)
	}
}

func TestACommandTheTwinCannotTakeIsRefusedWithErrTwinRefused(t *testing.T) {
	t.Parallel()
	twin := thermalTwin(t)
	at := twinStartNS + twinTickNS
	if err := twin.Apply(fanLease("c-1", at, "30")); err != nil {
		t.Fatal(err)
	}
	unknownDevice := fanLease("c-2", at, "30")
	unknownDevice.Device = "ghost-9"
	unknownEffector := fanLease("c-3", at, "30")
	unknownEffector.Name = "no_such_effector"
	tests := []struct {
		name    string
		command domain.TwinCommand
		also    error
	}{
		{"a device with no twin entity", unknownDevice, nil},
		{"an effector the entity does not have", unknownEffector, nil},
		{"the raw command the kind file does not offer", domain.TwinCommand{ID: "c-5", Device: "fan-01", Name: "set_pwm_lease", AtNS: at,
			Args: map[string]any{"duty_permille": json.Number("300"), "lease_ms": json.Number("5000")}}, domain.ErrNotTranslated},
		{"arguments outside the effector's schema", fanLease("c-4", at, "999"), nil},
		{"a command id reused with other arguments", fanLease("c-1", at, "50"), world.ErrCommandIDReused},
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
	if _, err := newTwinWorld(thermalModel(t), -1, nil); err == nil || !strings.Contains(err.Error(), "outside") {
		t.Fatalf("newTwinWorld(-1) = %v, want the world's refusal of a start outside its range", err)
	}
}
