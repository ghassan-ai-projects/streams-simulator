# Testing Context

## Authoritative Commands

Use these commands unless the task is documentation-only:

- `go test ./...`
- `go vet ./...`
- `make build`
- `make lint`
- `make ci-check`
- `make function-length`
- `git diff --check`
- `pre-commit run --all-files` if `pre-commit` is installed

## Repository-Specific Behavior

- `make build` builds `cmd/streamsim` into `bin/`.
- `make ci-check` runs `function-length -> docs-check -> tidy -> build -> vet -> lint-ci -> test-short -> test-simdet -> deadcode -> vulncheck -> fuzz-soak`.
- In restricted environments, lint and loopback MCP tests can fail because they need cache writes or local sockets.
- `deadcode` and `govulncheck` are release-gate tools; the Makefile fails closed when they are missing.

The strict source checker covers production declarations, methods and callbacks,
including build-tagged files, with no exemptions. Architecture tests enforce the
300-line file limit and direct package ownership. During the current source-only
refactoring program, the user requires test files to remain unchanged; validate
behavior with existing regression, replay and analytic-oracle tests.

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
