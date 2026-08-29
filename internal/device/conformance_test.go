package device

import (
	"os"
	"path/filepath"
	"testing"
)

// The vendored fixtures under contract/conformance/v1/ are copied verbatim from
// Agentic Stream (contract/SOURCE.md). Proving the emulator decodes every valid
// frame and rejects every invalid one is the cross-repo handshake: it means the
// two ends agree on the wire, byte for byte, and cannot silently diverge.

func TestConformanceValidFramesDecode(t *testing.T) {
	files, err := filepath.Glob("contract/conformance/v1/valid/*.json")
	if err != nil || len(files) == 0 {
		t.Fatalf("no valid conformance fixtures found: %v", err)
	}
	for _, file := range files {
		file := file
		t.Run(filepath.Base(file), func(t *testing.T) {
			frame, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := DecodeRecord(frame); err != nil {
				t.Fatalf("valid fixture %s must decode: %v", filepath.Base(file), err)
			}
		})
	}
}

func TestConformanceInvalidFramesRejected(t *testing.T) {
	files, err := filepath.Glob("contract/conformance/v1/invalid/*.json")
	if err != nil || len(files) == 0 {
		t.Fatalf("no invalid conformance fixtures found: %v", err)
	}
	for _, file := range files {
		file := file
		t.Run(filepath.Base(file), func(t *testing.T) {
			frame, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := DecodeRecord(frame); err == nil {
				t.Fatalf("invalid fixture %s must be rejected, but decoded", filepath.Base(file))
			}
		})
	}
}
