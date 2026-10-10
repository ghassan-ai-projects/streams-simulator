package domain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The vendored fixtures under contract/conformance/v1/ are copied verbatim from
// Agentic Stream (contract/SOURCE.md). Proving the emulator decodes every valid
// frame and rejects every invalid one is the cross-repo handshake: it means the
// two ends agree on the wire, byte for byte, and cannot silently diverge.

func TestConformanceValidFramesDecode(t *testing.T) {
	t.Parallel()
	files, err := filepath.Glob("../../contract/conformance/v1/valid/*.json")
	if err != nil || len(files) == 0 {
		t.Fatalf("no valid conformance fixtures found: %v", err)
	}
	for _, file := range files {
		file := file
		t.Run(filepath.Base(file), func(t *testing.T) {
			t.Parallel()
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
	t.Parallel()
	files, err := filepath.Glob("../../contract/conformance/v1/invalid/*.json")
	if err != nil || len(files) == 0 {
		t.Fatalf("no invalid conformance fixtures found: %v", err)
	}
	rejections := map[string]string{
		"command-bad-idempotency-key.json":           "validate command record: /idempotency_key: string does not match pattern",
		"command-expires-after-ms-zero.json":         "validate command record: /expires_after_ms: number 0 below minimum 1",
		"command-missing-operation.json":             `validate command record: missing required property "operation"`,
		"command-missing-policy-digest.json":         `validate command record: missing required property "policy_digest"`,
		"command-protocol-version-out-of-range.json": "validate command record: /protocol_version: value does not match const",
		"command-unknown-field.json":                 `validate command record: additional property "pin" not allowed`,
		"command-wrong-message-type.json":            `validate receipt record: missing required property "accepted"`,
		"receipt-missing-boot-id.json":               `validate receipt record: missing required property "boot_id"`,
		"receipt-rejected-without-reject-code.json":  `validate receipt record: missing required property "reject_code"`,
		"receipt-unknown-reject-code.json":           "validate receipt record: /reject_code: value not in enum",
		"result-bad-status.json":                     "validate result record: /status: value not in enum",
		"state-bad-firmware-digest.json":             "validate state record: /firmware_digest: string does not match pattern",
		"state-missing-capability-digest.json":       `validate state record: missing required property "capability_digest"`,
	}
	for _, file := range files {
		file := file
		name := filepath.Base(file)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			want, known := rejections[name]
			if !known {
				t.Fatalf("invalid fixture %s has no expected rejection: add it to the table", name)
			}
			frame, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := DecodeRecord(frame); err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("invalid fixture %s: err = %v, want %q", name, err, want)
			}
		})
	}
}
