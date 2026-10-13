// Package entropy is the randomness edge of the canvas bridge: pairing codes
// and session ids from the operating system's cryptographic source.
package entropy

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/hex"
	"fmt"
)

const (
	codeBytes    = 8
	codeLength   = 12
	sessionBytes = 32
)

func PairingCode() (string, error) {
	raw, err := random(codeBytes)
	if err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString(raw)[:codeLength], nil
}

func SessionID() (string, error) {
	raw, err := random(sessionBytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}

func random(n int) ([]byte, error) {
	raw := make([]byte, n)
	if _, err := rand.Read(raw); err != nil {
		return nil, fmt.Errorf("canvasbridge: entropy: %w", err)
	}
	return raw, nil
}
