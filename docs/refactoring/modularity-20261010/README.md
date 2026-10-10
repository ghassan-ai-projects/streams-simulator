# Modularity program — 2026-10-10

Goal: bring `streams-simulator` to the modularity bar of the sibling
repository `agentic-stream` (facade → use cases → pure domain → named edge
adapters, enforced by architecture gates), without changing behaviour.

| File | Purpose |
| --- | --- |
| [STANDARD.md](STANDARD.md) | The patterns, quality bar (M1–M11), package kinds and constraints for this repository |
| [FINDINGS.md](FINDINGS.md) | Structural verdict per package, mapped to rounds |
| [PLAN.md](PLAN.md) | Rounds, hazards, proof, **status table with commit hashes** |
| [ROUNDS.md](ROUNDS.md) | Per-round evidence log: what moved, review findings, results |
| [DEFERRED.md](DEFERRED.md) | Bugs, risks and improvements found but **not** fixed here |
| [BEHAVIOUR_PIN.txt](BEHAVIOUR_PIN.txt) | Baseline output hashes; `scripts/behaviour-pin` must reproduce it after every round |
| [survey/](survey/) | Raw package surveys (read-only, 2026-10-10) |

Reference material in the sibling repo: `AGENTS.md`,
`.agents/context/quality-bar.md`, `.agents/context/architecture-bar.md`,
`.agents/prompts/reference-module-refactor.md`, `internal/architecture/README.md`.
(The task called it `agentic-streams`; the directory is `agentic-stream`.)

Builds on [clean-code-20261002](../clean-code-20261002/BAR.md) (file/function
limits and the direct-import allowlist, already complete).
