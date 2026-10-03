package perturb

import (
	"encoding/json"
	"fmt"
)

// paramCheck validates one declared parameter value. Perturbation parameters
// are fail-closed: a misspelled or unknown key is an error, not a silently
// ignored default, and a declared key must have the declared type and range.
type paramCheck func(v any) error

func numInRange[T int64 | float64](lo, hi T) paramCheck {
	return func(v any) error {
		f, ok := asFloat(v)
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

// payloadsCheck validates the injection_probe payload list: a non-empty array
// of strings. Payloads are the adversarial text families for the probe.
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

// paramRules maps each perturbation to its declared parameter keys. A
// perturbation with no entry accepts no parameters. Every key here is read
// by applyOne; anything else is rejected so a misspelled option can never be
// silently approximated by a default.
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

func validateParams(name string, params map[string]any) error {
	if params == nil {
		return nil
	}
	rules, declared := paramRules[name]
	for key := range params {
		if err := validateParameter(name, key, params[key], rules, declared); err != nil {
			return err
		}
	}
	return nil
}

// asFloat coerces a JSON number to float64 without accepting non-numeric
// types. JSON numbers arrive as float64 or int64 in-process, as json.Number
// when decoded from a command-log artifact, or as Go int from callers;
// strings are never numeric.
func asFloat(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case int64:
		return float64(x), true
	case int:
		return float64(x), true
	case json.Number:
		f, err := x.Float64()
		return f, err == nil
	}
	return 0, false
}

func validateParameter(name, key string, value any, rules map[string]paramCheck, declared bool) error {
	if !declared {
		return fmt.Errorf("perturb: %s accepts no parameters (got %q)", name, key)
	}
	check, ok := rules[key]
	if !ok {
		return fmt.Errorf("perturb: %s has no parameter %q", name, key)
	}
	if err := check(value); err != nil {
		return fmt.Errorf("perturb: %s %s: %w", name, key, err)
	}
	return nil
}
