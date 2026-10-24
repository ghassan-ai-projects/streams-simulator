package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func assumptionsOf(t *testing.T, lib *Library, input AssumptionsInput) *Assumptions {
	t.Helper()
	got, err := DescribeAssumptions(lib, input)
	if err != nil {
		t.Fatalf("DescribeAssumptions = %v, want a document", err)
	}
	return got
}

func fanZoneInput(t *testing.T, change func(choices string) string) (*Library, AssumptionsInput) {
	t.Helper()
	library, kinds := baseLibrary(t), kindDocuments(t, zoneKindJSON, fanKindJSON)
	choices := change(fanZoneChoicesJSON)
	composed, err := ScaffoldTwin(library, kinds, []byte(fanZoneChoicesJSON), "fan-zone.twin.json")
	if err != nil {
		t.Fatal(err)
	}
	return library, AssumptionsInput{Domain: composed.Bytes, Source: "fan-zone.domain.json",
		Twin: &TwinFiles{Choices: []byte(choices), Source: "fan-zone.twin.json", Kinds: kinds}}
}

func (a *Assumptions) kind(t *testing.T, name string) KindAssumptions {
	t.Helper()
	for _, kind := range a.Kinds {
		if kind.Kind == name {
			return kind
		}
	}
	t.Fatalf("no kind %q in %v", name, a.Kinds)
	return KindAssumptions{}
}

func (k KindAssumptions) reading(t *testing.T, channel string) ReadingAssumption {
	t.Helper()
	for _, reading := range k.Readings {
		if reading.Channel == channel {
			return reading
		}
	}
	t.Fatalf("no reading %q in kind %s", channel, k.Kind)
	return ReadingAssumption{}
}

func (k KindAssumptions) command(t *testing.T, name string) CommandAssumption {
	t.Helper()
	for _, command := range k.Commands {
		if command.Name == name {
			return command
		}
	}
	t.Fatalf("no command %q in kind %s", name, k.Kind)
	return CommandAssumption{}
}

func (a *Assumptions) codes(scope string) []string {
	var codes []string
	for _, entry := range a.NotModeled {
		if entry.Scope == scope {
			codes = append(codes, entry.Code)
		}
	}
	return codes
}

func TestAssumptionsOfATwinNameTheInstrumentAndWhereEachChoiceIsRecorded(t *testing.T) {
	t.Parallel()
	library, input := fanZoneInput(t, func(choices string) string { return choices })
	got := assumptionsOf(t, library, input)
	if got.Origin.Kind != OriginTwin || got.Origin.Twin == nil || got.Origin.Twin.Status != StatusInSync || got.Origin.Twin.ID != "fan-zone-twin" {
		t.Fatalf("origin = %+v, want a twin in sync", got.Origin)
	}
	temp := got.kind(t, "zone").reading(t, "temp")
	if temp.Basis != BasisLibraryInstrument || temp.RecordedIn != RecordedTwinChoices || temp.Instrument == nil || temp.Instrument.Ref != "acme/probe@1" ||
		temp.Instrument.Version != 1 || temp.Instrument.Digest == "" || temp.Tag != "TT" || temp.Noise.Sigma != 0.05 || temp.Initial == nil || *temp.Initial != 30 {
		t.Fatalf("zone temp = %+v, want the acme probe recorded in the twin choices with its noise, tag and initial value", temp)
	}
	if temp.EqualsBinarySuggestion == nil || !*temp.EqualsBinarySuggestion {
		t.Fatalf("zone temp suggestion = %v, want the choice to equal the binary's suggestion", temp.EqualsBinarySuggestion)
	}
	duty := got.kind(t, "fan").reading(t, "duty")
	if duty.Basis != BasisPlainChannel || duty.Instrument != nil || duty.RecordedIn != RecordedTwinChoices || duty.Noise.Sigma != 0 {
		t.Fatalf("fan duty = %+v, want a plain channel without noise recorded in the twin choices", duty)
	}
}

func TestAssumptionsOfATwinStateEachResponseWithItsGainAndTimeConstant(t *testing.T) {
	t.Parallel()
	library, input := fanZoneInput(t, func(choices string) string { return choices })
	fan := assumptionsOf(t, library, input).kind(t, "fan")
	set := fan.command(t, "set_duty")
	if set.Response != ResponseFirstOrder || set.RecordedIn != RecordedTwinChoices || len(set.Drives) != 1 || set.Drives[0].Gain != -0.02 ||
		set.Drives[0].TimeConstantS != 100 || set.Drives[0].Kind != "zone" || set.Drives[0].Metric != "temp" {
		t.Fatalf("set_duty = %+v, want a first order response of gain -0.02 and time constant 100 s on zone temp", set)
	}
	if set.EqualsBinarySuggestion == nil || *set.EqualsBinarySuggestion || !strings.Contains(set.Statement, "100 s") {
		t.Fatalf("set_duty = %+v, want a statement with the time constant and a choice that differs from the suggested direct response", set)
	}
	if len(set.Sets) != 1 || set.Sets[0].Argument != "duty" || set.Sets[0].State != "duty_value" || set.Lease == nil || set.Lease.Argument != "lease_ms" {
		t.Fatalf("set_duty sets = %+v lease = %+v, want the duty argument on duty_value under a lease_ms lease", set.Sets, set.Lease)
	}
	stop := fan.command(t, "stop_fan")
	if stop.Response != ResponseFixedSafeStop || !stop.SafeStop || stop.EqualsBinarySuggestion != nil {
		t.Fatalf("stop_fan = %+v, want the fixed safe stop with no suggestion to compare", stop)
	}
}

func TestAssumptionsOfATwinListWhatItDoesNotReproduceWithStableCodes(t *testing.T) {
	t.Parallel()
	library, input := fanZoneInput(t, func(choices string) string { return choices })
	got := assumptionsOf(t, library, input)
	for scope, want := range map[string][]string{
		"domain":       {"kind_file_members_not_reproduced", "first_order_response_only", "device_behavior_not_modeled", "instrument_semantics_unchecked", "no_calibration_evidence"},
		"fan.duty":     {"no_measurement_noise"},
		"fan.stop_fan": {"safe_stop_is_a_step", "acknowledgement_never_fails"},
		"fan.set_duty": {"acknowledgement_never_fails"},
		"zone.temp":    nil,
	} {
		if have := got.codes(scope); !slices.Equal(have, want) {
			t.Errorf("not_modeled codes of %s = %v, want %v", scope, have, want)
		}
	}
}

func TestAssumptionsOfATwinWhoseChoicesNoLongerProduceTheDomainTakeNothingFromThem(t *testing.T) {
	t.Parallel()
	library, input := fanZoneInput(t, func(choices string) string {
		return strings.Replace(choices, `"version": "0.1.0"`, `"version": "0.2.0"`, 1)
	})
	got := assumptionsOf(t, library, input)
	if got.Origin.Kind != OriginDomain || got.Origin.Twin == nil || got.Origin.Twin.Status != StatusOutOfSync || got.Origin.Twin.Version != "0.2.0" {
		t.Fatalf("origin = %+v, want the domain with an out of sync twin of version 0.2.0", got.Origin)
	}
	temp := got.kind(t, "zone").reading(t, "temp")
	if temp.Basis != BasisDomainDeclared || temp.Instrument != nil || temp.RecordedIn != RecordedDomainFile || temp.EqualsBinarySuggestion != nil {
		t.Fatalf("zone temp = %+v, want only what the domain file declares", temp)
	}
	if set := got.kind(t, "fan").command(t, "set_duty"); set.Response != ResponseDeclared || len(set.Drives) != 0 {
		t.Fatalf("set_duty = %+v, want a declared response and no drives from the stale choices", set)
	}
}

func TestAssumptionsOfAComposedDomainNameTheInstrumentsItsInstancesRecord(t *testing.T) {
	t.Parallel()
	library := baseLibrary(t)
	composed, err := ComposeDomain(library, []byte(compositionJSON), "test.composition.json")
	if err != nil {
		t.Fatal(err)
	}
	got := assumptionsOf(t, library, AssumptionsInput{Domain: composed.Bytes, Source: "test.domain.json"})
	if got.Origin.Kind != OriginComposition || got.Origin.Twin != nil {
		t.Fatalf("origin = %+v, want a composition", got.Origin)
	}
	temp := got.kind(t, "kiln").reading(t, "kiln.temp")
	if temp.Basis != BasisLibraryInstrument || temp.RecordedIn != RecordedComposition || temp.Instrument == nil || temp.Instrument.Ref != "acme/probe@1" ||
		temp.Instrument.Digest == "" || temp.Range == nil || temp.Cadence.PeriodS == nil || *temp.Cadence.PeriodS != 10 {
		t.Fatalf("kiln temp = %+v, want the probe recorded by the composition with its range and cadence", temp)
	}
	if !slices.Equal(got.codes("kiln.temp"), []string{"saturation_not_declared"}) {
		t.Fatalf("kiln.temp codes = %v, want saturation_not_declared: the dynamics declare no clamp", got.codes("kiln.temp"))
	}
	command := got.kind(t, "kiln").command(t, "kiln__set_setpoint")
	if command.Response != ResponseDeclared || command.RecordedIn != RecordedDomainFile || command.Effect.StateChanges != 1 {
		t.Fatalf("set_setpoint = %+v, want the effect the domain declares", command)
	}
}

func TestAssumptionsOfAHandAuthoredDomainStateOnlyWhatTheFileDeclares(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(testsupport.Domain("aquaculture-pond"))
	if err != nil {
		t.Fatal(err)
	}
	got := assumptionsOf(t, baseLibrary(t), AssumptionsInput{Domain: raw, Source: "aquaculture-pond.domain.json"})
	if got.Origin.Kind != OriginDomain || got.Domain.ID != "aquaculture-pond" || len(got.Kinds) == 0 {
		t.Fatalf("assumptions = %+v, want the pond domain as the domain file declares it", got.Origin)
	}
	for _, kind := range got.Kinds {
		for _, reading := range kind.Readings {
			if reading.Basis != BasisDomainDeclared || reading.Instrument != nil || reading.RecordedIn != RecordedDomainFile {
				t.Fatalf("reading %+v of %s, want domain declared", reading, kind.Kind)
			}
		}
	}
}

func TestAssumptionsAreTheSameBytesEveryTime(t *testing.T) {
	t.Parallel()
	library, input := fanZoneInput(t, func(choices string) string { return choices })
	first, err := json.Marshal(assumptionsOf(t, library, input))
	if err != nil {
		t.Fatal(err)
	}
	second, err := json.Marshal(assumptionsOf(t, library, input))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) || model.ValidateDomainAssumptions(first) != nil {
		t.Fatalf("two documents differ or break their contract: %v", model.ValidateDomainAssumptions(first))
	}
}

func TestAssumptionsRefuseWhatTheyCannotRead(t *testing.T) {
	t.Parallel()
	library, input := fanZoneInput(t, func(choices string) string { return choices })
	notJSON := AssumptionsInput{Domain: []byte("{"), Source: "bad.domain.json"}
	missingKind := input
	missingKind.Twin = &TwinFiles{Choices: input.Twin.Choices, Source: "fan-zone.twin.json", Kinds: kindDocuments(t, zoneKindJSON)}
	badChoices := input
	badChoices.Twin = &TwinFiles{Choices: []byte(`{"id":1}`), Source: "bad.twin.json", Kinds: input.Twin.Kinds}
	for name, tc := range map[string]struct {
		input    AssumptionsInput
		contract model.Contract
		cause    error
	}{
		"a domain that is not JSON":            {notJSON, model.ContractDomainSpec, nil},
		"twin choices that break their schema": {badChoices, model.ContractTwinChoices, nil},
		"a twin whose kind file is missing":    {missingKind, model.ContractTwinChoices, ErrNoKindFile},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := DescribeAssumptions(library, tc.input)
			invalid := asViolation(t, err, tc.contract)
			if tc.cause != nil && !errors.Is(invalid, tc.cause) {
				t.Fatalf("error = %v, want %v", err, tc.cause)
			}
		})
	}
}

func heaterInput(t *testing.T, change func(choices string) string) (*Library, AssumptionsInput) {
	t.Helper()
	library, kinds := baseLibrary(t), kindDocuments(t, heaterKindJSON)
	choices := change(heaterChoicesJSON)
	composed, err := ScaffoldTwin(library, kinds, []byte(choices), "heater.twin.json")
	if err != nil {
		t.Fatal(err)
	}
	return library, AssumptionsInput{Domain: composed.Bytes, Source: "heater.domain.json", Twin: &TwinFiles{Choices: []byte(choices), Source: "heater.twin.json", Kinds: kinds}}
}

func TestAssumptionsOfATwinSayWhatADirectAndANoneResponseDoAndNameACommandThatChangesNothing(t *testing.T) {
	t.Parallel()
	direct := func(choices string) string {
		changed := strings.Replace(choices, `"response": "first_order", "sets": {"setpoint": "setpoint"},
        "drives": [{"from": "setpoint", "kind": "heater", "metric": "temperature", "gain": 1, "time_constant_s": 120}]`, `"response": "direct", "sets": {"setpoint": "setpoint"}`, 1)
		if changed == choices {
			t.Fatal("the heater choices no longer hold the response to change")
		}
		return changed
	}
	library, input := heaterInput(t, direct)
	heater := assumptionsOf(t, library, input).kind(t, "heater")
	set, ping := heater.command(t, "set_setpoint"), heater.command(t, "ping")
	if set.Response != ResponseDirect || len(set.Drives) != 0 || set.Statement != "Sets setpoint to the commanded value at once." {
		t.Fatalf("set_setpoint = %+v, want a direct response", set)
	}
	if ping.Response != ResponseNone || !strings.Contains(ping.Statement, "changes nothing") {
		t.Fatalf("ping = %+v, want a response that changes nothing", ping)
	}
	if codes := assumptionsOf(t, library, input).codes("heater.ping"); !slices.Equal(codes, []string{"command_changes_nothing", "acknowledgement_never_fails"}) {
		t.Fatalf("heater.ping codes = %v, want command_changes_nothing and acknowledgement_never_fails", codes)
	}
}

func TestAssumptionsOfADomainStateAReadingsDeadbandAndAnAcknowledgementThatCanFail(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(testsupport.Domain("open-field-irrigation"))
	if err != nil {
		t.Fatal(err)
	}
	got := assumptionsOf(t, baseLibrary(t), AssumptionsInput{Domain: raw, Source: "open-field-irrigation.domain.json"})
	var deadband, failing bool
	for _, kind := range got.Kinds {
		for _, reading := range kind.Readings {
			deadband = deadband || reading.Cadence.Deadband != nil
		}
		for _, command := range kind.Commands {
			failing = failing || !slices.Contains(got.codes(kind.Kind+"."+command.Name), "acknowledgement_never_fails")
		}
	}
	if !deadband || !failing {
		t.Fatalf("deadband = %v, failing acknowledgement = %v, want both stated by the irrigation domain", deadband, failing)
	}
}
