// Package device is a wire-faithful emulator of the serial device at the far
// end of the Agentic Stream serial-effector boundary (Real-World Sensor HIL-0).
//
// It speaks the device wire contract — command / receipt / result / state
// records — owned by Agentic Stream (see contract/SOURCE.md), enforces the
// safety envelope a real firmware+gateway would (boot identity, expiry, target
// and operation allowlist, parameter bounds, idempotency), and drives a plant
// so that state and verification reflect what the world actually did rather
// than what a command asked for. It is a test instrument: the effect a command
// has on the plant is the ground truth, and an acknowledgement is never proof
// of that effect.
//
// This package owns no transport framing beyond one-record-per-line NDJSON;
// raw serial bytes, reconnection, and device identity remain a gateway concern.
package device

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
)

//go:embed contract/schemas/*.json
var schemaFiles embed.FS

// ProtocolVersion is the device wire protocol version this emulator speaks. It
// must match the vendored contract (contract/schemas + contract/SOURCE.md).
const ProtocolVersion = 1

// maxFrameBytes bounds a single decoded record, matching the Agentic Stream
// codec so the two ends agree on the oversize-frame boundary.
const maxFrameBytes = 64 * 1024

var (
	schemaMu sync.Mutex
	compiled = map[string]*jsonschema.Schema{}
)

// schemaForMessageType returns the compiled schema for a device message_type.
func schemaForMessageType(messageType string) (*jsonschema.Schema, error) {
	file, err := messageSchemaFile(messageType)
	if err != nil {
		return nil, err
	}
	return cachedMessageSchema(messageType, file)
}

func messageSchemaFile(messageType string) (string, error) {
	file, ok := map[string]string{"command": "device-command-v1.json", "receipt": "device-receipt-v1.json",
		"result": "device-result-v1.json", "state": "device-state-v1.json"}[messageType]
	if !ok {
		return "", fmt.Errorf("device: unsupported message_type %q", messageType)
	}
	return file, nil
}

func compileMessageSchema(file string) (*jsonschema.Schema, error) {
	doc, err := decodeMessageSchema(file)
	if err != nil {
		return nil, err
	}
	schema, err := jsonschema.Compile(doc)
	if err != nil {
		return nil, fmt.Errorf("device: compile schema %s: %w", file, err)
	}
	return schema, nil
}

func decodeMessageSchema(file string) (any, error) {
	raw, err := schemaFiles.ReadFile("contract/schemas/" + file)
	if err != nil {
		return nil, fmt.Errorf("device: read embedded schema %s: %w", file, err)
	}
	var doc any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("device: decode schema %s: %w", file, err)
	}
	return doc, nil
}

func cachedMessageSchema(messageType, file string) (*jsonschema.Schema, error) {
	schemaMu.Lock()
	defer schemaMu.Unlock()
	if schema := compiled[messageType]; schema != nil {
		return schema, nil
	}
	schema, err := compileMessageSchema(file)
	if err != nil {
		return nil, err
	}
	compiled[messageType] = schema
	return schema, nil
}
