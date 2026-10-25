package app

import (
	"flag"
	"fmt"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/component"
)

const lineageUsage = "usage: streamsim domain lineage --domains-dir <dir> [--kinds-dir <dir>] [--twins-dir <dir>] [--compositions-dir <dir>] [--runs-dir <dir>] [--domain <id>] [--components-dir <dir>]"

type lineageOptions struct {
	dirs      component.WorkspaceDirs
	domain    string
	libraries listFlag
}

func cmdLineage(s *session, args []string) (any, error) {
	options, err := parseLineageOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	library, err := component.NewLibrary(s.componentDirs(options.libraries))
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	lineage, err := component.Lineage(library, options.dirs, options.domain)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return lineage, nil
}

func parseLineageOptions(args []string, stderr io.Writer) (lineageOptions, error) {
	var options lineageOptions
	fs := newFlagSet("domain lineage", stderr)
	registerLineageFlags(fs, &options)
	if err := parseFlags(fs, args); err != nil {
		return lineageOptions{}, err
	}
	if fs.NArg() != 0 || options.dirs.Domains == "" {
		return lineageOptions{}, usagef(lineageUsage)
	}
	return options, nil
}

func registerLineageFlags(fs *flag.FlagSet, options *lineageOptions) {
	registerDomainsDir(fs, &options.dirs.Domains, "")
	fs.StringVar(&options.dirs.Kinds, "kinds-dir", "", "directory of kind files (*.kind.json)")
	fs.StringVar(&options.dirs.Twins, "twins-dir", "", "directory of twin choices files (*.twin.json)")
	fs.StringVar(&options.dirs.Compositions, "compositions-dir", "", "directory of compositions (*.composition.json)")
	fs.StringVar(&options.dirs.Runs, "runs-dir", "", "directory of run folders")
	fs.StringVar(&options.domain, "domain", "", "report this domain id only")
	registerComponentDirs(fs, &options.libraries)
}
