package deviceworld

import (
	"errors"

	"github.com/ghassan-ai-projects/streams-simulator/internal/deviceworld/internal/app"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/deviceworld/internal/domain"
)

// Binding is a validated data-defined mapping from one device target to a
// world effector invocation. LoadBindings is the supported construction path;
// domain mappings are not supplied as Go callbacks.
type Binding = layer.Binding

// ErrNoWorld is returned by New when no world is given: the plant is the
// device's view of that world.
var ErrNoWorld = errors.New("deviceworld: a world is required")

// ErrNoPlant is returned by a nil or zero Plant that was not built by New.
var ErrNoPlant = errors.New("deviceworld: a plant built by New is required")

// OpenOptions say which device to open: its identity, its capability catalog
// and fault schedule, and, when WorldBindings is set, the world it is bound to.
// Now is the clock the plant counts device microseconds from.
type OpenOptions = app.OpenOptions
