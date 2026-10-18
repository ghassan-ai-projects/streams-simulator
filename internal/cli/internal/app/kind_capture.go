package app

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire"
)

const captureUsage = "usage: streamsim kind capture (<nbirth-or-journal-file> | --stdin) [--kinds-dir <dir>]"

type captureOptions struct {
	stdin bool
	kinds listFlag
	paths []string
}

func cmdKindCapture(s *session, args []string) (any, error) {
	options, err := parseCaptureOptions(s, args)
	if err != nil {
		return nil, err
	}
	if options.stdin {
		return captureStdin(s, options)
	}
	return captureFile(options)
}

func captureStdin(s *session, options captureOptions) (any, error) {
	captured, err := plantwire.CaptureStream(s.input(), options.kinds.items())
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return captured, nil
}

func captureFile(options captureOptions) (any, error) {
	captured, err := plantwire.CaptureFile(options.paths[0], options.kinds.items())
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return captured, nil
}

func parseCaptureOptions(s *session, args []string) (captureOptions, error) {
	var options captureOptions
	fs := newFlagSet("kind capture", s.stderr)
	fs.BoolVar(&options.stdin, "stdin", false, "read the NBIRTH record, or a journal holding it, from standard input (at most 1 MiB) instead of a file")
	fs.Var(&options.kinds, kindsDirFlag, "directory of kind files (*.kind.json), repeatable; each captured kind is compared with the file of its name")
	if err := parseFlags(fs, args); err != nil {
		return captureOptions{}, err
	}
	options.paths = fs.Args()
	if options.stdin != (len(options.paths) == 0) || len(options.paths) > 1 {
		return captureOptions{}, usagef(captureUsage)
	}
	return options, nil
}
