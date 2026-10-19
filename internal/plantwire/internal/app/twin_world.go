package app

import (
	"errors"
	"fmt"

	simdomain "github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

var ErrTwinRefused = errors.New("plantwire: the twin refused the command")

type twinWorld struct {
	world    *world.World
	spec     *simdomain.Compiled
	entities map[string]string
}

func newTwinWorld(spec *simdomain.Compiled, startNS int64, devices []domain.BoardDevice) (*twinWorld, error) {
	opts := world.Options{Noiseless: true, EmitDisabled: true, Unobserved: true, ForceEffectorOK: true}
	w, err := world.New(spec, 1, "twin", startNS, opts)
	if err != nil {
		return nil, fmt.Errorf("plantwire twin-report: %w", err)
	}
	twin := &twinWorld{world: w, spec: spec, entities: map[string]string{}}
	for _, device := range devices {
		twin.bind(device)
	}
	return twin, nil
}

func (t *twinWorld) bind(device domain.BoardDevice) {
	for _, id := range t.world.EntityIDs() {
		if domain.WireID(id) == device.Name && t.kindOf(id).Implements == device.Kind {
			t.entities[device.Name] = id
			return
		}
	}
}

func (t *twinWorld) kindOf(entity string) *model.Kind {
	return t.spec.InstanceKind(t.spec.InstanceSetFor(entity))
}

func (t *twinWorld) Support(device, metric string) string {
	entity, bound := t.entities[device]
	switch {
	case !bound:
		return domain.SkipNoEntity
	case t.kindOf(entity).ChannelDecl(metric) == nil:
		return domain.SkipNoChannel
	}
	return ""
}

func (t *twinWorld) Apply(command domain.TwinCommand) error {
	entity, bound := t.entities[command.Device]
	if !bound {
		return fmt.Errorf("%w: no twin entity for device %s", ErrTwinRefused, command.Device)
	}
	result, err := t.world.InvokeEffector(command.Name, entity, command.ID, command.Args, command.AtNS)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrTwinRefused, err)
	}
	if !result.Accepted {
		return fmt.Errorf("%w: %s", ErrTwinRefused, result.Reason)
	}
	return nil
}

func (t *twinWorld) Predict(device, metric string, atNS int64) float64 {
	return t.world.Reading(t.entities[device], metric, atNS)
}
