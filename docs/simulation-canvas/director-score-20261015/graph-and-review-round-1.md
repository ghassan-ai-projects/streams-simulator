# Director score, dimension: Information design of the plant graph and the review, round 1

- Director: UI/UX director (brief `brief-director-dim.md`), Chrome at http://127.0.0.1:8766.
- Previous score: 3 (round 10, `ui-polish-sim-authoring-20261015/evaluation-round-10.md`); round 8 findings in `evaluation-round-8.md`.
- Bar: `docs/simulation-canvas/UI-EXPECTATIONS.md` (P-2, P-4, P-14, E-14, V-3, V-4, V-14, G-10, G-4, R-4, R-5, R-9, VIS-4).
- Wire routing is in scope this round (owner reversed the exclusion).
- Coordinator clarification (owner, mid-round): (a) the live run display and (b) the static plant graph are scored separately; the dimension score is the LOWER of the two. Actions are ordered live first, each tagged [live] or [plant].

## Coverage and method

- Own tab 770933672, reloaded from the app. Viewport proof (`javascript_tool`): innerWidth 1474, innerHeight 874, devicePixelRatio 2, scrollWidth 1474 (no horizontal scroll). Not covered: 1280, 1920 and the 1440x614 window of round 10.
- Plants read (plant view, light unless noted): Food cold chain (64 %), Multi-hall campus (29 %), Office tower HVAC (47 %), Data centre cooling (63 %), Hunt & "Ålborg" pumps (dark, 64 %, and its editor, and 76 % after two zoom steps).
- Runs read: w-33 (Multi-hall campus, Evaluate, light, 50 %; Play, seek by mark), w-13 (Water pumping station, Evaluate, dark, 50 %; Play and seek by marks, effector state), w-36 (Food cold chain, Explore, created in this round, light then dark).
- Not read this round: Compressed-air works, Water pumping plant view, Packaging factory, Boiler steam loop, Greenhouse ventilation, Power distribution standby, and the other runs. Round-10 items for those plants are not re-verified here except where the food chain check below covers the same fix.
- Running state: sampled once, about 1 s after Start. The run finished within about 2 s, so progress and speed were not observed past the first frame.
- Side effects, disclosed: (1) one `navigate` call went to tab 770933647 (the tab the brief forbids) before my own tab existed; it only loaded the app's home page and no action was taken on it. (2) I created scratch run `w-36` (Food cold chain, 30 min, Explore, no faults) by Run plant and Start the run, to observe the running state. It was not deleted; deleting needs your say. (3) The hunt plant editor was opened and left with "No changes"; nothing saved. (4) The theme was switched through Light and Dark and restored to System. (5) My tab 770933672 is closed.

## Score

**Overall for this dimension: 4/10** (lower of (a) 5 and (b) 4). Previous score 3; delta +1.

- (a) Live run display (playback graph with readouts, scrubber, fault and effector marks, running state, update over time): **5/10**. Readouts update during playback, the scrubber marks and fault chips read well, and the kind line no longer overprints readout labels (ISS-10D-03 holds). Against it: wires draw over readout text in w-13 at 50 % (VFM-1 "1199 rpm", VFM-3 "motor current 56.7 A"); effector calls leave no mark on their node; the running state cannot be watched because the run ends in about 2 s.
- (b) Static plant graph (plant view and editor): **4/10**. The kind line now has its own row (food cold chain, 64 %: "pump" row under "pump-cond", no overprint), and wire ends land on pin rows. Against it: wires run under node bodies on four plants, a dashed wire loops around the left edge of production-line-a and pl-h1, arrowheads land on the far side from their pin, the campus opens at 29 % with kind lines cut, zoom is not anchored (the Site frame is clipped at 76 %), and the editor outline still leads with four empty areas.
- Evidence (screen, plant, theme, quoted label, value):
  1. w-13, 50 %, dark, 00:05:00: "CP-1 · Fault" chip with "impeller wear" in red; playback at 60x; the node is marked by shape and colour (V-4 holds).
  2. w-33, 50 %, light, 00:11:24 after the 00:10 mark: rcv-u1 "header pressure 737 kPa" against 715 kPa three seconds earlier (readouts update).
  3. w-13, dark: the teal wire from CP-3 crosses the value text of VFM-1 "1199 rpm" and VFM-3 "motor current 56.7 A".
  4. Hunt plant, 64 %, dark: the packaging-line-b "air out" to production-line-a link ends on the right edge of production-line-a, while its "air in" pin is on the left edge.
  5. Campus, 29 %, light: "cooling to…", "packaging…", "production…", "feeder met…" cut; cal-site shows no kind line.

## What holds

- Kind line has its own row under the id on the plant nodes (food cold chain at 64 %, hunt plant at 64 %); the overprint of round 10 (ISS-10D-01) is not reproduced on these plants.
- In the review, the kind line is separate from the readout labels (w-33 "chiller" over "chw supply temp" with a gap; w-13 "pump" over "flow" clear). ISS-10D-03 holds.
- Scrubber: fault marks are red circles, effector calls blue squares, chips centred under their marks, a legend ("Fault starts", "Effector call") in the toolbar, end labels 00:00:00 and 00:30:00. The 00:10 mark seek selects the node and opens Inspect with "OBSERVED · ch-u2 at 00:11:24" (V-3 holds).
- Fault on the node: a "Fault" chip with a wrench mark, a red dashed border, and the cause in red ("condenser fouling", "impeller wear", "strainer blockage"). Marked by shape and colour, not colour alone.
- Readouts: label above value, units spaced ("7.00 °C", "1199 rpm"), a true minus sign ("−0.50 m³/h"), thousands separators; the same node size across the review.
- Review header line: "19 089 observations · opened at 00:01:00, once every channel had reported" explains where playback starts; "Verified: the files match and the replay reproduces." is one line.
- Running state chrome: top-bar pill "Running Food cold chain · 00:00:00 of 00:30:00", the panel text "Stop now ends the run at the next step and keeps what ran as a shorter run you can review", and a single "Run w-36 of Food cold chain finished · Review" toast.
- Plant opening views are fitted with margin on the food chain (64 %), data centre (63 %) and office tower (47 %).
- The review's Navigator shows the viewport frame.

## Findings

### F-1 Wires draw over readout text in the live review
- Severity: High
- Where: Run review w-13 (Water pumping station, Evaluate), 50 %, dark, playhead 00:05:00 (zoom of the Pump house frame). The teal wire from CP-3 crosses VFM-1 "speed 1199 rpm" and VFM-3 "motor current 56.7 A".
- Observed: a wire is drawn over the value and unit glyphs of a readout node. During playback these values are what the user reads.
- Expected: readouts legible at every tier and during playback (V-4, V-14); no wire over node text (E-14 "never passes behind a node body" applies to the same routing, reversed here as wires on top).

### F-2 Wires pass under node bodies in the live review
- Severity: Medium
- Where: Run w-36 (Food cold chain, Explore), 50 %, dark: the orange wire from pump-cond's outlet runs under the receiver body (x about 630 to 838 in the zoom) and emerges at its lower right; the pink wire from chiller-1 runs under the cooling-tower body. Run w-33 (Multi-hall campus, Evaluate), 50 %, light: the dashed site-calendar wire runs under ch-u2, pmp-u2, pmp-u3 and ch-u1; the teal wire from pmp-u1 runs under pmp-u3. Run w-13: the teal wire from CP-3 runs under VFM-1's body, the orange wire from FM-1 under PCV-1.
- Observed: a wire disappears behind one node and reappears on another, so the eye cannot trace it; in the campus several wires share one gap.
- Expected: wires never pass behind a node body (E-14); wires are clearly coloured by port family (P-4).

### F-3 Two wire ends meet on the same point in the review
- Severity: Medium
- Where: Run w-36, 50 %, dark, chiller-1 (left edge, about y 450 in the zoom): the teal wire from pump-chw and the orange wire from drive-chw end at the same point.
- Observed: the review draws only the observed readout rows of a node, so the input ports that are not shown have no row; the wires land on the body's vertical centre and merge. The user cannot tell which input each wire feeds.
- Expected: each wire ends on its own port row (E-14, V-4).

### F-4 Effector calls leave no mark on the node they act on
- Severity: Medium
- Where: Run w-13, 50 %, dark, playhead 00:18:00 (blue square on the scrubber, "00:18:00 CP-3"). CP-3 is selected by the mark; the node shows a blue selection ring, "flow 166 m³/h", "power 16.5 kW", and nothing about the call.
- Observed: the effector "stop, applied · 371 ms" is visible only in the report's EFFECTOR CALLS table and as the timeline square. At the node that the call changed, the user sees no effector state.
- Expected: marks on the timeline and the node (V-3, V-4), with the mark shape and colour distinct from faults (VIS-4).

### F-5 The running state cannot be followed; progress and speed are not shown
- Severity: Medium
- Where: Run panel after Start on Food cold chain (light), "Creating the world..." with a thin progress bar and "00:00:00 of 00:30:00 simulated"; the top-bar pill "Running Food cold chain · 00:00:00 of 00:30:00"; the toolbar "Running · 00:00:00". About 1 s later: "Run w-36 finished."
- Observed: the run is at 00:00:00 in the one sample taken, and then finished. No speed (simulated seconds per second), no tick for faults, and no running indicator on the nodes. A user who starts a 30 min run sees a finished run within 2 s.
- Expected: progress with speed beside the bar, and ticks for injected faults and effector calls (R-4, R-9). Speed cannot be shown on a run that finishes in 2 s; see A-5 for the ceiling.

### F-6 A finished run states its name twice
- Severity: Low
- Where: Run panel after Start on Food cold chain, light: the top toast "Run w-36 of Food cold chain finished · Review" and the panel "Run w-36 finished." with "Open the review". Round 10 ISS-10D-12 still holds.
- Observed: the same status appears twice in two wordings.
- Expected: one status line that names the run once (R-5).

### F-7 Plant wires pass under node bodies and arrowheads sit on the far side of their pin
- Severity: Medium
- Where: Hunt & "Ålborg" pumps, 64 %, dark, plant view, the link from packaging-line-b "air out" to production-line-a "air in": the arrowhead sits on production-line-a's right edge at the "air out" row, while "air in" is on its left edge, so the wire passes under the body. Food cold chain, 64 %, light: the brown wires from compressor "power out" run under tank-chw; the wires into meter-main "loads in" run under gen-1; two brown wires into ahu-b run under its body.
- Observed: the arrowhead reads as feeding the node's output row; the wire is hidden for its length inside the node.
- Expected: each wire ends on its own pin row, enters from the side that faces its source and never passes behind a node body (E-14, P-4).

### F-8 Same-column wires loop round the node edge
- Severity: Medium
- Where: Hunt plant, 64 %, dark: the dashed site-calendar wire loops round the left edge of production-line-a and ends at its kind row, not at a pin. Multi-hall campus, 29 %, light: the dashed wire to pl-h1 runs under cv-h1 and pk-h1 and hooks round the left edge of pl-h1.
- Observed: the wire hooks back at the node's left edge; the end is not on a pin row.
- Expected: no loops between same-column nodes; each wire ends on its pin (E-14, P-14).

### F-9 A dashed wire touches the "Production hall A" title
- Severity: Medium
- Where: Hunt plant, 64 %, dark: the dashed wire from site-calendar reaches the frame title at the last glyph "A" of "Production hall A" (zoom of x 255-655, y 420-690).
- Observed: the wire comes to the title's end and runs on the title's edge. Round 8 ISS-8D-02 and round 10 ISS-10D-06 were about this kind of crossing; at 64 % it still touches the title.
- Expected: area titles are never crossed by wires (P-14).

### F-10 The campus opens at 29 % and cuts its kind lines
- Severity: Medium
- Where: Multi-hall campus opening view, 29 %, light, 1474x874: "cooling to…" (ct-u1), "packaging…" (pk-h1 to pk-h3), "production…" (pl-h1 to pl-h3), "feeder met…" (mtr-h1 to mtr-h3), "air compre…" (comp-u1). cal-site shows the id only. At this tier the wires end at node edges with no pins drawn.
- Observed: the largest plant opens with kind lines cut and node text about 7 to 8 px; the plant fills the height and leaves about 180 px empty on each side. This is round 10 ISS-10D-04, reproduced on this window.
- Expected: at the opening zoom every node shows its id and its full kind in words (P-2, G-12, VIS-6).

### F-11 Zooming in is not anchored, and the plant is cut by the left rail
- Severity: Medium
- Where: Hunt plant, dark: from the opening view (64 %), two presses of "+" (to 76 %) move the plant so the Site frame and the "Production hall A" title are cut at the canvas's left edge at x 246, under the outline rail. No pan cue.
- Observed: the zoom centres on the canvas, not on the content, so the top-left frames leave the view.
- Expected: the plant stays fitted and in view when zoomed; no clipping at the rail (G-10, P-3).

### F-12 The editor outline lists four empty areas before the populated group
- Severity: Medium
- Where: Editor of Hunt & "Ålborg" pumps, outline rail, dark: PRODUCTION HALL 0, COMPRESSOR ROOM 0, CHILLER PLANT 0 and SUBSTATION 0 come first, each with "No equipment yet" on its own line (about 240 px together); COMPRESSOR HOUSE 5 starts below them. Same pattern as round 10 ISS-10D-05 on compressed-air works.
- Observed: the populated group sits below four empty groups; the plant view lists the same areas as one line each on the graph's frames.
- Expected: populated areas first; an empty area on one line with its heading (G-4, ISS-6D-37).

## Actions to reach 9

Live run display first, then the static plant graph. The wire router is one presentation function in the canvas app layer; the review and the plant view both call it (principle: one use case, one implementation). The canvas computes no binary-owned fact here: routing and marks are presentation over data the run already carries.

### A-1 [live] Mark each effector call on the node it acts on
- **Change:** at the playhead time of an effector call, and for 60 s of simulated time after it, the target node shows a blue square badge (the same mark as the scrubber) beside its id and a row "stop · applied 371 ms" (outcome text from the run's EFFECTOR CALLS table) under the header. Inspect lists the call with its outcome.
- **Where:** review graph, node header, right of the id and beside the fault chip.
- **Acceptance:** in w-13 seek to 00:18:00: CP-3 shows the blue square and "stop · applied"; at 00:19:00 it still shows; at 00:00:30 it does not. Checked by screenshot zoom in dark and light.
- **Impact:** closes F-4; +1 on (a).
- **Status:** done, 55996f0. Measured in Chrome (w-13): at 00:18:00 and 00:19:00 CP-3 carries the blue square and "stop · applied · 371 ms"; at 00:19:10 and 00:00:30 it does not.

### A-2 [live] Make one finished status line
- **Change:** the toast reads "Run w-36 finished · Open review"; the panel body shows the plan summary only, and the "Run w-36 finished." line is removed.
- **Where:** run panel and top toast after Start.
- **Acceptance:** after Start on any plant the word "finished" appears once on screen (count with get_page_text).
- **Impact:** closes F-6; +0.5 on (a).
- **Status:** done, 16f42e4. The e2e launch spec counts "finished" once on the page after a run; the toast reads "Run w-N finished · Open review".

### A-3 [live] Give every port a pin row in the review
- **Change:** a review node lists every port of its kind as a row. A row with no observed value shows the port name in grey and "not observed" in place of the value. Each wire lands on the centre of its port's row. Two wire ends on one row are spaced 6 px apart along it.
- **Where:** review graph nodes at every zoom tier.
- **Acceptance:** in w-36 chiller-1 the teal (load in) and orange (flow in) wires end on different rows, with at least 6 px between endpoints (DOM check on the wire path end coordinates). No two wire ends within 6 px anywhere in w-33 and w-36 at 50 %.
- **Impact:** closes F-3; +1 on (a).
- **Status:** done, 0f60815 (rows) and 0646771 (6 px spacing in the router). Measured in Chrome (w-36, w-33, 50 %): chiller-1 wires end on different rows; no two wire ends within 6 px. At the compact tier an unobserved row shows "–" (full text "not observed" for screen readers and the tooltip), because the words do not fit beside the label at that size.

### A-4 [live] Route review wires round node rectangles, beneath all node text
- **Change:** wires move to a layer under every node body; node bodies are opaque. The router is orthogonal: a wire leaves its source port horizontally for 16 px, runs vertically in a gutter between columns (at least 24 px wide, lanes 6 px apart, lane order follows the vertical order of the two ends to avoid crossings), and enters the target horizontally for 16 px. Obstacles: every node rectangle inflated by 8 px, every area title box inflated by 4 px. When a gutter has no free lane, the wire takes the area's top margin (at least 24 px above the frame). Curves are dropped.
- **Where:** review graph (all tiers) and, through the same router, the plant view and editor (A-9).
- **Acceptance:** in w-13 at 50 %, dark, playhead 00:05:00: no wire pixel over the text of VFM-1 or VFM-3. In w-33 and w-36 at 50 %: zero wire segments under a node rectangle (DOM check: sample each segment against node rectangles inflated by 4 px). Wire crossings in w-33 drop from the current count to at most 2 per lane. Checked by zoom screenshot in both themes.
- **Impact:** closes F-1 (High) and F-2; +2.5 on (a). This is the largest single gain on the live surface.
- **Status:** done, 0646771. Measured in Chrome (w-13, w-33, w-36 at 50 %): 0 wire samples within 4 px of any node rectangle; 12 crossings among 19 wires in w-33. e2e asserts both at every zoom tier of the factory plant and a review.

### A-5 [live] Show run progress and speed while a run is live; ask the binary for pacing if the run is too short to watch
- **Change:** the progress panel shows "Running · 00:12:00 of 00:30:00 simulated · 60 s per s" (speed from the last 5 s of wall-clock time), a 6 px bar, and a tick on the bar for each fault injected and each effector call as it is reached. The top-bar pill shows the same progress.
- **Where:** run panel progress block and the top-bar pill.
- **Acceptance:** start a 30 min run; a sample at 1 s and at 3 s shows the simulated time increasing and a speed value; the ticks appear at their minutes. If the run finishes within 2 s, the acceptance cannot be met in the UI.
- **Impact:** closes F-5; +1 on (a). **Ceiling:** a 30 min Explore run completes in about 2 s in this lab, so no speed or progress is visible. The fix needs a pacing option in the binary (a run speed setting, for example 60 s per s) or a longer run for review; this is a binary change first (the canvas computes no run timing).
- **Status:** done, e3e9874, without a binary change: the canvas paces the run by pausing between steps ("Run speed": Watch about 30 s, default, or As fast as possible). Measured in Chrome on a 30 min Food cold chain run: samples 1 s apart read 00:04:00, 00:05:00, 00:06:00, 00:07:00 at 79, 75, 60, 60 s per s; bar 6 px; Stop now wakes a waiting run.

### A-6 [plant] Keep the plant fitted and in view when zooming
- **Change:** a zoom step keeps the point under the centre of the visible canvas fixed, then clamps the content so that the outer frame titles (margin 16 px) stay inside the canvas. Opening zoom is the fit.
- **Where:** plant view and editor, zoom controls and keyboard zoom.
- **Acceptance:** from the opening view of the hunt plant, press "+" twice: the "Site" and "Production hall A" titles remain fully visible at 76 %; the "Fit" button returns the opening view exactly.
- **Impact:** closes F-11; +0.5 on (b).
- **Status:** done, 046b353. Hunt plant: two presses of + keep "Site" and "Production hall A" at left 16 px; Fit returns the opening view exactly.

### A-7 [plant] Order the editor outline by equipment, populated first, empty areas on one row
- **Change:** areas are ordered by equipment count (populated first, plant order within each group). An empty area renders as one row: the heading, the count "0" and "No equipment yet" on the same row, at most 28 px high.
- **Where:** editor outline rail.
- **Acceptance:** on the compressed-air and hunt editors the first group is COMPRESSOR HOUSE 5 and each empty area takes one row of at most 28 px.
- **Impact:** closes F-12; +0.5 on (b).
- **Status:** done, b88c08f. Hunt editor: Site 2, Compressor house 5, Production hall A 1, Production hall B 1, then four empty rows of 28 px. Site is the root that holds the others, so it stays first.

### A-8 [plant] Open the campus with every kind in full and wrap kinds at the overview tier
- **Change:** the opening zoom is the larger of the fit and 40 %; the view is centred on the plant's top-left frame, and the canvas scrolls. Below 40 % each kind wraps to two lines, as the ids do; the kind text is at least 9 px at every tier.
- **Where:** plant view opening zoom; overview tier of all plants (campus, power distribution standby at 33 %).
- **Acceptance:** the campus opens with every kind in full (no "…") in a 1474x874 window and a 1440x614 window; the power distribution standby opens with "production line" in full.
- **Impact:** closes F-10; +1.5 on (b).
- **Status:** done, 98368b3. Campus opens at 40 % with every rendered kind uncut at 10.5 px; Power distribution standby shows "production line" in full. The 1440x614 window was not tried (windows are not resized in this loop).

### A-9 [plant] Route plant wires with the same router and put each input on the side that faces its source
- **Change:** A-4's router applies to the plant view and the editor. Each input pin sits on the side of its node that faces its source (left when the source is left, right when the source is right). A same-column link leaves the source's bottom (or top) edge and enters the target's top (or bottom) edge through a 16 px gutter. A wire never enters a node on the far side. Output ports are on the right, inputs on the left, as now.
- **Where:** plant view and editor graph.
- **Acceptance:** hunt plant at 64 %: the packaging-line-b to production-line-a wire ends on production-line-a's left edge at "air in" and passes no node body; food cold chain at 64 %: zero wire segments under tank-chw, gen-1 or ahu-b; the campus shows no same-column loop (count of wires whose path returns to the column of their source = 0).
- **Impact:** closes F-7 and F-8; +2 on (b).
- **Status:** done, c28cb4d (router in 0646771). Inputs stay on the left and every wire enters the left edge at its pin, going round through the gutters when the source lies to the right; a same-column link joins bottom to top. Food cold chain at 72 %: 0 wires within 4 px of any node.

### A-10 [plant] Keep area titles clear of wires
- **Change:** the title box of each area (text plus 4 px padding) is an obstacle in A-4's router; a wire that would cross a title goes round it through the area's top margin.
- **Where:** every area frame in the plant view and the review.
- **Acceptance:** hunt plant at 64 %: the dashed site-calendar wire is at least 8 px from the "Production hall A" text at 30 %, 64 % and 100 % (measured by DOM rect distance).
- **Impact:** closes F-9; +0.5 on (b).
- **Status:** partly met, c28cb4d. The site-calendar wire is 11.3 px from the "Production hall A" text at 72 % and 26 px at 125 %. At 29 % it is not met: the title wraps to two lines (117 units) over an 84-unit frame gap and overlaps the site-calendar node, so no route keeps 8 px; the layout's 48-unit gap between areas would have to grow with the title.

## Projected score after the actions

- (a) live: 5 now; A-1 to A-5 close F-1 to F-5 and F-6 (+1, +0.5, +1, +2.5, +1, to about 9 when all are in). A-5 has a binary ceiling for the speed value.
- (b) static: 4 now; A-6 to A-10 close F-7 to F-12 (+0.5, +0.5, +1.5, +2, +0.5, to about 8.5 to 9 on the same ten plants).
- Dimension: the lower of the two, so 9 needs both groups of actions in full, and a re-check of all ten plants on 1280, 1440 and 1920 widths.

## Summary for the coordinator

- Score: 4/10 (live 5, static 4; lower of the two).
- Findings: 12 total (High 1: F-1; Medium 10: F-2 to F-5 and F-7 to F-12; Low 1: F-6).
- Actions: 10 total (live A-1 to A-5, then static A-6 to A-10).
- File: /Users/ghassan/my-projects/streams-simulator/docs/simulation-canvas/director-score-20261015/graph-and-review-round-1.md
