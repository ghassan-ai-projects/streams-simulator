package domain

func (e *evaluator) report() TwinReport {
	metrics, divergences := e.collect()
	summary := e.summarize(metrics, divergences)
	verdict, reason := judgement(summary)
	if stop := e.nodeStopReason(); stop != "" {
		verdict, reason = VerdictCannotJudge, stop
	}
	twin := e.input.Twin
	twin.Kinds = append([]TwinKind{}, twin.Kinds...)
	return TwinReport{SchemaVersion: TwinReportVersion, Verdict: verdict, Reason: reason, Summary: summary, Run: e.input.Run, Twin: twin,
		ToleranceRule: ToleranceRule{Default: "span_fraction", SpanFraction: e.fraction, Description: toleranceRuleText},
		Commands:      append([]TwinCommand{}, e.commands...), Metrics: metrics, Divergences: divergences}
}

func (e *evaluator) collect() ([]MetricReport, []Divergence) {
	metrics, divergences := []MetricReport{}, []Divergence{}
	for _, tally := range e.tallies {
		tally.closeRun()
		metrics = append(metrics, tally.report())
		divergences = append(divergences, tally.divergences...)
	}
	return metrics, divergences
}

func (e *evaluator) summarize(metrics []MetricReport, divergences []Divergence) TwinSummary {
	summary := TwinSummary{Devices: len(e.input.Trace.Devices), Metrics: len(metrics), Unattributed: e.input.Trace.Unattributed,
		Divergences: len(divergences)}
	for _, metric := range metrics {
		summary.add(metric)
	}
	summary.CommandsApplied, summary.CommandsNotApplied = e.countCommands()
	return summary
}

func (s *TwinSummary) add(metric MetricReport) {
	s.Compared += metric.Summary.Compared
	s.Within += metric.Summary.Within
	s.Diverged += metric.Summary.Diverged
	s.NotCompared += metric.Summary.NotCompared
	if metric.Verdict == VerdictCannotJudge {
		s.MetricsCannotJudge++
	}
}

func (e *evaluator) countCommands() (applied, notApplied int) {
	for _, command := range e.commands {
		if command.Status == CommandApplied {
			applied++
		} else {
			notApplied++
		}
	}
	return applied, notApplied
}

func (e *evaluator) nodeStopReason() string {
	for _, command := range e.commands {
		switch {
		case command.Origin != OriginNode:
		case command.Status == CommandTwinRefused:
			return ReasonNodeStopNotApplied
		case command.Status == CommandNotApplied && (command.Reason == NotAppliedNeverForwarded || command.Reason == NotAppliedNotAccepted):
			return ReasonNodeStopUnconfirmed
		}
	}
	return ""
}

func judgement(summary TwinSummary) (verdict, reason string) {
	switch {
	case summary.Diverged > 0:
		return VerdictDiverges, ""
	case summary.Compared > 0:
		return VerdictAgrees, ""
	case summary.CommandsApplied == 0:
		return VerdictCannotJudge, ReasonNoCommands
	}
	return VerdictCannotJudge, ReasonNoReadings
}
