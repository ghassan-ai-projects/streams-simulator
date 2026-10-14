# P8a evidence — canvas foundation

| Exit criterion | Result |
|---|---|
| `canvas/` toolchain with every QUALITY_BAR §3 gate on the shell | `make canvas-check` from a clean install: format, `tsc --build` (strict, `noUncheckedIndexedAccess`, `exactOptionalPropertyTypes`, `noImplicitOverride`), type-aware ESLint with size and complexity limits and comment rules (0 warnings), dependency-cruiser layers, jscpd (0 clones), knip (0 unused), Vitest with per-layer coverage thresholds (10 tests, axe in component tests), Vite build (72 kB gzip), Playwright in Chrome with axe in both themes; `npm audit --audit-level=high` clean; 17 s |
| Canvas CI job | `.github/workflows/ci.yml` job `canvas` (Node 24, pinned Chromium) |
| Go isolation gate | `TestGoNeverReachesIntoTheCanvas`; `go.mod` `ignore ./canvas` keeps `canvas/node_modules` out of `./...`; the repository walkers skip `node_modules` |
| Customer-workspace marker, ignore rules and check (REVIEW I3) | `STREAMSIM_CUSTOMER_WORKSPACE` marker; `TestNoCustomerWorkspaceInsideTheRepository`; `/workspaces/` ignored; both gates shown to fire on a planted violation |
| `AGENTS.md` canvas section, `canvas/README.md` architecture, `canvas/DEPENDENCIES.md` | Written |
| End goal, acceptance tests and review checklist | `governance/CANVAS_UI_ACCEPTANCE.md` |

Design review (Chrome, 1440 × 900, light and dark): stable shell (top bar, workspace navigation, main), empty state that
says what to do, AA contrast, visible focus ring, no clipping. Monospace values render without ligatures.
