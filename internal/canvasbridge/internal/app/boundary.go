package app

import (
	"context"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

type boundary struct{ *Bridge }

func (o boundary) Workspace() (domain.Listing, error) {
	listing, err := o.Bridge.Workspace()
	if err != nil {
		return domain.Listing{}, o.redacted(err)
	}
	return domain.ScrubValue(listing, o.workspaceRoots())
}

func (o boundary) ReadFile(rel string) (domain.FileDocument, error) {
	document, err := o.Bridge.ReadFile(rel)
	return document, o.redacted(err)
}

func (o boundary) WriteFile(rel, content, etag string) (domain.FileDocument, error) {
	document, err := o.Bridge.WriteFile(rel, content, etag)
	return document, o.redacted(err)
}

func (o boundary) OpenRun(worldID string) error {
	return o.redacted(o.Bridge.OpenRun(worldID))
}

func (o boundary) Follow(ctx context.Context, worldID string, after int64, emit func(domain.Record) error) error {
	return o.redacted(o.Bridge.Follow(ctx, worldID, after, emit))
}

func (o boundary) redacted(err error) error {
	if err == nil {
		return nil
	}
	return redact(err, o.workspaceRoots())
}
