package domain

import "maps"

var channelValueTypes = map[string]string{"number": "number", "integer": "counter", "boolean": "boolean"}

type kindOutput struct {
	states, dynamics, channels, faults, effectors []any
}

type metricReading struct {
	channel map[string]any
	faults  []any
	biases  []any
}

func valueName(metric string) string { return metric + "_value" }

func restName(metric string) string { return metric + "_rest" }

func (t *twinBuild) kindDocuments() []any {
	documents := make([]any, len(t.kinds))
	for i, k := range t.kinds {
		documents[i] = kindDocument(k)
	}
	return documents
}

func kindDocument(k *twinKind) map[string]any {
	out := &kindOutput{}
	for _, plan := range k.metrics {
		out.addMetric(k, plan)
	}
	for _, command := range k.commands {
		out.effectors = append(out.effectors, effectorDocument(k, command))
	}
	return out.document(k)
}

func (o *kindOutput) document(k *twinKind) map[string]any {
	document := map[string]any{"name": k.file.Name, "implements": k.file.Name, "state": o.states, "channels": o.channels, "faults": o.faults}
	if len(o.dynamics) > 0 {
		document["dynamics"] = o.dynamics
	}
	if len(o.effectors) > 0 {
		document["effectors"] = o.effectors
	}
	return document
}

func (o *kindOutput) addMetric(k *twinKind, plan *metricPlan) {
	o.states = append(o.states, stateOf(valueName(plan.file.Name), plan))
	if len(plan.driven) > 0 {
		o.states = append(o.states, stateOf(restName(plan.file.Name), plan))
		o.dynamics = append(o.dynamics, dynamicsOf(k, plan))
	}
	observed := readingOf(k, plan)
	o.channels = append(o.channels, observed.channel)
	o.faults = append(o.faults, observed.faults...)
	o.states = append(o.states, observed.biases...)
}

func stateOf(name string, plan *metricPlan) map[string]any {
	state := map[string]any{"name": name, "initial": initialOf(plan)}
	if plan.file.Unit != "" {
		state["unit"] = plan.file.Unit
	}
	if eu := plan.file.EURange; eu != nil {
		state["min"], state["max"] = eu.Min, eu.Max
	}
	return state
}

func initialOf(plan *metricPlan) float64 {
	switch eu := plan.file.EURange; {
	case plan.choice.Initial != nil:
		return *plan.choice.Initial
	case eu == nil:
		return 0
	case plan.file.Writable:
		return eu.Min
	default:
		return (eu.Min + eu.Max) / 2
	}
}

func dynamicsOf(k *twinKind, plan *metricPlan) map[string]any {
	f1 := map[string]any{"form": "first_order_lag", "gain": 1, "time_constant_s": plan.tauS, "inputs": inputsOf(k, plan)}
	if eu := plan.file.EURange; eu != nil {
		f1["clamp"] = map[string]any{"min": eu.Min, "max": eu.Max}
	}
	return map[string]any{"target": valueName(plan.file.Name), "tier": "F1", "dt_ms": stepMS, "f1": f1}
}

func inputsOf(k *twinKind, plan *metricPlan) []any {
	inputs := []any{map[string]any{"state": restName(plan.file.Name), "coef": 1}}
	for _, c := range plan.driven {
		input := map[string]any{"state": c.state, "coef": c.gain}
		if c.driver != k {
			input["link"] = map[string]any{"type": linkTypeName(c.driver, k), "aggregate": "sum"}
		}
		inputs = append(inputs, input)
	}
	return inputs
}

func readingOf(k *twinKind, plan *metricPlan) metricReading {
	if plan.instrument == nil {
		return plainReading(k, plan)
	}
	return instrumentedReading(k, plan)
}

func baseChannel(metric kindMetric) map[string]any {
	channel := map[string]any{"name": metric.Name, "observes": valueName(metric.Name), "value_type": channelValueTypes[metric.DataType], "resolution": metric.Resolution}
	if metric.Unit != "" {
		channel["unit"] = metric.Unit
	}
	if eu := metric.EURange; eu != nil {
		channel["range"] = map[string]any{"min": eu.Min, "max": eu.Max}
	}
	return channel
}

func plainReading(k *twinKind, plan *metricPlan) metricReading {
	channel := baseChannel(plan.file)
	channel["fidelity"], channel["absence"] = "F1", "signal"
	channel["cadence"] = map[string]any{"mode": "periodic", "period_s": plainPeriodS}
	channel["noise"] = map[string]any{"model": "none", "sigma": 0}
	return metricReading{channel: channel, faults: dropoutFault(k, plan)}
}

func dropoutFault(k *twinKind, plan *metricPlan) []any {
	dropout := InstrumentFault{ID: "dropout", Form: "dropout"}
	return []any{instrumentFault(k.file.Name, MeasurementPoint{Name: plan.file.Name}, dropout, reading{}, plan.file.Name)}
}

func instrumentedReading(k *twinKind, plan *metricPlan) metricReading {
	instrument, eu := plan.instrument, plan.file.EURange
	measured := reading{conv: conversion{gain: 1}, lo: eu.Min, hi: eu.Max, scale: inMetricUnit(plan)}
	channel := cloneMap(instrument.Document["channel"])
	maps.Copy(channel, baseChannel(plan.file))
	channel["tag"] = instrument.Tag
	scaleSigma(channel, instrument.Scaling, measured.scale)
	scaleAbsoluteMembers(channel, instrument, measured.scale)
	copyStates(channel, instrument)
	return withInstrumentFaults(k, plan, channel, measured)
}

func withInstrumentFaults(k *twinKind, plan *metricPlan, channel map[string]any, measured reading) metricReading {
	point := MeasurementPoint{Name: plan.file.Name}
	faults, biased := instrumentFaults(k.file.Name, point, plan.instrument, measured, point.Name)
	if len(faults) == 0 {
		faults = dropoutFault(k, plan)
	}
	out := metricReading{channel: channel, faults: faults}
	if biased {
		out.biases = []any{biasState(channel, point, plan.file.Unit)}
	}
	return out
}
