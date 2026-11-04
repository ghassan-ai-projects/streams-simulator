# Evidence: canvas modularity refactoring

- **Date:** 2026-11-04 · **Branch:** `canvas-modularity-audit` · **Plan:** [PLAN.md](PLAN.md)

## Rounds

| Round            | Commits                                                                                 | Bundle against the previous round        |
| ---------------- | --------------------------------------------------------------------------------------- | ---------------------------------------- |
| R1 duplicates    | `690d5dd7` `829f4554` `3c1c97b2` `6c3a7ddd` `b70a62b6` `1bf1d197` `0c6f6ff2` `9f39ba54` | CSS byte-identical; JS differs by design |
| R4 folders       | `9323e87d` `e8fa0416` `a76eb0bd` `8e8696e2` `b8fd8d8f` `e801fddf`                       | JS and CSS byte-identical                |
| R5 selectors     | `5e803aff` `a12c86d3` `dc4132f7` `982ceeeb` plus one unused-export fix                  | CSS byte-identical; JS differs by design |
| Bridge cycle fix | see below                                                                               | JS and CSS byte-identical                |

The final stylesheet is byte-identical to the baseline's, after normalising CSS-module hashes.

No file under `canvas/scripts`, `canvas/architecture-fixtures`, `test/architecture` or root `scripts`, no Go source, and neither `architecture.json`, `ARCHITECTURE.md`, `eslint.config.js`, `vitest.config.ts` nor `.dependency-cruiser.cjs` differs from `origin/main`.

## Gates

| Gate                        | Baseline                            | Final                                                                                                                                                     |
| --------------------------- | ----------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Static canvas gates, lint   | pass                                | pass                                                                                                                                                      |
| Vitest                      | 2,902 passed, 5 skipped             | 2,921 passed, 5 skipped; the 19 new tests cover code that moved to a new owner                                                                            |
| e2e, real binary            | 170 passed, 1 failed (passes alone) | 171 passed                                                                                                                                                |
| `make behaviour-pin-check`  | not run                             | matches                                                                                                                                                   |
| `make architecture` (Enola) | baseline pinned before R1           | pass, after the fix below                                                                                                                                 |
| `make ci-check`             | not run                             | **stops at `vulncheck`:** GO-2026-6809 in `crypto/x509` of go1.27.2. Unrelated, as this branch changes no Go; `fuzz-soak`, the step after it, did not run |

## Found during execution

- **A folder cycle the plan's check missed.** Enola reported `adapters/bridge/internal` → `devices`, `launch`, `provenance` → `internal`. A reviewer's late suggestion moved `decoders` and `follow` to the root, and the cycle check was not re-run on the changed table; it also skipped edges into the root. Fixed by moving them to `documents` and `review`, not by re-pinning the baseline.
- **A false "identical".** An early oracle trial compared stale output after a failed build. The oracle now removes `dist` first and fails on a failed build. Two later negative controls were also vacuous at first, one renaming only types and one changing nothing, and were redone with a real string and a real declaration.
- **A commit with failing tests, amended before anything else built on it.** The `adapters/bridge` move skipped `from ".."` specifiers in 14 tests, and that commit was gated only on the bundle. Production was unaffected.

## Limits of this evidence

Moves are proved by bundle identity. The R1 and R5 JavaScript changes rest on reading the diffs, the unit suite and e2e. `LiveTop` in `ReviewCentre.tsx` had no failing unit test when its `startOf` was removed, and only the type checker caught it, so running-session rendering relies on e2e. No design review in Chrome was done, because the stylesheet and DOM-producing code are unchanged by intent and no visual change is expected.
