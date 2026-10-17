# Canvas architecture (generated)

This table is generated from [architecture.json](architecture.json) by
`node scripts/write-architecture-doc.ts` and checked by `npm run architecture`, so the two cannot drift. It names the
modules that exist, what each may import and who calls it. The target structure, the reasons and the migration order are
in [docs/simulation-canvas/ui-architecture-20261016/README.md](../../docs/simulation-canvas/ui-architecture-20261016/README.md).

| Module | Kind | Layer | Lifetime | Responsibility | May import | Callers |
| --- | --- | --- | --- | --- | --- | --- |
| `adapters` | adapter | adapters | singleton | bridge and browser edges: transport, protocol encoding and boundary decoding | `app`, `app/ports`, `model`, `model/workspace`, `test` | `root` |
| `app` | application | app | singleton | use cases, state machines and the ports the outside must implement | `app/ports`, `app/session`, `app/shared`, `model`, `model/workspace`, `test` | `root`, `test`, `ui` |
| `app/ports` | foundation | app | singleton | dependency interfaces and time/browser ports required by application modules | `model`, `model/workspace` | `adapters`, `app`, `app/session`, `test`, `ui` |
| `app/session` | application | app | singleton | bridge pairing state and recovery | `app/ports`, `app/shared` | `root`, `ui` |
| `app/shared` | foundation | app | singleton | observable state and operation identity primitives shared by application modules | — | `app`, `app/session` |
| `model` | model | model | values | immutable product values, pure transformations and parsing of canvas-owned documents | `model/graph`, `model/workspace` | `adapters`, `app`, `app/ports`, `model/graph`, `test`, `ui`, `ui/graph` |
| `model/graph` | model | model | values | plant-view projection, graph geometry, wire routing and graph presentation values | `model` | `model`, `test`, `ui`, `ui/graph` |
| `model/workspace` | model | model | values | workspace catalogue records and run listing projections | `model/shared`, `test` | `adapters`, `app`, `app/ports`, `model`, `test`, `ui` |
| `root` | composition | root | singleton | the composition root: construct adapters and application objects and render the shell | `adapters`, `app`, `app/session`, `ui`, `model` | — |
| `test` | test-support | test | per-test | shared test drivers, fixtures and harnesses | — | `adapters`, `app`, `app/session`, `model`, `model/graph`, `model/workspace`, `ui`, `ui/graph` |
| `ui` | ui | ui | components | rendering, input binding and component-local ephemeral state | `app`, `app/ports`, `app/session`, `model`, `model/graph`, `model/workspace`, `test`, `ui/graph`, `ui/primitives`, `ui/session` | `root`, `test`, `ui/graph` |
| `ui/graph` | ui | ui | components | the streamsim plant graph: nodes, edges, controls, outline and viewport | `ui`, `model`, `model/graph` | `ui` |
| `ui/primitives` | ui | ui | components | generic state subscription hook and visually stable shared building blocks | — | `test`, `ui` |
| `ui/session` | ui | ui | components | pairing screen and its presentation | `app/session`, `ui/primitives` | — |
