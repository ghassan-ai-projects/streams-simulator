package deviceworld

import (
	"errors"

	"github.com/ghassan-ai-projects/streams-simulator/internal/deviceworld/internal/app"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/deviceworld/internal/domain"
)

type Binding = layer.Binding

var ErrNoWorld = errors.New("deviceworld: a world is required")

var ErrNoPlant = errors.New("deviceworld: a plant built by New is required")

type OpenOptions = app.OpenOptions
