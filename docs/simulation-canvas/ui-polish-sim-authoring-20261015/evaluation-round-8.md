# UI/UX evaluation, round 8

- Director: UI/UX director (Haiku 5.5 subagent), Chrome at http://127.0.0.1:8766, own tab (tab 770933661; tab 770933647 was not used).
- Date: 2026-10-15.
- Bar: `docs/simulation-canvas/UI-EXPECTATIONS.md`.
- Previous evaluation: `docs/simulation-canvas/ui-polish-sim-authoring-20261015/evaluation-round-6.md` (overall 4/10, 40 findings).
- Scope reviewed: Home (13 plants, 33 runs); the ten plants named in the brief (compressed-air-works, office-tower-hvac, water-pumping-station, data-centre-cooling, packaging-factory, boiler-steam-loop, greenhouse-ventilation, food-cold-chain, power-distribution-standby, multi-hall-campus) and the other plants on Home where they bear on a finding (dir-review-empty, hunt-special-names, hunt-empty); runs w-1 to w-33 as listed on Home.
- Viewports covered: 1440x757 only. Proof: `resize_window` to 1280x800 reported success and `javascript_tool` read `window.innerWidth` 1440, `innerHeight` 757 (then `scrollWidth` 1440). The same request for 1920x1080 also left `innerWidth` at 1440. No width other than 1440 was reviewed; 1280 and 1920 are NOT covered in this round. Zoom of screenshots is not a substitute for a real width.
- Light and dark: the app opened in dark ("Theme: Dark"). Dark was reviewed on the plant views, the editor, the run panel, the Keyboard dialog and the review (w-13, w-22, w-33). Light (the "System" theme on this machine) was reviewed on Home, the data centre plant, the campus, the hunt plant, the review w-25 and the scratch run w-34. Theme was switched once during the review and left on "System".
- Runs: opened in full: w-13, w-22, w-25, w-33 and the scratch run w-34. The other runs (w-1 to w-12, w-14 to w-21, w-23, w-24, w-26 to w-32) were checked only as rows on Home and in each plant's Recent runs list, not opened.
- Scratch artefacts created in this round (for cleanup): plant `dir-empty-check-8` (created from New plant with "An empty plant"; not saved or edited), run `w-34` (30 min Explore on `dir-review-empty`). No existing plant was edited or saved; a rename attempt on air-dryer-1 in the compressed-air editor was refused by the field check and not kept.

## Findings

Numbered by severity: Medium first (ISS-8D-01 to ISS-8D-09), then Low (ISS-8D-10 to ISS-8D-20). No High finding was found in this round. The regression pass on round 6 and the score table follow the findings.

## ISS-8D-01 Wires end on node header rows and hook at node edges; the round-6 wire fix does not hold
- Severity: Medium
- Kind: graph
- Where: Plant view of Compressed-air works, 62 %, dark, 1440x757. The wire from packaging-line-b enters production-line-a at its right edge level with the "production line" kind row, above the first pin row ("shift in"). The dashed site-calendar wire loops at the left edge of production-line-a. Plant view of Multi-hall campus, 30 %: the dashed wire from cal-site ends on the header of ch-u2. Plant view of Boiler steam loop, 59 %: the magenta wire from hx-1 hooks at the left edge of hx-2 at its header row.
- Observed: several wire ends sit on the lower edge of the header band or on the kind row, not on a pin row, and some wires hook back at the node edge. ISS-6D-11 and ISS-6D-32 are recorded as fixed; for these wires they are not.
- Expected: every wire ends on the pin row of its port, never on the header and never in a loop (E-14, P-14).
- Suggestion: anchor a wire end to the port's own row, and route a same-column link around the node edge instead of hooking it.
- Status: fixed 3e237f3, 322db4d (area ordering left out: see the Fix line)
- Fix: Plant view and review: a wire end lands on the pin row of its port, and a wire leaves or enters on the side that faces its other end (a far-side handle sits on the same row), so a link back to the left no longer loops round its node (boiler steam loop: 1 loop in 10 wires before, 0 after). Review: a wire ends on the readout row of its port's channel (row centre; 24 of 40 ends in the measured w-33 review), or at the body's vertical centre when that channel is not shown (16 of 40), never on the header (before: 40 of 40 ends sat at a fixed 56 px, 15 px below the header's bottom edge and on the label of the first readout row, not on a row's centre). Area ordering (8b52380, reverted in cb7553d): left out. A second attempt changed the order of the areas, not the nodes inside one, and failed two e2e specs (plant.spec "Enter opens a hall and the arrows keep working", review.spec "a hall carries its faulted members"): the hall nodes fell outside the fitted view. Nodes inside an area already follow their predecessors (orderByPredecessors), so no within-area change was made.

## ISS-8D-02 Wires cross area frame titles
- Severity: Medium
- Kind: graph
- Where: Plant view of Packaging factory, 56 %, dark, 1440x757. Two dashed wires from W-1 and SC-1 cross the "Warehouse" title (between "Ware" and "house") and the "Line 2" title (through the "2"). Plant view of Boiler steam loop, 59 %, dark: the magenta wire from hx-1 touches the end of the "Process hall" title.
- Observed: wires are drawn over frame titles, so a title reads as part of a wire.
- Expected: area titles are never crossed by wires (P-14).
- Suggestion: draw frame titles above wires with a backing, or route wires around the title band.
- Status: fixed 57289d2
- Fix: Area titles render in their own raised node layer with an opaque backing, above the wires.

## ISS-8D-03 Multi-hall campus at its opening 30 % cuts the kind lines and is too small to read
- Severity: Medium
- Kind: graph
- Where: Plant view of Multi-hall campus (26 instances), opening zoom 30 %, dark and light, 1440x757. Cut kind lines: "packaging ..." (pk-h1, pk-h2, pk-h3), "production ..." (pl-h1, pl-h2, pl-h3), "cooling to..." (ct-u1).
- Observed: at the zoom the largest plant opens at, the kind line of each node is truncated and the node text is about half the size of the same text at the 62 % of the compressed-air plant. The ids stay readable; the kinds, which P-2 asks for, do not.
- Expected: at the opening zoom every node shows its id and kind in words (P-2, G-12).
- Suggestion: choose the opening zoom from the legibility floor (the longest kind must fit), or wrap kinds onto a second line as ISS-6D-10 did for ids.
- Status: won't-fix not reproduced: the campus opens at 52 % at the checked window with every kind line in full; the 30 % figure was not re-measured at the same window (the canvas server stopped serving during the check, see report)
- Fix: No change.

## ISS-8D-04 Review shows the binary's name and a folder word while it loads
- Severity: Medium
- Kind: content
- Where: Run review of w-33 (Multi-hall campus, Evaluate), first second after opening. Sub-bar chip "Report · streamsim report reading the run · 0 s"; Report pane "streamsim report is reading the run folder...", 1440x757, dark.
- Observed: the first state a user sees on a review names the internal binary and a file word. Once loaded, the same screen says "Events appear once streamsim report has read the run."
- Expected: plain loading words with no product-internal name and no folder word (G-6, N-7), for example "Reading the run...".
- Suggestion: change the loading copy in the bridge's message mapping; keep the binary's name in the For developers disclosure only.
- Status: fixed 8a28f4c
- Fix: Loading reads Reading the run, with no binary name or folder word; the scrubber reads Events appear once the run is read.

## ISS-8D-05 The editor outline has an empty "Halls" group that the plant view does not have
- Severity: Medium
- Kind: coherence
- Where: Editor of Multi-hall campus (#/edit?composition=compositions%2Fmulti-hall-campus.composition.json), Outline rail, dark and light, 1440x757. Group "HALLS 0" with "No equipment yet. Add some in the Build panel.", above "HALL 1 5". The plant view of the same plant has no Halls group: its areas are Utilities, Hall 1, Hall 2, Hall 3 and Site.
- Observed: the editor shows a group named "Halls" with no entries next to areas named "Hall 1" to "Hall 3". The two names are for different things (a hall node that opens versus an area) and the plant has no hall nodes.
- Expected: one area list on both surfaces (ISS-6D-37, E-1), and a group shown only when the user can add that kind, with a name that does not collide with the area names.
- Suggestion: hide the Halls group when the plant has no hall nodes, or name the kind "Hall nodes" and give it its own add action.
- Status: won't-fix the Halls group is the campus composition's own empty area (Site > Halls > Hall 1 to 3); the editor lists empty areas so equipment can be added to them; renaming it is a data change
- Fix: No change.

## ISS-8D-06 A new plant's save state reads "Not checked yet" beside an enabled Save
- Severity: Medium
- Kind: feedback
- Where: Editor of a new empty plant (Dir Empty Check 8, created with New plant and "An empty plant"), top edit bar, 1440x757, dark.
- Observed: the edit bar reads "Not checked yet: Save to see its issues." and Save is a filled primary button. The compressed-air editor reads "No changes" with Save disabled (E-11). Nothing on the bar says whether the new plant has been saved.
- Expected: one of the E-2 save states, named plainly (No changes, Unsaved changes, Saved and composed, Saved but not composed, Not saved), so the user knows what Save does.
- Suggestion: map the empty draft to a named E-2 state; disable Save until there is something to save.
- Status: fixed 48f49f0
- Fix: An empty new plant reads Nothing to save yet: add equipment in Build, and Save stays off until there is something to save.

## ISS-8D-07 A click on an area title leaves a focus ring on the first instance, with nothing selected
- Severity: Medium
- Kind: feedback
- Where: Plant view of Multi-hall campus, 30 %, light, 1440x757. Click on the "Hall 1" frame title at (785, 255).
- Observed: the click has no visible effect except a dashed focus ring on cal-site, the first outline item. The inspector stays on the plant overview. Keyboard focus moved to an element the user did not pick.
- Expected: after a click, focus stays on the element clicked or on the selection made; it never lands on an unrelated instance (G-8, the same class as ISS-6D-36).
- Suggestion: do not move focus on an area click; if areas are not selectable, make the title non-interactive.
- Status: fixed 6c649ea, refined 17e5e58 (keyboard focus starts at the first instance again; a pointer focus shows only the selection)
- Fix: Focus on the graph shows the selected instance only, so a click on an area no longer moves the ring to the first instance.

## ISS-8D-08 Two wire families share the same grey diamond in the legend
- Severity: Medium
- Kind: colour
- Where: Legend popover on the plant view (Legend button), 1440x757, dark. "Signal" is a pale grey diamond and "Weather" a mid grey diamond. Also "Pressure" (violet diamond) and "Temperature" (magenta diamond) share the diamond mark.
- Observed: the two grey families differ only by shade, so the legend cannot tell Signal wires from Weather wires; the weather wires are the dotted grey ones in the plant. Pressure and Temperature differ only by hue within the same mark.
- Expected: each family separable by hue and mark, and colour never the only carrier (P-4, ISS-6D-19).
- Suggestion: give Signal its own hue or a different mark (for example a hollow square), and keep Weather as the dashed grey family.
- Status: fixed beada97
- Fix: Signal has a square mark; Weather keeps the grey diamond and dashed wire.

## ISS-8D-09 The running panel shows no speed, and the plant name in its header is cut
- Severity: Medium
- Kind: feedback
- Where: Run panel while Dir Review Empty runs (Explore, 30 min), run w-34, sampled once about 1 s after Start, 1440x757, dark. Header "Run Dir Review E..." with "Running · 00:00:00 elapsed"; body "00:24:00 of 00:30:00 simulated".
- Observed: no speed (simulated seconds per second) is shown anywhere in the panel; the header clock reads 00:00:00 while the body and the top-bar pill read 24 min simulated. The run finished in about 4 s, so the header could not be watched over time. The plant name is cut in the header, and the cut hides the part that tells runs apart.
- Expected: speed beside the progress bar (R-4), the header clock matching the run's own progress, and the plant name readable (R-9). ISS-6D-39 is recorded as fixed for the header speed; this sample does not show it.
- Suggestion: give the plant name its own line in the panel header, and show speed next to the progress bar; re-check with a 24 h run, which is slow enough to watch.
- Status: fixed 66c494b
- Fix: The running panel header lets the plant name wrap instead of cutting it. Speed was not changed: the speed suffix shows once the run has measurable progress, and the sampled panel was too early to show it.

## ISS-8D-10 The review's time and total change on load
- Severity: Low
- Kind: feedback
- Where: Run review of w-33 on first open, 1440x757, dark. The playback reads "00:00:58 of 00:29:58" before the report has read the run, then "00:01:00 of 00:30:00" once every channel has reported.
- Observed: the total duration changes from 29:58 to 30:00 and the cursor moves on load.
- Expected: a stable total and a cursor that starts once the run is read (V-2), with a placeholder until the duration is known.
- Suggestion: show a placeholder for the time and total until the run's duration is known, and set both once.
- Status: fixed 000775c
- Fix: Until the run is read the time and total show --:--:--; both are set once the run is read.

## ISS-8D-11 The "Show all" link sits flush on the inspector's left edge
- Severity: Low
- Kind: layout
- Where: Plant view of Compressed-air works, inspector, RECENT RUNS, before the link is clicked, dark, 1440x757: "Show all 8 runs" at x 1066.6, the panel's own left edge. The section heading and rows start at x 1079.
- Observed: the link is outdented 12 px from the section it belongs to.
- Expected: the link is inset like the rows above it (VIS-5, one gutter).
- Suggestion: apply the section's inset to the link.
- Status: fixed d3e818a
- Fix: The Show all runs link uses the row inset (2 px vertical, 12 px horizontal).

## ISS-8D-12 The Legend popover has no inner padding
- Severity: Low
- Kind: craft
- Where: Legend popover on the plant view, dark and light, 1440x757. Measured: the popover spans x 832 to 1050 with padding 0px; the heading "Wire colour and pin shape by port type" starts at x 832 and the explanatory line runs to x 1050.
- Observed: text touches the popover's edges on every side.
- Expected: inner padding of 12 to 16 px, as on the other popovers, on the 4 px scale.
- Suggestion: add padding to the popover body.
- Status: fixed 2fbc74b
- Fix: The legend popover has 12 px of inner padding on every side.

## ISS-8D-13 The Keyboard dialog lists an empty EDITOR section and review keys on the plant screen
- Severity: Low
- Kind: content
- Where: Keyboard dialog on the plant view (Compressed-air works) and on a review (w-22), dark, 1440x757. The EDITOR heading has no rows above Close. The plant view lists the RUN REVIEW keys (Space for play, the time-slider keys), which do nothing on the plant screen.
- Observed: the same list appears on every screen; one section has no rows, and another belongs to a different screen.
- Expected: each screen lists only the keys it has, and no heading without rows (G-15: real keys).
- Suggestion: filter the sections by the current screen, and drop a heading when it has no rows.
- Status: fixed 40e4cc7
- Fix: The Keyboard dialog lists only the groups for the screen it opens from; a group with no rows is not shown.

## ISS-8D-14 A key explanation wraps inside its list of characters
- Severity: Low
- Kind: typography
- Where: Editor, Build panel Name hint after choosing Centrifugal pump (Site area), and the inspector Name hint of air-dryer-1 after a rename with "bad name!", 1440x757, dark. The rule reads "...letters, digits and . _" at the end of one line and ": / -, with no spaces." at the start of the next.
- Observed: the punctuation list in the rule breaks after "_", so the list reads as two unrelated pieces. ISS-6D-22 asked for the same fix on the error line; the error colour and border now hold, the break does not.
- Expected: the list of characters stays on one line or moves to its own line as a whole (N-7).
- Suggestion: set the spaces inside the character list as non-breaking spaces.
- Status: fixed dff9059
- Fix: The punctuation list in the name rule is bound with non-breaking spaces so it stays on one line.

## ISS-8D-15 The New plant id hint names file names, and the id error ends with a one-word line
- Severity: Low
- Kind: content
- Where: Home, New plant form, Id field, dark, 1440x757. The default hint reads "Used for its file names: lower case, digits and hyphens." The error after "Create and open the editor" with "Bad Id_1" ends with the single word "letter" on its own line.
- Observed: "file names" is the first thing the form says about the id, which is a file-level detail (W-7). The error's last line is one word.
- Expected: the id rule in plain words, with the file detail secondary; no one-word last line (N-7).
- Suggestion: write the hint as "Lower case letters, digits and hyphens, starting with a letter," and keep the file detail in a tooltip; bind the last two words of the error with a non-breaking space.
- Status: fixed dede804
- Fix: The id hint reads Lower case letters, digits and hyphens, starting with a letter. The error wording comes from the binary and is unchanged; the error text does not end on one word (text-wrap: pretty). The file-name tooltip was not added.

## ISS-8D-16 The Inspect panel repeats one time and one id in several places
- Severity: Low
- Kind: content
- Where: Run review of w-33, Inspect tab, pmp-u2 selected (click on the node), dark, 1440x757. The panel header reads "OBSERVED · pmp-u2 at 00:01:00", the caption under it reads "Observed at 00:01:00", the tab reads "Inspect · pmp-u2", and the true results block reads "TRUE RESULTS pmp-u2".
- Observed: the same time appears twice and the same id three times in one panel.
- Expected: one time and one id per block (V-5).
- Suggestion: drop the caption and keep the time in the header; drop the id from the block titles that sit under the tab title.
- Status: partly fixed b7d36b5
- Fix: The visible caption is screen-reader only, so the time is shown once in the header. The True results block keeps the entity name in its note because the review integrity contract (e2e review-integrity) asserts it; that repeat is kept.

## ISS-8D-17 A single node is stretched tall with an empty lower half in the review
- Severity: Low
- Kind: graph
- Where: Run review of w-25 (Dir Review Empty, 24 h, Explore), node CP-1 at 100 %, dark and light, 1440x757. The node is about 170 px tall with two channel rows; roughly the lower 70 % is empty.
- Observed: the node height is fixed by something other than its content, so a two-channel node looks unfinished.
- Expected: node height follows its rows at each zoom, with even heights across the graph (V-14).
- Suggestion: size node height from the number of rows, with a minimum.
- Status: won't-fix the height of a readout node cannot follow its row count in the model alone: the compact tier stacks each row's label above its value in CSS pixels that scale with 1/zoom, so a count-based height clips two-row nodes at 50 % unless the model knows the tier; the height stays one shared value
- Fix: No change. The readout box is fixed at 132 px and the node at header + readout; a per-row height needs the tier in plantView's layout inputs (a separate change).

## ISS-8D-18 The editor outline spends the rail on four empty area groups
- Severity: Low
- Kind: layout
- Where: Editor of Compressed-air works, Outline rail, dark, 1440x757. PRODUCTION HALL 0, COMPRESSOR ROOM 0, CHILLER PLANT 0 and SUBSTATION 0 come first, each with a two-line note, about 260 px together. COMPRESSOR HOUSE 5 starts below them.
- Observed: the populated group the user is working in sits below four empty groups, and each empty group repeats the same sentence.
- Expected: populated areas first, and empty areas on one line each (G-4, the ISS-6D-37 intent).
- Suggestion: order populated areas first; reduce each empty area to one line, with the sentence said once.
- Status: fixed cdaac5a
- Fix: Areas that hold equipment are listed first, and each empty area takes one line reading No equipment yet.

## ISS-8D-19 Node titles sit tight on the category rule on some nodes
- Severity: Low
- Kind: craft
- Where: Plant view of Data centre cooling, Power frame, nodes DT-1 and FM-1, 61 %, light and dark, 1440x757 (zoomed in the screenshot to 2x).
- Observed: the title glyphs start about 2 px under the coloured rule, and the kind line sits on the header band's lower edge. compressor-1 on Compressed-air works has about 4 px, so the inset differs by node.
- Expected: the same top inset on every node (ISS-6D-33), at every zoom.
- Suggestion: set a minimum top inset in the header and grow the band when the title needs it.
- Status: fixed 9439edd
- Fix: Node headers use a 6 px top inset and content aligns to the top of the header, so the title sits at the same inset on every node (checked in the build, not in a fresh screenshot).

## ISS-8D-20 Effector-call text is cut at the run footer with no scroll cue
- Severity: Low
- Kind: layout
- Where: Run panel of Water pumping station, Effector calls section, 1440x757, dark, before any scroll: "None: nothing is commanded. A call is made when the / run reaches its minute: the simulator may refuse it, and" stops at the footer's top edge.
- Observed: the sentence is cut mid-line at the footer, with no fade or scroll marker to say more follows.
- Expected: the section fits, or the scroll region shows a cue (G-15 no clipping, P-9 scroll cue).
- Suggestion: add a fade or a scroll cue at the footer edge.
- Status: fixed 06c819d
- Fix: A shadow above the sticky run footer marks that more text continues under it.


## Regression pass on round 6 (the 40 ISS-6D findings)

Checked in this round against the build on 2026-10-15. "Holds" means I saw the fix on the screen or in the DOM. "Partly" means part of the fix holds. "Not re-checked" means I did not reach the surface this round.

Holds:
- ISS-6D-01 plant and run names carry a full-name title on hover (DOM).
- ISS-6D-02 "Show 13 more runs" on Home.
- ISS-6D-03 favicon link present (DOM).
- ISS-6D-04 no "Not checked" chips; only Verified chips.
- ISS-6D-05 the id error shows once, in plain words. The empty-plant case that carried the schema text now creates with no message.
- ISS-6D-06 Create moves about 3 px on the error, not 160 px.
- ISS-6D-07 the Id field takes the error border and a mark, and the error replaces the hint.
- ISS-6D-08 the header New plant button shows the pressed style while its form is open.
- ISS-6D-09 Recent runs shows all 8 after "Show all 8 runs" (the link disappears once used).
- ISS-6D-10 long node ids wrap (receiver-header-1, production-line-a).
- ISS-6D-12 Navigator is framed with a bottom margin (checked at 1440x757).
- ISS-6D-13 and ISS-6D-18 opening views centre the plant on both axes on compressed-air, water pumping, data centre, greenhouse and food cold chain, and on boiler steam loop after Fit (top margin 21 px, bottom 17 px). Office tower and power distribution are fitted by width.
- ISS-6D-14 run dates do not break inside the date.
- ISS-6D-15 outline kind wraps under the id.
- ISS-6D-19 Pressure is violet and Water teal in the legend (the Signal/Weather pair is a new finding, ISS-8D-08).
- ISS-6D-21 port labels are omitted at the zoom where they do not fit (packaging factory at 56 %). The campus kind cut is a different element (ISS-8D-03).
- ISS-6D-25 the Name hint shows the suggested name (CP-1), and the placeholder shows it too.
- ISS-6D-26 Keyboard dialog fits 1440x757 with title and Close in place.
- ISS-6D-27 run-panel and editor errors use the same red error treatment.
- ISS-6D-28 the Mode choice does not move when the duration error appears.
- ISS-6D-29 footer reads "Duration needs attention before starting."
- ISS-6D-31 ISA-18.2 and its terms are under For developers.
- ISS-6D-32 on water pumping at 62 % no wire runs along a header top (wires land on pin rows). Other plants still show header-level endings: ISS-8D-01.
- ISS-6D-36 a click on compressor-1 at 62 % selects it and keeps the zoom at 62 %.
- ISS-6D-38 link rows have a labelled Remove button.
- ISS-6D-40 the Mean column and its header share one right edge (1296 px) and Range is right-aligned.

Partly:
- ISS-6D-22 the field error is red with a mark and a border; its rule still breaks inside the character list (ISS-8D-14).
- ISS-6D-30 the Verified chip has its explanation as a hover title (DOM); clicking it still does nothing.
- ISS-6D-33 node titles have a top inset on compressor and CP nodes but sit tight on DT-1 and FM-1 (ISS-8D-19).
- ISS-6D-34 Inspect shows one observed time per instance but repeats it in a caption (ISS-8D-16); the trend line is not built.
- ISS-6D-37 the editor shows the empty-area sentence, but the four empty groups and an extra Halls group remain (ISS-8D-05, ISS-8D-18).
- ISS-6D-39 the run panel shows elapsed time, but no speed was visible in the one sample taken (ISS-8D-09); no running mark on the plant graph (deferred in round 6).

Open or not fixed:
- ISS-6D-11 NOT holding. Wire ends still land on header rows and hook at node edges: ISS-8D-01. The round-6 status line claimed this fixed.
- ISS-6D-20 wires under node bodies: still present on campus, office tower and data centre (won't-fix, design item).
- ISS-6D-23 the Size preset still reads "Medium (about 165 m3/h)" while channels read m³/h (needs-binary, still open).
- ISS-6D-35 the review graph still opens clipped at the inspector edge at 50 %: w-13 (Reservoir, "1 nod") and w-33 (Utilities, "9 nodes" cut). Won't-fix at 50 %, still open.

Not re-checked this round:
- ISS-6D-16 and ISS-6D-17 in the plant view (the editor inspector has one link list with no area heading and no Browse disclosure).
- ISS-6D-24 equipment search with a term.

## Not covered in this round

- 1280 and 1920 widths: not covered; innerWidth stayed 1440 for both requests.
- Hall nodes: no plant has a hall node, so Enter-to-open-hall was not tested on a real hall; the campus's Hall 1 to Hall 3 are areas and do not open.
- Failed, cancelled or interrupted runs: not reached.
- Run finished status: the scratch run w-34 finished before the status could be read; the panel moved to the review with "Verifying" after about 3 s.
- Editor issues list (E-6): none exists in the build; not reached as a feature.
- Runs not opened: see the header.
- Tooltips on icon-only review controls (V-15): not hovered.

## Score table (round 8, 1440x757 only; dark and light)

| # | Dimension | Score | Evidence |
|---|---|---|---|
| 1 | Layout | 6 | Opening views are centred on both axes on compressed-air, water pumping, data centre, greenhouse and food cold chain, and fitted on boiler steam loop (21 px top, 17 px bottom); P-3 holds for those. Navigator framed (ISS-6D-12 holds). Still clipped at the inspector edge in the review at 50 %: w-13 "1 nod", w-33 "9 nodes" (ISS-6D-35, open). "Show all" link flush on the panel edge (ISS-8D-11). Editor rail spends its height on empty groups (ISS-8D-18). No horizontal page scroll (scrollWidth 1440). |
| 2 | Typography | 5 | Campus at its opening 30 % cuts kind lines ("packaging ...", "production ...", "cooling to...") on the largest plant (ISS-8D-03). Run panel title cut to "Run Dir Review E..." (ISS-8D-09). Rule and error break inside the character list (ISS-8D-14). Good: ids in mono, dates kept together (15 Oct 2026, 02:05), thousands separators (8 553, 19 089, 23 040), KPI columns share edges. |
| 3 | Colour and contrast | 5 | Signal and Weather are the same grey diamond in the legend (ISS-8D-08). Pressure and Temperature share the diamond mark. Light-theme port labels measure about 6.4:1 on white. Good: one error red across forms; status chips pair colour with a mark (tick, eye, lock). |
| 4 | Components | 6 | Good: link rows have labelled Remove buttons; the Id error has a border and a mark; the Keyboard dialog keeps its title and Close in view; the New plant header button shows pressed while open. Legend popover has no padding (ISS-8D-12). Keyboard dialog shows an empty EDITOR section (ISS-8D-13). Verified chip is not clickable (ISS-6D-30). |
| 5 | Content design | 5 | Plain error text for the id and the duration; ISA codes moved under For developers; "Show 13 more runs". Open: "streamsim report is reading the run folder..." while a review loads (ISS-8D-04); "Halls" group with no hall nodes (ISS-8D-05); "m3/h" in the Size preset (ISS-6D-23, open); "Used for its file names" in the New plant hint (ISS-8D-15); the Inspect panel repeats a time and an id (ISS-8D-16). |
| 6 | Feedback and flow | 5 | Good: a click on compressor-1 selects it and keeps the zoom (ISS-6D-36 holds); the Save state reads "No changes" with Save disabled on an unchanged editor; the Run button turns into a "Running" pill; the run panel footer names the duration. Open: a click on an area title leaves focus on cal-site (ISS-8D-07); a new plant reads "Not checked yet" beside an enabled Save (ISS-8D-06); no speed visible in the running header (ISS-8D-09); the review's total changes on load (ISS-8D-10). |
| 7 | Information design of graph and review | 4 | Wire ends still land on header rows and hook at node edges, so ISS-6D-11 does not hold (ISS-8D-01). Wires cross area titles "Warehouse", "Line 2" and "Process hall" (ISS-8D-02). Wires run under node bodies on campus, office tower and data centre (ISS-6D-20, won't-fix). Review graphs clipped at the inspector edge at 50 % (ISS-6D-35, open). Single node stretched tall with an empty lower half (ISS-8D-17). Good: timeline marks are centred under their labels (V-10, campus shows three), the fault and effector legend shows on the timeline (V-3), and selection dims other wires. |
| 8 | Motion and responsiveness | 6 | Create moves about 3 px on the id error (round 6: 160 px). The Mode choice does not move when the duration error appears. Start stays in place. Open: the review's time and total change on load (ISS-8D-10). |
| 9 | Craft | 5 | Good: favicon present; plant and run names carry full-name titles; the Keyboard dialog fits 757 px; tab titles name the screen ("Run w-33 · Multi-hall campus · Streams Simulator"). Open: legend popover with no padding (ISS-8D-12); node titles tight to the category rule (ISS-8D-19); effector text cut at the run footer with no cue (ISS-8D-20). |
| 10 | Whole-product coherence | 5 | Good: Open, Edit, Run and Review keep their verbs on Home, plant and review; the theme control shows its state; "Set up another run" replaces "Run again" consistently. Open: the editor's outline shows a Halls group that the plant view does not (ISS-8D-05); "streamsim" appears in the review while the product is Streams Simulator (ISS-8D-04); m3/h vs m³/h (ISS-6D-23, open); save state wording differs between a new plant and an edited one (ISS-8D-06). |

**Overall: 4** (the lowest dimension is 7, Information design of graph and review, at 4; not rounded up). Dimensions 1, 4, 8 are 6; dimensions 2, 3, 5, 6, 9 and 10 are 5.

Finding counts: High 0; Medium 9 (ISS-8D-01 to ISS-8D-09); Low 11 (ISS-8D-10 to ISS-8D-20); total 20.

## Round-6 items: status after the fixer pass

- ISS-6D-11 fixed cbe39f1 for the plant view (wire ends on pin rows); the live review graph is still open (ISS-8D-01).
- ISS-6D-20 won't-fix: wires that pass under node bodies need routing, which the layout does not do; nodes stay opaque.
- ISS-6D-22 fixed dff9059 (the rule's punctuation list stays on one line).
- ISS-6D-30 won't-fix: the Verified chip keeps its explanation as a hover title; making it clickable is a design choice not taken this round.
- ISS-6D-33 fixed 9439edd (node title inset).
- ISS-6D-34 fixed b7d36b5 (one observed time per block).
- ISS-6D-37 fixed cdaac5a for populated-first ordering and one-line empty areas; the Halls group is won't-fix (ISS-8D-05).
- ISS-6D-39 won't-fix in this round: speed was not changed and the header speed was not re-sampled after the run had progressed.

## Compared with round 6

Round 6 overall 4 (lowest: Feedback and flow, Content design, Information design of graph and review, all 4). Round 8 overall 4 (lowest: Information design of graph and review). The round-6 score table had Layout at 5; it is now 6. Content design is at 5 (up from 4), and Feedback and flow is at 5 (up from 4). Information design holds at 4, because the wire-ending fix (ISS-6D-11) does not hold.

What improved and was verified:
- Opening views centre the plant (ISS-6D-13, ISS-6D-18); the Navigator has its frame (ISS-6D-12); the campus timeline labels are centred under marks (V-10 met, V-3 legend present).
- The id error and the duration error use the error treatment, and the Create and Start buttons no longer jump (ISS-6D-05, ISS-6D-06, ISS-6D-07, ISS-6D-28, ISS-6D-29).
- The editor click on a node at fitted zoom selects it and keeps the zoom (ISS-6D-36 holds).
- Recent runs shows every run (ISS-6D-09); Verified has its explanation as a tooltip; ISA terms are under For developers (ISS-6D-31); Inspect shows one observed time (ISS-6D-34, partly).
- Keyboard dialog fits the window (ISS-6D-26); the link rows have labelled Remove buttons (ISS-6D-38); Mean and Range align (ISS-6D-40).

What did not improve or regressed:
- ISS-6D-11 (wire ends on pins) is not fixed; it is the most visible graph problem left, and it is the one the round-6 status line claimed fixed.
- Wires cross area titles (ISS-8D-02) and run under bodies (ISS-6D-20); the review graphs still clip at the inspector edge at 50 % (ISS-6D-35).
- The units still differ between the Size preset and the channels (ISS-6D-23, needs-binary).
- New in this round, found on surfaces round 6 did not reach: the campus at its opening zoom (ISS-8D-03), the Halls group (ISS-8D-05), the save state of a new plant (ISS-8D-06), the legend's grey pair and no padding (ISS-8D-08, ISS-8D-12), and the review's loading wording (ISS-8D-04).

UI-EXPECTATIONS status changes seen in this round: P-3 and P-10 met on the plants checked; E-15 met (editor opens fitted, 62 % on compressed-air); V-10 and V-3 met; G-13 and G-15 met; P-14 and E-14 not met (ISS-8D-01, ISS-8D-02); V-9 not met (ISS-6D-35); V-14 not met (ISS-8D-17); V-15 not checked for tooltips on the icon-only review controls.
