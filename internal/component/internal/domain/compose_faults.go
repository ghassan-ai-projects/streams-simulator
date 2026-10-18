package domain

func instrumentFaults(owner string, point MeasurementPoint, instrument *Instrument, measured reading, channel string) ([]any, bool) {
	var faults []any
	biased := false
	for _, fault := range instrument.Faults {
		document := instrumentFault(owner, point, fault, measured, channel)
		biased = biased || actsThroughBias(fault)
		faults = append(faults, document)
	}
	return faults, biased
}

func actsThroughBias(fault InstrumentFault) bool {
	return fault.Form == "offset" || fault.Form == "drift"
}

func instrumentFault(owner string, point MeasurementPoint, fault InstrumentFault, measured reading, channel string) map[string]any {
	document := map[string]any{"id": owner + "__" + point.Name + "__" + fault.ID, "observability": detectorOn(channel)}
	if actsThroughBias(fault) {
		document["affects"], document["onset"] = biasEffect(point, fault, measured)
	} else {
		document["observation"] = observationOf(fault, channel, measured)
		document["onset"] = map[string]any{"magnitude": 1, "shape": "step"}
	}
	if fault.IsNegativeClass {
		document["is_negative_class"] = true
	}
	return document
}

func detectorOn(channel string) map[string]any {
	return map[string]any{
		"detector":             map[string]any{"form": "single_channel_snr", "channel": channel},
		"first_observable_snr": 3,
		"unavoidable_snr":      10,
	}
}

func biasEffect(point MeasurementPoint, fault InstrumentFault, measured reading) ([]any, map[string]any) {
	magnitude := fault.Magnitude * measured.scale
	bias := point.Name + "__bias"
	if fault.Form == "offset" {
		return []any{map[string]any{"state": bias, "delta": magnitude}}, map[string]any{"magnitude": 1, "shape": "step"}
	}
	delta := magnitude * fault.DurationS / 3600
	onset := map[string]any{"magnitude": 1, "shape": "ramp", "rate_per_hour": 3600 / fault.DurationS}
	return []any{map[string]any{"state": bias, "delta": delta}}, onset
}

func observationOf(fault InstrumentFault, channel string, measured reading) map[string]any {
	observation := map[string]any{"channel": channel, "form": fault.Form}
	switch fault.Form {
	case "stuck":
		observation["value"] = stuckValue(fault, measured)
	case "scale", "noise":
		observation["factor"] = fault.Factor
	}
	if fault.Quality != "" {
		observation["quality"] = fault.Quality
	}
	return observation
}

func stuckValue(fault InstrumentFault, measured reading) float64 {
	if fault.StuckAt == "high" {
		return measured.hi
	}
	return measured.lo
}
