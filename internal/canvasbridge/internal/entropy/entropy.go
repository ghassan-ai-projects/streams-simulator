// Package entropy is the randomness edge of the canvas bridge: pairing codes,
// session ids and the run ids of connected sessions from the operating
// system's cryptographic source, which
// crypto/rand reads without ever failing (it ends the process instead).
package entropy

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/hex"
)

const (
	codeBytes    = 8
	codeLength   = 12
	sessionBytes = 32
	runBytes     = 4
)

func PairingCode() string {
	return base32.StdEncoding.EncodeToString(random(codeBytes))[:codeLength]
}

func SessionID() string {
	return hex.EncodeToString(random(sessionBytes))
}

func RunID() string {
	return "c-" + hex.EncodeToString(random(runBytes))
}

func random(n int) []byte {
	raw := make([]byte, n)
	_, _ = rand.Read(raw)
	return raw
}
