package app

type componentOptions struct {
	dirs listFlag
	unit string
	verb string
	args []string
}

func parseComponentOptions(s *session, args []string) (componentOptions, error) {
	var options componentOptions
	fs := newFlagSet("component", s.stderr)
	registerComponentDirs(fs, &options.dirs)
	fs.StringVar(&options.unit, "unit", "", "list: only the instruments that measure in this unit or in a convertible one, the exact unit first")
	verb, err := parseCommandVerb(fs, args, "list")
	if err != nil {
		return componentOptions{}, err
	}
	options.verb, options.args = verb, fs.Args()
	if options.unit != "" && verb != "list" {
		return componentOptions{}, usagef("component %s: --unit applies to list only", verb)
	}
	return options, nil
}
