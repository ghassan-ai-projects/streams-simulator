package domain

import (
	"math"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonvalue"
)

func inMetricUnit(plan *metricPlan) float64 {
	eu := plan.file.EURange
	if plan.instrument.Scaling == "span" {
		return eu.Max - eu.Min
	}
	return 1 / math.Abs(plan.conv.gain)
}

func scaleAbsoluteMembers(channel map[string]any, instrument *Instrument, factor float64) {
	if instrument.Scaling == "span" || factor == 1 {
		return
	}
	scaleMember(channel, "noise", "sigma", factor)
	scaleMember(channel, "cadence", "deadband", factor)
	scaleMember(channel, "drift", "rate_per_hour", factor)
}

func scaleMember(channel map[string]any, group, member string, factor float64) {
	section, _ := channel[group].(map[string]any)
	if value, numeric := jsonvalue.Number(section[member]); numeric {
		section[member] = value * factor
	}
}
