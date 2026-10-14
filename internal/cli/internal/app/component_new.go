package app

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/cli/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/component"
)

const newPlantUsage = "usage: streamsim component new <id> --title <title> (--from <composition.json> | --example <id>) [--empty]"

func componentCommand(s *session, args []string) (any, error) {
	if len(args) > 0 && args[0] == "new" {
		return cmdNewPlant(s, args[1:])
	}
	return cmdComponent(s, args)
}

type newPlantOptions struct {
	title, from, example string
	empty                bool
}

func cmdNewPlant(s *session, args []string) (any, error) {
	fs := newFlagSet("component new", s.stderr)
	var options newPlantOptions
	fs.StringVar(&options.title, "title", "", "the new plant's title")
	fs.StringVar(&options.from, "from", "", "the composition to start from")
	fs.StringVar(&options.example, "example", "", "the built-in example to start from")
	fs.BoolVar(&options.empty, "empty", false, "keep only the plant's frame: no equipment or connections")
	if err := parseFlags(fs, args); err != nil {
		return nil, err
	}
	if fs.NArg() != 1 || options.title == "" || (options.from == "") == (options.example == "") {
		return nil, usagef(newPlantUsage)
	}
	return newPlant(fs.Arg(0), options)
}

func newPlant(id string, options newPlantOptions) (any, error) {
	base, source, err := newPlantBase(options)
	if err != nil {
		return nil, wrapCLI(err)
	}
	plant := component.NewPlant{ID: id, Title: options.title, Empty: options.empty}
	composition, err := component.NewComposition(base, plant, source)
	return map[string]any{"composition": composition}, wrapCLI(err)
}

func newPlantBase(options newPlantOptions) ([]byte, string, error) {
	if options.example != "" {
		example, err := component.Example(options.example)
		return example, "example " + options.example, err
	}
	raw, err := files.Read(options.from)
	return raw, options.from, err
}
