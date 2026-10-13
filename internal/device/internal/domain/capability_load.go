package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
)

func LoadCapabilities(data []byte) (*Capabilities, error) {
	doc, err := decodeCapabilities(data)
	if err != nil {
		return nil, err
	}
	digest, err := digestRouteCapabilities(doc)
	if err != nil {
		return nil, err
	}
	caps := &Capabilities{targets: make(map[string]TargetCapability), safeStops: make(map[string]safeStopCapability, len(doc.SafeStops)), digest: digest}
	if err := caps.loadCatalog(doc); err != nil {
		return nil, err
	}
	return caps, nil
}

func decodeCapabilities(data []byte) (capabilitiesDoc, error) {
	var doc capabilitiesDoc
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&doc); err != nil {
		return capabilitiesDoc{}, fmt.Errorf("device: decode capabilities: %w", err)
	}
	if err := rejectTrailingCapabilities(decoder); err != nil {
		return capabilitiesDoc{}, err
	}
	if err := admitCapabilitiesDocument(doc); err != nil {
		return capabilitiesDoc{}, err
	}
	return doc, nil
}

func canonicalCatalog(doc capabilitiesDoc) canonicalCatalogDocument {
	routes := make(map[string]canonicalRouteDocument, len(doc.Routes))
	// determinism-safe: copies a map into a map.
	for name, route := range doc.Routes {
		routes[name] = canonicalRoute(route)
	}
	return canonicalCatalogDocument{ProtocolVersion: doc.ProtocolVersion, Routes: routes, SafeStops: doc.SafeStops}
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

func (c *Capabilities) loadCatalog(doc capabilitiesDoc) error {
	if err := c.loadRoutes(doc.Routes); err != nil {
		return err
	}
	return c.loadSafeStops(doc.SafeStops)
}

func rejectTrailingCapabilities(decoder *json.Decoder) error {
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return fmt.Errorf("device: capabilities contain trailing JSON")
		}
		return fmt.Errorf("device: decode trailing capabilities: %w", err)
	}
	return nil
}

func admitCapabilitiesDocument(doc capabilitiesDoc) error {
	if doc.ProtocolVersion != ProtocolVersion {
		return fmt.Errorf("device: capabilities protocol_version %d != %d", doc.ProtocolVersion, ProtocolVersion)
	}
	if len(doc.Routes) == 0 {
		return fmt.Errorf("device: capabilities declare no routes")
	}
	return nil
}

func digestRouteCapabilities(doc capabilitiesDoc) (string, error) {
	value, err := canonicalCatalogValue(doc)
	if err != nil {
		return "", fmt.Errorf("device: prepare capabilities digest: %w", err)
	}
	digest, err := canonical.DigestDomain(canonical.CapabilityCatalogDomain, value)
	if err != nil {
		return "", fmt.Errorf("device: digest capabilities: %w", err)
	}
	return digest, nil
}

func canonicalRoute(route routeDocument) canonicalRouteDocument {
	bounds := make(map[string]canonicalBoundDocument, len(route.Bounds))
	// determinism-safe: copies a map into a map.
	for field, bound := range route.Bounds {
		bounds[field] = canonicalBoundDocument(bound)
	}
	if len(route.Bounds) == 0 {
		bounds = nil
	}
	return canonicalRouteDocument{Operation: route.Operation, Target: route.Target, TargetBindings: route.TargetBindings,
		SelectorField: route.SelectorField, ExpiresAfterMS: route.ExpiresAfterMS, Presets: route.Presets, Bounds: bounds}
}
