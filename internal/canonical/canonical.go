// Package canonical implements RFC 8785 canonical JSON serialization,
// which every digest the simulator computes is based on. Canonical JSON
// makes the digest a pure function of the data, independent of map
// iteration order, number formatting, or whitespace.
//
// Supported input values are exactly what the simulator's own JSON path
// produces: nil, bool, json.Number, float64, int/uint of any width, string,
// []any, map[string]any, and time.Time (rendered RFC 3339 with nanoseconds).
// Anything else is rejected rather than guessed at.
package canonical

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// Marshal returns the RFC 8785 canonical serialization of v.
//
// Rules applied (RFC 8785 §3.2):
//   - object keys sorted lexicographically by UTF-16 code units;
//   - numbers serialized per ECMAScript Number::toString (shortest round-trip,
//     no exponent outside the ES bounds, no trailing ".0");
//   - strings escaped per JSON with U+2028/U+2029 additionally escaped;
//   - no insignificant whitespace;
//   - only finite numbers are representable; NaN and Inf are errors.
func Marshal(v any) ([]byte, error) {
	var b strings.Builder
	if err := writeValue(&b, v); err != nil {
		return nil, fmt.Errorf("canonical: %w", err)
	}
	return []byte(b.String()), nil
}

// MarshalString is Marshal returning a string.
func MarshalString(v any) (string, error) {
	b, err := Marshal(v)
	return string(b), err
}

// CapabilityCatalogDomain is the domain separator used by the paired device
// capability catalog contract. The trailing newline is part of the contract.
const CapabilityCatalogDomain = "situation-runtime/capability-catalog/v1\n"

// Digest returns the RFC 8785 canonical digest of v as "sha256:<hex>".
func Digest(v any) (string, error) {
	return DigestDomain("", v)
}

// DigestDomain returns a domain-separated RFC 8785 canonical digest of v as
// "sha256:<hex>". The domain is prepended to the canonical JSON bytes before
// hashing. An empty domain preserves the simulator's historical unscoped
// digest behavior.
func DigestDomain(domain string, v any) (string, error) {
	b, err := Marshal(v)
	if err != nil {
		return "", err
	}
	preimage := make([]byte, 0, len(domain)+len(b))
	preimage = append(preimage, domain...)
	preimage = append(preimage, b...)
	sum := sha256.Sum256(preimage)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// DigestBytes returns "sha256:<hex>" of raw bytes.
func DigestBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:])
}
