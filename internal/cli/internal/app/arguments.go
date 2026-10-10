package app

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

// usageExit ends a command with an exit status after the flag package has
// already told the user what was wrong (or printed help for -h).
type usageExit struct{ code int }

func (u usageExit) Error() string { return fmt.Sprintf("usage exit %d", u.code) }

func asUsageExit(err error) (usageExit, bool) {
	var exit usageExit
	return exit, errors.As(err, &exit)
}

// newFlagSet is a flag set whose diagnostics go to stderr and whose failures
// surface as a usageExit, matching flag.ExitOnError without exiting.
func newFlagSet(name string, stderr io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	return fs
}

// parseFlags parses args; a bad flag exits 2 and -h exits 0, as the flag
// package's ExitOnError mode did.
func parseFlags(fs *flag.FlagSet, args []string) error {
	err := fs.Parse(args)
	if err == nil {
		return nil
	}
	if errors.Is(err, flag.ErrHelp) {
		return usageExit{code: 0}
	}
	return usageExit{code: 2}
}

func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// parseTriple splits "a=b@c" by two separators into three parts.
func parseTriple(s, sep1, sep2 string) (string, string, float64, error) {
	index := strings.Index(s, sep1)
	if index < 0 {
		return "", "", 0, fmt.Errorf("bad triple %q", s)
	}
	first := s[:index]
	second, offset := parseOffset(s[index+len(sep1):], sep2)
	return first, second, offset, nil
}

func parseF(s string) float64 {
	var f float64
	_, _ = fmt.Sscanf(s, "%f", &f)
	return f
}

func parseOffset(rest, separator string) (string, float64) {
	if index := strings.Index(rest, separator); index >= 0 {
		return rest[:index], parseF(rest[index+len(separator):])
	}
	return rest, 0
}
