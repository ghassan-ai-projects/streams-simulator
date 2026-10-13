# Canvas UI — end goal, acceptance and review checklist

- **Date**: 2026-10-13
- **Owner direction**: desktop first (Chrome, 1280–1920 px wide); no mobile at this stage. First make the canvas work
  completely, at production level; then fix every detail. Review every round in Chrome as a director of design; commit each
  round.
- **Applies with**: [CANVAS_QUALITY_BAR.md](CANVAS_QUALITY_BAR.md) (gates), [CANVAS_PLAN.md](CANVAS_PLAN.md) (phases),
  the canvas design and technology records under `docs/simulation-canvas/`.

## 1. End goal

One person, at a desktop browser, with `streamsim canvas` running against a workspace, can do all of this without a terminal
and without the canvas computing anything the binary owns:

| # | Capability | What it means in the UI |
|---|---|---|
| G1 | **Open a workspace** | Domains, compositions, scenarios and runs of a workspace folder listed and searchable; a file changed on disk (by a person or a coding agent) reloads in place |
| G2 | **See a plant** | Plant view of any domain: instances as nodes grouped by area, typed links as edges, readable at the size of the 2,032-tag campus; select an instance to see its kind: states, sensors, actions, faults, KPIs, ports |
| G3 | **Build a plant from the library** | Browse the component library (equipment, instruments, presets, parameters, ports); place equipment, name instances, pick presets and parameters, connect ports (type-checked by the binary), set areas and site; compose, validate and lint through the binary with every issue pinned to its element; save the composition and the composed domain |
| G4 | **Run and control** | Start a run of a domain in Explore or Evaluate mode with a seed; play at a chosen speed, pause, step, reset, end; live values on the plant, live charts for chosen channels, a timeline of observations, commands, faults and perturbations |
| G5 | **Disturb** | Inject and clear faults, environment events and perturbations mid-run from a panel listing exactly what the domain declares; every disturbance lands in the command log; export the session as a scenario file |
| G6 | **Review** | Run list with a Verified / Not verified / Incomplete badge from `streamsim verify`; playback of any run folder with seek; the report: counts, timeline, faults, effector outcomes, KPI truth and alarm KPIs, true vs reported after sealing |
| G7 | **Truth stays sealed** | In Evaluate mode no truth-bearing value reaches the page before the run is sealed; Explore runs are stamped unblinded and say so on every screen |

## 2. Acceptance tests (Playwright, real `streamsim` binary, real files)

| # | Test | Passes when |
|---|---|---|
| A1 | **Build the factory in the UI** | From an empty composition, through the UI only: site and areas, ≥ 18 equipment entries (lead and lag compressors with their bands, receiver, dryer, chiller, tower, pumps, four lines, an air handler, sub-meters and a main meter, transformer, generator, weather, calendar) and ≥ 40 connections including the sampled feedback loops; compose, validate and lint clean; the saved composition, formatted, is byte-identical to `examples/compositions/factory-utilities.composition.json` and the composed domain to `examples/factory-utilities.domain.json` |
| A2 | **Build a second plant with an error and fix it** | A cooling plant built in the UI with a deliberate type mismatch in one connection shows the binary's issue on that connection; fixing it composes clean |
| A3 | **Run, disturb, review** | The factory runs live in Explore mode; a cooling tower fan failure injected mid-run shows on the tower, the chiller and the cooling meter in the live view; the run ends; its badge is Verified; the report shows the fault, the cascade's KPIs and the alarm KPIs |
| A4 | **Evaluate stays blind** | An Evaluate run with an injected fault: no true state, label or fault command reaches any bridge payload or the page until the run is sealed (canary string); after sealing the report shows truth |
| A5 | **Play back without the binary** | A recorded run folder plays back with seek and the report renders with `streamsim` absent from `PATH`; a crashed-run fixture shows Incomplete, an edited folder shows Not verified |
| A6 | **Scale** | The 2,032-tag campus opens in the plant view and runs live within the budgets of §4 |
| A7 | **Agent parity** | Every action the bridge exposes maps to a documented CLI command or director MCP tool (registry test); a file edited on disk while open reloads |

## 3. Review checklist (every round)

Each round ends with a review in Chrome at 1440 × 900 and 1920 × 1080, light and dark, recorded in the phase review. A round
is done only when every item holds.

### 3.1 Function

- [ ] Every control does what its label says, against the real binary; nothing is a placeholder or a mock
- [ ] Every async action has loading, success and failure states; a failed command shows the binary's message and code
- [ ] Empty states say what to do next; no blank panels
- [ ] Validation issues from the binary are pinned to the element they name (JSON pointer) and listed in one place
- [ ] Reload on file change works without losing the user's selection or unsaved edits (conflict shown, never overwritten)
- [ ] No console error or warning; no failed network request in normal use
- [ ] The truth boundary holds: Evaluate pages and payloads carry no sealed value before sealing

### 3.2 Design (director of design)

- [ ] **Hierarchy**: one primary action per screen; the eye lands on the plant or the run, not the chrome
- [ ] **Layout**: a stable shell (top bar, left navigation, main canvas, right inspector, bottom timeline) that never jumps;
      panels resize; nothing overlaps or clips at 1280 px
- [ ] **Spacing**: a 4 px grid; consistent paddings per component type; aligned edges
- [ ] **Typography**: one UI typeface and one monospace for values and ids; a clear type scale (no more than 6 sizes);
      numbers tabular; units always shown, in a lighter tone
- [ ] **Color**: semantic tokens only (surface, text, accent, ok, warn, alarm, info, truth); WCAG AA contrast in both themes;
      color never the only signal (icons or text beside it); quality and alarm states consistent everywhere
- [ ] **Data density**: an engineering tool: dense but calm; tables aligned on decimals; no decorative noise
- [ ] **States**: hover, focus, active, selected, disabled and read-only visibly distinct; focus ring always visible
- [ ] **Consistency**: the same thing looks the same everywhere (an instance, a fault, a channel, a badge)
- [ ] **Copy**: sentence case; plain words; verbs on buttons; no jargon the domain file does not use; errors say what
      happened and what to do
- [ ] **Motion**: purposeful and short (≤ 150 ms); respects reduced motion; live updates never flash the layout
- [ ] **Charts**: labeled axes with units, readable ticks, legends that match line colors, no chart junk
- [ ] **Icons**: one icon set, one stroke weight, each with a text label or tooltip

### 3.3 Accessibility and input

- [ ] Every screen passes axe with no violation
- [ ] Every action is reachable by keyboard in a sensible order; shortcuts listed in a help overlay
- [ ] Graph nodes and edges have accessible names; the inspector is a labeled region

### 3.4 Performance (desktop)

- [ ] Budgets in §4 hold and are measured by the tests, not estimated

## 4. Budgets (desktop, development machine; enforced in CI from P9)

| Budget | Value |
|---|---|
| First render of the shell | ≤ 1.5 s |
| Plant view of the factory (82 tags) interactive | ≤ 1 s after the domain loads |
| Plant view of the campus (2,032 tags, ~600 instances) interactive | ≤ 3 s; pan and zoom at ≥ 50 fps |
| Live update | keeps up with 2,000 observations per second without dropping a frame below 30 fps |
| Bundle (gzip) | ≤ 600 kB initial |

## 5. Order of work

1. **Function first** (P8a–P11): every capability G1–G7 working end to end with its acceptance test, at production quality
   (gates of the quality bar green), with a sound but plain visual design.
2. **Then every detail**: rounds of design review against §3.2 until every item holds in both themes at both widths, each
   round committed with its screenshots described in the review record.
