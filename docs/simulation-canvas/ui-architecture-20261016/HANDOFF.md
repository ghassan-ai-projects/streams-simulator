# Handoff — Canvas modular architecture migration

- **Date:** 2026-10-17 · **Branch:** `refactor-ui-code`
- **State:** the migration (P0–P5) and the whole-UI review are delivered; the authoritative status, the deviations from the plan and the open items are in [PLAN.md §0](PLAN.md). The finding dispositions are in [reviews/P5.md](reviews/P5.md); the module table is [canvas/ARCHITECTURE.md](../../../canvas/ARCHITECTURE.md), generated from `canvas/architecture.json`.

## Next steps

None are blocking. See PLAN.md §0 "Open items": merge the branch, re-pin the Enola baseline, optionally turn the class-token component tests into e2e style checks.

## Open items after the documentation audit (2026-10-17)

Optional hygiene, none blocking:

- The canvas ratchet (`npm run architecture`) finds the commit that introduced `architecture.json` in git history, so it needs full history: the CI checkout must use `fetch-depth: 0` (set in `.github/workflows/ci.yml`; the ratchet reports a shallow clone as a failure instead of passing silently).
- The documents under `docs/simulation-canvas/ui-architecture-20261016/` other than PLAN.md §0 and this file describe the plan as accepted on 2026-10-16; their pre-migration names (`ui/home`, `ui/run`, `ui/pairing`, `PlantScreen.module.css`) are not rewritten. The current names are in `canvas/ARCHITECTURE.md`.
- Counts quoted in `evidence/` are as of each phase; the current suite is 146 Vitest files, 721 tests and 70 Playwright tests.

## Working notes

- Gates: `cd canvas && npm run check` (format, types, lint, layers, architecture register and ratchet, duplication, dead code, unit tests with coverage, build, emitted-stylesheet oracle, e2e); the repository gate is `PATH="$PATH:$HOME/go/bin" GOTOOLCHAIN=go1.26.9 GOCACHE=/tmp/ss-gocache make phase-gate`.
- The e2e suite shares one workspace across parallel workers and is load-sensitive: `review.spec.ts` "fits at 40 %" and `areas.spec.ts` header zoom each failed once under a busy machine and passed alone; re-run a failure alone before treating it as a regression. (The keyboard and live-run flakes were real defects and are fixed.)
- `node scripts/sync-callers.ts` and `node scripts/write-architecture-doc.ts` regenerate the derived parts of the register and `ARCHITECTURE.md`; run both after any module or import change, then `npx prettier --write architecture.json ARCHITECTURE.md`.
- Do not add a re-export "for completeness": knip fails on an unused export.
