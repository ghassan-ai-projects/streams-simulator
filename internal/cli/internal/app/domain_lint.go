package app

import (
	"flag"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/lint"
)

type lintFlags struct {
	hours float64
	seed  uint64
}

func registerLintFlags(fs *flag.FlagSet) *lintFlags {
	flags := &lintFlags{}
	fs.Float64Var(&flags.hours, "hours", 24, "lint: simulated hours of the nominal day")
	fs.Uint64Var(&flags.seed, "seed", 1, "lint: world seed")
	return flags
}

func (flags *lintFlags) admit(fs *flag.FlagSet, verb string) error {
	var stray []string
	fs.Visit(func(set *flag.Flag) { stray = append(stray, set.Name) })
	if verb != "lint" && len(stray) > 0 {
		return usagef("domain %s: --%s applies to lint only", verb, stray[0])
	}
	if !(flags.hours > 0 && flags.hours <= lint.MaxHours) {
		return usagef("domain lint: --hours must be positive and at most %d, got %v", lint.MaxHours, flags.hours)
	}
	return nil
}

func lintDomainFile(path string, flags *lintFlags) (any, error) {
	compiled, err := domain.Load(path)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	report, err := lint.Check(compiled, lint.Options{Seed: flags.seed, Hours: flags.hours})
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return lintOutcome(report)
}

func lintOutcome(report *lint.Report) (any, error) {
	result := lintResult(report)
	if report.Errors == 0 {
		return result, nil
	}
	message := fmt.Sprintf("domain lint: %d error findings in %s", report.Errors, report.Domain)
	return nil, &outcomeError{code: codeDomainLint, message: message, result: result}
}

func lintResult(report *lint.Report) map[string]any {
	return map[string]any{
		"domain": report.Domain, "digest": report.Digest, "seed": report.Seed, "hours": report.Hours,
		"samples": report.Samples, "findings": report.Findings, "errors": report.Errors, "warnings": report.Warnings,
	}
}
