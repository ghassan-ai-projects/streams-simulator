# Director score, dimension: Information design of the plant graph and the review, round 4

- Director: UI/UX director (brief `brief-director-dim.md`), Chrome at http://127.0.0.1:8766.
- Previous score: 6 (round 3, `graph-and-review-round-3.md`: live 6, plant 7). Every round-3 action was re-checked in Chrome by me; the fixer's Status lines were not taken as evidence.
- Bar: `docs/simulation-canvas/UI-EXPECTATIONS.md` (P-2, P-4, P-7, P-11, P-14, E-14, V-3, V-4, V-6, V-9, V-14, V-15, C-1, G-3, G-6, G-9, G-17, R-4, VIS-2, VIS-4).
- Action IDs restart at A-1; round-3 IDs are cited as R3-A-n.

## Coverage and method

- Viewport (javascript_tool, measured at the start and again before finishing): `window.innerWidth` 1624, `innerHeight` 1013, devicePixelRatio 2, `scrollWidth` 1624 (no horizontal scroll), theme System (light by OS). Own tab 770933688; tab 770933647 was not used. No resize_window called.
- Not covered: 1280 x 720 or 800, 1440 x 614, 1920 x 1080 (the brief forbids resize_window; only this window was used).
- Method: DOM of the React Flow graph (`.react-flow__node-instance`, `.react-flow__node-frame`) read with javascript_tool. For each node I listed the leaf text boxes and checked pairwise box overlap (more than 2 px both ways), scrollWidth over clientWidth, and visible text size as computed font size times the zoom. Visual checks used zoom captures of the screen. Zoom was read from the toolbar label.
- Live run: I started two real runs on boiler-steam-loop (Watch, Explore, 30 min, seed 1, one `set running` effector call on boiler-1 at minute 10): w-41 (the first, sampled late) and w-42 (sampler armed before Start). Sampling every 100 ms through a javascript interval on the boiler-1 node text and the top-bar clock. Both runs are in the workspace; deleting them needs your say.
- Side effects, disclosed: two runs created (w-41, w-42). The theme was cycled to Dark and back to System. Seeking the review scrubber changed the open review tab (Report to Inspect) as a result, which is itself finding F-4. Tab 770933688 is mine and is closed at the end.

## Score

**5/10** (the lower of (a) live 5 and (b) plant 7). Previous 6, delta -1.

- (a) Live run display and review: **5/10** (was 6). The effector mark and the effector outcome now work in a live run (R3-A-3 and the effector half of R3-A-2 pass in Chrome). But the review's first screen is mostly dashes (33 of 46 readout cells at 00:01:00 on w-36), and the fault note on the node now overlaps the node's own kind line at the review's default zoom (boiler-1 on w-40 at 48 %, light and dark; ch-u2 on w-33 at 40 %). Round 3 had clipping; this is collision, and it lands on the one node the review is about. Two High findings and a repeated pattern on the main review screen keep it in the 5-6 band; the effector and anchoring work keep it above 4.
- (b) Static plant graph: **7/10** (unchanged). The plant opening views are clean: hunt at 72 % and 60 %, office tower at 52 % with port names on all 15 nodes, no horizontal scroll, the area headers are one line at 42 % and 35 %, the dark-theme text contrast is 6.05:1 (readout) and 5.91:1 (fault note). What stops 8 or 9: mid-word id breaks below 40 %, the "Production ..." header ellipsis that hides the A and B, a dashed wire that ends on the top edge of a node and runs behind it, and the office tower's dashed wires that still run along the plant-room frame edge.
- Evidence (screen, plant, theme, quoted label, measured value, viewport 1624 x 1013):
  1. Review w-36 Food cold chain, opens at 00:01:00 at 41 % (light, then dark): 13 nodes, 46 readout cells, **33 of them read "–"** (72 %); 13 cells hold a value. Chiller-1 at 14:58 shows seven "–" rows (load, chw flow, cond water in, tariff, heat rejected, chw supply, chw return) and Inspect lists four observed channels for the same node (chw supply temp 7.00 °C, chw return temp 9.50 °C, cond temp 31.9 °C, power 65.9 kW). The "–" is the visible form of a screen-reader-only "not observed" label, and the legend has no entry for it.
  2. Review w-40 Boiler steam loop at 00:10:00, 48 % (light and dark): boiler-1 reads "Fault · scaling" struck across its kind line "boiler" and touching "steam demand –". Measured overlap: "boiler" with "Fault · scaling", 97 x 10 px; "boiler" with "steam demand", 88 x 4 px. Same collision on review w-33 Multi-hall campus at 40 % (ch-u2, "chiller" with "Fault · condenser fouling", 79 x 8 px).
  3. Review w-40 at 40 % (one zoom step below the 48 % fit, label shows "40 %" and tier is overview): boiler-1 shows only the wrench chip "Fault"; the cause "scaling" exists in the DOM as a 0-px-wide note and a screen-reader label. Nothing visible explains the fault on the node.
  4. Live run w-42 (sampled every 100 ms, Watch): top-bar clock 00:10:00 with boiler-1 "Effector call" absent; 00:11:00 present with "Effector call: set running · accepted"; 00:12:00 and 00:13:00 present; 00:14:00 gone. Wall time from the first present sample to the first absent sample: 3.00 s (11.44 s to 14.44 s). Run w-41 gave the same hold (sampled late, its present samples at 00:11:00 to 00:13:00 and gone at 00:14:00).
  5. Plant hunt-special-names at 42 % (light): "Production ..." and "Production ..." both read "1 node" on the same line; at 35 % the ids read "compresso / r-1", "compresso / r-2", "productio / n-line-a", "packaging / -line-b" (mid-word breaks).

## What holds

- Review opens fitted or at 40 %, anchored on the first event: w-33 opens at 40 % with ch-u2 fully on screen (fit 35.5 %); w-40 and w-42 open at the 48 % fit with the whole plant visible; no node is outside the canvas at those openings except the campus frames noted in F-11.
- Timeline marks: red fault chips ("00:10:00 boiler-1") and blue effector chips ("00:15:00 pmp-u3") sit on the scrubber track, and the legend "Fault starts" and "Effector call" sits by the speed buttons. This is a clear first-five-seconds cue (V-3 met).
- Effector outcome wraps on the node: boiler-1 at 00:10:00 reads "set running · applied · 233 ms" on two lines with no clipping at 48 %; pmp-u3 on w-33 reads "set speed · applied · 360 ms" cleanly at 48 % (at 40 % its box overlaps the word "pump" by 3 px, see F-1). The blue square mark sits by the node id.
- Live run effector mark (R3-A-3): verified in two runs, see Evidence 4. The outcome text on the node is the binary's wording, and the run panel says "Every effector call was accepted. 00:10:00 · boiler-1 · set running: accepted".
- Live run copy is plain: "Running Boiler steam loop as run w-41 · 30 min of simulated time · Explore · no faults · 1 effector call"; "Stop now ends the run at the next step and keeps what ran as a shorter run you can review"; the finished toast reads "Run w-42 finished · Open review".
- Report panel: "DISTURBANCES AND EFFECTOR CALLS" is a real table with headers Time, Instance, Effector, Outcome (for example "00:10:00 | boiler-1 | set running | applied · 233 ms"); TRUE RESULTS is labelled "What really happened in the plant" and says when it was hidden (Evaluate) or visible (Explore). V-6 and V-13 hold.
- Area headers: one text line each at 42 % and 35 % (the name ellipsises, the count stays on the line). R3-A-4 holds for header line count.
- Plant opening views: all ports named at 52 % on office-tower-hvac ("heat out", "basin out", "supply out", "load in", "cond water"); the hunt plant at 72 % wraps names at hyphens only.
- Wire legend: "Wire colour and pin shape by port type", with the port families and "Port not recorded" / "Filled: connected" / "Hollow: not connected" (P-11 met).
- Dark theme: readout labels 6.05:1 and fault note 5.91:1 on the node background (WCAG AA for text).
- No horizontal scroll at 1624 px in the review or the plant.

## Findings

### F-1 The fault note sits on the node's kind line at the review's default zoom
- Severity: High
- Where: review w-40 Boiler steam loop at 00:10:00, boiler-1, 48 % (the fit, the review's default), light and dark. Same at 40 % on review w-33 Multi-hall campus (ch-u2, "chiller", with "Fault · condenser fouling"), light.
- Observed: the note "Fault · scaling" starts on the kind line and overlaps "boiler" (97 x 10 px box overlap, the red text runs through the word); the first readout row "steam demand" touches it (88 x 4 px). On ch-u2 the "chiller" word is struck through by "Fault · condenser" (79 x 8 px). At 48 % on ch-u2 the overlap shrinks to 3 px but is still present. The note reserves its wrapped lines below the header, but the header block does not move down to make room.
- Expected: V-14 (readout nodes show their full kind with even heights), V-4 (the fault marked on the node), R3-A-2 acceptance "no overflow, no truncation", and the principle "nothing truncated that matters". R3-A-2 measured text containment, not collisions, so it passed with a visible defect.

### F-2 Most readout cells read "–" on the review's first screen and stay "–"
- Severity: High
- Where: review w-36 Food cold chain, opening at 00:01:00, 41 % (light and dark); also w-40 and w-33 openings. Chiller-1 at 00:14:58 on w-36 (light).
- Observed: 33 of 46 readout cells show "–" at open (72 %), and the same count at 14:58. The "–" hides the DOM text "not observed" (a screen-reader-only label, `sr-only`), so for a sighted reader it is a dash with no meaning. Nodes list channels the run never reports (chiller-1 shows "load", "chw flow", "cond water in", "tariff", "heat rejected" as "–" at both 00:01:00 and 14:58; Inspect lists four observed channels with different names, for example "chw supply temp 7.00 °C" against the node's "chw supply –"). The Legend has no entry for "–". The review's status line reads "opened at 00:01:00, once every channel had reported", which a stranger will read as contradicting the dashes.
- Expected: V-4 (live readouts at the playback time on the graph), R-4 and G-6 (wording a user can read), G-17 (where values come from, here "no reading" versus "not measured"), UI-EXPECTATIONS V-5 (the observed channels in Inspect match the node).

### F-3 One zoom step below 40 % the review drops to a tier that hides the fault cause
- Severity: Medium
- Where: review w-40, boiler-1 at 00:10:00, light. The toolbar reads "40 %" (rounded; the tier switches below 0.40, measured at 0.396 after one zoom-out step from the 48 % fit). Tier: overview.
- Observed: readout rows, port rows and the fault note disappear. boiler-1 shows the wrench chip "Fault" and its kind "boiler" only. The note element "Fault · scaling" measures 0 px wide; the text "Fault: scaling" exists only as a screen-reader label. R3-A-2 said the cause is "also spoken text on the node at the overview tier": that is the screen-reader copy, not a visible cause. The step from 48 % to 40 % is a cliff, with no transition or label.
- Expected: P-2 (the id and kind legible at the opening zoom), V-4 (fault visible on the node), R3-A-2 (cause visible on the node at 35 % and above, as the round-3 finding F-2 asked).

### F-4 Seeking to a mark swaps the Report for Inspect without a cue
- Severity: Medium
- Where: review w-33 at 00:15:00 (click on the "00:15:00 pmp-u3" chip) and review w-40 at 00:10:00 (click on "00:10:00 boiler-1"), both themes.
- Observed: the right panel changes from "Report" to "Inspect · pmp-u3" (or "Inspect · boiler-1") and the report with the disturbances, effector calls and KPI summary is one tab away. The user who came to read what happened loses the summary at the first click on a mark. The Inspect tab has no explanation that it was opened by the click.
- Expected: V-6 (the report leads with what happened), V-3 (clicking a mark seeks and selects its instance, not replaces the report), G-11 (navigation keeps context).

### F-5 The effector mark appears one run minute after its call
- Severity: Low
- Where: live run w-42 (and w-41), boiler-1, set running at minute 10, Watch.
- Observed: the call at 00:10:00 is not on the node at the 00:10:00 tick (sample at 10.44 s: no "Effector call"). It first appears at the 00:11:00 tick (11.44 s) with "Effector call: set running · accepted". The R3-A-3 acceptance ("appear within one sample after the minute") is met, but a stranger watching the clock at 10:00 sees no call.
- Expected: VIS-4 and R-4 (faults and calls ticked off as they happen), RUN-4 (acknowledged at once).

### F-6 Overview-tier ids break mid-word below 40 %
- Severity: Medium
- Where: plant hunt-special-names at 35 % (light, zoom label "35 %"), and the campus plant at the same tier.
- Observed: "compressor-1" and "compressor-2" break as "compresso / r-1" and "compresso / r-2"; "production-line-a" breaks as "productio / n-line-a"; "packaging-line-b" breaks as "packaging / -line-b". At 42 % and 50 % names break only at hyphens. This is the known ceiling (round 3 said below 34 %); it already shows at 35 %, which is the fit of the campus and the level a user reaches by zooming out from most plants.
- Expected: G-3 (ids do not break inside a label), P-2 (id legible at the opening zoom), R2-A-10 (overview names).

### F-7 Two area headers read the same at 42 %
- Severity: Medium
- Where: plant hunt-special-names at 42 % (light), headers "Production hall A" and "Production hall B", each with "1 node".
- Observed: both read "Production ..." with "1 node". The ellipsis removes the only word that tells them apart. The full name is in a title attribute (hover only).
- Expected: P-14 (area titles clear), G-3 (names are not cut where the difference lies), R3-A-4 (name ellipsis is accepted, but the distinguishing part must stay).

### F-8 A dashed wire ends on the top edge of a node and disappears behind it
- Severity: Medium
- Where: plant hunt-special-names at 35 % (light), the dashed link from site-calendar to production-line-a. The link list says it lands on "shift in" (Inspector: "site-calendar shift out → production-line-a shift in").
- Observed: the dashed route goes down from site-calendar, left and down to the top of Production hall A, then its arrowhead sits on the top edge of the production-line-a node, above the header, not on the left-side "shift in" pin. The rest of the wire is hidden by the node body.
- Expected: P-4 (wires readable, routed between areas without crossing node text), E-14 ("every wire visibly ends on its two pins and never passes behind a node body").

### F-9 Office tower wires still hug the plant-room edge and share a lane at CH-2
- Severity: Medium
- Where: plant office-tower-hvac at 52 % (light; dark checked too).
- Observed: the dashed SC-1 and W-1 wires run down the left edge of the Plant room frame (about x 258 to 275, y 205 to 640 on screen) before turning into Floors 1-3, so they read as the frame's border. The AHU-2 to CH-2 wire runs up the plant-room body at about x 862 beside CH-2's left edge, and the red "load in" wire and the green "flow in" wire share a vertical about 3 to 4 px apart at CH-2 (zoomed). The fixer's R3-A-5 check (wires inside their source-target rectangle) passes; the picture still reads as a tangle at the chiller pins.
- Expected: P-4 (wires readable), P-14 (wires not on frame edges), E-14 (every wire visibly ends on its pin). The chillers-to-air-handler wire through the plant room, named in round 3 as known, is judged here: it crosses the plant-room body and is a Medium issue, not an excused one.

### F-10 Readout label cut in the review: "steam pres…"
- Severity: Low
- Where: review w-40 and w-42 at 48 %, boiler-1 at 10:00 (light and dark). Also "steam pressure" in w-40 at 40 % by the fixer.
- Observed: the readout label is clipped to "steam pres…" with "998 kPa" beside it; the full label is only in the title attribute.
- Expected: R3-A-2 (labels may end in an ellipsis only when the cause and outcome are complete; labels are not exempt from "nothing truncated that matters").

### F-11 A review opened at 40 % on the campus opens cut with no cue
- Severity: Low
- Where: review w-33 Multi-hall campus at 00:01:00, 40 % (fit 35.5 %), light.
- Observed: frame "Hall 1" runs past the right canvas edge (right at 1302 px, canvas ends at 1202 px), and frames "Hall 2" and "Hall 3" run below the canvas (bottom at 1047 px, canvas at 1013 px). No edge cue says more is there, except the navigator minimap in the bottom left. The Fit control is the only way to see it all. This is the designed opening (R3-A-1), but a stranger does not know the plant is bigger than the screen.
- Expected: V-9 (opens fitted, nothing cut), G-10 (nothing important clipped), R3-A-1 (the anchor is fully on screen; the rest should show a cue).

### F-12 The zoom and fit controls are icon-only
- Severity: Low
- Where: review and plant toolbars, "Zoom out", "Zoom in", "Fit the plant in view" (aria-labels only; the buttons have no text and no `title`, so no hover tooltip is shown on the button itself; I did not hover-test every one).
- Observed: three icon-only buttons; the Legend button has text.
- Expected: V-15 ("header and playback controls carry text or visible tooltips").

## Actions to reach 9

Ordered by points per effort. Live first (A-1 to A-5), then plant (A-6 to A-9), then polish (A-10). Each is a presentation or routing change over data the binary and the run already carry; I did not read the source (per the brief), so where a change touches the router or the node layout, the fixer must locate the owner.

### A-1 Lay the fault note inside the node's flow, below the kind line, and push the rows down by its height
- **Change:** at every tier from 40 % up, the node's header block is `id`, then `kind`, then the fault note in normal flow (not absolutely placed over the kind line). The rows below start after the note's measured height (one line for "Fault · scaling", two lines for "Fault · condenser fouling" at the compact tier). The wrapped text keeps its 8 CSS px floor and sits in the node's own box.
- **Where:** review and running nodes at 40 % to 106 %, the review graph of w-40 (boiler-1 at 00:10:00) and w-33 (ch-u2 at 00:10:00).
- **Acceptance:** in Chrome at 1624 x 1013, light and dark, on w-40 at 00:10:00 at 48 % and at 40 %, and on w-33 at 00:10:00 at 40 %: for every pair of leaf text boxes in the node, no overlap greater than 1 px in either direction (the same check as F-1's measure). The fault note's top is at or below the bottom of the kind text.
- **Impact:** closes F-1 (High) and the visible half of F-3. Live 5 to 7.

### A-2 Show only channels the run observes, and key the rest
- **Change:** a node lists the channels that the run actually reports (the same set Inspect uses); channels the run never reports are dropped from the node and counted in "+N more in Inspect". A row with an observed channel but no value at the playback time reads "–" and the Legend gets one line: "– no reading yet at this time". The review status line replaces "once every channel had reported" with "readings start at 00:01:00" (wording, no jargon).
- **Where:** review graph at every opening zoom (w-36, w-40, w-33 and the other review runs); Legend popover.
- **Acceptance:** w-36 at 00:01:00 at 41 %: the count of "–" readout cells is at most 1 in 4 of the readout cells (from 33 of 46), and for chiller-1 every row name matches a channel in the Inspect list at the same time. The Legend lists "–". Measure with the same DOM count.
- **Impact:** closes F-2 (High). Live 7 to 8.5.

### A-3 Show the cause on the node at the overview tier
- **Change:** below 40 % the node keeps the fault chip and shows the cause as one line (for example "scaling" in the chip's row, or "Fault · scaling" in 8 CSS px), truncated with the full cause in the title. Readouts may stay hidden below 40 %.
- **Where:** review nodes at 35 % to 39 % (w-40 at 39.x %, w-33 at 35 %).
- **Acceptance:** w-40 at 00:10:00 at 39 % and w-33 at 00:15:00 at 35 %: the node's visible text (not sr-only, width above 0 px) contains "scaling" and "condenser", and the element has a computed opacity of 1 and width above 0.
- **Impact:** closes F-3. Live 8.5 to 9 with A-1 and A-2.

### A-4 Keep the Report when a mark is clicked
- **Change:** clicking a mark on the scrubber (or a chip) seeks the playback and selects the instance on the graph, but leaves the active right-hand tab as the user had it. If the Inspect tab is opened, a one-line note at the top of the panel says "Opened by the mark at 00:10:00. Report is one tab away." When the user is on Report, it stays on Report.
- **Where:** review right panel (Report and Inspect tabs), any mark on the scrubber.
- **Acceptance:** on w-33, click the "00:15:00 pmp-u3" chip while on Report: the Report tab stays active, pmp-u3 is selected on the graph, and the Report's DISTURBANCES and EFFECTOR CALLS tables remain on screen. Repeat on w-40 at 00:10:00.
- **Impact:** closes F-4 (Medium). First-five-seconds read improves. Live +0.5 (capped; live reaches 9 with A-1 to A-4).

### A-5 Show the time and mark cue on the first screen of a review: one line under the scrubber
- **Change:** at 00:00:00 to the first event, the scrubber line under the top bar reads "Next: fault at 00:10:00 on boiler-1 (Play or click the mark)"; with no events it reads "No faults or calls in this run. Read the true results on the right." This replaces "No events in this run" and gives the stranger the first step.
- **Where:** review top bar, under the scrubber, at open.
- **Acceptance:** on w-40 at open: the line names the first mark with its time and instance; on w-36 at open: the line reads "No faults or calls in this run". Both in light and dark.
- **Impact:** supports the first-five-seconds read (V-2, G-6). Live +0.5 (part of the 9).

### A-6 Keep the distinguishing part of a header name when it is ellipsised
- **Change:** when a frame or area header name is too long at the overview and compact tiers, the ellipsis goes in the middle, not at the end ("Production … A" or "Prod… hall A"), and the full name stays in the title attribute.
- **Where:** plant and review frame headers at 42 % and below (hunt-special-names: "Production hall A" and "Production hall B").
- **Acceptance:** on hunt-special-names at 42 %: the two header texts differ in their visible characters (not only the title), and each still fits on one line. Both themes.
- **Impact:** closes F-7 (Medium). Plant +0.5.

### A-7 Stop ids breaking mid-word below 40 %: shrink the text, not the word
- **Change:** at the overview tier (below 40 %), an id is set on one line and its font shrinks (down to the 8 CSS px floor) before it wraps; wrapping may happen only at a hyphen or underscore. If it still does not fit at 8 px, the id is ellipsised with the full id in the title.
- **Where:** plant and review nodes at 35 % (hunt-special-names, the campus).
- **Acceptance:** on hunt-special-names at 35 %: "compressor-1", "compressor-2", "production-line-a" and "packaging-line-b" each render as text whose line breaks, if any, fall after a hyphen (DOM check: no line boundary between two alphanumerics that are not separated by a hyphen).
- **Impact:** closes F-6 (Medium). Plant +0.5.

### A-8 Route the dashed site-calendar wire into the shift-in pin, not the node's top edge
- **Change:** a wire whose target pin is on a node's left side is routed to enter that pin from the left, through the gutter, and never ends on a node's top or bottom edge. A wire that cannot reach its pin without entering a node body takes the lane above the area frame, not behind the node.
- **Where:** hunt-special-names (site-calendar to production-line-a "shift in"), and the same rule in the plant and the editor.
- **Acceptance:** on hunt-special-names at 35 % and at 72 %: for every wire, the last path point lies within 1 px of its target pin's centre, and no wire sample lies inside any node's rectangle except its endpoints (DOM check, the same method as round 3).
- **Impact:** closes F-8 (Medium). Plant +0.5.

### A-9 Office tower: keep long wires out of the frame edges and give pins their own lanes
- **Change:** the router keeps wires that run beside a frame edge at least 12 px from it (or moves them to the next gutter), and gives wires that end on neighbouring pins of the same node distinct lanes of at least 6 px apart before they reach the pin. The AHU-2 to CH-2 wire is routed in the gutter between the Plant room and the Floors frames, not up the plant-room body beside CH-2.
- **Where:** office-tower-hvac plant, opening zoom 52 %, and the editor on the same plant.
- **Acceptance:** at 52 %, light and dark: no wire sample lies within 12 px of a frame edge for more than 40 px of its length; wire samples at each CH-2 input pin are at least 6 px apart; the AHU-2 to CH-2 wire does not pass inside the Plant room frame's rectangle except in its gutter. Measured by the round-3 method.
- **Impact:** closes F-9 (Medium). Plant +1 (7 to 8).

### A-10 Polish for 9: readout labels, campus cue, control labels, effector lag
- **Change:** (a) readout labels keep their full text at 48 % by wrapping to two lines in the row (not ellipsis) or by a minimum label width (F-10); (b) a review that opens cut (campus at 40 %) shows an edge cue on the cut sides, for example a faded edge and "more to the right / below" (F-11); (c) the zoom and fit buttons get a visible text label or a tooltip on hover (F-12); (d) an effector call appears on the node at its own minute tick, not one tick later (F-5), or the call's minute is shown in its outcome row ("10:00 set running applied").
- **Where:** review and plant toolbars, review nodes, the campus review opening.
- **Acceptance:** (a) boiler-1 "steam pressure" fully visible at 48 % and 41 %; (b) the campus at 40 % shows the cue on the right and bottom sides; (c) hovering "Zoom in" shows the word "Zoom in" or a labelled control; (d) on a live run, the mark is on the node at the 10:00 sample (or the outcome row carries "10:00").
- **Impact:** closes F-10, F-11, F-12 and F-5 (all Low). Needed to reach 9 (trivial polish left is a 9).

### Not fixable in the UI: none identified from the canvas
- No action is blocked by the binary as far as I can see from Chrome. The routing (A-8, A-9) may live in a layout module the fixer must locate; I did not read the source, per the brief.

## Projected score after the actions

- Live: 5 now. A-1 (F-1) to about 7, A-2 (F-2) to about 8.5, A-3 (F-3) to about 9 with A-1 and A-2. A-4 and A-5 hold the 9 against regressions. Live reaches 9 only with A-1, A-2, A-3 all passing.
- Plant: 7 now. A-6 (F-7) and A-7 (F-6) to about 8, A-8 (F-8) to about 8.5, A-9 (F-9) to about 9. Plant reaches 9 only with A-6, A-7, A-8 and A-9.
- The dimension is the lower of the two: 9 needs every action above. Projections are estimates; the measured numbers are in Evidence.

## Not verified this round

- Widths 1280, 1440 x 614 and 1920 (no resize, per the brief). Round 3 reported no horizontal scroll at 1280 and 1920 by its own script; I did not repeat it.
- Light-theme dimmed-wire contrast (round 3 measured 3.56:1; not repeated).
- Hover tooltips on the icon-only buttons (F-12): aria-labels confirmed, hover not tested on each.
- The review of w-41 at its own 00:10:00 (sampled late, see F-5); the second run gave the full evidence.
- Editor canvas at 1624 x 1013.

## Summary for the coordinator

- Score: 5/10 (live 5, plant 7). Previous 6, delta -1.
- Round-3 actions, re-checked by me:
  - R3-A-1 (40 % tier, readouts on every node, open at max(fit, 40 %)): pass for the open zoom and anchoring; the readouts are present but 33 of 46 read "–" at open (F-2).
  - R3-A-2 (wrap cause and outcome): outcome wraps and passes (boiler-1 "set running · applied · 233 ms"); cause wraps but collides with the kind line at 40 % and 48 % (F-1) and disappears at the overview tier (F-3). Partial.
  - R3-A-3 (live effector mark): pass. Held about 3.0 s wall in two runs; appears one minute tick after the call (F-5, Low).
  - R3-A-4 (area header one line): pass for line count at 42 % and 35 %; the ellipsis hides A and B (F-7).
  - R3-A-5 (office routing): partial. The rectangle rule holds; wires still run along the plant-room edge and share lanes at CH-2 (F-9).
  - R3-A-6 (untested widths and tiers): 1280, 1440 and 1920 not covered by me. The 34 % mid-word break is confirmed at 35 % (F-6).
- Findings: 12 total. High 2 (F-1, F-2); Medium 6 (F-3, F-4, F-6, F-7, F-8, F-9); Low 4 (F-5, F-10, F-11, F-12).
- Actions: 10 total (live A-1 to A-5; plant A-6 to A-9; polish A-10).
- Viewport: 1624 x 1013, devicePixelRatio 2. Not covered: 1280, 1440 x 614, 1920.
- Runs created: w-41 and w-42 (boiler-steam-loop, Watch, one set running call on boiler-1 at minute 10).
- File: /Users/ghassan/my-projects/streams-simulator/docs/simulation-canvas/director-score-20261015/graph-and-review-round-4.md
