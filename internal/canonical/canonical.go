// Package canonical implements RFC 8785 canonical JSON serialization,
// which every digest the simulator computes is based on. Canonical JSON
// makes the digest a pure function of the data, independent of map
// iteration order, number formatting, or whitespace.
//
// Supported input values are exactly what the simulator's own JSON path
// produces: nil, bool, json.Number, float64, int/uint of any width, string,
// []any, map[string]any, and time.Time (rendered RFC 3339 with nanoseconds).
// Anything else is rejected rather than guessed at.
//
// One deliberate departure: a json.Number integer literal is written as it
// came, not through the IEEE-754 double form RFC 8785 prescribes. Seeds and
// nanosecond timestamps are 64-bit integers; the double form would map
// distinct values above 2^53 to the same text, and so to the same digest.
// Non-integer numbers follow the RFC.
package canonical

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

func Marshal(v any) ([]byte, error) {
	var b strings.Builder
	if err := writeValue(&b, v); err != nil {
		return nil, fmt.Errorf("canonical: %w", err)
	}
	return []byte(b.String()), nil
}

func MarshalString(v any) (string, error) {
	b, err := Marshal(v)
	return string(b), err
}

const CapabilityCatalogDomain = "situation-runtime/capability-catalog/v1\n"

func Digest(v any) (string, error) {
	return DigestDomain("", v)
}

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

func DigestBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:])
}
