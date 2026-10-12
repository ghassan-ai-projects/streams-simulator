package domain

import (
	"fmt"
	"math"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (checks domainChecks) dynamicSchedule(d *model.Dynamics) error {
	if d.F0 == nil || d.F0.Schedule == nil {
		return nil
	}
	schedule := d.F0.Schedule
	for i, point := range schedule.Points {
		if i > 0 && point.AtS <= schedule.Points[i-1].AtS {
			return checks.bad(fmt.Sprintf("/f0/schedule/points/%d/at_s", i), "schedule points must be in strictly increasing at_s")
		}
		if schedule.PeriodS > 0 && point.AtS >= schedule.PeriodS {
			return checks.bad(fmt.Sprintf("/f0/schedule/points/%d/at_s", i), "schedule point at %v s lies outside its period of %v s", point.AtS, schedule.PeriodS)
		}
	}
	return nil
}

func (checks domainChecks) dynamicPeak(d *model.Dynamics) error {
	if d.F1 == nil || d.F1.Form != model.FormPeak {
		return nil
	}
	switch {
	case len(d.F1.Inputs) != 1:
		return checks.bad("/f1/inputs", "peak measures exactly one input")
	case !multipleOf(d.F1.BlockS*1000, float64(d.DTMs)):
		return checks.bad("/f1/block_s", "peak block of %v s is not a whole number of %d ms steps", d.F1.BlockS, d.DTMs)
	case !multipleOf(d.F1.PeriodS, d.F1.BlockS):
		return checks.bad("/f1/period_s", "peak period of %v s is not a whole number of %v s blocks", d.F1.PeriodS, d.F1.BlockS)
	}
	return nil
}

func multipleOf(whole, part float64) bool {
	if part <= 0 {
		return false
	}
	ratio := whole / part
	return ratio >= 1 && math.Abs(ratio-math.Round(ratio)) < 1e-9
}

func (checks domainChecks) f1Selector(f1 *model.F1Dyn) error {
	if !checks.hasState(f1.Selector) {
		return checks.bad("/selector", "regime selector %q is not a declared state", f1.Selector)
	}
	return nil
}
