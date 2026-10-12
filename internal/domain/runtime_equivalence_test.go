package domain_test

import (
	"encoding/json"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

var v01RuntimeDigests = map[string]string{
	"aquaculture-pond":      "sha256:efb583f5c24fd2897d7d85c245949a30d3b833b214668fb1be7fe494c3ca4420",
	"cold-chain-transit":    "sha256:2657324d9bd69afb345b104f807c40bb16291bc2134cb729f42c534d0486d81b",
	"discrete-line-oee":     "sha256:2f88ce0f8a153156d361f7b098b2fbf43eae60c3d8f3a279cee0ff6a2b00726d",
	"host-system-health":    "sha256:db7136029d8d189bda986373761fa5ffbbc5c8553e50f4e159b8fdaadd45095c",
	"k8s-cluster":           "sha256:4ec72d2d90943aa8fa321a307c99010ad0a315bdfd68c4cf0e932351ce24d08f",
	"open-field-irrigation": "sha256:d68c2ceb8f043f2c15272e26986474b03f7dc36e20255cfb5068586f952002cc",
	"rotating-machinery":    "sha256:4a34e8662c04c28f677011cf66e14b392ebb36ffb290bcb41d4071c53fcb0874",
	"shipment-exception":    "sha256:d50e7cb1facbdec18299f39551e77a5680ecba5fb91091cf6ec6561cbace3263",
}

func TestEveryShippedDomainCompilesToItsV01RuntimeWorld(t *testing.T) {
	t.Parallel()
	for id, want := range v01RuntimeDigests {
		t.Run(id, func(t *testing.T) {
			t.Parallel()
			got, err := canonical.Digest(v01Form(t, shipped(t, id).Spec))
			if err != nil || got != want {
				t.Fatalf("runtime spec digest = %s (%v), want the v0.1 runtime %s", got, err, want)
			}
		})
	}
}

func v01Form(t *testing.T, spec *model.DomainSpec) map[string]any {
	t.Helper()
	if len(spec.Kinds) != 1 || len(spec.Instances) != 1 {
		t.Fatalf("%s declares %d kinds and %d instance sets, but the v0.1 form holds one of each", spec.ID, len(spec.Kinds), len(spec.Instances))
	}
	whole, kind, instances := generic(t, spec), generic(t, spec.Kinds[0]), generic(t, spec.Instances[0])
	flat := map[string]any{}
	for _, member := range []string{"id", "version", "title", "description", "axes", "stresses", "profiles", "ground_truth"} {
		copyMember(flat, whole, member)
	}
	for _, member := range []string{"state", "dynamics", "channels", "faults", "effectors"} {
		copyMember(flat, kind, member)
	}
	entities := map[string]any{"entity_type": spec.Kinds[0].Name}
	for _, member := range []string{"id_template", "count", "churn"} {
		copyMember(entities, instances, member)
	}
	flat["entities"] = entities
	return flat
}

func generic(t *testing.T, value any) map[string]any {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := model.DecodeBytes(raw, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func copyMember(into, from map[string]any, member string) {
	if value, ok := from[member]; ok {
		into[member] = value
	}
}

func TestTheEquivalenceTableNamesEveryShippedDomain(t *testing.T) {
	t.Parallel()
	files, err := filepath.Glob(filepath.Join(testsupport.DomainsDir(), "*.domain.json"))
	if err != nil {
		t.Fatal(err)
	}
	var shippedIDs, tabled []string
	for _, file := range files {
		shippedIDs = append(shippedIDs, strings.TrimSuffix(filepath.Base(file), ".domain.json"))
	}
	for id := range v01RuntimeDigests {
		tabled = append(tabled, id)
	}
	slices.Sort(shippedIDs)
	slices.Sort(tabled)
	if !slices.Equal(shippedIDs, tabled) {
		t.Fatalf("shipped domains %v, equivalence table %v: a domain is unchecked or stale", shippedIDs, tabled)
	}
}
