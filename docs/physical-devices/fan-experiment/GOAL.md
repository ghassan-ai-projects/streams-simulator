# Goal: the fan experiment works, simulated first, then real

- **Set by the owner, 2026-10-20:** implement [PLAN.md](PLAN.md) until the real fan runs through Tamoz, Agentic Stream and
  streamsim, switching from simulated to real by the session plan only. Commit each round, reviewed before the commit.
- **Model roles (owner, 2026-10-20):** **Opus writes** (production code and its tests, documents); **Sonnet investigates**
  (reads code, traces behaviour, writes findings and briefs' facts); **Haiku runs** commands, gates, lint and lint fixes, and
  any small mechanical task. The main session briefs, reviews every diff and commits.
- **Branches:** streamsim rounds on a branch per phase cut from the previous tip, starting from `real-run`. Agentic Stream:
  `fan-h0` (H0), then `fan-plantwire-host` (A0 to A4). Gateway (`agent-research-lab`): `fan-gateway`. Tamoz: no branch (no
  code). Untracked files the owner keeps in those checkouts are left alone. Merge to `main` only when the owner says so.

## Outcome

[PLAN.md](PLAN.md) section 7's pass criteria hold on the simulated rung (P4) and on the real rung (P6), with Agentic Stream and
Tamoz not restarted between them.

## Rounds

The phases and work items of [PLAN.md](PLAN.md) section 6, in that order: P0, P0.5, P0.7 (H0.1 to H0.5), then S1 to S9,
A0 to A4 and G1 to G2 (in parallel where they touch different repositories), P4, P5 (the owner), P6. A round that turns out
too large is split; a deviation from the plan that improves it is made and recorded in the round log.

## Quality bar (every round)

1. **Brief.** The main session gives one Opus writer a self-contained brief: the deliverable, the files or modules, the rules
   below, what done looks like. When the facts are not yet known, a Sonnet investigator reads the code first and returns
   them; the brief is written from its findings. A new package or contract states kind, layer, responsibility and callers
   first (AGENTS.md MUST 4; the other repository's AGENTS.md when the round is there).
2. **Code rules.** The repository's AGENTS.md in full. For streamsim: logic in its module, files at most 300 lines, functions
   at most 15 lines, no consumer knowledge in the binary, domains and devices are data, no new dependency, minimal comments.
3. **Never loosen a gate**, and no `//nolint`: fix the cause. Lint findings are fixed by a Haiku agent that edits and re-runs
   the linter itself.
4. **Tests at the layer they prove**, written by the same Opus writer as the code. Contract changes have fixtures and schema
   tests; a change to what a run produces has a replay-determinism test.
5. **Every command by Haiku**, started by the main session with a fixed list (`make build` first, focused `go test` for the
   touched packages, the thermal-experiment tests, `make quick`), reporting failures only. Failures go back to the same Opus
   writer by `SendMessage`; then Haiku runs the gate again. Long runs (`make phase-gate`, joined runs) in the background.
6. **Review before commit** by the main session: `git diff --stat`, then every change to `test/architecture`, `scripts/`,
   `contracts/`, module facades, `AGENTS.md`, the Makefile, and the round's own diff against its brief.
7. **Thermal experiment** (`go test ./internal/device/... ./internal/deviceworld/... ./test/contract/...`) stays green.
8. **Commit.** One commit per round, naming the round (for example `fan P0.7 H0.1: …`) and listing any gate change with its
   reason.
9. **Phase end** (after P0.7, the S rounds, A4, G2, P4): `make phase-gate` (or the other repository's full gate) by Haiku in
   the background, and two Sonnet reviewers with different lenses: (A) correctness and architecture against this bar and
   PLAN.md; (B) adversarial: hostile input, bypassable checks, tests that would not fail if the code broke, the truth
   boundary. High findings fixed, Medium fixed or deferred with a reason in the round log.
10. **Failures in a joined run** are triaged by PLAN.md section 4.3 before anything is fixed.

## Done means

- PLAN.md section 7 holds on P4 and P6, with retained run folders and Agentic Stream `verify-run` passes.
- Every round committed with its review; the full gate green on the last commit of each repository's branch.
- PHYSICAL_DEVICES.md, `documentation/` and the other repositories' docs describe what was built; deviations in the round log.

## Round log

| Round | Repository | Commit | Notes |
|---|---|---|---|
| P0 | streams-simulator | 9b95c33a | Plan and goal committed; the old demo canvas (PID 27234) stopped, its workspace left in place; commits at start: streams-simulator `real-run` 9b95c33a, agentic-stream `real-run` 91c2079d, tamoz `real-run` 9170538c, agent-research-lab `main` faf457b; `make build` and `make phase-gate` green on 9b95c33a (354 s; the Enola step reported coupling changes as information only) |
