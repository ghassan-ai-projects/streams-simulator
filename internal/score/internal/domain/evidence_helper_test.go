package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/run"

// evidenceOf packs a finished run into the scorer's input, as a host does.
func evidenceOf(r *run.Run) Evidence {
	return Evidence{
		RunID: r.ID, Domain: r.Domain(), Verdict: r.Verdict(), Ledger: r.Ledger(),
		Calls: r.World.EffectorCalls(), Perturbations: r.AppliedPerturbations(),
		Emitted: r.World.EmittedCount(), History: r.History(),
		Reproducible: r.Reproducible(), Unblinded: r.UnblindedStamp(),
	}
}
