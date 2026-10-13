// Package domain holds the lint rules: the sampling of a nominal day and the
// judgement of its series, with no file, clock or process access.
package domain

import (
	"cmp"
	"errors"
	"slices"
	"strconv"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
)

const (
	maxHours = 720

	DefaultHours = 24
	DefaultSeed  = 1

	severityError   = "error"
	severityWarning = "warning"

	codeNonFinite     = "non_finite"
	codeOutOfRange    = "out_of_range"
	codePinnedAtBound = "pinned_at_bound"
	codeNotSettled    = "not_settled"
	codeConservation  = "conservation_residual"
)

var (
	ErrNoSpec       = errors.New("lint: a domain spec is required")
	ErrInvalidHours = errors.New("lint: hours must be positive and at most 720")
)

type Options struct {
	Seed  uint64
	Hours float64
}

type Finding struct {
	Severity string  `json:"severity"`
	Code     string  `json:"code"`
	Entity   string  `json:"entity"`
	Channel  string  `json:"channel,omitempty"`
	State    string  `json:"state,omitempty"`
	Fault    string  `json:"fault,omitempty"`
	AtS      float64 `json:"at_s"`
	Detail   string  `json:"detail"`
}

type Report struct {
	Domain   string    `json:"domain"`
	Digest   string    `json:"digest"`
	Seed     uint64    `json:"seed"`
	Hours    float64   `json:"hours"`
	Samples  int       `json:"samples"`
	Findings []Finding `json:"findings"`
	Errors   int       `json:"errors"`
	Warnings int       `json:"warnings"`
}

func newReport(spec *domain.Compiled, options Options, g *grid, findings []Finding) *Report {
	slices.SortFunc(findings, compareFindings)
	report := &Report{
		Domain: spec.Spec.ID, Digest: spec.Digest, Seed: options.Seed, Hours: options.Hours,
		Samples: len(g.timesNS), Findings: append([]Finding{}, findings...),
	}
	for _, finding := range findings {
		report.count(finding)
	}
	return report
}

func Failed(report *Report) bool {
	return report.Errors > 0
}

func (r *Report) count(finding Finding) {
	if finding.Severity == severityError {
		r.Errors++
		return
	}
	r.Warnings++
}

func compareFindings(a, b Finding) int {
	return cmp.Or(
		cmp.Compare(a.Entity, b.Entity),
		cmp.Compare(a.subject(), b.subject()),
		cmp.Compare(a.Code, b.Code),
		cmp.Compare(a.AtS, b.AtS),
	)
}

func (f Finding) subject() string {
	return cmp.Or(f.Channel, f.State, f.Fault)
}

func number(v float64) string {
	return strconv.FormatFloat(v, 'g', 6, 64)
}
