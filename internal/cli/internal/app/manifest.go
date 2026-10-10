package app

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
	"path/filepath"
	"sort"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/cli/internal/files"
)

func cmdManifest(s *session, args []string) (any, error) {
	options, err := parseManifestOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	manifest, err := buildReleaseManifest(options, s)
	if err != nil {
		return nil, err
	}
	return signAndPublishManifest(manifest, options, s.build)
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
	names, err := files.FileNames(dir)
	if err != nil {
		return nil, fmt.Errorf("manifest: %s dir: %w", kind, err)
	}
	return digestManifestFiles(dir, manifestFileNames(names))
}

func readSigningSeed(path string) ([]byte, error) {
	raw, err := files.Read(path)
	if err != nil {
		return nil, fmt.Errorf("manifest: key: %w", err)
	}
	seed, err := hex.DecodeString(strings.TrimSpace(string(raw)))
	if err != nil || len(seed) != ed25519.SeedSize {
		return nil, fmt.Errorf("manifest: key must be %d hex chars", ed25519.SeedSize*2)
	}
	return seed, nil
}

func manifestFileNames(entries []string) []string {
	names := make([]string, 0, len(entries))
	for _, name := range entries {
		if !strings.HasPrefix(name, ".") {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func digestManifestFiles(dir string, names []string) (map[string]any, error) {
	out := map[string]any{}
	for _, name := range names {
		raw, err := files.Read(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("manifest: read %s: %w", name, err)
		}
		out[name] = canonical.DigestBytes(raw)
	}
	return out, nil
}

func publishManifest(body []byte, signature, out string, build Build) (any, error) {
	doc := map[string]any{"manifest": json.RawMessage(body), "signature": signature}
	raw, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("manifest: %w", err)
	}
	if err := files.Write(out, raw); err != nil {
		return nil, fmt.Errorf("manifest: %w", err)
	}
	return map[string]any{"manifest_written": out, "signed": signature != "", "sim": build.Version + "@" + build.Commit}, nil
}

func signAndPublishManifest(manifest map[string]any, options manifestOptions, build Build) (any, error) {
	body, err := canonical.Marshal(manifest)
	if err != nil {
		return nil, fmt.Errorf("manifest: canonical: %w", err)
	}
	sig, err := signManifest(body, options.keyFile)
	if err != nil {
		return nil, err
	}
	return publishManifest(body, sig, options.out, build)
}
