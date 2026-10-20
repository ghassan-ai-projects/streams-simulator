package app

import (
	"errors"
	"fmt"
	"math"

	simdomain "github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

var ErrTwinRefused = errors.New("plantwire: the twin refused the command")

type twinWorld struct {
	world    *world.World
	spec     *simdomain.Compiled
	faces    domain.Interfaces
	entities map[string]string
}

func newTwinWorld(model twinModel, startNS int64, devices []domain.BoardDevice) (*twinWorld, error) {
	opts := world.Options{Noiseless: true, EmitDisabled: true, Unobserved: true, ForceEffectorOK: true}
	w, err := world.New(model.spec, 1, "twin", startNS, opts)
	if err != nil {
		return nil, fmt.Errorf("plantwire twin-report: %w", err)
	}
	twin := &twinWorld{world: w, spec: model.spec, faces: model.faces, entities: map[string]string{}}
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
	if !bound {
		return domain.SkipNoEntity
	}
	if kind, _ := t.wireKind(entity); kind.ChannelDecl(metric) == nil {
		return domain.SkipNoChannel
	}
	return ""
}

func (t *twinWorld) Apply(command domain.TwinCommand) error {
	entity, bound := t.entities[command.Device]
	if !bound {
		return fmt.Errorf("%w: no twin entity for device %s", ErrTwinRefused, command.Device)
	}
	name, args, err := t.rawCommand(entity, command)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrTwinRefused, err)
	}
	return accepted(t.world.InvokeEffector(name, entity, command.ID, args, command.AtNS))
}

func accepted(result *world.InvokeResult, err error) error {
	if err != nil {
		return fmt.Errorf("%w: %w", ErrTwinRefused, err)
	}
	if !result.Accepted {
		return fmt.Errorf("%w: %s", ErrTwinRefused, result.Reason)
	}
	return nil
}

func (t *twinWorld) Predict(device, metric string, atNS int64) float64 {
	entity := t.entities[device]
	_, mapped := t.wireKind(entity)
	if mapped == nil {
		return t.world.Reading(entity, metric, atNS)
	}
	raw := mapped.Metrics[metric].Raw
	_, value, carried := domain.SemanticReading(*mapped, raw, t.world.Reading(entity, raw, atNS))
	number, isNumber := value.(float64)
	if !carried || !isNumber {
		return math.NaN()
	}
	return number
}

func (t *twinWorld) wireKind(entity string) (*model.Kind, *domain.MappedKind) {
	effective := t.kindOf(entity)
	if face, mapped := t.faces.Of(t.spec.Kind(effective.Name)); mapped != nil {
		return face, mapped
	}
	return effective, nil
}

func (t *twinWorld) rawCommand(entity string, command domain.TwinCommand) (string, map[string]any, error) {
	_, mapped := t.wireKind(entity)
	if mapped == nil {
		return command.Name, command.Args, nil
	}
	return domain.RawCommand(*mapped, command.Name, command.Args)
}
