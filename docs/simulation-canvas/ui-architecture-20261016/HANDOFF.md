# Handoff — Canvas modular architecture migration

- **Date:** 2026-10-17 · **Branch:** `refactor-ui-code`
- **State:** the migration (P0–P5) and the whole-UI review are delivered; the authoritative status, the deviations from the plan and the open items are in [PLAN.md §0](PLAN.md). The finding dispositions are in [reviews/P5.md](reviews/P5.md); the module table is [canvas/ARCHITECTURE.md](../../../canvas/ARCHITECTURE.md), generated from `canvas/architecture.json`.

## Next steps (in order of value)

1. Move the rules still held in the UI into model/app (review warnings, launch draft converter, `useLive` status, `NewPlant` readiness, editor status text) and the cross-store coordinators into `app/shell`/`app/review`.
2. Application-layer simplifications: one operation token across stores, editor save-state union, `LaunchStore` run cursor, `PlantMaker` result union, a transport-level `FakeBridge`.
3. A Chrome design pass for the accent-button unification and the `PlantScreen.module.css` split.
4. P6: the test ledger.

## Working notes

- Gates: `cd canvas && npm run check` (format, types, lint, layers, architecture, duplication, dead code, unit tests with coverage, build, emitted-stylesheet oracle, e2e); the repository gate is `PATH="$PATH:$HOME/go/bin" GOTOOLCHAIN=go1.26.9 GOCACHE=/tmp/ss-gocache make phase-gate`.
- The e2e suite is load-sensitive: `review.spec.ts` "fits at 40 %" and `launch.spec.ts` "holds the mark for 3 s" occasionally fail under a busy machine and pass in isolation; re-run a failure alone before treating it as a regression.
- `node scripts/sync-callers.ts` and `node scripts/write-architecture-doc.ts` regenerate the derived parts of the register and `ARCHITECTURE.md`; run both after any module or import change, then `npx prettier --write architecture.json ARCHITECTURE.md`.
- Do not add a re-export "for completeness": knip fails on an unused export.
