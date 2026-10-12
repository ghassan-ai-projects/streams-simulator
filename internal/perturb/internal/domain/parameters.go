package domain

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
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

// CheckApplication reports whether applying name with params over the window
// from fromNS to untilNS is admitted, without applying it: the name is in
// the catalog, the window can act (CheckWindow) and params holds only
// declared parameters, in range (CheckParams). The error states the
// violation.
func CheckApplication(name string, params map[string]any, fromNS, untilNS int64) error {
	if err := CheckParams(name, params); err != nil {
		return err
	}
	return CheckWindow(name, fromNS, untilNS)
}

// CheckParams reports whether name is in the catalog and params are
// parameters it accepts.
func CheckParams(name string, params map[string]any) error {
	if !isName(name) {
		return fmt.Errorf("unknown perturbation %q", name)
	}
	return paramsViolation(name, params)
}

// CheckWindow reports whether an activity window can act: untilNS is 0 (no
// end) or not before fromNS.
func CheckWindow(name string, fromNS, untilNS int64) error {
	if untilNS != 0 && untilNS < fromNS {
		return fmt.Errorf("%s window would end at %d, before it starts at %d", name, untilNS, fromNS)
	}
	return nil
}

// paramsViolation is the first parameter, in key order, that the
// perturbation does not accept.
func paramsViolation(name string, params map[string]any) error {
	rules, declared := paramRules[name]
	for _, key := range slices.Sorted(maps.Keys(params)) {
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
