package app

import (
	"flag"
	"io"
)

const twinUsage = "usage: streamsim kind twin [<file.kind.json>...] (--choices <choices.json> [--out <file>] | --suggest) [--kinds-dir <dir>] [--components-dir <dir>]"

type twinOptions struct {
	choices   string
	out       string
	suggest   bool
	kindFiles []string
	kindDirs  listFlag
	libraries listFlag
}

func parseTwinOptions(args []string, stderr io.Writer) (twinOptions, error) {
	var options twinOptions
	fs := newFlagSet("kind twin", stderr)
	registerTwinFlags(fs, &options)
	if err := parseFlags(fs, args); err != nil {
		return twinOptions{}, err
	}
	options.kindFiles = fs.Args()
	if err := options.check(); err != nil {
		return twinOptions{}, err
	}
	return options, nil
}

func (o twinOptions) check() error {
	if o.suggest && (o.choices != "" || o.out != "") {
		return usagef("kind twin: --suggest takes no --choices and no --out: it prints the choices and writes nothing")
	}
	if !o.suggest && o.choices == "" {
		return usagef(twinUsage)
	}
	return nil
}

func registerTwinFlags(fs *flag.FlagSet, options *twinOptions) {
	fs.StringVar(&options.choices, "choices", "", "the choices document: the instrument of each metric and the response of each command")
	fs.StringVar(&options.out, "out", "", "file to write the twin domain to (default: print it inside the output document)")
	fs.BoolVar(&options.suggest, "suggest", false, "print a valid choices document with the binary's defaults for the kind files, and write nothing")
	registerKindsDirs(fs, &options.kindDirs)
	registerComponentDirs(fs, &options.libraries)
}
