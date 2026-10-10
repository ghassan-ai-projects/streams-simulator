# Testing Context

## Authoritative Commands

Use these commands unless the task is documentation-only:

- `go test ./...`
- `go vet ./...`
- `make build`
- `make lint`
- `make ci-check`
- `make function-length`
- `make coverage-check` (per-package floor, default 70%, ratchet file `scripts/coverage-floors.txt`)
- `make behaviour-pin` (diff against `docs/refactoring/modularity-20261010/BEHAVIOUR_PIN.txt` after any structural change)
- `git diff --check`
- `pre-commit run --all-files` if `pre-commit` is installed

## Repository-Specific Behavior

- `make build` builds `cmd/streamsim` into `bin/`.
- `make ci-check` runs `function-length -> docs-check -> tidy -> build -> vet -> lint-ci -> test-short -> test-simdet -> coverage-check -> deadcode -> vulncheck -> fuzz-soak`.
- In restricted environments, lint and loopback MCP tests can fail because they need cache writes or local sockets.
- `deadcode` and `govulncheck` are release-gate tools; the Makefile fails closed when they are missing.

The strict source checker covers production declarations, methods and callbacks,
including build-tagged files, with no exemptions. Architecture tests (`test/architecture`) enforce the
300-line file and 15-line function limits, the direct-import allowlist, the declared
package kind and layer table, the I/O and clock edge inventory (`ioEdges`), package
comments and the package map.

## Test Organisation

The bar is T1–T10 in
[`docs/refactoring/modularity-20261010/STANDARD.md`](../../docs/refactoring/modularity-20261010/STANDARD.md#3-test-organisation-t1t10).
In short:

- A test lives at the lowest layer that owns the behaviour: `domain` proves
  rules over values, `app` proves use-case order, edges prove files, sockets
  and framing, the facade proves its public contract. Cross-module flows that
  run the real binary live in `test/acceptance`; repository-wide gates in
  `test/architecture`. Nothing at the repository root.
- Shipped domains, adapters and short socket paths come from
  `internal/testsupport`; a test never spells `../../domains` itself.
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
