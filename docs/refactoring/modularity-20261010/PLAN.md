# Plan

Baseline: `867026a` (`main` merge of PR #16), branch `improve-modularity`.
Tests, vet and `golangci-lint` are green at baseline; every package passes
`go test -short`. Production functions and files already meet the 15-line and
300-line bars; the program adds *structure*, not more splitting.

## Rules of every round

1. Plan the round here (scope, hazards, proof) before editing.
2. Change structure only (STANDARD M10). Intentional corrections are not in
   this program: they are in [DEFERRED.md](DEFERRED.md).
3. Proof, in this order, before the commit:
   `gofmt -l` → `go build ./...` → `go vet ./...` → `go test -short ./...` →
   `go test -tags simdet ./...` → `make function-length` → `golangci-lint run` →
   `scripts/behaviour-pin | diff docs/refactoring/modularity-20261010/BEHAVIOUR_PIN.txt -`.
4. An independent review subagent reads the diff against STANDARD.md and the
   behaviour hazards; its findings are fixed or recorded in `ROUNDS.md`.
5. One commit per round (`refactor(<pkg>): …`), staged by explicit path (the
   working tree carries unrelated user changes in `docs/README.md` and
   `docs/simulation-ideas/`). The commit hash goes into the status table.
6. A relocated declaration keeps its body byte-for-byte; `ROUNDS.md` records
   the source comparison. `make ci-check` runs before handoff, not each round.

## Rounds

Phase A (R0–R7) established the standard, gates and local clean-ups on the
flat package layout. **Direction change 2026-10-10:** every module becomes a
facade over private layers with no leaks, and tests are organised per layer
(STANDARD v2). Phase B migrates one module per round, bottom-up, so callers
keep importing the same facade path with the same exported API; Phase C
finishes test organisation.

### Phase A — done

| Round | Scope | Commit |
| --- | --- | --- |
| R0 | Standard, surveys, plan, deferred register, behaviour pin | `3e81cf2` |
| R1 | Kind/layer table, I/O-edge inventory, doc gates, 15-line gate, lint ratchet, coverage floors | `52ea773` |
| R2 | Foundations: shared schema compile/format helpers, dead `Picker`/`TimeNS` | `a3f6858` |
| R3 | `domain` file loading isolated in one declared file | `33d757f` |
| R4 | `adapter/conformance` extracted; `adapter.Load` isolated | `f735d8e` |
| R5 | `world`: invocation values, dead API, responsibility files | `ddf8535` |
| R6 | `perturb`: golden streams, table-driven dispatch | `87363c0` |
| R7 | `truth`: store/record split, fail-closed store, `SetupCall` to `model` | `b9b7909` |

### Phase B — module migration (facade · layers · no leaks)

Per round: write the module's `UBIQUITOUS_LANGUAGE.md`; create the layers
from its row in STANDARD §1; move code with bodies unchanged; write the
facade (aliases in `api.go`, delegating operations); move each test to the
layer it proves; register packages, allowlist, `facadeSpecs`; keep the
exported API other modules use (narrow only what nobody uses); behaviour pin
and goldens unchanged.

| Round | Module | Shape | Specific hazards |
| --- | --- | --- | --- |
| M0 | Gates and template | `facadeSpecs`, `TestFacadesOnlyDelegate`, `TestFacadeSignaturesNameNoInternalTypes`, `TestModuleShapeMatchesItsKind`, `TestEveryModuleHasUbiquitousLanguage`; test-bar gates T1/T3/T5/T9; `.agents/prompts/module-refactor.md` | gates start with the migrated-module list and grow with it |
| M1 | `perturb` | facade · domain | RNG draw order; golden digests (R6) |
| M2 | `truth` | facade · domain | label bytes; sealed-store semantics |
| M3 | `world` | facade · domain | substream names and draws; call log; idempotency order |
| M4 | `score` | facade · domain; `score.Evidence` replaces `*run.Run` input (drops the `run` import) | scorecard JSON byte-identical; nil-vs-empty ledger/calls |
| M5 | `audit` | facade · domain | audit verdict bytes |
| M6 | `suite` | facade · domain | generation RNG order, suite JSON bytes |
| M7 | `deviceworld` | facade · domain | binding validation, safe-stop plant |
| M8 | `sink` | facade · inproc/file/httppush edges | byte streams, `Close` returns |
| M9 | `domain` | facade · domain · files edge | domain digest, error text, catalog order |
| M10 | `adapter` | facade · domain · files edge; `conformance` becomes its own module | adapter digest bytes, render output |
| M11 | `device` | facade · app · domain · contract · wire/uds edges | admission order, frame bytes, lock scope |
| M12a | `run` | facade · app (whole package body moved verbatim; ledger/artifact/timer I/O declared as debt) | RunID, replay, `-race` quiescence |
| M12b | `run` | `durable` and `quiesce` and `clock` edges; pure `domain` rules (identity digests, replay command decoding, artifact assembly); app free of I/O | ledger flush/fsync points, artifact write order, permissions |
| M12 | `run` (summary row; delivered by M12a and M12b) | facade · app · domain · durable · quiesce edges | RunID, ledger flush/fsync points, artifact write order, replay |
| M13 | `refconsumer` | facade · app · domain · mcpclient edge | detection bytes, verdict JSON |
| M14 | `mcp` | facade · app (director/operator use cases) · protocol edge | tool schemas, error codes and precedence |
| M15 | `cli` | facade · app (commands) · edges; `cmd` unchanged | flag defaults, exit codes, output text, manifest bytes |

A module that proves larger than one reviewable diff is split (`M12a`,
`M12b`) and the table updated, not rewritten.

### Phase C — test organisation

| Round | Scope |
| --- | --- |
| T1 | `internal/testsupport`: one named loader for the shipped/example domains, adapters and capability catalogs replaces seven copies of the fixture path; `TestFixturePathsAreDefinedOnce` |
| T2 | Hygiene: remove the one `time.Sleep`; `paralleltest`, `tparallel`, `usetesting`, `thelper` on; `t.Parallel` everywhere or a reasoned nolint |
| T3 | Assertions name the error they expect (ratchet), no-op/duplicate tests deleted, names follow T3 |
| T4 | `test/acceptance`: cross-module flows (run → verify → score, mcp loop) live there, not in a module |
| T5 | Coverage to ≥ 70 % per package (`cli`, `model`, `mcp`, `world`, `refconsumer`, `adapter`), floors file empty |
| F | Final: full `make ci-check`, uncached race suite, module/test ratings, handoff |

## Deliberate changes (STANDARD M10)

Each is unreachable on production paths or refuses where the old code
dereferenced nil; every one has a regression test.

| Round | Change | Test |
| --- | --- | --- |
| R7 | `truth.Store` takes its open-run check at construction; a nil check counts every run as open (reveal refused) | `TestStoreWithoutAnOpenRunCheckOrBackingStoreFailsClosed` |
| M1 | `perturb.New` returns `(*Layer, error)`, `ErrNoSpec` for a nil spec; `perturb.Names` is a function returning a copy | `TestNewRefusesAMissingSpec`, `TestNamesIsTheCatalogAndAModifiableCopy` |
| M2 | `truth.NewSolver` returns `(*Solver, error)`; `BuildRecord` refuses a nil spec or solver; a zero or nil `Store` returns `ErrNoStore`; the record no longer aliases the caller's perturbation slice | `TestBuildRecordRefusesMissingInputsAndUnknownFaults`, `TestBuildRecordLabelsAScenarioFromTheSolvedOnsets` |
| M3 | `world.New` returns `ErrNoSpec` for a nil spec | `TestNewRefusesAMissingSpec` |
| M4 | `score.Score` returns `ErrNoLabel` for a nil label; a nil domain in the evidence skips fault recovery levels | `TestScoreRefusesAMissingLabelAndToleratesAMissingDomain` |
| M7 | `deviceworld.New` returns `(*Plant, error)`, `ErrNoWorld` for a nil world | `TestNewRefusesAMissingWorld` |
| M6 | `suite.Generate` returns `ErrNoDomain` for a missing domain | `TestGenerateRefusesAMissingDomain` |
| M10 | `adapter.NewEngine` returns `ErrNoAdapter` for a nil adapter; `Engine.Meta` (no caller) removed | `TestNewEngineRefusesAMissingAdapter` |
| M7–M10 | zero-value `deviceworld.Plant`, `domain.Catalog`, `sink.File`, `sink.HTTPPush` return `ErrNoPlant`/`ErrNoCatalog`/`ErrNotConstructed` instead of panicking (the old `Plant{}` returned `ErrPlantUnavailable`) | `TestZeroValueFileAndHTTPPushRefuseInsteadOfPanicking` and facade guards |
| M5 | `audit.NewPanel` returns `(*Panel, error)`, `ErrNoSpec` for a nil spec | `TestNewPanelRefusesAMissingSpec`; `audit.Audit` on a nil/zero panel returns `ErrNoPanel` (`TestAuditRefusesAMissingPanel`) |

## Not in this program

Everything in [DEFERRED.md](DEFERRED.md), in particular: any fix to D-01…D-40,
the wall-clock seam decision (P-01), RFC 8785 adapter digests (D-36), unifying
the two delivery paths (D-18), and world read-purity (D-14).

## Status

| Round | State | Commit | Notes |
| --- | --- | --- | --- |
| R0–R7 | done | see Phase A | R6 `87363c0`, R7 `b9b7909` |
| M0 | done | see git log | gates, template, prompt |
| M1 | done | `4a05a3b`, hardening `see git log` | perturb template; gates hardened after review |
| M2 | done | `b5674df` | review follow-up in M3 commit |
| M3 | done | `463fb65`, follow-up `e88cc57` | gate hardened, doc links gate |
| M4 | done | `2f5cf24`, follow-up `see git log` | score takes Evidence |
| M5 | done | `820e025` | audit; review follow-up in `see git log` |
| M6 | done | `ff2f3a3` | suite; review follow-up with M5 |
| M7 | done | `c522b77` | deviceworld |
| M8 | done | `28c648a` | sink |
| M9 | done | `59c3a45` | domain |
| M10 | done | `2d44673` | adapter |
| M11 | done | `42d0dbf` | device |
| M12a | done | `92d803b` | run facade over app |
| M12b | done | `3b40255` | run edges and domain |
| M13 | done | `13fa2e2` | refconsumer |
| M14 | done | `7b9def3` | mcp |
| M15 | done | `f28d88e` | cli |
| M16 | done | `see git log` | adapter conformance |
| T1–T5, F | pending | | |
