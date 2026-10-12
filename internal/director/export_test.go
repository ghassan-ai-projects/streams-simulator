package director

import "github.com/ghassan-ai-projects/streams-simulator/internal/run"

// RunOf is the composition tests' handle on the run behind a world: the
// closed-loop and covert-channel tests observe and park the run beneath the
// tools, which no surface may do.
func RunOf(d *Director, worldID string) *run.Run {
	return d.impl.World(worldID).Run
}
