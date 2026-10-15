# Director score, dimension: Information design of the plant graph and the review, round 2

- Director: UI/UX director (brief `brief-director-dim.md`), Chrome at http://127.0.0.1:8766.
- Previous score: 4 (round 1, `graph-and-review-round-1.md`: live 5, plant 4).
- Bar: `docs/simulation-canvas/UI-EXPECTATIONS.md` (P-2, P-4, P-14, E-14, V-3, V-4, V-9, V-14, G-4, G-9, R-4, R-5, R-9, VIS-2, VIS-4).
- Action IDs in this round restart at A-1; round-1 IDs are cited as R1-A-n.

## Coverage and method

- Viewport (javascript_tool): innerWidth 1624, innerHeight 1013, devicePixelRatio 2, scrollWidth 1624 (no horizontal scroll). Round 1 was 1474 wide. Not covered: 1280, 1440x614, 1920 (the brief forbids resize_window, so only this window was used).
- Themes: System (light) for most screens; dark for the w-13 review at 00:01 and 00:05. The plants were read in light only.
- Plants read at their opening zoom: all ten real plants plus the Hunt test plant (boiler 74 %, compressed-air 72 %, water pumping 72 %, packaging 70 %, greenhouse 57 %, power distribution 54 %, data centre 72 %, office tower 56 %, food cold chain 72 %, campus 40 %, hunt 72 %). Editors: hunt and compressed-air (both left with "No changes").
- Runs read: w-13 (water pumping, Evaluate, light and dark), w-33 (campus, Evaluate, light), w-36 (food cold chain, Explore, light), w-38 (boiler, Explore, one fault, started by me), w-39 (food cold chain, Explore, no faults, started by me).
- Measurement: DOM of the React Flow graph. Each wire path is sampled along its length and mapped to screen coordinates. Clearances are reported in canvas units (screen px divided by zoom), ignoring the first and last 18 units of each wire (the pin stub). Earlier round-2 notes that counted 4 px at 50 % were the router's 8-unit rule at 50 %, so they are not failures; the figures below use canvas units. Other checks: wire samples inside a node body, over leaf text (node text, frame names, count labels), inside a frame title box, and wire-end pairs within 6 px on screen.
- Side effects, disclosed: (1) I started two real runs. w-38 (Boiler steam loop, Explore, 30 min, one fault: scaling on boiler-1 at minute 10, Watch about 30 s) for the running-state check, and w-39 (Food cold chain, Explore, 30 min, no faults) to catch the finish status line, which I had missed on w-38. Both remain in the workspace; deleting them needs your say. (2) A placeholder background-task chip was created by mistake and withdrawn at once. (3) The theme was cycled and restored to System. (4) Tab 770933677 is mine and is closed at the end; tab 770933647 and the brief's forbidden tab were not used.

## Score

**5/10** (lower of (a) live 6 and (b) plant 5). Previous 4, delta +1.

- (a) Live run display and review: **6/10**. The round-1 fixes hold where I measured them: no wire passes through a node body in any review, no wire crosses readout text at 00:05 (light and dark) or 00:18, effector calls mark their node, the finish line appears once, and progress shows speed and fault ticks. Against it, every review opens at 50 % with the right-hand frames and one whole node cut behind the Report panel (V-9 fails), wire ends merge on shared input rows, and fault and effector text is cut on the node.
- (b) Static plant graph: **5/10**. Wires cross a frame title or count label at the opening zoom on 7 of the 11 plants read (packaging, data centre, food cold chain, office tower, greenhouse, boiler, hunt). The editor outline still wraps its empty areas. The office tower wires route round the plant perimeter. A-10 (the hunt title clearance) fails below 64 %. What works: opening zoom fits, the zoom anchor, kinds in full on the campus, and no body crossings.
- Evidence (screen, plant, theme, quoted label, value):
  1. Review w-36, 50 %, light: "Room B" and "Power" cut at the canvas edge; the four wires into meter-main "loads in" end at x 1268 (viewport), under the Report panel, so the node is not on screen at the opening view. Fit (toolbar) gives 41 % with every node visible.
  2. Review w-33, 50 %, light: 10 wire-end pairs 3 px apart on the "loads in" rows of mtr-h1, mtr-h2, mtr-h3 and mtr-u1.
  3. Plant packaging-factory, 70 %, light: the pink wire crosses "3 nodes" of Line 1; the brown wire crosses "3 nodes" of Line 2; the dashed wire crosses "1 node" of Warehouse.
  4. Editor hunt, 72 %, light: "PRODUCTION HALL" and "No equipment yet" each break onto two lines in a 220 px rail; the four empty areas run at 28 px pitch with "yet" beside the next heading.
  5. Run w-38 (my run), boiler plant, light: at 00:13 "Running · 00:13:00 of 00:30:00 simulated · 60 s per s" and a fault tick at the 10-minute point on the 6 px bar; boiler-1 shows no fault chip on the graph at any point of the run.

## What holds

- A-1 (effector mark on the node): CP-3 in w-13 shows the blue square and "stop · applied · 371 ms" at 00:18:00; the mark is gone by 00:19:08 and at 00:00:32. pmp-u3 in w-33 shows "set speed · applied · 360 ms" at 00:15:00 and not at 00:16:06 (after the 60 s window).
- Review readouts at the playhead: no wire sample over any readout text at 00:05:00 (light and dark) or 00:18:00 (light) in w-13.
- No wire passes through a node body in any review measured (w-13, w-33, w-36, w-38) or in any plant at its opening zoom.
- Node clearance: every wire keeps at least 8 canvas units from node cards outside its pin stub, in every plant and review measured.
- The finish status: one line "Run w-39 finished · Open review"; the panel no longer repeats "finished".
- Live progress: the text updates about once per second ("Running · 00:04:00" to "00:13:00" to "00:18:00", then "00:24:00"), with the speed shown in the panel and the top bar ("79 s per s", "60 s per s"), a Stop now control, and the fault tick on the bar.
- Opening zoom fits the plant (hunt 72 %, boiler 74 %, data centre 72 %, food 72 %); the Fit button restores the opening view exactly.
- Campus opens at 40 % with every kind in full (no "…"), and power distribution shows "production line" in full.
- Review header: "Verified: the files match and the replay reproduces." on one line, and the Report leads with disturbances and effector calls.

## Findings

### F-1 Every review opens clipped behind the Report panel
- Severity: High
- Where: all reviews at this window (1624 x 1013, light). w-36 Food cold chain: the "Room B" and "Power" frames are cut at the canvas edge (x 1202); the four wires into meter-main "loads in" end at x 1268, behind the panel, so meter-main is not on screen. w-38 Boiler: the "Process hall" title is cut at the edge. w-13 Water pumping: "Intake" is cut at the edge. The review opens at 50 %.
- Observed: the opening view does not fit the space left of the Report panel. The Fit control (toolbar, third icon) fits the whole plant at 41 % in the same window.
- Expected: the review graph opens fitted with no clipping (V-9). Round-1 A-6 fixed this for the plant and editor, not the review.

### F-2 Wire ends share one input row and cannot be told apart
- Severity: Medium
- Where: review w-33 (campus, 50 %, light): at least 10 pairs of ends 3 px apart on the "loads in" rows of mtr-h1, mtr-h2, mtr-h3 and mtr-u1. Review w-38 (boiler, 50 %): meter-gas "loads in" receives two wires 3 px apart. Plant views at their opening zoom: hunt 72 % (6 pairs), compressed-air 72 % (6 pairs), power distribution 54 % (8 pairs), boiler 74 % (2), food cold chain 72 % (2), packaging 70 % (2).
- Observed: several wires land on one pin and their arrowheads stack. Round-1 A-3 said two ends on a row are spaced 6 px apart; the router spaces them 6 canvas units, which is 3 px at 50 %. In w-36 the chiller-1 example from round 1 cannot be reproduced: only pump-chw reaches chiller-1 (flow in), and no "load in" wire exists.
- Expected: each wire ends on its own pin row, at least 6 px apart on screen (A-3; E-14, P-4).

### F-3 Fault and effector text is cut on the node
- Severity: Medium
- Where: review w-33 at 00:15:00, 50 %, light: ch-u2 shows "Fault · condenser fou…" and pmp-u3 shows "set speed · applied · 3…" in its header area.
- Observed: the cause and outcome are the only fault and effector facts on the node during playback. The Inspect panel has them in full, but a user reading the graph cannot read the cause.
- Expected: V-4 (fault marked and readable on the node), VIS-4.

### F-4 The graph shows no run plan while a live run is in progress
- Severity: Medium
- Where: run w-38 (boiler plant, Explore, one fault at minute 10 on boiler-1), plant view at 74 %, light, sampled at 00:04, 00:13 and 00:18 of the run.
- Observed: the only running indicators are the top-bar pill, the toolbar chip "Running · 00:13:00" and the panel. boiler-1 has no fault chip at minute 10 nor any mark for its planned fault, although the fault is in the plan and the run passed its minute. The review shows the fault chip afterwards.
- Expected: R-4 (the graph shows a running indicator and the faults already injected), V-3 and V-4 (marks on the node). The plan is known before the run starts, so this needs no binary change.

### F-5 Wires cross frame titles and count labels at the opening zoom on 7 of 11 plants
- Severity: Medium
- Where: packaging-factory 70 %: pink wire through "3 nodes" of Line 1; brown wire through "3 nodes" of Line 2; dashed wire through "1 node" of Warehouse. Data centre cooling 72 %: pink wire through "3 nodes" of Power and "1 node" of Hall A. Food cold chain 72 %: brown wire through "1 node" of Room A; pink wire through "2 nodes" of Power. Office tower 56 %: pink wire through "1 node" of Floors 1-3 and of Floors 4 to 6. Greenhouse 57 %: wire through "2 nodes" of Bay 1. Boiler 74 %: wire through "3 nodes" of Distribution. Hunt 72 %: wire through "1 node" of Production hall A.
- Observed: the router's title obstacle covers the area name only. The count label in the same header row is not an obstacle, so wires run through it.
- Expected: area frame titles are never crossed by wires (P-14). Also the plant and the review differ: no review showed a count-label crossing.

### F-6 Office tower wires route round the plant perimeter
- Severity: Medium
- Where: office tower HVAC, 56 %, light. The dashed wire from SC-1 runs down the left edge of every area (x about 234, from y 192 to 760) and along the bottom of the plant to AHU-2. The pink wire from CH-2 runs down the right edge (x about 998) to the bottom and back. 23 wires share the same column gutters.
- Observed: the two long wires enclose the plant, so they are hard to tell from the plant frame. Every wire keeps 8 canvas units from node cards (no clearance fault), so the issue is routing shape, not spacing.
- Expected: P-4 (wires are readable and routed between areas), E-14.

### F-7 The editor outline still wraps its empty areas
- Severity: Medium
- Where: editor of Hunt & "Ålborg" pumps (72 %, light) and of Compressed-air works (72 %, light), outline rail on the left. The four empty areas (PRODUCTION HALL, COMPRESSOR ROOM, CHILLER PLANT, SUBSTATION in hunt) render as: "PRODUCTION / HALL" and "No equipment / yet"; "COMPRESSOR / ROOM"; "CHILLER / PLANT"; "SUBSTATION" stays on one line.
- Observed: the rows are at 28 px pitch but each two-line label overflows into the next, so "yet" reads as part of the next heading. The first group is SITE (2 in hunt, 1 in compressed-air), not COMPRESSOR HOUSE as round-1 A-7's acceptance says. The populated-first order holds.
- Expected: G-4 and round-1 A-7 (one line per empty area, at most 28 px high).

### F-8 Production hall A is still crossed at 34 % to 63 %
- Severity: Medium
- Where: hunt plant, light. The dashed site-calendar wire passes through the "hall A" name at 34 % (distance 0) and through the "1 node" label at 48 % and 63 %. At 63 % the wire is 16 px from the name, but it still crosses the count. At 48 % "Production hall B" and "1 node" run together with no gap ("Production hall B1 node").
- Observed: round-1 A-10 is met only at 72 % (11.3 px) and 100 % (9.5 px).
- Expected: P-14 and round-1 A-10 (at least 8 px at 30 %, 64 % and 100 %).

### F-9 At 34 % names break mid-word and wires cross the kind text
- Severity: Low
- Where: plant overview tier, 34 %: hunt "producti / on-line-a", "packagin / g-line-b"; compressed-air "compress / or-2" and "receiver / -header-1". Wires cross "air compressor" in compressor-1 and compressor-2 and "air receiver" in receiver-header-1.
- Observed: names break inside a word and wires pass over kind text.
- Expected: G-3 (ids are not broken), E-14, P-4 (no wire over node text), VIS-2.

### F-10 Grey wires are faint in dark
- Severity: Low
- Where: review w-13, dark, 00:05:00. Grey and olive wires almost disappear against the dark canvas; the teal and orange ones read.
- Observed: wire colours with lower contrast disappear in dark.
- Expected: G-9 (3:1 for state colours). Reported here because wires are part of reading the graph.

### F-11 A dashed wire crosses the "Hall 1" title band in the campus
- Severity: Low
- Where: multi-hall campus, 40 %, light. The cal-site to pl-h1 wire passes through the header row of "Hall 1" (28 samples) to the right of the name, not through the text.
- Observed: the header row is crossed but the name is clear.
- Expected: P-14 (area frame titles are not crossed).

## Actions to reach 9

Live first (A-1 to A-4), then plant (A-5 to A-11). Each change is presentation over data the run and the binary already carry; the canvas computes no binary-owned fact.

### A-1 [live] Open the review fitted to the space left of the Report panel
- **Change:** on open, use the Fit zoom computed for the canvas width minus the Report panel (at 1624 px the fit is about 41 %). Later zoom changes stay the user's. The graph is then fully on screen at open.
- **Where:** review screen, opening view of every run.
- **Acceptance:** in w-36, w-38 and w-13 at 1624 x 1013 the first paint is at the Fit zoom, and every node and title rectangle lies inside the canvas (DOM check: right edge at most the canvas right, bottom at most the canvas bottom). Fit then gives the same view.
- **Impact:** closes F-1 (High). Expected +1.5 on live.

### A-2 [live] Spread wire ends that share an input row
- **Change:** when k wires end on one input row, place their ends along the row 12 canvas units apart (6 px at 50 %). If the row is too short, the row grows to k slots. Each wire keeps its own pin.
- **Where:** review graph and plant views (the same router).
- **Acceptance:** in w-33 and w-38 at 50 % every pair of wire ends that does not share a source port is at least 6 px apart (DOM check of path end points). In the hunt plant at 72 % the same holds.
- **Impact:** closes F-2 on the live side. Expected +1 on live.

### A-3 [live] Carry the run plan onto the node while the run is in progress
- **Change:** when the run clock reaches a planned fault, the node shows the Fault chip and its cause (the same as in the review), and it keeps the chip. When it reaches an effector call, the node shows the blue square and the outcome row for 60 s of simulated time, as the review does.
- **Where:** plant view during a live run (top-bar and panel progress already exist).
- **Acceptance:** start a run with one fault at minute 10 on boiler-1 (as in w-38). Sampling each second while the clock passes 10:00, boiler-1 shows the Fault chip within one sample after 10:00 and keeps it to the end of the run.
- **Impact:** closes F-4 and the node-side part of R-4. Expected +1 on live.

### A-4 [live] Show the full fault cause and effector outcome on the node
- **Change:** the node's fault line wraps to a second line instead of cutting (for example "Fault · condenser fouling"), and the effector row wraps (for example "set speed · applied · 360 ms"). The node grows by at most one line.
- **Where:** review and running nodes at 50 % and above.
- **Acceptance:** w-33 at 00:15:00, 50 %: ch-u2 shows "condenser fouling" and pmp-u3 shows "set speed · applied · 360 ms" with no "…" (DOM text check of the node).
- **Impact:** closes F-3. Expected +0.5 on live.

### A-5 [plant] Make frame title headers obstacles, count labels included
- **Change:** the router's title obstacle is the full header row (name, count label and 4 px padding). A wire that would cross the header goes round through the area's top margin, at least 24 px above the frame, as round-1 A-4 already specifies for gutters.
- **Where:** every area header in plant view, editor and review.
- **Acceptance:** at the opening zoom of each of the 11 plants read, zero wire samples lie inside any header rectangle (name plus count label). DOM check of text rectangles.
- **Impact:** closes F-5 (seven plants) and F-11. Expected +1.5 on plant.

### A-6 [plant] Keep wire ends apart on shared input rows
- **Change:** as A-2, in the plant view and editor.
- **Where:** plant view and editor.
- **Acceptance:** at the opening zoom of hunt, compressed-air, power distribution, boiler, food cold chain and packaging, every pair of ends not sharing a source port is at least 6 px apart.
- **Impact:** closes F-2 on the plant side. Expected +1 on plant.

### A-7 [plant] Route the office tower wires in their own gutters, not round the perimeter
- **Change:** the router refuses a route that leaves the rectangle spanned by its source and target nodes, inflated by one gutter (48 canvas units). A wire that cannot be routed inside that rectangle takes the nearest free lane inside its area margin.
- **Where:** plant view and editor (office tower is the reference plant).
- **Acceptance:** office tower at 56 %: for each wire, every path point lies inside its source-target rectangle inflated by 48 canvas units (DOM check); the dashed SC-1 wire no longer reaches the bottom edge of the plant.
- **Impact:** closes F-6. Expected +1 on plant.

### A-8 [plant] Outline: one line per empty area, no wrapping
- **Change:** each outline row is a grid of name, count and status. The name is nowrap with an ellipsis and the full name in a title attribute. An empty area reads "0 · empty" on one line, at most 28 px high. Site stays first because it is the root; the round-1 A-7 acceptance text should say so.
- **Where:** editor outline rail.
- **Acceptance:** hunt and compressed-air editors: each empty row is at most 28 px high and its text sits on one line (DOM check of each row's height and text lines).
- **Impact:** closes F-7. Expected +1 on plant.

### A-9 [plant] Clear the hunt "Production hall A" header at every tier
- **Change:** the site-calendar route runs in the gutter between Site and Production hall A and not through the header (with A-5 applied it is kept clear of the count label too). At 34 % to 63 % a name that wraps keeps its count on the line below, and "Production hall B" keeps at least 6 px from its "1 node" label.
- **Where:** hunt plant at 30 % to 100 %.
- **Acceptance:** DOM check at 34 %, 48 %, 63 %, 72 % and 100 %: distance from the site-calendar wire to the name and to the count is at least 8 px; name and count rectangles do not overlap; "Production hall B" and its count are at least 6 px apart.
- **Impact:** closes F-8. Expected +0.5 on plant.

### A-10 [plant] Overview tier: break names only at spaces or hyphens, and keep wires off the kind text
- **Change:** below 40 % a name breaks only at a space or a hyphen. The router's obstacle for each node adds its kind line and name text (not only the card) at this tier.
- **Where:** plant view and editor below 40 %.
- **Acceptance:** at 34 %: no name is split inside a word (DOM text check); zero wire samples inside any leaf text rectangle of a node.
- **Impact:** closes F-9. Expected +0.5 on plant.

### A-11 [plant] Make wires legible in dark
- **Change:** raise the dark-theme colour of grey and olive wires so each reaches at least 3:1 against the canvas, as G-9 requires for state colours.
- **Where:** review and plant views, dark theme.
- **Acceptance:** in dark, the contrast of each wire colour against the canvas is at least 3:1 (computed from the stroke and background values).
- **Impact:** closes F-10. Expected +0.5 on plant and live.

### A-12 [plant] Verify the hunt link named in the round-1 A-9 acceptance (test data, not UI)
- **Change:** the hunt domain no longer has the packaging-line-b air out to production-line-a air in link that round-1 A-9 cites. Either the acceptance is rewritten for the link that exists (receiver-header-1 pressure out to production-line-a air in, which now enters the left edge at air in), or the fixture is restored.
- **Where:** test data (the hunt-special-names domain), not the canvas.
- **Acceptance:** the acceptance names a link present in the domain, and the check passes on it.
- **Impact:** no points; it makes the round-1 A-9 acceptance checkable.

## Projected score after the actions

- (a) live 6 now. A-1 to A-4 close F-1 to F-4 (+1.5, +1, +1, +0.5) to about 9.
- (b) plant 5 now. A-5 to A-11 close F-5 to F-10 (+1.5, +1, +1, +1, +0.5, +0.5, +0.5) to about 9 on the same 11 plants.
- The dimension is the lower of the two, so 9 needs every A-1 to A-11 item. The projection is an estimate; it is not measured.

## Not verified this round (for the next round)

- Widths 1280, 1440 x 614 and 1920 (no resize, per the brief).
- Dark theme on the plants other than the w-13 review.
- Live effector tick on the progress bar (no live run with an effector call was started).
- The exact 00:19:00 boundary of the 60 s effector window (I sampled 00:19:08, where the mark was gone).

## Summary for the coordinator

- Score: 5/10 (live 6, plant 5; lower of the two). Previous 4, delta +1.
- Round-1 actions: A-1 (effector mark) pass; A-2 (one finish line) pass; A-3 (pin rows, 6 px) fail (3 px at 50 %); A-4 (routing under bodies) pass; A-5 (progress and speed) pass (effector tick not sampled); A-6 (fit and zoom anchor) pass; A-7 (outline order) fail (empty rows wrap); A-8 (campus opens at 40 %, kinds in full) pass; A-9 (plant wires, same-column links) pass on measured parts; A-10 (title clearance) fail below 64 %.
- Findings: 11 total (High 1: F-1; Medium 7: F-2 to F-8; Low 3: F-9 to F-11).
- Actions: 12 total (live A-1 to A-4; plant A-5 to A-12, of which A-12 is test data).
- File: /Users/ghassan/my-projects/streams-simulator/docs/simulation-canvas/director-score-20261015/graph-and-review-round-2.md
