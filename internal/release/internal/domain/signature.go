package domain

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

func Seed(text []byte) ([]byte, error) {
	seed, err := hex.DecodeString(strings.TrimSpace(string(text)))
	if err != nil || len(seed) != ed25519.SeedSize {
		return nil, fmt.Errorf("manifest: key must be %d hex chars", ed25519.SeedSize*2)
	}
	return seed, nil
}

func Sign(body, seed []byte) string {
	priv := ed25519.NewKeyFromSeed(seed)
	digest := sha256.Sum256(body)
	return hex.EncodeToString(ed25519.Sign(priv, digest[:]))
}
