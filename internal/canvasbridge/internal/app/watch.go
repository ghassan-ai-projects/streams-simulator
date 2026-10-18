package app

import (
	"context"
	"slices"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

const watchInterval = time.Second

var watchedFolders = slices.Concat(workspaceFolders, []string{"kinds", "twins"}, domain.ComponentFolders())

func (b *Bridge) Watch(ctx context.Context, changed func() error) error {
	last := b.edges.Workspace.Fingerprint(watchedFolders)
	return b.everyTick(ctx, watchInterval, func() (bool, error) {
		current := b.edges.Workspace.Fingerprint(watchedFolders)
		if current == last {
			return false, nil
		}
		last = current
		return false, changed()
	})
}
