package domain

import (
	"errors"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const (
	ComparisonVersion = "0.1"
	ScopeWholeRun     = "whole_run"
	NotAvailable      = "not_available"
	windowedReason    = "The run report carries whole-run KPI truth only (mean, minimum, maximum and samples over the whole run), so no event-relative window can be computed from it and none is shown."
)

var ErrNotEnded = errors.New("report: a run to compare has not ended")

type Comparison struct {
	SchemaVersion   string          `json:"schema_version"`
	A               RunInputs       `json:"a"`
	B               RunInputs       `json:"b"`
	Plan            PlanDiff        `json:"plan"`
	Limits          []Limit         `json:"comparability_limits"`
	Results         Results         `json:"results"`
	WindowedResults WindowedResults `json:"windowed_results"`
}

type WindowedResults struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type RunInputs struct {
	RunID             string             `json:"run_id,omitempty"`
	Name              string             `json:"name,omitempty"`
	Purpose           string             `json:"purpose,omitempty"`
	Mode              string             `json:"mode"`
	Seed              *uint64            `json:"seed,omitempty"`
	SimVersion        string             `json:"sim_version,omitempty"`
	Domain            *model.ArtifactRef `json:"domain,omitempty"`
	Adapter           *model.ArtifactRef `json:"adapter,omitempty"`
	TimeMode          string             `json:"time_mode,omitempty"`
	StartTimeNS       int64              `json:"start_time_ns"`
	RequestedEndNS    int64              `json:"requested_end_ns,omitempty"`
	EndTimeNS         int64              `json:"end_time_ns"`
	SimulatedLengthNS int64              `json:"simulated_length_ns"`
	Termination       *model.Termination `json:"termination,omitempty"`
	Rate              float64            `json:"rate,omitempty"`
	Complete          bool               `json:"complete"`
	IncompleteReason  string             `json:"incomplete_reason,omitempty"`
	Verification      string             `json:"verification"`
	KPITruth          bool               `json:"kpi_truth"`
}

func Compare(a, b *Report) *Comparison {
	comparison := &Comparison{SchemaVersion: ComparisonVersion, A: inputsOf(a), B: inputsOf(b),
		Plan: planOf(a, b), Results: resultsOf(a, b), WindowedResults: WindowedResults{Status: NotAvailable, Reason: windowedReason}}
	comparison.Limits = limitsOf(comparison)
	return comparison
}

func NotComparable(report *Report) string {
	switch {
	case report.Run.Ended && report.Truth != nil:
		return ""
	case report.TruthUnavailable != "":
		return report.TruthUnavailable
	}
	return "the run has not ended (no run.json), so its truth may not be read"
}

func inputsOf(report *Report) RunInputs {
	run := report.Run
	return RunInputs{RunID: run.RunID, Name: run.Name, Purpose: run.Purpose, Mode: run.Mode, Seed: run.Seed, SimVersion: run.SimVersion,
		Domain: run.Domain, Adapter: run.Adapter, TimeMode: run.TimeMode, StartTimeNS: run.StartTimeNS, RequestedEndNS: run.RequestedEndNS,
		EndTimeNS: run.EndTimeNS, SimulatedLengthNS: run.EndTimeNS - run.StartTimeNS, Termination: run.Termination, Rate: run.Rate,
		Complete: run.Complete, IncompleteReason: run.IncompleteReason, Verification: report.Verification.Status, KPITruth: kpisOf(report) != nil}
}

func kpisOf(report *Report) *model.KPITruth {
	if report.Truth == nil {
		return nil
	}
	return report.Truth.KPIs
}
