package app

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

var fixedNow = time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)

func options(t *testing.T) Options {
	t.Helper()
	return Options{Out: filepath.Join(t.TempDir(), "manifest.json"), DomainsDir: testsupport.DomainsDir(),
		AdaptersDir: testsupport.AdaptersDir(), Author: "A <a@x>", Reviewer: "R <r@x>",
		Version: "v-test", Commit: "c-test", Now: func() time.Time { return fixedNow }}
}

func readDocument(t *testing.T, path string) (manifest map[string]any, signature string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Manifest  map[string]any `json:"manifest"`
		Signature string         `json:"signature"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	return doc.Manifest, doc.Signature
}

func TestPublishRecordsBuildClockAndDigestsOfEveryInputFile(t *testing.T) {
	t.Parallel()
	o := options(t)
	result, err := Publish(o)
	if err != nil || result.Path != o.Out || result.Signed || result.Sim != "v-test@c-test" {
		t.Fatalf("result = %+v, %v", result, err)
	}
	manifest, signature := readDocument(t, o.Out)
	domains, _ := manifest["domains"].(map[string]any)
	raw, err := os.ReadFile(testsupport.Domain("aquaculture-pond"))
	if err != nil {
		t.Fatal(err)
	}
	if signature != "" || manifest["created_at"] != "2026-10-10T12:00:00Z" || domains["aquaculture-pond.domain.json"] != canonical.DigestBytes(raw) {
		t.Fatalf("manifest = %v, signature %q", manifest, signature)
	}
}

func TestPublishSignsTheCanonicalBodyWhenAKeyIsGiven(t *testing.T) {
	t.Parallel()
	o := options(t)
	seed := make([]byte, ed25519.SeedSize)
	o.KeyFile = filepath.Join(t.TempDir(), "key")
	if err := os.WriteFile(o.KeyFile, []byte(hex.EncodeToString(seed)), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := Publish(o)
	if err != nil || !result.Signed {
		t.Fatalf("result = %+v, %v", result, err)
	}
	manifest, signature := readDocument(t, o.Out)
	body, err := canonical.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := hex.DecodeString(signature)
	public, ok := ed25519.NewKeyFromSeed(seed).Public().(ed25519.PublicKey)
	if err != nil || !ok || !ed25519.Verify(public, sha256Sum(body), decoded) {
		t.Fatalf("signature %q does not cover the canonical body (%v)", signature, err)
	}
}

func TestPublishNamesWhatItCannotRead(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "absent")
	for name, tc := range map[string]struct {
		change func(*Options)
		want   string
	}{
		"domains dir":  {func(o *Options) { o.DomainsDir = missing }, "manifest: domain dir: "},
		"adapters dir": {func(o *Options) { o.AdaptersDir = missing }, "manifest: adapter dir: "},
		"key file":     {func(o *Options) { o.KeyFile = missing }, "manifest: key: "},
		"short key":    {func(o *Options) { o.KeyFile = shortKey(t) }, "manifest: key must be 64 hex chars"},
		"output":       {func(o *Options) { o.Out = filepath.Join(missing, "m.json") }, "manifest: "},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			o := options(t)
			tc.change(&o)
			if _, err := Publish(o); err == nil || !strings.HasPrefix(err.Error(), tc.want) {
				t.Fatalf("err = %v, want prefix %q", err, tc.want)
			}
		})
	}
}

func shortKey(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "short")
	if err := os.WriteFile(path, []byte("abcd"), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func sha256Sum(body []byte) []byte {
	digest := sha256.Sum256(body)
	return digest[:]
}
