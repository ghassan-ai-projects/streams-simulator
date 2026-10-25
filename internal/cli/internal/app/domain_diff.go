package app

import (
	"fmt"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/component"
)

const (
	diffUsage      = "usage: streamsim domain diff <a> <b> [--format json|markdown] (each a domain file or a run folder)"
	diffFormatJSON = "json"
	diffFormatText = "markdown"
)

func cmdDiff(s *session, args []string) (any, error) {
	format, sides, err := parseDiffOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	diff, err := component.DiffDomains(sides[0], sides[1])
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	if format == diffFormatText {
		return map[string]any{"format": format, "markdown": string(component.DiffMarkdown(diff))}, nil
	}
	return diff, nil
}

func parseDiffOptions(args []string, stderr io.Writer) (string, []string, error) {
	var format string
	fs := newFlagSet("domain diff", stderr)
	fs.StringVar(&format, "format", diffFormatJSON, "diff format: json|markdown")
	if err := parseFlags(fs, args); err != nil {
		return "", nil, err
	}
	if fs.NArg() != 2 || (format != diffFormatJSON && format != diffFormatText) {
		return "", nil, usagef(diffUsage)
	}
	return format, fs.Args(), nil
}
