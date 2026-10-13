package domain

import (
	"cmp"
	"fmt"
	"math"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func nonFinite(s *series, g *grid) (Finding, bool) {
	hits := indexesWhere(s.values, func(v float64) bool { return !finite(v) })
	if len(hits) == 0 {
		return Finding{}, false
	}
	detail := fmt.Sprintf("value %s, %d of %d samples are not finite", number(s.values[hits[0]]), len(hits), len(s.values))
	return s.finding(severityError, codeNonFinite, hits[0], g, detail), true
}

func outOfRange(s *series, g *grid) (Finding, bool) {
	if s.channel == nil || s.channel.Range == nil {
		return Finding{}, false
	}
	declared := s.channel.Range
	hits := indexesWhere(s.values, func(v float64) bool { return excess(declared.Min, declared.Max, v) > 0 })
	if len(hits) == 0 {
		return Finding{}, false
	}
	worst := worstOutside(declared, s.values, hits)
	detail := fmt.Sprintf("reading %s outside the declared range [%s, %s] (worst), %d of %d samples outside",
		number(s.values[worst]), number(declared.Min), number(declared.Max), len(hits), len(s.values))
	return s.finding(severityError, codeOutOfRange, hits[0], g, detail), true
}

func worstOutside(declared *model.Range, values []float64, hits []int) int {
	return slices.MaxFunc(hits, func(a, b int) int {
		return cmp.Compare(excess(declared.Min, declared.Max, values[a]), excess(declared.Min, declared.Max, values[b]))
	})
}

func excess(low, high, v float64) float64 {
	return math.Max(low-v, v-high)
}

func indexesWhere(values []float64, test func(float64) bool) []int {
	var hits []int
	for i, v := range values {
		if test(v) {
			hits = append(hits, i)
		}
	}
	return hits
}

func finite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}
