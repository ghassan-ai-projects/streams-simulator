package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestIdenticalDomainsHaveNoChangesAndTheSameDigest(t *testing.T) {
	t.Parallel()
	raw := twinBytes(t, fanZoneChoicesJSON)
	got := diffOf(t, fileSide(raw, "a.domain.json"), fileSide(raw, "b.domain.json"))
	if len(got.Kinds) != 0 || len(got.DomainChanges) != 0 || got.Summary.Changes != 0 {
		t.Fatalf("diff = %v %v, want no change", got.Kinds, got.DomainChanges)
	}
	if !got.Identity.SameID || !got.Identity.SameVersion || !got.Identity.SameDigest || got.Identity.SameIDAndVersionDifferentDigest {
		t.Fatalf("identity = %+v, want the same model", got.Identity)
	}
	if got.Unchanged.Kinds != 2 || got.Unchanged.Readings == 0 || got.Unchanged.Commands == 0 || got.Unchanged.Links != 1 {
		t.Fatalf("unchanged = %+v, want every item counted as unchanged", got.Unchanged)
	}
}

func TestAChangedGainNamesTheAffectedRelationItsInstancesAndTheSilentEdit(t *testing.T) {
	t.Parallel()
	a, b := twinBytes(t, fanZoneChoicesJSON), twinBytes(t, gainChoices("-0.03"))
	got := diffOf(t, fileSide(a, "before.domain.json"), fileSide(b, "after.domain.json"))
	if len(got.Kinds) != 1 {
		t.Fatalf("kinds = %v, want only the driven kind", got.kindNames())
	}
	zone := got.kind(t, "zone")
	gain := zone.change(t, CodeGainChanged)
	if gain.Field != "coef" || gain.Before != -0.02 || gain.After != -0.03 || gain.Section != SectionDynamics ||
		gain.Subject != "zone/dynamics/temp_value/inputs/duty_value@fan_drives_zone" || !strings.Contains(gain.Sentence, "-0.02") || !strings.Contains(gain.Sentence, "-0.03") {
		t.Fatalf("gain change = %+v, want the coupling coefficient from -0.02 to -0.03", gain)
	}
	if zone.Status != StatusChanged || zone.Instances.After.Count != 1 || !slices.Equal(zone.Instances.After.Names, []string{"zone-1"}) {
		t.Fatalf("zone = %+v, want it changed with its instance zone-1", zone)
	}
	if !got.Identity.SameID || !got.Identity.SameVersion || got.Identity.SameDigest || !got.Identity.SameIDAndVersionDifferentDigest {
		t.Fatalf("identity = %+v, want the silent edit flagged: same id and version, another digest", got.Identity)
	}
	if want := []Relation{{Type: "fan_drives_zone", FromKind: "fan", ToKind: "zone", Links: []string{"fan-1>zone-1"}}}; !reflect.DeepEqual(zone.Relations, want) {
		t.Fatalf("relations = %+v, want the relation the changed coupling runs over: %+v", zone.Relations, want)
	}
}

func TestANewVersionIsNotASilentEdit(t *testing.T) {
	t.Parallel()
	a := twinBytes(t, fanZoneChoicesJSON)
	b := editedDomain(t, a, func(doc map[string]any) { doc["version"] = "0.2.0" })
	got := diffOf(t, fileSide(a, "a"), fileSide(b, "b"))
	if got.Identity.SameVersion || got.Identity.SameIDAndVersionDifferentDigest || !got.Identity.SameID || got.Identity.SameDigest {
		t.Fatalf("identity = %+v, want a new version with a new digest", got.Identity)
	}
}

func TestACommandResponseChangedFromFirstOrderToDirectRemovesTheDynamicsItDroveAndTheRelation(t *testing.T) {
	t.Parallel()
	direct := strings.Replace(fanZoneChoicesJSON, `"response": "first_order"`, `"response": "direct"`, 1)
	direct = strings.Replace(direct, `,
        "drives": [{"from": "duty", "kind": "zone", "metric": "temp", "gain": -0.02, "time_constant_s": 100}]`, ``, 1)
	got := diffOf(t, fileSide(twinBytes(t, fanZoneChoicesJSON), "a"), fileSide(twinBytes(t, direct), "b"))
	removed := got.kind(t, "zone").change(t, CodeDynamicsRemoved)
	if removed.After != nil || removed.Before == nil || removed.Subject != "zone/dynamics/temp_value" {
		t.Fatalf("change = %+v, want the zone dynamics, which the drive of the fan made, removed", removed)
	}
	if links := codesOf(got.DomainChanges); !slices.Contains(links, CodeLinkRemoved) {
		t.Fatalf("domain changes = %v, want the relation between fan and zone removed", links)
	}
}

func TestASwappedInstrumentIsReportedOnTheReadingItMeasures(t *testing.T) {
	t.Parallel()
	library := libraryWith(t, instrumentVariant(t, "acme/probe-k", "degC", func(doc map[string]any) { doc["range"] = map[string]any{"min": 0, "max": 100} }))
	composed := func(composition string) []byte {
		out, err := ComposeDomain(library, []byte(composition), "test.composition.json")
		if err != nil {
			t.Fatal(err)
		}
		return out.Bytes
	}
	got := diffOf(t, fileSide(composed(compositionJSON), "a"), fileSide(composed(withEquipment(t, replacing("temp", "acme/probe-k@1"))), "b"))
	kiln := got.kind(t, "kiln")
	swapped := kiln.change(t, CodeInstrumentChanged)
	if swapped.Subject != "kiln/readings/kiln.temp" || swapped.Field != "instrument" || !strings.Contains(string(mustJSON(t, swapped.Before)), "acme/probe@1") ||
		!strings.Contains(string(mustJSON(t, swapped.After)), "acme/probe-k@1") {
		t.Fatalf("instrument change = %+v, want probe@1 to probe-k@1 on kiln.temp", swapped)
	}
	if rng := kiln.change(t, CodeRangeChanged); rng.Field != "range" {
		t.Fatalf("range change = %+v", rng)
	}
}

func TestAKindAddedAndAKindRemovedAreEachOneGroupWithItsInstances(t *testing.T) {
	t.Parallel()
	a := twinBytes(t, fanZoneChoicesJSON)
	b := editedDomain(t, a, func(doc map[string]any) { addKindCopy(doc, "fan", "fan_b", "fan-2") })
	added := diffOf(t, fileSide(a, "a"), fileSide(b, "b")).kind(t, "fan_b")
	if added.Status != StatusAdded || !slices.Equal(codesOf(added.Changes), []string{CodeKindAdded}) || added.Instances.After.Count != 1 || added.Instances.Before.Count != 0 {
		t.Fatalf("added = %+v, want one kind_added change and the instance fan-2", added)
	}
	removed := diffOf(t, fileSide(b, "b"), fileSide(a, "a")).kind(t, "fan_b")
	if removed.Status != StatusRemoved || !slices.Equal(codesOf(removed.Changes), []string{CodeKindRemoved}) || removed.Instances.Before.Count != 1 {
		t.Fatalf("removed = %+v, want one kind_removed change", removed)
	}
}

func TestACommandBoundTheAcknowledgementAndSafeStopAreComparedField(t *testing.T) {
	t.Parallel()
	a := twinBytes(t, fanZoneChoicesJSON)
	b := editedDomain(t, a, func(doc map[string]any) {
		effector := effectorOf(doc, "fan", "set_duty")
		effector["risk_class"] = "high"
		effector["ack"] = map[string]any{"failure_modes": []any{map[string]any{"mode": "ok", "probability": 1}}, "latency_ms": map[string]any{"mean": 90, "sigma": 10}}
		properties := effector["args_schema"].(map[string]any)["properties"].(map[string]any)
		properties["duty"].(map[string]any)["maximum"] = 300
	})
	fan := diffOf(t, fileSide(a, "a"), fileSide(b, "b")).kind(t, "fan")
	for _, code := range []string{CodeRiskClassChanged, CodeLatencyChanged, CodeArgumentChanged} {
		fan.change(t, code)
	}
	if bound := fan.change(t, CodeArgumentChanged); bound.Field != "maximum" || bound.Before != 600.0 || bound.After != 300.0 || !strings.HasSuffix(bound.Subject, "/arguments/duty") {
		t.Fatalf("bound = %+v, want the maximum of duty from 600 to 300", bound)
	}
}

func TestAChangeTheViewDoesNotModelIsStillReportedAsDefinitionAndNeverHiddenByTheDigest(t *testing.T) {
	t.Parallel()
	a := twinBytes(t, fanZoneChoicesJSON)
	b := editedDomain(t, a, func(doc map[string]any) {
		doc["stresses"] = "another stress statement for the plant that is longer than forty characters"
		kind := kindOf(doc, "zone")
		kind["dynamics"].([]any)[0].(map[string]any)["dt_ms"] = 500
	})
	got := diffOf(t, fileSide(a, "a"), fileSide(b, "b"))
	if section := got.DomainChanges; len(section) != 1 || section[0].Code != CodeSectionChanged || section[0].Field != "stresses" {
		t.Fatalf("domain changes = %+v, want the stress statement named", section)
	}
	if definition := got.kind(t, "zone").change(t, CodeDynamicsChanged); definition.Field != "definition" {
		t.Fatalf("dynamics change = %+v, want the raw declaration under definition", definition)
	}
}

func TestARunFolderIsComparedByTheDomainItRecordedAndNamesItsSource(t *testing.T) {
	t.Parallel()
	older, current := twinBytes(t, fanZoneChoicesJSON), twinBytes(t, gainChoices("-0.03"))
	got := diffOf(t, runSide(t, older, "runs/w-1"), fileSide(current, testsupport.WorkspaceDomain("fan-zone-twin")))
	if got.A.Source != (DiffSource{Kind: SourceRun, Label: "runs/w-1"}) || got.B.Source != (DiffSource{Kind: SourceFile, Label: testsupport.WorkspaceDomain("fan-zone-twin")}) {
		t.Fatalf("sources = %+v %+v", got.A.Source, got.B.Source)
	}
	if !got.Identity.SameIDAndVersionDifferentDigest {
		t.Fatalf("identity = %+v, want the run to have used another model than the file under the same version", got.Identity)
	}
	got.kind(t, "zone").change(t, CodeGainChanged)
}

func TestADiffIsTheSameEveryTimeAndValidatesAgainstItsContract(t *testing.T) {
	t.Parallel()
	a, b := twinBytes(t, fanZoneChoicesJSON), twinBytes(t, gainChoices("-0.03"))
	first, second := diffOf(t, fileSide(a, "a"), fileSide(b, "b")), diffOf(t, fileSide(a, "a"), fileSide(b, "b"))
	one, two := mustJSON(t, first), mustJSON(t, second)
	if !bytes.Equal(one, two) {
		t.Fatal("two diffs of the same inputs differ")
	}
	if err := model.ValidateDomainDiff(one); err != nil {
		t.Fatalf("ValidateDomainDiff = %v", err)
	}
	reversed := mustJSON(t, diffOf(t, fileSide(b, "b"), fileSide(a, "a")))
	if err := model.ValidateDomainDiff(reversed); err != nil {
		t.Fatalf("ValidateDomainDiff of the reverse = %v", err)
	}
}

func TestADiffRefusesWhatItCannotCompareWithTheCodeOfTheContractItBreaks(t *testing.T) {
	t.Parallel()
	good := twinBytes(t, fanZoneChoicesJSON)
	noSpec := readRunTemplate(t)
	delete(noSpec, "domain_spec")
	noSpecRaw := mustJSON(t, noSpec)
	for name, tc := range map[string]struct {
		input    DiffInput
		contract model.Contract
	}{
		"a file that is not JSON":        {fileSide([]byte(`{`), "bad.domain.json"), model.ContractDomainSpec},
		"a file that breaks the schema":  {fileSide([]byte(`{"id":"X"}`), "bad.domain.json"), model.ContractDomainSpec},
		"a run.json that is not JSON":    {DiffInput{Raw: []byte(`{`), Label: "runs/r", Run: true}, model.ContractRunArtifact},
		"a run.json without domain_spec": {DiffInput{Raw: noSpecRaw, Label: "runs/r", Run: true}, model.ContractRunArtifact},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			for _, pair := range [][2]DiffInput{{tc.input, fileSide(good, "g")}, {fileSide(good, "g"), tc.input}} {
				_, err := DescribeDiff(pair[0], pair[1])
				if invalid, ok := model.AsValidationError(err); !ok || invalid.Contract != tc.contract || !strings.Contains(err.Error(), tc.input.Label) {
					t.Fatalf("DescribeDiff = %v, want a violation of %s naming %s", err, tc.contract, tc.input.Label)
				}
			}
		})
	}
}

func TestARunWhoseRecordedDigestIsNotItsDomainSpecIsRefused(t *testing.T) {
	t.Parallel()
	good := twinBytes(t, fanZoneChoicesJSON)
	side := runSide(t, good, "runs/r")
	var run map[string]any
	if err := json.Unmarshal(side.Raw, &run); err != nil {
		t.Fatal(err)
	}
	run["domain"].(map[string]any)["digest"] = digestSample
	side.Raw = mustJSON(t, run)
	_, err := DescribeDiff(side, fileSide(good, "g"))
	if invalid, ok := model.AsValidationError(err); !ok || invalid.Contract != model.ContractRunArtifact || !errors.Is(err, ErrRunDomainDigest) {
		t.Fatalf("DescribeDiff = %v, want ErrRunDomainDigest as an artifact violation", err)
	}
}
