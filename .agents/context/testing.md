# Testing Context

## Authoritative Commands

Use these commands unless the task is documentation-only:

- `go test ./...`
- `go vet ./...`
- `make build`
- `make lint`
- `make ci-check`
- `make phase-gate` (`ci-check` plus `make architecture`; the simulation-canvas phase gate)
- `make function-length`
- `make coverage-check` (per-package floor, default 70%, ratchet file `scripts/coverage-floors.txt`)
- `make behaviour-pin` (diff against `test/pin/BEHAVIOUR_PIN.txt` after any structural change)
- `git diff --check`
- `pre-commit run --all-files` if `pre-commit` is installed

## Repository-Specific Behavior

- `make build` builds `cmd/streamsim` into `bin/`.
- `make ci-check` runs `function-length -> fmt-check -> docs-check -> tidy -> build -> vet -> lint-ci -> gate-ratchet -> test-ci -> deadcode -> vulncheck -> fuzz-soak`.
- `make test-ci` (`scripts/test-ci`) is the whole pipeline test tier: the suite runs once,
  `-short -shuffle=on -tags simdet` with cheap `-cover`; the per-package coverage floors are checked from
  that run's output (`scripts/check-coverage REPORT...`). Every pipeline (`make ci-check`, `make phase-gate`,
  each CI job) stays under 2 minutes of wall time.
- Long simulations run on demand only: a test that simulates hours of plant time or sweeps the whole
  component library must start with `if testing.Short() { t.Skip("<what it simulates>; on demand: make test-long") }`.
  `make test-long` runs the whole suite without `-short`; `make test-race` runs it under the race detector.
  Neither is part of `ci-check`, `phase-gate` or CI; run them when a change touches what they cover
  (long-horizon dynamics, the component library, goroutines and locks).
- `make test-short`, `make test-simdet` and `make coverage-check` remain as standalone
  targets for working on one concern; `ci-check` does not call them.
- `make fuzz-soak` runs every fuzz target side by side (`FUZZSOAKTIME`, default 5s each); `make fuzz`
  runs them one after another (`FUZZTIME`, default 10s). Add a target by appending to `FUZZ_NAMES` and
  setting `FUZZ_PKG_<name>` in the Makefile. `make soak` stays the explicit slow tier.
- In restricted environments, lint and loopback MCP tests can fail because they need cache writes or local sockets.
- `deadcode` and `govulncheck` are release-gate tools; the Makefile fails closed when they are missing.

The strict source checker covers production declarations, methods and callbacks,
including build-tagged files, with no exemptions. Architecture tests (`test/architecture`) enforce the
300-line file and 15-line function limits, the direct-import allowlist, the declared
package kind and layer table, the I/O and clock edge inventory (`ioEdges`), package
comments and the package map.

## Test Organisation

The bar is T1–T10 in
[`governance/MODULARITY_STANDARD.md`](../../governance/MODULARITY_STANDARD.md#3-test-organisation-t1t10).
In short:

- A test lives at the lowest layer that owns the behaviour: `domain` proves
  rules over values, `app` proves use-case order, edges prove files, sockets
  and framing, the facade proves its public contract. Cross-module flows that
  run the real binary live in `test/acceptance`; repository-wide gates in
  `test/architecture`. Nothing at the repository root.
- Shipped domains, adapters and short socket paths come from
  `internal/testsupport`; a test never spells `../../domains` itself.
- A test that only runs one deterministic simulation on one goroutine (generating, auditing or scoring a
  world) may call `testsupport.SkipUnderRace(t, "<what it computes>")` right after `t.Parallel()`: the race
  detector slows such a run ten- to twenty-fold and has nothing to find in it, and the test still runs in
  full in the non-race pass. `test/architecture` refuses the call in a test that starts a goroutine. Tests
  with goroutines, locks under contention, sockets or processes stay under `-race`.
- Tests that must generate suites or audit worlds keep them small: a cheap shipped domain
  (`shipment-exception`), `n=1`, or a shrunken profile (entities, trace length) when the property is
  independent of scale. Audits run two full worlds per candidate, so scale is the cost.
- Acceptance tests share one binary build per ldflags value (`test/acceptance/binary_test.go`).
- Every `Test*` calls `t.Parallel()` first (or `//nolint:paralleltest // why`);
  no test sleeps; negative tests assert which error (`errors.Is/As` or the
  message), not only that one occurred.

## Test Quality Bar

For production-code changes:

- write a failing or expectation-setting test before implementation when feasible
- add or update `*_test.go` files in every modified package
- cover new exported behavior and meaningful branches
- use table-driven tests where that improves clarity
- use `t.Helper()` in test helpers
- use `t.Context()` for test contexts when appropriate
- avoid network calls in unit tests

If test-first is not feasible:

- say why explicitly in the plan or handoff
- still add the proving test in the same change
- make sure the test demonstrates the behavior change, not just execution

For documentation-only changes:

- run the narrowest useful checks
- still run `git diff --check`
- explain skipped commands in the final handoff

## Failure Handling

- Fix failures caused by your changes before stopping.
- If a command fails for an existing repo issue or environment restriction, say so clearly and do not misreport it as validated.
