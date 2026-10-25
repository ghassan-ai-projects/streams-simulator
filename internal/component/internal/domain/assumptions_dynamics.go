package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

const formFirstOrderLag = "first_order_lag"

func dynamicsAssumptions(kind *model.Kind) []DynamicsAssumption {
	dynamics := make([]DynamicsAssumption, 0, len(kind.Dynamics))
	for i := range kind.Dynamics {
		dynamics = append(dynamics, dynamicsAssumption(&kind.Dynamics[i]))
	}
	return dynamics
}

func dynamicsAssumption(dynamics *model.Dynamics) DynamicsAssumption {
	value := DynamicsAssumption{Target: dynamics.Target, Tier: dynamics.Tier, Inputs: []InputValue{}}
	if dynamics.F1 == nil {
		return value
	}
	value.Form, value.Inputs = dynamics.F1.Form, inputsValues(dynamics.F1.Inputs)
	if dynamics.F1.Form == formFirstOrderLag {
		value.Gain, value.TimeConstantS, value.DeadTimeS = &dynamics.F1.Gain, &dynamics.F1.TimeConstantS, &dynamics.F1.DeadTimeS
	}
	if clamp := dynamics.F1.Clamp; clamp != nil {
		value.Clamp = &ClampValue{Min: clamp.Min, Max: clamp.Max}
	}
	return value
}

func inputsValues(inputs []model.F1Input) []InputValue {
	values := make([]InputValue, len(inputs))
	for i, input := range inputs {
		values[i] = InputValue{State: input.State, Coef: input.Coef}
		if input.Link != nil {
			values[i].Link = input.Link.Type
		}
	}
	return values
}

func faultAssumptions(kind *model.Kind) []FaultAssumption {
	faults := make([]FaultAssumption, 0, len(kind.Faults))
	for _, fault := range kind.Faults {
		faults = append(faults, faultAssumption(fault))
	}
	return faults
}

func faultAssumption(fault model.Fault) FaultAssumption {
	value := FaultAssumption{ID: fault.ID, Description: fault.Description, NegativeClass: fault.IsNegativeClass, OnsetShape: fault.Onset.Shape, Affects: []string{}}
	for _, effect := range fault.Affects {
		value.Affects = append(value.Affects, effect.State)
	}
	if fault.Observation != nil {
		value.OnReading = fault.Observation.Channel
	}
	return value
}
