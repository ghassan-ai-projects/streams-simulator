package contract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/component"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

const (
	beltLength  = 60.0
	beltSpeed   = 2.0
	beltEmptyN  = 1000.0
	beltPerKgN  = 0.494
	breakawayN  = 1630.0
	driveRunCap = 1.5
	tripTheta   = 1.1025
	ambientC    = 35.0
)

type beltRegime struct {
	name                    string
	breakaway, rated, boost float64
	wear                    float64
}

var beltRegimes = []beltRegime{
	{name: "a", breakaway: 1.424, rated: 920, boost: 1.5, wear: 0.55},
	{name: "b", breakaway: 1.152, rated: 607, boost: 1.75, wear: 0.135},
}

func beltPlant(t *testing.T, regime string) *domain.Compiled {
	t.Helper()
	spec, err := domain.Load(testsupport.BeltConveyorDomain(regime))
	if err != nil {
		t.Fatal(err)
	}
	return spec
}

func beltVariant(t *testing.T, regime string, edit func(doc map[string]any)) *domain.Compiled {
	t.Helper()
	raw, err := os.ReadFile(testsupport.BeltConveyorComposition(regime))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	edit(doc)
	if raw, err = json.Marshal(doc); err != nil {
		t.Fatal(err)
	}
	library, err := component.NewLibrary(nil)
	if err != nil {
		t.Fatal(err)
	}
	composed, err := component.Compose(library, raw, "variant.composition.json")
	if err != nil {
		t.Fatal(err)
	}
	return composed.Domain
}

func driveParameters(doc map[string]any) map[string]any {
	return doc["equipment"].([]any)[2].(map[string]any)["parameters"].(map[string]any)
}

type beltWorld struct {
	t        *testing.T
	w        *world.World
	nowS     float64
	commands int
}

func newBeltWorld(t *testing.T, spec *domain.Compiled) *beltWorld {
	t.Helper()
	return newSeededBeltWorld(t, spec, 5)
}

func newSeededBeltWorld(t *testing.T, spec *domain.Compiled, seed uint64) *beltWorld {
	t.Helper()
	w, err := world.New(spec, seed, "belt", model.DefaultStartTimeNS, world.Options{Noiseless: true, EmitDisabled: true})
	if err != nil {
		t.Fatal(err)
	}
	return &beltWorld{t: t, w: w}
}

func beltNS(seconds float64) int64 { return model.DefaultStartTimeNS + int64(math.Round(seconds*1e9)) }

func (b *beltWorld) advance(seconds float64) {
	b.t.Helper()
	if _, _, err := b.w.Advance(beltNS(seconds)); err != nil {
		b.t.Fatal(err)
	}
	b.nowS = seconds
}

func (b *beltWorld) state(entity, name string) float64 {
	return b.w.StateValue(entity, name, beltNS(b.nowS))
}

func (b *beltWorld) inject(entity, fault string, severity float64) {
	b.t.Helper()
	if _, err := b.w.InjectFault(entity, fault, beltNS(b.nowS), map[string]any{"severity": severity}); err != nil {
		b.t.Fatal(err)
	}
}

func (b *beltWorld) invoke(effector, entity string, args map[string]any) bool {
	b.t.Helper()
	b.commands++
	result, err := b.w.InvokeEffector(effector, entity, fmt.Sprintf("c-%d", b.commands), effectorArgs(effector, entity, args), beltNS(b.nowS))
	if err != nil {
		b.t.Fatal(err)
	}
	return result.Accepted && result.Reason == ""
}

// firstTime steps one second at a time until the state reaches the value, and returns that time.
func (b *beltWorld) firstTime(entity, name string, value, untilS float64) float64 {
	b.t.Helper()
	for b.nowS < untilS {
		b.advance(b.nowS + 1)
		if b.state(entity, name) == value {
			return b.nowS
		}
	}
	return math.Inf(1)
}

// beltDemand is the force the belt needs while running: empty belt, material, accelerating the feed, and the drives' drag.
func beltDemand(b *beltWorld, rated float64, enabled ...float64) float64 {
	demand := beltEmptyN + beltPerKgN*b.state("B-1", "m") + b.state("B-1", "v")*b.state("B-1", "feed")
	for _, on := range enabled {
		demand += 0.03 * rated * (2 - on)
	}
	return demand
}

func TestComposingTheBeltConveyorCompositionsReproducesTheExamplesByteForByte(t *testing.T) {
	t.Parallel()
	for _, regime := range []string{"a", "b"} {
		want, err := os.ReadFile(testsupport.BeltConveyorDomain(regime))
		if err != nil {
			t.Fatal(err)
		}
		if got := composedExample(t, testsupport.BeltConveyorComposition(regime)); !bytes.Equal(got, want) {
			t.Fatalf("plant %s: the committed example is not what streamsim domain compose writes; regenerate it with domain compose --out", regime)
		}
	}
}

func TestBothBeltPlantsStartAtTheirNominalSteadyState(t *testing.T) {
	t.Parallel()
	for _, regime := range beltRegimes {
		b := newBeltWorld(t, beltPlant(t, regime.name))
		before := map[string]float64{}
		for _, entity := range b.w.EntityIDs() {
			for _, name := range b.w.StateNamesOf(entity) {
				before[entity+"."+name] = b.state(entity, name)
			}
		}
		b.advance(600)
		for key, a := range before {
			entity, name, _ := strings.Cut(key, ".")
			if after := b.state(entity, name); math.Abs(after-a) > 0.03*math.Max(math.Abs(a), math.Abs(after))+0.05 {
				t.Errorf("plant %s: %s moves from %.4g to %.4g in ten minutes", regime.name, key, a, after)
			}
		}
	}
}

func TestTheWeakDriveAloneStartsSteadyLintsCleanAndKeepsAHardFault(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("simulates the weak drive for a day and audits its faults")
	}
	testsupport.SkipUnderRace(t, "catalog checks of one drive")
	library, err := component.NewLibrary(nil)
	if err != nil {
		t.Fatal(err)
	}
	ref := "streamsim/conveyor-drive@1"
	entry := map[string]any{"use": ref, "names": []string{"U-1"}, "parameters": map[string]any{"margin": "weak"}}
	composed, err := component.Compose(library, catalogDocument(t, []any{entry}, nil), "weak-drive.composition.json")
	if err != nil {
		t.Fatal(err)
	}
	plant := catalogPlant{catalogCase: catalogCase{ref: ref, doc: loadLibrary(t).files[ref]}, spec: composed.Domain, entities: []string{"U-1"}}
	for _, check := range catalogChecks {
		if check.name != "effectors" {
			t.Run(check.name, func(t *testing.T) {
				t.Parallel()
				check.run(t, plant)
			})
		}
	}
}
