package app

import (
	"context"
	"time"
)

const watchInterval = time.Second

func (b *Bridge) Watch(ctx context.Context, changed func() error) error {
	last := b.edges.Workspace.Fingerprint(workspaceFolders)
	return b.everyTick(ctx, watchInterval, func() (bool, error) {
		current := b.edges.Workspace.Fingerprint(workspaceFolders)
		if current == last {
			return false, nil
		}
		last = current
		return false, changed()
	})
}
