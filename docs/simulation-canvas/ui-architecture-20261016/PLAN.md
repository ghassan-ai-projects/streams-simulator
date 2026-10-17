# Canvas modular architecture — implementation plan

- **Date:** 2026-10-16
- **Implements:** [README.md](README.md) — the accepted Canvas modular architecture of 2026-10-16
- **Quality bar:** [BAR.md](BAR.md) (job and per-phase bar; written after this plan is approved)
- **Evidence and reviews:** `docs/simulation-canvas/ui-architecture-20261016/evidence/P<n>.md` and
  `docs/simulation-canvas/ui-architecture-20261016/reviews/P<n>.md`
- **Scope:** `canvas/` only. No Go source, contract, fixture, pin, `test/architecture` file or root `scripts/` file changes.
- **Hard rule:** no behaviour change outside the corrections this plan enumerates, each written as a failing-first
  regression test and listed in its phase's evidence. They are: the Phase 0 lifecycle guards (§5 P0.8), the Phase 3
  malformed-required-response correction (§5 P3), and the Phase 4 editor save-transaction rules (§5 P4). Every other
  phase is behaviour-preserving, and its evidence must show the pre-change suite passing unchanged.

## 0. Execution status — 2026-10-17 (resumed)

The user resumed the migration with a standing delegation: deviate from this plan wherever a better design is clear, and record why. Order followed: finish production migration → repair tests and run the full chain → resolve Enola cycles → browser verification → whole-UI quality review.

### Delivered

| Phase | State |
| --- | --- |
| P0–P3 | Committed earlier; P3 test/double migration committed in `bc831583` (typed gateways, 486 tests). |
| P4 | Editor save transaction committed; dirty tracking stays identity-based (documents are immutable) and the revision counters were removed. |
| P5 | Delivered: 203 files moved into `index.ts` + `internal/` modules (28 product modules), facades and contracts, feature session factories, CSS ownership, register rebuilt from the real tree. |

### Deviations from the plan (each deliberate)

| Plan text | What was done instead | Why |
| --- | --- | --- |
| `app/shell` owns `createPlantModule`/`createEditorModule`/`createReviewModule` and the session interfaces | Each factory and its `*Session` interface live in the owning feature (`app/plant`, `app/editor`, `app/review`); `app/shell` keeps navigation, panels, theme | The session type belongs to the feature; shell-owned session types forced every UI feature to import `app/shell` and broke the dependency direction |
| Six component primitives (`button`, `panel`, `disclosure`, `metadata`, `scroll-region`, `breadcrumb`) | `ui/primitives` exports `Breadcrumb`, hooks, `cx` and three stylesheet objects (`panel`, `shared`, `screen`); `ui/theme` is its own module for the three global entrypoints | Only the breadcrumb had a shared component shape; the other callers share class names, so a component per class would only wrap `className` |
| Cross-module `composes` | Dissolved by applying the shared class in TSX; the screen side-hiding rule is native CSS nesting inside the existing `:global` selector | No new `:global` allowance and no cross-feature stylesheet import |
| `adapters/bridge` and `adapters/browser` as two modules | One `adapters` module | The protocol-owner ratchet names the module; splitting it would loosen the rule's meaning without a benefit |
| Register keeps `model`, `app`, `ui`, `adapters` | Kept as folder-only umbrella modules with empty allow-lists | The ratchet compares against the introducing commit; umbrellas make every new module a refinement instead of a regression |
| Ratchet unchanged | One extension with a fixture: a global-stylesheet owner that moved between modules under the same basename is not a new owner | `shared.module.css` moved into `ui/primitives`; the old rule only knew caps and allowances moving, not the owner list |
| `.tsbuild` output inside `canvas/` | tsc output goes to `node_modules/.cache/tsbuild` | Enola scanned the gitignored declaration output and reported its cycles as regressions |
| Fakes keep their own failure mapping | `fakeFailures` calls the adapter's real mappers through the adapter facade | The duplicate could drift from production and tripped the duplication gate |

Defects found and fixed during the move (each with a regression test where it is a behaviour change):

- The P3 report decoder required `run.run_id`; the Go report omits it for an open run, so a still-open run reviewed as "No report". `reportDecoder.test.ts` fails without the fix; the superseded assertion (`{ run: {}, ... }` refused) now asserts a missing `run` section instead.
- `wirePins.test.ts` iterated a tuple whose mode expression was always truthy, so `pins` was never exercised; the intended mode matrix is restored.

### Whole-UI review (after the migration)

Three independent read-only reviewers (model + adapters, application + test support, UI + stylesheets) produced 85 verified findings; two implementation waves applied the low-risk majority. Dispositions are in [reviews/P5.md](reviews/P5.md). Highlights:

- Lifecycle defects fixed failing-first: discard during an in-flight save, workspace watcher subscribed after the first load, plant load dropped for a newer workspace, double publish on navigation.
- A new gate rejects references to undefined stylesheet classes and defined-but-unused classes (`check-css-architecture`, two fixtures). It immediately found four tier classes that were applied but never defined.
- Dead code and duplication removed: one JSON-narrowing helper, one plural/underscore helper, shared failure mappers, shared `groupBy`/`tally`, dead review-row fields, 20 dead or empty CSS rules, duplicate CSS blocks, optional props production never omits.
- P4 dirty tracking is closed by decision: `Composition` is immutable, so `doc !== saved` is the correct dirty test; the revision counters were removed.

### Open items (recorded, not blocking)

- Rules still held in the UI that belong in model/app: review warnings and alarm copy (`RunStatus`, `alarmWords`), the launch draft converter (`draft.ts`), `useLive` status derivation, `NewPlant` readiness, editor status text; cross-store coordinators (`useRunPlant`, `showBuild`, rerun proposal) belong in `app/shell`/`app/review`.
- State-model simplifications in the application layer: one operation token type for all stores, `LaunchStore` run cursor, `PlantMaker` result union, the editor save-state union (`saveStatus`/`composeFailure`/`outcome`/`failure` overlap), replacing `FakeBridge` by a transport-level fake.
- Visual unification that needs a Chrome design pass: six accent-button definitions into `shared`, `PlantScreen.module.css` split by owner.
- P6 test ledger.
- Enola: `make architecture` passes (no cycle or layer regression).

## 1. Why this plan exists

The accepted architecture keeps the four technical layers and the existing `Observable` + `useSyncExternalStore`
mechanism, and adds: product modules with facades over private `internal/` trees, a shared `graph` module, typed
simulator gateways, explicit operation identity, a transactional editor save, Clock/Scheduler ports, CSS ownership per
module, and enforced architecture checks. The following evidence describes the **pre-migration baseline captured on
2026-10-16**, not the current worktree; see §0 for the live checkpoint and resume order. At that baseline, there was
one flat layer per technical layer, one raw `Bridge.action` surface, no operation identity, and `ui/editor`/`ui/run`
imported `ui/plant` files and `PlantScreen.module.css` directly.

## 2. Baseline facts (measured, not asserted)

| Fact | Value |
| --- | --- |
| Canvas source | 185 modules, 24 007 lines under `canvas/src` |
| Tests | 96 files / 445 unit and component tests, all green; `npm run test` exit 0 |
| Coverage | `model` 98.82 %, `app` 97.47 % (floors 90 / 80); `adapters` has no floor |
| Static gates | `format:check`, `typecheck`, `lint`, `deps` (0 violations, tests excluded by config), `dup` (0 clones), `knip` clean |
| Go gates | `make ci-check` passes with `PATH="$PATH:$HOME/go/bin" GOCACHE=/tmp/ss-gocache`. `deadcode`/`govulncheck` live in `~/go/bin` |
| Behaviour pin | already mismatched at baseline on host Go 1.27.2 (`scripted-day/run.json`, `manifest-file`); CI pins Go 1.26.9, which is installed locally (`GOTOOLCHAIN=go1.26.9 go version` → go1.26.9). Go is untouched by this plan, so the pin is used only as "Go unchanged" evidence |
| End-to-end | 70 real-Chrome tests, ~2 min against the real binary. One wall-clock test (`launch.spec.ts:122`) failed once at 2 399 ms against a 2 500 ms bound under parallel load and passed in isolation; recorded as pre-existing, re-measured per phase |
| Raw bridge surface | `Bridge.action(name, args): Promise<Outcome<unknown>>` with 12 production call sites in 5 app files (`componentSpecs.ts:12`; `editorStore.ts:102,185,206`; `launchStore.ts:217,223`; `plantMaker.ts:51,57,70,93`; `reviewStore.ts:233,280`) plus the port declaration (`ports.ts:19`); `ReviewState.runDomain: Load<unknown>`; `PlantStore.show(key, domain: unknown)` |
| Protocol leakage | 12 behaviour-controlling reads of `BridgeFailure.code`/`.issues`/`.result` outside `adapters/`, plus `pairingBlocked`/`isMissing`/`pairingMessage` in `model/bridge.ts` |
| Malformed-success tolerance | `parseLibrary(null)`, `parseComponent(null)`, `parsePresets(null)`, `parseExamples(null)`, `parseWorkspace(null)`, `parseReport(null)`, `parseWritten(null)`, `worldOf(null)`, `summaryOf(null)` decode a malformed required payload into an empty-but-successful value |
| Operation identity | none — no generation, epoch, session id or operation token in `canvas/src/app` or `canvas/src/ui`. Guards are resource equality only: `worldId`, `PlantStore.wanted` path, `EditorStore.current(path)` |
| Timers in `app` | `ReviewStore` `setTimeout(flush, 250)` + `Date.now`; `LaunchStore` `Date.now` + `window.setTimeout`; both constructor-injected with production defaults |
| Fixed sleeps / uncontrolled flushes | `plantStore.test.ts:114` `setTimeout(resolve, 5)`; 40+ `await Promise.resolve()` flushes; `LaunchPanel`/`LaunchRunning` tests poll a real 2 000 ms hold with a 3 000 ms timeout |
| Vacuous stale-callback test | `reviewStore.test.ts:135` cannot invoke a stale callback: `FakeBridge.followRun` deletes the follower on dispose, so `stream("w-1", …)` dispatches to nothing |
| Policy-bearing fake | `FakeBridge` implements `component.new` plant creation, etag fabrication, read-after-write, 404 defaults and `siteArea` |
| Missing regressions | no A→B→A test for review, workspace, plant or editor; no workspace close-while-reload-pending; no editor save-then-route-switch, edit-during-write or edit-during-compose |
| Strict Mode | `main.tsx:32` is the only `StrictMode` in the tree; no test harness renders one; e2e serves the production build, where React does not double-invoke effects |
| Editor writes | `EditorStore.save()` writes `serialise(doc)`; no test asserts the written bytes |

Cross-module facts that shape the migration:

| Fact | Detail |
| --- | --- |
| Graph consumers | `ui/run/RunScreen.tsx` and `ui/editor/EditorScreen.tsx` import `PlantGraph`, `GraphControls`, `Outline`, `canvasAspect`, `PlantScreen.module.css`; `ui/editor/EditorBar.tsx` imports `GraphControls` and `PlantScreen.module.css`; `ui/run/useLive.ts` imports `NodeLive` from `ui/plant/flow` |
| Graph→review coupling | `model/plantView.ts:9-11` imports `shortChannel` (playback), `reviewRowCount` (reviewRows) and `kindLabel` (report); `model/nodeWords.ts:2-3` imports `NodeStatus`/`statusWords`; `model/planMarks.ts:1,3` imports `model/launch` and `humanise`; `model/readout.ts:2-3` imports `playback`/`reviewRows`; `ui/plant/Readout.tsx:3-4` imports `reviewBody`/`ReviewRow`. The doc registers `review -> shared, workspace, plant, graph`, so these edges must be inverted, not moved |
| `Outline` coupling | `ui/plant/Outline.tsx:3-4,6` imports `model/plant` (`Plant`, `areaOf`) and `model/report` (`kindLabel`), neither of which `ui/graph` may import |
| `PlantScreen.module.css` reach | editor and run use `.screen`, `.side`, `.center`, `.bar`, `.graphHost`, `.failure`, `.loading` (`:1,15,34,42,81,265`) — shell/screen layout, not graph. The graph-owned remainder is `.graph`, `.panCue`, and everything `Outline` uses |
| CSS contracts | `@xyflow/react/dist/base.css` is imported at `ui/plant/PlantGraph.tsx:3` and must precede the graph sheets in the emitted stylesheet; `PlantGraph.tsx:252` sets `tier-${tierOf(zoom)}`, consumed by 63 `:global(.tier-*)` selectors and asserted at 7 e2e sites; `flow.ts:106,144` emits `wire-on`/`wire-off`/`family-*` strings styled in `ui/theme/families.css`; `PlantScreen.module.css:420-421` selects `:global([data-left|data-right="closed"])` (shell state) and `:2,24,30` use cross-file `composes` from `shared.module.css`. `nodes.module.css:117-120` uses `!important` against a React Flow handle rule; `PlantScreen.module.css:88-94` sets the pan-cue `z-index` |
| Emitted stylesheet | `canvas/dist/assets/index-*.css` today: 34 `!important`, 17 `z-index`, 33 positioned declarations, with vendor rules first, graph sheets second, tokens/base/families last — a deterministic file that decides cascade ties |
| Audit vacuity risk | `e2e/support/idFit.ts:23` and `noteFit.ts:5,16` select by class substring (`[class*='id_']`) and assert only `toEqual([])`, so a class rename disables them silently |
| No hidden state | no module-level mutable state anywhere in the graph or feature scope; the move risk is CSS cascade and the global class contract |

Environment facts: `make build` fails under the harness file sandbox unless `GOCACHE` is inside a writable path; all Go
targets run as `PATH="$PATH:$HOME/go/bin" GOCACHE=/tmp/ss-gocache …`, and the pin check adds
`GOTOOLCHAIN=go1.26.9`.

## 3. Decisions

| # | Decision | Reason |
| --- | --- | --- |
| D1 | Implement the doc's phases 0–6 in order, one phase per round, each independently reviewable and committed. A handoff checkpoint is reported after Phase 3 and after each Phase 5 integration unit | The doc's order separates behaviour correction (P0, P4) from structural movement (P1, P5). P0–P4 are bounded; P5 is five features plus shell and primitives, so it is executed as units and can be stopped safely between them |
| D2 | No existing gate is loosened: no threshold falls, no allow-list grows, no exclusion is added, no rule is removed, and every existing check keeps running. New checks, new files, and the tooling configuration needed to lint and typecheck those new files are additions. `test/architecture` and root `scripts/` are not edited | `AGENTS.md` §MUST-3. `make gate-ratchet` covers only Go, so a canvas ratchet makes this enforced rather than review-only |
| D3 | `canvas/architecture.json` is the single machine-readable register: module kinds, entrypoints, lifetimes, callers, exact per-module import allow-lists (no globs), the current→target mapping for every directory, CSS caps and exception registers, coverage floors, and the exact migration debt with the phase that removes it | F3/F4. The Go gates use a package table the same way |
| D4 | `canvas/scripts/check-architecture.ts`, `check-css-architecture.ts` and `check-ratchet.ts` are TypeScript, parsed with the TypeScript compiler API (already a devDependency: `typescript` 6.0.3), and run with Node's native type stripping from `npm run architecture`, inside `npm run check` | A regex gate cannot decide "workflow in an entrypoint" or "internal type in a public signature"; README:1153 asks for the compiler API. TypeScript also keeps the checkers typechecked and lintable without a new dependency |
| D5 | The register and its ratchet live in `canvas/` and never in `test/architecture`; the ratchet compares the working tree against the commit that introduced `canvas/architecture.json`, so caps may only fall, floors may only rise, ignores/exclusions may only shrink, and debt entries may only disappear | Keeps `make gate-ratchet` the untouched Go ratchet while giving canvas the same property. Comparing against the introducing commit (found with `git log --diff-filter=A`) cannot be reset by a later commit |
| D6 | Operation identity is a tiny `RequestBook`/`SessionSlot` in `app/shared` that manages identity only | A3's explicit limit: no task framework, cache or cancellation library |
| D7 | Typed gateways follow the doc's capability grouping (session, workspace, component/plant, editor, launch, review), implemented by one bridge adapter over a private generic HTTP helper. Feature failure unions are closed, feature-owned, and never a renamed protocol union | A2's limit plus F2's intent: the UI must not branch on protocol codes |
| D8 | CSS stays CSS Modules over the existing semantic tokens; primitives are created only where two real callers share semantics, each as `primitives/<name>/index.ts` over `internal/` | F4 and the doc's "no second styling framework" decision |
| D9 | Visual evidence is: the full real-Chrome e2e suite, the emitted-stylesheet oracle (D11), audit considered-element floors, and a Chrome review of captured screens recorded in `reviews/P<n>.md`. A capture helper (`canvas/e2e/tools/capture.ts`, not a collected spec) writes PNGs to a directory outside the repository | The quality bar permits pixel baselines only in the pinned Linux CI container and forbids local baseline comparison; `playwright/no-skipped-test` is an error with `allowConditional` off, so a self-skipping spec is a lint violation. The capture helper gives the reviewer real images without a committed baseline |
| D10 | Reviewers read the diff and the specs, not the author's summary. Two lenses minimum (architecture/boundary; correctness and behaviour preservation); a third UX/accessibility/visual lens for the phases that move CSS or UI structure (1 and 5) | Repository quality bar §6.6 and the canvas bar's UI-phase rule |
| D11 | `canvas/scripts/check-emitted-css.ts` inspects the built `canvas/dist/assets/index-*.css` and compares, against a baseline recorded in the register: the relative order of vendor, graph and theme rules, and the exact multiset of `!important`, `z-index` and positioned declarations. It fails when the build output is missing, runs after `npm run build` in the `check` chain, and every change is named as a deliberate correction in the register | This file decides cascade ties, is byte-stable across machines, needs no browser, and is the only cheap oracle for the cascade/z-order risk the CSS split creates |
| D12 | Delegated work is heavily used, but every structural decision and every diff to a facade boundary, the register, the checkers or a gate file is reviewed by the lead agent | `AGENTS.md` §MUST-5 |
| D13 | The plan stops at a phase or unit boundary and reports if the bar cannot be met. It never lowers a bar to finish | Quality bar §6 |
| D14 | Phase 1 begins with a presentation inversion, not with file moves, and Phase 0 settles the file→module assignment for every existing module so that no phase depends on a deferred ownership decision | The survey found `graph → review` edges in `plantView`, `nodeWords`, `planMarks`, `readout` and `Readout.tsx`, and that `Outline` cannot be neutral otherwise. README:306 makes the inversion a precondition |

Constraints the checkers cannot enforce, assigned to review explicitly: "shared code lists its callers" (README:760),
"cross-module orchestration has one application owner" (:759), React effect resource and cleanup (:756), Context
purpose (:755), simulator-ownership parity (:761), and CSS selector semantics (:1134).

## 4. Test ladder (when to run what)

| Moment | Command | Cost |
| --- | --- | --- |
| Every edit | `npx tsc --build --noEmit` and `npx vitest run <touched test files>` | seconds |
| Every logical step | `npm run lint && npm run deps && npm run architecture` (source checkers) | ~40 s |
| Every phase, before review | `GOCACHE=/tmp/ss-gocache make canvas-check` — the authoritative canvas gate: clean `npm ci`, format, types, lint, layers, architecture, duplication, dead code, unit tests with coverage, build, e2e | ~7 min |
| Every phase that moves CSS or UI | `CANVAS_CAPTURE=<dir> node e2e/tools/capture.ts` before and after, then Chrome inspection of the images, recorded in `reviews/P<n>.md` | ~2 min |
| Every phase that moves files | `make architecture` (Enola) after re-pinning at Phase 0 | ~1 min |
| Every phase exit | `PATH="$PATH:$HOME/go/bin" GOTOOLCHAIN=go1.26.9 GOCACHE=/tmp/ss-gocache make phase-gate` (ci-check + canvas-check + architecture + behaviour-pin-check) and `git diff --check` | ~11 min |
| End of job | the same `make phase-gate`, plus a `test/architecture`/`scripts` diff review | ~11 min |

Heavy gates are never run in parallel: the one wall-clock e2e test is load-sensitive, and a false failure costs more
than the saved time. A timing failure is re-run in isolation before it is treated as a regression, and the result is
recorded.

## 5. Phases

Every phase ends with: gates green, `evidence/P<n>.md` recording real command output (including the observed
failing-first output for every correction), `reviews/P<n>.md` recording each reviewer's findings and dispositions, and a
commit. Structural movement and behaviour correction are separate commits.

### Phase 0 — architecture checks, controlled test drivers, lifecycle safety

No feature or graph file moves.

**P0.1 Register.** `canvas/architecture.json`: every module's kind, entrypoint, lifetime and callers; exact
per-module import allow-lists (never a glob); the current→target mapping for every directory (`home`→`workspace`,
`run`→`review`, `pairing`→`session`, and the rest as they exist today); the file→module assignment for all `model/*`
files, including `model/bridge.ts` exports (protocol documents → `adapters/bridge`; shared values and `Outcome` →
`model/shared` or the owning module; failure unions → feature contracts) and the folding of `model/ports.ts` into
`model/plant`; CSS caps and exception registers; coverage floors; and the exact migration debt, one entry per file or
import edge, each naming the phase that removes it.

**P0.2 `check-architecture.ts`.** Test-inclusive AST scan of `canvas/src` (production and tests). It rejects: an
unregistered module or file; an import that is not on the exact allow-list; any import of another module's private
path, including type-only imports; a missing entrypoint; a wildcard facade export, an exported concrete controller, an
internal type in a public signature, or a statement in `index.ts` other than imports, re-exports and factory
delegation; production `as unknown as` double assertions; simulator action names and `/api/` routes outside
`adapters/bridge`; `Outcome<unknown>` in an app public API; `react`/DOM imports under `src/app`; `src/ui → src/adapters`;
exported boolean-flag parameters; and JSX `style` attributes carrying static colour, spacing, typography or layout
declarations. It cross-checks `.dependency-cruiser.cjs`: every layer rule named in the register still exists with the
same severity and no `from`/`to` exclusion was added.

**P0.3 `check-css-architecture.ts`.** Fails on: a feature importing another feature's stylesheet; cross-module `composes`;
`:global` or `!important` outside the registered vendor file; raw colour literals outside the theme and the registered
mask exception; global custom-property declarations outside the theme contract; any CSS file above its cap, and any
unlisted CSS file above the 300-line bar. Exception and cap registers may only shrink.

**P0.4 `check-ratchet.ts`.** Compares the working tree against the introducing commit of `canvas/architecture.json` and
against the gate configuration: caps may only fall, coverage floors may only rise, `knip` ignores and dependency-cruiser
exclusions may only shrink, the `npm run check` chain may not lose a command, and debt entries may only disappear. A
replacement debt entry is legal only when it is not broader than the entry it replaces and the old entry is deleted in
the same commit.

**P0.5 Emitted-stylesheet oracle.** `canvas/scripts/check-emitted-css.ts` records in the register the baseline order and
declaration multisets described in D11 and checks them against the built stylesheet; it is added to the `npm run check`
chain after `npm run build`, and fails when the build output is absent.

**P0.6 Fixtures.** `canvas/architecture-fixtures/` (outside `src`, excluded from tsc, ESLint, depcruise, knip and jscpd,
scanned only in the checkers' fixture mode): one failing fixture per rule, positive fixtures for same-module private
access and for a legal cross-module entrypoint import, a type-only private-import fixture, and a fixture proving the
debt list is exhaustive (removing one debt entry must make the checker fail). `evidence/P0.md` lists rule → fixture →
observed message.

**P0.7 Tooling accommodation.** `canvas/tsconfig.scripts.json` (referenced from `tsconfig.json`) for the checkers, a
`disableTypeChecked` ESLint block for `scripts/**`, and `npm run architecture` added to `npm run check`. These are
additions; no existing rule, ignore or threshold changes.

**P0.8 Controlled test drivers and lifecycle corrections.** `src/test/deferred.ts`, `src/test/streamDriver.ts`
(retains a disposed follower so a late callback can be invoked deliberately), and `src/test/fakeClock.ts` implementing
the README A5 `Clock`/`Scheduler` signatures. Replace the 40+ microtask-flush loops and the fixed sleep in
`plantStore.test.ts` with deferred drivers, and replace `LaunchPanel`/`LaunchRunning` real-time polling of the 2 000 ms
hold with fake timers. Add an explicit `<StrictMode>` harness in `src/test/`. Then, each observed failing for the
intended reason and corrected in the same unit:

1. review: a retained callback from a superseded run must not add observations to the current playback index (today
   `ReviewStore.take` adds before its `worldId` check) nor refresh the current run;
2. review: A→B→A with the first request still pending must not publish the older answer;
3. plant: two deferred overlapping loads resolve in the required order without a sleep, plus A→B→A;
4. workspace: close while a reload is pending must not publish and must not install the watcher; two overlapping
   reloads publish in request order;
5. editor: A→B→A open ordering and a late load that must not install state.

**P0.9 Audit floors.** Every `e2e/support/*Audit.ts` returns the number of elements it considered, and every spec that
uses one asserts a floor, so a renamed class cannot disable geometry auditing silently.

**Exit:** checkers reject new production, test and CSS boundary violations; the ratchet passes and the debt register is
baselined and may only shrink; every P0.8 test fails before its guard and passes after, with output in the evidence;
the strict-mode harness renders a route under `StrictMode`; all gates green; no feature or graph file moved.

### Phase 1 — extract the shared graph module

Prerequisite commit (behaviour-preserving, before any file moves):

- Invert the `graph → review` edges: `model/plantView` takes neutral presentation inputs (a row-count function, a
  note-height function, a channel namer, a kind namer) with the current defaults; `model/readout` and `model/nodeWords`
  take `NodeStatus`/`Shown` as values; `model/planMarks` takes launch marks and words as neutral inputs;
  `ui/plant/Readout` receives the review body builder. Move caller-counted shared display words (`kindLabel`,
  `humanise`) to `model/shared`.
- Give `Outline` a `model/graph` data source: an outline projection (areas, items, labels, issue marks) computed in
  `model/graph`, so `ui/graph` imports only `model/graph` and `ui/primitives`.

Then:

- Declare `model/graph` and `ui/graph` in the register with their responsibility and callers.
- Move, behaviour-unchanged, into `ui/graph/internal/`: `PlantGraph`, `GraphControls`, `Outline`, `canvasAspect`,
  `flow`, `nodes`, `PanCue`, `WireEdge`, `viewportFit`, `labelFit`, `Readout`, `Marks` and their tests; the public
  `ui/graph/index.ts` exports only what the three consumers name (`PlantGraph` + props, `Outline` + props, `Navigator`,
  `ZoomBar`, `useCanvasAspect`, `NodeLive`).
- Move into `model/graph/internal/`: `plantView`, `layout`, `wireView`, `wireRoute`, `wirePath`, `wireGrid`,
  `wireSearch`, `wireEnds`, `viewport`, `headerFit`, `graphKeys`, `nodeWords`, `frameTitle`, and `readout` where the
  inversion shows it is graph-owned presentation; `model/graph/index.ts` exposes the values plant, editor and review
  name, and nothing more.
- Move graph-owned styles out of `PlantScreen.module.css`/`nodes.module.css` into `graph/internal/graph.module.css`,
  `nodes.module.css` and `vendor.module.css`; keep the emitted order (vendor base, then graph sheets, then the theme
  entrypoint); reduce the React Flow `:global`/`!important` overrides to the selectors React Flow requires; keep the
  `.tier-*` presentation classes graph-owned and registered; dissolve the three cross-file `composes` lines; lower the
  CSS caps and remove the graph debt entries without adding a broader one.
- Update `ui/editor` and `ui/run` to import only `ui/graph` and `model/graph`.

**Exit:** plant, editor and review import only the public graph API; no `graph → review` or `ui/graph → model/plant`
edge remains; the geometry, layout and e2e suites are green; the emitted-stylesheet oracle passes or every change is
named as a deliberate correction; the Chrome review of captured screens is recorded; graph-owned debt is gone and the
residual `PlantScreen.module.css` edge used by editor and run remains as a P5-owned entry with its cap.

### Phase 2 — operation identity, route sessions, Clock and Scheduler

- Add `app/shared/operationIdentity.ts` (identity only).
- Introduce `PlantModule`/`EditorModule`/`ReviewModule` factories with route sessions owning `open`/`close`, reusable
  and Strict Mode-safe; `close` invalidates in-flight work and releases streams, watchers and pending requests.
- Move the unsaved-navigation guard from `main.tsx` to the editor route, registered through a shell capability that
  returns a disposer and removed on cleanup.
- Add the `Clock` and `Scheduler` ports and the browser implementation; migrate the launch and review application
  timers; then activate the no-direct-app-timer rule.
- Replace the P0.8 direct guards with the shared identity primitive where it makes ownership clearer, keeping those
  regression tests green with their own assertions.
- Tests: A→B, A→B→A, close-while-pending, close/open replay, repeated close, Strict Mode open/close counts through the
  `StrictMode` harness, unsaved-guard disposal, disposed-callback rejection, and `Scheduler`/`Clock` schedule, cancel
  and delay semantics against `fakeClock.ts`. No sleeps.

**Exit:** a closed session cannot publish or install any side effect; app code uses no direct clock or timer; the
listed tests pass without timing coordination; `evidence/P2.md` records the measured launch hold from an isolated e2e
run before and after, so that after this phase a lower-bound failure is treated as a regression rather than flakiness.

### Phase 3 — type the bridge boundary

- `app/ports/` grows typed feature gateways (session, workspace, component/plant, editor, launch, review) replacing
  `Bridge.action(string, Record<string, unknown>) -> Outcome<unknown>` at every application call site.
- `adapters/bridge/index.ts` over `internal/` (and `adapters/browser/index.ts` over `internal/`) implement them, keep
  one private generic HTTP helper, and own request encoding, SSE framing, journal decoding and payload decoding:
  `parseWorkspace`, `parseFile`, `parseWritten`, `parseLibrary`, `parsePresets`, `parseExamples`, `parseComponent`,
  `parseReport`, `parseReplay`, `parseJournalRecord`, and the envelope and status mapping.
- A malformed **required** response becomes an explicit unreadable/protocol failure, never an empty success. Each
  decoder change is failing-first with the superseded assertion quoted in `evidence/P3.md`, and the phase states which
  rendered codes and messages change. The `DomainDocument` the review receives stays opaque and unvalidated.
- A verification verdict returned inside an unsuccessful response stays a typed result case.
- Protocol failures become closed feature-owned failure unions with the diagnostic preserved. `BridgeFailure`, HTTP
  status, protocol codes, action names and `/api/` routes disappear from `app` and `ui`; no feature union is a renamed
  protocol union, and no app or UI code compares a protocol code or status. The journal reaches the app as decoded
  entries.
- `vitest.config.ts` gains a `src/adapters/**` coverage floor no lower than the moved files' current coverage, and
  `evidence/P3.md` records per-moved-file coverage before and after.
- The P0.8 regression tests are ported to the typed gateways with unchanged assertions.

**Exit:** no application public API returns `unknown` or exposes `BridgeFailure`; action names, HTTP status and protocol
codes live only in the adapter; the adapter coverage floor holds; mapping, malformed-payload and real-binary parity
tests green.

### Phase 4 — transactional editor saving

- Add draft revision identity and immutable save attempts captured from one revision.
- The write uses the captured path, document and composition etag; compose uses the captured source/output paths and
  output etag; a completion updates transport metadata only while its operation is active.
- Edits made while saving stay dirty; navigating away invalidates the attempt; a conflict keeps the draft and exposes
  recovery; write success followed by compose failure is a defined partial-success state with retry.
- Documented guarantees match the current compose protocol: the composition etag is not re-checked atomically at
  compose time, so the UI does not claim a strict captured-source guarantee. A bridge contract change for that is out
  of scope.
- Tests, failing first: save-then-route-switch, edit-during-write, edit-during-compose, conflict, partial success and
  retry; plus a byte-compatibility assertion that `EditorStore.save()` writes the fixture bytes (and, once in e2e, the
  example composition after a round trip).

**Exit:** the listed tests pass, failure precedence is specified and tested, the written composition bytes are asserted,
and existing compose behaviour is byte-compatible (Playwright editor specs unchanged).

### Phase 5 — feature module boundaries and CSS ownership

One integration unit per feature: `session`, `workspace`, `plant`, `editor`, `launch`, `review`, then `shell` and
`primitives`. Each unit ends with its own gates, evidence, review and commit.

- `home`→`workspace`, `run`→`review`, `pairing`→`session`; each application feature becomes `index.ts` over
  `internal/` (concrete controller, transitions, workflow helpers, caches), each UI feature becomes `index.ts` over
  `internal/` (screens, parts, styles), each with `contract.ts` and public contract tests that import only `index.ts`.
- `app/shell` owns navigation, theme, panel state, session gating, the `ApplicationModules` composition contract and
  the cross-module coordinators (`showBuild`, the run-opens-recorded-plant flow, session-to-workspace opening), so
  `ui/shell` no longer passes a whole `Stores` object to screens. Context stays limited to stable tree-scoped
  presentation facilities.
- Extract only proven primitives (`button`, `panel`, `disclosure`, `metadata`, `scroll-region`, `breadcrumb`) into
  `ui/primitives/<name>/index.ts` over `internal/`, and dissolve `shared.module.css` and `panel.module.css` as their
  last responsibilities move.
- Register `ui/theme/tokens.css`, `base.css` and `families.css` as the only global entrypoints, and move the graph
  selectors that `families.css` currently owns into the graph module.
- Split the oversized feature stylesheets by named owner as the owning module moves; lower each cap as it falls.
- Adopt the shared JSON readers only in the parsers touched by the migration.
- Add `canvas/ARCHITECTURE.md`: the human module table, checked against `canvas/architecture.json` so the two cannot
  drift.
- Activate each migrated module's facade, private-import, CSS-ownership and file-cap rules and remove its debt entries.

**Exit (per unit, and for the phase):** architecture rules prevent private cross-feature imports in production and
tests; the CSS debt register is empty and its exception list is frozen; no circular modules; no gate weakened; the
emitted-stylesheet oracle passes or names each correction; the Chrome review is recorded.

### Phase 6 — test audit

- Build the outcome/owner/unique-failure/replacement ledger for the suite.
- Remove only proven duplicates; keep distinct oracles, browser rendering tests and real-binary tests.
- Simplify fixtures so general fakes carry no server policy; replace `FakeBridge`'s plant-creation policy with canned
  typed outcomes backed by the real-binary tests.
- Measure suite time before and after; never trade failure coverage for a smaller count.

**Exit:** every remaining test states its purpose and its lowest owning layer, the ledger is committed, and the full
canvas and phase gates pass.

## 6. Risks

| Risk | Mitigation |
| --- | --- |
| A CSS split flips cascade order, drops an `!important` guard or loses a `z-index` | D11's emitted-stylesheet oracle, audit considered-element floors, e2e geometry/contrast suites, Chrome review of captured screens |
| A refactor changes behaviour with no test noticing | Corrections are failing-first with recorded output; reviewers check that each new test fails against the pre-change code; `evidence/P<n>.md` quotes the observed failure |
| A gate, cap, floor or exclusion is quietly weakened | D5's canvas ratchet against the introducing commit; the register is the single list; reviewers read the register diff and `package.json`/`eslint.config.js`/`vitest.config.ts`/`knip.json`/`.dependency-cruiser.cjs` diffs |
| The typed-gateway phase touches many test files | Ports are typed and fakes follow the port; the Playwright suite is the unchanged real-wire oracle; the phase is committed per feature; P0.8 tests are ported with unchanged assertions |
| Phase 3 moves ~1 000 decoder lines out from under the model floor | A new `src/adapters/**` floor no lower than today's coverage, plus per-file before/after coverage in the evidence |
| Phase 5 is very large | One unit per feature with its own gates, evidence, review and commit; the register's debt list shrinks monotonically; a checkpoint is reported after each unit |
| Strict Mode behaviour is unverifiable through e2e (production build) | P0.8's `StrictMode` harness with open/close count assertions |
| A gate must change to make a phase pass | Not allowed. The phase stops and reports |
| The wall-clock e2e test flakes under load | Heavy gates are serialized; a failure is re-run in isolation and recorded |
| Sandbox cannot write the Go cache | `GOCACHE=/tmp/ss-gocache` on every Go target |

## 7. Definition of done for the job

1. Every phase's exit criteria met, evidenced in `evidence/P<n>.md`, and reviewed in `reviews/P<n>.md` with no open
   P1/P2 finding.
2. `canvas/architecture.json` describes the delivered tree with an empty debt list; both checkers and the ratchet pass;
   `make canvas-check` and `make phase-gate` pass.
3. No Go file changed, and the behaviour pin is byte-identical to the baseline at the pinned toolchain
   (`GOTOOLCHAIN=go1.26.9`).
4. No behaviour change outside the corrections enumerated in §5 P0.8, P3 and P4, each proven by a regression test and
   recorded in that phase's evidence.
5. Every round committed with scope, gates and review outcome in the message.

## 8. Plan review record

This plan was reviewed by two independent reviewers before implementation:

- **Architecture and sequencing lens** — `approve with changes`; the changes are folded in as D3–D5, D11, P0.1–P0.7,
  P0.9, the P1 prerequisite commit and scoped exit, P3's adapter floor, P5's unit structure, and the restated hard rule.
- **Risk and verifiability lens** — `approve with changes`; the changes are folded in as the canvas ratchet (D5), the
  emitted-stylesheet oracle (D11), the audit floors (P0.9), the Strict Mode harness (P0.8), the malformed-response
  correction named in the hard rule, the composition byte assertion (P4), the `fakeClock` scheduler tests and the
  measured-hold requirement (P2), and one fixture per rule (P0.6).
