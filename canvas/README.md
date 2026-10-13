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

| Layer          | Holds                                                                    | May import                                                |
| -------------- | ------------------------------------------------------------------------ | --------------------------------------------------------- |
| `src/model`    | Types and pure rules; generated contract types under `src/generated`     | nothing outside itself                                    |
| `src/app`      | Stores and use cases; `ports.ts` declares what it needs from the outside | `model`                                                   |
| `src/adapters` | The bridge client (HTTP JSON and Server-Sent Events), browser storage    | `model`, `app`                                            |
| `src/ui`       | Components, styles, design tokens                                        | `app`, `model`; never `adapters` (no transport in the UI) |

`dependency-cruiser` enforces the table. The canvas reaches the simulator only through the local bridge
(`streamsim canvas`), which speaks to the director over MCP and serves workspace files; the Go build never depends on this
folder (`test/architecture/canvas_test.go`, `go.mod` `ignore ./canvas`).

## Design system

Semantic tokens in `src/ui/theme/tokens.css` (surfaces, text, accent, status ok/warn/alarm/info, truth), light and dark,
a 4 px spacing grid, six type sizes, system UI and monospace faces (tabular numbers for values). Desktop only: the shell
needs at least 1280 px.

## Commands

| Command                               | Does                                                                                             |
| ------------------------------------- | ------------------------------------------------------------------------------------------------ |
| `npm ci --ignore-scripts`             | Install the pinned dependencies                                                                  |
| `npm run dev`                         | Development server on 127.0.0.1:5173                                                             |
| `npm run check`                       | Every gate: format, types, lint, layers, duplication, dead code, tests with coverage, build, e2e |
| `make canvas-check` (repository root) | The same, from a clean install, plus `npm audit`                                                 |

End-to-end tests run against the real `streamsim` binary and real files; locally they use the installed Chrome, in CI the
pinned Chromium.
