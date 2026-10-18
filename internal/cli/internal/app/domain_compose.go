package app

import (
	"fmt"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/cli/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/component"
)

type composeOptions struct {
	composition, out string
	dirs, kinds      listFlag
}

func cmdCompose(s *session, args []string) (any, error) {
	options, err := parseComposeOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	raw, err := files.Read(options.composition)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	library, err := component.NewLibrary(s.componentDirs(options.dirs))
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return composeDomain(library, raw, options)
}

func parseComposeOptions(args []string, stderr io.Writer) (composeOptions, error) {
	var options composeOptions
	fs := newFlagSet("domain compose", stderr)
	fs.StringVar(&options.out, "out", "", "file to write the composed domain to (default: print it inside the output document)")
	registerComponentDirs(fs, &options.dirs)
	registerKindsDirs(fs, &options.kinds)
	if err := parseFlags(fs, args); err != nil {
		return composeOptions{}, err
	}
	if fs.NArg() != 1 {
		return composeOptions{}, usagef("usage: streamsim domain compose <composition.json> [--out <file>] [--components-dir <dir>] [--kinds-dir <dir>]")
	}
	options.composition = fs.Arg(0)
	return options, nil
}

type composedDomain struct {
	component.DomainSummary
	Composed string `json:"composed,omitempty"`
	Domain   any    `json:"domain,omitempty"`
}

func composeDomain(library *component.Library, raw []byte, options composeOptions) (any, error) {
	composed, err := component.Compose(library, raw, options.composition)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	if _, err := checkImplements(composed.Domain.Spec, options.composition, options.kinds); err != nil {
		return nil, err
	}
	result := composedDomain{DomainSummary: composed.Summary}
	if err := placeComposed(&result, composed, options.out); err != nil {
		return nil, err
	}
	return result, nil
}

func placeComposed(result *composedDomain, composed *component.Composed, out string) error {
	if out == "" {
		result.Domain = composed.Domain.Spec
		return nil
	}
	if err := files.Write(out, composed.Bytes); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	result.Composed = out
	return nil
}
