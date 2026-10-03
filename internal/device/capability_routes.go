package device

import (
	"fmt"
)

func (c *Capabilities) loadRoutes(routes map[string]routeDocument) error {
	for _, name := range sortedStringKeys(routes) {
		if err := c.loadRoute(name, routes[name]); err != nil {
			return err
		}
	}
	return nil
}

func routeStringValues(route routeDocument) map[string]map[string]struct{} {
	values := map[string]map[string]struct{}{}
	for _, preset := range route.Presets {
		appendRouteStrings(values, preset)
	}
	return values
}

func routeBounds(routeName string, route routeDocument) (map[string][2]float64, error) {
	values := routePresetValues{numeric: map[string][]float64{}, kinds: map[string]string{}}
	if err := values.collect(routeName, route.Presets); err != nil {
		return nil, err
	}
	if err := values.validateBounds(routeName, route.Bounds); err != nil {
		return nil, err
	}
	return values.effectiveBounds(routeName, route.Bounds)
}

func (c *Capabilities) loadRoute(name string, route routeDocument) error {
	if err := validateDeclaredRoute(name, route); err != nil {
		return err
	}
	if _, exists := c.targets[route.Target]; exists {
		return fmt.Errorf("device: target %q is declared by multiple routes", route.Target)
	}
	bounds, err := routeBounds(name, route)
	if err != nil {
		return err
	}
	c.targets[route.Target] = TargetCapability{Operation: route.Operation, EnergizeField: firstBoundField(bounds),
		ExpiresAfterMS: route.ExpiresAfterMS, Bounds: bounds, StringValues: routeStringValues(route)}
	return nil
}

func validateRouteIdentity(name string, route routeDocument) error {
	if name == "" {
		return fmt.Errorf("device: capabilities contain an empty route")
	}
	if route.Operation == "" || route.Target == "" || route.SelectorField == "" {
		return fmt.Errorf("device: route %q must set operation, target, and selector_field", name)
	}
	if route.ExpiresAfterMS < 1 || route.ExpiresAfterMS > 86400000 {
		return fmt.Errorf("device: route %q expires_after_ms must be between 1 and 86400000", name)
	}
	if len(route.Presets) == 0 {
		return fmt.Errorf("device: route %q has no presets", name)
	}
	return nil
}

func validateRoutePresets(name string, route routeDocument) error {
	for field := range route.Bounds {
		for presetName, preset := range route.Presets {
			if _, ok := preset[field]; !ok {
				return fmt.Errorf("device: route %q preset %q does not produce bounded parameter %q", name, presetName, field)
			}
		}
	}
	return nil
}

func validateTargetBindings(name string, route routeDocument) error {
	for logical, physical := range route.TargetBindings {
		if logical == "" || physical == "" {
			return fmt.Errorf("device: route %q contains an empty target binding", name)
		}
		if physical != route.Target {
			return fmt.Errorf("device: route %q target binding %q resolves to %q, want %q", name, logical, physical, route.Target)
		}
	}
	return nil
}

func appendRouteStrings(values map[string]map[string]struct{}, preset map[string]any) {
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

func validateDeclaredRoute(name string, route routeDocument) error {
	if err := validateRouteIdentity(name, route); err != nil {
		return err
	}
	if err := validateRoutePresets(name, route); err != nil {
		return err
	}
	if err := validateTargetBindings(name, route); err != nil {
		return err
	}
	return nil
}
