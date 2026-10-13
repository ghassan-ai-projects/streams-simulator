package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestADailyProfileFollowsTheSiteClock(t *testing.T) {
	t.Parallel()
	day := &model.Schedule{PeriodS: 86400, Interpolation: "step", Points: []model.SchedulePoint{{AtS: 0, Value: 0}, {AtS: 8 * 3600, Value: 1}, {AtS: 17 * 3600, Value: 0}}}
	midnightUTC := int64(20000) * 86400 * secondsPerNS
	cases := []struct {
		offsetS float64
		utcHour int64
		want    float64
	}{
		{0, 7, 0}, {0, 8, 1}, {3 * 3600, 5, 1}, {3 * 3600, 4, 0}, {3 * 3600, 14, 0}, {-5 * 3600, 13, 1}, {-5 * 3600, 22, 0},
	}
	for _, tc := range cases {
		day.UTCOffsetS = tc.offsetS
		if got := scheduleValue(day, midnightUTC+tc.utcHour*3600*secondsPerNS, 0); got != tc.want {
			t.Fatalf("offset %v s, %02d:00 UTC: %v, want %v", tc.offsetS, tc.utcHour, got, tc.want)
		}
	}
}
