package device

import (
	"fmt"
)

func (c *Capabilities) loadRoutes(routes map[string]routeDocument) error {
	for _, routeName := range sortedStringKeys(routes) {
		route := routes[routeName]
		if routeName == "" {
			return fmt.Errorf("device: capabilities contain an empty route")
		}
		if route.Operation == "" || route.Target == "" || route.SelectorField == "" {
			return fmt.Errorf("device: route %q must set operation, target, and selector_field", routeName)
		}
		if route.ExpiresAfterMS < 1 || route.ExpiresAfterMS > 86400000 {
			return fmt.Errorf("device: route %q expires_after_ms must be between 1 and 86400000", routeName)
		}
		if len(route.Presets) == 0 {
			return fmt.Errorf("device: route %q has no presets", routeName)
		}
		for field := range route.Bounds {
			for presetName, preset := range route.Presets {
				if _, ok := preset[field]; !ok {
					return fmt.Errorf("device: route %q preset %q does not produce bounded parameter %q", routeName, presetName, field)
				}
			}
		}
		for logicalTarget, physicalTarget := range route.TargetBindings {
			if logicalTarget == "" || physicalTarget == "" {
				return fmt.Errorf("device: route %q contains an empty target binding", routeName)
			}
			if physicalTarget != route.Target {
				return fmt.Errorf("device: route %q target binding %q resolves to %q, want %q", routeName, logicalTarget, physicalTarget, route.Target)
			}
		}
		if _, exists := c.targets[route.Target]; exists {
			return fmt.Errorf("device: target %q is declared by multiple routes", route.Target)
		}
		bounds, err := routeBounds(routeName, route)
		if err != nil {
			return err
		}
		stringValues := routeStringValues(route)
		c.targets[route.Target] = TargetCapability{
			Operation:      route.Operation,
			EnergizeField:  firstBoundField(bounds),
			ExpiresAfterMS: route.ExpiresAfterMS,
			Bounds:         bounds,
			StringValues:   stringValues,
		}
	}
	return nil
}

func routeStringValues(route routeDocument) map[string]map[string]struct{} {
	values := map[string]map[string]struct{}{}
	for _, preset := range route.Presets {
		for field, raw := range preset {
			value, ok := raw.(string)
			if !ok {
				continue
			}
			if values[field] == nil {
				values[field] = map[string]struct{}{}
			}
			values[field][value] = struct{}{}
		}
	}
	return values
}

func routeBounds(routeName string, route routeDocument) (map[string][2]float64, error) {
	values := map[string][]float64{}
	kinds := map[string]string{}
	for _, preset := range route.Presets {
		for field, raw := range preset {
			value, numeric := finiteNumber(raw)
			kind := "non-numeric"
			if numeric {
				kind = "numeric"
				values[field] = append(values[field], value)
			}
			if previous, exists := kinds[field]; exists && previous != kind {
				return nil, fmt.Errorf("device: route %q preset field %q mixes numeric and non-numeric values", routeName, field)
			}
			kinds[field] = kind
		}
	}
	for field, bound := range route.Bounds {
		if field == "" {
			return nil, fmt.Errorf("device: route %q contains an empty bounds parameter", routeName)
		}
		if kinds[field] != "numeric" {
			return nil, fmt.Errorf("device: route %q bound %q is not produced as a numeric preset parameter", routeName, field)
		}
		if bound.Min != nil && !finite(*bound.Min) || bound.Max != nil && !finite(*bound.Max) {
			return nil, fmt.Errorf("device: route %q bound %q is non-finite", routeName, field)
		}
		if bound.Min != nil && bound.Max != nil && *bound.Max < *bound.Min {
			return nil, fmt.Errorf("device: route %q bound %q has max < min", routeName, field)
		}
	}

	bounds := make(map[string][2]float64, len(values))
	for field, fieldValues := range values {
		min, max := fieldValues[0], fieldValues[0]
		for _, value := range fieldValues[1:] {
			if value < min {
				min = value
			}
			if value > max {
				max = value
			}
		}
		bound := route.Bounds[field]
		if bound.Min != nil {
			min = *bound.Min
		}
		if bound.Max != nil {
			max = *bound.Max
		}
		if max < min {
			return nil, fmt.Errorf("device: route %q field %q has max < min", routeName, field)
		}
		bounds[field] = [2]float64{min, max}
	}
	return bounds, nil
}
