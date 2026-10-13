package contract

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

const hourNS = int64(3600e9)

func factoryUtilities(t *testing.T) *domain.Compiled {
	t.Helper()
	spec, err := domain.Load(filepath.Join(testsupport.RepositoryRoot(), "examples", "factory-utilities.domain.json"))
	if err != nil {
		t.Fatal(err)
	}
	return spec
}

func TestComposingTheFactoryUtilitiesCompositionReproducesTheExampleByteForByte(t *testing.T) {
	t.Parallel()
	examples := filepath.Join(testsupport.RepositoryRoot(), "examples")
	want, err := os.ReadFile(filepath.Join(examples, "factory-utilities.domain.json"))
	if err != nil {
		t.Fatal(err)
	}
	if got := composedExample(t, filepath.Join(examples, "compositions", "factory-utilities.composition.json")); !bytes.Equal(got, want) {
		t.Fatalf("the committed factory-utilities example is not what streamsim domain compose writes; regenerate it with domain compose --out (%d bytes composed, %d committed)", len(got), len(want))
	}
}

func TestEveryKindOfTheFactoryComesFromTheBuiltInLibrary(t *testing.T) {
	t.Parallel()
	spec := factoryUtilities(t).Spec
	if len(spec.Instances) != len(spec.Kinds) {
		t.Fatalf("%d instance sets for %d kinds, want one set per composed kind", len(spec.Instances), len(spec.Kinds))
	}
	for _, set := range spec.Instances {
		if set.Provenance == nil || !strings.HasPrefix(set.Provenance.Component, "streamsim/") {
			t.Errorf("instances of %s: provenance %+v, want a component of the built-in library", set.Kind, set.Provenance)
		}
	}
}

func factoryMeans(t *testing.T, spec *domain.Compiled, fault []string, reads [][2]string) []float64 {
	t.Helper()
	w, err := world.New(spec, 7, "factory", model.DefaultStartTimeNS, world.Options{Noiseless: true, EmitDisabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if fault != nil {
		if _, err := w.InjectFault(fault[0], fault[1], model.DefaultStartTimeNS+8*hourNS, nil); err != nil {
			t.Fatal(err)
		}
	}
	means := make([]float64, len(reads))
	for at := 9 * hourNS; at < 10*hourNS; at += 10e9 {
		for i, read := range reads {
			means[i] += w.StateValue(read[0], read[1], model.DefaultStartTimeNS+at) / 360
		}
	}
	return means
}

type cascadeStep struct {
	entity, state string
	rise          float64
}

func TestAUtilityFaultCascadesThroughTheLinksToEveryMeterAndConsumerDownstream(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("simulates the factory for ten hours, four times")
	}
	testsupport.SkipUnderRace(t, "ten noiseless factory hours per case")
	spec := factoryUtilities(t)
	cases := []struct {
		name  string
		fault []string
		steps []cascadeStep
	}{
		{"a cooling tower fan failure warms the condenser, loads the chiller and shows on the cooling and main meters",
			[]string{"CT-1", "cooling_tower__fan_failure"},
			[]cascadeStep{{"CT-1", "basin_temp", 5}, {"CH-1", "cond_temp", 5}, {"CH-1", "power", 10}, {"MB-COOL", "power", 10}, {"MM-1", "power", 10}}},
		{"a receiver drain stuck open keeps the lag compressor loaded longer and shows on the air and main meters",
			[]string{"R-1", "air_receiver__drain_stuck_open"},
			[]cascadeStep{{"AC-2", "auto_load", 0.02}, {"AC-2", "power", 1}, {"MB-AIR", "power", 1}, {"MM-1", "power", 1}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			reads := make([][2]string, len(tc.steps))
			for i, step := range tc.steps {
				reads[i] = [2]string{step.entity, step.state}
			}
			clean, faulted := factoryMeans(t, spec, nil, reads), factoryMeans(t, spec, tc.fault, reads)
			for i, step := range tc.steps {
				t.Logf("%s.%s: %.3f clean, %.3f faulted", step.entity, step.state, clean[i], faulted[i])
				if faulted[i]-clean[i] < step.rise {
					t.Errorf("%s.%s: %.3f clean, %.3f faulted, want a rise of at least %v", step.entity, step.state, clean[i], faulted[i], step.rise)
				}
			}
		})
	}
}

func TestTheFactoryClosesItsAirAndCoolingLoopsWithSampledReads(t *testing.T) {
	t.Parallel()
	spec := factoryUtilities(t)
	want := map[[2]string]float64{{"compressor_lead", "header_pressure"}: 1, {"compressor_lag", "header_pressure"}: 1,
		{"molder", "air_pressure"}: 10, {"packaging_line", "coolant_temp"}: 10, {"chiller", "cond_water_in"}: 10, {"ahu", "chw_supply_in"}: 10}
	for key, sample := range want {
		dyn := spec.Kind(key[0]).DynamicsOf(key[1])
		if dyn == nil || dyn.Follow == nil || dyn.Follow.SampleS != sample {
			t.Errorf("%s.%s: follow %+v, want a read sampled every %v s", key[0], key[1], dyn, sample)
		}
	}
}
