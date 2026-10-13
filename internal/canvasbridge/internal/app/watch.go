package app

import (
	"context"
	"time"
)

const watchInterval = time.Second

var watchedFolders = []string{"domains", "compositions", "scenarios", "layouts", "runs"}

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
