package app

import "flag"

var listOnlyFlags = map[string]bool{"unit": true, "min": true, "max": true}

type readingFlags struct {
	low, high float64
	given     map[string]bool
}

func (r *readingFlags) register(fs *flag.FlagSet) {
	fs.Float64Var(&r.low, "min", 0, "list: the lowest value of the reading, in --unit; every instrument then states whether it fits")
	fs.Float64Var(&r.high, "max", 0, "list: the highest value of the reading, in --unit")
}

func (r *readingFlags) record(fs *flag.FlagSet) {
	r.given = map[string]bool{}
	fs.Visit(func(set *flag.Flag) { r.given[set.Name] = true })
}

func (r readingFlags) bounds() (low, high *float64) {
	if r.given["min"] {
		low = &r.low
	}
	if r.given["max"] {
		high = &r.high
	}
	return low, high
}

func misplacedListFlags(fs *flag.FlagSet, verb string) error {
	var misplaced error
	fs.Visit(func(set *flag.Flag) {
		if listOnlyFlags[set.Name] && verb != "list" && misplaced == nil {
			misplaced = usagef("component %s: --%s applies to list only", verb, set.Name)
		}
	})
	return misplaced
}
