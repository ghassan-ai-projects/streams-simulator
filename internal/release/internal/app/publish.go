// Package app is the use-case layer of the release module: it digests the
// release inputs, builds and signs the manifest and writes it.
package app

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/release/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/release/internal/files"
)

type Options struct {
	Out, DomainsDir, AdaptersDir, Author, Reviewer, KeyFile string
	Version, Commit                                         string
	Now                                                     func() time.Time
}

type Result struct {
	Path   string
	Signed bool
	Sim    string
}

func Publish(options Options) (Result, error) {
	manifest, err := build(options)
	if err != nil {
		return Result{}, err
	}
	body, err := domain.Canonical(manifest)
	if err != nil {
		return Result{}, err
	}
	signature, err := sign(body, options.KeyFile)
	if err != nil {
		return Result{}, err
	}
	return write(body, signature, options)
}

func build(options Options) (map[string]any, error) {
	domains, err := fileDigests(options.DomainsDir, "domain")
	if err != nil {
		return nil, err
	}
	adapters, err := fileDigests(options.AdaptersDir, "adapter")
	if err != nil {
		return nil, err
	}
	identity := domain.Identity{Version: options.Version, Commit: options.Commit, Author: options.Author,
		Reviewer: options.Reviewer, CreatedAt: options.Now().UTC().Format(time.RFC3339)}
	return domain.Body(identity, domains, adapters), nil
}

func fileDigests(dir, kind string) (map[string]any, error) {
	names, err := files.Names(dir)
	if err != nil {
		return nil, fmt.Errorf("manifest: %s dir: %w", kind, err)
	}
	return digestFiles(dir, domain.FileNames(names))
}

func digestFiles(dir string, names []string) (map[string]any, error) {
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

func sign(body []byte, keyFile string) (string, error) {
	if keyFile == "" {
		return "", nil
	}
	text, err := files.Read(keyFile)
	if err != nil {
		return "", fmt.Errorf("manifest: key: %w", err)
	}
	seed, err := domain.Seed(text)
	if err != nil {
		return "", err
	}
	return domain.Sign(body, seed), nil
}

func write(body []byte, signature string, options Options) (Result, error) {
	doc, err := domain.Document(body, signature)
	if err != nil {
		return Result{}, err
	}
	if err := files.Write(options.Out, doc); err != nil {
		return Result{}, fmt.Errorf("manifest: %w", err)
	}
	return Result{Path: options.Out, Signed: signature != "", Sim: options.Version + "@" + options.Commit}, nil
}
