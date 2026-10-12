package plantwire

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/app"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/uds"
)

type Options = app.Options

var (
	ErrRefused          = domain.ErrRefused
	ErrHostLost         = app.ErrHostLost
	ErrHostGap          = app.ErrHostGap
	ErrSocketNotPrivate = uds.ErrSocketNotPrivate
)
