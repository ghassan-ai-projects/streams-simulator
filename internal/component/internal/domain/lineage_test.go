package domain

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	specs "github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const zoneChoicesJSON = `{
  "schema_version": "0.1", "id": "zone-twin", "version": "0.1.0", "title": "Zone twin",
  "kinds": [{"kind": "zone", "instance": "zone-1", "metrics": {"temp": {"instrument": "acme/probe", "initial": 30}}}]
}`

func workspaceFile(name, raw string) NamedFile { return NamedFile{Name: name, Raw: []byte(raw)} }

func composedPlant(t *testing.T, library *Library, composition string) NamedFile {
	t.Helper()
	composed, err := ComposeDomain(library, []byte(composition), "test.composition.json")
	if err != nil {
		t.Fatal(err)
	}
	return NamedFile{Name: "test-plant.domain.json", Raw: composed.Bytes}
}

func twinPlant(t *testing.T, library *Library) NamedFile {
	t.Helper()
	composed, err := ScaffoldTwin(library, kindDocuments(t, zoneKindJSON, fanKindJSON), []byte(fanZoneChoicesJSON), "fan-zone.twin.json")
	if err != nil {
		t.Fatal(err)
	}
	return NamedFile{Name: "fan-zone-twin.domain.json", Raw: composed.Bytes}
}

func runJSON(id, digest string) string {
	return fmt.Sprintf(`{"run_id":"r","mode":"explore","domain":{"id":%q,"version":"0.1.0","digest":%q}}`, id, digest)
}

func lineageOf(t *testing.T, library *Library, workspace Workspace, only string) *Lineage {
	t.Helper()
	got, err := DescribeLineage(library, workspace, only)
	if err != nil {
		t.Fatalf("DescribeLineage = %v, want a document", err)
	}
	return got
}

func (l *Lineage) plant(t *testing.T, id string) PlantLineage {
	t.Helper()
	for _, plant := range l.Plants {
		if plant.Domain.ID == id {
			return plant
		}
	}
	t.Fatalf("no plant %q in %v", id, l.Plants)
	return PlantLineage{}
}

func (l *Lineage) deviceKind(t *testing.T, kind string) DeviceKindLineage {
	t.Helper()
	for _, entry := range l.DeviceKinds {
		if entry.Kind == kind {
			return entry
		}
	}
	t.Fatalf("no device kind %q in %v", kind, l.DeviceKinds)
	return DeviceKindLineage{}
}

func TestLineageOfACompositionBackedPlantIsEditableThroughItsCompositionWhileTheyAgree(t *testing.T) {
	t.Parallel()
	library := baseLibrary(t)
	workspace := Workspace{Domains: []NamedFile{composedPlant(t, library, compositionJSON)}, Compositions: []NamedFile{workspaceFile("test-plant.composition.json", compositionJSON)}}
	plant := lineageOf(t, library, workspace, "").plant(t, "test-plant")
	if plant.Composition == nil || plant.Composition.Status != SourceInSync || plant.Composition.File != "compositions/test-plant.composition.json" ||
		plant.Composition.ProducesDigest != plant.Domain.Digest || plant.Twin != nil {
		t.Fatalf("composition = %+v, want in sync with the domain digest %s", plant.Composition, plant.Domain.Digest)
	}
	if plant.Editable.Verdict != VerdictEditable || plant.Editable.Reason != ReasonComposition || plant.Editable.Route.Code != RouteEditComposition || plant.Domain.File != testsupport.WorkspaceDomain("test-plant") {
		t.Fatalf("editable = %+v file %s, want editable through the composition", plant.Editable, plant.Domain.File)
	}
}

func TestLineageDetectsACompositionThatNoLongerProducesTheDomainFile(t *testing.T) {
	t.Parallel()
	library := baseLibrary(t)
	edited := mutate(t, compositionJSON, func(doc map[string]any) { doc["version"] = "0.2.0" })
	broken := mutate(t, compositionJSON, func(doc map[string]any) {
		doc["equipment"] = []any{map[string]any{"use": "acme/none@1", "names": []any{"X-1"}}}
	})
	for name, tc := range map[string]struct {
		composition string
		status      string
		reason      bool
	}{"edited after composing": {edited, SourceOutOfSync, false}, "no longer composes": {broken, SourceDoesNotCompose, true}} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			workspace := Workspace{Domains: []NamedFile{composedPlant(t, library, compositionJSON)}, Compositions: []NamedFile{workspaceFile("test-plant.composition.json", tc.composition)}}
			plant := lineageOf(t, library, workspace, "").plant(t, "test-plant")
			if plant.Composition == nil || plant.Composition.Status != tc.status || (plant.Composition.Reason != "") != tc.reason {
				t.Fatalf("composition = %+v, want status %s", plant.Composition, tc.status)
			}
			if plant.Editable.Reason != ReasonCompositionOutOfSync || plant.Editable.Route.Code != RouteRecompose || plant.Editable.Verdict != VerdictEditable {
				t.Fatalf("editable = %+v, want editable through the composition, which must be composed again", plant.Editable)
			}
		})
	}
}

func TestLineageOfATwinPlantRoutesToTheDeviceEditorAndNamesTheKindFilesItImplements(t *testing.T) {
	t.Parallel()
	library := baseLibrary(t)
	workspace := Workspace{Domains: []NamedFile{twinPlant(t, library)}, Twins: []NamedFile{workspaceFile("fan-zone.twin.json", fanZoneChoicesJSON)},
		Kinds: []NamedFile{workspaceFile("zone.kind.json", zoneKindJSON), workspaceFile("fan.kind.json", fanKindJSON)}}
	plant := lineageOf(t, library, workspace, "").plant(t, "fan-zone-twin")
	if plant.Twin == nil || plant.Twin.Status != SourceInSync || plant.Twin.File != "twins/fan-zone.twin.json" || plant.Twin.Version != "0.1.0" {
		t.Fatalf("twin = %+v, want in sync", plant.Twin)
	}
	if plant.Editable.Verdict != VerdictReadOnly || plant.Editable.Reason != ReasonTwin || plant.Editable.Route.Code != RouteOpenDeviceBehavior {
		t.Fatalf("editable = %+v, want read only with the device editor route", plant.Editable)
	}
	if len(plant.DeviceKinds) != 2 || plant.DeviceKinds[0].Kind != "fan" || plant.DeviceKinds[0].File != "kinds/fan.kind.json" || plant.DeviceKinds[0].Status != KindFound ||
		!strings.HasPrefix(plant.DeviceKinds[0].Digest, "sha256:") {
		t.Fatalf("device kinds = %+v, want fan and zone found with digests", plant.DeviceKinds)
	}
}

func TestLineageSaysWhichTwinsIncludeADeviceKindAndWhichKindHasATwinOfItsOwn(t *testing.T) {
	t.Parallel()
	library := baseLibrary(t)
	workspace := Workspace{Domains: []NamedFile{twinPlant(t, library)},
		Twins: []NamedFile{workspaceFile("fan-zone.twin.json", fanZoneChoicesJSON), workspaceFile("zone.twin.json", zoneChoicesJSON)},
		Kinds: []NamedFile{workspaceFile("zone.kind.json", zoneKindJSON), workspaceFile("fan.kind.json", fanKindJSON)}}
	got := lineageOf(t, library, workspace, "")
	fan, zone := got.deviceKind(t, "fan"), got.deviceKind(t, "zone")
	if fan.HasOwnTwin || len(fan.TwinsIncluding) != 1 || fan.TwinsIncluding[0] != (TwinMembership{Twin: "fan-zone-twin", File: "twins/fan-zone.twin.json", Kinds: 2}) {
		t.Fatalf("fan = %+v, want included in the two-kind twin only, with no twin of its own", fan)
	}
	if !zone.HasOwnTwin || len(zone.TwinsIncluding) != 2 || zone.TwinsIncluding[1].Twin != "zone-twin" || !slices.Equal(fan.ImplementedBy, []string{"fan-zone-twin"}) {
		t.Fatalf("zone = %+v, want a twin of its own beside the shared one", zone)
	}
}

func TestLineageNamesTheKindFilesATwinNeedsButTheWorkspaceLacks(t *testing.T) {
	t.Parallel()
	library := baseLibrary(t)
	workspace := Workspace{Domains: []NamedFile{twinPlant(t, library)}, Twins: []NamedFile{workspaceFile("fan-zone.twin.json", fanZoneChoicesJSON)},
		Kinds: []NamedFile{workspaceFile("zone.kind.json", zoneKindJSON)}}
	plant := lineageOf(t, library, workspace, "").plant(t, "fan-zone-twin")
	if plant.Twin.Status != SourceKindFilesMissing || !strings.Contains(plant.Twin.Reason, "fan") || plant.Editable.Reason != ReasonTwinOutOfSync {
		t.Fatalf("twin = %+v editable = %+v, want kind_files_missing naming fan", plant.Twin, plant.Editable)
	}
	if plant.DeviceKinds[0].Kind != "fan" || plant.DeviceKinds[0].Status != KindMissing || plant.DeviceKinds[0].File != "" {
		t.Fatalf("device kinds = %+v, want fan missing", plant.DeviceKinds)
	}
}

func TestLineageNamesATwinThatNoLongerScaffoldsTheDomain(t *testing.T) {
	t.Parallel()
	library := baseLibrary(t)
	edited := strings.Replace(fanZoneChoicesJSON, `"gain": -0.02`, `"gain": -0.5`, 1)
	broken := strings.Replace(fanZoneChoicesJSON, `"gain": -0.02`, `"gain": 0`, 1)
	for name, tc := range map[string]struct{ choices, status string }{"edited": {edited, SourceOutOfSync}, "refused": {broken, SourceDoesNotScaffold}} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			workspace := Workspace{Domains: []NamedFile{twinPlant(t, library)}, Twins: []NamedFile{workspaceFile("fan-zone.twin.json", tc.choices)},
				Kinds: []NamedFile{workspaceFile("zone.kind.json", zoneKindJSON), workspaceFile("fan.kind.json", fanKindJSON)}}
			plant := lineageOf(t, library, workspace, "").plant(t, "fan-zone-twin")
			if plant.Twin.Status != tc.status || plant.Editable.Reason != ReasonTwinOutOfSync || plant.Editable.Route.Code != RouteRescaffoldTwin {
				t.Fatalf("twin = %+v editable = %+v, want status %s routed to scaffold the twin again", plant.Twin, plant.Editable, tc.status)
			}
		})
	}
}

func TestLineageOfAHandAuthoredDomainOffersACopy(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(testsupport.Domain("aquaculture-pond"))
	if err != nil {
		t.Fatal(err)
	}
	plant := lineageOf(t, baseLibrary(t), Workspace{Domains: []NamedFile{{Name: "aquaculture-pond.domain.json", Raw: raw}}}, "").plant(t, "aquaculture-pond")
	if plant.Composition != nil || plant.Twin != nil || plant.Editable.Verdict != VerdictReadOnly || plant.Editable.Reason != ReasonHandAuthored || plant.Editable.Route.Code != RouteCopyDomainFile {
		t.Fatalf("plant = %+v, want a read only hand authored domain", plant)
	}
}

func TestLineageRecognisesAShippedExampleOnlyWhileItIsTheExampleComposedNow(t *testing.T) {
	t.Parallel()
	library := baseLibrary(t)
	examples := []NamedFile{workspaceFile("test-plant.composition.json", compositionJSON)}
	same := Workspace{Domains: []NamedFile{composedPlant(t, library, compositionJSON)}, Examples: examples}
	edited := mutate(t, compositionJSON, func(doc map[string]any) { doc["version"] = "0.2.0" })
	other := Workspace{Domains: []NamedFile{composedPlant(t, library, edited)}, Examples: examples}
	if plant := lineageOf(t, library, same, "").plant(t, "test-plant"); plant.Editable.Reason != ReasonExample || plant.Editable.Route.Code != RouteCreateFromExample {
		t.Fatalf("editable = %+v, want the example route", plant.Editable)
	}
	if plant := lineageOf(t, library, other, "").plant(t, "test-plant"); plant.Editable.Reason != ReasonHandAuthored {
		t.Fatalf("editable = %+v, want hand authored once the file differs from the example", plant.Editable)
	}
}

func TestLineageListsTheRunsOfAPlantAndFlagsThoseThatUsedAnotherVersion(t *testing.T) {
	t.Parallel()
	library := baseLibrary(t)
	domain := composedPlant(t, library, compositionJSON)
	digest := plantDigest(t, domain)
	other := "sha256:" + strings.Repeat("a", 64)
	workspace := Workspace{Domains: []NamedFile{domain}, Runs: []RunFolder{
		{ID: "r-3", RunJSON: []byte(runJSON("test-plant", digest))},
		{ID: "r-1", RunJSON: []byte(runJSON("test-plant", other))},
		{ID: "r-2", Observer: []byte(`{"type":"run.opened","mode":"evaluate","domain":{"id":"test-plant","version":"0.1.0","digest":"` + digest + `"}}` + "\n{}")},
		{ID: "r-4", RunJSON: []byte(runJSON("another-plant", digest))},
	}}
	got := lineageOf(t, library, workspace, "")
	runs := got.plant(t, "test-plant").Runs
	want := []RunLineage{{ID: "r-1", Mode: "explore", Recorded: true, DomainVersion: "0.1.0", DomainDigest: other},
		{ID: "r-2", Mode: "evaluate", DomainVersion: "0.1.0", DomainDigest: digest, SameAsCurrent: true},
		{ID: "r-3", Mode: "explore", Recorded: true, DomainVersion: "0.1.0", DomainDigest: digest, SameAsCurrent: true}}
	if !slices.Equal(runs, want) || len(got.Unreadable) != 0 {
		t.Fatalf("runs = %+v unreadable = %+v, want %+v", runs, got.Unreadable, want)
	}
}

func plantDigest(t *testing.T, file NamedFile) string {
	t.Helper()
	compiled, err := specs.Parse(file.Raw, file.Name)
	if err != nil {
		t.Fatal(err)
	}
	return compiled.Digest
}
