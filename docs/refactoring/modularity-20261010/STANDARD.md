# Modularity standard (v2)

Adapted from the sibling repository `agentic-stream`: its quality bar Q1–Q8,
architecture bar A1–A12, test bar T1–T12 and the "reference module" pattern of
`internal/authority`. That repository owns SQLite tables; this one owns
deterministic computation, JSON/JSONL artifacts, a wire transport and an MCP
surface. The *principles* carry over, the *layer names* are adapted. Nothing
here changes behaviour: the nine non-negotiables, the determinism tuple and
digest inputs are fixed.

**v2 (2026-10-10, owner direction):** every module is a facade over private
internal layers with no internal leaks, and tests are organised per layer.
This replaces v1's "pure cores stay one package".

## 1. Modules and shapes

A **module** is `internal/<name>`: a public *facade* package plus private
layers under `internal/<name>/internal/…`. Go itself forbids importing those
layers from outside the module; gates forbid the facade from leaking them.

| Module kind | Shape | Modules |
| --- | --- | --- |
| **Pure core** (rules, no I/O) | facade · `internal/domain` | `world`, `perturb`, `truth`, `audit`, `score`, `suite`, `deviceworld` |
| **Core with an edge** (rules plus files, sockets, timers) | facade · `internal/app` · `internal/domain` · one edge package per external system | `run`, `device`, `adapter`, `domain`, `sink`, `refconsumer` |
| **Surface** (protocol or CLI) | facade · `internal/app` · edge/protocol packages | `mcp`, `cli` |
| **Foundation** (stateless library every module uses) | one package, no layers | `canonical`, `jsonschema`, `model`, `randutil`, `schemas`, `wall` |

Foundations are the standard's explicit "N/A": the sibling repository leaves
its stateless libraries (`ids`, `clock`, `contractsv1`, …) single-package for
the same reason — a facade over a library whose whole export is the API only
forwards. They keep their gates (purity, no business imports) and are
protected by `ioEdges`.

### Layer responsibilities

- **Facade** (`internal/<m>`): the module's whole public contract. Files:
  `doc.go` (package comment: the business capability), `api.go` (aliases of
  domain values, constants, sentinel errors), `service.go` (`Config`, `New`,
  the facade type), `operations.go` (one documented delegating line per
  operation). No loops, no I/O, no decisions. Exported signatures name facade
  types only: a type defined in an internal layer appears through an alias
  declared in `api.go`, never as `domain.X` in a signature (**no leaks**).
  Required dependencies are constructor arguments; a missing one fails closed
  with a typed error (`New(…) (*T, error)`). The facade imports its own domain
  layer under the alias `layer` (the repository also has a module named
  `domain`). Re-exported catalogs are functions returning copies, never
  exported mutable variables. A constructor may instead fail closed by
  *substituting the refusing default* when that default simply refuses the
  guarded operation (`truth.NewStore(nil)` counts every run as open, so reveal
  is refused); it says so in its doc comment and has a test.
- **domain** (`internal/<m>/internal/domain`): vocabulary and every rule,
  including the module's in-memory aggregate and its invariants. Pure: no
  `os`, `net`, `os/exec`, clock read, entropy or goroutine. Time, randomness
  and sizes arrive as parameters or injected sources.
- **app** (`internal/<m>/internal/app`): use cases as a short sequence of
  domain verbs: validate → open resources → load → decide (domain) → persist
  or send (edge) → record. Never imports `os`/`net`/`os/exec` itself; talks to
  edges through the module's edge packages.
- **edge** (`internal/<m>/internal/<system>`): the only code that touches a
  file system, socket, process, timer or the wall clock, named after the
  external system (`durable`, `quiesce`, `wire`, `uds`, `files`). Decides
  nothing: a `WHERE`-equivalent selects, it never authorises.

Dependency direction inside a module: facade → app → domain; app → edges;
edges → domain types only; domain → nothing of the module. Between modules:
facade → facade, downward by layer (STANDARD M2).

## 2. Rules

Each rule has a named enforcement. "Gate" is a test in `test/architecture`.

| ID | Rule | Enforced by |
| --- | --- | --- |
| M1 | Every package states its responsibility in a package comment beginning `Package <name>`, is listed in the package map, and every module has `UBIQUITOUS_LANGUAGE.md` (terms with code names, retired words). | Gates `TestEveryPackageDocumentsItsResponsibility`, `TestPackageMapListsEveryPackage`, `TestEveryModuleHasUbiquitousLanguage` |
| M2 | Imports point only to a strictly lower layer; the direct-import allowlist is complete with no stale edges; nothing imports a surface except surfaces. | Gates `TestPackageDependencies`, `TestAllowedImportsHaveNoStaleEdges`, `TestImportsPointToStrictlyLowerLayers`, `TestSurfacesAreImportedOnlyBySurfaces` |
| M3 | Domain layers and foundations are pure: no I/O, entropy, `go` statement, wall-clock call or function value. Exceptions are `ioEdges` entries with scheduled debt. | Gates `TestIOStaysInDeclaredEdges`, `TestPureKindsHoldOnlyScheduledIODebt` |
| M4 | File, socket, process, entropy, goroutine and clock use sits only in files declared in `ioEdges` with a reason. App layers hold none. | Gate `TestIOStaysInDeclaredEdges` |
| M5 | A facade only delegates: every function (exported or not, `init` forbidden) is one statement calling `<layer>.F(…)` or `<receiver>.<field>.M(…)` where the field was declared with a layer type; arguments contain no function literal or nested call; a constructor is nil-guards then one return. `cli` and `mcp` hold wiring and protocol only. | Gate `TestFacadesFollowTheFacadeRules` (`facadeViolations`), proven by `TestFacadeRulesRejectEveryKnownBypass` |
| M6 | No leaks: exported signatures, fields and values contain no internal-layer type except an alias of a plain value record declared exported in `api.go` (the layer type must have no methods); no exported mutable variable (only `Err…` sentinels and constants); dot or blank imports of a layer are rejected; exported symbols used by nobody outside tests are removed. | Gate `TestFacadesFollowTheFacadeRules`, `make deadcode` |
| M7 | Records cross a boundary typed and parsed once with a closed field set; a raw document is kept only where a digest depends on its bytes. | Review |
| M8 | Production functions ≤ 15 body lines at one abstraction level, entry points first; files ≤ 300 lines; cognitive ≤ 15, cyclomatic ≤ 20, nested-`if` ≤ 3; no token clone of 75+. | Gates `TestProductionFunctionsStayWithinTheBodyLimit`, `TestGoFileSize`; lint |
| M9 | Every package with statements is ≥ 70 % covered (`-short`); every exported facade operation and every error branch that enforces an invariant has a test. | `make coverage-check` |
| M10 | A round changes structure only: RNG draw order, digest inputs, JSON shapes, error precedence, locks, command and delivery order are identical. A deliberate change is listed in PLAN first and proven by a regression test. | Existing oracles unchanged, `scripts/behaviour-pin`, golden digests, round review |
| M11 | No duplicated implementation of a job the repo already does; a wrapper whose body is one call is deleted unless a gate requires it. | Lint `dupl`; duplication scan |
| M12 | A module's internal layers match its row in §1 and are registered in the `packages` table with kind and layer; a module that gains an edge adds an edge package, not code in app or domain. | Gates `TestEveryPackageIsClassified`, `TestModuleShapeMatchesItsKind` |

## 3. Test organisation (T1–T10)

Adapted from the sibling's test bar. Current baseline (2026-10-10): 285
tests in 87 files, 14 files call `t.Parallel`, one `time.Sleep`, the same
fixture path constant copied into 7 packages.

| ID | Rule | Enforced by |
| --- | --- | --- |
| T1 | **Place.** A test lives with the code it proves, at the lowest layer that owns the behaviour, plus at most one integration path through the layer above. The facade keeps wiring and contract tests, not domain rules. Nothing at the repository root. Repository-wide gates live in `test/architecture`; cross-module acceptance flows live in `test/acceptance`. | Gate `TestNoTestsAtRepositoryRoot`; module rounds move tests |
| T2 | **Layers prove different things.** domain: every branch, boundary and rejection over values; app: use-case ordering and error precedence with fakes of edges; edge: framing, files, sockets; facade: configuration and the public contract; surfaces: product flows and CLI/MCP contracts, few and fast. | Review; coverage per package |
| T3 | **Name.** `TestSubjectDoesObservableThing`; the file is named for the subject; no round, phase or ticket number. | Gate `TestTestNamesCarryNoPlanningVocabulary` |
| T4 | **Assert.** Observable outcomes, got and want printed; error tests assert which error (`errors.Is/As` or the domain message), never `err != nil` alone. | Review; `TestErrorAssertionsNameTheErrorTheyExpect` (ratchet, added with the module that needs it) |
| T5 | **Deterministic.** A test never sleeps to wait for work; waits on a channel, condition or bounded poll bound to `t.Context()`; time comes from a virtual or injected clock. | Gate `TestTestsNeverSleep` |
| T6 | **Isolated and parallel.** Every test and subtest calls `t.Parallel()`; `t.TempDir`, `t.Context`, `t.Cleanup`, `t.Setenv`. A test that cannot be parallel says why in a `//nolint:paralleltest // <reason>`. No mutable package-level test state. | Lint `paralleltest`, `tparallel`, `usetesting`, `thelper` |
| T7 | **Fast.** `go test -short -race ./...` stays fast; no test sleeps; product acceptance tests are made faster, never skipped under `-short`. | Review; `make coverage-check` timings |
| T8 | **Covered.** ≥ 70 % per package; coverage is never raised with assertion-free tests. | `make coverage-check`; review |
| T9 | **Shared fixtures.** Inputs come from `testdata/`, the shipped `domains/` and `adapters/`, or a named builder in `internal/testsupport`; a path constant lives in one place; helpers call `t.Helper()`; setup repeated in two places moves into one helper. | Gate `TestFixturePathsAreDefinedOnce`; review |
| T10 | **No dead tests.** Delete tests that repeat another assertion at the same layer, test removed behaviour or assert a constant equals its literal. `t.Skip` only for `-short` or a missing optional tool, and it says which. | Review |

## 4. What was taken, what was changed

| agentic-stream | This repository | Why |
| --- | --- | --- |
| facade · app · domain · `store` | facade · app · domain · named edge packages (`durable`, `quiesce`, `wire`, `uds`, `files`) | No SQL; the edge is named after the external system. |
| No comments inside modules | Not adopted (DEFERRED P-02) | Churn without structural value; exported symbols stay documented. |
| `durableOwners` table ownership gate | `ioEdges` inventory plus an artifact-file ownership rule in `run` | The durable state here is run artifacts and the append-only ledger. |
| Per-module dated folders | One program folder; `ROUNDS.md` per round | Tracking in one place. |

## 5. Method

- Go-aware tooling for Go changes (`gopls rename`, `gofmt -r`, `goimports`,
  small `go/ast` programs); `sed` only for non-Go text.
- Move code with declaration bodies unchanged; record source comparison in
  `ROUNDS.md`. One module per round; tests move with the code they prove.
- A round = focused tests + `gofmt` + `go vet` + lint + architecture gates +
  `scripts/behaviour-pin` + independent review + one commit, hash in PLAN.
- `make ci-check` before handoff.

## 6. Constraints carried over

`AGENTS.md` Forbidden Changes remain: domains/adapters/effectors are data; no
consumer knowledge in the binary; no broker, physics engine, ORM, expression
language or plugin system; zero external dependencies except the MCP Go SDK;
no per-domain code branches.
