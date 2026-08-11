package run

// Fuzz target: artifact loading (JSON + schema validation) must never panic
// on adversarial bytes.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func FuzzArtifactLoad(f *testing.F) {
	seeds := [][]byte{
		[]byte(`{}`),
		[]byte(`null`),
		[]byte(`{"schema_version":"0.1","sim_version":"0.1.0","seed":1,"sink":"inproc","time_mode":"stepped","command_log":[]}`),
		[]byte(`not json`),
		[]byte(`{"command_log":[{"op":"clock.advance"}]}`),
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, raw []byte) {
		// Neither decoding nor schema validation may panic; an error is a
		// valid outcome for adversarial bytes.
		var doc any
		_ = model.DecodeBytes(raw, &doc)
		_ = model.ValidateRunArtifact(raw)
		var art model.RunArtifact
		_ = jsonDecodeArtifact(raw, &art)
		_ = bytes.Contains(raw, []byte("x"))
	})
}

func jsonDecodeArtifact(raw []byte, art *model.RunArtifact) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(art); err != nil {
		return fmt.Errorf("fuzz artifact decode: %w", err)
	}
	return nil
}
