package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func validateArtifactDocument(raw []byte, path string) error {
	var doc any
	if err := model.DecodeBytes(raw, &doc); err != nil {
		return fmt.Errorf("run: %s not valid JSON: %w", path, err)
	}
	if err := model.ValidateRunArtifact(raw); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return nil
}

func decodeArtifact(raw []byte) (*model.RunArtifact, error) {
	var artifact model.RunArtifact
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&artifact); err != nil {
		return nil, fmt.Errorf("run: decode artifact: %w", err)
	}
	if err := rejectArtifactTrailingJSON(decoder); err != nil {
		return nil, err
	}
	return &artifact, nil
}

func rejectArtifactTrailingJSON(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("run: artifact has trailing JSON")
		}
		return fmt.Errorf("run: artifact trailing data: %w", err)
	}
	return nil
}
