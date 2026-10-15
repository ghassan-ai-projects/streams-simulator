# Director score, dimension: Information design of the plant graph and the review, round 3

- Director: UI/UX director (brief `brief-director-dim.md`), Chrome at http://127.0.0.1:8766.
- Previous score: 5 (round 2, `graph-and-review-round-2.md`: live 6, plant 5). Round 2 priority list: A-1, A-8, A-2, A-5, A-6, A-3 (plus the rest of A-1 to A-11).
- Bar: `docs/simulation-canvas/UI-EXPECTATIONS.md` (P-2, P-4, P-7, P-14, E-14, V-3, V-4, V-9, V-14, C-1, G-3, G-4, G-9, R-4, R-5, R-9, VIS-2, VIS-4).
- Action IDs restart at A-1; round-2 IDs are cited as R2-A-n.

## Coverage and method

- Viewport (javascript_tool): `window.innerWidth` 1624, `innerHeight` 1013, devicePixelRatio 2, `scrollWidth` 1624 (no horizontal scroll). Not covered: 1280, 1440 x 614, 1920 (the brief forbids resize_window; only this window was used).
- Themes: light (System) for the review and plant opening-view measurements; dark for the w-13 review, the live run's review checks, the dimmed-wire contrast and the hunt zoom series (42 % to 72 %). Light dimmed-wire contrast was not measured this round.
- Method: DOM of the React Flow graph. Wire paths are sampled every 1.5 canvas units and mapped to screen coordinates with `getScreenCTM`. Checks: (a) every instance node, frame and frame title inside the canvas rectangle; (b) wire samples inside the text boxes of frame names and count labels (leaf elements, not the whole header box); (c) minimum distance from a wire sample to any header text box and to any node text box; (d) wire-end pairs under 6 screen px where the two wires do not share a source port; (e) wire samples inside a node body.
- Correction to my own first pass: the first check used the whole header box, which is wider than the text and caught the vertical part of a wire that is 11 to 18 px right of "1 node". The text-box measure above is the acceptance rule of R2-A-5 ("name plus count label"), and it is the one reported here.
- Side effects, disclosed: (1) I started one real run: w-40 (Boiler steam loop, Explore, 30 min, Watch about 30 s, one fault: scaling on boiler-1 at minute 10). It is in the workspace; deleting it needs your say. (2) I opened the hunt editor once with no changes and left it without saving. (3) The theme was cycled and restored to System. (4) Tab 770933686 is mine and is closed at the end; tab 770933647 and the forbidden tab were not used. (5) No source was read, no git was run, nothing was fixed.

## Score

**6/10** (the lower of (a) live 6 and (b) plant 7). Previous 5, delta +1.

- (a) Live run display and review: **6/10** (unchanged). Four round-2 problems are now fixed in Chrome: the review clipping (R2-A-1), the shared wire ends (R2-A-2), the plan mark on the running node (R2-A-3) and the dimmed-wire contrast (R2-A-11). The review now opens with no readout value on any node (the fit moved the loss from clipping to hidden values), and fault causes and effector outcomes are still cut or absent on the node (R2-A-4 not done). Those two keep it at 6.
- (b) Static plant graph: **7/10** (up from 5). All 11 plant opening views pass the header and shared-end checks: zero wire samples in any name or count text box, and zero wire-end pairs under 6 px. The outline empty rows now read "0 · EMPTY" on one line. What stops 8 or 9: the office tower still routes two long wires round the perimeter (R2-A-7 not done); the overview tier wraps area headers so that the count reads as part of the name (hunt at 42 %); and the office tower opens at 49 %, one point under the tier boundary, so its port rows disappear.
- Evidence (screen, plant, theme, quoted label, value, viewport 1624 x 1013):
  1. Review w-36 Food cold chain, light, opens at 41 %: 13 instances and 4 frames all inside the canvas (256 to 1202 x 255 to 1013); 0 readouts visible (`display: none`); 0 wire-end pairs under 6 px.
  2. Review w-33 Multi-hall campus, light, opens at 35 %: 26 instances, 5 frames, 0 outside the canvas, 0 pairs under 6 px, 0 readouts visible. At 00:15:00 the node pmp-u3 reads "Effector call: set speed · applied · 360 ms", and its text box overflows (scrollWidth above clientWidth). ch-u2 reads "Fault" only; the cause "condenser fouling" is not on the node.
  3. Live run w-40 Boiler steam loop, sampled every 250 ms: boiler-1 reads no fault at 00:08:00 and "Fault · scaling" at the first sample that reads 00:10:00, and at every sample to 00:24:00 (top bar "Running Boiler steam loop · 00:10:00 of 00:30:00 · 60 s per s").
  4. Plant office-tower-hvac, light, opens at 49 % (tier class `tier-overview`): the nodes show id and kind only, with no port rows (for example "P-1 pump"). Boiler at 68 % shows port rows ("shift out", "demand in").
  5. Plant hunt-special-names, dark, at 42 % (tier-overview): frame header "Production hall A" wraps to "Production hall" / "A", and "1 node" wraps to "1" / "node". At 50 % (tier-compact) the same header is one line with "1 node" on the baseline.

## What holds

- Review opens fitted: no instance node, frame or frame title lies outside the canvas in w-36 (41 %), w-33 (35 %), w-40 (48 %) and w-13 (48 %). The Fit control gives the same view.
- Shared wire ends: zero end pairs under 6 screen px in those four reviews and in all 11 plants at their opening zoom (R2-A-2 and R2-A-6 met).
- Header clearance (R2-A-5 met in the text-box sense): at the opening zoom of 11 plants, zero wire samples inside any name or count text box. Minimum distance from a wire to header text: boiler 8.6 px (68 %), compressed-air 10.2 (72 %), food cold chain 9.8 (72 %), data centre 9.7 (72 %), packaging 11.8 (66 %), greenhouse 10.3 (57 %), hunt 10.2 (72 %), campus 45.9 (40 %), office 20.3 (49 %), power 8.4 (51 %), water 9.8 (72 %). Review w-13 13.9 px.
- Plan mark on the node during a live run (R2-A-3 met): see evidence 3. The fault mark is sampled, not the effector mark (see Not verified).
- Empty outline rows (R2-A-8 met, hunt editor): "PRODUCTION HALL 0 · EMPTY", "COMPRESSOR ROOM 0 · EMPTY", "CHILLER PLANT 0 · EMPTY", "SUBSTATION 0 · EMPTY", each one line, 28 px high.
- Dimmed wires in dark (R2-A-11 met): the dimmed stroke is rgb(149,157,171) at stroke-opacity 0.65, which over the canvas rgb(15,17,21) is about 3.6:1. Active power (242,193,78) and water (47,184,154) strokes read well.
- No wire crosses a node body, and no wire sample lies inside any node's text box at the opening zoom in any review or plant measured.
- Overview-tier fault mark: at 35 % the faulted node shows a red wrench mark and the word "Fault" (shape and colour, V-4 met at this tier).
- Timeline legend and marks in the review: "Fault starts" and "Effector call" legend (w-13), labelled marks "00:05:00 CP-1", "00:12:00 CP-2", "00:18:00 CP-3" (V-3 partly met).
- Hunt "Production hall A" at 72 %: the site-calendar dashed wire turns down in the gutter above the header, not through it.

## Trade-off: readouts hidden until 50 % in a review (judgement)

- Measured: the readout layer is `display: none` below 50 % (49 % hidden; 58 % and 59 % shown, 10 of 13 visible at 59 %). Opening views at 41 %, 35 %, 48 % and 48 % show zero readouts on every node. Port rows of plant nodes use the same boundary (hunt at 42 % has none, at 50 % has them).
- Judgement: **not acceptable as shipped.** A review exists to read what happened, and the per-node readouts are the only values on the graph. A user who opens w-33 sees the topology, the frame names and the fault mark, and no value. Fitting to the canvas (R2-A-1) fixed clipping and created this gap. The graph-only view cannot be read until the user zooms in past 50 %, and the first screen does not say so.
- Acceptance I would sign: lower the compact tier (readouts and port rows) from 50 % to 40 %; floor the on-screen size of readout and port text at 8 CSS px (canvas font at least 8 divided by the zoom); open a review at max(fit, 40 %), anchored on the first fault or effector node (the first area when there is none); keep Fit as the whole-plant view. Under that rule w-36 (fit 41 %) opens with values visible; only plants whose fit is below 40 % (the campus at 35 %) open cut, and then the anchor node is fully on screen. This also fixes the office tower (49 %, port rows back).
- Overview-tier text, separately: below 50 % the kind text is 6.4 CSS px and the id 7.6 px on the campus (canvas font 17.9 and 21.35 at the 0.355 zoom). It is legible at 2 x zoom and tiny at 1 x, so it needs the same 8 px floor.

## Findings

### F-1 A review opens with no readout value on any node
- Severity: High
- Where: review w-36 Food cold chain (41 %, light), w-33 Multi-hall campus (35 %, light), w-40 Boiler steam loop (48 %, light), w-13 Water pumping station (48 %, dark and light). Every node card shows id and kind only, for example "compressor / air compressor" in w-36.
- Observed: the readout layer is hidden below 50 % (`display: none`, 0 visible readouts in all four). The user sees the fault marks but no channel value until zoomed in past 50 %. Round-2 F-1 (clipping) is fixed; this is what the fix left behind.
- Expected: V-4 (live readouts at the playback time on the graph), R-4 (what the run is doing), and the review opening state of V-9 (opens fitted with nothing cut). Both cannot hold at once with the current 50 % threshold.

### F-2 Fault cause and effector outcome are cut or absent on the node
- Severity: Medium
- Where: review w-33 at 00:15:00, 35 %, light. pmp-u3: "Effector call: set speed · applied · 360 ms" (DOM text box overflows; cut on the node). ch-u2: "Fault" only; "condenser fouling" is not on the node.
- Observed: the cause and outcome sit in the Inspect panel only, not on the graph. The round-2 A-4 fix is not in this build (the fixer reports it skipped).
- Expected: V-4, VIS-4, round-2 R2-A-4 (no "…" on the node text).

### F-3 Office tower opens below the 50 % tier boundary and loses its port rows
- Severity: Medium
- Where: plant office-tower-hvac, light, opens at 49 % (`tier-overview`). Every node shows id and kind only (for example "P-1 pump", "CH-1 chiller large"). Compare boiler-steam-loop at 68 % and hunt at 50 %, which show port rows.
- Observed: one point of fit moves the plant across the tier boundary, and the channel and port names disappear on opening. The plant is readable at its kinds, but a user cannot see which port a wire lands on.
- Expected: P-2 (kind and id legible at the opening zoom), P-7 and C-1 (ports show a plain name), and the single tier rule across screens.

### F-4 Office tower still routes two long wires round the plant perimeter
- Severity: Medium
- Where: plant office-tower-hvac, 49 %, light (the same routing at 56 % in round 2). The dashed SC-1 wire runs down the left edge of the Site and Plant room frames (about x 280, y 300 to 775) and then along the bottom to AHU-2. The pink CH-2 wire runs down the right edge (about x 948, y 368 to 776) and back left along the bottom.
- Observed: the two wires enclose the plant frame and are hard to tell from it. Clearance to nodes is fine (no wire crosses a body); the shape is the problem.
- Expected: P-4 (wires routed between areas and readable), E-14, round-2 R2-A-7 (not done).

### F-5 Overview-tier area headers wrap, and the count reads as part of the name
- Severity: Medium
- Where: plant hunt-special-names, dark, 42 % (`tier-overview`). "Production hall A / 1 node" wraps to "Production hall" / "A" with "1" / "node" on a second line; "Production hall B / 1 node" wraps the same way. At 50 % (`tier-compact`) the header is one line.
- Observed: a 12 px header box holds a two-line name, so "1" sits on the name line and "node" on the next, and the reader cannot tell the count from the title.
- Expected: P-14 (area titles clear), G-3 (ids and names do not break inside a label), round-2 R2-A-9 at this tier.

### F-6 Overview-tier node text is too small to read at 1 x
- Severity: Low
- Where: review w-33 at 35 % and campus plant at 40 %, light. Kind text 6.4 CSS px and id 7.6 CSS px on the screen (canvas font 17.9 and 21.35 at zoom 0.355 to 0.40).
- Observed: legible at 2 x in a zoomed capture; at 1 x the kind line is a smudge.
- Expected: P-2 ("id and kind in words at the opening zoom"), VIS-2, G-3.

## Actions to reach 9

Live first (A-1 to A-3), then plant (A-4 to A-6). Each is presentation over data the binary and the run already carry; nothing here needs the binary.

### A-1 [live + plant] One readout and port tier from 40 %, with an 8 px floor, and open reviews at max(fit, 40 %) anchored on the event
- **Change:** move the compact tier (readout layer, port rows) from 50 % to 40 % (the constant in the tier rule, one place). Floor the on-screen size of readout, port and kind text at 8 CSS px: canvas font = max(current, 8 / zoom). Open a review at max(fit, 40 %), centred on the first fault or effector node (or the first area when none). Fit is unchanged and shows the whole plant.
- **Where:** review screen (opening view of every run) and plant and editor screens (the same tier rule, the office tower at 49 %).
- **Acceptance:** at 1624 x 1013, the first paint of w-36 shows a readout value on every node in view (readout `display` not none, innerText non-empty, font at least 8 CSS px). w-33 opens at max(35 %, 40 %) = 40 % centred on the first event node, and that node is fully inside the canvas. Fit then gives the 35 % whole-plant view. The office tower at its opening zoom shows port rows on every node. Measure: DOM text and computed font of the nodes listed.
- **Impact:** closes F-1 (High) and F-3, and F-6 by the floor. Expected live +1.5 (6 to 7.5), plant +1 (7 to 8).

### A-2 [live] Wrap the fault cause and the effector outcome on the node (R2-A-4)
- **Change:** the node's fault line wraps to a second line instead of cutting (for example "Fault · condenser fouling"), and the effector row wraps (for example "Effector call: set speed · applied · 360 ms"). The node grows by at most one line at each tier.
- **Where:** review and running nodes at every tier from 35 % up.
- **Acceptance:** w-33 at 00:15:00, at 35 % and at 50 %: ch-u2 innerText contains "condenser fouling" and pmp-u3 innerText contains "applied · 360 ms", and for each of these elements scrollWidth is at most clientWidth (no overflow, no "…").
- **Impact:** closes F-2. Expected live +1 (7.5 to 8.5).

### A-3 [live] Verify the effector mark on a live run (no code expected)
- **Change:** start a Watch run on boiler-steam-loop with one effector call at minute 10 on boiler-1 (for example stop). Sample every 250 ms.
- **Where:** plant screen during a live run.
- **Acceptance:** the blue effector mark and its outcome row appear within one sample after the minute, stay for 60 s of simulated time, and go away after. Record the timestamps.
- **Impact:** verifies the R2-A-3 effector half (not sampled in round 2 or this round). Expected live +0.5 (8.5 to 9) if it passes; if it fails, it is a live finding and the live score stays at 8.5.

### A-4 [plant] Area headers stay one line at every tier (R2-A-9 at overview)
- **Change:** the area header row is a single nowrap line: the name is ellipsised with the full name in a title attribute when it does not fit, and the count keeps its own position on the same line. Below 50 % the header does not wrap at all.
- **Where:** every area header in plant view, editor and review, below 50 %.
- **Acceptance:** hunt at 42 %: each frame title element is one text line (height at most the line height), and "Production hall A" and "1 node" are on the same baseline. Measure: getClientRects count of text lines per header.
- **Impact:** closes F-5. Expected plant +0.5 (8 to 8.5).

### A-5 [plant] Route the office tower wires in their own gutters, not round the perimeter (R2-A-7)
- **Change:** the router refuses a route that leaves the rectangle spanned by its source and target nodes, inflated by one gutter (48 canvas units). A wire that cannot be routed inside that rectangle takes the nearest free lane inside its area margin.
- **Where:** plant view and editor (office tower is the reference plant).
- **Acceptance:** office tower at its opening zoom: for every wire, every path point lies inside its source-target rectangle inflated by 48 canvas units (DOM check). The dashed SC-1 wire no longer runs down the left edge of the plant frame, and the pink CH-2 wire no longer reaches the bottom edge.
- **Impact:** closes F-4. Expected plant +1 (8.5 to 9.5, capped at 9).

### A-6 [verification] Check the untested widths and tiers
- **Change:** no code change expected. Repeat the review and plant checks at 1280 x 800 and 1920 x 1080 (or the nearest sizes the window allows), at 34 % on hunt (round-2 F-9, mid-word breaks), and the light-theme dimmed-wire contrast.
- **Where:** all review and plant screens.
- **Acceptance:** no horizontal scroll at 1280; the opening review at 1280 shows the same tier rule as A-1; no name breaks inside a word at 34 %; light dimmed contrast at least 3:1.
- **Impact:** no points on its own. It removes the ceilings listed below; a failure there would lower the score.

## Projected score after the actions

- Live: 6 now. A-1 (F-1, High) to about 7.5, A-2 (F-2) to about 8.5, A-3 (effector check) to about 9 if it passes. Live reaches 9 only with A-1, A-2 and A-3 all passing.
- Plant: 7 now. A-1 (F-3, F-6) to about 8, A-4 (F-5) to about 8.5, A-5 (F-4) to about 9. Plant reaches 9 only with A-1, A-4 and A-5.
- The dimension is the lower of the two, so 9 needs every action above. The projection is an estimate; only the findings marked "measured" in this report are measured.

## Not verified this round

- Widths 1280, 1440 x 614 and 1920 (no resize, per the brief).
- Live effector mark: the effector call's 60 s window is unit-tested only; the live run sampled the fault only.
- Light-theme dimmed-wire contrast (dark measured at 3.6:1).
- 34 % mid-word breaks (round-2 F-9): at 42 % and 50 % names break only at hyphens, so it is probably fixed at those tiers, not measured at 34 %.
- Editor canvas at 1624 x 1013 beyond the outline row check.

## Summary for the coordinator

- Score: 6/10 (live 6, plant 7; the lower is the dimension). Previous 5, delta +1.
- Round-2 actions: R2-A-1 (review fit) pass, but it created F-1; R2-A-2 (shared ends) pass; R2-A-3 (plan mark) pass for the fault (effector not sampled); R2-A-4 (wrap cause text) not done; R2-A-5 (header clearance, 11 plants) pass; R2-A-6 (shared ends in plant) pass; R2-A-7 (office routing) not done; R2-A-8 (outline rows) pass; R2-A-9 (hunt header) partly, the overview tier garbles it; R2-A-10 (overview names) partly, names break at hyphens at 42 % and 50 %; R2-A-11 (dimmed wires) pass in dark.
- Trade-off judgement: opening readouts hidden below 50 % is not acceptable as shipped. Accept with A-1 (40 % tier, 8 px floor, open at max(fit, 40 %) anchored on the first event).
- Findings: 6 total (High 1: F-1; Medium 4: F-2 to F-5; Low 1: F-6).
- Actions: 6 total (live A-1 to A-3; plant A-4 to A-6, A-6 is verification).
- Viewport: 1624 x 1013 only. Not covered: 1280, 1440 x 614, 1920.
- File: /Users/ghassan/my-projects/streams-simulator/docs/simulation-canvas/director-score-20261015/graph-and-review-round-3.md
