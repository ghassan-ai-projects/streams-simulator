package deviceworld

import (
	"errors"

	layer "github.com/ghassan-ai-projects/streams-simulator/internal/deviceworld/internal/domain"
)

// Binding is a validated data-defined mapping from one device target to a
// world effector invocation. LoadBindings is the supported construction path;
// domain mappings are not supplied as Go callbacks.
type Binding = layer.Binding

// ErrNoWorld is returned by New when no world is given: the plant is the
// device's view of that world.
var ErrNoWorld = errors.New("deviceworld: a world is required")
