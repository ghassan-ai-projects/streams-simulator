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
	"aquaculture-pond":      "sha256:07239180fb86a73b75393acaf02c6618972f0927de6ed6a65b325868cafe7908",
	"cold-chain-transit":    "sha256:527b81cf10a6e8dd3a5f19d199efc415780c7bea071ef6d2a6707ba84eb2cbfe",
	"discrete-line-oee":     "sha256:c5312b22db603334c249c1e30313cae0ac71d42a929c36365a1c9fa95a1d644b",
	"host-system-health":    "sha256:da36450f72d9e46cdcf0999b9f38a2704f9535d9524b0c0c03bdd50153e2904c",
	"k8s-cluster":           "sha256:f823d642be85b19560d02bcda39ab435d087016e296f1bf6b1e211a3bd8388ab",
	"open-field-irrigation": "sha256:936ae49fc4c7b0f182a5368b642e764127f2541b4d40ad8ecf7bf8c1f5060121",
	"rotating-machinery":    "sha256:a871d6ec8ca644afcfd44946238cdb1617684afe58b4a6e68c573018470e0683",
	"shipment-exception":    "sha256:930a4750e526700d1c1346c61f0da4c8161848dbadf8d70587941c497f89e8e8",
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
