# Round log

One entry per round: what moved, behaviour pin result, review findings, the
commit. Newest last.

## R0 — standard, surveys, plan

- Added STANDARD, FINDINGS, PLAN, DEFERRED, survey/, `scripts/behaviour-pin`
  and its baseline (146 lines; two consecutive runs identical).
- Baseline: `go test -short ./...` green (5.6 s), `golangci-lint` 0 issues,
  coverage below 70 % in `cli` 26.8, `model` 16.4, `mcp` 62.9, `world` 66.2,
  `refconsumer` 68.2, `randutil` 68.8, `adapter` 69.9.
- Lint probe with `gocognit`/`gocyclo`/`nestif`/`dupl`/`paralleltest` enabled:
  production has 1 `nestif` (`jsonschema/validate_composition.go:51`) and 1
  `dupl` pair (`perturb/transform_timing.go:78-98`); the rest are test files
  (9 `gocognit`, 73 `paralleltest`, 4 `tparallel`).
