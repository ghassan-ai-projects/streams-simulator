package domain

import (
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestASimulatedSpanReadsAsAnHoursMinutesSecondsClockRoundedToTheSecond(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		ns   int64
		want string
	}{
		"zero":                {0, "00:00:00"},
		"four minutes":        {240e9, "00:04:00"},
		"half an hour":        {1800e9, "00:30:00"},
		"half a second up":    {1_500_000_000, "00:00:02"},
		"under half a second": {499_999_999, "00:00:00"},
		"an hour and change":  {3_725e9, "01:02:05"},
		"past a hundred":      {360_000e9, "100:00:00"},
		"before the start":    {-5e9, "00:00:00"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := clockText(tc.ns); got != tc.want {
				t.Fatalf("clockText(%d) = %q, want %q", tc.ns, got, tc.want)
			}
		})
	}
}

func TestTheMarkdownScopeSentenceNamesTheSimulatedTimeAsAClock(t *testing.T) {
	t.Parallel()
	report := endedReport()
	report.Truth = &Truth{KPIs: &model.KPITruth{FromNS: report.Run.StartTimeNS, ToNS: report.Run.StartTimeNS + 7320e9}}
	if text := string(Markdown(report)); !strings.Contains(text, "Scope: whole run, 02:02:00 of simulated time.") {
		t.Fatalf("markdown lacks the scope as a clock:\n%s", text)
	}
}
