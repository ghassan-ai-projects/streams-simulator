package domain

import (
	"bytes"
	"reflect"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
)

func sited(t *testing.T, change func(doc map[string]any)) string {
	t.Helper()
	return mutate(t, withEquipment(t, kilnEntry(map[string]any{"area": "Hall"})), func(doc map[string]any) {
		doc["site"] = map[string]any{"name": "Plant", "timezone": "Europe/Berlin"}
		doc["areas"] = []any{map[string]any{"name": "Works"}, map[string]any{"name": "Hall", "parent": "Works"}}
		change(doc)
	})
}

func TestComposeCopiesTheSiteAreasAndTheEntryArea(t *testing.T) {
	t.Parallel()
	doc := composeOK(t, baseLibrary(t), sited(t, func(map[string]any) {}))
	if member(t, doc, "site", "name") != "Plant" || member(t, doc, "site", "timezone") != "Europe/Berlin" {
		t.Errorf("site = %v", doc["site"])
	}
	hall := named(t, member(t, doc, "site", "areas"), "area", "name", "Hall")
	if hall["parent"] != "Works" {
		t.Errorf("Hall = %v, want parent Works", hall)
	}
	if got := member(t, doc, "instances", 0, "area"); got != "Hall" {
		t.Errorf("instance area = %v, want Hall", got)
	}
}

func TestComposeRefusesAreasThatDoNotHoldTogether(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, pointer, message string
		change                 func(doc map[string]any)
	}{
		{"an entry area that is not declared", "/equipment/0/area", `area "Hall" is not declared`,
			func(doc map[string]any) { doc["areas"] = []any{map[string]any{"name": "Works"}} }},
		{"an area declared twice", "/areas/1/name", `area "Works" is declared twice`,
			func(doc map[string]any) {
				doc["areas"] = []any{map[string]any{"name": "Works"}, map[string]any{"name": "Works"}, map[string]any{"name": "Hall"}}
			}},
		{"a parent declared later", "/areas/0/parent", `parent "Hall" must be declared before area "Works"`,
			func(doc map[string]any) {
				doc["areas"] = []any{map[string]any{"name": "Works", "parent": "Hall"}, map[string]any{"name": "Hall"}}
			}},
		{"areas without a site", "/areas", "areas need a site", func(doc map[string]any) { delete(doc, "site") }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := compose(t, baseLibrary(t), sited(t, tc.change))
			wantIssue(t, err, "composition-v0.1", tc.pointer, tc.message)
		})
	}
}

func TestComposeRecordsTheDigestsOfTheFilesItCameFrom(t *testing.T) {
	t.Parallel()
	library := baseLibrary(t)
	doc := composeOK(t, library, withEquipment(t, kilnEntry(nil)))
	provenance := member(t, doc, "instances", 0, "provenance")
	want, err := canonical.Digest(library.equipment["acme/kiln@1"].Document)
	if err != nil {
		t.Fatal(err)
	}
	if member(t, provenance, "digest") != want || member(t, provenance, "component") != "acme/kiln" || number(t, member(t, provenance, "version")) != 1 {
		t.Errorf("provenance = %v, want component acme/kiln@1 with digest %s", provenance, want)
	}
	for _, ref := range []string{"acme/probe@1", "acme/meter@1"} {
		digest, err := canonical.Digest(library.instruments[ref].Document)
		if err != nil || member(t, provenance, "instrument_digests", ref) != digest {
			t.Errorf("instrument digest of %s = %v, want %s (%v)", ref, member(t, provenance, "instrument_digests", ref), digest, err)
		}
	}
}

func TestComposeDigestsDependOnTheBytesOfTheLibraryFile(t *testing.T) {
	t.Parallel()
	changed := mutate(t, kilnJSON, func(doc map[string]any) { doc["title"] = "Another kiln" })
	files := baseFiles()
	files.equipment = []File{file("equipment/kiln.component.json", changed)}
	other := composeOK(t, loadLibrary(t, files.directory("lib")), withEquipment(t, kilnEntry(nil)))
	base := composeOK(t, baseLibrary(t), withEquipment(t, kilnEntry(nil)))
	if member(t, other, "instances", 0, "provenance", "digest") == member(t, base, "instances", 0, "provenance", "digest") {
		t.Fatal("two different files recorded one digest")
	}
}

func composeBytes(t *testing.T, library *Library) []byte {
	t.Helper()
	entries := []map[string]any{kilnEntry(map[string]any{"preset": "big"}), kilnEntry(map[string]any{"as": "kiln_b", "names": []any{"K-2"}})}
	composed, err := ComposeDomain(library, []byte(withEquipment(t, entries...)), "test.composition.json")
	var out []byte
	if composed != nil {
		out = composed.Bytes
	}
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestComposeGivesTheSameBytesEveryTime(t *testing.T) {
	t.Parallel()
	library := baseLibrary(t)
	first, second := composeBytes(t, library), composeBytes(t, library)
	if !bytes.Equal(first, second) {
		t.Fatal("two composes of one composition differ")
	}
}

func TestComposeDoesNotDependOnTheOrderFilesAreListed(t *testing.T) {
	t.Parallel()
	forward := baseFiles()
	backward := baseFiles()
	slices.Reverse(backward.instruments)
	extra := file("equipment/another.component.json", mutate(t, kilnJSON, func(doc map[string]any) { doc["id"] = "acme/another" }))
	forward.equipment = append(forward.equipment, extra)
	backward.equipment = []File{extra, backward.equipment[0]}
	a, b := loadLibrary(t, forward.directory("lib")), loadLibrary(t, backward.directory("lib"))
	if !bytes.Equal(composeBytes(t, a), composeBytes(t, b)) {
		t.Fatal("the composed domain depends on the order the files were read")
	}
	if !reflect.DeepEqual(a.List(), b.List()) {
		t.Fatalf("listings differ:\n%v\n%v", a.List(), b.List())
	}
}

func TestQualifyingNamesPrefixesEveryReferenceTheDomainResolves(t *testing.T) {
	t.Parallel()
	kind := map[string]any{
		"faults": []any{map[string]any{"id": "split", "expected_effector": "reset",
			"observability": map[string]any{"detector": map[string]any{"channel_a": "left", "channel_b": "right"}}},
			map[string]any{"id": "frozen", "observation": map[string]any{"channel": "left"}}},
		"effectors": []any{map[string]any{"name": "reset", "confirmation_channels": []any{"left", "right"}}},
	}
	qualifyFaults(kind, "unit")
	qualifyEffectors(kind, "unit")
	faults := kind["faults"].([]any)
	detector := member(t, faults[0], "observability", "detector")
	if member(t, faults[0], "id") != "unit__split" || member(t, faults[0], "expected_effector") != "unit__reset" ||
		member(t, detector, "channel_a") != "unit.left" || member(t, detector, "channel_b") != "unit.right" {
		t.Errorf("fault = %v, want qualified id, effector and channels", faults[0])
	}
	if member(t, faults[1], "observation", "channel") != "unit.left" {
		t.Errorf("observation = %v, want unit.left", faults[1])
	}
	effector := kind["effectors"].([]any)[0]
	if member(t, effector, "name") != "unit__reset" || member(t, effector, "confirmation_channels", 1) != "unit.right" {
		t.Errorf("effector = %v, want qualified name and confirmation channels", effector)
	}
}
