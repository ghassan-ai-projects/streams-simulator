# Handoff — Canvas modular architecture migration

- **Date:** 2026-10-17 (stop point) · **Branch:** `refactor-ui-code`
- **Plan:** [PLAN.md](PLAN.md) · **Bar:** [BAR.md](BAR.md) · **Architecture:** [README.md](README.md) · **Module table:** [canvas/ARCHITECTURE.md](../../../canvas/ARCHITECTURE.md)
- **Read first:** `AGENTS.md` (Canvas section, the MUST rules), this file, then `PLAN.md` §4 (test ladder) and §5.

## 1. Position at the stop point

| Phase | State |
| --- | --- |
| Plan + bar | **Committed**: `c9fb428`, `539fddd` |
| **P0** register, four checkers, ratchet, drivers, lifecycle guards | **Committed**: `542b95c`, `6583f06`; `make phase-gate` passed; three review rounds, findings fixed |
| **P1a** presentation inversion | **Committed**: `962fc43` |
| **P1b/P1c** graph extraction + graph CSS ownership | **Committed**: `417d745` (graph module, facades, vendor sheet, local tier classes, target-edge rule, checker corrections with fixtures). One open item: the vendor-selector repair needs a positive **browser** rerun (`reviews/P1.md` finding 1) |
| **P2** clock ports, route sessions, operation identity | **Committed**: `be47ec7`. **No independent review and no test/browser evidence yet** (`evidence/P2.md`, `reviews/P2.md` say so) |
| **P3** typed gateways, decoders in the adapter, feature failure unions | **Committed**: `56e80fb` (production) · **uncommitted working tree**: the whole test/double migration (`tsc` clean, **104 files / 486 tests green**) |
| **P4** editor save transaction | **Committed**: `d6a08ed` + `src/app/editorSave.ts` (uncommitted). `Draft.savedRevision` is written but `isDirty` is identity-based again (see §4) |
| **P5** feature module boundaries + CSS ownership | **Partially committed**: graph CSS partition (`1420eca`), `model/workspace`, `app/session`, `ui/primitives` facades, `ARCHITECTURE.md`. **Remaining: the 134-file feature move, the primitives, the stylesheet dissolution** |
| **P6** test audit | Not started |

**Everything is uncommitted in the working tree except the six commits above.** At the stop point the tree is coherent:
`npx tsc --build --noEmit` clean, `npx vitest run` 104/104 files and 486/486 tests green, `npm run architecture`
**ok** (fixtures 45, architecture debt 2, css debt 36, ratchet ok, architecture doc ok, callers ok).

## 2. Immediate next steps

1. **Commit the P3 test migration** (the current working tree) once the remaining gates pass:
   `npm run lint && npm run format:check && npm run knip && npm run dup && npm run build && npm run css:emitted && npx playwright test`
   (none of these has been run since the migration; expect small fallout).
2. **Close the P3 coverage gap** (the subagent flagged it explicitly): the success-path decoder assertions that moved out
   of the model tests exist nowhere yet. Write `canvas/src/adapters/bridge/internal/decoders.test.ts` and
   `components.test.ts` covering: `workspace()` (`duration_ns`→`durationS`, `recorded_at`→`recordedAt`, malformed→`null`),
   `fileContent()`, `written()`, `paired()`, `composeSummary()`, `acknowledgement()` (`effect_eta_ns`→`effectEtaNs`),
   `runDomain()` (`domain_spec` extraction), `library()` (sorting/filtering/ref/words), `presets()`, `examples()`,
   `component()` (port direction `provides`→`out`, `state`→`channel`, category from class then tags, title/description
   fallback).
3. **Run the P5 move** — the tooling is built and validated (see §3): `node scripts/draft-facades.ts /tmp/p5-moves.json --write`
   → move the 134 files in `/tmp/p5-moves.json` (regenerate it first, see §3) → `node scripts/codemod-imports.ts <plan> --apply`
   → register transform → `npx tsc` → `npm run architecture`.
4. **P5 primitives and CSS dissolution**: `ui/primitives/{panel,button,disclosure,metadata,scroll-region,breadcrumb}` with
   `internal/`, then delete `ui/shared.module.css` and `ui/panel.module.css` and migrate their ~25 callers (usage map in
   §5). Register caps fall as the files shrink.
5. **P2/P1 validation that was deferred**: an isolated browser run of the vendor-selector fix, the launch hold measured
   before/after the scheduler migration, and the P2 evidence/review files filled with real output.
6. **P6**: the test-retirement ledger (outcome, owner, unique failure, replacement) and the `FakeBridge` policy cleanup.
7. **Final round**: `make canvas-check`, `make phase-gate` (with `PATH`, `GOTOOLCHAIN=go1.26.9`, `GOCACHE` per §4), Enola
   (three inferred cycles are an open item), two independent reviews per phase lens, Chrome capture review, and the
   evidence/review files updated with real results.

## 3. Tooling built for the remaining phases (in `canvas/scripts/`)

| Tool | Purpose |
| --- | --- |
| `codemod-imports.ts` | AST import/export rewriter for moves: rewrites relative specifiers through a move map, routes imports from outside a module to that module's entrypoint, dry-run by default (`--apply` writes). Validated end to end |
| `draft-facades.ts` | Computes from the import graph exactly which names each module's callers need and writes `contract.ts` (public types) + `index.ts` (values + contract) per module |
| `sync-callers.ts` | Derives the register's `callers`; `--check` fails on drift (wired into `npm run architecture`) |
| `write-architecture-doc.ts` | Generates `canvas/ARCHITECTURE.md` from the register; `--check` is wired into `npm run architecture` |
| `capture-screens.ts` | Captures 20 Chrome screens against a real fixture for visual review |

The register currently has folder-level targets and per-module `folders`/`entrypoint`, so the P5 move only needs the map
regenerated from the live tree (`node -e` snippet in this session's transcript) plus a register transform that re-paths
`debt`, `css.caps`, `css.globalCaps`, `css.colorExceptions`, `features` and `targets` through the map (drafted at
`/tmp/p5-register.mjs`; re-check it before use).

## 4. Known open items (do not lose these)

- **P1 review finding 1**: `vendor.module.css` selectors are module-scoped and only work because `vendor.dot` is applied
  to handles in `nodes.tsx` and `vendor.minimap` to the `MiniMap`; a positive browser rerun is outstanding.
- **P3 coverage gap** (§2 item 2) — assertions currently exist nowhere.
- **P4 dirty tracking**: `Draft.savedRevision` is written but `isDirty` is `doc !== saved` again, so undo/redo back to the
  saved document reads clean (the pre-refactor rule). Revision-based dirty tracking needs the history to carry revisions
  (`past`/`future` as `{doc, revision}`) — a P4-sized change with tests.
- **`componentSpecs.ts` fallback category** was restored from `"prod"` (a P3 rewrite) to `"env"` (pre-P3 value and the
  decoder's unknown-class default) to keep an existing assertion; confirm that is the intended default.
- **Enola**: three inferred cycles are an open item, not a clean result.
- **Register**: `callers` are derived; run `node scripts/sync-callers.ts` after any move, and
  `node scripts/write-architecture-doc.ts` after any register change.

## 5. Facts that save time

- **Environment**: Go targets need `PATH="$PATH:$HOME/go/bin"`, `GOCACHE=/tmp/ss-gocache`, and
  `GOTOOLCHAIN=go1.26.9` for the behaviour pin (host Go 1.27.2 is compiler-sensitively different).
- **e2e** runs ~70 tests in ~2 min on the real binary + Chrome; it is load-sensitive (the wall-clock launch hold, and
  occasionally `alarm-playback`/`keyboard`). Re-run a failure in isolation before calling it a regression. Zoom
  animations drop keypresses: use `stepZoom`/`zoomInTo` from `e2e/support/viewport.ts`, never a bare `press("+")` in a poll.
- **Patch with asserted replacements**: `str.replace` that silently misses has broken three checker edits in this
  session. The architecture checkers are the fastest reviewer — run `npm run architecture` after every register or import
  change.
- **Do not add a re-export "for completeness"**: `knip` fails on an unused export; export exactly what callers name.
- **A file's target assignment is live data**, enforced by the `target-edge` rule for every extracted module.
- **shared/panel usage map for the CSS dissolution** (callers, `composes` included):
  `shared`: `.crumb` 1 (+Breadcrumb), `.sectionTitle` 4, `.cut` 7, `.scrolls` via composes (LaunchPanel, RunScreen,
  PlantScreen), `.panels` via composes (PlantScreen, RunScreen), `.button` via composes (HomeScreen, PlantScreen,
  RunScreen), `.metaPart` via MetaParts. `panel`: `.panel` 3, `.head` 3, `.body` 4 (+composes in 11 sheets), `.group` 10,
  `.field` 5, `.pair` 3, `.check` 0, `.hint` 8, `.problems` 0, `.primary` 2, `.secondary` 8, `.icon` 4.
