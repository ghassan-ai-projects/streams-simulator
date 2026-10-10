package domain

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
)

func (c *Capabilities) loadLegacyTargets(targets map[string]legacyTargetDocument) error {
	for _, name := range sortedStringKeys(targets) {
		target, err := compileLegacyTarget(name, targets[name])
		if err != nil {
			return err
		}
		c.targets[name] = target
	}
	return nil
}

func (c *Capabilities) loadSafeStops(stops map[string]safeStopDocument) error {
	for _, target := range sortedStringKeys(stops) {
		stop := stops[target]
		if target == "" || stop.Operation != "safe_stop" {
			return fmt.Errorf("device: safe stop %q must use operation safe_stop", target)
		}
		if stop.ExpiresAfterMS < 1 || stop.ExpiresAfterMS > 86400000 {
			return fmt.Errorf("device: safe stop %q expires_after_ms must be between 1 and 86400000", target)
		}
		c.safeStops[target] = safeStopCapability{operation: stop.Operation, expiresAfterMS: stop.ExpiresAfterMS}
	}
	return nil
}

func firstBoundField(bounds map[string][2]float64) string {
	fields := make([]string, 0, len(bounds))
	for field := range bounds {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

func finiteNumber(raw any) (float64, bool) {
	switch value := raw.(type) {
	case float64:
		return value, finite(value)
	case json.Number:
		parsed, err := value.Float64()
		return parsed, err == nil && finite(parsed)
	default:
		return 0, false
	}
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func sortedStringKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// Digest returns the canonical identity of the loaded capability catalog.
func (c *Capabilities) Digest() string {
	if c == nil {
		return ""
	}
	return c.digest
}

func (c *Capabilities) hasSafeStop(target string) bool {
	if c == nil {
		return false
	}
	_, ok := c.safeStops[target]
	return ok
}

// TargetNames returns the catalog's device targets in deterministic order.
func (c *Capabilities) TargetNames() []string {
	if c == nil {
		return nil
	}
	return sortedStringKeys(c.targets)
}

// SafeStopNames returns the catalog's explicit safe-stop targets in
// deterministic order.
func (c *Capabilities) SafeStopNames() []string {
	if c == nil {
		return nil
	}
	return sortedStringKeys(c.safeStops)
}

func (c *Capabilities) target(name string) (TargetCapability, bool) {
	if c == nil {
		return TargetCapability{}, false
	}
	tc, ok := c.targets[name]
	return tc, ok
}

func compileLegacyTarget(name string, target legacyTargetDocument) (TargetCapability, error) {
	if err := validateLegacyTarget(name, target); err != nil {
		return TargetCapability{}, err
	}
	bounds, err := legacyTargetBounds(name, target.Bounds)
	if err != nil {
		return TargetCapability{}, err
	}
	if _, ok := bounds[target.EnergizeField]; !ok {
		return TargetCapability{}, fmt.Errorf("device: target %q energize_field %q is not a bounded parameter", name, target.EnergizeField)
	}
	return TargetCapability{Operation: target.Operation, EnergizeField: target.EnergizeField, Bounds: bounds}, nil
}

func legacyTargetBounds(name string, declared map[string]legacyBoundDocument) (map[string][2]float64, error) {
	bounds := make(map[string][2]float64, len(declared))
	for field, bound := range declared {
		if !finite(bound.Min) || !finite(bound.Max) {
			return nil, fmt.Errorf("device: target %q field %q has a non-finite bound", name, field)
		}
		if bound.Max < bound.Min {
			return nil, fmt.Errorf("device: target %q field %q has max < min", name, field)
		}
		bounds[field] = [2]float64{bound.Min, bound.Max}
	}
	return bounds, nil
}

func validateLegacyTarget(name string, target legacyTargetDocument) error {
	if target.Operation == "" {
		return fmt.Errorf("device: target %q declares no operation", name)
	}
	if target.EnergizeField == "" {
		return fmt.Errorf("device: target %q declares no energize_field", name)
	}
	return nil
}
