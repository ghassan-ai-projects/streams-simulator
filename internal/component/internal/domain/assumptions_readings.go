package domain

import (
	"strconv"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func readingAssumptions(spec *model.DomainSpec, kind *model.Kind, overlay *twinOverlay) []ReadingAssumption {
	provenance := provenanceOf(spec, kind.Name)
	readings := make([]ReadingAssumption, 0, len(kind.Channels))
	for i := range kind.Channels {
		readings = append(readings, readingAssumption(kind, &kind.Channels[i], provenance, overlay))
	}
	return readings
}

func provenanceOf(spec *model.DomainSpec, kind string) *model.Provenance {
	for _, set := range spec.Instances {
		if set.Kind == kind && set.Provenance != nil {
			return set.Provenance
		}
	}
	return nil
}

func readingAssumption(kind *model.Kind, channel *model.Channel, provenance *model.Provenance, overlay *twinOverlay) ReadingAssumption {
	reading := declaredReading(kind, channel)
	if fact, found := overlay.metric(kind.Name, channel.Name); found {
		return twinReading(reading, fact)
	}
	if ref := recordedInstrument(kind.Name, channel.Name, provenance); ref != "" {
		reading.Basis, reading.RecordedIn = BasisLibraryInstrument, RecordedComposition
		reading.Instrument = instrumentValue(ref, provenance.InstrumentDigests[ref])
	}
	return reading
}

func declaredReading(kind *model.Kind, channel *model.Channel) ReadingAssumption {
	reading := ReadingAssumption{Channel: channel.Name, Observes: channel.Observes, Unit: channel.Unit, ValueType: channel.ValueType,
		Resolution: channel.Resolution, Noise: NoiseValue{Model: channel.Noise.Model, Sigma: channel.Noise.Sigma}, Absence: channel.Absence,
		Tag: channel.Tag, Cadence: cadenceOf(channel.Cadence), Basis: BasisDomainDeclared, RecordedIn: RecordedDomainFile}
	if channel.Range != nil {
		reading.Range = &RangeValue{Min: channel.Range.Min, Max: channel.Range.Max}
	}
	if channel.Availability != nil {
		reading.Availability = &AvailabilityValue{Uptime: channel.Availability.Uptime, MTTRS: channel.Availability.MTTRS}
	}
	if state := kind.StateDecl(channel.Observes); state != nil {
		reading.Initial = &state.Initial
	}
	return reading
}

func cadenceOf(cadence model.Cadence) CadenceValue {
	value := CadenceValue{Mode: cadence.Mode}
	if cadence.PeriodS != 0 {
		value.PeriodS = &cadence.PeriodS
	}
	if cadence.Deadband != 0 {
		value.Deadband = &cadence.Deadband
	}
	return value
}

func twinReading(reading ReadingAssumption, fact metricFact) ReadingAssumption {
	reading.RecordedIn, reading.Basis = RecordedTwinChoices, BasisPlainChannel
	if fact.instrument != nil {
		reading.Basis = BasisLibraryInstrument
		reading.Instrument = &InstrumentValue{ID: fact.instrument.ID, Version: fact.instrument.Version,
			Ref: reference(fact.instrument.ID, fact.instrument.Version), Digest: fact.instrument.Digest}
	}
	equals := fact.equals
	reading.EqualsBinarySuggestion = &equals
	return reading
}

func recordedInstrument(kind, channel string, provenance *model.Provenance) string {
	if provenance == nil {
		return ""
	}
	return provenance.Instruments[strings.TrimPrefix(channel, kind+".")]
}

func instrumentValue(ref, digest string) *InstrumentValue {
	id, version, _ := strings.Cut(ref, "@")
	number, _ := strconv.Atoi(version)
	return &InstrumentValue{ID: id, Version: number, Ref: ref, Digest: digest}
}
