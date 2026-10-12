package app

import (
	"flag"
	"io"
)

type manifestOptions struct{ out, domainsDir, adaptersDir, author, reviewer, keyFile string }

func parseManifestOptions(args []string, stderr io.Writer) (manifestOptions, error) {
	var options manifestOptions
	fs := newFlagSet("manifest", stderr)
	registerManifestFlags(fs, &options)
	if err := parseOnlyFlags(fs, args); err != nil {
		return manifestOptions{}, err
	}
	if options.author == "" || options.reviewer == "" {
		return manifestOptions{}, usagef("manifest requires --author and --reviewer (independent review identity)")
	}
	return options, nil
}

func registerManifestFlags(fs *flag.FlagSet, options *manifestOptions) {
	fs.StringVar(&options.out, "out", "release-manifest.json", "output path")
	fs.StringVar(&options.domainsDir, "domains-dir", "domains", "domains directory")
	fs.StringVar(&options.adaptersDir, "adapters-dir", "adapters", "adapters directory")
	fs.StringVar(&options.author, "author", "", "release author identity (name <email>)")
	fs.StringVar(&options.reviewer, "reviewer", "", "independent review identity (name <email>)")
	fs.StringVar(&options.keyFile, "key", "", "ed25519 private key (64 hex seed chars) to sign the manifest digest")
}
