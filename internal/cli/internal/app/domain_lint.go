package app

import (
	"errors"
	"flag"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/lint"
)

type lintFlags struct {
	hours float64
	seed  uint64
	kinds listFlag
}

func registerLintFlags(fs *flag.FlagSet) *lintFlags {
	flags := &lintFlags{}
	fs.Float64Var(&flags.hours, "hours", lint.DefaultHours, "lint: simulated hours of the nominal day")
	fs.Uint64Var(&flags.seed, "seed", lint.DefaultSeed, "lint: world seed")
	registerKindsDirs(fs, &flags.kinds)
	return flags
}

func (flags *lintFlags) admit(fs *flag.FlagSet, verb string) error {
	var stray []string
	fs.Visit(func(set *flag.Flag) { stray = append(stray, set.Name) })
	if name, applies := misplacedFlag(stray, verb); name != "" {
		return usagef("domain %s: --%s applies to %s only", verb, name, applies)
	}
	return nil
}

func lintDomainFile(path string, flags *lintFlags) (any, error) {
	compiled, err := domain.Load(path)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	report, err := lint.Check(compiled, lint.Options{Seed: flags.seed, Hours: flags.hours})
	if errors.Is(err, lint.ErrInvalidHours) {
		return nil, &usageError{message: err.Error(), cause: err}
	}
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return lintOutcome(report)
}

func lintOutcome(report *lint.Report) (any, error) {
	if !lint.Failed(report) {
		return report, nil
	}
	message := fmt.Sprintf("domain lint: %d error findings in %s", report.Errors, report.Domain)
	return nil, &outcomeError{code: codeDomainLint, message: message, result: report}
}
