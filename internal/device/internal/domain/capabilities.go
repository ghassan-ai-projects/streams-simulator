package domain

type Capabilities struct {
	targets   map[string]TargetCapability
	safeStops map[string]safeStopCapability
	digest    string
}

type TargetCapability struct {
	Operation      string
	EnergizeField  string
	ExpiresAfterMS int
	Bounds         map[string][2]float64
	StringValues   map[string]map[string]struct{}
}

type safeStopCapability struct {
	expiresAfterMS int
}

type capabilitiesDoc struct {
	ProtocolVersion int                         `json:"protocol_version"`
	Routes          map[string]routeDocument    `json:"routes"`
	SafeStops       map[string]safeStopDocument `json:"safe_stops"`
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
