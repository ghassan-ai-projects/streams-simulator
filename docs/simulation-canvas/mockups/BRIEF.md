# Canvas mockup brief — constraints and acceptance criteria

The mockups in this folder are judged against this brief, nothing else. A reviewer cites the criterion id for every finding.
Approval rule: §6.

## 1. Who and what for

| User | Wants to |
|---|---|
| Process engineer | Build a plant from the component library, wire it correctly, and see it compile clean |
| Test engineer | Run a plant, disturb it at the right moment, and watch the consequences spread |
| Reviewer | Judge a finished run: was it verified, what happened, what did the KPIs and alarms do |

Desktop only: 1440 × 900 (primary) and 1920 × 1080; the layout is fluid and only the graph grows. Dark is the hero theme; light must work equally.

## 2. Style reference

- **Unreal Engine Blueprints** for the graph: nodes with category-coloured headers, typed pins on both sides, curved wires
  coloured by type, comment frames, collapsed macro nodes, minimap, details panel, searchable palette, compile button with
  status, compiler results.
- **ISA-101 restraint** for status: normal operation is calm; colour that means "look here" is reserved for abnormal states.

## 3. Hard constraints (measurable; a violation is a High finding)

| Id | Constraint |
|---|---|
| C1 | Text ≥ 11 px; body text 12–13 px; at most two type families (UI sans, monospace) and three weights. Zoom never scales text: semantic zoom switches the node tier (full, compact) instead |
| C2 | Contrast: text ≥ 4.5:1 against its background; pins, wires (dimmed ones included), icons and focus rings ≥ 3:1 against the canvas, in both themes. Measured: `node tools/contrast.mjs bp.css tools/pairs.json` passes |
| C3 | Colour is never the only signal: every status and every port type also has a label, icon or shape |
| C4 | Reserved colours: red only for alarms and active faults, amber only for warnings and stopped equipment, purple only for sealed truth and Explore mode, the accent blue only for selection, focus and the primary action |
| C5 | At most 9 port-type colours, named in a legend and in every pin tooltip, with a pin shape per family (○ flow, □ energy, ◇ state and signals). Measured: `node tools/deltae.mjs bp.css 11 13` passes (CIEDE2000 ≥ 11 between port colours, ≥ 13 from the reserved colours, both themes) |
| C6 | Numbers: monospace, tabular figures, fixed decimals per channel, unit always shown in a lighter tone |
| C7 | No truncated instance id; component names may truncate with an ellipsis and a tooltip |
| C8 | 4 px spacing grid; one node width per zoom level; pin rows on a fixed pitch |
| C9 | Honesty: every control and every value shown maps to something the binary can do or report today (CANVAS_BRIDGE.md action registry, director tools, CLI commands). Anything else is drawn as clearly marked "not available yet" or left out |
| C10 | The canvas never acts as the plant host: no control that invokes an equipment effector |
| C11 | Truth boundary: in an Evaluate run, no true state, director-side fault record or sealed value is shown before the run ends; the canvas may list the commands it sent itself, labelled as sent. Explore shows truth, in purple, and says so on screen |
| C12 | Selection, hover, focus, disabled and read-only states are visibly distinct for nodes, pins, wires and controls (specimens: `states.html`) |
| C13 | Nothing clipped or overlapping: no cut text, no node over another node or outside the graph, at 1440 × 900 and 1920 × 1080. Measured: `node tools/layoutcheck.mjs $PWD <pages>` prints `layout clean` |

## 4. Screens and their acceptance criteria

| Id | Screen | Must show | Accept when |
|---|---|---|---|
| S1 | **Builder** (`builder*.html`, composition graph) | Library palette with search and categories; nodes with typed pins (hollow unconnected, filled connected, sampled marked); wires by type; comment frames by area; details panel (name, area, preset, parameters, instruments, ports, faults); Compile with status; compiler results pinned to nodes and wires; minimap; breadcrumbs | A reviewer can tell in under 5 s: which pins of the selected node are unconnected, what type a wire carries, which ports close a loop by sampling, whether the plant compiles, and where the one issue is |
| S2 | **Live run, Explore** (`live-run.html`) | Run bar (mode, pause, step, canvas-driven auto-step with its step size, simulated clock, disturb, end); KPI strip; the graph at compact tier with live readouts and change since the fault; the faulted node unmistakable; inspector with trend (axes, units, fault-onset marker), readings, true state, active disturbances; bottom timeline (disturbances, alarms, observation density) | After injecting `fan_failure` on CT-1, a reviewer can follow the effect (tower → chiller → cooling meter → main meter) on one screen without scrolling |
| S2b | **Live run, Evaluate** (`live-evaluate.html`) | The same layout with truth sealed; a real annunciator alarm (AC-1 discharge temperature, `alarm_state`) | The alarmed node is found in under 3 s; nothing true or sealed is on screen |
| S3 | **Site overview at scale** (`campus.html`, the 26-hall campus, 548 instances, 2,032 tags) | Macro nodes for repeated sections with summary pins and health; bundled wires between them; KPI strip; alarm list; search and filters; minimap | Every hall with an active alarm is found in at most two interactions; nothing on screen is unreadable at the default zoom |
| S4 | **Inside a macro** (`macro.html`, one hall opened) | Breadcrumb back to the site; the hall's nodes with live state; the boundary pins drawn as portals | A reviewer knows where they are and how to go back |
| S5 | **Review** (`review.html`, a finished run) | Verification badge (Verified, Not verified, Incomplete); playback scrubber over the timeline with the graph at the scrubbed time; report: counts, faults, disturbances, effector outcomes, KPI truth, alarm KPIs; true vs reported after sealing | A reviewer can answer "was this run verified, what was injected when, and what did the chiller's COP do" in under 30 s |
| S6 | **Pairing and workspace entry** (`pairing*.html`, `home*.html`) | Pairing-code entry with wrong-code and locked states; workspace contents (plants, compositions, runs with badges); the empty workspace | A first-time user reaches a plant in at most three interactions |
| S7 | **State sheet** (`states.html`) | Every C12 state of controls, nodes, pins and wires; compile-button states; status chips; the port-type legend; the keyboard model | Each state is told apart from its neighbours without colour alone |

## 5. Task walkthroughs (each reviewer walks all six and times them on the mockups)

| Id | Task | Target |
|---|---|---|
| W1 | Add a fourth chilled-water pump (P-5) and connect it to CH-1 (`builder-connect.html`) | ≤ 4 interactions, the compatible pin is obvious |
| W2 | Find why the plant does not compile and get to the offending wire (`builder-error.html`) | ≤ 2 interactions |
| W3 | Start an Explore run from the builder (`builder-simulate.html`) | ≤ 3 interactions |
| W4 | Inject `fan_failure` on CT-1 mid-run and see its effect on the main meter | ≤ 3 interactions, effect visible without navigation |
| W5 | In the campus, open the hall with the active alarm | ≤ 2 interactions |
| W6 | From the run list, decide whether run w-2 can be trusted and what its KPIs were | ≤ 2 interactions |

## 6. Approval

- Each reviewer scores their angle 1–5 and lists findings as High (violates §3, fails a §4 acceptance, or a §5 target),
  Medium (hurts a task or consistency), Low (polish).
- **Approved** when every reviewer scores at least 4, no High finding remains, and every Medium finding is fixed or
  accepted with a written reason in the review record.
- Each round's findings and dispositions are recorded in `docs/simulation-canvas/reviews/mockups.md`.
