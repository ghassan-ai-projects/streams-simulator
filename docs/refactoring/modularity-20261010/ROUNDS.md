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

## R2 — foundations

- `randutil`: `Picker` deleted (no production caller; its map-order float sum
  was a determinism trap, D-37). Tests for `Intn`, `Exp`, `Lognormal` added
  (coverage 72.9 → 92.3 %, floor entry removed; the 68.8 % of the first survey was a single-run variance, the Picker tests sat in the denominator).
- `jsonschema`: `CompileJSON` and `FormatErrors` (new `document.go`, with
  tests) replace three `mustAny` and two `formatErrs` copies in `domain`,
  `adapter` and `model`. Bodies were identical (the `model` copy differed only in its reader type); the 10-error bound is now a
  named constant. Behaviour difference: a malformed *embedded* schema returned
  by `CompileJSON` is an error rather than a panic — unreachable, the
  embedded schemas are proven byte-identical to the contracts and compiled by
  every test.
- `model`: unused `TimeNS` type deleted; package comment corrected (it
  validates the consumer verdict and run artifact). Decision: validators stay in
  `model`. The `model → jsonschema/schemas` edge is foundation→foundation (the
  standard permits it), both callers are in `run`, and moving them would only
  relocate error strings that are part of the contract. Schema-enum constants
  (`SinkBroker`, `TimeScaled`, `AdmissionAccepted`, `SolveAnalytic`) are
  as unreferenced as `TimeNS` but stay as a deliberate exception: they mirror
  the committed contract enums (`run-artifact-v0.1`, verdict, ground truth) and
  document the closed vocabulary; `TimeNS` mirrored nothing.
- Not done on purpose: caching compiled embedded schemas (performance only,
  adds shared state), moving `canonical.CapabilityCatalogDomain` (R9, with the
  device split), `wall` (decision P-01).
- Coverage floor for `model` raised 18 → 20.

## R3 — domain

- `Load`, `LoadAll`, `domainPaths`, `loadPaths` moved verbatim into
  `internal/domain/file.go`, now the package's only `os` user and its single
  `ioEdges` entry (no debt). `domain` is reclassified `edge-core`.
- Decision change vs the R0 plan: moving the loaders to `cli` was rejected. about 29
  `domain.Load/LoadAll` references (3 in `cli`, 17 in tests of 8 packages, 9 in
  `domain`'s own tests) use it; relocating them buys no
  gate (the file-level gate already isolates the I/O) and the sibling pattern
  keeps small readers beside their rules. Revisit only if a second external
  system appears.
- Deleted `Compiled.FaultNames/EffectorNames/ProfileNames` (no callers).
  `HasProfile` stays (used by shipped-domain tests).
- New `file_test.go`: sorted top-level `.json` only, nested dir and non-JSON
  ignored, one bad domain fails the whole load, error text names the missing
  path (`domain: read …`, `domain: list …`).

## R4 — adapter

- New package `internal/adapter/conformance` (layer 4, edge-core) owns
  `adapter verify`: `Verify`, `Result` (was `adapter.VerifyResult`; JSON keys
  unchanged), output-schema validation, golden comparison and the fixture-file
  branch. Files moved with bodies unchanged except package, the
  `adapter.Load`/`adapter.NewEngine`/`adapter.FixtureEvents` qualifiers and the
  type rename. `cli/catalog.go` is the only caller.
- `adapter` keeps the embedded fixture (`FixtureEvents`, used by its own
  tests) and `Load` now sits in `file.go`, the package's one file-system edge;
  `adapter` is reclassified `edge-core`.
- Tests moved with the code: `TestShippedAdaptersConform`,
  `TestVerifyDetectsTampering`, `TestValidateStrictObservedOrder`,
  `TestVerifyRejectsNonMonotonicFixture`. Added: golden divergence (first byte
  reported), schema-violating record, missing golden file, `adapter.Load`
  success and missing file.
- Coverage: moving verification out lowered `adapter` from 69.9 to 68.3 (the
  moved code was well covered); floor reset to 68 and tracked for R15.
  `conformance` 82.4 %.
- Known and untouched (DEFERRED D-30..D-32): empty-fixture panic, `hash_suffix`,
  verify semantics for absent `output_schema`/`golden`.
- One error-text edit, unreachable: the embedded-fixture failure in
  `conformance.loadFixture` now reads `adapter: embedded fixture: …` (wrapcheck;
  the embedded fixture is covered by tests). All reachable `adapter verify`
  messages are unchanged.

## R3/R4 review follow-up

Independent review found no High issue. Fixed: the LoadAll ordering test did
not test order (now two invalid files, the error must name `a.json` only); the
bad-domain test now pins the `streamsim: ` prefix; the misnamed
`TestVerifyDetectsTampering` (it tampered with nothing) was removed in favour
of the real tamper tests, which now assert the exact line/column and add a
truncated-golden case. Doc nits corrected (82.4 %, reference counts, stale
`verify*.go` paths in DEFERRED, domains-and-adapters.md). The PLAN hazard
"VerifyResult JSON keys" is vacuous: no caller marshals `Result` (cli prints a
hand-built map); the tags are kept.

## R5 — world

- Deleted (no caller anywhere, tests included): `PendingEvents`,
  `HiddenStateSnapshot` with `sortedStateNames`, `ErrEffectorRefused`.
  `DynamicsFor` unexported (`dynamicsFor`), `AddEntity` and `addEntity` lose the
  ignored `params` argument (one production caller in `run`).
- Unexported by `gopls rename` (build, vet and tests green): `World.Seed`,
  `ClockNS`, `Noiseless`, `EmitDisabled`; `Entity.RetiredNS`, `States`. None was
  read outside the package.
- `recordCall`'s 11 positional parameters and the repeated
  `(effector, entityID, commandID, args, atNS)` tuple became `invocation` and
  `callOutcome` values; evaluation order and every call-log field are
  unchanged. `Advance` has named results (`emitted, effectsApplied`) and a doc
  line; dead `atNS` removed from `pickFailureMode` and `ackLatency`. The
  failure-mode substream names and draw order are untouched.
- File moves (declarations verbatim): `registry.go` (entity registry),
  `churn.go` (birth/retire/lifetime), kick counting into `effects.go`,
  `StateValue` into `observe.go`, `quantize` into `reading_values.go`,
  `dynamicsFor` into `dynamics.go`.
- New `effector_calls_test.go` pins the call-log contract: identity and
  outcome fields, interlock refusal record, idempotent replay adds no record
  and re-executes after the window. It exposed D-42 (replay loses the effect
  ETA), recorded, not fixed.
- Coverage 66.2 → 68.0 %; floor 68.

### R5 review follow-up

No High/Med finding. Fixed: write-only `Entity.retiredNS` removed (the field
and its one assignment); PLAN wording corrected to *replay-before-interlock*;
added the refusal-is-not-cached test. Not pinned: replay-before-interlock with
the interlock newly holding (needs a time-varying hidden state in the test
spec); covered today by `effector_order_test` and the behaviour pin.

## R6 — perturb

- Characterization first: `golden_test.go` pins the digest of the delivered
  stream for each of the 19 perturbations alone and for all together over a
  60-event stream (fixed seed, flush every 15 events). Recorded on the
  unmodified code, re-run unchanged after the refactor. (A first draft left
  `unit_mismatch`, `non_monotonic` and `gross_backfill` as no-ops; the stream
  now carries a unit, observed-after-event times and a backfill window so every
  digest is distinct.)
- `transforms.go`: the four-level switch ladder (`applyMultiplicity`,
  `applyDelivery`, `applyTiming`, `applyPayload`, `applyPayloadContent`)
  becomes one `transforms` name→function table; `Apply` admits only catalog
  names so every name has an entry. Application order and RNG consumption are
  untouched (the layer's `order` slice, not the map, orders application).
- `Active` unexported (`applied`; never returned or used outside); a stale
  "Post-pass" comment removed from `Process`.
- Left: unused `atNS`/receiver parameters (the table needs one uniform
  signature), the whole-second reorder rule (D-26).
- Weak pin, known: with the golden stream `producer_flap` delivers nothing
  (every record is held and the flap window never closes), so its digest is
  that of an empty stream; its buffering is covered by `perturb_test`.

## R7 — truth

- `truth.go` split by responsibility: `record.go` (label building) and
  `store.go` (sealed store). `BuildRecord` takes an `Injection` value instead
  of ten positional parameters (four adjacent same-typed); label bytes are
  unchanged (suite golden, analytic cross-check and pin identical).
- `SetupCall` moved to `model` (plain record used by `truth`, `audit`,
  `suite`); `audit` no longer imports `truth`, so `audit` drops to layer 5 and
  `suite` to 6 in the layer table and the allowlist lost the edge.
- `NewStore(runIsOpen)` takes the open-run check at construction and
  `OpenChecker` is gone. Production wiring is identical (`mcp` passed
  `d.runIsOpen`). A nil check counts every run as open, so a store built
  without one refuses to reveal unless unblinded: D-01's fail-open default is
  closed. Behaviour change confined to constructions that never occur in
  production; pinned by `TestStoreWithoutOpenRunCheckFailsClosed`.
- `Solve`/`Result` unexported (only `BuildRecord` and tests used them).
- Not changed: D-11 epoch-zero sentinel, D-12 zero sigma, D-13 peer sigma.
- One error-text edit: the two `Solve: %w` wrap prefixes in
  `truth/solver_scan.go` became `solve: %w` (staticcheck ST1005 stopped
  exempting the capitalised name once `Solve` was unexported). The prefix
  appears only inside `streamsim: …` wraps of oracle-world failures.

## Direction change and M0 — module shape and test organisation

Owner direction (mid-program): every module follows the sibling's pattern — a
facade in each, private internal layers, no leaks — and tests are organised
per layer. Phase B/C of PLAN replace the old R8–R15; STANDARD is v2. The
"pure cores stay one package" decision of v1 is withdrawn; foundations stay
single packages (the sibling's own N/A precedent), flagged for owner
confirmation.

M0 added to `test/architecture` (all green; module gates are vacuous until the
first module migrates in M1, where their injection proofs are recorded):

- Layer model for nested packages: `splitModule`, `layerRank`
  (domain < edges < app < facade), `infoOf`; cross-module layering by module
  layer, intra-module by rank; allowlist and stale-edge gates keyed by module.
- `moduleShapes`, `TestModuleShapeMatchesItsKind`,
  `TestEveryModuleHasUbiquitousLanguage`, `TestFacadesOnlyDelegate`,
  `TestFacadeSignaturesNameNoInternalTypes` (the no-leak gate; aliases are the
  sanctioned exposure).
- Test bar: `TestNoTestsAtRepositoryRoot`, `TestTestsNeverSleep` (ratchet: one
  sleep in `device/uds_test.go`), `TestTestNamesCarryNoPlanningVocabulary`.
- `.agents/prompts/module-refactor.md`, AGENTS.md modularity section.
- Debt rounds for `run`/`device` I/O renumbered to M12/M11.

## M1 — perturb (template module)

- Shape: facade `internal/perturb` (`doc.go`, `api.go`, `service.go`,
  `operations.go`) over `internal/perturb/internal/domain`. Every production
  and test file moved with `git mv`; only the package clause, the package
  comment and the import graph changed in them. The facade keeps the exported
  API the other modules use (`New`, `Layer.{Apply,Clear,Process,Flush}`,
  `Delivered`, the 19 name constants, `Names`); `ActiveIDs` stays internal
  (no caller outside the module).
- `Layer` is a facade struct holding the domain `*Layer` (explicit delegating
  methods, not an alias of the aggregate); `Delivered` is an alias because it
  is a plain value record. Constants and `Names` re-export the domain values.
- Tests: all 285-line behaviour tests, the golden digests and the
  conformance matrix now sit in the domain layer where the private helpers
  live; the facade keeps four black-box contract tests (`perturb_test`
  package): catalog listing, admission of names/params, apply→process→clear,
  flap hold and flush.
- `UBIQUITOUS_LANGUAGE.md` added; architecture map updated.
- Gate corrections found while migrating: the no-leak gate must ignore
  unexported struct fields (the facade holds the internal object in one); a
  nil `Recv` made the inspector panic.
- Injection proofs: a two-statement facade method fails
  `TestFacadesOnlyDelegate`; an exported method returning `*rules.Layer`
  fails `TestFacadeSignaturesNameNoInternalTypes` (and delegation); removing
  `UBIQUITOUS_LANGUAGE.md` fails `TestEveryModuleHasUbiquitousLanguage`; an
  undeclared `internal/zzextra` layer fails `TestModuleShapeMatchesItsKind`.
- Lint: `wrapcheck` ignores errors from a module's private layers
  (`*/internal/*/internal/*`, the sibling's rule), so the facade returns layer
  errors unchanged; layers wrap with their own context.

### M1 review follow-up (template hardened before M2+ copy it)

Independent review found no behaviour change but a weak template:

- **Gates (H1).** Every bypass the reviewer built passed the first gates. The
  facade rules are now `facade_rules_test.go`/`facade_declarations_test.go`:
  all functions (not only exported), `init` forbidden, callee must be
  `<layer>.F` or `<recv>.<field>.M` with the field declared in a struct as a
  layer type (resolved across the module's facade files), no function literal
  or nested call in arguments, constructors are nil-guards plus one return,
  exported non-error variables rejected, aliases only exported, in `api.go`
  and of layer types without methods, dot/blank layer imports rejected.
  `TestFacadeRulesRejectEveryKnownBypass` builds each bypass as a snippet.
- **Mutable catalog (H2).** `perturb.Names` shared the admission slice; it is
  now `perturb.Names()` returning a copy (`domain.CatalogNames`), the domain
  variable is private, and the domain layer tests uniqueness and
  copy-on-return. Callers in `suite` updated.
- **Fail closed (M1).** `perturb.New` returns `(*Layer, error)` and refuses a
  nil spec with `ErrNoSpec`. Deliberate change: callers `run.newRunState` and
  `audit.newAuditPipeline` already returned errors; the nil case is unreachable
  in production (`world.New` precedes it) so no observable difference.
- **Surface (M2).** The 19 name constants had no outside user and are
  removed from the facade; `Names()` is the catalog.
- **Tests (M3, T1).** The catalog-uniqueness test moved to the domain layer;
  the facade test asserts the copy semantics.
- **Docs (L1–L4).** Language file corrected (`delivery_reason`,
  `applied_perturbations`); alias name `layer` mandated; `wrapcheck` ignore
  narrowed to `internal/*/internal/{domain,app}` (edge layers still wrap);
  `ActiveIDs` unexported in the domain layer; DEFERRED path updated.
- Two production functions that crossed 15 lines with the new error path
  (`run.New`, `audit.prepareWorld`) were split into named steps.

## M2 — truth

- Facade `internal/truth` (`doc.go`, `api.go`, `service.go`, `operations.go`)
  over `internal/truth/internal/domain`; files `git mv`-ed, package clause and
  package comment changed, domain tests' fixture path deepened by two levels.
- Facade surface = what other modules used: `Solver`, `NewSolver`, `Store`,
  `NewStore`, `BuildRecord`, `Injection` (alias of a plain record). `Solver`
  and `Store` are facade structs; `Store` has explicit delegating methods,
  `Solver` is an opaque handle passed to `BuildRecord`.
- Deliberate, listed in PLAN: `NewSolver` returns `(*Solver, error)` and
  refuses a nil spec with `ErrNoSpec` (STANDARD fail-closed constructors); the
  one production caller (`suite.prepareGeneration`) already returned errors.
- Facade contract tests (`truth_test`): nil spec, positive/negative class
  label, seal/reveal lifecycle. `UBIQUITOUS_LANGUAGE.md` added.

## M3 — world

- Facade `internal/world` (`doc.go`, `api.go`, `service.go`, `operations.go`)
  over `internal/world/internal/domain` (every file and test `git mv`-ed;
  package clause and comment only). `World` is a facade struct with `ID`,
  `Spec`, `StartNS` fixed at construction and 20 explicit delegating methods;
  `Options`, `InvokeResult`, `EffectorCall`, `FaultInfo`, `Entity` are aliases
  of plain records (the gate checks they have no exported methods); the seven
  `Mode*` constants stay as the vocabulary of `InvokeResult.Mode`;
  `ErrInterlockRefused` is the same error value, so `errors.Is` is unchanged.
- Deliberate, listed in PLAN: `world.New` refuses a nil spec with `ErrNoSpec`.
- The repository-wide map-iteration determinism test moved out of `world`
  into `test/architecture` (T1: a whole-repository rule). It now ranges over
  every production package (layers included) with the same name heuristic
  and marker window (a marker on the loop line or the line above or below). `device`
  and `deviceworld`, which it did not cover before, have 15 map ranges
  (DEFERRED D-38, message order only) and sit in `determinismDebt` until
  their modules migrate (M7, M11).
- Facade gate refinements found here: `constructs` accepts definitions from
  one layer call and nil guards in either direction (`== nil`/`!= nil`) that
  return; an alias counts a layer type's *exported* methods only.
- Facade contract tests (`world_test`, black-box over the shipped
  aquaculture-pond domain): nil spec, identity and entities, ordered emission
  and clock refusal, faults and effectors, entity lifecycle, failure-mode
  override. Facade coverage 96 %, domain layer 68 % (floor moved to the
  layer package).

### M2 review follow-up (applied in the M3 round)

Isolated-snapshot review: no High. Fixed: nil-handle panics (`BuildRecord`
with nil spec or solver, zero `Store`) now return `ErrNoSpec`/`ErrNoSolver`/
`ErrNoStore`; the record clones the caller's perturbation slice; the facade
gate learned guarded delegations (nil guards that only refuse) and now judges
constructor guard bodies (one return of nil/zero/`Err…`/`err`) and every
returned value or composite-literal field (plain references, layer calls with
plain arguments only), proven by new bypass snippets (repairing guard,
computed arguments); facade tests restructured so the facade proves its own
contracts (missing inputs, unknown fault, unblind stamping, double seal,
unknown run, fail-closed stores) instead of repeating domain rules; language
file and DEFERRED paths corrected; `SealStatus` documented (`sealed` is always
true for a known run — D-note: the flag is vestigial). `NewStore(nil)` keeps
the refusing-default substitution (STANDARD §1 exception, documented). Coverage
gaps in the domain layer (divergence/conservation detector paths) remain for
T5.

## M4 — score

- `score` no longer imports `run`: the scorer's input is a `score.Evidence`
  (run id, domain, verdict, ledger, effector calls, applied perturbations,
  emitted count, hidden-state history, reproducible/unblinded flags). The one
  production caller (`mcp.scoreClosedRun`) packs it with `scoringEvidence`;
  the hidden-state record became `model.StateSnapshot` (JSON tags unchanged,
  `run.History()` now returns it). Every metric function reads the same data
  through the same accessors as before, once instead of per call.
- Structure: facade `internal/score` (aliases of the five plain records and
  `Evidence`; `Score` and `Offline` delegate) over
  `internal/score/internal/domain`; files and tests `git mv`-ed (fixture paths
  deepened). Dead `Scorecard.Marshal` (no caller) removed so the record can be
  aliased. The `internal/score` layer drops from 6 to 5 and the allowlist
  edge `score -> run` is gone.
- `Offline` stays a separate entry point sharing the same policy functions
  (D-04: offline cannot compute loop metrics without history/calls); unifying
  the paths changes scorecards and is deferred.
- Facade contract tests: no-verdict refusal, identity/flag carry-over, bundle
  parity online vs offline. Domain-layer scoring tests unchanged (85 %).

### M3 review follow-up

Isolated review: no High. Fixed: the invariants page pointed at the moved
oracle test and its evidence command ran no tests; `TestDocumentationLinksResolve`
now fails on any dead relative link in `documentation/`, `.agents/` and the
root pages (it also caught and fixed three older broken links in
`limitations.md`); the determinism marker window is restored to the original
line above/on/below; the constructor and delegation gate no longer accepts
computed return values or arguments (arithmetic, index, slice, dereference),
guard conditions with calls, or more than one layer definition (five new bypass
snippets); `PendingKicks() < 0` replaced by an assertion that can fail, and
`Options` pass-through (initial entities, emit disabled, forced failure mode)
is tested at the facade.

Recorded, not fixed: the map-iteration gate is a name heuristic. A type-aware
scan finds 46 unmarked map ranges it cannot see (DEFERRED D-43), among them
three whose order reaches output or error text. The `Entity` alias exposes the
live record (parity with before; D-44). `EffectorCall.Args` is the caller's
map (D-45).

## M5 — audit

- Facade `internal/audit` over `internal/audit/internal/domain` (files and
  tests `git mv`-ed; fixture paths deepened). `Panel` is a facade struct with
  one delegating `Audit`; `Verdict` and `Perturbation` are aliases of plain
  records. `DetectorNames` and `BalancedAccuracyCutoff` had no caller outside
  the module and are not re-exported (the exported mutable `DetectorNames`
  slice is gone from the public surface).
- Deliberate, listed in PLAN: `audit.NewPanel` returns `(*Panel, error)` and
  refuses a nil spec (`ErrNoSpec`). Callers `suite.generationTools` and
  `mcp.auditPanel` already return errors; both were split into named steps to
  stay within 15 lines.
- Facade contract tests (nil spec; verdict shape and best-score consistency;
  unknown fault refused). Domain layer 90 %.

## M6 — suite

- Facade `internal/suite` (aliases of `Scenario`, `TrivialCase`, `Suite`,
  `Config`; `Generate` delegates) over `internal/suite/internal/domain`.
- Duplicates removed first, behaviour-neutral and proven by the suite tests,
  the generation goldens and the behaviour pin: `suite.Perturbation` is now an
  alias of `audit.Perturbation` (identical record; the audit receives the
  scenario's slice unchanged instead of a field-by-field copy), and the local
  `renderID` with its hand-written `replaceAll`/`indexOf`/`entityParameter`
  is replaced by `world.RenderID` (same output for nil parameters, the only
  case reached).
- `Generate` is unchanged; facade tests cover the unknown-profile error,
  determinism of the suite bytes and label/scenario parity. Domain layer 70.8 %.
- Doc link repaired (`domains-and-adapters.md` now points at `admission.go`,
  where the profile-name branches live; see the follow-up).

### M4 review follow-up

Differential test by the reviewer: byte-identical scorecards over 4 failure
modes × 2 faults plus an empty-ledger run. Fixed: `mcp.scoringEvidence` is now
pinned field by field against the run accessors (perturbed run, blind and
unblinded; a swapped or dropped field fails) because the behaviour pin never
scores; `Score` returns `ErrNoLabel` for a nil label and `faultFor` tolerates a
nil domain (deliberate, unreachable from `mcp`); the facade's constant-equals-
itself bundle test became a real online/offline agreement test over the shared
consumer and judgment metrics; the architecture map no longer says `score ->
run`. Not done: avoiding the evidence clone before the no-verdict check
(performance only).

## M7 — deviceworld

- Facade `internal/deviceworld` (`Binding` alias; `LoadBindings`,
  `ValidateBindings` delegate; `Plant` struct with `Apply`/`SafeStop`) over
  `internal/deviceworld/internal/domain`. The binding fixture stays at the
  module level (`internal/deviceworld/testdata/`), because the `cli` device
  test reads it too; the layer's tests read it by relative path.
- Deliberate, listed in PLAN: `deviceworld.New` returns `(*Plant, error)` and
  refuses a nil world (`ErrNoWorld`); the one caller (`cli.deviceWorldPlant`)
  already returned an error.
- `determinismDebt` key follows the layer (`.../internal/domain`).
- Facade tests: nil world, unknown catalog field, binding validation against a
  world (missing required target named), unmapped target is
  `ErrPlantUnavailable`. Facade 86 %, layer 83 %.

## M8 — sink

- Facade `internal/sink` (`Sink` alias of the contract; `Inproc`, `File`,
  `HTTPPush` facade structs with explicit delegating methods; `NewFile`,
  `NewHTTPPush`) over `internal/sink/internal/domain` (contract and the
  in-memory sink), `internal/sink/internal/files` and
  `internal/sink/internal/httppush` (the only code with `os` and `net/http`).
  `ioEdges` entries moved to the two edge files; the zero-value `Inproc`
  remains usable (`&sink.Inproc{}` in `run`).
- Verbatim moves; one equivalent simplification in `files.New`: the
  unreachable inner branch of the mkdir check was dropped (same control flow:
  an error with a directory other than `.` fails, otherwise create proceeds).
- Tests by layer: domain (`Inproc`), files (round trip, mkdir failure, flush,
  closed file), httppush (equivalence, unreachable endpoint — moved), facade
  (inproc/file equality, flush visibility, zero value, HTTP delivery order).
- Known and untouched: D-28 (synchronous POST under the run lock),
  `File.Close` re-reads the file, non-idempotent Close, file permissions
  (`0o666 & umask` vs `0o600`).
- Lint: the wrapcheck ignore is `internal/*/internal/*` again (all private
  layers, including edges). It governs calls *into* a layer from the facade;
  an edge still has to wrap the `os`/`net/http` errors it receives because
  those packages are not ignored.

### M5/M6 review follow-up

Isolated review: no High. Suite bytes were compared old vs new binary over
8 domains × 3 profiles (24 files identical) and `renderID` vs `world.RenderID`
over 4488 cases (0 diffs). Fixed: `mcp.AuditScenario` had no test at all (the
M5 change is now covered: result keys, unknown domain, unknown fault);
`suite.Generate` refuses a missing domain (`ErrNoDomain`) and `audit.Panel`
refuses a nil or zero panel (`ErrNoPanel`), both listed in PLAN; the false
"aquaculture setup exception" sentence in `domains-and-adapters.md` is
replaced by the real remaining non-data gap (profile-name branches in
`admission.go`); facade tests assert messages and a non-empty perturbation
list, the suite determinism test uses one scenario (suite package 3.8 s →
2 s), and the scenario perturbation JSON shape is pinned.

## M9 — domain

- Facade `internal/domain` (`doc.go`, `api.go`, `service.go`, `operations.go`)
  over `internal/domain/internal/domain` (parse, validate, compile, digest,
  catalog; the module named `domain` therefore has the one stuttering path
  `internal/domain/internal/domain`) and `internal/domain/internal/files`
  (`Load`, `LoadAll`: the only `os` user). Callers keep `*domain.Compiled`
  unchanged: ~65 references, 171 `.Spec` reads.
- `Compiled` is an **alias** of the layer type although it has methods. It is
  a value type whose whole method set is read-only lookup over its own data
  and is the public contract used by every module, so it is listed in the
  gate's reviewed `aliasedValueTypes` table with that reason; any other alias
  of a method-bearing type is still rejected. `Catalog` is a facade struct
  with delegating `List`/`Describe`/`Coverage`; `Entry` and `CoverageReport`
  are plain-record aliases.
- Tests by layer: rules tests stay in the domain layer (75 %, with a test-only
  `Load` helper), loader tests moved to `files` (100 %), and the facade holds
  the shipped-domain integrity test plus catalog and parse contract tests
  (100 %). Fuzz targets in the Makefile and two documented test commands
  point at the new paths.

## M10 — adapter

- Facade `internal/adapter` (`Load`, `LoadBytes`, `FixtureEvents`, `Engine`
  with `Begin`/`RenderRecord`/`RenderStreamRecord`/`End`/`RenderRun`,
  `NewEngine`) over `internal/adapter/internal/domain` (validation, engine;
  embedded fixture moved with it) and `internal/adapter/internal/files`
  (`Load`: the only `os` user). The loader's `separator` parameter no longer
  leaks: `DecodeFile` (multi-line error layout) and `LoadBytes` (single-line)
  are the two entry points; error text is unchanged.
- Deliberate, listed in PLAN: `NewEngine` refuses a nil adapter
  (`ErrNoAdapter`). `Engine.Meta` (no caller) removed.
- `adapter/conformance` is still a flat edge-core package that uses only the
  adapter facade; it is listed with the other unmigrated packages for the
  final shape gate (a facade/domain split would only forward ~200 lines).
- Tests by layer: rules (66 %, floor moved to the layer), files (100 %),
  facade (86 %: nil adapter, `LoadBytes`, streaming session equals whole-run
  render over the shipped adapter and fixture).

## M11 — device

- Facade `internal/device` (`Device`, `New`, `ApplyCommand`, `State`,
  `SetFaults`, `LoadCapabilities`, `ParseFaultSpec`, `ValidateFaultSchedule`,
  `Listen`; aliases `Config`, `Faults`, `Outcome`, `FaultInjection`, `Plant`,
  `SafeStopper`, `PlantCommand`, `PlantEffect`, `Capabilities`, `WireFaults`;
  the three sentinel errors) over
  - `internal/device/internal/domain`: state machine, admission, capability
    catalog, fault schedule, plant port and the wire record codec (the codec
    is pure and the rules call it, so it cannot sit in a higher layer);
  - `internal/device/internal/uds`: socket listener, session loop, exchange
    and the deterministic wire-fault gate (the only `net`/`os` user);
  - `internal/device/contract`: the vendored schemas as an embedded
    `contract.Schemas` (a tiny foundation package so the module keeps the
    fixture directory it shares with `deviceworld` tests and the Agentic Stream
    provenance in `SOURCE.md`).
- `Capabilities` is an alias of a method-bearing type: listed in the reviewed
  `aliasedValueTypes` table (read-only lookups plus the digest).
- Test seams (`EncodeRecord`, `DecodeRecord`, `ServeConn*`, `QueryStateControl`,
  `AcceptedCommandCount`, `Fault*`) exist only in `export_test.go`; they had no
  production caller outside the module. `domain.ApplyFrame` is the exported
  form of the old private `handleCommand` the edge needs.
- Tests by layer: domain 85 % (including the admission-order test, moved from
  the transport file because it exercises `admit`), uds 76 % (session, wire
  faults, ack loss, injected disconnect, listen) plus the wire-gate unit
  tests, facade 100 % (black-box transport flows over the shipped catalog and
  contract fixtures, plus the facade contract tests). `determinismDebt` key
  follows the layer.
- Untouched, recorded: D-05/D-06 (safe-stop id window, one-shot flags in dedup
  replay), `Listen`'s unjoined accept loop (D-38 family), typed
  receipt/result records (P-04).

## M12a — run: facade over the app layer

- Facade `internal/run` (`Run` with `ID`, `Config`, `World` fixed at
  construction and 25 explicit delegating methods; `Config`, `ReplayResult`
  aliases; `LoadArtifact`, `ReplayArtifact`, `ErrConsumerNotQuiesced`) over
  `internal/run/internal/app`: every file and test moved verbatim (package
  clause, comment, fixture depth only). The four hooks other modules' tests
  use (`SetEvidenceRecorder`, `SetQuiesceParkedHook`, `SetFailureMode`,
  `Trace`) stay public and documented as test-harness hooks: an
  `export_test.go` seam cannot serve another package's tests.
- Removed from the public surface (no caller outside the module):
  `RecordHistory`, `RenderRecord`, `Unblinded`, `TraceDigest`, `AddEntity`,
  `ConfigureEnvTarget`, the quiescence clock types and
  `DefaultQuiescenceTimeout`, and the exported `Perturb`/`Engine`/`Sink`
  fields. They remain in the app layer for the module's own tests.
- The app layer still holds the ledger/artifact/timer I/O; `ioEdges` entries
  are re-keyed with debt `M12b`, which extracts the edges and the pure rules.
  `run` joins `moduleShapes` in M12b (the shape gate requires a domain layer).
- Makefile fuzz, soak and benchmark targets and the invariants evidence
  commands point at the new paths; `go test ./internal/run/...`.
- Facade tests (97 %): identity, command recording, publication and replay of
  an artifact, hooks. App layer 72 %.

### M7–M10 review follow-up

Isolated review: no High; suite/pin/fuzz identical. Fixed: the facade
`deviceworld.Plant` is now pinned as `device.Plant` and `device.SafeStopper`
and a lease expiry is driven through it (a drifting `SafeStop` would have
skipped safe stops silently, since the device discovers it by assertion);
the alias gate no longer accepts composite aliases or an exemption keyed by
name alone: `aliasedValueTypes` pins the target and the exact exported method
set and has a stale-entry test, which immediately caught the unused
`Capabilities.TargetNames` (removed); stale regeneration and test commands
point at the layer packages and the false wall-mode claims are corrected
(`httppush`, `wall`, DECISIONS erratum); zero-value facade types refuse
instead of panicking; a formally racy sink test is locked; the adapter layer
tests the two error layouts; PLAN and STANDARD text corrected (the `app`
layer exists only where use cases orchestrate, and the exemption is pinned,
not "read-only whole method set").

## M12b — run: edges and rules

- `internal/run/internal/durable`: the append-only ledger file (`OpenLedger`,
  `Append`, `Flush`, `Finish`, `Closed`), evidence publication
  (`PublishEvidence`, `WriteRunArtifact`) and `ReadArtifact`; `quiesce`: the
  `Clock`/`Timer` interfaces, `RealClock` and `DefaultTimeout` (`app` keeps
  type aliases so its tests and `Config` are unchanged); `clock`: `Stamp()`,
  the one place `time.Now` is read for the artifact's `created_at` and the
  unblinding stamp (behaviour-neutral; not the `wall` seam, P-01).
- `internal/run/internal/domain`: the pure free functions — `AdapterDigest`,
  `CanonicalHash`/`fnv` (the RunID derivation; now pinned by a test),
  `ValidateArtifactDocument`/`DecodeArtifact`, `CommandString`/`CommandTime`/
  `AsMap`, `ValidateSubmittedVerdict`, `CloneVerdict`, `CountLedger`,
  `ErrorString`. Bodies verbatim; the functions that read `*Run` stay in
  `app` (they are orchestration over the aggregate).
- `app` no longer imports `os`/`bufio`/`net`; the debt entries are gone and
  `ioEdges` lists the four edge files. Error text is unchanged, including the
  `End:`/`run:`/`streamsim:` prefixes of the ledger, evidence and artifact
  paths and the order: trace, ledger, history, verdict, `run.json`.
- `run` joins `moduleShapes`; the facade rules, no-leak and alias gates now
  cover it.
- Tests by layer: `durable` (ledger lifecycle, unusable path, evidence files,
  artifact round trip), `domain` (RunID hash pin, digests, decoding,
  argument decoding, helpers), `quiesce`, `clock`; the app's finalization
  regression now asks the ledger edge whether it is closed.
- Recorded, not changed: the `Run` aggregate in `app` still mixes command
  admission, delivery and replay (D-18 delivery-path divergence stays);
  D-15/D-16/D-17 untouched.

## M13 — refconsumer

- Facade `internal/refconsumer` (`Runner` with `Process`, `New`,
  `DefaultConfig`; `MCPOperator` with explicit methods, `NewMCPOperator`;
  aliases of `Config`, `Nameplate`, `EntityInfo`, `ChannelInfo`,
  `EffectorInfo` and the three port interfaces) over
  `internal/refconsumer/internal/domain` (detector, series, statistics,
  verdict, the Runner and its ports) and
  `internal/refconsumer/internal/mcpclient` (the MCP operator client; the
  only package that speaks to an operator endpoint).
- Moves verbatim; the edge qualifies the domain types it returns.
- Tests: domain 93 % (moved), `mcpclient` 77 % (new: an in-memory fake
  operator serving the four tools — token and run id on every call, decoded
  nameplate and effector list, refusal text), facade 100 % (default tuning,
  empty-trace verdict reaches the sink once, a real streamable HTTP operator
  endpoint, argument refusal). The package had 68 % as one unit; its floor
  entry is gone.
- Recorded, not changed: the consumer identity literal is still repeated
  (DEFERRED duplicates); `NewMCPOperator` uses `context.Background()` for
  its calls (no caller cancellation).

## M14 — mcp

- Facade `internal/mcp` (`Director`, `NewDirector`, `NewDirectorServer`,
  `NewOperatorServerResolver`, `SetOperatorEndpoint`) over
  `internal/mcp/internal/app` (every handler, schema and tool: the
  director and operator use cases, moved verbatim) and
  `internal/mcp/internal/capability` (the entropy edge: the only place
  that reads `crypto/rand`, minting `t-` + 32 random bytes, error text
  unchanged).
- Kind: surface (facade · app · edge); no domain layer because the module
  holds no rules of its own. `ioEdges` names the capability edge; the
  protocol-edge split of `app` stays open (D-46).
- Tests: facade 100 % (roles expose disjoint tool surfaces, the recorded
  operator endpoint reaches `sim.world.create`), `capability` (prefix,
  width, uniqueness), `app` 65.5 % (moved, floor re-keyed to the app layer).
- Documentation links and `go test` commands now name the app layer.

## Review follow-ups — M11 device, M12 run

Isolated review of `42d0dbf`, `92d803b`, `3b40255`: no behaviour regression
(behaviour pin identical, race/shuffle clean, fuzz pass); 0 high, 1 medium,
6 low.

- Fixed: the durable edge's error branches now have tests (`Finish` closes
  the file when the flush fails, `Flush` and `PublishEvidence` error text);
  the exported run rules have doc comments and `CanonicalHash`'s comment
  names the right symbol; `ApplyFrame` is documented; dead production
  `Device.HandleCommand` is gone (its test bridge encodes the outcome
  itself); the unreachable `ServeConnWithFaults` test seam on the facade is
  removed; `SOURCE.md` names the moved conformance test; the verdict test
  now asserts refusal instead of logging.
- Recorded, not changed: D-47 (device `Advance`/`SetFaultSchedule` are
  unreachable by production code but give the default manual clock its only
  way to move), D-48 (`run.Config.QuiescenceClock` is public through the
  alias yet typed by a private interface; `Ledger.Closed` is a test-only
  query on a production type).

## M15 — cli

- Facade `internal/cli` (`Main(args, Build)`, `Build`) over
  `internal/cli/internal/app` (every command: option parsing and use-case
  orchestration, no direct process access) and three edges:
  `process` (standard streams and wall clock as `process.Env`), `files`
  (reads, writes, directory listing) and `serve` (the device socket and the
  MCP director session with its operator endpoint, both until interrupted).
- Commands now return the JSON document to print; the dispatcher prints it
  once. Flag errors surface as a `usageExit` (status 2, `-h` status 0) instead
  of `flag.ExitOnError` exiting inside the command, which is what makes every
  command runnable in-process under test.
- `cmd/streamsim` passes `cli.Build{Version, Commit}` to `cli.Main`; the
  facade exports no mutable variables, so `cli.Version`/`cli.Commit` are gone.
  This is the only change to `cmd`, deliberate and listed below.
- Deliberate changes: (1) the interrupt handler is registered before the
  device socket is bound, so a signal during startup is no longer lost;
  (2) `score` reads the ledger with one file read instead of an open and a
  stream, which changes the text only for a ledger path that is a directory.
- Tests: `app` 81 % (was 26 % for the package as one unit; 22 in-process
  command cases: usage and exit statuses, every error prefix, catalog,
  domain, adapter, run → replay/verify, refconsumer → score, suite, manifest
  with the injected clock and build), `files`/`process` 100 %, `serve` 79 %
  (injected stop channel and in-memory transport; no signals, no sleeps),
  facade 100 %. The cross-process determinism test stays beside the facade.
- Architecture gates: the "surfaces are imported only by surfaces" rule now
  reads the module's kind, so a surface's own layers may import surfaces; the
  twelve per-file `internal/cli` ioEdges entries became three edge entries.
