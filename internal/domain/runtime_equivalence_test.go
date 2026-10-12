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
			raw, err := json.Marshal(shipped(t, id).Spec)
			if err != nil {
				t.Fatal(err)
			}
			var doc any
			if err := model.DecodeBytes(raw, &doc); err != nil {
				t.Fatal(err)
			}
			got, err := canonical.Digest(doc)
			if err != nil || got != want {
				t.Fatalf("runtime spec digest = %s (%v), want the v0.1 runtime %s", got, err, want)
			}
		})
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
