package domain

import (
	"errors"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestDecodeDocumentChecksTheContractThenDecodes(t *testing.T) {
	t.Parallel()
	var verdict model.Verdict
	err := DecodeDocument([]byte(`{}`), "verdict.json", model.ValidateVerdict, &verdict)
	var invalid *model.ValidationError
	if !errors.As(err, &invalid) || invalid.Source != "verdict.json" {
		t.Fatalf("err = %v, want a contract violation naming verdict.json", err)
	}
	accept := func([]byte) error { return nil }
	if err := DecodeDocument([]byte("not json"), "label.json", accept, &verdict); err == nil || !strings.HasPrefix(err.Error(), "decode label.json: ") {
		t.Fatalf("an undecodable document: err = %v", err)
	}
}
