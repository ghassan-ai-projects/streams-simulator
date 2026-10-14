package contract

import (
	"slices"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

var catalogSkip = map[string]string{}

type catalogCase struct {
	ref, preset string
	doc         map[string]any
}

func (c catalogCase) name() string {
	if c.preset == "" {
		return c.ref
	}
	return c.ref + "/" + c.preset
}

type catalogPlant struct {
	catalogCase
	spec     *domain.Compiled
	entities []string
}

func (p catalogPlant) kind() *model.Kind { return &p.spec.Spec.Kinds[0] }

func catalogCases(t *testing.T) []catalogCase {
	t.Helper()
	var cases []catalogCase
	lib := loadLibrary(t)
	for _, ref := range sortedKeys(lib.files) {
		doc := lib.files[ref]
		if _, equipment := doc["measurement_points"]; !equipment {
			continue
		}
		for _, preset := range presetNames(doc) {
			cases = append(cases, catalogCase{ref: ref, preset: preset, doc: doc})
		}
	}
	return cases
}

func presetNames(doc map[string]any) []string {
	presets, _ := doc["presets"].(map[string]any)
	if len(presets) == 0 {
		return []string{""}
	}
	return sortedKeys(presets)
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

func TestCatalogSkipListStaysEmpty(t *testing.T) {
	t.Parallel()
	if len(catalogSkip) != 0 {
		t.Fatalf("the catalog harness skips %v; fix the component instead", sortedKeys(catalogSkip))
	}
}

func TestCatalogEveryEquipmentPreset(t *testing.T) {
	t.Parallel()
	testsupport.SkipUnderRace(t, "simulates every component of the library")
	if testing.Short() {
		t.Skip("simulates every component of the library; on demand: make test-long")
	}
	cases := catalogCases(t)
	if len(cases) == 0 {
		t.Fatal("the library holds no equipment")
	}
	for _, c := range cases {
		t.Run(strings.ReplaceAll(c.name(), "/", "_"), func(t *testing.T) {
			t.Parallel()
			if reason, skip := catalogSkip[c.name()]; skip {
				t.Skip(reason)
			}
			plant := composeCatalogPlant(t, c)
			for _, check := range catalogChecks {
				t.Run(check.name, func(t *testing.T) { check.run(t, plant) })
			}
		})
	}
}

var catalogChecks = []struct {
	name string
	run  func(*testing.T, catalogPlant)
}{
	{"lint", checkCatalogLint},
	{"steady_start", checkCatalogSteadyStart},
	{"replay", checkCatalogReplay},
	{"effectors", checkCatalogEffectors},
	{"faults_observable", checkCatalogFaultsObservable},
	{"trivial_baseline_audit", checkCatalogAudit},
}
