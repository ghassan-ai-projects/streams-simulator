package device

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
)

// EncodeRecord validates one device record and returns exactly one canonical
// NDJSON line (sorted keys, trailing newline) — the byte layout the Agentic
// Stream codec also produces. Raw serial framing is layered above this by the
// gateway.
func EncodeRecord(document map[string]any) ([]byte, error) {
	if document == nil {
		return nil, fmt.Errorf("device: record is required")
	}
	if err := validateRecord(document); err != nil {
		return nil, err
	}
	encoded, err := canonical.Marshal(document)
	if err != nil {
		return nil, fmt.Errorf("device: encode record: %w", err)
	}
	encoded = append(encoded, '\n')
	if len(encoded) > maxFrameBytes {
		return nil, fmt.Errorf("device: record exceeds %d bytes", maxFrameBytes)
	}
	return encoded, nil
}

// DecodeRecord decodes and validates exactly one device record from a frame. It
// fails closed on empty, oversized, non-object, trailing-content, unknown-type,
// and schema-invalid input — an invalid frame must never yield a usable record.
func DecodeRecord(frame []byte) (map[string]any, error) {
	if len(frame) == 0 {
		return nil, fmt.Errorf("device: frame is empty")
	}
	if len(frame) > maxFrameBytes {
		return nil, fmt.Errorf("device: frame exceeds %d bytes", maxFrameBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(frame))
	var document map[string]any
	if err := decoder.Decode(&document); err != nil {
		return nil, fmt.Errorf("device: decode frame: %w", err)
	}
	if document == nil {
		return nil, fmt.Errorf("device: frame must be a JSON object")
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("device: frame contains trailing JSON")
		}
		return nil, fmt.Errorf("device: decode trailing frame data: %w", err)
	}
	if err := validateRecord(document); err != nil {
		return nil, err
	}
	return document, nil
}

func validateRecord(document map[string]any) error {
	messageType, ok := document["message_type"].(string)
	if !ok || messageType == "" {
		return fmt.Errorf("device: message_type is required")
	}
	schema, err := schemaForMessageType(messageType)
	if err != nil {
		return err
	}
	if errs := schema.Validate(document); len(errs) > 0 {
		return fmt.Errorf("device: validate %s record: %w", messageType, jsonschema.Error(errs[0]))
	}
	return nil
}
