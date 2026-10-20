package app

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	wire "github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

func (d *devices) effector(entity, name string) *model.Effector {
	kind, mapped := d.kindOf(entity)
	if mapped == nil {
		return d.run.World.EffectorOf(entity, name)
	}
	return kind.EffectorDecl(name)
}

func (d *devices) checkArgs(entity, name string, args map[string]any) error {
	kind, mapped := d.kindOf(entity)
	if mapped == nil {
		return worldArgs(d.run.World.CheckEffectorArgs(name, args))
	}
	if err := wire.CheckArguments(kind.EffectorDecl(name).ArgsSchema, args); err != nil {
		return err
	}
	_, _, err := wire.RawCommand(*mapped, name, args)
	return err
}

func (d *devices) rawCommand(entity, name string, args map[string]any) (string, map[string]any, error) {
	_, mapped := d.kindOf(entity)
	if mapped == nil {
		return name, args, nil
	}
	return wire.RawCommand(*mapped, name, args)
}

func worldArgs(err error) error {
	if err != nil {
		return fmt.Errorf("plantwire: %w", err)
	}
	return nil
}
