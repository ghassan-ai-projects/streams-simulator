package domain

import (
	"encoding/json"
	"errors"
	"maps"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	specs "github.com/ghassan-ai-projects/streams-simulator/internal/domain"
)

type twinOverlay struct {
	origin *TwinOrigin
	facts  map[string]*kindFacts
}

type kindFacts struct {
	metrics  map[string]metricFact
	commands map[string]commandFact
}

type metricFact struct {
	instrument *Instrument
	equals     bool
}

type commandFact struct {
	response string
	drives   []DriveValue
	equals   *bool
}

func (o *twinOverlay) inSync() bool {
	return o != nil && o.origin.Status == StatusInSync
}

func (o *twinOverlay) metric(kind, channel string) (metricFact, bool) {
	if !o.inSync() || o.facts[kind] == nil {
		return metricFact{}, false
	}
	fact, found := o.facts[kind].metrics[channel]
	return fact, found
}

func (o *twinOverlay) command(kind, name string) (commandFact, bool) {
	if !o.inSync() || o.facts[kind] == nil {
		return commandFact{}, false
	}
	fact, found := o.facts[kind].commands[name]
	return fact, found
}

func overlayOf(lib *Library, compiled *specs.Compiled, files *TwinFiles) (*twinOverlay, error) {
	if files == nil {
		return nil, nil
	}
	build, scaffold, err := scaffoldTwinFiles(lib, files)
	if err != nil {
		return nil, err
	}
	origin := &TwinOrigin{ID: build.choices.ID, Version: build.choices.Version, Status: StatusOutOfSync}
	if scaffold.Domain.Digest != compiled.Digest {
		return &twinOverlay{origin: origin}, nil
	}
	return inSyncOverlay(build, origin, files.Kinds)
}

func scaffoldTwinFiles(lib *Library, files *TwinFiles) (*twinBuild, *Composed, error) {
	build, err := resolveTwin(lib, files.Kinds, files.Choices, files.Source)
	if err != nil {
		return nil, nil, err
	}
	scaffold, err := build.compose()
	return build, scaffold, err
}

func inSyncOverlay(build *twinBuild, origin *TwinOrigin, kinds map[string]map[string]any) (*twinOverlay, error) {
	origin.Status = StatusInSync
	facts, err := build.facts(kinds)
	return &twinOverlay{origin: origin, facts: facts}, err
}

func (t *twinBuild) facts(files map[string]map[string]any) (map[string]*kindFacts, error) {
	suggested, err := t.suggestion(files)
	facts := make(map[string]*kindFacts, len(t.kinds))
	for _, k := range t.kinds {
		facts[k.file.Name] = t.kindFacts(k, suggested[k.file.Name])
	}
	return facts, err
}

func (t *twinBuild) suggestion(files map[string]map[string]any) (map[string]kindChoice, error) {
	chosen := map[string]map[string]any{}
	for _, k := range t.kinds {
		chosen[k.file.Name] = files[k.file.Name]
	}
	document, err := t.lib.SuggestChoices(chosen)
	byKind, decodeErr := decodedChoices(document)
	return byKind, errors.Join(err, decodeErr)
}

func decodedChoices(document map[string]any) (map[string]kindChoice, error) {
	raw, encodeErr := canonical.Marshal(document)
	var choices twinChoices
	decodeErr := json.Unmarshal(raw, &choices)
	byKind := make(map[string]kindChoice, len(choices.Kinds))
	for _, choice := range choices.Kinds {
		byKind[choice.Kind] = choice
	}
	return byKind, errors.Join(encodeErr, decodeErr)
}

func (t *twinBuild) kindFacts(k *twinKind, suggested kindChoice) *kindFacts {
	facts := &kindFacts{metrics: map[string]metricFact{}, commands: map[string]commandFact{}}
	for _, plan := range k.metrics {
		facts.metrics[plan.file.Name] = metricFact{instrument: plan.instrument, equals: t.sameInstrument(plan.choice, suggested.Metrics[plan.file.Name])}
	}
	for _, command := range k.commands {
		facts.commands[command.file.Name] = commandFacts(command, suggested.Commands[command.file.Name])
	}
	return facts
}

func (t *twinBuild) sameInstrument(chosen, suggested metricChoice) bool {
	if chosen.Instrument == instrumentNone || suggested.Instrument == instrumentNone {
		return chosen.Instrument == suggested.Instrument
	}
	one, oneErr := t.lib.resolve(chosen.Instrument)
	other, otherErr := t.lib.resolve(suggested.Instrument)
	return oneErr == nil && otherErr == nil && one == other
}

func commandFacts(command *commandPlan, suggested commandChoice) commandFact {
	if command.file.SafeStop {
		return commandFact{response: ResponseFixedSafeStop}
	}
	equals := command.choice.Response == suggested.Response && maps.Equal(command.choice.Sets, suggested.Sets)
	return commandFact{response: command.choice.Response, drives: driveValues(command.choice.Drives), equals: &equals}
}

func driveValues(drives []driveChoice) []DriveValue {
	values := make([]DriveValue, len(drives))
	for i, drive := range drives {
		values[i] = DriveValue(drive)
	}
	return values
}
