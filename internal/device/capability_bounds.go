package device

import "fmt"

type routePresetValues struct {
	numeric map[string][]float64
	kinds   map[string]string
}

func (values *routePresetValues) collect(name string, presets map[string]map[string]any) error {
	for _, preset := range presets {
		for field, raw := range preset {
			if err := values.record(name, field, raw); err != nil {
				return err
			}
		}
	}
	return nil
}

func (values *routePresetValues) record(name, field string, raw any) error {
	number, numeric := finiteNumber(raw)
	kind := "non-numeric"
	if numeric {
		kind = "numeric"
		values.numeric[field] = append(values.numeric[field], number)
	}
	if previous, exists := values.kinds[field]; exists && previous != kind {
		return fmt.Errorf("device: route %q preset field %q mixes numeric and non-numeric values", name, field)
	}
	values.kinds[field] = kind
	return nil
}

func (values *routePresetValues) validateBounds(name string, bounds map[string]routeBoundDocument) error {
	for field, bound := range bounds {
		if err := validateRouteBound(name, field, values.kinds[field], bound); err != nil {
			return err
		}
	}
	return nil
}

func validateRouteBound(name, field, kind string, bound routeBoundDocument) error {
	if field == "" {
		return fmt.Errorf("device: route %q contains an empty bounds parameter", name)
	}
	if kind != "numeric" {
		return fmt.Errorf("device: route %q bound %q is not produced as a numeric preset parameter", name, field)
	}
	if bound.Min != nil && !finite(*bound.Min) || bound.Max != nil && !finite(*bound.Max) {
		return fmt.Errorf("device: route %q bound %q is non-finite", name, field)
	}
	if bound.Min != nil && bound.Max != nil && *bound.Max < *bound.Min {
		return fmt.Errorf("device: route %q bound %q has max < min", name, field)
	}
	return nil
}

func (values *routePresetValues) effectiveBounds(name string, declared map[string]routeBoundDocument) (map[string][2]float64, error) {
	bounds := make(map[string][2]float64, len(values.numeric))
	for field, numbers := range values.numeric {
		minimum, maximum := presetExtremes(numbers, declared[field])
		if maximum < minimum {
			return nil, fmt.Errorf("device: route %q field %q has max < min", name, field)
		}
		bounds[field] = [2]float64{minimum, maximum}
	}
	return bounds, nil
}

func presetExtremes(numbers []float64, bound routeBoundDocument) (float64, float64) {
	minimum, maximum := numericExtremes(numbers)
	if bound.Min != nil {
		minimum = *bound.Min
	}
	if bound.Max != nil {
		maximum = *bound.Max
	}
	return minimum, maximum
}

func numericExtremes(numbers []float64) (float64, float64) {
	minimum, maximum := numbers[0], numbers[0]
	for _, number := range numbers[1:] {
		if number < minimum {
			minimum = number
		}
		if number > maximum {
			maximum = number
		}
	}
	return minimum, maximum
}
