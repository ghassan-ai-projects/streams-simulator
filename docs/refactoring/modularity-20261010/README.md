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

## Outcome (2026-10-10)

- Every `internal/` package is a facade over private layers
  (`internal/<m>/internal/{domain,app,<edges>}`) except the six foundations
  (`canonical`, `jsonschema`, `model`, `randutil`, `schemas`, `device/contract`),
  the wall seam (P-01) and the test aid `testsupport`;
  `TestEveryModuleIsMigrated` keeps it that way.
- Behaviour is unchanged: `scripts/behaviour-pin` reproduces the baseline;
  the deliberate changes (fail-closed constructors, M15 items) are in
  PLAN with regression tests.
- Tests: every package at or above 70 % except the entrypoint; all tests
  parallel and sleep-free; shipped fixtures named once; four cross-module
  acceptance flows against the real binary.
- Open work is in [DEFERRED.md](DEFERRED.md): defects D-01…D-51 (found, not
  fixed), policy items P-01…P-07, and the T-01 ratchet.
