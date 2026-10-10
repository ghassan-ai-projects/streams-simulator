// Package domain holds the device emulator's rules: the wire record codec and
// contract schemas, the capability catalog, admission and the boot-identity,
// expiry, allowlist, bounds and idempotency checks, the protocol fault
// schedule, and the plant port the world drives. It performs no I/O; the
// socket and the frame gate are the uds edge's job.
package domain

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ghassan-ai-projects/streams-simulator/internal/device/contract"
	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
)

// ProtocolVersion is the device wire protocol version this emulator speaks. It
// must match the vendored contract (contract/schemas + contract/SOURCE.md).
const ProtocolVersion = 1

// MaxFrameBytes bounds a single decoded record, matching the Agentic Stream
// codec so the two ends agree on the oversize-frame boundary.
const MaxFrameBytes = 64 * 1024

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
	raw, err := contract.Schemas.ReadFile("schemas/" + file)
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
