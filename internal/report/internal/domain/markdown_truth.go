package domain

import (
	"strconv"
	"strings"
)

const withheld = "Truth withheld: the run has not ended (no run.json)."

func truthSections(report *Report) []section {
	truth, start := report.Truth, report.Run.StartTimeNS
	if truth == nil {
		return []section{{"Truth", []string{withheldReason(report)}}}
	}
	return []section{
		{"Faults", listOrNone(len(truth.Faults), faultLines(truth.Faults, start))},
		{"Perturbations", listOrNone(len(truth.Perturbations), perturbationLines(truth.Perturbations, start))},
		{"Environment", listOrNone(len(truth.Environment), environmentLines(truth.Environment, start))},
		{"Effector calls", listOrNone(len(truth.EffectorCalls), effectorLines(truth.EffectorCalls, start))},
	}
}

func withheldReason(report *Report) string {
	if report.TruthUnavailable != "" {
		return "Truth unavailable: " + escapeText(report.TruthUnavailable)
	}
	return withheld
}

func listOrNone(count int, lines []string) []string {
	if count == 0 {
		return []string{"None."}
	}
	return lines
}

func faultLines(faults []Fault, start int64) []string {
	rows := make([][]string, len(faults))
	for i, f := range faults {
		rows[i] = []string{seq(f.Seq), instant(f.AtNS, start), f.Action, f.EntityID, first(f.Fault, f.FaultID),
			optionalInstant(f.OnsetNS, start), codeJSON(f.Params)}
	}
	return table([]string{"Seq", "At", "Action", "Entity", "Fault", "Onset", "Params"}, rows)
}

func perturbationLines(items []Perturbation, start int64) []string {
	rows := make([][]string, len(items))
	for i, p := range items {
		rows[i] = []string{seq(p.Seq), instant(p.AtNS, start), p.Action, first(p.Perturbation, p.PerturbID),
			optionalInstant(p.FromNS, start), optionalInstant(p.UntilNS, start), codeJSON(p.Params)}
	}
	return table([]string{"Seq", "At", "Action", "Perturbation", "From", "Until", "Params"}, rows)
}

func environmentLines(items []Environment, start int64) []string {
	rows := make([][]string, len(items))
	for i, e := range items {
		rows[i] = []string{seq(e.Seq), instant(e.AtNS, start), e.Target, e.Fault, codeJSON(e.Params)}
	}
	return table([]string{"Seq", "At", "Target", "Fault", "Params"}, rows)
}

func effectorLines(calls []EffectorCall, start int64) []string {
	rows := make([][]string, len(calls))
	for i, c := range calls {
		rows[i] = []string{seq(c.Seq), instant(c.AtNS, start), c.CommandID, c.Effector, c.EntityID,
			codeJSON(c.Args), outcomeText(c.Outcome)}
	}
	return table([]string{"Seq", "At", "Command", "Effector", "Entity", "Args", "Outcome"}, rows)
}

func outcomeText(outcome *Outcome) string {
	if outcome == nil {
		return ""
	}
	if outcome.Refusal != "" {
		return "refused: " + outcome.Refusal
	}
	return strings.Join(append([]string{"mode=" + outcome.Mode}, acknowledgement(outcome)...), " ")
}

func acknowledgement(outcome *Outcome) []string {
	var parts []string
	if outcome.Accepted != nil {
		parts = append(parts, "accepted="+yesNo(*outcome.Accepted))
	}
	if outcome.EffectApplied != nil {
		parts = append(parts, "effect_applied="+yesNo(*outcome.EffectApplied))
	}
	if outcome.AckLatencyMS != nil {
		parts = append(parts, "ack_latency_ms="+strconv.FormatFloat(*outcome.AckLatencyMS, 'f', -1, 64))
	}
	return parts
}

func seq(n int64) string { return strconv.FormatInt(n, 10) }

func first(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
