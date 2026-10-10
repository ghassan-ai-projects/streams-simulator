package deviceworld

import (
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/deviceworld/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// Plant is a device.Plant backed by a *world.World: the world's effector is
// the oracle behind the device's wire loop.
type Plant struct {
	plant *layer.Plant
}

// New returns a world-backed plant. bindings maps device targets (for example
// "fan-01") to world effector invocations. A nil world is refused with
// ErrNoWorld.
func New(w *world.World, bindings map[string]Binding) (*Plant, error) {
	if w == nil {
		return nil, ErrNoWorld
	}
	return &Plant{plant: layer.New(w, bindings)}, nil
}
