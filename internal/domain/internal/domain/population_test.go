package domain

import (
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func populationFixture() *Compiled {
	return &Compiled{Spec: &model.DomainSpec{Instances: []model.InstanceSet{
		{Kind: "tank", Names: []string{"tank-a", "tank-b"}},
		{Kind: "pond", IDTemplate: "site-{site}/pond-{n}", Count: model.EntityCount{Default: 2}},
		{Kind: "pump", IDTemplate: "pump-{n}"},
	}}}
}

func TestPopulationIDsExpandNamesAndTemplatesInSetOrder(t *testing.T) {
	t.Parallel()
	got := PopulationIDs(populationFixture(), 0)
	want := []string{"tank-a", "tank-b", "site-a/pond-1", "site-a/pond-2", "pump-1"}
	if !slices.Equal(got, want) {
		t.Fatalf("ids = %v, want %v", got, want)
	}
}

func TestPopulationIDsOverrideOnlyTheFirstTemplatedSet(t *testing.T) {
	t.Parallel()
	got := PopulationIDs(populationFixture(), 3)
	want := []string{"tank-a", "tank-b", "site-a/pond-1", "site-a/pond-2", "site-a/pond-3", "pump-1"}
	if !slices.Equal(got, want) {
		t.Fatalf("ids = %v, want %v", got, want)
	}
}

func TestRenderInstanceIDFillsTheIndexAndUnknownValues(t *testing.T) {
	t.Parallel()
	if got := RenderInstanceID("rack-{rack}/server-{n}", 7); got != "rack-a/server-7" {
		t.Fatalf("id = %q", got)
	}
}
