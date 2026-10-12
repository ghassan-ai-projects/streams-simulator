package domain

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestSignCoversTheDigestOfTheCanonicalBody(t *testing.T) {
	t.Parallel()
	seed := make([]byte, ed25519.SeedSize)
	body := []byte(`{"test":true}`)
	decoded, err := hex.DecodeString(Sign(body, seed))
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(body)
	public, ok := ed25519.NewKeyFromSeed(seed).Public().(ed25519.PublicKey)
	if !ok || !ed25519.Verify(public, digest[:], decoded) {
		t.Fatal("signature must cover the canonical body digest")
	}
}

func TestSeedAcceptsExactlyTheHexOfAFullSeed(t *testing.T) {
	t.Parallel()
	good := []byte(hex.EncodeToString(make([]byte, ed25519.SeedSize)) + "\n")
	if seed, err := Seed(good); err != nil || len(seed) != ed25519.SeedSize {
		t.Fatalf("seed = %x, %v", seed, err)
	}
	for name, text := range map[string]string{"short": "abcd", "not hex": strings.Repeat("zz", ed25519.SeedSize), "empty": ""} {
		if _, err := Seed([]byte(text)); err == nil || err.Error() != "manifest: key must be 64 hex chars" {
			t.Fatalf("%s: err = %v", name, err)
		}
	}
}
