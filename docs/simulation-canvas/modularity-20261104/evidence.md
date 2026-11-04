# Evidence: canvas modularity refactoring

- **Date:** 2026-11-04 · **Branch:** `canvas-modularity-audit` · **Plan:** [PLAN.md](PLAN.md)

## Rounds

| Round                 | Commits                                                                                 | Bundle against the previous round               |
| --------------------- | --------------------------------------------------------------------------------------- | ----------------------------------------------- |
| R1 duplicates         | `690d5dd7` `829f4554` `3c1c97b2` `6c3a7ddd` `b70a62b6` `1bf1d197` `0c6f6ff2` `9f39ba54` | CSS byte-identical; JS differs by design        |
| R4 folders            | `9323e87d` `e8fa0416` `a76eb0bd` `8e8696e2` `b8fd8d8f` `e801fddf`                       | JS and CSS byte-identical                       |
| R5 selectors          | `5e803aff` `a12c86d3` `dc4132f7` `982ceeeb` plus one unused-export fix                  | CSS byte-identical; JS differs by design        |
| Bridge cycle fix      | see below                                                                               | JS and CSS byte-identical                       |
| R2 `ui/capture`       | `72ec5f48`                                                                              | JS and CSS byte-identical                       |
| R3 `model/compare`    | `1af1615f`                                                                              | CSS byte-identical; unminified statements match |
| R3 `model/twinReport` | `7886365b`                                                                              | CSS byte-identical; unminified statements match |
| R3 `ui/compare`       | `75f36a60`                                                                              | CSS byte-identical; unminified statements match |

The final stylesheet is byte-identical to the baseline's, after normalising CSS-module hashes.

R3 reorders modules, which reshuffles minified identifiers, so the minified JS differs without any change in meaning. For those commits both sides were first built unminified and compared as multisets of top-level statements, ignoring module order, path comments and the bundler's `$N` rename suffixes: all 5,811 match, and a one-string negative control fails. That check cannot see a statement rebound to a different function of the same base name, and it says nothing about evaluation order.

### R3 proof with resolved bindings

A second comparison closes both gaps. Each commit was extracted from git and built from clean source, unminified and minified, and each pair was compared on its syntax trees: locals are numbered by first use, top-level bindings are identified by refining over the reference graph, so a reference counts by what it binds and not by what it is called. Evaluation order of the top-level items is then compared.

| Pair (each both unminified and minified) | Items only on one side | Moved in evaluation order | Moved with an eager effect | New early reads | Order flips |
| ---------------------------------------- | ---------------------- | ------------------------- | -------------------------- | --------------- | ----------- |
| `72ec5f48` → `1af1615f`                  | 0                      | 16                        | 0                          | 0               | 0           |
| `1af1615f` → `7886365b`                  | 0                      | 22                        | 0                          | 0               | 0           |
| `7886365b` → `75f36a60`                  | 0                      | 0                         | 0                          | 0               | 0           |
| `72ec5f48` → `75f36a60`                  | 0                      | 38                        | 0                          | 0               | 0           |

CSS is byte-identical and `index.html` is identical apart from the script's file hash in all of them. The minified rows cover the bundle that ships: it is the same program up to names. The 38 moved items are data and function values, listed by the tool: no call runs when they are set, nothing reads them before they are set, and no top-level call that could reach them through a function changes order against them.

Planted faults, each one change to the `75f36a60` unminified bundle:

| Planted change                                     | Expected | This comparison               | Statement comparison |
| -------------------------------------------------- | -------- | ----------------------------- | -------------------- |
| Rebuild of the same commit (`72ec5f48`)            | match    | match                         | not run              |
| `list$5` renamed everywhere                        | match    | match                         | match                |
| One call to `scopeOf$2` now calls `scopeOf`        | differ   | differs, names `optionalScope` | **match**            |
| One character in a string                          | differ   | differs                       | differs              |
| Two statements swapped, the second reads the first | differ   | differs, early read           | **match**            |
| A value moved after a call reaching it indirectly  | differ   | differs, order flip           | differs              |

The first planted rebind passed, and was right to: the pair it chose, `unescaped$1` and `unescaped`, are the same function, which also exposed a duplicate R1 missed (below). The planted fault was then changed to pick a pair with different bodies.

Limits: functions with the same structure count as the same, so a rebind between two of them is invisible, which matters only if they hold different module state. Eager effects are found syntactically: a getter is not seen, and global reads and calls are counted only outside function bodies. Equality shows the bundle is the same program, not that the program is right. The comparison scripts are not committed, so this table cannot be regenerated from the repository.

No file under `canvas/scripts`, `canvas/architecture-fixtures`, `test/architecture` or root `scripts`, no Go source, and none of `eslint.config.js`, `vitest.config.ts` or `.dependency-cruiser.cjs` differs from `origin/main`.

Register changes, all approved by the owner on 2026-11-04 and named in `Owner-approved-gate-change:` commit lines: new modules `ui/capture`, `ui/compare`, `model/compare`, `model/twinReport`; `ui/shell` gains `ui/capture`; `app/ports`, `app/review` and `adapters` gain `model/compare` and `model/twinReport`; `ui/review` gains `ui/compare` and `model/twinReport`. `ui/review` held `model/compare` only between two commits. The `model/review` facade fell from 234 to 167 lines, its contract from 83 to 55.

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
- **Two duplicates R1 missed.** `model/devices/internal/issues.ts` kept its own `unescaped`, and `captureDraft.ts` wrote the same logic inline, although `829f4554` is titled "one unescaped". The R3 proof found the first and a search found the second. Both now use `model/shared/json.ts`. Against `75f36a60` the bundle loses that one function and `pointerValue` calls the shared one; the other items that differ only call into those, and the CSS is byte-identical. `make canvas-check` passes on the fix: 2,921 tests passed and 5 skipped, e2e 171 of 171.
- **A commit with failing tests, amended before anything else built on it.** The `adapters/bridge` move skipped `from ".."` specifiers in 14 tests, and that commit was gated only on the bundle. Production was unaffected.

## Limits of this evidence

Moves are proved by bundle identity. The R1 and R5 JavaScript changes rest on reading the diffs, the unit suite and e2e. `LiveTop` in `ReviewCentre.tsx` had no failing unit test when its `startOf` was removed, and only the type checker caught it, so running-session rendering relies on e2e. No design review in Chrome was done, because the stylesheet and DOM-producing code are unchanged by intent and no visual change is expected.
