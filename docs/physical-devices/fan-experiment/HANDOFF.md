# Handoff: the fan experiment (rewritten 2026-10-21, after P4)

Read this first, then [GOAL.md](GOAL.md) (quality bar, round log, owner-waiting list W1 to W15) and
[PLAN.md](PLAN.md) (decisions D1 to D11, all taken by the owner).

## 1. Where the goal stands

The **simulated rung (P4) passes all seven criteria of PLAN section 7**. What is left is the real rung:
**P5 (bench safety checks) and P6 (the real run) need the owner at the board**. The simulated rung was run
twice; both run folders are retained under `~/.e2e-run/`.

| Repo | Branch | Tip | State |
|---|---|---|---|
| `/Users/ghassan/my-projects/streams-simulator` | `fan-sim` (cut from `real-run`) | `61b8223d` (the P4 result) | clean; `make phase-gate` green on `2004dafb` (S11b) |
| `/Users/ghassan/my-projects/agentic-stream` | `fan-a` (adapter) | `3f5694b6` (A5) | clean; `make ci-check` green. The branch is checked out only in a linked worktree in the scratchpad: `/private/tmp/claude-501/-Users-ghassan-my-projects-streams-simulator/517cecf1-e1a3-4525-abf0-112e2348451e/scratchpad/wt-fan-a`. If the worktree vanishes: `git -C agentic-stream worktree prune` and `git worktree add <path> fan-a`. The main checkout is on `fan-h0` with the owner's untracked `documentation/visual-guide/` (leave it) |
| `/Users/ghassan/my-projects/agent-research-lab` | `fan-gateway` (cut from `main`) | `895d3387` (G1j) | clean, gateway complete |
| `/Users/ghassan/my-projects/tamoz` | `real-run` | `9170538c` | no code; untracked `PROFILE.md`, `outreach/` are the owner's: leave them |

Nothing is pushed or merged. Merge to `main` only when the owner says so.

### What was done since the previous handoff (2026-10-20)

- **S11b gated** on streamsim: `make build` then `make phase-gate` green on `2004dafb` (the environment needs
  `$(go env GOPATH)/bin` on `PATH` for `deadcode` and `govulncheck`, and the newer `golangci-lint` 2.14.0 wins over
  the older one in that folder).
- **The contract re-synced** in the two repositories that vendor it: gateway `895d3387` (G1j, 322 tests green) and
  agentic-stream `20312a1e` (A4.1).
- **Review B written** (`reviews/A-review-B.md`: one High, six Medium, seven Low) and **round A5** (`3f5694b6`)
  fixed both reviews: the claim is renewed while a command waits at the node, the consent is checked again when the
  record is written, every birth is admitted again, a command the host cannot encode is "not sent", the site's
  devices are the ones the node must announce, the actuation consent is one fact of the `device` module.
- **D3 answered**: one `serve` serves both rungs — `--effect-profile physical --live-actuation --owner-authorized
  --plant-wire-environment mixed --hardware-boards mega-01`. The switch to the real board is the session plan alone.
- **P4 run** (twice, see GOAL.md): the fan command reached the twin, was executed, and was verified by the fan's own
  reading (`level_percent` 45 for its 5 s lease, 0 after it).

## 2. Open work, in order

1. **P5 (the owner at the bench).** The gateway runbook `agent-research-lab/real-world-sensor/assessment/RUNBOOK-PLANTWIRE-GATEWAY.md`,
   the fan-experiment README (`examples/devices/thermal-board/fan-experiment/README.md`), the bench checks in
   `docs/physical-devices/briefs/bench-checks-m2-m5.md` and PLAN.md section 5 "Outside the repositories" (O1 to O5).
   The real run uses `real.plan.json`; the Mega may reset when the serial port opens (UNVERIFIED, on the checklist).
2. **P6 (the owner at the bench, after P5).** The same chain with `real.plan.json`; Agentic Stream and Tamoz keep
   running (the consent for `mega-01` was given at start).
3. **W15 (the owner's call).** A refused command leaves its verification `awaiting`, so the run artifact's
   zero-tolerance check counts it and `verify-run` says `fail` on a clean run. It is in Agentic Stream's
   `internal/actions` (`completeClassification`).
4. **Later goals (PLAN.md 8.1).** Three rooms in one session, the A/B/C evaluation, rules-first decisions, Tamoz
   proposing rule changes.

## 3. Owner-waiting list

W1 to W15 are in [GOAL.md](GOAL.md). W13 (the board digest on the wire) and W14 (a contract text change needs a
version bump or a re-sync everywhere) were added on 2026-10-21; W15 is new from the P4 run.

## 4. Rules and pitfalls

- Binding: each repo's AGENTS.md. In streamsim: files ≤ 300 lines, functions ≤ 15 lines, minimal comments, no
  consumer knowledge in the binary, domains/devices are data, **never loosen a gate** (caps, allow-lists, coverage
  floors, test hooks, anything under `test/architecture` or `scripts/`): floors may only rise; a gate change needs the
  owner's approval line `Owner-approved-gate-change: <reason>`. New-module registrations are not loosening. No
  `//nolint`. US spelling in strings (`misspell` runs in lint).
- Owner rules (memory): commit each improvement, review before commit; the writing model writes code and its tests, a
  second model investigates and reviews, a third runs commands and gates; never block on long scripts (run them in
  the background); branch per milestone, merge to main only at the end; customer data never in GitHub; browser
  reviews in the owner's real Chrome (canvas UI rounds).
- Never print or copy secrets (`~/.tamoz/service.env` values, keys). The permission classifier blocks reading that
  file; source it only inside the worker's shell as the runbooks say.
- zsh has no `PIPESTATUS`: when a gate agent pipes to `tail`, the exit code is lost; write output to a log file and
  check `$?` separately.
- Gate timing is noisy when several things run at once (Agentic Stream's T7 suite-time budget, streamsim canvas e2e).
  Rerun alone before declaring a failure real.
- The streamsim binary is `bin/streams-simulator` (the old `bin/streamsim` is stale). `make build` refreshes it; tests
  that use it skip when it is older than HEAD.
- Writers drift in size: a writer whose context is huge should be replaced by a fresh one briefed from the findings
  files.

## 5. Pointers

- Findings per area (facts with file:line): `H0-FINDINGS.md`, `P0.5-FINDINGS.md`, `S-FINDINGS.md`, `G-FINDINGS.md`,
  `A-FINDINGS.md`. Reviews: `reviews/*.md` (H0, G, S, A-review-A, A-review-B).
- The runbook of the simulated rung: `P4-RUNBOOK.md` here and the kit `/Users/ghassan/my-projects/.e2e-run/fan-p4-kit/`
  (the kit copy is the one to edit; `scripts/prep-run.sh` builds both binaries from the committed branches).
- Design docs: `docs/design/PHYSICAL_DEVICES.md` (streamsim), `documentation/design/plant-wire-adapter.md` in the
  `fan-a` worktree (Agentic Stream ADR), contract `contracts/plant-wire/SPEC.md` (1.4.0, section 15 board mapping).
