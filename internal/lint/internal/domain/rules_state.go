package domain

import (
	"fmt"
	"math"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const (
	pinnedShare       = 0.5
	settleSpanPerHour = 0.05
)

var discreteForms = []string{model.FormTrip, model.FormAntiRecycle, model.FormSequencer, model.FormPID, model.FormAlarm}

func pinnedAtBound(s *series, g *grid) (Finding, bool) {
	if s.state == nil || s.dynamic == nil || spanOf(s.values) == 0 {
		return Finding{}, false
	}
	hits, bound := dominantBound(s.state, s.values)
	if float64(len(hits)) <= pinnedShare*float64(len(s.values)) {
		return Finding{}, false
	}
	detail := fmt.Sprintf("state sits at its declared %s in %d of %d samples", bound, len(hits), len(s.values))
	return s.finding(severityWarning, codePinnedAtBound, hits[0], g, detail), true
}

func dominantBound(state *model.State, values []float64) ([]int, string) {
	atMin := indexesWhere(values, func(v float64) bool { return state.Min != nil && v <= *state.Min })
	atMax := indexesWhere(values, func(v float64) bool { return state.Max != nil && v >= *state.Max })
	if len(atMax) > len(atMin) {
		return atMax, "max " + number(*state.Max)
	}
	if len(atMin) > 0 {
		return atMin, "min " + number(*state.Min)
	}
	return nil, ""
}

func notSettled(s *series, g *grid) (Finding, bool) {
	from, last := g.lastTenthStart(), len(s.values)-1
	if !settles(s) || from >= last {
		return Finding{}, false
	}
	rate, span := math.Abs(s.values[last]-s.values[from])/g.hoursBetween(from, last), spanOf(s.values)
	if !(rate > settleSpanPerHour*span) {
		return Finding{}, false
	}
	detail := fmt.Sprintf("still moves %s per hour over the last tenth of the run, above %s%% of its observed span %s per hour",
		number(rate), number(settleSpanPerHour*100), number(span))
	return s.finding(severityWarning, codeNotSettled, from, g, detail), true
}

func settles(s *series) bool {
	dyn := s.dynamic
	if s.state == nil || dyn == nil || !dyn.Integrated() || slices.Contains(discreteForms, dyn.F1.Form) {
		return false
	}
	return !cyclic(s.kind, s.name, map[string]bool{})
}

func spanOf(values []float64) float64 {
	return slices.Max(values) - slices.Min(values)
}
