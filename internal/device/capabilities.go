package device

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"sort"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
)

// Capabilities is the device's declared, data-defined capability catalog. The
// canonical route catalog is reduced to the target-level admission data the
// device needs, while its safe-stop declarations remain available to the
// device lifecycle. The legacy targets shape is also accepted for backwards
// compatibility with existing simulator fixtures.
type Capabilities struct {
	targets   map[string]TargetCapability
	safeStops map[string]safeStopCapability
	digest    string
}

// TargetCapability is one target's declared capability.
type TargetCapability struct {
	Operation      string
	EnergizeField  string
	ExpiresAfterMS int
	Bounds         map[string][2]float64
}

type safeStopCapability struct {
	operation      string
	expiresAfterMS int
}

type capabilitiesDoc struct {
	ProtocolVersion int                             `json:"protocol_version"`
	Routes          map[string]routeDocument        `json:"routes"`
	SafeStops       map[string]safeStopDocument     `json:"safe_stops"`
	Targets         map[string]legacyTargetDocument `json:"targets"`
}

type routeDocument struct {
	Operation      string                        `json:"operation"`
	Target         string                        `json:"target"`
	TargetBindings map[string]string             `json:"target_bindings"`
	SelectorField  string                        `json:"selector_field"`
	ExpiresAfterMS int                           `json:"expires_after_ms"`
	Presets        map[string]map[string]any     `json:"presets"`
	Bounds         map[string]routeBoundDocument `json:"bounds"`
}

type routeBoundDocument struct {
	Min *float64 `json:"min"`
	Max *float64 `json:"max"`
}

type safeStopDocument struct {
	Operation      string `json:"operation"`
	ExpiresAfterMS int    `json:"expires_after_ms"`
}

type legacyTargetDocument struct {
	Operation     string                         `json:"operation"`
	EnergizeField string                         `json:"energize_field"`
	Bounds        map[string]legacyBoundDocument `json:"bounds"`
}

type legacyBoundDocument struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// canonicalCatalogDocument mirrors the typed representation used by the
// paired Agentic Stream loader. Omitted route bounds stay omitted from the
// digest preimage; this is part of the cross-repository contract rather than a
// property of the source file's whitespace.
type canonicalCatalogDocument struct {
	ProtocolVersion int                               `json:"protocol_version"`
	Routes          map[string]canonicalRouteDocument `json:"routes"`
	SafeStops       map[string]safeStopDocument       `json:"safe_stops,omitempty"`
}

type canonicalRouteDocument struct {
	Operation      string                            `json:"operation"`
	Target         string                            `json:"target"`
	TargetBindings map[string]string                 `json:"target_bindings,omitempty"`
	SelectorField  string                            `json:"selector_field"`
	ExpiresAfterMS int                               `json:"expires_after_ms"`
	Presets        map[string]map[string]any         `json:"presets"`
	Bounds         map[string]canonicalBoundDocument `json:"bounds,omitempty"`
}

type canonicalBoundDocument struct {
	Min *float64 `json:"min,omitempty"`
	Max *float64 `json:"max,omitempty"`
}

// LoadCapabilities parses and validates a device capability catalog from
// JSON. The canonical route form is the cross-repository contract; the legacy
// target form remains accepted so old simulator fixtures continue to load.
// Unknown fields and trailing JSON are rejected so a typo cannot silently
// change the device's safety envelope. The canonical digest is retained as
// the identity advertised in the device.state handshake.
func LoadCapabilities(data []byte) (*Capabilities, error) {
	var doc capabilitiesDoc
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&doc); err != nil {
		return nil, fmt.Errorf("device: decode capabilities: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("device: capabilities contain trailing JSON")
		}
		return nil, fmt.Errorf("device: decode trailing capabilities: %w", err)
	}
	if doc.ProtocolVersion != ProtocolVersion {
		return nil, fmt.Errorf("device: capabilities protocol_version %d != %d", doc.ProtocolVersion, ProtocolVersion)
	}
	if len(doc.Routes) > 0 && len(doc.Targets) > 0 {
		return nil, fmt.Errorf("device: capabilities must use routes or targets, not both")
	}
	if len(doc.Routes) == 0 && len(doc.Targets) == 0 {
		return nil, fmt.Errorf("device: capabilities declare no routes or targets")
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("device: decode capabilities for digest: %w", err)
	}
	var digest string
	var err error
	if len(doc.Routes) > 0 {
		digestValue, prepareErr := canonicalCatalogValue(doc)
		if prepareErr != nil {
			return nil, fmt.Errorf("device: prepare capabilities digest: %w", prepareErr)
		}
		digest, err = canonical.DigestDomain(canonical.CapabilityCatalogDomain, digestValue)
	} else {
		// Legacy target catalogs retain their historical unscoped identity.
		digest, err = canonical.Digest(raw)
	}
	if err != nil {
		return nil, fmt.Errorf("device: digest capabilities: %w", err)
	}

	caps := &Capabilities{
		targets:   make(map[string]TargetCapability),
		safeStops: make(map[string]safeStopCapability, len(doc.SafeStops)),
		digest:    digest,
	}
	if len(doc.Routes) > 0 {
		if err := caps.loadRoutes(doc.Routes); err != nil {
			return nil, err
		}
	} else {
		if err := caps.loadLegacyTargets(doc.Targets); err != nil {
			return nil, err
		}
	}
	if err := caps.loadSafeStops(doc.SafeStops); err != nil {
		return nil, err
	}
	return caps, nil
}

func canonicalCatalog(doc capabilitiesDoc) canonicalCatalogDocument {
	routes := make(map[string]canonicalRouteDocument, len(doc.Routes))
	for name, route := range doc.Routes {
		bounds := make(map[string]canonicalBoundDocument, len(route.Bounds))
		for field, bound := range route.Bounds {
			bounds[field] = canonicalBoundDocument{Min: bound.Min, Max: bound.Max}
		}
		if len(route.Bounds) == 0 {
			bounds = nil
		}
		routes[name] = canonicalRouteDocument{
			Operation:      route.Operation,
			Target:         route.Target,
			TargetBindings: route.TargetBindings,
			SelectorField:  route.SelectorField,
			ExpiresAfterMS: route.ExpiresAfterMS,
			Presets:        route.Presets,
			Bounds:         bounds,
		}
	}
	return canonicalCatalogDocument{
		ProtocolVersion: doc.ProtocolVersion,
		Routes:          routes,
		SafeStops:       doc.SafeStops,
	}
}

func canonicalCatalogValue(doc capabilitiesDoc) (map[string]any, error) {
	encoded, err := json.Marshal(canonicalCatalog(doc))
	if err != nil {
		return nil, fmt.Errorf("encode typed catalog: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	var value map[string]any
	if err := decoder.Decode(&value); err != nil {
		return nil, fmt.Errorf("decode typed catalog: %w", err)
	}
	return value, nil
}

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
		c.targets[route.Target] = TargetCapability{
			Operation:      route.Operation,
			EnergizeField:  firstBoundField(bounds),
			ExpiresAfterMS: route.ExpiresAfterMS,
			Bounds:         bounds,
		}
	}
	return nil
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

func (c *Capabilities) loadLegacyTargets(targets map[string]legacyTargetDocument) error {
	for _, name := range sortedStringKeys(targets) {
		target := targets[name]
		if target.Operation == "" {
			return fmt.Errorf("device: target %q declares no operation", name)
		}
		if target.EnergizeField == "" {
			return fmt.Errorf("device: target %q declares no energize_field", name)
		}
		bounds := make(map[string][2]float64, len(target.Bounds))
		for field, bound := range target.Bounds {
			if !finite(bound.Min) || !finite(bound.Max) {
				return fmt.Errorf("device: target %q field %q has a non-finite bound", name, field)
			}
			if bound.Max < bound.Min {
				return fmt.Errorf("device: target %q field %q has max < min", name, field)
			}
			bounds[field] = [2]float64{bound.Min, bound.Max}
		}
		if _, ok := bounds[target.EnergizeField]; !ok {
			return fmt.Errorf("device: target %q energize_field %q is not a bounded parameter", name, target.EnergizeField)
		}
		c.targets[name] = TargetCapability{Operation: target.Operation, EnergizeField: target.EnergizeField, Bounds: bounds}
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

func (c *Capabilities) target(name string) (TargetCapability, bool) {
	if c == nil {
		return TargetCapability{}, false
	}
	tc, ok := c.targets[name]
	return tc, ok
}
