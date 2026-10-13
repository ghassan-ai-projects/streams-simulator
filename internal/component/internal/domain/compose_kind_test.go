package domain

import (
	"strings"
	"testing"
)

func TestComposeWritesOneKindWithQualifiedNamesAndAttachedChannels(t *testing.T) {
	t.Parallel()
	doc := composeOK(t, baseLibrary(t), withEquipment(t, kilnEntry(map[string]any{"as": "kiln_a"})))
	kind := member(t, doc, "kinds", 0)
	if got := member(t, kind, "name"); got != "kiln_a" {
		t.Fatalf("kind name = %v, want kiln_a", got)
	}
	channels := member(t, kind, "channels").([]any)
	if len(channels) != 2 {
		t.Fatalf("channels = %d, want the two non-optional points", len(channels))
	}
	channel := named(t, channels, "channel", "name", "kiln_a.temp")
	for member_, want := range map[string]string{"observes": "temp", "unit": "degC", "tag": "TT"} {
		if got := member(t, channel, member_); got != want {
			t.Errorf("channel %s = %v, want %s", member_, got, want)
		}
	}
	if got := member(t, channel, "observation_bias", "state"); got != "temp__bias" {
		t.Errorf("observation_bias state = %v, want temp__bias", got)
	}
	named(t, member(t, kind, "faults"), "fault", "id", "kiln_a__fouling")
	named(t, member(t, kind, "faults"), "fault", "id", "kiln_a__temp__offset")
	named(t, member(t, kind, "effectors"), "effector", "name", "kiln_a__set_setpoint")
	if got := member(t, kind, "faults", 0, "observability", "detector", "channel"); got != "kiln_a.temp" {
		t.Errorf("detector channel = %v, want kiln_a.temp", got)
	}
}

func TestComposeNamesTheKindAfterTheComponentWhenNoAsIsGiven(t *testing.T) {
	t.Parallel()
	doc := composeOK(t, baseLibrary(t), withEquipment(t, kilnEntry(nil)))
	if got := member(t, doc, "kinds", 0, "name"); got != "kiln" {
		t.Fatalf("kind name = %v, want kiln", got)
	}
}

func TestComposeKeepsLibraryDescriptionsOutOfTheDomain(t *testing.T) {
	t.Parallel()
	composed, err := ComposeDomain(baseLibrary(t), []byte(withEquipment(t, kilnEntry(nil))), "test.composition.json")
	var out []byte
	if composed != nil {
		out = composed.Bytes
	}
	if err != nil {
		t.Fatal(err)
	}
	for _, leaked := range []string{"library text that must not leak", "state text", "fault text", "effector text", "args title text", "comment text", "nested arg text", "example text"} {
		if strings.Contains(string(out), leaked) {
			t.Errorf("the composed domain carries library text %q", leaked)
		}
	}
	if !strings.Contains(string(out), "A composition for tests.") {
		t.Error("the composition's own description is missing")
	}
}

func TestComposeKeepsTheArgumentSchemaOfAnEffectorWithoutItsAnnotationsAtAnyDepth(t *testing.T) {
	t.Parallel()
	doc := composeOK(t, baseLibrary(t), withEquipment(t, kilnEntry(nil)))
	schema := member(t, doc, "kinds", 0, "effectors", 0, "args_schema").(map[string]any)
	setpoint := member(t, schema, "properties", "setpoint").(map[string]any)
	if len(schema) != 4 || number(t, setpoint["maximum"]) != 200 || len(setpoint) != 3 {
		t.Fatalf("args_schema = %v, want type, additionalProperties, required and properties with the bounds kept and the annotations gone", schema)
	}
	if _, annotated := member(t, schema, "properties", "kiln_id").(map[string]any)["$comment"]; annotated {
		t.Error("a nested $comment reached the composed domain")
	}
}

func TestComposeAddsABiasStateForEveryPointWhoseInstrumentActsThroughOne(t *testing.T) {
	t.Parallel()
	doc := composeOK(t, baseLibrary(t), withEquipment(t, kilnEntry(nil)))
	states := member(t, doc, "kinds", 0, "state")
	for _, name := range []string{"temp__bias", "power__bias"} {
		bias := named(t, states, "state", "name", name)
		if bias["unit"] == nil || number(t, bias["initial"]) != 0 {
			t.Errorf("bias state %s = %v, want a unit and initial 0", name, bias)
		}
	}
	if got := len(states.([]any)); got != 7 {
		t.Errorf("states = %d, want the five of the kiln and two biases", got)
	}
}

func TestComposeWritesOnlyTheMembersTheComponentUses(t *testing.T) {
	t.Parallel()
	doc := composeOK(t, baseLibrary(t), withEquipment(t, kilnEntry(nil)))
	kind := member(t, doc, "kinds", 0).(map[string]any)
	for _, absent := range []string{"flows", "boundaries", "system"} {
		if _, present := kind[absent]; present {
			t.Errorf("kind carries %s, which the kiln does not declare", absent)
		}
	}
	if _, present := doc["site"]; present {
		t.Error("a composition without a site wrote one")
	}
	if _, present := doc["links"]; present {
		t.Error("a composition without connections wrote links")
	}
}

func TestComposeRejectsAnUnknownReference(t *testing.T) {
	t.Parallel()
	_, err := compose(t, baseLibrary(t), withEquipment(t, kilnEntry(map[string]any{"use": "acme/forge@1"})))
	wantIssue(t, err, "composition-v0.1", "/equipment/0/use", "acme/forge@1 is not an equipment")
}

func TestComposeRejectsAnInstrumentUsedAsEquipment(t *testing.T) {
	t.Parallel()
	_, err := compose(t, baseLibrary(t), withEquipment(t, kilnEntry(map[string]any{"use": "acme/probe@1"})))
	wantIssue(t, err, "composition-v0.1", "/equipment/0/use", "not an equipment")
}

func TestComposeRejectsTwoEntriesWithOneKindName(t *testing.T) {
	t.Parallel()
	second := kilnEntry(map[string]any{"names": []any{"K-2"}})
	_, err := compose(t, baseLibrary(t), withEquipment(t, kilnEntry(nil), second))
	wantIssue(t, err, "composition-v0.1", "/equipment/1/as", `kind name "kiln" is used by another entry`)
}

func TestComposeRejectsADefaultKindNameBeyondTheLimit(t *testing.T) {
	t.Parallel()
	long := mutate(t, kilnJSON, func(doc map[string]any) { doc["id"] = "acme/a-very-long-component-name" })
	files := baseFiles()
	files.equipment = []File{file("equipment/long.component.json", long)}
	_, err := compose(t, loadLibrary(t, files.directory("lib")), withEquipment(t, kilnEntry(map[string]any{"use": "acme/a-very-long-component-name@1"})))
	wantIssue(t, err, "composition-v0.1", "/equipment/0/as", "at most 20 characters")
}

func TestComposeReportsAnInvalidComposedDomainAsADomainViolation(t *testing.T) {
	t.Parallel()
	raw := mutate(t, withEquipment(t, kilnEntry(nil)), func(doc map[string]any) {
		doc["profiles"] = []any{map[string]any{"name": "nominal", "duration_s": 3600},
			map[string]any{"name": "sensor_pathology", "duration_s": 3600, "fault_weights": map[string]any{"kiln__no_such_fault": 1}},
			map[string]any{"name": "correlated_cascade", "not_applicable": "the test plant has no cascade at all"}}
	})
	_, err := compose(t, baseLibrary(t), raw)
	invalid := asViolation(t, err, "domain-spec-v0.2")
	if !strings.Contains(err.Error(), "test.composition.json") || len(invalid.Issues) == 0 {
		t.Fatalf("error = %v, want the domain issues naming the composition", err)
	}
}

func TestComposeRejectsACompositionThatBreaksItsSchema(t *testing.T) {
	t.Parallel()
	_, err := compose(t, baseLibrary(t), `{"schema_version": "0.1"}`)
	wantViolation(t, err, "composition-v0.1")
	_, err = compose(t, baseLibrary(t), `not json`)
	wantViolation(t, err, "composition-v0.1")
}

func TestComposeSummarisesTheDomainItComposed(t *testing.T) {
	t.Parallel()
	composed, err := ComposeDomain(baseLibrary(t), []byte(withEquipment(t, kilnEntry(nil))), "test.composition.json")
	if err != nil {
		t.Fatal(err)
	}
	if composed.Summary.ID != composed.Domain.Spec.ID || composed.Summary.Digest != composed.Domain.Digest || composed.Summary.Channels == 0 {
		t.Fatalf("summary = %+v, want the composed domain's identity and counts", composed.Summary)
	}
}
