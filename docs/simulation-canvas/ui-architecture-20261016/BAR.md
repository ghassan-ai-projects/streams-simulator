# Quality bar — Canvas modular architecture migration

- **Date:** 2026-10-16
- **Applies to:** every phase and every round of [PLAN.md](PLAN.md), in `canvas/` only
- **Precedence:** [AGENTS.md](../../../AGENTS.md), [governance/MODULARITY_STANDARD.md](../../../governance/MODULARITY_STANDARD.md)
  and [governance/CANVAS_QUALITY_BAR.md](../../../governance/CANVAS_QUALITY_BAR.md) apply unchanged. This file adds to
  them and never relaxes them. [CANVAS_UI_ACCEPTANCE.md](../../../governance/CANVAS_UI_ACCEPTANCE.md) §3 remains the
  checklist for the Chrome review of every UI-moving round.

## 1. Job bar

A round is a phase, or one integration unit of Phase 5. The job is done only when all of the following hold.

| # | Bar | Enforced by |
| --- | --- | --- |
| J1 | `canvas/architecture.json` describes the delivered tree, its debt list is empty, and every module in it has a kind, an entrypoint, a lifetime, its callers and an exact (never glob) import allow-list | `npm run architecture` |
| J2 | `check-architecture`, `check-css-architecture`, `check-ratchet` and `check-emitted-css` pass; the ratchet shows caps only fallen, floors only risen, ignores and exclusions only shrunk, debt entries only removed, and the `npm run check` chain with no command lost | `npm run check`, `make canvas-check` |
| J3 | No existing gate is loosened: every threshold, allow-list, lint rule, exclusion, `knip` scope, dependency-cruiser rule and test-hook list is unchanged or stricter; `test/architecture` and root `scripts/` are untouched | ratchet, plus `git diff` review of `canvas/package.json`, `canvas/eslint.config.js`, `canvas/vitest.config.ts`, `canvas/knip.json`, `canvas/.dependency-cruiser.cjs`, `canvas/.jscpd.json` |
| J4 | No behaviour change outside the corrections PLAN.md enumerates; each correction was observed failing first, and the superseded assertion and the new assertion are quoted in that phase's evidence | evidence file, reviewer check |
| J5 | No Go file changed; the behaviour pin matches the baseline at the pinned toolchain | `make phase-gate` with `GOTOOLCHAIN=go1.26.9` |
| J6 | Every phase/unit ended with `make canvas-check` and `make phase-gate` green at the recorded invocation, and `git diff --check` clean | evidence file |
| J7 | Two independent reviewers with different lenses read the diff and the specs; a CSS/UI-moving phase adds the UX/accessibility/visual lens. No P1 or P2 finding is open: each is fixed with evidence, or recorded as not a defect with evidence | `reviews/P<n>.md` |
| J8 | Every UI-moving round has a Chrome review of captured screens at 1440 × 900 and 1920 × 1080, light and dark, recorded with the images | `reviews/P<n>.md` |
| J9 | Source limits hold without exception: files ≤ 300 lines, functions and components ≤ 40 physical lines, cyclomatic ≤ 10, cognitive ≤ 15, no clone ≥ 75 tokens, no unused export or file, zero lint warnings, no `TODO`/`FIXME`, no skipped or focused test, no fixed sleep in a test | `npm run lint`, `npm run dup`, `npm run knip` |
| J10 | Coverage floors hold or rise, including the `src/adapters/**` floor added in Phase 3 | `npm run test`, `vitest.config.ts` |
| J11 | No simulator-owned validation, composition, simulation, truth, scoring or verification appears in TypeScript; a missing capability goes into the simulator first | review, `docs/simulation-canvas/README.md` ownership list |
| J12 | Every new module has a stated responsibility, public API, callers and permitted dependencies before it exists; shared code has at least two real callers and neither is its natural owner | register, review |
| J13 | Every round is committed with its scope, gates and review outcome in the message; structural movement and behaviour correction are separate commits | `git log` |

## 2. Per-phase bar

Each phase must additionally satisfy its PLAN.md exit criteria, evidenced as follows.

| Phase | Extra bar | Evidence |
| --- | --- | --- |
| 0 | Every checker rule has a failing fixture and a passing counter-fixture; the debt list is proven exhaustive; every P0.8 regression fails before its guard; the emitted-stylesheet oracle has a recorded baseline; audit floors are asserted; the `StrictMode` harness renders a route | `evidence/P0.md` with rule → fixture → observed message, the pre-fix failure output, and the baseline declaration multisets |
| 1 | No `graph → review` or `ui/graph → model/plant` edge remains; the public graph API exports only what the three consumers name; the emitted order of vendor, graph and theme rules is unchanged or each change is named; graph debt is gone and the residual screen-layout edge is a P5 entry with its cap | `evidence/P1.md`, `reviews/P1.md`, Chrome images |
| 2 | A closed session cannot publish or install a side effect; app code has no direct clock or timer; A→B, A→B→A, close-while-pending, close/open replay, repeated close and disposed-callback tests pass with no sleeps; `Scheduler`/`Clock` semantics are unit-tested; the P0.8 assertions are still green; the launch hold is measured in isolation before and after | `evidence/P2.md` with the measured hold and the test list |
| 3 | No app public API returns `unknown` or exposes `BridgeFailure`; no app or UI code compares a protocol code or status; feature failure unions are closed and feature-owned; a malformed required response is an explicit failure; the `src/adapters/**` floor holds; per-moved-file coverage is recorded; the P0.8 tests are ported with unchanged assertions | `evidence/P3.md` with the decoder table, superseded assertions and coverage before/after |
| 4 | Save uses the captured revision; edits during a save stay dirty; navigation invalidates the attempt; conflict and partial success are defined states with retry; the written composition bytes are asserted; compose output stays byte-compatible | `evidence/P4.md` with the state machine and the byte assertions |
| 5 | Per unit: one public facade per module, `internal/` unreachable across modules in production and tests, CSS owned by the module that uses it, each unit's debt entries removed with no broader replacement, `canvas/ARCHITECTURE.md` in sync with the register, and no circular module | one evidence and one review section per unit, Chrome images |
| 6 | Every remaining test has an outcome, an owner, a unique failure and, where merged, a replacement in the ledger; no distinct oracle, browser test or real-binary test was deleted; suite time recorded before and after | `evidence/P6.md` with the ledger |

## 3. Round protocol

1. **Implement** the smallest coherent unit; keep behaviour-preserving movement separate from corrections.
2. **Test** with the ladder: focused `vitest` per edit, then `npm run lint && npm run deps && npm run architecture`, then
   the authoritative `make canvas-check`.
3. **Capture** Chrome images and inspect them when the unit moves CSS or UI structure.
4. **Gate**: `make phase-gate` at the phase or unit exit with the recorded invocation.
5. **Review**: two independent reviewers (three for CSS/UI-moving work) read the diff and the specs, not the author's
   summary. Findings enter `reviews/P<n>.md` with severity, evidence and disposition.
6. **Resolve**: fix every P1/P2 with evidence, or record it as not a defect with evidence; re-run the affected gates.
7. **Record** `evidence/P<n>.md` with real command output, including the observed failing-first output for corrections.
8. **Commit** with scope, gates and review outcome in the message.

## 4. Stop conditions

- A phase that cannot meet its bar stops and reports the concrete condition. It never lowers a bar, widens an
  allow-list, raises a cap, removes a check or deletes a test to finish.
- A gate that must change for a phase to pass is a stop condition, not a decision: report it and wait.
- Findings that are not defects are recorded with their evidence; they are not silently dropped.
- Heavy gates never run in parallel with each other. A timing failure in the one wall-clock e2e test is re-run in
  isolation; after Phase 2 it is treated as a regression, not as flakiness.
