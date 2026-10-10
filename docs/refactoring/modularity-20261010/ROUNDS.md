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
  every production package (layers included) with the same marker. `device`
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
