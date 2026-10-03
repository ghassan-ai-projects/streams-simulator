package device

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
	StringValues   map[string]map[string]struct{}
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
