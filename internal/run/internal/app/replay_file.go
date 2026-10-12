package app

import (
	"context"
	"fmt"
)

// Directories say where the domains and adapters a run names are looked up.
type Directories struct{ Domains, Adapters string }

// ReplayFile reproduces the run artifact at artifactPath against the domain
// and adapter it recorded, found in the directories, and reports whether the
// trace digest matches.
func ReplayFile(ctx context.Context, artifactPath string, dirs Directories) (*ReplayResult, error) {
	art, err := LoadArtifact(artifactPath)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	spec, adap, err := LoadInputs(dirs.Domains, dirs.Adapters, art.Domain.ID, art.Adapter.ID)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	res, err := ReplayArtifact(ctx, art, spec, adap, "")
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return res, nil
}
