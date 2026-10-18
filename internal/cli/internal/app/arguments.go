package app

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

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

func newFlagSet(name string, stderr io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	return fs
}

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

func parseOnlyFlags(fs *flag.FlagSet, args []string) error {
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	return refuseExtraArguments(fs, 0)
}

func refuseExtraArguments(fs *flag.FlagSet, allowed int) error {
	if fs.NArg() > allowed {
		return usagef("%s: unexpected argument %q", fs.Name(), fs.Arg(allowed))
	}
	return nil
}

func keepPositional(fs *flag.FlagSet, positional []string) error {
	return parseFlagSet(fs, append([]string{"--"}, positional...))
}

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

type listFlag []string

func (l *listFlag) String() string { return strings.Join(*l, ",") }

func (l *listFlag) Set(value string) error {
	*l = append(*l, value)
	return nil
}

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

func parseTriple(s, sep1, sep2 string) (string, string, string, error) {
	first, rest, found := strings.Cut(s, sep1)
	if !found {
		return "", "", "", usagef("bad triple %q", s)
	}
	second, third, _ := strings.Cut(rest, sep2)
	return first, second, third, nil
}

func registerInputDirs(fs *flag.FlagSet, domains, adapters *string) {
	registerDomainsDir(fs, domains, "domains")
	registerAdaptersDir(fs, adapters, "adapters")
}

func registerDomainsDir(fs *flag.FlagSet, dst *string, def string) {
	fs.StringVar(dst, "domains-dir", def, "directory of domain specs")
}

func registerAdaptersDir(fs *flag.FlagSet, dst *string, def string) {
	fs.StringVar(dst, "adapters-dir", def, "directory of adapter files")
}

func registerComponentDirs(fs *flag.FlagSet, dst *listFlag) {
	fs.Var(dst, "components-dir", "customer component library directory, repeatable; the built-in library is always loaded first")
}

const kindsDirFlag = "kinds-dir"

func registerKindsDirs(fs *flag.FlagSet, dst *listFlag) {
	fs.Var(dst, kindsDirFlag, "directory of kind files (*.kind.json), repeatable; a domain kind that implements a kind file needs it")
}

func verbOfFlag(name string) string {
	if name == kindsDirFlag {
		return "validate"
	}
	return "lint"
}

func registerRunFolder(fs *flag.FlagSet, dst *string, def string) {
	fs.StringVar(dst, "out", def, "run folder for the artifact and the journal")
}
