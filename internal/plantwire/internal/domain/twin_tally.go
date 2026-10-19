package domain

import (
	"cmp"
	"fmt"
	"slices"
)

type metricTally struct {
	metric      BoardMetric
	facts       MetricFacts
	tolerance   Tolerance
	support     string
	compared    int
	within      int
	diverged    int
	skipped     []ReasonCount
	series      Series
	open        *Divergence
	divergences []Divergence
	force       func(atNS int64) []string
}

func newTally(metric BoardMetric, facts MetricFacts, tolerance Tolerance, force func(int64) []string) *metricTally {
	return &metricTally{metric: metric, facts: facts, tolerance: tolerance, force: force, series: Series{Points: []SeriesPoint{}}}
}

func (t *metricTally) skip(reason string) {
	at := slices.IndexFunc(t.skipped, func(c ReasonCount) bool { return c.Reason == reason })
	if at < 0 {
		t.skipped = append(t.skipped, ReasonCount{Reason: reason})
		at = len(t.skipped) - 1
	}
	t.skipped[at].Count++
}

func (t *metricTally) point(atNS int64, board, twin float64) {
	t.compared++
	t.series.Total++
	if len(t.series.Points) >= MaxSeriesPoints {
		t.series.Truncated = true
		return
	}
	t.series.Points = append(t.series.Points, SeriesPoint{AtNS: atNS, Board: board, Twin: twin})
}

func (t *metricTally) agree() {
	t.within++
	t.closeRun()
}

func (t *metricTally) differ(atNS int64, board, twin, gap float64) {
	t.diverged++
	if t.open == nil {
		t.open = &Divergence{Name: fmt.Sprintf("%s.%s#%d", t.metric.Device, t.metric.Metric, len(t.divergences)+1),
			Device: t.metric.Device, Metric: t.metric.Metric, FirstNS: atNS, Tolerance: *t.tolerance.Value, WorstError: -1}
	}
	t.extend(atNS, board, twin, gap)
}

func (t *metricTally) extend(atNS int64, board, twin, gap float64) {
	run := t.open
	run.LastNS = atNS
	run.Readings++
	if gap > run.WorstError {
		run.WorstError, run.WorstAtNS, run.BoardValue, run.TwinValue = gap, atNS, board, twin
		run.Commands = t.force(atNS)
	}
}

func (t *metricTally) closeRun() {
	if t.open != nil {
		t.divergences = append(t.divergences, *t.open)
		t.open = nil
	}
}

func (t *metricTally) verdict() string {
	switch {
	case t.diverged > 0:
		return VerdictDiverges
	case t.compared > 0:
		return VerdictAgrees
	}
	return VerdictCannotJudge
}

func (t *metricTally) report() MetricReport {
	skipped := slices.SortedFunc(slices.Values(t.skipped), func(a, b ReasonCount) int { return cmp.Compare(a.Reason, b.Reason) })
	return MetricReport{Device: t.metric.Device, Kind: t.metric.Kind, Metric: t.metric.Metric, Unit: t.facts.Unit,
		Tolerance: t.tolerance, Verdict: t.verdict(), Summary: MetricSummary{Compared: t.compared, Within: t.within,
			Diverged: t.diverged, NotCompared: totalSkipped(skipped)}, NotCompared: append([]ReasonCount{}, skipped...), Series: t.series}
}

func totalSkipped(counts []ReasonCount) int {
	total := 0
	for _, count := range counts {
		total += count.Count
	}
	return total
}
