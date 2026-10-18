package kernel

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel/internal/sockets"
)

var (
	ErrNotPrivate = sockets.ErrNotPrivate
	ErrInUse      = sockets.ErrInUse
	ErrNotRegular = files.ErrNotRegular
	ErrTooLarge   = files.ErrTooLarge
)
