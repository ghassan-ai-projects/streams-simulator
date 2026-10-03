package cli

import (
	"flag"
	"fmt"
	"runtime"
	"time"
)

type manifestOptions struct{ out, domainsDir, adaptersDir, author, reviewer, keyFile string }

func parseManifestOptions(args []string) (manifestOptions, error) {
	var options manifestOptions
	fs := flag.NewFlagSet("manifest", flag.ExitOnError)
	registerManifestFlags(fs, &options)
	if err := fs.Parse(args); err != nil {
		return manifestOptions{}, fmt.Errorf("streamsim: %w", err)
	}
	if options.author == "" || options.reviewer == "" {
		return manifestOptions{}, fmt.Errorf("manifest requires --author and --reviewer (independent review identity)")
	}
	return options, nil
}

func buildReleaseManifest(options manifestOptions) (map[string]any, error) {
	domains, err := fileDigests(options.domainsDir, "domain")
	if err != nil {
		return nil, err
	}
	adapters, err := fileDigests(options.adaptersDir, "adapter")
	if err != nil {
		return nil, err
	}
	return releaseManifestIdentity(options, domains, adapters), nil
}

func releaseManifestIdentity(options manifestOptions, domains, adapters map[string]any) map[string]any {
	return map[string]any{"schema_version": "release-manifest-v0.1",
		"sim":     map[string]any{"version": Version, "commit": Commit, "go": runtime.Version()},
		"domains": domains, "adapters": adapters,
		"consumer": map[string]any{"name": "streamsim-refconsumer", "version": "0.1.0"},
		"suite":    map[string]any{"negative_class_fraction": 0.4, "trivial_cutoff": 0.9},
		"author":   options.author, "reviewer": options.reviewer, "created_at": time.Now().UTC().Format(time.RFC3339)}
}

func registerManifestFlags(fs *flag.FlagSet, options *manifestOptions) {
	fs.StringVar(&options.out, "out", "release-manifest.json", "output path")
	fs.StringVar(&options.domainsDir, "domains-dir", "domains", "domains directory")
	fs.StringVar(&options.adaptersDir, "adapters-dir", "adapters", "adapters directory")
	fs.StringVar(&options.author, "author", "", "release author identity (name <email>)")
	fs.StringVar(&options.reviewer, "reviewer", "", "independent review identity (name <email>)")
	fs.StringVar(&options.keyFile, "key", "", "ed25519 private key (64 hex seed chars) to sign the manifest digest")
}
