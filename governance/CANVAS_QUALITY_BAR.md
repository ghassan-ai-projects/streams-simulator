# Quality bar — simulation canvas goal

- **Date**: 2026-10-12 (revised after the P0 review, [reviews/P0.md](../docs/simulation-canvas/reviews/P0.md))
- **Applies to**: every phase of [GOAL.md](../docs/simulation-canvas/GOAL.md), in this repository: Go simulator changes
  and the `canvas/` TypeScript application.
- **Precedence**: [AGENTS.md](../AGENTS.md) and the modularity
  [STANDARD.md](MODULARITY_STANDARD.md) apply unchanged; this file adds
  to them and never relaxes them.

## 1. Principles

1. **No workarounds, no shortcuts.** No stub shipped as complete, no `TODO`/`FIXME` in shipped
   code, no disabled or focused test, no suppressed lint without a specific rule name and a
   reason, no mock of the binary in end-to-end or acceptance tests, no "temporary" path. All
   enforced by the lint rules in §2 and §3.
2. **The instrument stays trustworthy.** The nine non-negotiables, the determinism tuple and
   sealed truth are never weakened. A change to an oracle, digest input, RNG draw order,
   artifact shape, CLI stdout contract or pipeline order is stated in the phase spec first,
   proven by a regression test, and recorded as deliberate in the behaviour pin.
3. **Design before code.** A phase that changes a contract, schema, protocol or trust boundary
   writes or corrects its spec first, in the same phase, before implementation.
4. **One source of truth.** The canvas computes nothing the binary owns. Domain validation,
   formatting, simulation, scoring, verification and scenario export happen only in
   `streamsim`.
5. **Evidence over assertion.** A phase is done when its gates pass, its evidence file records
   the real results and its reviewers' findings are resolved.
6. **The device path stays green.** Until Agentic Stream migrates (next goal), the simulator's
   device emulator, `deviceworld`, the vendored device contract and the adapters are not
   removed, and the device-path gate (§4) passes in every phase.
7. **No consumer knowledge in products.** No consumer name, schema or behaviour in Go sources
   or canvas sources; external parties exist only as data (`connections.json`) and in docs.
   The vendored device contract is a recorded, time-boxed exception that ends with the
   Agentic Stream migration.

## 2. Go (simulator) bar

Everything in AGENTS.md and STANDARD.md (function ≤ 15 body lines, file ≤ 300 lines, module
kinds and layers, facade rules, no leaks, `ioEdges`, import allowlist, T1–T10, ≥ 70 % coverage
per package), plus:

| Rule | Enforced by |
|---|---|
| No `TODO`/`FIXME`/`XXX` in production Go | `godox` in `.golangci.yml` (added in P1) |
| Every `//nolint` names its linter and gives a reason | `nolintlint` with `require-explanation`, `require-specific` (added in P1) |
| New modules (journal, report, scenario, plant-wire node, reference host) ≥ 85 % statement coverage | rows in `scripts/coverage-floors.txt`, `make coverage-check` |
| Every new parser of untrusted input (scenario file, journal reader, plant-wire records) has a fuzz target in `make fuzz` | Makefile, review |
| New package registered: package comment, package map, `UBIQUITOUS_LANGUAGE.md`, packages table (kind, layer), import allowlist, `ioEdges` for any file/socket/clock use | `test/architecture` gates |
| Every behaviour change recorded in `CHANGELOG.md`; `model.SimVersion` (part of the determinism tuple) bumped when a change can alter a run's trace, artifact or digest | review |
| No new Go module dependency | `go.mod` diff, review |

## 3. Canvas (TypeScript) bar

| Area | Requirement | Enforced by |
|---|---|---|
| Language | `strict`, `noUncheckedIndexedAccess`, `exactOptionalPropertyTypes`, `noImplicitOverride`; no `any`, no non-null assertion, no `@ts-ignore`/`@ts-expect-error`. Library types that conflict with `exactOptionalPropertyTypes` are wrapped once in the `adapters` layer, each wrapper documented | `tsc --noEmit`, lint |
| Lint | Type-aware strict rule set; zero warnings; `no-warning-comments` for TODO/FIXME; every disable comment names the rule and has a description | `eslint --max-warnings 0` with `eslint-comments/require-description` |
| Format | One formatter, checked in CI | `prettier --check` |
| Architecture | Layers `model` (pure types and rules) → `app` (state, use cases) → `adapters` (bridge client, journal follower, file access) → `ui` (components); no upward import; `ui` never touches transport; `model` imports nothing outside itself | dependency-cruiser in CI |
| Size | Files ≤ 300 lines; functions and components ≤ 40 physical lines from the line of the opening brace to the closing brace, including comments and blank lines; cyclomatic ≤ 10, cognitive ≤ 15 | lint (`max-lines`, `max-lines-per-function` with `skipBlankLines: false, skipComments: false`, `complexity`, `sonarjs/cognitive-complexity`) |
| Duplication, dead code | No clone ≥ 75 tokens; no unused export, file or dependency | `jscpd`, `knip` |
| Data at boundaries | Bridge messages, journal records, layout and connections files are parsed once in `adapters` against their schema; the app sees typed values. Domain documents are never validated in TypeScript: the binary validates them. Where a schema is shared with Go, the same valid/invalid fixtures run through both validators | unit tests, review |
| Tests | Unit tests for `model` (≥ 90 % lines) and `app` (≥ 80 %), generated files excluded from coverage; component tests for `ui`; end-to-end tests against the **real** `streamsim` binary and real files. Bridge unit tests may use a fake director only if a contract test runs the same calls against the real binary | Vitest thresholds, Playwright |
| Test hygiene | No `.only`, no disabled test, no `waitForTimeout` or fixed sleeps; time through fake timers or awaited conditions | lint (`vitest/no-disabled-tests`, `vitest/no-focused-tests`, `playwright/no-wait-for-timeout`, `playwright/no-focused-test`) |
| Accessibility | WCAG 2.2 AA. Automated axe checks on every screen in component and e2e tests, plus a manual keyboard and screen-reader checklist in every UI phase review | axe, review record |
| Visual | Visual baselines generated and compared only in the pinned Linux CI container (no local baseline comparison: no container runtime on the development machine); every UI phase updates its baselines | Playwright in CI |
| Performance | A generated stress fixture (≥ 55 channels, ≥ 19 faults, a high-rate journal) from P8b; numeric budgets (first render, live update throughput, bundle size, cumulative layout shift) written into this file in the P9 spec and enforced in CI from P9 | Playwright traces, bundle check |
| Security (local bridge) | Threat model written in the P8b spec, one test per threat: `127.0.0.1` only; token never in URLs or logs; Origin and Host checks; workspace-scoped file access refusing traversal and symlink escape, with atomic writes and conflict detection; size limits; spawned `streamsim` path pinned and environment scrubbed; operator capability tokens never reach the browser or logs; strict CSP on the served build; outbound health checks only to URLs in `connections.json` | bridge tests |
| Dependencies | Exact versions, lockfile committed, `npm ci --ignore-scripts`, each dependency justified in `canvas/DEPENDENCIES.md`, `npm audit --audit-level=high` clean | CI |
| Isolation | Go never imports `canvas/` | `test/architecture` |

All canvas gates run through `make canvas-check`.

## 4. Standing gates (every phase)

| Gate | Content |
|---|---|
| `make ci-check` | The repository's full pipeline |
| `make canvas-check` | From P8a onward |
| `make architecture` | Enola, against a baseline pinned before the phase's first structural edit |
| Device path | `internal/device`, `internal/deviceworld` and CLI device tests, the device conformance fixtures, and the pinned `device serve` transcript (`test/acceptance/device_serve_test.go`, added in P1), all inside `make ci-check` |
| Findings | No finding in [FINDINGS.md](FINDINGS.md) assigned to a finished phase is open (`test/architecture`) |
| Consumer knowledge | No consumer name in Go or canvas sources outside the recorded exception (`test/architecture`, added in P1) |
| Whitespace | `git diff --check` |

`make phase-gate` runs all of them (added in P1). CI runs on pull requests to `main`; until a
pull request exists for this branch, the phase evidence file records the local `make
phase-gate` output.

## 5. Contracts

| Contract | Location | Embedded in the binary | Change rule |
|---|---|---|---|
| CLI output (`cli-output-v0.1`) | `contracts/` | No (tests validate outputs) | Pre-release: additive changes in place; breaking changes bump the minor version and remove the old file |
| Scenario file (`scenario-v0.1`) | `contracts/` | Yes | Same |
| Run journal (`run-journal-v0.1`) | `contracts/` | Yes | Same |
| Run report (`run-report-v0.1`) | `contracts/` | No | Same |
| Domain spec (`domain-spec-v0.2`) | `contracts/` | Yes | Same |
| Canvas layout, connections | `canvas/schemas/` | — (canvas-owned) | Same |
| Plant wire v1 | `contracts/plant-wire/` | Yes (byte-identical copy) | Exception to `contracts/`: self-contained for cross-repository sync; pinned digest per repository and a safety-field floor test |
| Device wire v1 | `internal/device/contract/` | Yes | Vendored from Agentic Stream; time-boxed exception (§1.7) |

## 6. Phase protocol

1. **Spec**: write or correct the design or contract the phase depends on, including the module
   kind of every new component and any deliberate behaviour change.
2. **Plan**: list files, tests, gates and the findings the phase closes in PLAN.md.
3. **Tests first** where feasible; otherwise the proving test lands in the same phase.
4. **Implement** within the layer rules, with documentation, registration and fuzz targets.
5. **Gates**: `make phase-gate`; record real results in `evidence/Pn.md`.
6. **Independent review** by at least two reviewers with different lenses (architecture and
   standards; correctness, safety and determinism; UI phases add UX and accessibility). They
   read the diff and the specs, not the author's summary. Findings and dispositions go to
   `reviews/Pn.md`.
7. **Resolve** every finding: fixed with evidence, or recorded as not a defect with evidence,
   or carried to a named later phase in FINDINGS.md. Re-run the gates after fixes.
8. **Commit** the phase with a message stating scope, gates and review outcome. The next
   phase's first commit records the hash in PLAN.md.

A phase that cannot meet the bar stops and reports why; it never lowers the bar to finish.
