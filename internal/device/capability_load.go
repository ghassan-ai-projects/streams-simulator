package device

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
)

// LoadCapabilities parses and validates a device capability catalog from
// JSON. The canonical route form is the cross-repository contract; the legacy
// target form remains accepted so old simulator fixtures continue to load.
// Unknown fields and trailing JSON are rejected so a typo cannot silently
// change the device's safety envelope. The canonical digest is retained as
// the identity advertised in the device.state handshake.
func LoadCapabilities(data []byte) (*Capabilities, error) {
	doc, err := decodeCapabilities(data)
	if err != nil {
		return nil, err
	}
	digest, err := capabilitiesDigest(data, doc)
	if err != nil {
		return nil, err
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

func decodeCapabilities(data []byte) (capabilitiesDoc, error) {
	var doc capabilitiesDoc
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&doc); err != nil {
		return capabilitiesDoc{}, fmt.Errorf("device: decode capabilities: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return capabilitiesDoc{}, fmt.Errorf("device: capabilities contain trailing JSON")
		}
		return capabilitiesDoc{}, fmt.Errorf("device: decode trailing capabilities: %w", err)
	}
	if doc.ProtocolVersion != ProtocolVersion {
		return capabilitiesDoc{}, fmt.Errorf("device: capabilities protocol_version %d != %d", doc.ProtocolVersion, ProtocolVersion)
	}
	if len(doc.Routes) > 0 && len(doc.Targets) > 0 {
		return capabilitiesDoc{}, fmt.Errorf("device: capabilities must use routes or targets, not both")
	}
	if len(doc.Routes) == 0 && len(doc.Targets) == 0 {
		return capabilitiesDoc{}, fmt.Errorf("device: capabilities declare no routes or targets")
	}

	return doc, nil
}

func capabilitiesDigest(data []byte, doc capabilitiesDoc) (string, error) {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return "", fmt.Errorf("device: decode capabilities for digest: %w", err)
	}
	var digest string
	var err error
	if len(doc.Routes) > 0 {
		digestValue, prepareErr := canonicalCatalogValue(doc)
		if prepareErr != nil {
			return "", fmt.Errorf("device: prepare capabilities digest: %w", prepareErr)
		}
		digest, err = canonical.DigestDomain(canonical.CapabilityCatalogDomain, digestValue)
	} else {
		// Legacy target catalogs retain their historical unscoped identity.
		digest, err = canonical.Digest(raw)
	}
	if err != nil {
		return "", fmt.Errorf("device: digest capabilities: %w", err)
	}

	return digest, nil
}

func canonicalCatalog(doc capabilitiesDoc) canonicalCatalogDocument {
	routes := make(map[string]canonicalRouteDocument, len(doc.Routes))
	for name, route := range doc.Routes {
		bounds := make(map[string]canonicalBoundDocument, len(route.Bounds))
		for field, bound := range route.Bounds {
			bounds[field] = canonicalBoundDocument(bound)
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
