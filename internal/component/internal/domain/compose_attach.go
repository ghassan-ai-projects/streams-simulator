package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/jsonvalue"

type reading struct {
	conv   conversion
	lo, hi float64
	scale  float64
}

func (b *entryBuild) attachInstruments() error {
	for _, point := range b.points {
		if err := b.attach(point); err != nil {
			return err
		}
	}
	return nil
}

func (b *entryBuild) attach(point MeasurementPoint) error {
	instrument := b.instruments[point.Name]
	measured, err := b.reading(point, instrument)
	if err != nil {
		return err
	}
	channel := b.channel(point, instrument, measured)
	faults, biased := b.instrumentFaults(point, instrument, measured, channel["name"].(string))
	if biased {
		b.addBias(point, instrument, channel)
	}
	b.channels = append(b.channels, channel)
	b.faults = append(b.faults, faults...)
	return nil
}

func (b *entryBuild) reading(point MeasurementPoint, instrument *Instrument) (reading, error) {
	unit, declared := b.stateUnit(point.State)
	if !declared || unit == "" {
		return reading{}, b.refuse("use", "state %q of measurement point %q has no unit", point.State, point.Name)
	}
	conv, err := b.c.lib.vocabulary.convert(point.UnitFamily, unit, instrument.Unit)
	if err != nil {
		return reading{}, b.refuse(relative("instruments", "replace", point.Name), "measurement point %q: %v", point.Name, err)
	}
	lo, hi := convertedRange(point, instrument, conv)
	if lo >= hi {
		return reading{}, b.refuse(relative("instruments", "replace", point.Name), "range of %q in %s is empty: %v to %v", point.Name, instrument.Unit, lo, hi)
	}
	return reading{conv: conv, lo: lo, hi: hi, scale: spanScale(instrument, hi-lo)}, nil
}

func convertedRange(point MeasurementPoint, instrument *Instrument, conv conversion) (float64, float64) {
	lo, hi := conv.apply(point.Range.Min), conv.apply(point.Range.Max)
	if instrument.Range != nil {
		lo, hi = max(lo, instrument.Range.Min), min(hi, instrument.Range.Max)
	}
	return lo, hi
}

func spanScale(instrument *Instrument, span float64) float64 {
	if instrument.Scaling == "span" {
		return span
	}
	return 1
}

func (b *entryBuild) stateUnit(state string) (string, bool) {
	for _, item := range members(b.ws.kind, "state") {
		if item["name"] == state {
			unit, _ := item["unit"].(string)
			return unit, true
		}
	}
	return "", false
}

func (b *entryBuild) channel(point MeasurementPoint, instrument *Instrument, measured reading) map[string]any {
	channel := cloneMap(instrument.Document["channel"])
	scaleSigma(channel, instrument.Scaling, measured.scale)
	channel["name"] = b.name + "." + point.Name
	channel["observes"], channel["unit"], channel["tag"] = point.State, instrument.Unit, point.Tag
	channel["range"] = map[string]any{"min": measured.lo, "max": measured.hi}
	if !measured.conv.identity() {
		channel["observation_gain"], channel["observation_offset"] = measured.conv.gain, measured.conv.offset
	}
	return channel
}

func scaleSigma(channel map[string]any, scaling string, scale float64) {
	noise, _ := channel["noise"].(map[string]any)
	if sigma, numeric := jsonvalue.Number(noise["sigma"]); scaling == "span" && numeric {
		noise["sigma"] = sigma * scale
	}
}

func (b *entryBuild) addBias(point MeasurementPoint, instrument *Instrument, channel map[string]any) {
	name := point.Name + "__bias"
	channel["observation_bias"] = map[string]any{"state": name, "coef": 1, "noise_scale_at_full": 1}
	b.biases = append(b.biases, map[string]any{"name": name, "unit": instrument.Unit, "initial": 0})
}

func (b *entryBuild) qualifyKind() error {
	qualifyFaults(b.ws.kind, b.name)
	qualifyEffectors(b.ws.kind, b.name)
	stripDescriptions(b.ws.kind)
	return b.checkBiasNames()
}

func (b *entryBuild) checkBiasNames() error {
	taken := map[string]bool{}
	for _, item := range members(b.ws.kind, "state") {
		name, _ := item["name"].(string)
		taken[name] = true
	}
	for _, bias := range b.biases {
		name, _ := bias.(map[string]any)["name"].(string)
		if taken[name] {
			return b.refuse("use", "%s already declares a state %q, the name of an instrument bias", b.equipment.ID, name)
		}
	}
	return nil
}
