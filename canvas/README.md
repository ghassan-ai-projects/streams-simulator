# Streams Simulator canvas

The visual client of the simulator: open a workspace, see a plant, build one from the component library, run it, disturb it
and review what happened. It is one client of the binary and the files, equal to a coding agent: it computes nothing the
binary owns (validation, composition, simulation, truth, scoring, verification) and has no capability the CLI and the
director MCP tools lack. Goal and acceptance: `governance/CANVAS_UI_ACCEPTANCE.md`; gates: `governance/CANVAS_QUALITY_BAR.md` §3.

## Architecture

```
ui (React components)  ──▶  app (state, use cases, ports)  ──▶  model (pure types and rules)
                                   ▲
adapters (bridge client, browser storage) ── implement the app's ports
main.tsx: the composition root wires adapters into the app and renders the ui
```

| Layer          | Holds                                                                  | May import                                                |
| -------------- | ---------------------------------------------------------------------- | --------------------------------------------------------- |
| `src/model`    | Types and pure rules                                                   | nothing outside itself                                    |
| `src/app`      | Stores and use cases; `ports/` declares what it needs from the outside | `model`                                                   |
| `src/adapters` | The bridge client (HTTP JSON and Server-Sent Events), browser storage  | `model`, `app`                                            |
| `src/ui`       | Components, styles, design tokens                                      | `app`, `model`; never `adapters` (no transport in the UI) |

`dependency-cruiser` enforces the table (the `deps` gate). The canvas reaches the simulator only through the local bridge
(`streamsim canvas`), which speaks to the director over MCP and serves workspace files; the Go build never depends on this
folder (`test/architecture/canvas_test.go`, `go.mod` `ignore ./canvas`).

## Modules and the register

Inside each layer the code is split into modules, `src/<layer>/<name>` (for example `src/app/review`, `src/ui/graph`).
A module exposes an `index.ts` and hides its implementation under `internal/`; other modules import the facade only. The
public types of a module are re-exported from its `contract.ts`.

`architecture.json` is the register of those modules and is enforced by `npm run architecture`:

- each module's allow-list of imports and its callers (callers are derived by `scripts/sync-callers.ts`);
- the facade rule: nothing reaches into another module's `internal/`;
- CSS ownership: which module owns which stylesheet, global entrypoints and owners, size caps, `!important` and
  `:global` allowances, colour exceptions, and the class-key rule (a class used in a component but not defined, or defined
  and never used, fails);
- a ratchet against the commit that introduced the register: a cap may fall, an allow-list may shrink, a floor may rise,
  never the reverse. It needs full git history (`fetch-depth: 0` in CI).

`ARCHITECTURE.md` is generated from the register (`scripts/write-architecture-doc.ts`) and checked for drift by the same
command; it lists every module, what it may import and who calls it. After any module or import change run
`node scripts/sync-callers.ts` and `node scripts/write-architecture-doc.ts`, then `npx prettier --write architecture.json ARCHITECTURE.md`.
`npm run css:emitted` checks the built stylesheet (`dist/assets`) against the register: rule order and the counts of
`!important`, `z-index` and positioned declarations.

## Design system

Semantic tokens in `src/ui/theme/tokens.css` (surfaces, text, accent, status ok/warn/alarm/info, truth), light and dark,
a 4 px spacing grid, six type sizes, system UI and monospace faces (tabular numbers for values). Desktop only: the shell
needs at least 1280 px.

## Commands

| Command                               | Does                                                                                                                                                                       |
| ------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `npm ci --ignore-scripts`             | Install the pinned dependencies                                                                                                                                            |
| `npm run dev`                         | Development server on 127.0.0.1:5173                                                                                                                                       |
| `npm run check`                       | Every gate, in order: format, types, lint, layers, architecture register, duplication, dead code, tests with coverage floors, build, emitted stylesheet oracle, end-to-end |
| `npm run architecture`                | The register: module allow-lists and facades, CSS ownership and class keys, the ratchet, `ARCHITECTURE.md` and callers in sync                                             |
| `npm run css:emitted`                 | The built stylesheet against the register (needs `npm run build` first)                                                                                                    |
| `npm run deps` / `dup` / `knip`       | Layer rules / duplication / dead code, one at a time                                                                                                                       |
| `npm run test`                        | Vitest with coverage; floors per layer in `vitest.config.ts` (`model` 90, `app` 80, `adapters` 99 lines)                                                                   |
| `make canvas-check` (repository root) | The same, from a clean install, plus `npm audit`                                                                                                                           |

End-to-end tests run against the real `streamsim` binary and real files; locally they use the installed Chrome, in CI the
pinned Chromium.
