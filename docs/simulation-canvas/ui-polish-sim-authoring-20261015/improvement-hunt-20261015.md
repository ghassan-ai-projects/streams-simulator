# Improvement hunt — the canvas as a working engineer's tool

- **Who**: an improvement-hunt round run by a delegated canvas-review agent (not a scoring round; the ten-dimension
  score loop is closed at 9 with no open findings). Scope was corrected twice mid-hunt by the owner:
  **UI and UX only, with no behaviour change** — look, read and feel (hierarchy, spacing, type, colour, density, copy,
  labels, states as they already exist, cross-screen consistency, icons, focus visibility and names, motion, perceived
  performance). Anything that would change what the app *does* is recorded under `## Bigger bets` as a proposal, not as a
  fix. Every improvement also carries the loop's `Dimension:` name (Components; Motion and responsiveness; Colour and
  contrast; Layout; Content design; Feedback and flow; Craft; Whole-product coherence; Typography; Information design of
  graph and review), and the `## Improvements` section is grouped by dimension with each dimension's items ranked by
  impact.
- **Instance**: `http://127.0.0.1:8768`, the review instance, in its own persistent profile
  `/tmp/canvas-review/pw-profile` (paired before this round; **no code entered, no re-pair, storage never cleared**).
  Ports 8766 and 8767 were never touched. Chromium via
  `chromium.launchPersistentContext('/tmp/canvas-review/pw-profile', { channel: 'chrome', viewport: … })`, cache disabled
  per navigation (`Network.setCacheDisabled`).
- **Date**: 2026-10-15.
- **Revision under review**: HEAD was `0a477da` when the hunt started. Three canvas fix commits and one docs commit landed
  while it ran (`a13c028` empty-editor first step, `969531d` links empty-state sentence, `6921832` wires e2e check,
  `3520ff3` expectations rows), so HEAD at hand-off is `3520ff3`. **The served bundle never changed**: the review instance
  served `canvas/dist` as built at 22:16 throughout — `index.html`
  `sha256 0f262832e0677ca8576852c5a5e4d8b53d0059cafb5dd0c52a96bbbd1a659bcb`,
  `assets/index-CS8g2xLX.js` `sha256 319b770313e20d8d6f8cc249bff82d5b47e5ad7cf3b0ff0dcc67a46c3faa7668` (660,743 B),
  `assets/index-Di5IAvBm.css` `sha256 4a538e49dc6493dc93ff9508b79b3b8758587f867a1eaadf4f8a272964b8636e` (82,977 B). Every
  on-screen claim below is a claim about that build. One consequence: `a13c028`'s empty-editor first step is **in source
  but not in the served dist**, so the editor's empty state seen today is the pre-fix one.
- **Sizes and themes**: 1440 × 900 primary, plus **1280 × 720** and **1920 × 1080** on every screen; **System (light) and
  Dark** through the "Theme:" button on Home, the plants, the editor, the run panel and the reviews. Width-specific
  findings are labelled with the width. **The theme was restored to "System" at the end** (final measured label
  "Theme: System", body background `rgb(233, 236, 241)`, light palette), both side panels were left expanded
  (`Outline=true`, `Inspector=true`, inspector 1066 → 1440) and the persistent context was closed cleanly
  ("CONTEXT_CLOSED").
- **Budget used**: **0 runs started, 0 plants created, edited or saved.** Every editor was opened read-only. One plan-local
  mutation is declared: the run panel's "+ Add fault" was clicked once to read the fault-row sentence — that mutates only
  the unsaved in-memory run plan, started no run and wrote nothing; the panel was left afterwards. The workspace listing is
  unchanged (`git status` clean apart from this document and one file another agent left untracked).
- **Where the evidence lives**: 64 screenshots and 30 driver scripts under `/tmp/canvas-review/hunt-20261015/` (never
  committed). Screenshots are cited by filename below.
- **axe**: ran on every surface listed below, in both themes at 1440 × 900 and at 1280/1920 for the four main screens —
  **33 runs, 0 violations**.
- **Console and HTTP**: **zero console errors or warnings and zero HTTP ≥ 400 on every surface in both themes**, with two
  exceptions, both the same 404 and both reproduced deliberately:
  `GET /api/file?path=domains%2Fdir-empty-check-8.domain.json` → 404 with
  `error: Failed to load resource: … 404 (Not Found)`. One instance is a direct URL navigation; the other is the **normal
  path** — Home → `Edit Dir Empty Check 8` — and is written up as IMP-16. No `net::ERR_ABORTED` and no page errors this
  hunt.
- **Reachability**: Home; the multi-hall-campus, dir-review-empty, water-pumping-station, hunt-empty and
  dir-empty-check-8 plants; the campus editor (read-only) and the empty editor; the run panel; reviews w-33 (Evaluate,
  verified), w-44 (did not finish) and w-45 (Explore, verified); the legend popover; the keyboard dialog on a plant and a
  review; the New plant flow. **Not reachable in this workspace**: `factory-utilities` (its routes 404 — its `domains/` and
  `compositions/` entries are absent, so V-8, C-4, C-5 and the CH-2/M-1 wire notes could not be re-checked), and any
  composition carrying a binary validation issue (so **E-6** is unverified for that reason). A run's *running* states
  (R-4, R-5, R-9, R-10, RUN-2, RUN-5) were not entered because the budget was spent on zero runs.

## Expectations still not met

Every row still marked ✗ or ◐ in `docs/simulation-canvas/UI-EXPECTATIONS.md` at `3520ff3` (85 rows), with what was seen
today. "Holds" = the note describes today's build. "Closed" = the note is stale; the row is met on screen now. "Partly" =
one clause of the row is met, another is not (the clause is named). "Not reached" = the surface or state was outside this
hunt's budget (declared, not guessed).

| ID | Mark | Today | Evidence, and what it would take to close |
|---|---|---|---|
| G-1 | ◐ | partly | The top bar itself is now one bar on every screen (product, workspace, breadcrumb, Outline/Inspector, Keyboard, Theme), but each screen still carries its own second bar (plant bar, editor bar, review bar) with its own control set. Close: fold the second row's actions into one shell pattern, or accept two rows as the pattern and say so. |
| G-2 | ◐ | holds | The rail is a fixed 255 px on all three screens, but the right panel is 336 @1280 / 374.4 @1440 / 460 @1920 (IMP-08). |
| G-3 | ◐ | holds | Ids are mono and secondary (`streamsim/air-receiver` under the instance kind line) and units read "12.9 °C"-style, but raw ids still appear in the editor's instance lists and area paths (IMP-18). |
| G-4 | ◐ | closed | Designed empty states are present everywhere reached: "No instances yet." + "No equipment yet / Find a piece of equipment in the Build panel…" (editor), "Select an instance on the graph to see every channel at the playback time." (Inspect), "This plant is not in the workspace any more. Back to the workspace" (missing domain), "None: the plant runs healthy.", "None: nothing is commanded…", "None in this run.". |
| G-5 | ◐ | not reached | No action was left waiting on the binary long enough to observe in-place progress; the slowest reads were the campus domain (541 kB, 7 ms) and `workspace` (99 ms). |
| G-6 | ◐ | closed | No developer jargon is visible in the default state: `native-jsonl`, `sim 0.3.2`, `streamsim` and `digest` exist only inside the closed "For developers" `<details>` (`open:false`) and the "About this run" popover. Visible copy is plain ("Explore: the truth is visible; not for scoring."). |
| G-7 | ◐ | closed | Review playback reads "00:01:00 of 00:30:00", ticks are `hh:mm:ss`, observation counts use thin grouping ("19 089"), and Home rows read "15 Oct 2026, 16:41". The one gap is the About-this-run time clause (IMP-10). |
| G-8 | ◐ | partly | Tab order on Home follows the layout (Keyboard → Theme → New plant → each plant row's Open/Edit/Run) with a visible 2 px `rgb(37,99,235)` focus ring on every stop. The Keyboard dialog's first focus is invisible (IMP-03). |
| G-9 | ◐ | closed | Measured today: light-theme text 6.62–17.49:1, dark 6.54–14.46:1; disabled buttons 4.59:1 (the note says "about 2:1"); wires 4.57:1 (olive), 4.62:1 (teal), 5.34:1 (dimmed grey) against the canvas. The one colour issue left is semantic, not contrast (IMP-05). |
| G-10 | ◐ | closed | No horizontal scroll at any width (`scrollWidth == innerWidth` at 1280/1440/1920 on every screen), the plant summary is not clipped at 1280 (`scrollWidth 135 == clientWidth 135`), and the breadcrumb separator is the same chevron SVG at every width. The remaining gap is tooltip-only copy on the review readouts (IMP-21). |
| G-11 | ◐ | closed | Panel collapse is remembered across screens and across a browser restart (VIS-9); no plant carried unsaved edits here, so the unsaved-leave half was not exercised. |
| G-12 | ◐ | closed | Campus plant view: first instance node 241 ms after DCL, DCL 52 ms. Frame timing over 30 ctrl+wheel zoom steps: median 16.7 ms, p95 17.4 ms, worst 17.6 ms. |
| G-13 | ✗ | closed | The control reads "Theme: System" with a visible label and its state on every screen. |
| G-14 | ✗ | closed | One separator (14 px chevron SVG) and one size (13 px) in the breadcrumb on Home, plant, editor and review, at 1280, 1440 and 1920. |
| G-15 | ✗ | holds | The dialog shows real keys, has no "(graph)"/"(playback)" inside key caps, and now names the event keys ("Run review: Space play or pause, , . go to the previous event, the next event"). But its `::backdrop` computes to `rgba(0, 0, 0, 0)` — the page is **not dimmed** (IMP-02) — and its first focus, the H2 "Keyboard", has `outline: none` (IMP-03). |
| G-16 | ✗ | closed | The zoom readout is `ui-monospace` 12 px on the plant, the editor and the review ("40 %", "45 %", "100 %"). |
| G-17 | ✗ | not reached | Value origins were not exercised (no preset/override/run-setting values were inspected against their source). |
| W-5 | ◐ | not reached | The New plant control was clicked; it navigates straight into the editor with focus already in a field rather than presenting a named form, so the "prominent in the Plants header / validates the id" clauses were not judged. |
| W-8 | ◐ | not reached | No empty *workspace* is reachable — this workspace holds 14 plants and 45 runs. |
| P-1 | ◐ | closed | At 1280 the bar shows the whole summary ("26 instances · 20 links", no overflow) with Edit plant and Run plant; nothing clips. |
| P-3 | ✗ | holds | Campus at 1440: area frames span y 262 → 765 inside a graph of y 84 → 900, so 353 of 856 px (41 %) is empty, split 178 above / 135 below. A one-instance plant opens at zoom 1 (`dir-review-empty`: content 384 → 624 in 84 → 900, 72 % empty). Closing it means changing the opening zoom rule, i.e. the round-16 decision — see `## Bigger bets`. |
| P-4 | ◐ | closed | Selecting `STW-2` on water-pumping-station puts its two wires at `rgb(7,120,102)` 3 px and dims the other nine to `rgb(86,96,112)` 1 px. Measured over five plants, **no wire passes through a node body** (screen-space sampling of every edge path against every node rect) and no wire crosses a frame title's painted text runs. |
| P-7 | ◐ | closed | Links are clickable: the link is a button with `aria-label="Select source comp-u1; comp-u1 air out → rcv-u1 air in"`; kind and area lead ("Compressed-air receiver and header · Utilities"); channels carry units (kPa, m³/h, m, status); faults (21) and states (22) are collapsed rows. |
| P-9 | ◐ | closed | The rail scrolls in its own region, fades its cut row (`data-cut="true"` + a `linear-gradient` mask) and clears both at the bottom. |
| P-10 | ✗ | closed | Navigator is 255 × 177, fully visible at 1280 (y 543 → 720) and 1440 (y 723 → 900). |
| P-11 | ✗ | closed | A Legend control opens a popover on the plant, the editor and the review with wire-colour and pin-shape rows plus "Filled: connected / Hollow: not connected". Its scope is IMP-22. |
| P-12 | ✗ | holds | Hall frames carry no chevron, no "Open", `cursor: default` and no button inside (`react-flow__node-frameTitle`, `aria-roledescription="node"`, no label). The rail also types "HALL 1" against the canvas frame's "Hall 1" (IMP-18). |
| P-14 | ✗ | closed | 0 frame-title-text crossings and 0 node-header crossings measured on five plants (water-pumping-station, multi-hall-campus, food-cold-chain, power-distribution-standby, office-tower-hvac) by sampling every edge path through its screen CTM against the painted text runs. The round-10 "yellow wire through Substation" is not reproducible here. |
| E-3 | ◐ | partly | A "Find equipment" search box exists, and the Name field carries "Leave empty to use the suggested name". Still missing: any description beside a library choice, and the Size select reads only "Default" until a component is picked (so preset sizes cannot be read without choosing one). The chooser's own affordance is IMP-01. |
| E-4 | ✗ | holds | The Link form's "To" select lists all 21 other instances as bare ids, unfiltered by compatibility (IMP-12); the compatibility rule itself is binary-owned (`B`). |
| E-5 | ✗ | partly | The links list has a visible, labelled "Remove" on every row, but selecting a wire on the graph and pressing Delete was not exercised. |
| E-6 | ✗ | not reached | No composition in this workspace carries a validation issue, so no issue list, marker or message could be inspected. |
| E-8 | ◐ | closed | Area editing exists: the AREAS section lists every area with Rename and Delete (with the reason when Delete is refused — "Move the 2 instances out of it first.", "Delete the 3 areas inside it first."), plus "Add area" with a name and a parent. The "move an instance onto an area frame" half was not exercised. |
| E-9 | ◐ | not reached | No unsaved state was created, so the leave-guard could not be observed. |
| E-14 | ✗ | not reproducible | 0 wires through node bodies on five plants; the CH-2 pass-behind and the M-1/M-2 loops name `factory-utilities`, which is not in this workspace. |
| R-1 | ◐ | closed | The run panel shows Seed with a dice button and its explanation, Duration with a unit select, Mode with one-line explanations, Run speed, Faults to inject and Effector calls. |
| R-2 | ◐ | closed | With a fault added, the row reads "At minute [0] [STW-1 · storage tank] has a [Choose a fault]" — a sentence, with the instance select showing the full instance and kind, and the fault list scoped to that instance. |
| R-3 | ✗ | not reached | Picking a fault instance by clicking a graph node was not tried. |
| R-4 | ◐ | not reached | No run was started, so the while-running progress/speed/ticks state was not seen. |
| R-5 | ◐ | not reached | No run was started or finished in this hunt. |
| R-6 | ◐ | not reached | No invalid plan was produced, so field-level problems were not seen. |
| R-7 | ✗ | closed | First focus on opening the run panel is `INPUT#launch-seed`; every control is 28 px tall at 13 px; the instance select reads "STW-1 · storage tank". |
| R-8 | ✗ | closed | Seed 215 × 28, Duration 80 × 28, unit select 83 × 28, fault minute input 28 — all 28 px at 13 px. |
| R-9 | ✗ | not reached | Requires a running run. |
| R-10 | ✗ | not reached | Requires a refused plan. |
| R-11 | ✗ | closed | "Explore" and "Evaluate" are radio rows whose explanations sit on their own indented lines in a lighter tone. |
| V-3 | ◐ | closed | The w-33 timeline shows two labelled fault marks and one effector mark, each with a coloured swatch, and a legend at the right of the playback row: "Fault starts" (dot) / "Effector call" (square). |
| V-5 | ◐ | partly | Inspect for an Evaluate run reads "Observed · rcv-u1 / at 00:01:00 / Latest observed at 00:01:00 / Channel \| Value \| Observed at" with units, plus "True results / No true results for rcv-u1." Still missing: the small trend line across the run and the fault/alarm marks. |
| V-8 | ✗ | not reached | Needs `factory-utilities` (absent). |
| V-9 | ✗ | holds | At 1280 the review opens at 0.4 with instance nodes 86 px past the graph's right edge and a right pan cue; Fit reaches 0.1871. At 1440 it is the floor with a right cue (nodes 36 px clear, frames cut) and Fit reaches 0.285; at 1920 the opening zoom 0.4 ≈ its own fit (0.3956). The rule is consistent — the note's "clipped at 1280" still describes the floor behaviour. |
| V-10 | ✗ | closed | Mark chips are centred on their mark within 6 px (00:10:00 chip centre 531.05 vs track x 531.2; 00:15:00 658.15 vs 660.8; 00:22:00 836.45 vs 842.2), each with its swatch. Only a separate 1 px tick *on* the track is still absent. |
| V-11 | ✗ | closed | Numeric cells are `ui-monospace` + `tabular-nums` + `text-align: right` ("Mean", "Range", "Count" columns); units sit in their own 11 px cell. |
| V-12 | ◐ | closed | Real column headers with words and units: "Starts / Instance / Disturbance", "Time / Instance / Effector / Outcome", "KPI / Mean / Unit / Range", "Measure / Count / What it means", "File / Check", "Records / Count". |
| V-13 | ✗ | closed | An Evaluate run's report says "Evaluate run: the true results below were hidden from the outside system while it ran and shown when it ended." and "TRUE RESULTS — What really happened in the plant. An outside system could not see it while the run was live; it is shown here once the run has ended." That agrees with the run panel's "while it runs". |
| V-14 | ✗ | holds | Kinds are whole words ("pump", "chiller") and rows are capped ("+3 more in Inspect"), but every readout node carries a visible empty band between the kind line and the first channel (IMP-17). |
| V-15 | ✗ | closed | Every icon-only playback control carries both `aria-label` and `title` ("Previous event: none", "Back 10 minutes", "Forward 10 minutes", "Next event: none", "Jump to alarm: none"), and the bell is "Jump to alarm". The three graph controls are the exception (IMP-04). |
| A-2 | ◐ | not reached | The bridge never stopped during this hunt. |
| N-1 | ✗ | holds | No ready-to-run example entry is offered from Home (14 plants, none labelled as an example). |
| N-2 | ✗ | not reached | "New plant" was opened but not submitted; the flow lands in the editor with focus in a field rather than showing a named creation form, so the "human name first / validate the id" clauses were not judged. |
| N-3 | ◐ | not reached | The copy-with-choices form was not reached. |
| N-4 | ✗ | holds | No guided journey exists; the editor's canvas hint is the closest thing ("No equipment yet / Find a piece of equipment in the Build panel, give it a name and an area, and add it. Then link an output to an input, and save to compose the plant."). |
| N-5 | ◐ | closed | Defaults are visible and runnable (seed 1, 30 minutes, Explore, Watch, no faults) and the run panel's footer states exactly what Start will do; the seed is not under an Advanced disclosure. |
| N-6 | ◐ | partly | Presets are listed by title only and the Size select shows "Default" until a component is chosen; no purpose or parameter is shown beside a library choice. |
| N-7 | ◐ | not reached | No error state with a recovery was produced. |
| RUN-1 | ◐ | closed | The footer reads "Start simulates 30 min of plant time in Explore mode with no faults, paced to take about 30 s, then saves the run and opens its review." — except that it is rendered under the "Effector calls" heading (IMP-06). |
| RUN-2 | ◐ | closed | Run speed is an explicit choice — "Watch: about 30 s / The run plays out in about 30 s, so you can follow its time, faults and calls while it runs." against "As fast as possible" — and no live controls are shown. |
| RUN-5 | ◐ | not reached | Complete/cancelled/interrupted/failed states were not produced (no run started). |
| RUN-8 | ✗ | closed | "The same seed, plant, faults and simulator version give the same run." sits under Seed. |
| C-1 | ◐ | closed | Ports read in words with direction and quantity: pins "flow out / cw flow out / liquid out / power out", inspector "header pressure · kPa, outlet flow · m³/h, condensate level · m, drain status", and the link form's output select "pressure out / air out". |
| C-2 | ✗ | not reached | Selecting a wire on the graph was not exercised. |
| C-3 | ✗ | not reached | Not exercised. |
| C-4 | ✗ | not reached | Needs the binary's compatibility answer and the factory plant; the presentational half is IMP-12. |
| C-5 | ✗ | not reached | The instance inspector does state per-port connection ("connected" / "not connected" in the node's accessible names), but the required/optional/default-fed vocabulary was not judged. |
| C-6 | ✗ | not reached | Model connections, physical wiring and consumer communication were not distinguished on screen within this hunt's reach. |
| VIS-1 | ◐ | holds | Title first (breadcrumb + H1), summary second, actions in the same place on plant, editor and review; the editor drops the instance/link summary where the save state sits. |
| VIS-2 | ◐ | closed | Kind in words at the opening zoom with a category band on each node; no icon. |
| VIS-3 | ◐ | closed | Quiet grid, flat nodes, area frames with a light border and a count; selection and abnormal states carry the emphasis. |
| VIS-4 | ◐ | closed | Category band, selection ring, focus ring, fault and alarm marks are visually distinct; combined states were not observed (no faulted node in this hunt). |
| VIS-5 | ◐ | closed | Rows and inspector counts align; actions keep their column on every row. |
| VIS-6 | ◐ | holds | Hall frames with a name, a node count and readable equipment at the opening zoom; the frame title case differs from the rail's (IMP-18). |
| VIS-7 | ◐ | partly | The review shows the timeline marks as one lane with a legend, and its report tables are aligned; there are no separate fault/alarm/effector lanes and no compact trends. |
| VIS-8 | ✗ | closed | Visible Outline and Inspector buttons collapse their panels independently (`aria-pressed`, titles "Hide the outline panel ([)" / "Hide the inspector panel (])"), and the keyboard dialog lists both keys. |
| VIS-9 | ✗ | closed | Collapsing the inspector on a plant persisted to the review and across a browser restart; the button then reads `aria-pressed="false"` with the title "Show the inspector panel (])". |
| VIS-10 | ◐ | holds | Panels start at consistent widths per screen but are not resizable, and the rail and the right panel use different width rules (IMP-08). |
| VIS-11 | ✗ | holds | No reference screens or state sheet exist in the product (a documentation artifact, not a UI change). |

## Per-dimension coverage

| Dimension | What was checked | Verdict |
|---|---|---|
| Components | Buttons and their states, the icon-only set, the equipment listbox, the legend popover, the keyboard `<dialog>`, the run panel's fieldsets and footer, the report tables, the state chips, the Outline/Inspector toggles. | Four improvements (IMP-01, IMP-02, IMP-03, IMP-04); the rest holds. |
| Motion and responsiveness | `transition-duration`/`animation-duration` on every element of Home, the plant and the editor; hover contrast deltas; theme-switch repaint; frame timing over 30 ctrl+wheel zoom steps; `prefers-reduced-motion` emulation. | **Nothing to improve.** 0 elements carry a non-zero transition or animation anywhere reached, so hover/selection/theme changes are instantaneous by construction and reduced motion has nothing to reduce; the theme switch repaints in 23.5–24.6 ms; zoom holds 16.7 ms median / 17.6 ms worst per frame (60 fps). Deliberate, consistent, within §3.2's ≤ 150 ms. |
| Colour and contrast | Sampled text in both themes (h1, summary, kind line, body, table header, section heading, outline row, legend button, chips), wire strokes against the canvas, disabled buttons, the state-chip tokens. | One improvement (IMP-05, a semantic-token clash). Contrast itself clears AA everywhere measured: text 6.54–17.49:1, wires 4.57–5.34:1, disabled 4.59:1. |
| Layout | Shell geometry at 1280/1440/1920 on Home, plant, editor, run panel and review; horizontal scroll; panel widths; panel scrolling and its bottom edge; the run panel's sticky footer; the report's tables; the Navigator. | Three improvements (IMP-06 footer occlusion, IMP-07 missing bottom cue, IMP-08 width rules). No horizontal scroll and no overlap anywhere else; the Navigator is fully visible at every width. |
| Content design | Every headline, hint, empty state, error and disclosure reached, in both themes: Home rows, the plant inspector, the editor's Build panel, the run panel, the review report and its "For developers" disclosures, the keyboard dialog, the About-this-run popover. | Six improvements (IMP-09, IMP-10, IMP-11, IMP-12, IMP-13, IMP-14) and one nit (IMP-15). Plain words throughout; no developer jargon in the default state. |
| Feedback and flow | The normal paths (Home → plant → run panel without starting; Home → editor; review → report/Inspect → playback), plus panel collapse and the empty/missing/finished/not-finished states. Network and console captured per navigation. | One improvement (IMP-16). The states that exist are designed and explained, the pan cue appears where the plant is cut and clears on Fit, and every empty state says what to do next. The running states (R-4/R-9/RUN-5) were not reachable inside the run budget. |
| Craft | Icon set and stroke weights, focus rings, table and number typography, node readout internals, the outline rail's fade and row truncation, popover geometry, the breadcrumb and its truncation. | One improvement (IMP-17, the readout's empty reserved band). One icon set (lucide) at a consistent 2 px stroke with a single 2.25 px logo exception; 2 px focus rings on every tab stop; numbers tabular and right-aligned. |
| Whole-product coherence | The same plant on the plant screen, in the editor and in a review; the legend on plant and review; the panel names; the area-name forms; the link notation; the opening viewport rule across screens and widths. | Three improvements (IMP-18, IMP-19, IMP-20). The opening viewport rule is coherent: each screen opens at its own fitted zoom when that is above the 0.4 floor and at the floor otherwise (campus review 0.3956 ≈ 0.4 at 1920 where it fits; campus plant 0.455 at 1920; both at the floor with a cue at 1280/1440 where they do not). |
| Typography | The type scale, font families, tabular figures, unit treatment, id treatment, the zoom readout, the breadcrumb, truncation with `text-overflow` and its `title`. | **Nothing to improve.** Sizes cluster at 11/12/13 px with a 13 px semibold H1; ids and values are `ui-monospace`; units are 11 px in a lighter tone and in their own table cell; numeric columns are `tabular-nums` and right-aligned; the zoom readout is one font and size on plant, editor and review; nothing measured clipped mid-token without an ellipsis and a `title`. |
| Information design of graph and review | Node readouts (values, units, "+N more in Inspect", unobserved channels), the instance inspector, the review timeline, scrubber, marks and legend, the report sections and tables, the plant overview, the graph's accessible names. | Two improvements (IMP-21, IMP-22). Node and edge accessible names are good (`role=group` "STW-1, storage tank, 2 inputs, 1 output"; edges `role=img` "pl-h1 to mtr-h1"), tables are aligned and headed, and the report leads with what happened. |

## Improvements

Grouped by dimension in the loop's own order; within each dimension, most impactful first. IMP numbers are stable
document order.

### Components

#### IMP-01 The equipment chooser hides most of the library with no cue
- **Kind**: affordance · **Dimension**: Components
- **Where**: editor → Build panel → "Add equipment" (the `select[aria-label="Equipment"]` listbox).
- **Observed**: the chooser is a native multi-row listbox — rect `1079, 167, 349 × 142`, `size="6"` — so 6 of the 13
  components of the campus composition (and 6 of the 20 offered on the empty editor) are painted at once, with
  `scrollHeight 448` against `clientHeight 140` (≈ 8 rows hidden). It has no visible border of its own and macOS draws its
  overlay scrollbar only while scrolling, so the block of six titles reads as the whole library
  (`30-editor-1440-light.png`, `C2-editor-dir-empty-check-8.png`).
- **Why it matters**: adding equipment is the first step of the authoring flow. An engineer who cannot see that a component
  exists will not search for it, and the "Find equipment" field only helps someone who already knows the name.
- **Suggestion**: give the listbox a visible 1 px boundary and its own scroll region and label it with its size
  ("13 components") so a partial view is legible as partial. Raising `size` to fit the library (13–20 rows) is the smallest
  change that removes the hidden state entirely; the boundary plus the count is the smaller one.
- **Effort**: S · **Impact**: medium · **Needs binary**: no (markup and CSS).

#### IMP-02 The keyboard dialog does not dim the page
- **Kind**: components · **Dimension**: Components
- **Where**: the "Keyboard" dialog, opened from every screen's top bar.
- **Observed**: `getComputedStyle(dialog, "::backdrop").backgroundColor` is `rgba(0, 0, 0, 0)` on a plant and on a review;
  in `A2-keyboard-dialog.png` the plant behind the dialog is at full contrast, so the dialog reads as a floating card
  rather than a modal. G-15 asks for exactly this dimming.
- **Why it matters**: the dialog is 633 × 492 over a dense graph; without the dim the graph competes with the help text and
  the dialog's own edges are the only thing separating them.
- **Suggestion**: one backdrop token (a light-theme `rgba(15, 17, 21, 0.28)`-class value and its dark counterpart) on
  `dialog::backdrop`, reusing the app's existing scrim token if it has one.
- **Effort**: S · **Impact**: medium · **Needs binary**: no.

#### IMP-03 The keyboard dialog's first focus is invisible
- **Kind**: a11y · **Dimension**: Components
- **Where**: the "Keyboard" dialog, on open (plant and review).
- **Observed**: `document.activeElement` on open is the H2 "Keyboard" and its computed outline is
  `rgb(20, 26, 34) none 3px` — `outline-style: none`, no box shadow. The same check on the top bar's tab stops gives a 2 px
  `rgb(37, 99, 235)` ring, so the dialog is the one place where focus lands invisibly.
- **Why it matters**: §3.3 and G-8 require focus to be always visible; a keyboard user who opens the dialog with `?` gets
  no confirmation of where they are.
- **Suggestion**: move initial focus to the dialog's Close button (the first actionable control, and the one G-15 names), or
  give the focused heading the same focus ring as everything else.
- **Effort**: S · **Impact**: low-medium · **Needs binary**: no.

#### IMP-04 The graph's zoom controls are the only icon-only controls without a tooltip
- **Kind**: consistency · **Dimension**: Components
- **Where**: the graph bar on the plant view, the editor and the review (Zoom out, Zoom in, Fit the plant in view).
- **Observed**: all three carry `aria-label` ("Zoom out", "Zoom in", "Fit the plant in view") and a 14 px lucide icon, and
  `title` is `null`; every other icon-only control on the same rows — the playback steps, the alarm jump — carries both
  `aria-label` and `title` ("Back 10 minutes", "Jump to alarm: none").
- **Why it matters**: §3.2 requires each icon to have a text label or a tooltip; the fit glyph (four corner brackets) is not
  self-evident, and it is the control a user needs most when the opening view is cut.
- **Suggestion**: set `title` from the existing `aria-label` on those three buttons.
- **Effort**: S · **Impact**: low · **Needs binary**: no.

### Motion and responsiveness

No improvements. See the per-dimension table and `## Checked and holding`.

### Colour and contrast

#### IMP-05 "Did not finish" shares the success green with "Verified"
- **Kind**: consistency · **Dimension**: Colour and contrast
- **Where**: the run state chips on Home and in a review header.
- **Observed**: the review's Verified badge is `color rgb(15, 107, 63)` on `rgba(15, 107, 63, 0.12)` with a matching border
  (`_verified_qn2s5_45`, 11 px/600, with an icon). Home's "Did not finish" chip measures exactly the same
  `rgb(15, 107, 63)` on `rgba(15, 107, 63, 0.12)` with the same border. "Explore" uses purple `rgb(106, 63, 200)` and
  "Evaluate" neutral `rgb(75, 85, 99)`.
- **Why it matters**: an abandoned run is the one run state an engineer must not mistake for a good one; sharing the ok
  token makes it read as a positive result at a glance, and colour is the fastest channel on a scanning screen. The word and
  the icon do distinguish it, so this is a token clash rather than an accessibility failure.
- **Suggestion**: give unfinished/incomplete its own neutral-or-attention token (the "Evaluate" neutral, or the warn token),
  leaving the ok green to Verified alone.
- **Effort**: S · **Impact**: low-medium · **Needs binary**: no.

### Layout

#### IMP-06 The run panel's sticky footer covers the last section's text and its button
- **Kind**: layout · **Dimension**: Layout
- **Where**: the run plan panel (right panel), after "Faults to inject"; measured at 1440 × 900.
- **Observed**: the panel's footer is `position: sticky`, rect y 785 → 900, white with a 1 px top border and a shadow. The
  last fieldset ("Effector calls", y 762 → 889) has its children **underneath it**: "None: nothing is commanded…" at
  y 786 → 843 and the **"Add effector call" button at y 859 → 889** both report `coveredByFooter: true`. The panel scrolls
  (scrollHeight 976 vs clientHeight 856, maxScroll 120) and scrolling to the bottom lifts the button clear to y 739 → 769,
  so nothing is permanently unreachable — but at the default position the section shows only its heading and the footer's
  own sentence, and that sentence ("Start simulates 30 min of plant time…") then reads as Effector-calls content
  (`C0-run-footer-2x.png`, `40-run-panel-1440-light.png`).
- **Why it matters**: the primary action's pre-flight sentence is read against the wrong heading, the section's own action is
  invisible on open, and at 1280 × 720 the panel is 212 px shorter so more of the last section is covered.
- **Suggestion**: reserve the footer's height as padding-bottom on the panel's scroll body so no section content can sit
  under it, and give the footer a full-bleed separator (or a surface change) so its sentence is visibly not part of the
  section above it.
- **Effort**: S · **Impact**: medium-high · **Needs binary**: no.

#### IMP-07 The right panel scrolls with no cue that anything is below
- **Kind**: affordance · **Dimension**: Layout
- **Where**: every right panel that scrolls: the plant inspector, the editor's Build panel, the run plan, the review report.
- **Observed**: the editor's `aside._panel` has `overflow: auto`, `scrollHeight 1423` against `clientHeight 856`,
  `mask-image: none`, `data-cut` absent and `padding-bottom: 0`; its content ends mid-list at the panel's bottom edge
  (`30-editor-1440-light.png` shows "Move the 2 instances out of it first." cut by the window edge). The **outline rail in
  the same shell does have the cue** — `data-cut="true"` with
  `linear-gradient(rgb(0,0,0) calc(100% - 32px), rgba(0,0,0,0))`.
- **Why it matters**: 567 px of the Build panel's content and 120 px of the run plan's are below the fold with no
  affordance, so a user can finish a task without ever seeing Areas or the Effector-calls action. The shell already owns the
  pattern; only one panel uses it.
- **Suggestion**: apply the rail's existing cut fade (or a 1 px inset shadow) to the right panel's bottom edge when it
  scrolls, driven by the same scroll listener the rail already has.
- **Effort**: S · **Impact**: low-medium · **Needs binary**: no.

#### IMP-08 The outline rail and the right panel size themselves by different rules
- **Kind**: consistency · **Dimension**: Layout
- **Where**: the shell, at 1280 / 1440 / 1920.
- **Observed**: the left rail is a fixed 255 px at all three widths; the right panel is 336 px at 1280, 374.4 px at 1440 and
  460 px at 1920, so the panel grows by 37 % between the narrowest and widest supported sizes while the rail never moves.
  G-2 and VIS-10 ask for one width token per panel kind.
- **Why it matters**: the same panel holds the same content (a labelled two-column table, a select, a value list) at three
  different widths; a table that fits at 1440 wraps at 1280 for no reason the user can see, and a scan pattern learned on
  one screen does not transfer to another.
- **Suggestion**: pick one rule for both panels — a fixed width (the rail's rule) or a proportional one applied to both —
  and let only the centre take the extra space. The fixed rule is the smaller change and matches the rail that already does
  it.
- **Effort**: S · **Impact**: low · **Needs binary**: no.

### Content design

#### IMP-09 One panel is called two things
- **Kind**: copy · **Dimension**: Content design
- **Where**: the keyboard dialog's "Everywhere" group, against the top-bar button and the review's panel label.
- **Observed**: the dialog says "hide or show the outline panel, the **details panel**"; the button it describes is labelled
  "Inspector" with `title="Hide the inspector panel (])"`; the review's panel region is `aria-label="Run details"`. The plant
  inspector is `aria-label="Inspector"`.
- **Why it matters**: G-6 asks one action to have one name, and this is the control that shows or hides the panel a user is
  reading. A keyboard user who reads "details panel" and looks for a "Details" button will not find it, and it is the first
  hint a screen-reader user gets about the panel.
- **Suggestion**: use "Inspector" in the keyboard help, in the button title and in the review's region label (or rename all
  four together); no new string is needed beyond the existing one.
- **Effort**: S · **Impact**: medium · **Needs binary**: no.

#### IMP-10 The About-this-run sentence does not parse
- **Kind**: copy · **Dimension**: Content design
- **Where**: the ⓘ "About this run" popover in the review header (440 × 69 px).
- **Observed**: the popover renders "Explore: the truth is visible; not for scoring. / dir-review-empty 0.1.0 · seed 1 ·
  native-jsonl · sim 0.3.2 · **run time 00:00:00 is 00:00:00 UTC simulated**" (identical shape on w-33 and w-45).
- **Why it matters**: this is the one place a user can check how a run was made; the closing clause states two facts as one
  equation with no labels, so neither the wall-clock time nor the simulated span can be read out of it even though both are
  known facts about the run.
- **Suggestion**: split it into labelled facts, e.g. "Simulated 30 min of plant time" and the wall-clock line in G-7's
  format once it exists; in the meantime, drop the duplicated "00:00:00 is" and label the one fact that is known.
- **Effort**: S · **Impact**: medium · **Needs binary**: partly — the wall-clock run date is G-7's recorded `B` gap; the
  sentence can be corrected now for the facts the bridge already provides.

#### IMP-11 The graph's screen-reader instructions promise deletion on read-only screens
- **Kind**: a11y · **Dimension**: Content design
- **Where**: the plant graph on the plant view, in the editor and in a review.
- **Observed**: a `display: none` text node inside `[role="application"]` reads "Press enter or space to select a node. Press
  delete to remove it and escape to cancel." (with the matching edge sentence) on all three screens; on the plant view and
  the review there is nothing to delete, and the app's own visible instruction beside it says only "Enter selects an instance
  or opens a hall".
- **Why it matters**: it is the first thing a screen-reader user hears about the graph, and it describes an action the screen
  does not support — the same principle as G-15's "describes no Play that does not exist".
- **Suggestion**: pass the editor's deletable description only where deletion exists, and a selection-only description on
  the plant view and the review.
- **Effort**: S · **Impact**: low-medium · **Needs binary**: no.

#### IMP-12 The Link form offers bare instance ids
- **Kind**: information · **Dimension**: Content design
- **Where**: editor → select an instance → "Link an output" → the "To" select.
- **Observed**: its 21 options are `["comp-u1","ch-u1","ch-u2","ct-u1","pmp-u1",…,"mtr-h3"]` — ids only, no kind, and the
  chosen output's family ("pressure out" / "air out") gives no clue which of them can take it. Every other place the app names
  an instance it appends the kind: the graph nodes' accessible names ("comp-u1, air compressor, 1 input, 2 outputs"), the
  outline rows ("comp-u1 air compressor") and the run panel's fault instance select ("STW-1 · storage tank").
- **Why it matters**: linking is the step where a wrong guess is expensive (the binary refuses it), and the one place the
  app drops the kind is the one place the user has to choose between 21 similar-looking ids.
- **Suggestion**: append the kind to each option ("ch-u1 · chiller"), matching the run panel's fault select. Full
  compatibility filtering is a binary capability (see `## Bigger bets`).
- **Effort**: S · **Impact**: medium · **Needs binary**: no.

#### IMP-13 The empty Halls group reads "0 · empty"
- **Kind**: copy · **Dimension**: Content design
- **Where**: the editor's outline rail, when the plant has no halls (and the editor's area group on an empty plant).
- **Observed**: the group header is "Halls" with the count element `_count_1swad_112` reading "0 · empty" (11 px, uppercase),
  while every sibling group shows a bare count ("2", "9", "5", "12"). The empty editor's area group reads
  "Factory 0 · empty".
- **Why it matters**: one group in a column of counts renders a value-and-state pair where a number is expected, which breaks
  the column's scan rhythm and reads as two facts about one thing that has one.
- **Suggestion**: drop the state word from the count slot (a bare "0" matches the column) and, if the empty case needs words,
  say "No halls yet" in the group's own row area where the Add-area action is.
- **Effort**: S · **Impact**: low · **Needs binary**: no.

#### IMP-14 The plant row's not-runnable sentence is a run-on
- **Kind**: copy · **Dimension**: Content design
- **Where**: Home → the plant row whose composition does not compose.
- **Observed**: "Not runnable yet: open Edit, save, and fix any issue it lists." sits in the row's meta slot, where every
  other row reads "14 instances · 7 runs · last 15 Oct 2026, 16:15".
- **Why it matters**: it is the row's only sentence and it is doing three jobs (state, destination, remedy) in one clause
  chain, which is why it wraps where the others read as a single fact.
- **Suggestion**: state first, action second — "Not runnable yet. Edit it, then save." — or move the remedy to the Edit
  button's tooltip and leave the row with the state.
- **Effort**: S · **Impact**: low · **Needs binary**: no.

#### IMP-15 The review says "Verified" twice in adjacent lines
- **Kind**: copy · **Dimension**: Content design
- **Where**: the review header chip and the first line of the report.
- **Observed**: the header chip reads "Verified" (green, with a tick icon), and directly under it the report's collapsed
  integrity line reads "Verified: the files match and the replay reproduces."
- **Why it matters**: two readings of one fact — a badge and a sentence — on the same screenful; a scanning user has to check
  whether the second adds anything to the first.
- **Suggestion**: let the badge own the word and the sentence own the explanation ("The files match and the replay
  reproduces."), which is also shorter inside a 374 px panel.
- **Effort**: S · **Impact**: low · **Needs binary**: no.

### Feedback and flow

#### IMP-16 A doomed domain probe leaves a 404 and a console error on the normal editor path
- **Kind**: feedback · **Dimension**: Feedback and flow
- **Where**: Home → the "Dir Empty Check 8" row → `Edit` (the row Home itself labels "Not runnable yet").
- **Observed**: the click issues `200 GET /api/file?path=compositions%2Fdir-empty-check-8.composition.json`, then
  **`404 GET /api/file?path=domains%2Fdir-empty-check-8.domain.json`**, then `200 POST /api/actions/component.list`, and the
  page logs `error: Failed to load resource: the server responded with a status of 404 (Not Found)`. The editor still renders
  correctly ("Nothing to save yet: add equipment in Build.") and correctly offers no "Open plant".
- **Why it matters**: §3.1 requires no failed request and no console error in normal use, and the workspace listing already
  told the canvas that this plant has no composed domain. A red console entry on the first click undermines the instrument's
  credibility, and this is exactly the state a user reaches to repair a broken plant.
- **Suggestion**: answer "is there a composed domain?" from the workspace listing the editor already holds — the same fact the
  Home row used — instead of re-probing for the file, and keep the existing not-composed panel. This removes one request
  rather than adding any, and needs no new capability or endpoint.
- **Effort**: S · **Impact**: medium · **Needs binary**: no.

### Craft

#### IMP-17 The readout node carries an empty band between its kind and its channels
- **Kind**: craft · **Dimension**: Craft
- **Where**: every review readout node (w-45's `CP-1`, w-33's `ch-u2`…).
- **Observed**: in the 2× crop `93-w45-node-2x.png` the node's header (id + kind, on the grey surface) ends and then ~30 CSS
  px of bare node surface runs before the first channel row, with no content and no separator; this is the band P-12-2
  recorded as 23 px reserved below the kind row for fault/effector/outcome lines, present whether or not the node has such a
  line.
- **Why it matters**: it is the largest piece of unexplained whitespace inside a dense data node, and at the opening zoom it
  is what makes readout nodes look half-built (V-14's "tall nodes with empty bottoms"); an engineer reading a node top to
  bottom has to decide whether the gap means something.
- **Suggestion**: collapse the band when the node has no fault/effector/outcome line to show, or paint it as part of the
  header (the id/kind surface) so it reads as structure rather than a hole.
- **Effort**: M · **Impact**: medium · **Needs binary**: no.

### Whole-product coherence

#### IMP-18 One area is typed three ways on one screen
- **Kind**: consistency · **Dimension**: Whole-product coherence
- **Where**: the editor (campus composition) and the plant view — rail, canvas frame, Area select, AREAS list.
- **Observed**: the rail shows "HALL 1" (uppercase via CSS, 11 px); the canvas frame shows "Hall 1"; the Area select and the
  AREAS list show "Site › Halls › Hall 1". On the plant view the rail shows "HALL 1" against the frame's "Hall 1".
- **Why it matters**: matching a rail row to a frame is the main way a user navigates a 26-instance plant; three spellings of
  one name force a re-read on every move. The path form is genuinely needed in the select (it disambiguates), so the fix is
  to keep the leaf name where the leaf is shown.
- **Suggestion**: leaf name in the rail and on the frames, full path only in the Area select and the AREAS list.
- **Effort**: S · **Impact**: medium · **Needs binary**: no.

#### IMP-19 The same plant's outline is ordered differently on the plant screen and in the editor
- **Kind**: consistency · **Dimension**: Whole-product coherence
- **Where**: the outline rail on the plant view and in the editor, same plant, same session.
- **Observed**: plant rail order — `wx-site`, `cal-site`, `comp-u1 … mtr-u1` (Utilities), then `pl-h1 … mtr-h1` (Hall 1) and
  the rest of the halls. Editor rail order — `wx-site`, `cal-site`, then `pl-h1 …` (Hall 1), Hall 2, Hall 3, and
  **Utilities last**, with an extra "Halls" group heading the hall block.
- **Why it matters**: the rail is the map of the plant; if the map reorders when the user switches to the editor to fix
  something, their position in it is lost and the two screens cannot be read side by side. Round 15 unified the area order
  between the rail and the AREAS list on the plant screen; this is the same class one screen over.
- **Suggestion**: order the editor rail by the plant's area order (as the plant rail and the AREAS list already do), keeping
  the "Halls" affordance as a heading at Hall 1's position rather than as a reordering.
- **Effort**: M · **Impact**: medium · **Needs binary**: no.

#### IMP-20 The same link is written in two directions
- **Kind**: consistency · **Dimension**: Whole-product coherence
- **Where**: the plant inspector's "Links in" row vs the editor's "Links" row.
- **Observed**: the plant inspector renders "comp-u1 air out → rcv-u1 air in" (source first, arrow right; its `aria-label`
  reads "Select source comp-u1; …"); the editor renders "air in ← comp-u1 air out" (destination first, arrow left).
- **Why it matters**: the two rows describe one relationship; reading one before the other means re-orienting, and the arrow
  changes meaning between them. A link is always "an output supplies an input" — the legend's own sentence — so one
  direction is enough.
- **Suggestion**: use the source → destination order in both places.
- **Effort**: S · **Impact**: low · **Needs binary**: no.

### Typography

No improvements. See the per-dimension table and `## Checked and holding`.

### Information design of graph and review

#### IMP-21 A channel with no reading is a bare dot whose meaning is only in a tooltip
- **Kind**: information · **Dimension**: Information design of graph and review
- **Where**: review readout nodes (all runs); the same markup is in the plant graph's readouts.
- **Observed**: the markup is
  `<div class="_unobserved_…"><dt title="suction level">suction level</dt><dd><span title="not observed"><span aria-hidden="true">·</span><span class="sr-only">not observed</span></span></dd></div>`
  — so the visible value is a single middle dot (`·`, painted at x 760.7 → 764 in a 237 px node) and the words exist only in a
  `title` and a screen-reader span. On the w-33 campus review the pattern repeats dozens of times per screen
  (`93-w45-node-2x.png`, `G1-review-w33.png`); at the opening zoom the dot is a speck at the row's right edge.
- **Why it matters**: G-10 says a tooltip must never be the only copy of a fact, and this is the fact that decides whether a
  reading is missing or zero. In a review the whole point is spotting what the plant did *not* report; a row of dots cannot be
  scanned, and a reader who does not hover concludes the value is blank or broken.
- **Suggestion**: paint the state — a dimmed em dash or "not observed" in the value slot, at the same 11 px unit tone —
  instead of the middle dot, keeping the `title` for the longer phrasing.
- **Effort**: S · **Impact**: medium-high · **Needs binary**: no.

#### IMP-22 The legend lists every port family in the product, not the plant's
- **Kind**: information · **Dimension**: Information design of graph and review
- **Where**: the Legend popover on the plant view, the editor and the review.
- **Observed**: the same 12 rows render everywhere reached — "Compressed air, Pressure, Water, Condenser water, Temperature,
  Heat, Power, Signal, Weather" plus "Port not recorded / Filled: connected / Hollow: not connected" — on the 26-instance
  campus, on the 1-instance `dir-review-empty` plant and on a review. water-pumping-station's wires were measured as exactly
  two families: `rgb(7, 120, 102)` (Water) and `rgb(143, 98, 0)` (Power).
- **Why it matters**: seven of the nine rows describe wire colours that are not on the screen, in the one popover a user
  opens *because* they cannot tell the wires apart; it makes a two-colour plant look like a nine-colour one and buries the
  rows that matter.
- **Suggestion**: keep the vocabulary but mark the plant's own families — full contrast for the families present in the
  current plant, dimmed for the rest — or lead the popover with "In this plant: Water, Power" and keep the full list under it.
  The canvas already computes each wire's family to colour it, so no new fact is needed.
- **Effort**: S · **Impact**: medium · **Needs binary**: no.

## Quick wins

Small, safe, no behaviour change, in the order they are worth doing:

- [ ] IMP-02 add a `dialog::backdrop` scrim (G-15's dimming) — one CSS rule.
- [ ] IMP-04 set `title` from the existing `aria-label` on Zoom in / Zoom out / Fit.
- [ ] IMP-15 let the integrity sentence drop its repeated "Verified:" now that the badge carries it.
- [ ] IMP-21 paint "not observed" (or a dimmed em dash) instead of the bare `·` in readout values.
- [ ] IMP-09 use one name for the right panel in the keyboard help, the button title and the review region label.
- [ ] IMP-03 move the keyboard dialog's initial focus to Close (or ring the focused heading).
- [ ] IMP-06 reserve the sticky footer's height in the run panel's scroll body.
- [ ] IMP-07 reuse the outline rail's cut fade on the right panel's bottom edge.
- [ ] IMP-13 drop "empty" from the rail's count slot.
- [ ] IMP-12 append the kind to each instance option in the Link form.
- [ ] IMP-18 use the leaf area name in the rail and on the frames; keep the path in the select.
- [ ] IMP-05 give "Did not finish" its own token instead of the ok green.
- [ ] IMP-01 give the equipment listbox a visible boundary and a component count.
- [ ] IMP-16 answer the composed-domain question from the workspace listing so no 404 is issued.
- [ ] IMP-10 split the About-this-run time clause into labelled facts.
- [ ] IMP-19 order the editor rail by the plant's area order.
- [ ] IMP-08 pick one width rule for the rail and the right panel.
- [ ] IMP-20 write links source-first in both places.
- [ ] IMP-11 pass a selection-only graph description on the plant view and the review.
- [ ] IMP-14 make the not-runnable plant row one state and one action.

## Bigger bets

Proposals only — each changes what the app does, or needs design or binary work, so none is a fix to make now.

1. **A tier-aware node header, or no reserved band at all** (P-12-2's parked item; needs binary: no, needs design). IMP-17
   is the CSS-sized half of it: the band can be collapsed, but the fault/effector/outcome lines that use it appear *while* a
   run plays, so collapsing it makes the node grow when a fault lands — the layout jump the reserved band exists to prevent.
   The real fix is a layout that measures the lines it will show, which round 16 deliberately deferred (ISS-8D-17).
   Tradeoff: a stable node that reserves space versus a compact node that moves.
2. **Fit-first opening on wide plants** (P-12-1 / P-3). Today every screen opens at the 0.4 floor and gets a pan cue when the
   plant is cut; P-3's 41 % empty vertical band on the campus at 1440 comes from the same rule. Letting wide plants open
   fitted re-introduces sub-0.4 tiers with 8 px kind lines — the regression round 10 closed. Tradeoff: no cut to explain
   versus illegible labels at the opening view; needs the overview tier to get genuinely better first.
3. **Compatibility-aware linking** (E-4, C-4, C-5; needs binary: yes). The Link form lists all 21 other instances because
   the compatibility rule belongs to the binary. IMP-12 is the presentational half available now; a real "only compatible
   inputs" list — and the required/optional/default-fed vocabulary of C-5 — needs the binary to expose the rule.
4. **Serve the bundle compressed** (performance; needs binary: yes — it is the `streamsim canvas` response, a Go edge). The
   §4 gzip budget is met with room to spare: the built JS is 660,743 B raw and **203,494 B gzip**, the CSS 82,977 B raw and
   **15,164 B gzip**, so 218,658 B (≈ 213 kB) against the 600 kB budget. But this instance serves both files **uncompressed**
   (`encodedBodySize == decodedBodySize == 660,743`), so a cold load actually transfers ≈ 726 kB, 3.3× the budget. Tradeoff:
   a one-line transport change against the appearance of a budget breach on every cold load.
5. **A run date in the About-this-run popover** (G-7's `B` item; needs binary: yes). IMP-10 can fix the sentence, but
   "15 Oct 2026, 16:41" for a run needs the binary to report when it ran.
6. **A ready-to-run example and a guided journey** (N-1, N-4). The Home screen offers 14 plants and no example, and the
   editor's canvas hint is the only guidance; both are new content and new interaction, not a UI polish. Tradeoff: the
   shortest first-use path versus more surface to maintain and keep honest.
7. **Reference screens and a state sheet** (VIS-11; a documentation artifact outside this document). Every state in this hunt
   was measured off the built product instead; a sheet would make the next review cheaper without changing the UI.

## Checked and holding

Do not redo these; they are measured and correct on this build.

- **Wires.** On five plants, sampling every edge path through its screen CTM against every node rect and every frame title's
  painted text runs: **0 wires through a node body, 0 wires across a frame title's text**. Wires are
  `vector-effect: non-scaling-stroke` at 1.6 px, so they stay 1.6 CSS px at every zoom; families are `rgb(7,120,102)` Water,
  `rgb(143,98,0)` Power, `rgb(86,96,112)` dimmed, all ≥ 3:1 against the canvas.
- **Selection highlight.** Selecting `STW-2` puts its two wires at 3 px in their family colour and dims the other nine to
  1 px grey; selecting a node also highlights its rail row and fills the inspector.
- **Opening viewport.** One rule, honoured: each screen opens at its own fitted zoom when that is above the 0.4 floor
  (campus plant 0.455 and campus review 0.3956 ≈ its fit at 1920, both with no cut) and at the floor with a cue where the
  content does not fit (campus plant and review at 1280/1440 with `[data-pan-cue="right"]`), and Fit lowers the zoom below the
  floor to fit (review 0.285 at 1440, 0.1871 at 1280) with every cue cleared. Plant and editor are identical to the pixel for
  the same plant.
- **Empty, missing and unavailable states.** "No instances yet." / "No equipment yet / Find a piece of equipment in the Build
  panel…" (editor), "Select an instance on the graph to see every channel at the playback time." (Inspect), "This plant is not
  in the workspace any more. Back to the workspace / Details" (domain file absent), "None: the plant runs healthy." / "None:
  nothing is commanded…" / "None in this run.", "Not checked: this run could not be replayed for its verdict." / "Did not
  finish" (w-44), "True results / No true results for rcv-u1." for a sealed Evaluate inspect.
- **Numbers and tables.** All numeric cells `ui-monospace` + `tabular-nums` + right-aligned; units in their own 11 px cell
  with a screen-reader-only "Unit" header (the visible header is intentionally absent because the unit sits with the value);
  report tables carry real headers ("Starts/Instance/Disturbance", "Time/Instance/Effector/Outcome", "KPI/Mean/Unit/Range",
  "Measure/Count/What it means", "File/Check", "Records/Count"); counts use thin grouping
  ("19 089 observations · readings start at 00:01:00").
- **Review playback.** Play + Space, speed chips 1×/10×/60×/600× with `aria-pressed`, "00:01:00 of 00:30:00", tick labels
  inset by half a label at each end (as they should be), mark chips centred on their mark within 6 px, and a "Fault starts /
  Effector call" legend at the right of the playback row.
- **Keyboard dialog content.** A real `<dialog>`, per-screen groups ("Everywhere", "Plant graph", "Run review"), clean key
  caps with no "(graph)"/"(playback)", `[`/`]` documented for the panels, the review group naming Space and `,`/`.` for
  previous/next event, and Escape returning to its button. (Its backdrop and focus ring are IMP-02 and IMP-03.)
- **Panel collapse.** Both panels collapse independently through visible toggles with correct `aria-pressed` and titles, the
  graph takes the freed space, and the collapsed state survives navigation and a browser restart.
- **Graph accessibility.** Instance nodes are `role="group"` with names like "STW-1, storage tank, 2 inputs, 1 output"; edges
  are `role="img"` with "pl-h1 to mtr-h1"; the graph itself is `role="application"`, labelled "Plant graph" and focusable; the
  zoom control group is labelled "Zoom"; the outline, Navigator, Inspector, plant and run-details regions all carry names.
- **Focus visibility and order.** Every top-bar and row control reached shows a 2 px `rgb(37,99,235)` ring, in tab order that
  follows the layout (Keyboard → Theme → New plant → each plant row's Open/Edit/Run).
- **Theme switching.** Repaints within 23.5–24.6 ms (two frames), no layout shift, System/Light/Dark all reachable, both
  themes sampled for contrast, restored to System.
- **Run panel content.** Dice button, seed explanation, duration + unit, Explore/Evaluate with one-line explanations on their
  own lines, Watch / As-fast-as-possible, "Faults to inject / None: the plant runs healthy. / Add fault" and "Effector calls /
  None: nothing is commanded… / Add effector call" (the last of these is IMP-06's occlusion), first focus on Seed, and every
  control 28 px at 13 px.
- **Editor craft.** Undo/redo controls, "No changes" beside a disabled Save/Revert (both `aria-disabled`, still focusable by
  design, text 4.59:1), "Open plant", the AREAS list with per-area Rename/Delete and the refusal reason ("Move the 2 instances
  out of it first.", "Delete the 3 areas inside it first."), the sensors-not-added-by-hand explanation, and the links row
  "air in ← comp-u1 air out" with a labelled Remove.
- **Icons.** One set (lucide), one stroke weight (2 px, with the 2.25 px product logo as the single exception), every
  icon-only control carrying at least an accessible name.
- **Performance.** DCL 52 ms with the 660 kB JS and 83 kB CSS fetched in ~25 ms combined; campus plant interactive 241 ms
  after DCL (budget 3 s); zoom 16.7 ms median / 17.6 ms worst frame; `workspace` 99 ms, campus domain 541 kB in 7 ms.
- **Console and HTTP.** Clean on every surface except IMP-16's single 404.

### One measurement that could not be taken cleanly

**Shell first render.** FCP measured 1,448 ms (Home, cache disabled) and 1,972 ms (campus navigation) in this headless
persistent profile, against DCL 52 ms, a fully fetched resource set in ~150 ms and a first header at 87–89 ms. The timestamps
look dominated by this harness's paint scheduling in headless Chrome rather than by app work — the same profile gave "89 ms"
to the same measurement in round 14, and the §4 budget is ≤ 1.5 s. Reported as measured, not as a finding.
