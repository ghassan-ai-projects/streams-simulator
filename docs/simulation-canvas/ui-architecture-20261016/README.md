# Canvas modular architecture design

- **Date:** 2026-10-16
- **Status:** accepted architecture; implemented on branch `refactor-ui-code` (2026-10-17), with the deviations recorded in [PLAN.md §0](PLAN.md)
- **Scope:** `canvas/` application architecture, module ownership, data and control flow, useful abstractions, enforceable constraints, and test design
- **Preserves:** simulator ownership of validation, composition, simulation, truth, scoring, and verification

> **Baseline notice.** This document is the design as accepted on 2026-10-16. Sections that name `ui/home`, `ui/run`, `ui/pairing`, `PlantScreen.module.css` or other pre-migration paths describe the baseline the migration started from, not the current tree (those are now `ui/workspace`, `ui/review`, `ui/session` and feature-owned stylesheets). The current module list, allow-lists and callers are in [canvas/ARCHITECTURE.md](../../../canvas/ARCHITECTURE.md), generated from `canvas/architecture.json`.

## 1. Decision

Keep the existing four technical layers and organize the product inside them as vertical modules.

The current `model`, `app`, `adapters`, and `ui` separation is sound. The repository already enforces its broad dependency direction, and the current checks report no dependency violations or token clones. A rewrite, a new state library, generic repositories, or a general service layer would add machinery without solving a demonstrated problem.

The architecture adds boundaries where the code now has real shared ownership or unsafe ambiguity:

1. Split the application into product modules for session, workspace, plant, editor, launch, and review, with shell and graph as explicit supporting modules.
2. Give each module a small public facade and hide its implementation under an enforced `internal/` boundary.
3. Give shared plant-graph presentation its own `graph` UI module instead of letting editor and review depend on `ui/plant` internals.
4. Replace `Bridge.action(name, args) -> Outcome<unknown>` at application call sites with typed simulator operations decoded by the adapter.
5. Give every asynchronous use case an explicit session and request identity so an earlier request or retained callback cannot publish or install side effects into a later session.
6. Model editor save and compose as one application transaction tied to the exact document revision and file versions it started with.
7. Replace timing-based test coordination and policy-bearing generic fakes with controlled, explicit test drivers.

The result is a modular monolith. It keeps one React application, one composition root, one bridge transport, and one state mechanism.

## 2. Evidence from the current code

This design is based on source inspection and focused checks. It does not claim a browser or runtime defect unless one was reproduced.

### What is already good

- `canvas/src/main.tsx` is a clear composition root.
- `model` contains pure rules and types.
- application stores expose external-store snapshots consumed through `useSyncExternalStore`.
- adapters implement browser and bridge edges.
- UI code does not import transport code.
- dependency-cruiser currently reports no violations across 185 modules.
- `jscpd` currently reports no clone of 75 tokens or more.
- 445 Vitest unit and component tests pass.
- current line coverage is 98.82% for `model` and 97.47% for `app`.
- Playwright exercises the real `streamsim` binary and real workspace files.

These are reasons to evolve the design in place.

### Boundaries that need improvement

| Evidence                                                                                                                                      | Architectural consequence                                                                                                                                                                           |
| --------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `EditorScreen` and `RunScreen` import `PlantGraph`, `GraphControls`, `Outline`, `canvasAspect`, and `PlantScreen.module.css` from `ui/plant`  | The graph is a shared UI capability, but its ownership currently says it belongs to the plant-viewing feature.                                                                                      |
| `Bridge.action` accepts a string and untyped argument record and returns `unknown`                                                            | Stores know protocol operation names and decode successful responses. This contradicts the quality bar's intent that boundary messages are parsed once before the app sees them.                    |
| `ReviewStore`, `EditorStore`, `PlantStore`, and `WorkspaceStore` use different late-result checks                                             | Async ownership is a repeated responsibility without one invariant or reusable mechanism.                                                                                                           |
| `ReviewStore.take` mutates the current playback index before its later update checks the run id                                               | A retained callback from an old run is a source-visible risk. The current fake removes the callback before its stale-callback test tries to call it, so that test does not exercise the risk.       |
| `EditorStore.save` captures a draft, awaits the write, then updates whichever ready draft is current; `compose` reads the current draft again | A route change or a new edit during the operation needs an explicit revision rule. Source inspection shows missing protection; it has not been reproduced as a user-visible failure in this review. |
| `WorkspaceStore.reload` has no request generation                                                                                             | Two overlapping reloads can publish in completion order instead of request order. This is a risk requiring a regression test.                                                                       |
| JSON record helpers and tolerant fallback readers recur across bridge, report, plant, composition, playback, library, and alarm parsing       | A small common reader vocabulary would remove mechanical repetition. It must not become a decoder framework or make malformed required data look valid.                                             |
| `FakeBridge` implements plant-creation policy                                                                                                 | A general test double reproduces application/server behavior, so tests can pass while the real binary behaves differently.                                                                          |
| `plantStore.test.ts` uses a five-millisecond delay to order overlapping reads                                                                 | The test is timing-dependent despite the repository rule against fixed sleeps.                                                                                                                      |
| several production files sit at 294–300 lines                                                                                                 | The size gate is containing growth, but the largest files combine multiple named responsibilities and should be split by ownership during the relevant module migration.                            |

## 3. Target architecture

Use two dimensions together:

- **Technical layers** control dependency direction and side effects.
- **Product modules** own use cases and vocabulary.

```text
composition root
      |
      +-------------------- constructs modules and adapters
      v
ui/<module>  ----------->  app/<module>  ----------->  model/<module>
   |                            |
   | renders state             | calls typed ports
   | sends user intent         v
   +----------------------  app/ports  <------------- adapters/bridge
                                                       |
                                                       v
                                                local Go bridge
```

The arrows are the only permitted directions. Runtime responses travel back through the same boundary as typed values; this does not reverse the source dependency.

### 3.1 Technical layers

| Layer      | Owns                                                                                                                              | Must not own                                                                                                           |
| ---------- | --------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| `model`    | immutable product values, pure transformations, selectors, parsing of canvas-owned documents, graph geometry, display projections | React, browser APIs, timers, transport, store mutation, simulator-owned validation                                     |
| `app`      | use cases, state machines, operation identity, orchestration, ports needed from the outside                                       | React, `fetch`, `EventSource`, DOM, CSS, protocol JSON, transport operation strings, direct wall-clock or timer access |
| `adapters` | HTTP/SSE/browser implementation, protocol request construction, boundary decoding into typed app/model values                     | product workflow, UI state, simulator rules, cross-feature orchestration                                               |
| `ui`       | rendering, input binding, focus, browser-owned visual timing, component-local ephemeral state                                     | transport calls, response decoding, simulation rules, multi-store business orchestration                               |
| `main.tsx` | construction and dependency injection only                                                                                        | use cases, parsing, policy, derived product state                                                                      |

The existing `Observable` plus `useSyncExternalStore` remains the state mechanism. A different state library is warranted only if a measured problem cannot be solved within this design.

### 3.2 Product modules

Each module that appears in a layer has one public facade. The facade is the module's `index.ts`: it exports a small interface, stable value types, and the intended public values or construction function. Its implementation lives under `internal/` and is never imported directly from another module.

```text
canvas/src/
  model/
    shared/
    shell/{index.ts,internal/}
    workspace/{index.ts,internal/}
    plant/{index.ts,internal/}
    editor/{index.ts,internal/}
    launch/{index.ts,internal/}
    review/{index.ts,internal/}
    graph/{index.ts,internal/}
  app/
    shared/
    ports/
    session/{index.ts,internal/}
    shell/{index.ts,internal/}
    workspace/{index.ts,internal/}
    plant/{index.ts,internal/}
    editor/{index.ts,internal/}
    launch/{index.ts,internal/}
    review/{index.ts,internal/}
  adapters/
    bridge/{index.ts,internal/}
    browser/{index.ts,internal/}
  ui/
    primitives/
    session/{index.ts,internal/}
    shell/{index.ts,internal/}
    workspace/{index.ts,internal/}
    plant/{index.ts,internal/}
    editor/{index.ts,internal/}
    launch/{index.ts,internal/}
    review/{index.ts,internal/}
    graph/{index.ts,internal/}
```

This is a target ownership map, not permission to create all folders at once. Migrate one module at a time and create a folder only when it has a concrete responsibility and callers.

| Module       | Responsibility                                                                    | Public API examples                                             |
| ------------ | --------------------------------------------------------------------------------- | --------------------------------------------------------------- |
| `session`    | bridge pairing state and recovery                                                 | `Session`, `SessionState`, `createSession`                      |
| `workspace`  | session-visible workspace catalog, reload lifecycle, home projections             | `Workspace`, `WorkspaceState`, `createWorkspace`                |
| `plant`      | load one plant projection, component specifications, selection, plant facts       | `PlantModule`, `PlantState`, `Plant`, `createPlantModule`       |
| `editor`     | composition draft, history, edits, save/compose transaction, conflict handling    | `EditorModule`, `Editor`, `EditorState`, `createEditorModule`   |
| `launch`     | plan validation and execution workflow, progress, early stop, rerun proposal      | `Launcher`, `LaunchPlan`, `LaunchState`, `createLaunch`         |
| `review`     | report/verify loading, observer playback, selection and review presentation       | `ReviewModule`, `Reviewer`, `ReviewState`, `createReviewModule` |
| `graph`      | plant-view projection, geometry, routing, React Flow mapping and interaction      | `PlantGraph`, `GraphView`, `GraphControls`                      |
| `shell`      | routing, global panels, theme, session/workspace coordination, module composition | `App`, `Navigation`, `Panels`, route factories                  |
| `primitives` | visually and semantically stable UI building blocks                               | panel chrome, metadata parts, status badge                      |

`shared` is not a dumping ground. Code enters it only after two product modules require the same stable concept and neither module is its natural owner.

### 3.3 Public facade and hidden internals

The application facade follows the same intent as a Go module facade: callers see capabilities and state, while concrete controllers, mutable bookkeeping, state transitions, and workflow helpers remain private.

```text
app/editor/
  index.ts                         public facade only
  contract.ts                      public declarations, re-exported by index.ts
  editor.contract.test.ts         public contract tests
  internal/
    controller.ts                 concrete implementation
    state.ts                      private state constructors/transitions
    saveAttempt.ts                save workflow
    history.ts                    undo/redo mechanics
    controller.test.ts            internal behavior tests
```

The public application API is an interface returned by a factory:

```ts
export interface ReadableState<S> {
  snapshot(): S;
  subscribe(listener: () => void): () => void;
}

export interface Editor extends ReadableState<EditorState> {
  open(path: string): Promise<void>;
  close(): void;
  select(instanceId: string | null): void;
  apply(command: EditCommand): void;
  undo(): void;
  redo(): void;
  save(): Promise<void>;
}

export interface EditorModule {
  createSession(): Editor;
}

export interface EditorDependencies {
  readonly gateway: EditorGateway;
  readonly operations: OperationIdentity;
}

export function createEditorModule(
  dependencies: EditorDependencies,
): EditorModule;
```

`internal/controller.ts` may define `EditorController`, but `index.ts` does not export that class. Larger public state and command declarations may live in `contract.ts`, but callers still import the directory entrypoint. UI accepts `Editor`, not `EditorController` or today's concrete `EditorStore`. This makes the public contract small and lets internals change without changing callers.

The facade mechanism differs slightly by layer:

```ts
// app/editor/index.ts: construction delegates without exporting the session controller.
export type {
  Editor,
  EditorDependencies,
  EditorModule,
  EditorState,
} from "./contract";
import type { EditorDependencies, EditorModule } from "./contract";
import { EditorController } from "./internal/controller";

export function createEditorModule(
  dependencies: EditorDependencies,
): EditorModule {
  return {
    createSession: () => new EditorController(dependencies),
  };
}

// ui/editor/index.ts: one intended component is an explicitly named public value.
export { EditorScreen } from "./internal/EditorScreen";
export type { EditorScreenProps } from "./contract";

// model/graph/index.ts: selected pure operations form the public model API.
export { layoutGraph, routeGraph } from "./internal/layout";
export type { GraphLayout, GraphView } from "./contract";
```

An explicitly named export from `internal/` makes that value public through the facade; it does not make the source path public. Its signature may reference only public contract types. Wildcard exports, exported concrete controllers, and signatures that expose internal declarations remain forbidden. This avoids redundant React or model wrappers while preserving one import path and a reviewable public surface.

Apply the same pattern at each layer:

| Layer    | Public facade                                                               | Hidden internals                                                      |
| -------- | --------------------------------------------------------------------------- | --------------------------------------------------------------------- |
| model    | stable values and pure operations needed by other modules                   | parsing mechanics, intermediate geometry, private selectors           |
| app      | capability interface, observable state union, dependency interface, factory | controller classes, transition helpers, caches, operation bookkeeping |
| adapters | typed implementation factory used by the composition root                   | HTTP helper, route strings, JSON decoders, SSE framing                |
| UI       | screen/component and its public props                                       | hooks, panel parts, local formatting, feature CSS                     |

Facade rules:

1. `index.ts` contains explicit exports and small construction delegation only. It holds no product rule or workflow.
2. A module exports the smallest capability its callers need. It does not export its internal controller, mutable collections, caches, protocol documents, or helper functions.
3. State exposed by a facade is immutable and uses a closed discriminated union.
4. Dependencies enter through the factory. Internals do not reach into a global container.
5. Cross-module calls target the public interface. A caller cannot downcast to the implementation.
6. One use case has one owning facade operation. The UI does not rebuild it from several lower-level methods.
7. Public contract tests import only `index.ts`. Private unit tests may live under `internal/` and test the rule they own.
8. `index.ts` may explicitly name an intended public UI component or pure model operation implemented under `internal/`. It never uses `export *`, exports a concrete application controller, or exposes an internal type in a public signature.

TypeScript has no directory-level equivalent of Go's `internal` enforcement inside one application package. The privacy guarantee therefore comes from repository gates: dependency-cruiser, restricted imports, API review, and tests. Splitting the canvas into npm packages only to obtain package export maps would add build and dependency-management cost without a current need.

### 3.4 Module dependency rules

Module dependencies are declared per layer. A single rule such as “features may use graph” is insufficient because model, application, adapter, and UI code have different responsibilities. The following is the exhaustive target contract; Phase 0 verifies it against concrete declarations and records it in the module register before files move.

```text
model:
  shared
  shell     -> shared
  workspace -> shared
  plant     -> shared
  graph     -> shared, plant
  editor    -> shared, plant, graph
  launch    -> shared, plant
  review    -> shared, workspace, plant, graph

app:
  shared
  ports     -> public model APIs used by port contracts
  session   -> shared, ports
  workspace -> shared, ports, model/workspace
  plant     -> shared, ports, model/plant
  editor    -> shared, ports, model/editor, model/plant, model/graph
  launch    -> shared, ports, model/launch, model/plant
  review    -> shared, ports, plant, model/review, model/workspace, model/plant, model/graph
  shell     -> shared, ports, session, workspace, plant, editor, launch, review, model/shell

adapters:
  browser   -> app/ports, model/shell
  bridge    -> app/ports, model/shared, model/workspace, model/plant,
               model/editor, model/launch, model/review

ui:
  primitives
  graph     -> primitives, model/graph
  session   -> primitives, app/session
  workspace -> primitives, app/workspace
  plant     -> primitives, graph, launch, app/plant, app/launch
  editor    -> primitives, graph, app/editor
  launch    -> primitives, app/launch
  review    -> primitives, graph, launch, app/review, app/launch
  shell     -> primitives, session, workspace, plant, editor, review, app/shell
```

`ports` is a foundation containing only dependency interfaces. Feature-specific gateway declarations remain grouped by feature inside it; it owns no workflow. Its model allow-list is the exact union of public values used in those signatures, recorded in the machine-readable register rather than a blanket `model/**` permission.

`app/shell` owns navigation, theme, panel state, session gating, the `ApplicationModules` composition contract, and application coordinators. This is why its registered edges name every public application module. `ui/shell` imports only that application contract plus public UI screens. The Show Build coordinator imports the public editor capability and the shell-owned panel capability. Session-to-workspace opening is another shell coordinator.

No feature imports `shell`; navigation and panel intents enter features as narrow callbacks or route contracts. The current breadcrumb slot moves to `ui/primitives/breadcrumb` so shell and feature screens can use it without a UI dependency cycle.

Before the first move, confirm every edge against the actual declarations. In particular, `plantView.ts` currently imports review-oriented helpers. The migration must invert that dependency: `model/graph` accepts feature-neutral presentation values; review builds its live/readout presentation and passes it into graph. `ui/graph` consumes the public `model/graph` projection rather than importing `model/plant` directly, and `graph` never imports `review`.

Within each layer:

- Feature modules do not import another feature's `internal/` files. They import its directory entrypoint only.
- editor, plant, and review UI may use the public UI graph API; model dependencies follow the map above.
- `launch` receives `Plant` and selection values through its API; it does not import plant UI.
- `shell` may compose module entrypoints but contains no feature policy.
- Cross-module workflows live in a named application coordinator owned by the initiating use case.
- Circular module imports are forbidden even when TypeScript can load them.

Examples:

- “Open a run from a plant” is navigation intent emitted by plant UI and handled through the navigation application API.
- “Show Build” currently coordinates editor selection and panel state inside `ui/shell/showBuild.ts`. Move that operation to an application coordinator because it changes two application stores.
- “Use the run's recorded plant for review” is owned by the review application module. Review requests the plant projection through a typed plant capability; `RunScreen` should not coordinate two store loading state machines.

### 3.5 TypeScript code rules

The compiler configuration already provides a strong baseline: strict mode, unchecked-index protection, exact optional properties, exhaustive switches, unused-code checks, and isolated modules. ESLint already rejects `any`, non-null assertions, TypeScript suppression comments, oversized functions/files, and excessive complexity. Preserve these gates.

Add conventions that express module ownership and product meaning:

| Construct                   | Rule                                                                                                        | Example                                                                   |
| --------------------------- | ----------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------- |
| React component and file    | PascalCase noun; one primary component per file                                                             | `ReviewTimeline.tsx`, `EditorToolbar.tsx`                                 |
| non-React source file       | camelCase named for one responsibility; no `utils.ts`, `helpers.ts`, `common.ts`, or unqualified `types.ts` | `operationIdentity.ts`, `saveAttempt.ts`, `reportDecoder.ts`              |
| module capability interface | role noun without `I`, `Service`, `Manager`, or implementation suffix                                       | `Editor`, `Reviewer`, `WorkspaceGateway`                                  |
| concrete private class      | capability plus implementation meaning; never exported from the module                                      | `EditorController`, `BridgeWorkspaceGateway`                              |
| factory                     | `create` plus the public capability                                                                         | `createEditor`, `createReviewer`                                          |
| command method              | imperative domain verb                                                                                      | `open`, `apply`, `save`, `stop`, `selectAlarm`                            |
| query or projection         | domain noun or an established `...Of` transformation                                                        | `snapshot`, `plantsOf`, `runFacts`, `openingViewport`                     |
| predicate                   | `is`, `has`, `can`, or a domain phrase that reads as a fact                                                 | `isDirty`, `hasSelection`, `canUndo`, `workspaceEmpty`                    |
| state union discriminator   | `kind` with short domain values                                                                             | `{ kind: "loading" }`, `{ kind: "ready"; plant: Plant }`                  |
| expected failure            | closed feature-owned union returned in `Outcome`; never encoded only in a message string                    | `{ kind: "conflict"; path: string; diagnostic: Diagnostic }`              |
| event callback prop         | `on` plus completed user intent                                                                             | `onSave`, `onSelect`, `onStop`; avoid exposing setters such as `setState` |

#### Model states instead of flag combinations

Public module state uses readonly discriminated unions so impossible combinations cannot be constructed:

```ts
type PlantState =
  | { readonly kind: "idle" }
  | { readonly kind: "loading"; readonly path: string }
  | { readonly kind: "ready"; readonly path: string; readonly plant: Plant }
  | {
      readonly kind: "failed";
      readonly path: string;
      readonly failure: PlantFailure;
    };
```

Do not represent this as independent `loading`, `ready`, `failed`, and `plant` fields. Booleans remain appropriate for local facts such as `disabled`, `selected`, or `systemPrefersDark`.

Avoid boolean flag arguments at public module boundaries because calls such as `create(source, id, title, true)` and `plantView(plant, true)` hide meaning and age poorly. Use a closed option when the choice changes behavior:

```ts
type PlantCreation = {
  readonly source: PlantSource;
  readonly mode: "copy" | "reference";
};
type GraphPurpose = "inspect" | "edit" | "review";
```

Local private helpers may accept a boolean when it represents one obvious fact and does not select an operating mode.

#### Keep public values immutable and narrow

- Public state, inputs, arrays, maps, and result records are readonly.
- Export the smallest type the caller needs. Do not export concrete controllers, internal state builders, caches, decoder intermediates, or mutable collections.
- Use `satisfies` to check maps and configuration without widening their values. Use `as const` only to preserve literal values.
- Type assertions do not substitute for decoding. Production `as unknown as T`, unchecked casts from `unknown`, and casts used to bypass module contracts are forbidden. Boundary decoders narrow one field at a time.
- Use `null` for an explicit domain absence in public state. Use optional properties for values legitimately omitted from an input or protocol document. Do not use both for the same meaning.
- Add a branded identifier only when two string identifiers are repeatedly interchangeable at compile time and a real defect or high-risk call site justifies it. Do not wrap every string speculatively.

#### Separate commands, state, and effects

Application facades expose commands and one observable snapshot. Derived values are pure model selectors rather than additional mutable fields. A command owns its entire use case; callers do not reproduce a workflow by ordering several lower-level mutations.

Expected product failures return typed outcomes and become state transitions. Exceptions are reserved for violated internal invariants or truly unexpected adapter/runtime failures, and the module boundary converts them into an explicit unexpected failure before publication. UI components do not inspect exception classes, HTTP status, or protocol error strings.

React components receive immutable values and intent callbacks or one narrow application capability. They do not construct stores, decode protocol data, mutate shared objects, or use an effect to calculate renderable derived state. Custom hooks may own browser lifecycles such as focus, resize observation, and animation; their cleanup must release every installed resource.

`index.ts` is the only barrel file. It explicitly exports the module contract and construction function. Nested catch-all barrels are forbidden because they obscure dependency direction and make accidental public APIs easy.

#### Enforcement

- Keep the existing TypeScript and ESLint strict rules unchanged.
- Extend the architecture checker to reject generic grab-bag filenames, deep imports, production double assertions, and workflow code in module entrypoints.
- Use `@typescript-eslint/naming-convention` only for objective boundaries such as exported types and booleans if it can be introduced without false positives. Domain vocabulary remains a review responsibility.
- Public facade contract tests compile against `index.ts` and use interface-shaped fakes, so tests cannot depend on concrete controllers.
- State-transition tests enumerate every `kind` and prove invalid late completions cannot change the current state.
- Do not add lint rules for subjective word preferences. The review checklist asks whether a name states product intent and whether the abstraction has a present owner and caller.

### 3.6 React composition patterns

Use a small set of React patterns that reinforce the module architecture:

| Pattern                                      | Use in the canvas                                                                                                                         | Limit                                                                                        |
| -------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| manual composition root                      | `main.tsx` constructs adapters and module facades, wires dependencies, and passes the completed application to React                      | no service locator, decorator container, or dependency-injection framework                   |
| external store adapter                       | keep `useSyncExternalStore` behind `useStore`; application state remains usable and testable without React                                | do not copy external state into component state                                              |
| route controller plus pure screen            | one thin route component opens/closes one feature facade and passes immutable state plus intent callbacks to a mostly pure screen         | do not create container/view pairs for every small component                                 |
| discriminated workflow state                 | render loading, ready, failed, conflict, and running branches from closed application state unions                                        | do not reproduce the state machine with local booleans                                       |
| render-time derivation                       | calculate cheap display values during render; keep expensive pure graph/report projections in model selectors and memoize when justified  | do not use an effect to derive renderable state from props or another state value            |
| effect as external synchronization           | use effects for subscriptions, timers, focus, observers, keyboard listeners, and route-resource lifetime                                  | every effect names the external resource, installs it once per identity, and returns cleanup |
| semantic primitive with slots                | panel, dialog, field, status, metadata, disclosure, and button primitives own accessibility, states, and styles                           | create one only for repeated semantics; do not build a generic component framework           |
| intent callbacks                             | children receive `onSave`, `onSelect`, or `onStop`; the route component translates intent to one facade command                           | do not pass setters, mutable stores, or several low-level operations to leaf components      |
| expected-error rendering plus error boundary | feature states render recoverable failures; one root boundary catches unexpected render defects and provides a stable recovery action     | an error boundary does not replace typed application failures                                |
| local reducer for complex ephemeral UI       | use `useReducer` only when one component owns several related temporary transitions, such as an open dialog with a multi-step local draft | application workflows remain in app modules                                                  |

#### Composition root and module access

Keep manual dependency injection because it is explicit and currently small. The composition root constructs application singletons and factories for route-scoped capabilities; it does not construct one permanent editor or reviewer session:

```tsx
const modules = {
  workspace: createWorkspace({ gateway: workspaceGateway }),
  editor: createEditorModule({ gateway: editorGateway, operations }),
  review: createReviewModule({ gateway: reviewGateway, clock, scheduler }),
};

root.render(<App modules={modules} />);
```

`EditorModule` and `ReviewModule` are long-lived factories with `createSession()`; they do not hold an active document, run subscription, or route-specific mutable state. `App` may know the set of modules because it is the React composition boundary. It passes only the relevant module factory or singleton to a route. The current `Screen` helper receives the entire `Stores` object, while `PlantScreen` and `RunScreen` each receive several concrete stores. Replace this gradually with one route-owned application facade or an explicitly declared coordinator per cross-module operation. Do not hide the same broad object in React Context.

Context is appropriate for stable tree-scoped presentation facilities such as the existing breadcrumb slot, theme presentation, or an accessibility id scope. It is not the default state mechanism and must not provide all modules to every descendant. Frequently changing feature state continues through `useSyncExternalStore` at the narrowest route or feature boundary that needs it.

#### Route controller and pure screen

A route component creates the route session once and owns its open resource lifetime. A screen renders state and emits intent:

```tsx
function ReviewRoute({ module, runId }: ReviewRouteProps) {
  const [reviewer] = useState(() => module.createSession());
  const state = useStore(reviewer);

  useEffect(() => {
    void reviewer.open(runId);
    return () => {
      reviewer.close();
    };
  }, [reviewer, runId]);

  return (
    <ReviewScreen
      state={state}
      onStop={() => reviewer.stop()}
      onSelectAlarm={(alarm) => reviewer.selectAlarm(alarm)}
    />
  );
}
```

`close()` releases every run-specific subscription, timer, and pending-operation identity and is safe to call repeatedly. The same route facade can be reopened after `close()`, which makes React Strict Mode's setup-cleanup-setup replay safe. When the route actually unmounts, React drops the closed facade and no external resource remains. A route facade that owns a resource outside this open/close lifetime must additionally expose idempotent `dispose()` and the route must prove Strict Mode-safe recreation before using it.

The facade owns request identity, error transitions, and coordination with plant data. The route owns mount lifetime. `ReviewScreen` does not load resources or inspect transport failures. Smaller descendants receive values and callbacks rather than the facade when they do not need its full capability.

The editor route follows the same rule. It creates one editor session for the route, installs the unsaved-navigation guard through a shell capability that returns a disposer, and removes that guard during cleanup. The composition root no longer closes over one permanent editor merely to answer the navigation guard.

This split is useful at route and substantial feature boundaries. A small button, row, or label remains one component; adding a controller pair would obscure it.

#### Effects and hooks

Treat every effect as synchronization with something outside the render calculation. Valid owners include:

- an application facade session opened for the current route;
- a DOM observer, focus operation, portal target, or keyboard listener;
- a browser animation or clock used only for presentation;
- a third-party graph instance.

An effect must be safe under React Strict Mode's mount-cleanup-mount sequence. Cleanup runs before the external identity changes, and late asynchronous work is rejected by the application session identity. Event-triggered work starts in the event handler; it does not set a boolean solely to make an effect notice and begin the operation.

Custom hooks package one browser-facing lifecycle or reusable React binding. Name the resource or behavior, such as `useResizeObserver`, `useFocusReturn`, `usePlaybackClock`, or `useStore`. A hook does not become a hidden application service, transport client, or cross-feature coordinator.

#### Derived data and memoization

Keep one source of truth. Do not store values that can be derived from the current snapshot and props. Cheap derivation happens directly during render. Pure expensive work such as graph layout, wire routing, or large report projections belongs in `model` and may use `useMemo` at the UI boundary when inputs are stable.

Use `useMemo` and `useCallback` for measured expensive work, third-party identity requirements, or effect dependency stability. Do not wrap every calculation or handler pre-emptively. Referential stability is an implementation requirement only where a consumer relies on it.

#### Forms and drafts

Temporary input that can be discarded with one component stays local and controlled. A draft that participates in undo/redo, navigation guards, save conflicts, or several panels belongs to the editor application module. Do not mirror an application draft into local form state without an explicit edit buffer and commit/cancel rule.

#### Error and loading boundaries

Expected failures are application state and render beside their recovery action. Add one root React error boundary for unexpected render exceptions. It shows a stable fallback, preserves a diagnostic for support, and offers reload or return-home recovery. Add a narrower graph boundary only if React Flow failures are shown to be independently recoverable.

Do not introduce Suspense for bridge loading in this migration. The external stores already model progress, cancellation, stale results, partial evidence, and recovery explicitly. Route-level code splitting is justified only by a measured bundle or startup problem.

#### React enforcement and tests

- Keep the React Hooks ESLint rules and Strict Mode enabled.
- Architecture checks prevent UI from importing adapters and leaf components from receiving concrete private controllers.
- Route integration tests prove open/close identity, Strict Mode cleanup, and late-completion rejection.
- Pure screen tests cover accessible branches and emitted intent without mocking transport.
- Primitive tests prove keyboard, focus, semantics, and state attributes; Chrome proves layout and visual integration.
- Review every new effect for its external resource, identity, cleanup, and late-callback behavior.
- Profile before adding `memo`, selector subscriptions, virtualization, or route-level lazy loading.

## 4. Data and control flow

There are no direct shortcuts around this path:

```text
user/browser event
  -> UI module translates it to intent
  -> one application use case owns the operation
  -> use case calls a typed port
  -> adapter encodes the bridge request
  -> Go bridge calls the simulator
  -> adapter validates and decodes the response once
  -> use case accepts the result only for its active operation identity
  -> store publishes a new immutable state
  -> UI renders the state
```

### 4.1 Commands

A component may call one application operation. It must not reproduce a workflow by sequencing low-level bridge calls.

Example:

```ts
// UI intent
editor.save();

// application-owned port
interface EditorGateway {
  writeComposition(
    input: WriteComposition,
  ): Promise<Outcome<WrittenComposition>>;
  composeDomain(input: ComposeDomain): Promise<Outcome<ComposedDomain>>;
}
```

The adapter may still use one generic internal HTTP helper. Generic transport mechanics belong in the adapter; generic business operations do not belong in the application.

### 4.2 Queries

Derived values follow one rule:

- pure product derivation belongs in `model` selectors;
- stateful loading and caching belong in the owning `app` module;
- display-only derivation belongs beside the UI component;
- simulator-owned facts are requested from the simulator and never recomputed in TypeScript.

The UI must not reach through a store to inspect adapter data or parse raw responses.

### 4.3 Events and streams

Every subscription returns a disposer and belongs to one application session identity.

For a run stream:

1. `ReviewStore.open(worldId)` creates a new session id and a new `PlaybackIndex` captured by that session.
2. Report requests, verify requests, recorded-plant reads, and stream callbacks capture the same session id.
3. A callback checks the session id before mutating any captured state or installing another side effect.
4. `close()` invalidates the session before disposing the transport subscription.
5. Reopening the same `worldId` creates a different session id; resource equality is not operation identity.

This rule also applies to workspace reloads, plant loads, editor opens, and save transactions. Resource sessions and individual requests have separate identities. For example, `WorkspaceStore.close()` invalidates the session; a reload that later completes cannot install a watcher even if it was the newest reload request in that now-closed session.

### 4.4 Failures

- Adapters distinguish transport failure, protocol failure, and malformed successful payload.
- The application translates typed failures into feature state.
- The UI renders the failure and recovery action.
- A malformed required response must not silently become an empty but successful report, plant, or catalog.
- Tolerant projection of domain documents remains separate from validation. The simulator continues to own domain validation.

## 5. Useful abstractions and how to implement them

### A1. Shared graph module

**Problem:** plant view, editor, and run review share the same graph implementation through files owned by `ui/plant`.

**Abstraction:** `ui/graph` owns `PlantGraph`, node/edge mapping, graph controls, viewport fitting, outline, canvas aspect, graph keyboard behavior, and neutral graph styles. `model/graph` owns `PlantView`, layout, routing, viewport rules, and `NodeLive`-style presentation values.

**How:**

1. Define `ui/graph/index.ts` and `model/graph/index.ts` as public APIs.
2. Move graph files with behavior unchanged and update imports.
3. Rename plant-specific CSS selectors only where ownership becomes misleading.
4. Keep feature orchestration in `PlantScreen`, `EditorScreen`, and `RunScreen`.
5. Add dependency rules preventing feature modules from importing graph internals.

**Limit:** do not build a generic diagram engine. The module renders streamsim plant graphs.

### A2. Typed simulator operations

**Problem:** `Bridge.action(string, Record<string, unknown>)` lets protocol names, argument shapes, and response decoding spread through application stores.

**Abstraction:** small capability interfaces grouped by owning feature.

```ts
interface ReviewGateway {
  report(runId: string): Promise<Outcome<Report>>;
  verify(runId: string): Promise<VerificationResult>;
  recordedPlant(runId: string): Promise<Outcome<DomainDocument>>;
  follow(runId: string, follower: JournalFollower): () => void;
}

interface LaunchGateway {
  createWorld(input: CreateWorld): Promise<Outcome<CreatedWorld>>;
  advanceClock(input: AdvanceClock): Promise<Outcome<void>>;
  injectFault(input: InjectFault): Promise<Outcome<void>>;
  invokeEffector(input: InvokeEffector): Promise<Outcome<EffectorOutcome>>;
  endRun(worldId: string): Promise<Outcome<void>>;
  destroyWorld(worldId: string): Promise<Outcome<void>>;
}
```

`DomainDocument` is an opaque, object-shaped domain document for display projection; accepting it does not validate the domain. `JournalFollower` receives the decoded `JournalEntry` union, not JSON text. `VerificationResult` is a typed union that preserves a replay verdict returned with an unsuccessful HTTP response, distinct from a transport or malformed-payload failure.

**How:**

1. Define typed ports in `app/ports` from current use cases.
2. Implement them in one bridge adapter using the existing generic HTTP request helper privately.
3. Move success-payload decoding from stores into adapter boundary decoders.
4. Run shared valid/invalid fixtures through TypeScript and Go contract tests where a schema is shared.
5. Decode journal lines in the adapter and expose an opaque domain document rather than raw transport JSON.
6. Preserve verification verdicts returned inside unsuccessful responses as a typed result case.
7. Remove raw `action` access from application modules after the last caller migrates.

**Limit:** do not duplicate simulator validation or introduce one interface per HTTP request mechanically. Capabilities follow use-case ownership.

### A3. Operation identity

**Problem:** path or run-id equality does not distinguish two visits to the same resource, and stores implement late-result protection inconsistently.

**Abstraction:** a tiny `OperationSlot` or generation token owned by `app/shared`.

```ts
const operation = requests.begin();
const result = await gateway.load(path);
if (!requests.active(operation)) return;
publish(result);
```

For streams, capture the token and session-owned accumulator in the callback. Invalidate the token before cleanup.

**How:** first write deferred-request and retained-callback tests. Then use the abstraction in workspace reload, plant load, editor open/save, and review open. The abstraction should manage identity only; it should not become a task framework, cache, or cancellation library.

### A4. Editor save transaction

**Problem:** save, file-version update, compose, validation issues, and domain-version update form one use case, but current code reads mutable store state again after awaits.

**Abstraction:** an immutable `SaveAttempt` captured from one draft revision.

```ts
interface SaveAttempt {
  readonly operation: OperationId;
  readonly path: string;
  readonly revision: number;
  readonly document: Composition;
  readonly compositionEtag: string;
  readonly domainPath: string;
  readonly domainEtag: string;
}
```

**Rules supported by the current protocol:**

- the write always uses the captured path, document, and composition etag;
- compose uses the captured source and output paths and the captured output etag;
- a completion may update transport metadata only when its operation is still active;
- edits made while saving remain dirty after that save completes;
- navigating to another document invalidates the attempt;
- a conflict keeps the current draft and exposes recovery state;
- write success followed by compose failure is represented as partial success: the composition file is saved, the domain output is not accepted as current, and retry is available;
- invalidation prevents later state publication but does not pretend to undo an already completed file write or compose command;
- save and compose failure precedence is specified in the state machine.

The current `domain.compose` operation rereads the composition from its path and guards the output etag. It does not accept an expected source etag. Therefore the UI cannot guarantee that an external writer did not replace the source between write and compose. A strict captured-source guarantee requires a separate bridge/simulator contract change that admits the expected composition etag and checks it atomically before composition. This design does not silently assume that change.

This is an application transaction with explicit partial-success states, not an ACID or database abstraction.

### A5. Clock and scheduler ports

**Problem:** launch pacing and review update coalescing use timers inside application stores. A target rule banning direct app timers would fail current code unless the edge is migrated deliberately.

**Abstraction:** small injected ports:

```ts
interface Clock {
  nowMs(): number;
}

interface Scheduler {
  schedule(delayMs: number, task: () => void): () => void;
  delay(delayMs: number, signal?: AbortSignal): Promise<void>;
}
```

The browser adapter implements them. Launch and review receive only the operation they need. Playback animation driven by `usePlayback` remains a UI effect because it is tied to component presence and browser rendering.

**Limit:** no global platform object or general task runtime. Cancellation and deterministic tests are the concrete reasons for these ports.

### A6. Controlled async test drivers

**Problem:** promise flushing by repeated `Promise.resolve()` and a five-millisecond delay make lifecycle tests indirect. `FakeBridge` also implements server policy.

**Abstraction:** deferred operations and retained subscriptions.

```ts
const read = deferred<Outcome<FileContent>>();
gateway.reads.enqueue(read.promise);
const load = store.load(path, workspace);
read.resolve(success(file));
await load;
```

The stream test driver retains the callback even after disposal so a test can deliberately invoke a late callback. The production adapter still disposes normally.

Replace `FakeBridge`'s plant-creation implementation with explicit canned typed outcomes. Test the real creation contract through the existing real-binary boundary tests.

### A7. Small JSON reader vocabulary

**Problem:** `object`, `list`, `text`, and number guards are repeated.

**Abstraction:** four or five pure functions in `model/shared/jsonValue.ts` for structural reading.

**How:** adopt only while touching a parser. Each boundary parser still decides which fields are required and what failure means.

**Limit:** no schema DSL, reflection framework, universal decoder, or dependency. A malformed successful bridge response must remain distinguishable from a legitimate empty value.

## 6. Constraints and enforcement

The architecture is incomplete unless violations fail locally and in CI.

| Constraint                                                                                          | Enforcement                                                                                                                                                                                             |
| --------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Technical imports follow `model <- app <- adapters`; UI imports only app/model and UI public APIs   | extend `.dependency-cruiser.cjs`; run in `npm run deps` and `make canvas-check`                                                                                                                         |
| A feature imports another feature only through `index.ts`                                           | dependency-cruiser rules matching private feature paths; a small architecture fixture proves a forbidden import fails                                                                                   |
| A module's `internal/` tree is reachable only from that module                                      | dependency-cruiser rule parameterized by module root; architecture fixtures prove same-module access and reject cross-module access                                                                     |
| A public facade exposes only its declared contract                                                  | TypeScript-AST architecture check rejects wildcard exports, concrete controller exports, internal types in public signatures, and workflow logic in `index.ts`; named UI/model values are permitted     |
| UI never imports adapters, calls `fetch`/`EventSource`, or names `/api/` routes                     | dependency-cruiser plus ESLint `no-restricted-imports` and `no-restricted-globals`; source architecture test for route literals outside adapters                                                        |
| App never imports React, DOM/browser APIs, CSS, direct timers, clocks, or transport implementations | dependency-cruiser and ESLint restricted imports/globals, enabled after launch/review migrate to Clock/Scheduler ports                                                                                  |
| Model stays pure and does not import npm packages                                                   | keep existing dependency-cruiser rules; add restricted browser globals to model files                                                                                                                   |
| Simulator action names and HTTP routes exist only in production bridge-adapter code                 | source architecture test for known action/route patterns outside `adapters/bridge`, excluding tests/fixtures so contract assertions remain possible; typed ports make production exceptions unnecessary |
| Raw `unknown` crosses only boundary decoder APIs                                                    | TypeScript rule/review: adapter decoders accept `unknown`; application port results are concrete types; no `Outcome<unknown>` in app public APIs                                                        |
| Public module state is readonly and represented by a closed `kind` union                            | facade contract tests plus TypeScript-AST architecture check; transition tests exhaust every state kind                                                                                                 |
| Public behavior choices use named unions or option objects rather than boolean flags                | TypeScript-AST check for exported function, method, and factory parameters; allow booleans that describe a single local fact                                                                            |
| Expected failures are typed by the owning feature before they reach UI                              | no exception-class, HTTP-status, or protocol-code branching in UI; gateway contract tests prove protocol-to-feature failure mapping                                                                     |
| Module files state one responsibility and expose only one explicit entrypoint                       | architecture check rejects generic grab-bag filenames, nested catch-all barrels, production double assertions, and workflow statements in `index.ts`                                                    |
| React Context does not expose the application module container or frequently changing feature state | source architecture check limits application capability access to the composition and route boundaries; review records each new context's stable tree-scoped purpose                                    |
| Every React effect synchronizes one named external resource and owns complete cleanup               | React Hooks lint, Strict Mode route tests, and review of identity, cleanup, and late-callback behavior                                                                                                  |
| Renderable derived data has one source of truth                                                     | React lint and review reject effects that mirror props or application snapshots into local state; model selector tests prove non-trivial derivation                                                     |
| Every async publication and side-effect installation proves active session/request identity         | focused lifecycle tests for A->B, A->B->A, close-while-pending, dispose-then-callback, late watcher installation, and edit-while-awaiting; review checklist item for new async use cases                |
| Cross-module orchestration has one application owner                                                | module API review and dependency rules; UI components may call one use case but may not coordinate multiple stores for a product operation                                                              |
| Shared code requires at least two real callers and a stable common meaning                          | review rule recorded in this design; `shared` additions list their callers in the change description                                                                                                    |
| No simulator-owned validation, composition, truth, scoring, or verification in TypeScript           | preserve parity registry and real-binary tests; review changes to model rules against the ownership list                                                                                                |
| Source limits remain 300 lines, functions/components 40 lines, complexity 10/15                     | existing ESLint gates; never raise thresholds to complete migration                                                                                                                                     |
| No token clone of 75+ and no dead exports                                                           | existing `jscpd` and `knip` gates                                                                                                                                                                       |
| No fixed sleeps in unit, component, or end-to-end tests                                             | existing Playwright lint plus an ESLint/source rule covering `setTimeout` in tests except fake-clock tests; use deferred drivers                                                                        |
| Coverage floors remain a backstop, not a test-retention rule                                        | keep current model/app thresholds; use the test ledger below for deletion decisions                                                                                                                     |
| Graph geometry and browser rendering remain independent oracles                                     | keep pure routing/layout tests and real Chrome fit/overflow/contrast checks; never compute expected test values with the production helper under test                                                   |
| New module has a stated responsibility, public API, callers, and permitted dependencies             | add a checked `canvas/ARCHITECTURE.md` module table or machine-readable module map consumed by an architecture test                                                                                     |
| A feature never imports or composes another feature's CSS                                           | CSS architecture check resolves TypeScript `.module.css` imports and CSS `composes` paths against the module register                                                                                   |
| Shared visual behavior is exposed through a primitive or module facade, not a shared class bag      | architecture check permits cross-module style imports only for theme entrypoints; review requires a named owner and at least two callers for a new primitive                                            |
| Raw colors and global custom properties live in the theme token contract                            | CSS source check rejects color literals outside token files and a small registered list of rendering exceptions such as masks                                                                           |
| Global selectors and `!important` are confined to owned third-party integration files               | CSS source check rejects `:global` and `!important` elsewhere; existing exceptions are registered as migration debt and the register may only shrink                                                    |
| Inline styles carry runtime geometry or CSS custom-property values only                             | TypeScript-AST check rejects static color, spacing, typography, and layout declarations in JSX `style`; review verifies each dynamic value comes from runtime geometry                                  |
| New CSS modules stay within the repository file-size bar                                            | new files are at most 300 lines; existing oversized style sheets receive ratcheted caps that fall as responsibilities are split and never rise                                                          |

### Dependency enforcement additions

Phase 0 adds the module register and architecture checker before the first move. The register describes the current tree and exact migration debt; a module entry is added atomically when its target folder exists. The target dependency contract above guides those additions but is not predeclared as nonexistent paths.

Rules should reject:

- `src/ui/editor/** -> src/ui/plant/**`
- `src/ui/review/** -> src/ui/plant/**`
- cross-module imports of any feature file other than that module's `index.ts`; private access is forbidden even when no `internal/` folder exists
- any import of `<other-module>/internal/**`, including type-only imports
- wildcard facade exports, concrete controller exports, and public signatures containing internal types; explicitly named public UI/model values may resolve to their own module's `internal/` implementation
- `src/app/** -> react`, `react-dom`, `@xyflow/react`, browser adapters, or UI
- `src/ui/** -> src/adapters/**`
- any circular dependency

Public entrypoint imports remain allowed, for example `ui/editor -> ui/graph` and `app/review -> app/ports`.

The current dependency-cruiser run excludes test files, so it cannot prove the test boundary by itself. `npm run architecture` adds a test-inclusive import scan that applies these rules to production and tests. Same-module private tests may import their own `internal/` files; public contract tests and every cross-module test must import the module entrypoint. Positive and negative fixtures cover same-module private access, external public access, type-only imports, and rejected cross-module private access.

### Constraint activation during migration

Existing TypeScript, dependency, duplication, unused-export, coverage, and browser gates remain active throughout. New gates are introduced without pretending the unmigrated tree already has the target shape:

1. Phase 0 records each current CSS and module-boundary violation as an exact file or import edge, adds architecture fixtures, and makes new violations fail. The debt register may only shrink.
2. Phase 1 creates the graph module, activates its facade/private-import rules, and removes its plant-owned style and import debt.
3. Phase 2 activates route-lifetime and no-direct-app-clock/timer rules after factories, route sessions, Clock, and Scheduler are migrated.
4. Phase 3 activates typed-gateway, typed-feature-failure, raw-`unknown`, action-name, and protocol-error rules.
5. Phase 4 activates the editor revision/transaction invariants through behavior tests rather than a static allow-list.
6. Phase 5 activates each remaining feature's facade, CSS ownership, and file-cap rules as that feature moves; `shared.module.css` and `panel.module.css` debt reaches zero before exit.
7. Phase 6 changes no architecture permission. It removes tests only through the evidence ledger.

An unmigrated file can remain only through its Phase 0 exact debt entry. The implementation may not add a broader pattern, new path, higher cap, or new exception to accommodate later work.

## 7. Test architecture

The goal is fewer redundant tests and stronger unique failure detection. Test count and line coverage alone do not decide value.

### 7.1 Test layers

| Layer              | Proves                                                                              | Typical doubles                                         |
| ------------------ | ----------------------------------------------------------------------------------- | ------------------------------------------------------- |
| model unit         | pure rules, boundaries, geometry, parsing outcomes                                  | values only                                             |
| app unit           | state transitions, operation ordering, failure precedence, stale-result rejection   | typed gateways, deferred operations, retained callbacks |
| adapter contract   | request encoding, response decoding, malformed payloads, SSE lifecycle              | controlled transport; shared protocol fixtures          |
| component          | accessible rendering and user intent for a component                                | real app store or narrow app facade                     |
| screen integration | one feature's main state-to-UI path                                                 | typed app module, minimal surrounding shell             |
| Playwright         | critical user journeys, real binary integration, browser layout/input/accessibility | real binary and real files; no binary mock              |

### 7.2 Tests to retain

- Deterministic layout, routing, pin, grid, and viewport tests. They diagnose pure geometry.
- Chrome fit, overflow, text-clash, contrast, keyboard, and axe tests. Browser layout is a different oracle.
- Distinct verification and integrity cases in `review.spec.ts` and `review-integrity.spec.ts`. They represent different binary outcomes and trust claims.
- App tests for state-machine transitions, error precedence, unsaved navigation, editor history, launch stop/failure, and run completion.
- Adapter tests for error parsing, credential handling, SSE close/failure behavior, and malformed responses.
- Tests such as `showBuild.test.ts` that prove a real cross-store product operation. Move the operation to app ownership; retain its behavioral assertion.

### 7.3 Tests to improve immediately

1. Replace the fixed delay in the PlantStore overlapping-load test with two deferred reads resolved in the required order.
2. Repair the ReviewStore stale-callback test so it invokes a retained old callback after a new run opens.
3. Add A -> B -> A generation tests; matching resource ids must not admit an old completion.
4. Add WorkspaceStore close-while-reload-pending and overlapping-reload tests.
5. Add EditorStore save-then-route-switch, edit-during-write, edit-during-compose, and conflict tests.
6. Add malformed-success adapter tests so required responses cannot decode as successful empty values.
7. Replace repeated microtask flushing with explicit deferred completion.

### 7.4 Test retirement rule

Before deleting or merging a test, record four fields:

| Field          | Question                                                      |
| -------------- | ------------------------------------------------------------- |
| outcome        | Which user-visible or architectural behavior does it protect? |
| owner          | What is the lowest layer that owns that behavior?             |
| unique failure | What regression fails this test but no retained test?         |
| replacement    | Which retained test proves the same outcome more directly?    |

Delete a test only when it has no unique failure or its replacement is already green. Never delete a test merely because it is small, because coverage is high, or because an end-to-end test touches the same screen.

Initial candidates for evaluation, not immediate deletion:

- DOM-structure assertions in `MetaParts.test.tsx`: replace them only after a browser assertion proves the no-orphan-separator wrapping outcome.
- component cases that repeat a model rule and assert no distinct rendering behavior: move the rule assertion to model and retain one component wiring case.
- large multi-assertion Playwright tests: split by failure domain where diagnosis is poor, but share setup rather than collapsing distinct outcomes.

## 8. Migration plan

Each phase is independently reviewable. Structural movement and behavior correction are separate commits so evidence stays attributable.

### Phase 0: close lifecycle safety gaps

- Record current dependency graph and behavior gates.
- Add the machine-readable module register, test-inclusive architecture scan, CSS architecture checker, and their positive/negative fixtures.
- Register current violations as exact, ratcheted migration debt; do not add target folders or broad exceptions before they exist.
- Add deferred test helpers.
- Add each missing async ownership regression, observe it fail for the intended reason, and correct it in the same integration unit.
- Cover stale review callbacks, plant load ordering, workspace close/reload and watcher installation, and editor open ordering.
- Keep red/green development inside this phase; do not hand off or merge a phase with failing tests.
- Do not move feature or graph files yet.

**Exit:** lifecycle risks have deterministic tests and direct guards; architecture checks reject new production, test, and CSS boundary violations; the exact debt register is baselined and may only shrink; all focused and full required tests are green; existing product behavior remains pinned.

### Phase 1: extract the graph module

- Declare module responsibility and callers.
- Move shared graph UI/model files without behavior changes.
- Replace plant-owned styles with graph-owned styles where shared.
- Register the graph module and activate its facade, private-import, vendor-style, and CSS ownership rules.
- Remove the graph-related entries from the migration debt register.

**Exit:** plant, editor, and review import only the public graph API; visual and geometry suites remain green.

### Phase 2: consolidate operation identity

- Introduce the minimal generation/session primitive.
- Replace the direct guards from Phase 0 where the shared primitive makes ownership clearer.
- Apply it consistently to review streams and loads, workspace reload, plant load, and editor open without changing behavior.
- Introduce plant/editor/review module factories and route sessions with Strict Mode-safe `open`/`close` ownership.
- Move the unsaved-navigation guard from the composition root to the editor route and require its registration to return a disposer.
- Introduce the Clock and Scheduler ports and migrate launch/review application timers before enabling the no-direct-app-clock/timer rule.

**Exit:** A -> B, A -> B -> A, close-while-pending, close/open replay, repeated close, Strict Mode cleanup, unsaved-guard disposal, and disposed callback tests pass without sleeps; app code uses no direct clocks or timers.

### Phase 3: type the bridge boundary

- Introduce typed feature gateways.
- Move response decoding into the bridge adapter.
- Translate protocol and transport failures into closed feature-owned failure unions while preserving diagnostics.
- Add malformed-payload contract tests.
- Remove raw action calls from stores.

**Exit:** no application public API returns `unknown` or exposes `BridgeFailure`; action names, HTTP status, and protocol codes live in the adapter; mapping and real-binary parity tests remain green.

### Phase 4: make editor saving transactional

- Add each editor save/navigation/concurrent-edit regression, observe the intended failure, and correct it in the same phase.
- Add draft revision identity.
- Capture immutable save attempts.
- Define conflict and concurrent-edit state transitions.
- Keep the current document dirty when later edits are not part of the completed save.

**Exit:** save/navigation/concurrent-edit and partial-success tests pass; documented guarantees match the current compose protocol; existing compose behavior remains byte-compatible.

### Phase 5: finish feature module boundaries

- Migrate workspace, plant, editor, launch, and review one at a time.
- Expose one public interface and factory per application module, plus the smallest required public model and UI APIs.
- Move concrete controllers, transitions, workflow helpers, caches, and feature-private components under the owning `internal/` tree.
- Activate each migrated feature's facade and private-import checks and remove its exact debt entries.
- Move cross-store orchestration from UI into named app use cases.
- Adopt shared JSON readers only in parsers touched by the migration.
- Extract only proven UI primitives and dissolve `shared.module.css` and `panel.module.css` as their final responsibilities move.

**Exit:** architecture rules prevent private cross-feature imports in production and tests; CSS ownership debt is zero; no circular modules or gate weakening.

### Phase 6: test audit

- Build the outcome/owner/unique-failure/replacement ledger.
- Remove only proven duplicates.
- Simplify test fixtures so general fakes carry no server policy.
- Measure suite time before and after; do not trade away failure coverage for a smaller count.

**Exit:** every remaining test has a stated purpose at its lowest owning layer; full canvas and phase gates pass.

## 9. Validation for every implementation phase

Run focused tests while working, then:

```text
npm run format:check
npm run typecheck
npm run lint
npm run deps
npm run architecture
npm run dup
npm run knip
npm run test
npm run build
npm run e2e
```

Before integration, run repository `make phase-gate` with an Enola baseline pinned before structural edits. Review `git diff main -- test/architecture scripts` and justify every gate change. No threshold, allow-list, coverage floor, or test rule may be weakened to finish the migration.

Chrome review remains required because module movement can change CSS ownership, import order, layout, focus, and interaction even when TypeScript and component tests pass.

## 10. Supporting decisions and later improvements

The first four items are required parts of the architecture and have explicit migration phases. The last two are conditional follow-ups.

| Item | Status      | Owning phase                                                                 |
| ---- | ----------- | ---------------------------------------------------------------------------- |
| F1   | required    | Phase 2 establishes route-session lifetimes                                  |
| F2   | required    | Phase 3 translates protocol failures while typing gateways                   |
| F3   | required    | Phase 0 establishes the module register and checker before any module moves  |
| F4   | required    | Phase 0 establishes checks; Phases 1 and 5 migrate graph and feature styles  |
| F5   | conditional | a separate contract-generation design if repository contract policy needs it |
| F6   | conditional | only after profiling demonstrates a missed performance budget                |

### F1. Define singleton and route-session lifetimes

The current composition root constructs one instance of every store. UI effects call `open()` and `close()` on workspace and review stores, while one `PlantStore` is shared by plant viewing and run review. The lifetimes work through convention rather than an explicit ownership model.

Declare the lifetime of every application object:

| Lifetime                   | Modules or objects                                                   | Rule                                                                                                                              |
| -------------------------- | -------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| application singleton      | session, navigation, theme, panels, workspace catalog, active launch | created once by the composition root; explicitly disposed at application shutdown when it owns a resource                         |
| application module factory | editor and review module factories                                   | created once; creates route sessions and holds only stable dependencies, never an active document, run, subscription, or route id |
| route session              | plant view, editor document, run review                              | one instance per route; `open`/`close` is reusable and Strict Mode-safe; close releases all route resources and invalidates work  |
| component local            | form input, disclosure, focus, temporary graph interaction           | owned by React and never stored in an application singleton                                                                       |

Every application singleton that permanently acquires an external resource exposes idempotent `dispose(): void`. A route facade exposes idempotent, reusable `open`/`close`; closing releases every acquired resource. If a route facade exceptionally owns something outside that interval, it also exposes `dispose()` and proves safe recreation under Strict Mode. Do not add a general scope container; explicit factories and React cleanup are sufficient.

**Enforcement:** the module manifest records singleton, factory, or route lifetime; public contract tests prove close/open replay, repeated close, disposal where applicable, and late-completion rejection; architecture review rejects singleton modules holding route-specific subscriptions or editor drafts.

### F2. Stop leaking protocol failures into features

`BridgeFailure` currently appears throughout app and UI code, so feature behavior depends directly on transport status, protocol codes, and issue documents.

Typed gateways should translate the protocol into feature-owned failure unions while retaining diagnostics:

```ts
type WorkspaceFailure =
  | { readonly kind: "unpaired"; readonly diagnostic: Diagnostic }
  | { readonly kind: "unreachable"; readonly diagnostic: Diagnostic }
  | { readonly kind: "unreadable"; readonly diagnostic: Diagnostic };
```

The adapter owns HTTP status and wire decoding. The application owns recovery meaning. The UI renders `kind` and may show the preserved diagnostic message in a developer detail. Raw protocol codes remain available for support without controlling feature behavior.

**Enforcement:** no `BridgeFailure`, HTTP status, or bridge error-code comparison outside adapters after the typed-gateway phase; contract tests cover every mapping and preserve unknown codes as an explicit `unexpected` case.

### F3. Add a machine-readable canvas module register

The Go architecture gates use a package table. Give canvas modules the equivalent source of truth, for example `canvas/architecture.json`:

```json
{
  "modules": {
    "app/editor": {
      "kind": "application",
      "entrypoint": "src/app/editor/index.ts",
      "lifetime": "factory-and-route-session",
      "mayImport": [
        "app/shared",
        "app/ports",
        "model/editor",
        "model/plant",
        "model/graph"
      ]
    }
  }
}
```

A small TypeScript-based checker validates that every module is registered, every entrypoint exists, imports match the allow-list, external imports use entrypoints, and `internal/` paths do not leak. Generate dependency-cruiser rules from this register or check both against it so two architecture lists cannot drift.

**Limit:** the register records ownership and permitted dependency direction. It is not a plugin registry or runtime module loader.

### F4. Make CSS ownership follow module ownership

Editor and run code currently import `PlantScreen.module.css`, and shell components share other feature style sheets. Moving TypeScript without moving CSS would leave the same coupling hidden in presentation files.

CSS needs an architecture, but adopting a second framework would add vocabulary without fixing ownership. Keep CSS Modules and the existing semantic custom properties. Organize them by the same module boundaries as TypeScript.

#### Architecture choice

| Approach                | Decision                                                                                                                          |
| ----------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| CSS Modules             | keep; local class scoping matches component and module ownership                                                                  |
| semantic design tokens  | keep and tighten; this is the public style contract for color, spacing, type, motion, radii, dimensions, and status               |
| BEM                     | do not add; CSS Modules already provide name isolation, so long encoded class names would duplicate the module boundary           |
| ITCSS or SMACSS         | borrow the separation of tokens, base, vendor integration, primitives, and feature styles; do not impose their full naming system |
| utility-first framework | do not add; it would create a second styling API and does not solve the present cross-feature ownership problem                   |
| CSS-in-JS               | do not add; runtime styling and another dependency have no demonstrated benefit here                                              |
| cascade layers          | defer; first remove cross-module imports and global leakage, then evaluate layers only if measured cascade-order defects remain   |

CSS Modules are the scoping mechanism. The architecture is the ownership model around them.

#### Target structure

```text
src/ui/
  theme/
    tokens.css                 # global semantic custom properties and light/dark values
    base.css                   # reset, element defaults, focus and accessibility base
  primitives/
    button/
      index.ts                 # public Button interface
      internal/
        Button.tsx
        Button.module.css
    panel/
      index.ts
      internal/
        Panel.tsx
        Panel.module.css
    disclosure/
    metadata/
    scroll-region/
  graph/
    index.ts                   # public graph UI interface
    internal/
      GraphRoot.tsx
      graph.module.css
      nodes.module.css
      vendor.module.css        # React Flow selectors scoped below the graph root
  editor/
    index.ts
    internal/
      EditorScreen.tsx
      EditorScreen.module.css
  review/
    index.ts
    internal/
      ReviewScreen.tsx
      ReviewScreen.module.css
```

`tokens.css` and `base.css` are the only application-global style entrypoints. `GraphRoot` imports the React Flow base CSS followed by its owned vendor overrides. Those overrides are scoped beneath the graph root wherever the third-party selector permits it. Feature entrypoints do not export class names or stylesheet objects.

The current `shared.module.css` mixes breadcrumbs, section headings, buttons, panel layout, metadata, disclosure, masks, and scrolling. The current `panel.module.css` mixes container layout, fields, messages, and three button variants. Replace each stable, semantic behavior with a small primitive that owns markup, accessibility, states, and styles together. Keep one-off layout in the feature that uses it. A few repeated declarations are cheaper than a false shared abstraction.

#### Style ownership rules

1. A feature imports only its own CSS Modules and public TypeScript facades from graph or primitives.
2. CSS `composes` stays within one module. Cross-module reuse goes through a component interface; no feature imports a primitive's internal stylesheet.
3. Theme tokens are the only cross-module CSS API. A feature may define private custom properties on its root for runtime geometry, but it must not redefine global semantic tokens.
4. Shell owns application chrome and panel placement. Graph owns canvas, nodes, ports, wires, zoom tiers, and React Flow integration. Features own their screen-specific content and arrangement.
5. State is expressed with native attributes such as `aria-disabled`, `aria-selected`, and `data-state`. Do not create a parallel JavaScript-to-CSS state vocabulary when an existing semantic attribute works.
6. `:global` and `!important` are integration escape hatches. They belong only in a registered vendor-integration stylesheet and must be tied to a third-party selector that cannot be controlled locally.
7. JSX inline styles are allowed for calculated coordinates, dimensions, fitted font size, and CSS custom-property injection. Static colors, spacing, typography, and layout belong in the owning stylesheet.
8. Raw color literals live in `tokens.css`. A rendering literal used for a mask or alpha calculation needs a named, checked exception when no semantic color is involved.

The existing inline styles for timeline positions, scrubber progress, graph zoom variables, outline depth, and fitted node text satisfy rule 7 because their values are runtime geometry. The broad `:global` and `!important` usage in plant node styles must move with the graph module and be reduced to the selectors required by React Flow.

#### Naming rules

Names describe the element's role in its owning component. They do not describe a current color, spacing value, DOM position, or implementation technique.

| Construct                             | Rule                                                                                                 | Examples                                                                                          |
| ------------------------------------- | ---------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------- |
| CSS Module file                       | match its component or named responsibility                                                          | `Button.module.css`, `ReviewTimeline.module.css`, `GraphPorts.module.css`                         |
| component root class                  | use `root`; module scoping supplies the component name                                               | `styles.root`                                                                                     |
| local structural class                | use concise camelCase role names                                                                     | `toolbar`, `panelHeader`, `fieldLabel`, `emptyState`, `eventLane`                                 |
| visual or interaction state           | prefer an existing `aria-*` attribute; otherwise use a typed `data-*` attribute                      | `[aria-selected="true"]`, `[data-state="failed"]`, `[data-zoom-tier="overview"]`                  |
| variant exposed by a public primitive | name the product meaning in the TypeScript prop and map it privately to an attribute                 | `tone="danger"` becomes `[data-tone="danger"]`                                                    |
| global design token                   | use kebab case with category, semantic role, and optional state                                      | `--surface-panel`, `--text-secondary`, `--status-warn-bg`, `--motion-fast`                        |
| private component property            | prefix it with the owning component or module                                                        | `--graph-zoom`, `--graph-text-floor`, `--scrubber-played`                                         |
| third-party global selector           | preserve the vendor's name and keep it in the owned vendor stylesheet                                | `.react-flow__edge` inside `graph/internal/vendor.module.css`                                     |
| test selector                         | use accessible role/name first and a dedicated `data-testid` only when no user-facing selector works | never locate an element by `styles.root`, a generated class name, or a visual class such as `red` |

Avoid names such as `blueButton`, `leftThing`, `margin8`, `bigText`, `row2`, and `flexContainer`. They encode a replaceable presentation choice. Avoid names such as `common`, `shared`, `misc`, and `utils`; they conceal ownership. A component may use `leftPanel` when left placement is part of the stable shell contract, but feature content should use its purpose, such as `outlinePanel`.

Do not use BEM suffixes inside CSS Modules. `Button.module.css` with `.root`, `.icon`, and `[data-tone="danger"]` already has an explicit scope and state contract. Do not construct class names dynamically from unchecked strings. Map a closed TypeScript union to CSS Module exports or typed `data-*` values.

Descendant selectors are allowed when the owner intentionally controls that semantic subtree, such as `.field input`. Positional selectors such as `:first-child` and `:nth-child()` must not coordinate independently owned components. Give shell slots a class or `data-slot` name instead.

The CSS architecture checker enforces module filenames, token and private-property naming, forbidden generic filenames, and closed locations for global selectors. Selector semantics remain a review concern until they justify a real CSS parser; do not add a fragile regular-expression linter for subjective names.

#### File boundaries and size

Several current style sheets combine multiple responsibilities: `plant/nodes.module.css` is 695 lines, `run/RunScreen.module.css` is 608, `plant/PlantScreen.module.css` is 453, `home/HomeScreen.module.css` is 434, and `editor/Editor.module.css` is 330. Split them by named owner, not by an arbitrary line interval. Examples are graph node shell, graph port integration, review timeline, review report panel, and editor toolbar.

New CSS files follow the existing 300-line source limit. Record each existing oversized file with its current cap, then ratchet the cap down as a module migrates. Never raise a cap or split selectors into meaningless fragments to satisfy it.

#### Enforcement

Add `scripts/check-css-architecture.mjs` in Phase 0 before the first CSS migration. It uses the same module register as the TypeScript boundary check and fails on:

- feature-to-feature stylesheet imports;
- cross-module CSS `composes`;
- `:global` or `!important` outside registered vendor integration;
- raw color literals outside tokens and registered rendering exceptions;
- new or enlarged CSS files above their allowed cap;
- global custom-property declarations outside the theme contract.

Use the TypeScript compiler API for the small JSX inline-style rule rather than adding a CSS framework. Run both checks from `npm run deps` or a dedicated `npm run architecture`, which remains part of `make canvas-check`. Prettier continues to own formatting. A style linter is justified later only if the source checker grows into a general CSS parser.

#### Tests and migration

Do not snapshot class names or CSS text. Test primitive behavior and accessible state in component tests. Keep visual layout, theme, focus, overflow, contrast, zoom tiers, and third-party integration in real Chrome because those outcomes depend on the cascade and rendering engine.

Migrate styles with their owning module:

1. extract graph styles and scope React Flow overrides during Phase 1;
2. introduce a primitive only when two current callers share its semantics and behavior;
3. replace editor and run imports of `PlantScreen.module.css` through graph or shell public interfaces;
4. dissolve `shared.module.css` and `panel.module.css` as their last owned responsibilities move;
5. run the complete light/dark desktop Chrome matrix after each style-owning module moves.

The result gives CSS the same property as the target TypeScript modules: callers use a stable interface while implementation selectors and files remain private.

### F5. Reconcile generated contract types with runtime decoding

The quality bar says shared schema types are generated, but the current canvas has no `src/generated/` tree while repository contracts exist for composition, domain, journal, report, and other documents.

Use contract generation only for transport data-transfer types. Keep application states and model concepts handwritten. Generated TypeScript types do not validate runtime JSON, so adapters still need closed runtime decoders and shared valid/invalid fixtures.

The generation command must be deterministic, use the pinned toolchain, write only `src/generated/`, and have a check mode that fails on drift. Do not make UI or app modules depend directly on large generated protocol documents; adapters translate them to the smaller public module contracts.

### F6. Add performance abstractions only after measurement

Several screens subscribe to whole store snapshots, and review playback can publish frequently. This is a reason to profile, not evidence of a performance defect.

Measure render counts and frame time on the factory and 2,032-tag campus acceptance cases. Introduce selector subscriptions, indexed projections, or batching only where traces show a missed budget. Keep selectors inside the owning model/app module and preserve one immutable state publication contract.

No caching, memoization framework, virtualized graph, or alternate state library is justified by this source review alone.

## 11. Explicit non-goals

- no Redux, MobX, event bus, command bus, dependency-injection framework, repository layer, or generic service layer
- no micro-frontends or separately deployed modules
- no plugin system
- no TypeScript reimplementation of simulator validation or truth
- no universal schema/decoder framework
- no speculative selector framework or performance refactor without profiling
- no collapsing of independent test oracles into production helpers
- no broad file moves without one-module scope, focused tests, and a reviewed boundary

## 12. Accepted decisions

The owner accepted this document as the target Canvas architecture on 2026-10-16:

1. Use the hybrid structure: technical layers plus vertical product modules.
2. Use a facade-plus-`internal/` boundary for each TypeScript module, enforced by architecture gates.
3. Use `graph` as a shared streamsim-specific module for plant, editor, and review.
4. Use typed feature gateways while retaining one private generic HTTP mechanism in the adapter.
5. A completed editor save marks only its captured revision saved; later edits remain dirty.
6. Require the test retirement ledger before deleting tests.
7. Use CSS Modules plus semantic tokens, private module styles, and public UI primitives; do not add a second styling framework.
8. Use readonly discriminated state, typed feature failures, named behavior options, and explicit module entrypoints as the TypeScript code contract.
9. Use manual composition, external stores, route controllers, pure screens, explicit effects, and restrained Context as the React pattern set.

Implementation starts with Phase 0 only: deterministic lifecycle regression tests, controlled test drivers, and architecture baselines/checkers. It must not mix module movement with behavior correction.

## 13. Review record

A senior frontend architecture review agreed with the central direction:

- preserve the current layers, `Observable`, and `useSyncExternalStore`;
- extract the shared graph capability;
- type bridge operations and decode at the boundary;
- make async operation identity consistent;
- keep pure geometry and real-browser rendering as independent test oracles;
- avoid a state-library migration, generic service/repository layers, and speculative performance abstractions.

The review also identified the stale-callback test gap, fixed-delay overlap test, policy-bearing `FakeBridge`, and editor/workspace concurrency cases included above. The review was source-based; it did not run browser simulations.

A final independent Astra frontend architecture review found no P1 blocker. It identified and this revision corrected five consistency issues:

- facades now permit explicitly named public UI/model values without exposing private import paths or concrete controllers;
- the composition root now owns module factories while route controllers own reusable, Strict Mode-safe route sessions;
- the dependency contract now includes foundations, adapters, session/shell ownership, and concrete coordinator edges;
- every required constraint now has a migration phase and exact ratcheted debt handling;
- private-import enforcement now includes tests and type-only imports through a test-inclusive scan and architecture fixtures.

The Astra review was read-only and source-based. It did not run tests or Chrome. The root documentation review ran the documentation and formatting gates recorded in the task handoff.
