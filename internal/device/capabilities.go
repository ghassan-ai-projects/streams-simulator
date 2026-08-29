package device

import (
	"encoding/json"
	"fmt"
)

// Capabilities is the device's declared, data-defined capability catalog: which
// targets exist, the single operation each accepts, its numeric parameter
// bounds, and which parameter's positive value means the output is energized.
// It is loaded from JSON — never hard-coded — so no target name, operation, or
// field is baked into the binary (AGENTS.md: effectors are data, not code; no
// per-domain code branches).
type Capabilities struct {
	targets map[string]TargetCapability
}

// TargetCapability is one target's declared capability.
type TargetCapability struct {
	Operation     string
	EnergizeField string
	Bounds        map[string][2]float64
}

type capabilitiesDoc struct {
	ProtocolVersion int `json:"protocol_version"`
	Targets         map[string]struct {
		Operation     string `json:"operation"`
		EnergizeField string `json:"energize_field"`
		Bounds        map[string]struct {
			Min float64 `json:"min"`
			Max float64 `json:"max"`
		} `json:"bounds"`
	} `json:"targets"`
}

// LoadCapabilities parses a device capability catalog from JSON.
func LoadCapabilities(data []byte) (*Capabilities, error) {
	var doc capabilitiesDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("device: decode capabilities: %w", err)
	}
	if doc.ProtocolVersion != ProtocolVersion {
		return nil, fmt.Errorf("device: capabilities protocol_version %d != %d", doc.ProtocolVersion, ProtocolVersion)
	}
	if len(doc.Targets) == 0 {
		return nil, fmt.Errorf("device: capabilities declare no targets")
	}
	caps := &Capabilities{targets: make(map[string]TargetCapability, len(doc.Targets))}
	for name, target := range doc.Targets {
		if target.Operation == "" {
			return nil, fmt.Errorf("device: target %q declares no operation", name)
		}
		if target.EnergizeField == "" {
			return nil, fmt.Errorf("device: target %q declares no energize_field", name)
		}
		bounds := make(map[string][2]float64, len(target.Bounds))
		for field, b := range target.Bounds {
			if b.Max < b.Min {
				return nil, fmt.Errorf("device: target %q field %q has max < min", name, field)
			}
			bounds[field] = [2]float64{b.Min, b.Max}
		}
		if _, ok := bounds[target.EnergizeField]; !ok {
			return nil, fmt.Errorf("device: target %q energize_field %q is not a bounded parameter", name, target.EnergizeField)
		}
		caps.targets[name] = TargetCapability{Operation: target.Operation, EnergizeField: target.EnergizeField, Bounds: bounds}
	}
	return caps, nil
}

func (c *Capabilities) target(name string) (TargetCapability, bool) {
	if c == nil {
		return TargetCapability{}, false
	}
	tc, ok := c.targets[name]
	return tc, ok
}
