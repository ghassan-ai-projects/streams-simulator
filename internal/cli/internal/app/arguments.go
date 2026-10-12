package app

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

// usageExit ends a command with an exit status after the flag package has
// already told the user what was wrong (or printed help for -h). message is
// the flag package's diagnosis, empty for help.
type usageExit struct {
	code    int
	message string
}

func (u usageExit) Error() string {
	if u.message == "" {
		return fmt.Sprintf("usage exit %d", u.code)
	}
	return u.message
}

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

// parseFlags parses args with flags before, between or after the positional
// arguments, which then remain in fs.Args; everything after "--" is
// positional. A bad flag exits 2 and -h exits 0, as the flag package's
// ExitOnError mode did.
func parseFlags(fs *flag.FlagSet, args []string) error {
	args, literal := splitAtTerminator(fs, args)
	var positional []string
	for {
		if err := parseFlagSet(fs, args); err != nil {
			return err
		}
		if fs.NArg() == 0 {
			break
		}
		positional, args = append(positional, fs.Arg(0)), fs.Args()[1:]
	}
	return keepPositional(fs, append(positional, literal...))
}

// parseOnlyFlags parses the flags of a command that takes no positional
// argument, refusing any.
func parseOnlyFlags(fs *flag.FlagSet, args []string) error {
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	return refuseExtraArguments(fs, 0)
}

// refuseExtraArguments refuses positional arguments beyond the first allowed.
func refuseExtraArguments(fs *flag.FlagSet, allowed int) error {
	if fs.NArg() > allowed {
		return usagef("%s: unexpected argument %q", fs.Name(), fs.Arg(allowed))
	}
	return nil
}

// keepPositional sets fs.Args to arguments without reading any as a flag.
func keepPositional(fs *flag.FlagSet, positional []string) error {
	return parseFlagSet(fs, append([]string{"--"}, positional...))
}

// splitAtTerminator separates the arguments after the "--" that ends flag
// parsing. It reads the arguments as the flag package does, so a "--" given
// as a flag's value is not the terminator.
func splitAtTerminator(fs *flag.FlagSet, args []string) ([]string, []string) {
	for i := 0; i < len(args); i++ {
		if args[i] == "--" {
			return args[:i], args[i+1:]
		}
		if takesNextArgument(fs, args[i]) {
			i++
		}
	}
	return args, nil
}

// takesNextArgument reports a flag written without "=" whose value is the
// next argument: any defined flag that is not boolean.
func takesNextArgument(fs *flag.FlagSet, arg string) bool {
	name := strings.TrimPrefix(strings.TrimPrefix(arg, "-"), "-")
	if !strings.HasPrefix(arg, "-") || name == "" || strings.Contains(name, "=") {
		return false
	}
	defined := fs.Lookup(name)
	if defined == nil {
		return false
	}
	boolean, ok := defined.Value.(interface{ IsBoolFlag() bool })
	return !ok || !boolean.IsBoolFlag()
}

func parseFlagSet(fs *flag.FlagSet, args []string) error {
	err := fs.Parse(args)
	if err == nil {
		return nil
	}
	if errors.Is(err, flag.ErrHelp) {
		return usageExit{code: 0}
	}
	return usageExit{code: 2, message: err.Error()}
}

// listFlag is a repeatable string flag: every occurrence adds one value.
type listFlag []string

func (l *listFlag) String() string { return strings.Join(*l, ",") }

// Set appends one occurrence of the flag.
func (l *listFlag) Set(value string) error {
	*l = append(*l, value)
	return nil
}

// items expands the occurrences into entries. An occurrence is a comma
// separated list of entries; one that carries a JSON object (an "@{"
// argument) is a single entry, because the object may contain commas.
func (l listFlag) items() []string {
	var out []string
	for _, value := range l {
		if strings.Contains(value, "@{") {
			out = append(out, value)
			continue
		}
		out = append(out, splitCSV(value)...)
	}
	return out
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

// parseTriple splits "a=b@c" by two separators into three parts; the third
// is empty when absent.
func parseTriple(s, sep1, sep2 string) (string, string, string, error) {
	first, rest, found := strings.Cut(s, sep1)
	if !found {
		return "", "", "", usagef("bad triple %q", s)
	}
	second, third, _ := strings.Cut(rest, sep2)
	return first, second, third, nil
}
