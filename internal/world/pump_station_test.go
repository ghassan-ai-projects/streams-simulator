package world_test

import (
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

const (
	hourS        = 3600.0
	nominalShaft = 3.1
	nominalAxial = 2.0
)

func stationSecond(s float64) int64 { return model.DefaultStartTimeNS + int64(s*1e9) }

func newPumpStation(t *testing.T) (*world.World, string) {
	t.Helper()
	spec, err := domain.Load(testsupport.PumpStationDomain())
	if err != nil {
		t.Fatal(err)
	}
	w, err := world.New(spec, 7, "w-pump-station", model.DefaultStartTimeNS, world.Options{Noiseless: true, ForceEffectorOK: true})
	if err != nil {
		t.Fatal(err)
	}
	return w, w.EntityIDs()[0]
}

func injectAndAdvance(t *testing.T, w *world.World, pump, fault string, hours float64) {
	t.Helper()
	if _, err := w.InjectFault(pump, fault, stationSecond(0), nil); err != nil {
		t.Fatal(err)
	}
	if _, _, err := w.Advance(stationSecond(hours * hourS)); err != nil {
		t.Fatal(err)
	}
}

func invoke(t *testing.T, w *world.World, effector, pump string, atHours float64, args map[string]any) {
	t.Helper()
	args["pump_id"] = pump
	if _, err := w.InvokeEffector(effector, pump, effector+"-1", args, stationSecond(atHours*hourS)); err != nil {
		t.Fatal(err)
	}
}

func stateAt(w *world.World, pump, state string, atHours float64) float64 {
	return w.StateValue(pump, state, stationSecond(atHours*hourS))
}

func TestMaintenanceNamingTheFaultRepairsIt(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		fault, state string
		nominal      float64
	}{{"bearing_wear", "vibration_true", nominalShaft}, {"misalignment", "axial_vibration_true", nominalAxial}} {
		t.Run(tc.fault, func(t *testing.T) {
			t.Parallel()
			w, pump := newPumpStation(t)
			injectAndAdvance(t, w, pump, tc.fault, 30)
			if worn := stateAt(w, pump, tc.state, 30); worn < tc.nominal+1 {
				t.Fatalf("%s at 30 h = %v, want a raised level", tc.state, worn)
			}
			invoke(t, w, "schedule_maintenance", pump, 30, map[string]any{"work_order": tc.fault})
			if _, _, err := w.Advance(stationSecond(40 * hourS)); err != nil {
				t.Fatal(err)
			}
			if w.ActiveFaultsCount() != 0 {
				t.Fatalf("%d faults active after the matching work order", w.ActiveFaultsCount())
			}
			if got := stateAt(w, pump, tc.state, 40); math.Abs(got-tc.nominal) > 0.1 {
				t.Fatalf("%s at 40 h = %v, want back near %v", tc.state, got, tc.nominal)
			}
		})
	}
}

func TestMaintenanceNamingAnotherFaultLeavesItActive(t *testing.T) {
	t.Parallel()
	w, pump := newPumpStation(t)
	injectAndAdvance(t, w, pump, "bearing_wear", 30)
	invoke(t, w, "schedule_maintenance", pump, 30, map[string]any{"work_order": "misalignment"})
	if _, _, err := w.Advance(stationSecond(40 * hourS)); err != nil {
		t.Fatal(err)
	}
	if w.ActiveFaultsCount() != 1 {
		t.Fatalf("%d faults active, want the wear to stay", w.ActiveFaultsCount())
	}
	if got := stateAt(w, pump, "vibration_true", 40); got < nominalShaft+1 {
		t.Fatalf("vibration at 40 h = %v, want it still high", got)
	}
}

func TestMaintenanceRepairsASensorFault(t *testing.T) {
	t.Parallel()
	for _, fault := range []string{"sensor_drift", "sensor_failure"} {
		t.Run(fault, func(t *testing.T) {
			t.Parallel()
			w, pump := newPumpStation(t)
			injectAndAdvance(t, w, pump, fault, 30)
			if bias := stateAt(w, pump, "vibration_bias", 30); bias < 0.5 {
				t.Fatalf("bias at 30 h = %v, want the probe lying", bias)
			}
			invoke(t, w, "schedule_maintenance", pump, 30, map[string]any{"work_order": fault})
			if _, _, err := w.Advance(stationSecond(32 * hourS)); err != nil {
				t.Fatal(err)
			}
			if bias := stateAt(w, pump, "vibration_bias", 32); bias > 0.01 {
				t.Fatalf("bias at 32 h = %v, want the probe recalibrated", bias)
			}
		})
	}
}

func TestReducingLoadMasksWearWithoutRepairingIt(t *testing.T) {
	t.Parallel()
	w, pump := newPumpStation(t)
	injectAndAdvance(t, w, pump, "bearing_wear", 30)
	before := stateAt(w, pump, "vibration_true", 30)
	invoke(t, w, "reduce_load", pump, 30, map[string]any{})
	if _, _, err := w.Advance(stationSecond(40 * hourS)); err != nil {
		t.Fatal(err)
	}
	if after := stateAt(w, pump, "vibration_true", 40); after > before-0.5 {
		t.Fatalf("vibration %v before and %v after reduce_load, want it lower", before, after)
	}
	if w.ActiveFaultsCount() != 1 {
		t.Fatalf("%d faults active, want the wear to go on", w.ActiveFaultsCount())
	}
}

func TestACavitationTransientSettlesOnItsOwn(t *testing.T) {
	t.Parallel()
	w, pump := newPumpStation(t)
	id, err := w.InjectFault(pump, "transient_none", stationSecond(0), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.ClearFault(id, stationSecond(hourS)); err != nil {
		t.Fatal(err)
	}
	if _, _, err := w.Advance(stationSecond(60)); err != nil {
		t.Fatal(err)
	}
	if burst := stateAt(w, pump, "vibration_true", 60.0/hourS); burst < nominalShaft+0.5 {
		t.Fatalf("vibration during the burst = %v, want a spike", burst)
	}
	if _, _, err := w.Advance(stationSecond(6 * hourS)); err != nil {
		t.Fatal(err)
	}
	if got := stateAt(w, pump, "vibration_true", 6); math.Abs(got-nominalShaft) > 0.01 {
		t.Fatalf("vibration at 6 h = %v, want it settled at %v with no action", got, nominalShaft)
	}
}
