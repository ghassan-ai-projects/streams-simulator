package contract

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/audit"
	"github.com/ghassan-ai-projects/streams-simulator/internal/component"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const (
	campusHalls      = 26
	campusMinTags    = 2000
	campusSpanNS     = int64(600e9)
	campusBudget     = 20 * time.Second
	campusRealTimeX  = float64(campusSpanNS) / float64(campusBudget)
	factoryAuditSpan = 4 * hourNS
)

func campusDomain(t *testing.T) *domain.Compiled {
	t.Helper()
	raw, err := os.ReadFile(testsupport.FactoryComposition())
	if err != nil {
		t.Fatal(err)
	}
	doc, err := testsupport.Campus(raw, campusHalls, "WX-1", "CAL-1")
	if err != nil {
		t.Fatal(err)
	}
	library, err := component.NewLibrary(nil)
	if err != nil {
		t.Fatal(err)
	}
	composed, err := component.Compose(library, doc, "campus.composition.json")
	if err != nil {
		t.Fatal(err)
	}
	return composed.Domain
}

func tagCount(spec *domain.Compiled) int {
	tags := 0
	for _, set := range spec.Spec.Instances {
		tags += len(set.Names) * len(spec.Kind(set.Kind).Channels)
	}
	return tags
}

func TestTheCampusFixtureCarriesTwoThousandTagsComposedOnlyFromTheFactory(t *testing.T) {
	t.Parallel()
	spec := campusDomain(t)
	if tags := tagCount(spec); tags < campusMinTags {
		t.Fatalf("the campus of %d halls has %d tags, want at least %d", campusHalls, tags, campusMinTags)
	}
	if got, want := len(spec.Spec.Kinds), len(factoryUtilities(t).Spec.Kinds); got != want {
		t.Fatalf("the campus has %d kinds, want the factory's %d: halls repeat instances, not kinds", got, want)
	}
}

//nolint:paralleltest // a wall-clock budget runs alone in its package
func TestTheCampusRunsTenMinutesThroughTheRunPipelineWithinItsSpeedBudget(t *testing.T) {
	if testing.Short() {
		t.Skip("the speed budget runs the 2,000-tag campus")
	}
	testsupport.SkipUnderRace(t, "a wall-clock budget means nothing under the race detector")
	spec := campusDomain(t)
	native, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	r, err := run.New(t.Context(), run.Config{Domain: spec, Adapter: native, Seed: 7, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: model.DefaultStartTimeNS})
	if err != nil {
		t.Fatal(err)
	}
	began := time.Now()
	if _, err := r.Advance(context.Background(), model.DefaultStartTimeNS+campusSpanNS, false); err != nil {
		t.Fatal(err)
	}
	art, err := r.End("")
	elapsed := time.Since(began)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%d tags, %d events in %v: %.0fx real time (budget %v, %.0fx)", tagCount(spec), art.Counts.Emitted, elapsed,
		float64(campusSpanNS)/float64(elapsed), campusBudget, campusRealTimeX)
	if elapsed > campusBudget {
		t.Fatalf("ten simulated minutes of the campus took %v, over the budget of %v", elapsed, campusBudget)
	}
}

func TestTheFactoryUtilityFaultsAreNotAllTrivialForAOneLineBaseline(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("audits three faults over four factory hours")
	}
	testsupport.SkipUnderRace(t, "single-goroutine audit worlds")
	spec := factoryUtilities(t)
	panel, err := audit.NewPanel(spec, 11, 60e9)
	if err != nil {
		t.Fatal(err)
	}
	entities := []string{"AC-1", "AC-2", "R-1", "CT-1", "CH-1"}
	nonTrivial := 0
	for _, fault := range [][2]string{{"R-1", "air_receiver__header_leak"}, {"AC-1", "compressor_lead__airend_wear"}, {"CT-1", "cooling_tower__fill_fouling"}} {
		start := model.DefaultStartTimeNS
		verdict, err := panel.Audit(fault[0], fault[1], start+hourNS, start, entities, factoryAuditSpan, nil, nil)
		if err != nil {
			t.Fatalf("%s: %v", fault[1], err)
		}
		t.Logf("%s: trivial=%v best=%s %.3f", fault[1], verdict.Trivial, verdict.Best, verdict.BestScore)
		if !verdict.Trivial {
			nonTrivial++
		}
	}
	if nonTrivial == 0 {
		t.Fatal("every audited factory fault is caught by a one-line baseline; the world is too easy to benchmark against")
	}
}
