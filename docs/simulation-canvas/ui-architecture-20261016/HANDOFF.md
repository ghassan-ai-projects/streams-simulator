# Handoff — Canvas modular architecture migration

- **Date:** 2026-10-16 (late) · **Branch:** `refactor-ui-code`
- **Plan:** [PLAN.md](PLAN.md) · **Bar:** [BAR.md](BAR.md) · **Architecture:** [README.md](README.md)
- **Read first:** `AGENTS.md` (Canvas section, the MUST rules), this file, then `PLAN.md` §4 (test ladder) and §5.

## 1. Position

| Phase | State |
| --- | --- |
| Plan + bar | **Committed**: `c9fb428` (plan + plan reviews), `539fddd` (bar). Owner approved the plan in conversation; `create_goal` was called with the seven-phase objective |
| **P0** — register, four checkers, ratchet, fixtures, test drivers, lifecycle guards | **Committed** `542b95c` (enforcement), `6583f06` (guards + drivers + evidence + review dispositions). `make phase-gate` **passed**. Three review rounds; every finding fixed |
| **P1a** — presentation inversion (prerequisite) | **Committed** `962fc43` |
| **P1b/P1c** — graph module extraction + CSS ownership | **UNCOMMITTED, 91 changed paths.** Almost done; two small defects and the closure steps below |
| **P2–P6** | Not started |

Green at the last full run *before* the current batch: `npm run check` = 70/70 e2e, 466 unit tests, all four checkers.
After the current batch (vendor sheet + local tier classes): `tsc` **has 2 errors** (below), and the previous full check
has not been re-run.

## 2. Immediate next steps (in this order)

1. **Fix the two defects introduced by the interrupted batch** (both in `canvas/e2e/plant.spec.ts`):
   - line 4–5: two imports from `./support/viewport`; merge into one (`import { stepZoom, zoomInTo, zoomOutTo } from "./support/viewport";`).
   - line 182: `tiers.add(await graph.getAttribute("data-zoom-tier"))` returns `string | null`; use `tiers.add((await graph.getAttribute("data-zoom-tier")) ?? "")`.
   Then `npx prettier --write e2e/plant.spec.ts && npx tsc --build --noEmit && npm run lint`.
2. **Re-run the CSS-sensitive e2e specs** (`plant`, `plant-labels`, `zoom`, `review`, `wires`) after `npm run build`.
   The batch renamed the zoom-tier classes from global `tier-full|compact|overview|map` to local CSS-Module classes and
   moved the tier onto a `data-zoom-tier` attribute on the graph root; e2e reads now use that attribute. Any other
   `tier-` reference left in `e2e/**` or `src/**` must be updated (grep for `tier-`).
3. **Full gate**: `npm run check` (must be 70/70 + 466 tests + four checkers), then
   `PATH="$PATH:$HOME/go/bin" GOTOOLCHAIN=go1.26.9 GOCACHE=/tmp/ss-gocache make phase-gate`.
4. **Close P1**: write `evidence/P1.md` (a draft exists — update it: the vendor sheet is now done, the tier classes are
   local, the façade was trimmed, the target-edge rule was added, the capture tool's marker changed) and `reviews/P1.md`
   with the three reviews' findings and dispositions (they are in this session's transcript; the boundary lens initially
   **rejected** P1 and its findings were then fixed: see §3).
5. **Commit P1** with the owner-approval lines (see §4). Then start P2.

## 3. What the current (uncommitted) P1 work contains

- **Move**: 50 files `git mv`'d — 19 UI files into `canvas/src/ui/graph/internal/`, 31 model files into
  `canvas/src/model/graph/internal/` — plus `ui/graph/index.ts`, `model/graph/index.ts`,
  `model/graph/contract.ts`, `ui/graph/internal/graph.module.css`, `ui/graph/internal/vendor.module.css`.
- **Facades**: `ui/graph` exports `Navigator`, `Outline`, `PlantGraph`, `useCanvasAspect`, `ZoomBar`, `NodeLive`;
  `model/graph` exports exactly what outside callers name (trimmed after `knip` flagged 11 unused re-exports) plus the
  port/plant vocabulary its own UI needs (`Plant`, `areaOf`, `Family`, `Pin`, `familyNames`, `shapeOf`, `kindLabel`).
- **Target-level dependency contract enforced**: new rule `target-edge` in `canvas/scripts/architecture/tsRules.ts` —
  a file whose *target module* is a registered module must obey that module's `targetModules[].mayImport` contract.
  It is scoped to extracted modules, skips test files, and exempts `test` support, the composition root and
  `test-support` modules. It immediately caught 8 real edges, now fixed: the graph UI takes plant/port vocabulary
  through the `model/graph` façade; the graph owns its minimal `ReadingValue` in `model/graph/contract.ts`;
  `src/model/areaTree.ts` was re-targeted from `model/editor` to `model/shared` (graph and editor both use it).
- **CSS ownership**: graph-owned rules moved verbatim to `graph.module.css`; the app's own 63 `:global(.tier-*)`
  selectors became local classes (`tierMap|tierOverview|tierCompact|tierFull`) mapped from the closed `Tier` union in
  `PlantGraph.tsx`, with `data-zoom-tier` on the graph root for tests; the `:global([role="application"]:focus)` rules
  became local attribute selectors; the React Flow connection hooks (`.connectable`, `.connectingfrom`, `.valid`) and
  the minimap panel override moved into the new **registered** `vendor.module.css`; the cross-file `composes: scrolls`
  was dissolved by adding a `--scroll-shadow-bg` token to `tokens.css` that both sheets now reference (this also
  removed a `jscpd` clone). `src/ui/graph` is registered in the register's `features` map.
- **Register**: the two graph modules registered (kind, layer, lifetime, folders, entrypoint, callers, `mayImport`);
  targets re-pathed; caps lowered (`PlantScreen.module.css` 453 → 314); `nodes.module.css` `globalCaps` now 0/30;
  `GraphControls.module.css` allowance removed; `vendorFiles` gained the vendor sheet; `emittedCss` sentinel for the
  graph is now `tierCompact` and the counts are the observed 32 `!important` / 17 `z-index` / 32 positioned with a
  `note` naming the correction; debt re-pathed/re-counted (one removed, two re-pathed, one counted 4 → 3, one added for
  the outline's remaining shared-stylesheet import, P5).
- **Checker corrections (owner-approved in conversation)**: the ratchet treats a module nested under the importer's own
  folders (or under an already-allowed module) as a refinement, and matches caps/allowances/colour exceptions/debt that
  only changed path with an equal-or-lower count; the façade rule allows `export { X } from "./internal/X"` value
  re-exports while still rejecting internal types, namespaces and `export *`. Fixtures: 42 cases
  (`facade-value-reexport`, `facade-internal-type-reexport`, `dynamic-import`, `import-alias`, `import-equals`,
  `target-missing`, `callers-recorded`, `protocol-template-tail`, …).
- **Test robustness fix**: `e2e/support/viewport.ts` gained `stepZoom`/`zoomInTo`/`settledTier` because the old test
  pressed `+` inside `expect.poll` while the zoom animation was still running, so presses were dropped and `tier-full`
  (zoom ≥ 1 from a 0.48 opening) was never reached. **Proven pre-existing**: the same spec fails at committed HEAD in a
  clean worktree (`/tmp/wt-p1`); it is a timing fragility, not a Phase 1 regression. Other `+` loops in
  `plant.spec.ts`, `plant-labels.spec.ts`, `review.spec.ts`, `wires.spec.ts` now use `stepZoom`.

## 4. Owner approvals already given (record these in the commit)

- Two Phase 0 checker corrections (ratchet refinement + moved-entry matching; façade value re-export) — approved when
  the extraction was blocked; the commit must carry
  `Owner-approved-gate-change: ratchet refinement and moved-entry matching, plus the sanctioned facade value re-export (owner approved in conversation)`.
- The **vendor stylesheet** registration (`css.vendorFiles` grows) is the F4-required end state and the owner said "do
  the right thing, no workaround" when asked to defer it; name it in the same approval line or a second one.
- The ratchet currently fails with exactly one line — `vendor stylesheet allow-list gained src/ui/graph/internal/vendor.module.css` —
  which those approval lines cover.

## 5. Environment facts the next agent needs

- Go targets need `PATH="$PATH:$HOME/go/bin"` (for `deadcode`/`govulncheck`) and `GOCACHE=/tmp/ss-gocache`;
  the behaviour pin needs `GOTOOLCHAIN=go1.26.9` (host Go is 1.27.2 and the pin is compiler-sensitive).
- `make canvas-check` = clean `npm ci` + `npm audit` + `npm run check`; `make phase-gate` adds Go CI, Enola and the pin.
- The e2e suite runs ~70 tests in ~2 min on the real binary and Chrome; it is load-sensitive: the wall-clock
  `launch.spec.ts` hold test and occasionally `alarm-playback`/`keyboard` flake under parallel load. **Re-run a failure in
  isolation before treating it as a regression** (BAR §4).
- `node scripts/capture-screens.ts <dir>` (canvas/) captures 20 screens against a fresh fixture. Known limits: the theme
  is a persisted preference so `*-1920-light` shots come out dark, and run timestamps differ per capture.
- `.enola/` holds the Enola baseline (ignored); `make architecture` reports new coupling but exits 0.
- The canvas checkers are TypeScript run by Node's native type stripping (`canvas/scripts/**`, typechecked by
  `tsconfig.scripts.json`); `npm run architecture` runs all four and the fixtures.

## 6. Remaining phases (plan §5)

- **P2** operation identity: `app/shared/operationIdentity.ts`, route sessions (`PlantModule`/`EditorModule`/`ReviewModule`
  with Strict-Mode-safe `open`/`close`), the unsaved-navigation guard moved to the editor route with a disposer,
  `Clock`/`Scheduler` ports replacing the launch/review timers, **and the launch test fake-timer migration carried from
  P0** (`evidence/P0.md` §7).
- **P3** typed bridge boundary: per-feature gateways in `app/ports`, decoding moved into `adapters/bridge`, closed
  feature failure unions, `BridgeFailure`/action names/HTTP codes out of app and UI, `src/adapters/**` coverage floor.
- **P4** transactional editor save: revision identity, immutable save attempts, conflict/partial-success states,
  composition byte-compatibility assertions.
- **P5** feature module boundaries and CSS ownership: `home`→`workspace`, `run`→`review`, `pairing`→`session`; facades +
  `internal/` per feature; `ui/primitives`; dissolve `shared.module.css` and `panel.module.css`; split the oversized
  sheets; `canvas/ARCHITECTURE.md`.
- **P6** test audit with the retirement ledger.

## 7. Lessons that save time (learned the hard way here)

- Patch files with the `edit` tool or with **asserted** replacements: `str.replace` that silently does not match has
  already caused three broken checker edits in this session. After every patch run `tsc` + the affected checker.
- The architecture checkers are the fastest reviewer: run `npm run architecture` after every register or import change;
  they name the exact file, rule and line.
- Do not add a re-export to a façade "for completeness": `knip` fails on an unused export. Export exactly what callers
  name.
- A file's target assignment is what enforces the dependency contract; `targets` is live data now, not documentation.
- When a phase's exit criterion depends on a gate that cannot express it, fix the gate with a fixture that proves the
  loosening case still fails, and ask the owner before changing any gate (BAR §4).
