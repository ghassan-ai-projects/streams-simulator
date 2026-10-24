package domain

import (
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type Limit struct {
	Code    string `json:"code"`
	Side    string `json:"side,omitempty"`
	Message string `json:"message"`
}

type difference struct {
	code    string
	differs func(a, b RunInputs) bool
	message func(a, b RunInputs) string
}

type sideRule struct {
	code    string
	applies func(RunInputs) bool
	message func(RunInputs) string
}

const singleRun = "Each side is one run. A difference between two single runs does not establish that a change caused it, and a seed is one draw of the world's randomness."

func limitsOf(c *Comparison) []Limit {
	limits := []Limit{}
	for _, d := range differences {
		if d.differs(c.A, c.B) {
			limits = append(limits, Limit{Code: d.code, Message: d.message(c.A, c.B)})
		}
	}
	limits = append(limits, sideLimits(c)...)
	if rowsOnOneSide(c.Results) {
		limits = append(limits, Limit{Code: "kpi_only_in_one_run", Message: "Some KPIs or alarms exist in only one of the runs; their rows have no delta."})
	}
	return append(limits, Limit{Code: "single_run_per_side", Message: singleRun})
}

func sideLimits(c *Comparison) []Limit {
	limits := []Limit{}
	for _, rule := range sideRules {
		for _, side := range []struct {
			name string
			in   RunInputs
		}{{"a", c.A}, {"b", c.B}} {
			if rule.applies(side.in) {
				limits = append(limits, Limit{Code: rule.code, Side: side.name, Message: "Run " + strings.ToUpper(side.name) + ": " + rule.message(side.in)})
			}
		}
	}
	return limits
}

func rowsOnOneSide(results Results) bool {
	for _, row := range results.KPIs {
		if row.Presence != PresenceBoth {
			return true
		}
	}
	for _, row := range results.Alarms {
		if row.Presence != PresenceBoth {
			return true
		}
	}
	return false
}

var sideRules = []sideRule{
	{"stopped_early", func(in RunInputs) bool {
		return in.Termination != nil && in.Termination.Reason == model.TerminationStoppedByOperator
	}, func(in RunInputs) string {
		return "the operator stopped it at " + offset(in.Termination.AtNS, in.StartTimeNS) + ", so its results cover only part of its plan."
	}},
	{"incomplete", func(in RunInputs) bool { return !in.Complete }, func(in RunInputs) string {
		return "it ended incomplete (" + in.IncompleteReason + "), so its evidence is partial."
	}},
	{"files_unverified", func(in RunInputs) bool { return in.Verification != FilesMatch }, func(in RunInputs) string {
		return "its files are not shown to match the digests the run recorded (" + in.Verification + ")."
	}},
	{"kpi_truth_absent", func(in RunInputs) bool { return !in.KPITruth }, func(RunInputs) string {
		return "it records no whole-run KPI results, so it contributes no result rows."
	}},
}
