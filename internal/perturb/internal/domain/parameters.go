package domain

import (
	"fmt"
	"maps"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonvalue"
)

type paramCheck func(v any) error

func numInRange[T int64 | float64](lo, hi T) paramCheck {
	return func(v any) error {
		f, ok := jsonvalue.Number(v)
		if !ok {
			return fmt.Errorf("expected a number, got %T", v)
		}
		if f < float64(lo) || f > float64(hi) {
			return fmt.Errorf("must be in [%v,%v], got %v", lo, hi, f)
		}
		return nil
	}
}

func strEnum(values ...string) paramCheck {
	return func(v any) error {
		s, ok := v.(string)
		if !ok {
			return fmt.Errorf("expected a string, got %T", v)
		}
		for _, want := range values {
			if s == want {
				return nil
			}
		}
		return fmt.Errorf("must be one of %v, got %q", values, s)
	}
}

func payloadsCheck(v any) error {
	arr, ok := v.([]any)
	if !ok {
		return fmt.Errorf("expected an array of strings, got %T", v)
	}
	if len(arr) == 0 {
		return fmt.Errorf("must contain at least one payload")
	}
	for i, p := range arr {
		if _, ok := p.(string); !ok {
			return fmt.Errorf("payload %d: expected a string, got %T", i, p)
		}
	}
	return nil
}

var paramRules = map[string]map[string]paramCheck{
	Drop:           {"rate": numInRange(0.0, 1.0)},
	DuplicateBurst: {"rate": numInRange(0.0, 1.0)},
	IDReuse:        {"rate": numInRange(0.0, 1.0)},
	OutOfEnum:      {"rate": numInRange(0.0, 1.0)},
	OutOfRange: {
		"rate":      numInRange(0.0, 1.0),
		"magnitude": numInRange(-1e9, 1e9),
	},
	Reorder: {"max_displacement": numInRange(0, int64(1<<30))},
	DelayTail: {
		"mean_s":  numInRange(0.0, 1e9),
		"sigma_s": numInRange(0.0, 1e9),
	},
	ClockSkew: {
		"offset_s": numInRange(0.0, 1e9),
		"sign":     strEnum("positive", "negative"),
	},
	Oversize:       {"bytes": numInRange(1, int64(1<<62))},
	Storm:          {"multiplier": numInRange(1, int64(1<<20))},
	InjectionProbe: {"payloads": payloadsCheck},
}

func CheckApplication(name string, params map[string]any, fromNS, untilNS int64) error {
	if err := CheckParams(name, params); err != nil {
		return err
	}
	return CheckWindow(name, fromNS, untilNS)
}

func CheckParams(name string, params map[string]any) error {
	if !isName(name) {
		return fmt.Errorf("unknown perturbation %q", name)
	}
	return paramsViolation(name, params)
}

func CheckWindow(name string, fromNS, untilNS int64) error {
	if untilNS != 0 && untilNS < fromNS {
		return fmt.Errorf("%s window would end at %d, before it starts at %d", name, untilNS, fromNS)
	}
	return nil
}

func paramsViolation(name string, params map[string]any) error {
	rules, declared := paramRules[name]
	for _, key := range slices.Sorted(maps.Keys(params)) {
		if err := validateParameter(name, key, params[key], rules, declared); err != nil {
			return err
		}
	}
	return nil
}

func validateParameter(name, key string, value any, rules map[string]paramCheck, declared bool) error {
	if !declared {
		return fmt.Errorf("%s accepts no parameters (got %q)", name, key)
	}
	check, ok := rules[key]
	if !ok {
		return fmt.Errorf("%s has no parameter %q", name, key)
	}
	if err := check(value); err != nil {
		return fmt.Errorf("%s %s: %w", name, key, err)
	}
	return nil
}
