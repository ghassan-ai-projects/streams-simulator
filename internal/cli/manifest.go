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
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
)

func cmdManifest(args []string) error {
	fs := flag.NewFlagSet("manifest", flag.ExitOnError)
	out := fs.String("out", "release-manifest.json", "output path")
	domainsDir := fs.String("domains-dir", "domains", "domains directory")
	adaptersDir := fs.String("adapters-dir", "adapters", "adapters directory")
	author := fs.String("author", "", "release author identity (name <email>)")
	reviewer := fs.String("reviewer", "", "independent review identity (name <email>)")
	keyFile := fs.String("key", "", "ed25519 private key (64 hex seed chars) to sign the manifest digest")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	if *author == "" || *reviewer == "" {
		return fmt.Errorf("manifest requires --author and --reviewer (independent review identity)")
	}
	domains, err := fileDigests(*domainsDir, "domain")
	if err != nil {
		return err
	}
	adapters, err := fileDigests(*adaptersDir, "adapter")
	if err != nil {
		return err
	}
	manifest := map[string]any{
		"schema_version": "release-manifest-v0.1",
		"sim": map[string]any{
			"version": Version, "commit": Commit, "go": runtime.Version(),
		},
		"domains":  domains,
		"adapters": adapters,
		"consumer": map[string]any{"name": "streamsim-refconsumer", "version": "0.1.0"},
		"suite": map[string]any{
			"negative_class_fraction": 0.4,
			"trivial_cutoff":          0.9,
		},
		"author":     *author,
		"reviewer":   *reviewer,
		"created_at": time.Now().UTC().Format(time.RFC3339),
	}
	// Canonical JSON is the signed body; the signature sits alongside.
	body, err := canonical.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("manifest: canonical: %w", err)
	}
	sig := ""
	if *keyFile != "" {
		raw, err := os.ReadFile(*keyFile)
		if err != nil {
			return fmt.Errorf("manifest: key: %w", err)
		}
		seed, err := hex.DecodeString(strings.TrimSpace(string(raw)))
		if err != nil || len(seed) != ed25519.SeedSize {
			return fmt.Errorf("manifest: key must be %d hex chars", ed25519.SeedSize*2)
		}
		priv := ed25519.NewKeyFromSeed(seed)
		digest := sha256.Sum256(body)
		sig = hex.EncodeToString(ed25519.Sign(priv, digest[:]))
	}
	doc := map[string]any{
		"manifest":  json.RawMessage(body),
		"signature": sig,
	}
	raw, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("manifest: %w", err)
	}
	if err := os.WriteFile(*out, raw, 0o600); err != nil {
		return fmt.Errorf("manifest: %w", err)
	}
	return printJSON(map[string]any{"manifest_written": *out, "signed": sig != "", "sim": Version + "@" + Commit})
}

// fileDigests digests every file in dir, keyed by base name. Values are
// map[string]any for the canonical marshaler.
func fileDigests(dir, kind string) (map[string]any, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("manifest: %s dir: %w", kind, err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
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
