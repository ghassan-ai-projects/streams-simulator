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

## R1 — gates and lint ratchet

**Added** (`test/architecture`): `packages` kind+layer table with
`TestEveryPackageIsClassified`, `TestImportsPointToStrictlyLowerLayers`,
`TestFoundationsImportOnlyFoundations`, `TestSurfacesAreImportedOnlyBySurfaces`,
`TestAllowedImportsHaveNoStaleEdges`; `ioEdges` inventory with
`TestIOStaysInDeclaredEdges` and `TestPureKindsHoldOnlyScheduledIODebt`;
`TestEveryPackageDocumentsItsResponsibility`, `TestPackageMapListsEveryPackage`;
the 60-line review table became a strict 15-line gate that `go test ./...`
enforces. Lint: `gocognit`, `gocyclo`, `nestif`, `dupl` on, `new-from-rev`
removed (the old config showed only issues new since `HEAD~1`; the whole tree
is now checked and reports 0). `make coverage-check` with a ratcheting floors
file, part of `ci-check`.

**Production edits (behaviour-neutral):** `perturb.rewriteObservedTimes`
replaces two cloned closures; `jsonschema.validateCondition` flattened;
`model.validateAgainst` replaces two cloned validators (error text identical).

**Injection proofs** (each gate failed on a deliberately added violation, then
the file was removed): `import "os"` in `world` → `TestIOStaysInDeclaredEdges`;
`var f = time.Now`, dot-import of `time`, a `go` statement → same gate;
`world` importing `perturb` → `TestPackageDependencies` and
`TestImportsPointToStrictlyLowerLayers`; a new unclassified package → classify,
dependency, documentation and map gates; a 17-line function →
`TestProductionFunctionsStayWithinTheBodyLimit`.

**Review (independent agent) — fixed before commit:** clock detection missed
function values and dot imports (HIGH); import map collapsed repeated blank/dot
imports; package-map match was prefix-based (`device` satisfied by
`deviceworld`); debt prefix check inverted; coverage script skipped packages
without tests and did not ratchet; two tautological tests removed; stale
comments in `.golangci.yml`, `ci.yml` and `AGENTS.md`; test names in STANDARD
corrected. Not fixed, recorded: mixed skip-directory rules between older gates
(`.git` only) and the new loader (`testdata`, `.enola`); `runtime`
(`model.CurrentPlatform`) is not treated as a capability.

**Proof:** `gofmt`, `go vet`, `go test -short`, `-tags simdet`,
`make function-length`, `golangci-lint` 0 issues, `scripts/check-coverage`,
`scripts/behaviour-pin` identical to baseline.
