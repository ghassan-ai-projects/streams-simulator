package cli

// The release manifest: simulator, domain, adapter, suite, toolchain and
// consumer digests plus author and independent review identity, optionally
// signed with stdlib ed25519. One artifact pins exactly what a benchmark
// release claims to be.

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
)

func cmdManifest(args []string) error {
	options, err := parseManifestOptions(args)
	if err != nil {
		return err
	}
	manifest, err := buildReleaseManifest(options)
	if err != nil {
		return err
	}
	return signAndPublishManifest(manifest, options)
}

func signManifest(body []byte, keyFile string) (string, error) {
	if keyFile == "" {
		return "", nil
	}
	seed, err := readSigningSeed(keyFile)
	if err != nil {
		return "", err
	}
	priv := ed25519.NewKeyFromSeed(seed)
	digest := sha256.Sum256(body)
	return hex.EncodeToString(ed25519.Sign(priv, digest[:])), nil
}

// fileDigests digests every file in dir, keyed by base name. Values are
// map[string]any for the canonical marshaler.
func fileDigests(dir, kind string) (map[string]any, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("manifest: %s dir: %w", kind, err)
	}
	return digestManifestFiles(dir, manifestFileNames(entries))
}

func readSigningSeed(path string) ([]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("manifest: key: %w", err)
	}
	seed, err := hex.DecodeString(strings.TrimSpace(string(raw)))
	if err != nil || len(seed) != ed25519.SeedSize {
		return nil, fmt.Errorf("manifest: key must be %d hex chars", ed25519.SeedSize*2)
	}
	return seed, nil
}

func manifestFileNames(entries []os.DirEntry) []string {
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names
}

func digestManifestFiles(dir string, names []string) (map[string]any, error) {
	out := map[string]any{}
	for _, name := range names {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("manifest: read %s: %w", name, err)
		}
		out[name] = canonical.DigestBytes(raw)
	}
	return out, nil
}

func publishManifest(body []byte, signature, out string) error {
	doc := map[string]any{"manifest": json.RawMessage(body), "signature": signature}
	raw, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("manifest: %w", err)
	}
	if err := os.WriteFile(out, raw, 0o600); err != nil {
		return fmt.Errorf("manifest: %w", err)
	}
	return printJSON(map[string]any{"manifest_written": out, "signed": signature != "", "sim": Version + "@" + Commit})
}

func signAndPublishManifest(manifest map[string]any, options manifestOptions) error {
	body, err := canonical.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("manifest: canonical: %w", err)
	}
	sig, err := signManifest(body, options.keyFile)
	if err != nil {
		return err
	}
	return publishManifest(body, sig, options.out)
}
