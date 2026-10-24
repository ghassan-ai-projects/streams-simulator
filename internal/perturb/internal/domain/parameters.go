package domain

import (
	"fmt"
	"maps"
	"math"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonvalue"
)

const (
	ParamNumber     = "number"
	ParamInteger    = "integer"
	ParamEnum       = "enum"
	ParamStringList = "string_list"
)

type paramSpec struct {
	name   string
	kind   string
	lo, hi float64
	values []string
}

func number(name string, lo, hi float64) paramSpec {
	return paramSpec{name: name, kind: ParamNumber, lo: lo, hi: hi}
}

func integer(name string, lo, hi float64) paramSpec {
	return paramSpec{name: name, kind: ParamInteger, lo: lo, hi: hi}
}

func enum(name string, values ...string) paramSpec {
	return paramSpec{name: name, kind: ParamEnum, values: values}
}

func stringList(name string) paramSpec {
	return paramSpec{name: name, kind: ParamStringList}
}

func (spec paramSpec) check(v any) error {
	switch spec.kind {
	case ParamEnum:
		return spec.checkEnum(v)
	case ParamStringList:
		return checkPayloads(v)
	}
	return spec.checkRange(v)
}

func (spec paramSpec) checkRange(v any) error {
	f, ok := jsonvalue.Number(v)
	if !ok {
		return fmt.Errorf("expected a number, got %T", v)
	}
	if spec.kind == ParamInteger && f != math.Trunc(f) {
		return fmt.Errorf("expected a whole number, got %v", f)
	}
	if f < spec.lo || f > spec.hi {
		return fmt.Errorf("must be in [%v,%v], got %v", spec.bound(spec.lo), spec.bound(spec.hi), f)
	}
	return nil
}

func (spec paramSpec) bound(limit float64) any {
	if spec.kind == ParamInteger {
		return int64(limit)
	}
	return limit
}

func (spec paramSpec) checkEnum(v any) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("expected a string, got %T", v)
	}
	if !slices.Contains(spec.values, s) {
		return fmt.Errorf("must be one of %v, got %q", spec.values, s)
	}
	return nil
}

func checkPayloads(v any) error {
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

var paramSpecs = map[string][]paramSpec{
	Drop:           {number("rate", 0, 1)},
	DuplicateBurst: {number("rate", 0, 1)},
	IDReuse:        {number("rate", 0, 1)},
	OutOfEnum:      {number("rate", 0, 1)},
	OutOfRange:     {number("rate", 0, 1), number("magnitude", -1e9, 1e9)},
	Reorder:        {integer("max_displacement", 0, 1<<30)},
	DelayTail:      {number("mean_s", 0, 1e9), number("sigma_s", 0, 1e9)},
	ClockSkew:      {number("offset_s", 0, 1e9), enum("sign", "positive", "negative")},
	Oversize:       {integer("bytes", 1, 1<<62)},
	Storm:          {integer("multiplier", 1, 1<<20)},
	InjectionProbe: {stringList("payloads")},
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
	specs, declared := paramSpecs[name]
	for _, key := range slices.Sorted(maps.Keys(params)) {
		if err := validateParameter(name, key, params[key], specs, declared); err != nil {
			return err
		}
	}
	return nil
}

func validateParameter(name, key string, value any, specs []paramSpec, declared bool) error {
	if !declared {
		return fmt.Errorf("%s accepts no parameters (got %q)", name, key)
	}
	index := slices.IndexFunc(specs, func(spec paramSpec) bool { return spec.name == key })
	if index < 0 {
		return fmt.Errorf("%s has no parameter %q", name, key)
	}
	if err := specs[index].check(value); err != nil {
		return fmt.Errorf("%s %s: %w", name, key, err)
	}
	return nil
}
