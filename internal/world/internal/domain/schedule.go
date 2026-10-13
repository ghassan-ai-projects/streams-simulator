package domain

import (
	"sort"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func scheduleValue(schedule *model.Schedule, t, startNS int64) float64 {
	if schedule == nil {
		return 0
	}
	offset := scheduleOffset(schedule, t, startNS)
	reached := sort.Search(len(schedule.Points), func(i int) bool { return schedule.Points[i].AtS > offset }) - 1
	return interpolate(schedule, reached, offset)
}

func interpolate(schedule *model.Schedule, reached int, offset float64) float64 {
	from := pointBefore(schedule, reached)
	if schedule.Interpolation == model.InterpolateStep {
		return from.Value
	}
	to := pointAfter(schedule, reached)
	if to.AtS == from.AtS {
		return from.Value
	}
	return from.Value + (to.Value-from.Value)*(offset-from.AtS)/(to.AtS-from.AtS)
}

func scheduleOffset(schedule *model.Schedule, t, startNS int64) float64 {
	if schedule.PeriodS <= 0 {
		return float64(t-startNS) / secondsPerNS
	}
	periodNS := int64(schedule.PeriodS * secondsPerNS)
	local := t + int64(schedule.UTCOffsetS*secondsPerNS)
	return float64((local%periodNS+periodNS)%periodNS) / secondsPerNS
}

func pointBefore(schedule *model.Schedule, reached int) model.SchedulePoint {
	points := schedule.Points
	switch {
	case reached >= 0:
		return points[reached]
	case schedule.PeriodS > 0:
		last := points[len(points)-1]
		return model.SchedulePoint{AtS: last.AtS - schedule.PeriodS, Value: last.Value}
	}
	return points[0]
}

func pointAfter(schedule *model.Schedule, reached int) model.SchedulePoint {
	points := schedule.Points
	switch {
	case reached+1 < len(points):
		return points[reached+1]
	case schedule.PeriodS > 0:
		return model.SchedulePoint{AtS: points[0].AtS + schedule.PeriodS, Value: points[0].Value}
	}
	return points[len(points)-1]
}
