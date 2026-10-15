# UI/UX evaluation, round 6

- Director: UI/UX director (Haiku 5.5 subagent), Chrome at http://127.0.0.1:8766, own tab.
- Date: 2026-10-15.
- Bar: `docs/simulation-canvas/UI-EXPECTATIONS.md`.
- Scope reviewed: plants compressed-air-works, data-centre-cooling, office-tower-hvac, water-pumping-station,
  packaging-factory and hunt-special-names (read only); runs w-2 to w-9, w-10 to w-21 as listed on Home, and w-13 (Evaluate).
  Scratch artefacts this review created: plant dir-review-empty (Explore, one pump CP-1), runs w-22 and w-25 on it.
  The author's plants (boiler-steam-loop, greenhouse-ventilation, food-cold-chain) were not opened, edited or saved.
- Viewports covered: 1440x757 only (the real window; window.innerWidth 1440, innerHeight 757, devicePixelRatio 2, screen 1728x1117). 1280 and 1920 were NOT covered: resize_window reported success for 1280x800 and 1920x1080 but innerWidth stayed 1440; the app refuses to load in an iframe, so no width emulation was possible. Light and dark were both reviewed on 1440.
- Previous evaluation: none exists in this folder (round 6 is the first evaluation file here).

## Findings

Ordered most severe first (High, Medium, Low); IDs are in the order found.

## ISS-6D-05 The invalid-id error shows raw schema text, twice, and names the wrong problem
- Severity: High
- Kind: content
- Where: Home, New plant form, Id field typed "Bad Id_1" (Name "Dir Bad Id Test"), clicked "Create and open the editor", 1440x757, light.
- Observed: two copies of one message appear. Under Id, in grey: `missing required property "equipment"; "Bad Id_1" is not a plant id: 3 to 64 lower-case letters, digits and hyphens, starting with a letter`. Then again in red, in a block about 160 px lower, after an empty gap. The first clause is a JSON-schema error about a field the user never sees, and it leads the message.
- Expected: one message, under the Id field, that names the field and the rule in plain words, for example "The id can use lower-case letters, digits and hyphens only, 3 to 64 characters, starting with a letter." No schema property names (N-2, N-7).
- Suggestion: show the binary's id rule as the field error and drop the schema clause before it reaches the canvas, through the bridge's message mapping.
- Status: fixed 772ab3c (residual needs-binary)
- Fix: The Id rule shows once, under the Id field; the general failure no longer repeats it. Residual: when the start plant is an empty plant, the binary's message still carries the schema text 'missing required property "equipment"'. Needs-binary: component new --empty writes a composition without the required equipment key; it should write equipment as an empty list, so the start plant validates and no schema message appears.

## ISS-6D-36 In the editor, a click on a node at the fitted zoom often zooms to 100 % on site-calendar and selects nothing
- Severity: High
- Kind: feedback
- Where: Editor of Compressed-air works (http://127.0.0.1:8766/?v=6d#/edit?composition=compositions%2Fcompressed-air-works.composition.json), fitted zoom 62 %, light and dark, 1440x757. Five clicks on nodes at 62 % in the editor after fresh loads: (545, 200) and (590, 200) on compressor-1, (530, 278) on compressor-2 and (545, 200) on compressor-1 once more. Three of the five changed the zoom readout to 100 %, moved the view to the site-calendar node (dashed focus ring on it), and left the Build panel open with no instance selected. The other two did nothing visible. A click on compressor-1 at 100 % selected it correctly. The same clicks in the plant view selected at once.
- Observed: the click that jumps selects nothing under the pointer and moves focus and zoom to the first instance in the outline. The plant view does not do this: a click on compressor-1 at 62 % selects it at once. The cause is not known; the pattern is not tied to a port, since body clicks also jumped.
- Expected: a click selects the instance under the pointer and keeps the zoom; focus follows the selection (E-15 opens fitted and stays fitted; G-8 focus never lands on the wrong element after an action).
- Suggestion: make the editor's canvas pointer handling match the plant view; check the focus handler that centres the outline's first item at 100 %.
- Status: fixed bcd00b1
- Fix: A selection moves the view only when the node is off screen; the first click selects compressor-1 at the fitted zoom (verified in Chrome: 52 % kept).

## ISS-6D-01 Truncated plant and run names have no full-name tooltip
- Severity: Medium
- Kind: craft
- Where: Home (workspace lab), 1440x757, light. Plants row "Hunt & "Ålborg" pumps / 冷却 test: 01 ünï…" (hunt-special-names) and every Runs row of that plant (w-14, w-15, w-17, w-19).
- Observed: the name is cut with an ellipsis and no element in the row carries a title attribute (checked: the text-bearing spans have title empty). The rest of the name, which is the only place the long title's ending differs, cannot be read anywhere on Home.
- Expected: a truncated name shows its full text on hover (title) or on focus, as the brief's craft list asks for tooltips that add information.
- Suggestion: set title to the full plant or run name on each truncated name span.
- Status: fixed ec21031
- Fix: Plant and run names carry their full text as a title on hover.

## ISS-6D-06 The Create button jumps about 160 px when the error appears
- Severity: Medium
- Kind: motion
- Where: Home, New plant form, after the invalid-id submit above, 1440x757, light.
- Observed: before submit the "Create and open the editor" button sits at y 389 and Cancel at y 389. After the error appears the buttons sit at y 548 and Start from moves from y 281 to y 320. The error block also leaves a gap of about 40 px above it, so the target a user has just pressed moves away.
- Expected: the error takes space in a reserved slot next to its field and the buttons stay where they were (no layout shift, brief point 8).
- Suggestion: render the Id error inline under the Id field inside the same row, and keep the action row at a fixed position.
- Status: fixed b24b39a
- Fix: The Id note reserves two lines and a general failure reads below the action row, so the buttons no longer move when an error appears (verified in Chrome: Create moved about 2 px).

## ISS-6D-07 The Id field never shows an error state, and its hint about file names stays in place
- Severity: Medium
- Kind: component
- Where: Home, New plant form, Id field with "Bad Id_1", 1440x757, light.
- Observed: the field border stays the normal focus blue; it is not red and carries no icon. Its hint "Used for its file names: lower case, digits and hyphens." stays under the field, and the error text is a separate line below the hint. The field is not marked invalid by the time Create is pressed (no validation while typing either).
- Expected: an invalid field has an error border and an icon with the message beside it, and the hint is replaced by the error (W-5 asks the id be validated as the binary would before writing).
- Suggestion: validate on blur and on typing with the binary's rule, mark the field with the error state, and replace the hint while the error is shown.
- Status: fixed e6f9a09
- Fix: The Id field gets the error border and a warning mark, and the error replaces the hint. Not done: a check on each keystroke; the rule is the binary's, so the id is checked when leaving the field and on Create.

## ISS-6D-09 The plant's Recent runs lists 5 of 8 runs, with no way to see the rest
- Severity: Medium
- Kind: feedback
- Where: Plant view of Compressed-air works (http://127.0.0.1:8766/#/plant?domain=domains%2Fcompressed-air-works.domain.json), inspector, RECENT RUNS, 1440x757, light, nothing selected.
- Observed: the section header count reads "8" and the rows are w-9, w-8, w-7, w-6 and w-4. w-5, w-3, w-2 and the 8th run are missing. No "Show all", "More" or link to Runs filtered by this plant is shown (DOM check: no button or link with show, more or all in the inspector).
- Expected: a count that says 8 leads to all 8 (or a "Show all 8 runs" control, or a link to Runs filtered to this plant), as P-6 asks for recent runs with Review links.
- Suggestion: add "Show all 8 runs" that opens Runs filtered by this plant, or list up to the count shown with paging.
- Status: fixed d1c31df
- Fix: Recent runs shows five; Show all N runs reveals the rest in the inspector.

## ISS-6D-10 Node titles are cut with an ellipsis at the opening zoom
- Severity: Medium
- Kind: graph
- Where: Plant view of Compressed-air works, zoom 62 %, 1440x757, light; nodes production-line-a (Production hall A, header "production-line…") and receiver-header-1 (header "receiver-header…"); also the port "dew point…" on production-line-a.
- Observed: the node id is cut in the header band, so the one word that identifies the node is partly hidden at the zoom the plant opens at. Full ids are not shown on hover (no title). The header band is about 16 px tall, and the kind line below it is not cut.
- Expected: P-2 wants the id and kind legible at the opening zoom; an id that is cut at the opening zoom fails that (the node must widen or the id must wrap).
- Suggestion: size node width from the longest id at the opening zoom, or wrap the id to two lines, before truncating.
- Status: won't-fix (design decision)
- Fix: Tried widening each node to its longest id or pin row: it fixed the cut labels, but the opening zoom of Compressed-air works dropped from 52 % to 49 %, which hides pin labels at the opening view, so it was reverted. The zoom-compensated text (22 px cap at compact zoom) against the 232-unit node width needs a decision on node size versus opening zoom.

## ISS-6D-11 Wire ends land on node headers and loop at the left edge
- Severity: Medium
- Kind: graph
- Where: Plant view of Compressed-air works, Production hall A and Production hall B, around production-line-a (left edge, y about 380) and packaging-line-b (left edge, y about 380), 1440x757, light.
- Observed: the teal and dashed wires arrive at the left edge of each node's header band and turn back in a small loop instead of ending on a pin row. The dashed site-calendar wire also runs through the Production hall A frame area. Wires cross the header row of the node, which UI-EXPECTATIONS P-14 and E-14 name as a failure.
- Expected: a wire ends on its pin, on the side of the node that carries the port, and never on the header or in a loop (E-14, P-14).
- Suggestion: route the end point to the port's row at the node edge and keep the header band clear of wire endings.
- Status: fixed b04efd1 (partly)
- Fix: The wire anchor at the compact zoom moved from the header band to the first pin row, so wires no longer land on headers. Remaining: wires that come from the right and enter from the left still loop at the node edge; that needs a routing pass.

## ISS-6D-12 The Navigator is cut off at the bottom edge at 1440x757
- Severity: Medium
- Kind: layout
- Where: Plant view, Navigator thumbnail at bottom-left of the outline rail, 1440x757, light.
- Observed: the NAVIGATOR heading is at y 608 and its thumbnail frame runs to the window bottom (y 757). The thumbnail has no bottom edge, so the frame reads as clipped. This is the case P-10 lists as not met.
- Expected: the Navigator is fully visible at a fixed size at 1280 and 1440 (P-10), with its own bottom edge and margin.
- Suggestion: give the Navigator a fixed height that fits beneath the outline with its bottom margin, and let the outline scroll above it.
- Status: fixed 55beaf5
- Fix: The Navigator has its own frame and a bottom margin; the outline above scrolls.

## ISS-6D-13 The opening view leaves an empty band below the plant frames
- Severity: Medium
- Kind: graph
- Where: Plant view of Compressed-air works at the opening zoom 62 %, 1440x757, light; canvas area y 84 to 757.
- Observed: the frames run from about y 100 to y 485. About 270 px under them is empty grey canvas, and the Site frame sits 15 px from the top. The plant is not fitted vertically, and the fit control (the square icon right of the zoom readout) is the only way to fit it.
- Expected: the opening view fits the plant with an even margin and no large empty band (P-3, not met).
- Suggestion: fit the content box to the canvas on open, both axes, with the same margin on each side.
- Status: fixed 688ea1c
- Fix: The opening view centres the plant on both axes, so the empty band is split evenly above and below.

## ISS-6D-16 The link browser groups every link under "Site", including links between two Compressor house instances
- Severity: Medium
- Kind: content
- Where: Plant view of Compressed-air works, inspector for compressor-1 then air-dryer-1, "Browse outgoing links" (after expanding), 1440x757, light.
- Observed: under "Browse outgoing links (3)" for compressor-1 the group heading reads "Site" above "compressor-1 air out → air-dryer-1 air in" and two more links, all between instances in Compressor house. air-dryer-1 shows the same "Site" heading above "air-dryer-1 power out → power-meter-air loads in". Neither instance is in Site (the Site area holds only site-calendar).
- Expected: the group heading names the area of the instance the list belongs to, or the links are not grouped at all, so the heading is always true.
- Suggestion: group by the other end's area and name it, or drop the heading when the links are in one area.
- Status: fixed f669179
- Fix: A link group is headed by a hall only when the instance is in a hall; no Site heading for areas.

## ISS-6D-18 The fit control changes nothing on the opening plant view
- Severity: Medium
- Kind: graph
- Where: Plant view of Compressed-air works, zoom readout 62 %, fit control (square icon, right of the zoom "−  62 %  +" group), 1440x757, light.
- Observed: clicking fit leaves the zoom at 62 % and the layout identical; the frames stay at y 100 to 485 with 270 px of empty canvas below. The fit appears to scale to the width only, so it cannot show a short, wide plant in the height available.
- Expected: fit uses both axes and sets the zoom so the whole plant is visible with a small even margin (P-3).
- Suggestion: compute the zoom from the larger of the width and height ratios and centre the content box.
- Status: fixed 688ea1c
- Fix: Fit uses the same rule as the opening view, centred on both axes. On a plant whose width already sets the zoom, Fit reproduces the opening view, which is correct; the empty band was the top alignment.

## ISS-6D-20 Wires pass behind node bodies in the Office tower HVAC plant room
- Severity: Medium
- Kind: graph
- Where: Plant view of Office tower HVAC (#/plant?domain=domains%2Foffice-tower-hvac.domain.json), zoom 44 %, Plant room frame, 1440x757, light. Visible at VFM-2 and VFM-3 (the red wire from AHU-1 and the olive wire from P-3 run under their bodies), and across the CH-3 and FM-1 area (the red wire from AHU-2 and the magenta wire from AHU-2 run under the chiller and feeder meter bodies).
- Observed: a wire runs under a node body and emerges on the other side, so the eye cannot tell whether it ends at that node or passes it. The red and magenta wires from the bottom are the longest and cross the most nodes. The plant is otherwise fitted well at this size (the frames span about y 100 to 740).
- Expected: wires route around node bodies and end on pins (P-4, E-14: a wire never passes behind a node body).
- Suggestion: route wires through the gaps between nodes or along the frame edge and draw nodes above wires.
- Status: won't-fix (design item)
- Fix: Wires that pass under node bodies need a routing pass in the layout (routing around bodies, drawing nodes above wires). That is a layout design change, not a fix in this round.

## ISS-6D-21 Port names are cut inside the nodes at the opening zoom
- Severity: Medium
- Kind: graph
- Where: Plant view of Data centre cooling (zoom 61 %: WCC-2 "cond wat…", WCC-1 "cond wat…", CT-1 "wet bulb in", W-1 "dry bulb …" and "wet bulb …"); Packaging factory (zoom 56 %: AR-1 "pressure…", DRY-1 "dew poi…", PL-1 "dew poi…", PL-2 "dew poi…"); both 1440x757, light.
- Observed: the port label is cut with an ellipsis where the node is narrower than the label, so the port that a link ends on is not identifiable from the node. The full name is not shown on hover (no title) and the inspector is the only other place it appears. Node width is the same for every node of a kind, so the cut repeats across the plant.
- Expected: port names are readable on the node at the opening zoom (C-1 plain port names; P-2 legibility), or the node widens to its longest port label.
- Suggestion: size each node's width to its longest port label at the opening zoom and show the full port name on hover.
- Status: won't-fix (design decision)
- Fix: Same cause as ISS-6D-10: port labels are sized in zoom-compensated text against a fixed node width; see that entry.

## ISS-6D-22 Field errors in the editor read as help text, not errors
- Severity: Medium
- Kind: component
- Where: Editor of dir-review-empty (scratch plant), instance CP-1, Name field typed "bad name!" then Rename, right panel, 1440x757, light.
- Observed: the message "Not a valid name. An instance name: a letter or digit, then letters, digits and . _ : / -, with no spaces." appears in the same grey, regular weight as the normal hint beneath the field. The Name field border and the Rename control show no error state. The message breaks after "." and leaves "/ -, with no spaces." on its own line.
- Expected: an error line in the error colour with an icon, next to a field that shows the error border, and the rule in plain words without a stray line break (N-7; brief point 3, colour never the only carrier).
- Suggestion: give field errors the error token, an icon and an error border, and keep the rule as a separate hint.
- Status: fixed 6d40b72
- Fix: An invalid name shows in the error colour with a mark, and the Name field takes the error border (shared rule for every invalid input).

## ISS-6D-27 Three different error styles for three field problems
- Severity: Medium
- Kind: colour
- Where: New plant form (red block under the form), editor Name field (grey regular text), run panel Duration field (amber text with a small dot), 1440x757, light.
- Observed: the same kind of message appears in red, in grey and in amber depending on the screen. The amber run-panel error sits in a column about 90 px wide, so "The duration is at least one minute." breaks onto three lines, with a dot about 4 px wide as its only icon.
- Expected: one error treatment (error token, icon, the field's error border, text in the same column as the field), used on every form (brief point 3 and 10).
- Suggestion: define one field-error component and use it in the New plant form, the editor and the Run panel, with the message spanning the field's width.
- Status: fixed 6d40b72, ba02b47
- Fix: Run-panel problems use the same error colour as the other forms (was amber); invalid inputs share the error border everywhere.

## ISS-6D-30 The Verified badge does not explain itself
- Severity: Medium
- Kind: feedback
- Where: Run review of w-22 (http://127.0.0.1:8766/#/run?world=w-22), sub-bar "Verified" chip beside "seed 1 · 30 min", 1440x757, light.
- Observed: hovering shows no tooltip and clicking does nothing (the chip has no title, no aria-label and no cursor). The explanation exists only as the collapsed line "Verified: the files match and the replay reproduces." in the report.
- Expected: the badge explains what Verified covers, on hover or click (V-1), or opens the integrity line.
- Suggestion: give the badge a title and make it open the integrity line in the Report.
- Status: fixed a56a803
- Fix: The Verified (and other status) chip carries the integrity line as its hover title. Clicking does not open the Report line; not added in this round.

## ISS-6D-31 ISA-18.2 codes and alarm terms remain in the Alarms section
- Severity: Medium
- Kind: content
- Where: Run review of w-22, Report, ALARMS section, right of the heading "ISA-18.2 measures"; row notes "Overload spells ... (ISA-18.2: floods)", "Repeating alarms ... (ISA-18.2: chattering)", "Stale alarms ... (ISA-18.2: stale)", "Still active at the end ... (ISA-18.2: standing)", 1440x757, light.
- Observed: the sealed-truth and alarm-standard wording from ISS-5H-05 was removed from the headings, but the standard code and its jargon words (floods, chattering, standing) remain in the heading and in the parentheses of four rows.
- Expected: the plain words for each measure with the standard reference in the details disclosure only (G-6: no jargon in headlines).
- Suggestion: move "ISA-18.2" and the parenthetical words into the "For developers" disclosure and keep the plain description in the row.
- Status: fixed e0bc038
- Fix: Alarm rows use plain words; ISA-18.2 and its terms are under For developers in the Alarms section.

## ISS-6D-32 Wires run along the top edges of node headers in the Water pumping station
- Severity: Medium
- Kind: graph
- Where: Plant view of Water pumping station (#/plant?domain=domains%2Fwater-pumping-station.domain.json), zoom 62 %, Pump house frame, dark theme, 1440x757. The yellow wire from FM-1 runs along the top edge of the CP-1, PCV-1 and VFM-2 headers at about y 157.
- Observed: one power wire is drawn on the top border of three node headers in a row, so the header reads as a wire and the node edge is lost. The wire is not visibly attached to a port on those headers.
- Expected: P-14 (wires do not cross node headers) and E-14 (wires end on their pins, not behind or along a node).
- Suggestion: route a wire through the gap between header rows, or end it on the port row of the node it feeds, and draw node headers above wires.
- Status: fixed b04efd1
- Fix: Same anchor change as ISS-6D-11: wires attach to the first pin row, not along the header edge.

## ISS-6D-33 Node titles sit on the node's top border with no padding
- Severity: Medium
- Kind: craft
- Where: Plant view nodes in light and dark, for example compressor-1 and compressor-2 (Compressor house), production-line-a (Production hall A), CP-1 (Water pumping station), at 62 % zoom, 1440x757.
- Observed: the id text ("compressor-1") starts at the top border of the node: its ascenders touch the coloured top rule and nothing sits above it. The header band is about 16 px tall, so the title takes the whole band and no margin is left above it. The kind line below it is set with more room, so the header looks unbalanced.
- Expected: the title has the same top inset as the kind and port rows, and the coloured category rule sits above the header band with clear space.
- Suggestion: add 4 px of top padding inside the node header and move the category rule outside the title area.
- Status: fixed a41208e
- Fix: The node header has 4 px of top padding under the category rule.

## ISS-6D-35 The review graph opens clipped at the inspector edge at 1440x757
- Severity: Medium
- Kind: graph
- Where: Run review of w-13 (Water pumping station, Evaluate), graph at 50 % zoom opened fitted, http://127.0.0.1:8766/#/run?world=w-13, dark theme, 1440x757. The Reservoir frame and its "1 node" label run under the inspector edge at x 1066 and the label reads "1 nod".
- Observed: the fit at the opening zoom does not include the whole plant; the rightmost frame is cut by the panel divider, and its label is cut mid-word. Nothing in the graph indicates more content to the right.
- Expected: the review graph opens fitted with all frame labels visible and no clipping at 1280 or 1440 (V-9, still open; "nothing clipped" in brief point 1).
- Suggestion: fit the content box to the visible canvas width minus a margin, and add a right-edge margin inside the canvas.
- Status: fixed b342b26, 7250ae3 (partly)
- Fix: The run graph's column was growing under the inspector; its grid is now constrained, so the graph stops at the panel edge. Remaining: at the 50 % minimum fit the Water pumping station is still wider than the canvas, so its right frame is cut until the user pans. A lower minimum would hide pin labels (overview tier), so the minimum stays.

## ISS-6D-37 The editor outline lists four empty areas that the plant view does not show
- Severity: Medium
- Kind: content
- Where: Editor of Compressed-air works, Outline rail, 1440x757, dark and light: groups PRODUCTION HALL 0, COMPRESSOR ROOM 0, CHILLER PLANT 0 and SUBSTATION 0. The plant view of the same plant, Outline, shows four areas: Site 1, Compressor house 5, Production hall A 1, Production hall B 1.
- Observed: the editor's outline carries eight area groups, four with no instances and no empty-state text, and the plant view carries four. The two surfaces disagree about what areas the plant has (E-1 says the editor looks like the plant screen it edits).
- Expected: the same area list on both surfaces, and empty areas either hidden or shown with a line that says what goes there and how to add equipment (G-4).
- Suggestion: hide zero-count areas in the editor outline or show them as "No equipment yet" with an Add action; make the plant view use the same list.
- Status: fixed 1d6efd5
- Fix: Empty areas in the editor outline say 'No equipment yet. Add some in the Build panel.' The editor keeps showing them on purpose, so it has more groups than the plant view.

## ISS-6D-39 The running panel shows no speed, and the graph shows no running indicator
- Severity: Medium
- Kind: feedback
- Where: Run panel while Dir Review Empty runs w-25 (24 h, Explore, no faults), 1440x757, dark, screenshot taken immediately after Start. The top bar pill reads "Running Dir Review Empty · 01:12:00 of 24:00:00"; the panel reads "01:12:00 of 24:00:00 simulated" with a thin progress bar; the plant bar shows "Running · 01:12:00"; the CP-1 node in the graph looks unchanged.
- Observed: the panel gives progress and the planned time but no simulated seconds per second, and no fault ticks (there are none to tick). The graph carries no running indicator at normal scale. The run finished within about half a minute and opened its review; no "Run w-25 finished" status was visible before the review took over (R-5 asks for it).
- Expected: the running panel shows the speed (simulated seconds per second), the graph shows a running state on the plant (R-4, R-9), and the finished status stays long enough to read before the review opens (R-5).
- Suggestion: show the speed beside the progress bar, add a running mark to the plant frame or node header, and hold the finished status for a few seconds or until the user dismisses it.
- Status: won't-fix (design item)
- Fix: The speed already appears in the Running line once a second of wall time has passed (launch panel, speedOf), so the screenshot was taken too early. Not done: a running mark on the plant graph and holding the finished status before the review opens; both are design changes.

## ISS-6D-02 "Show 1 more of 1" is awkward wording
- Severity: Low
- Kind: content
- Where: Home, Runs card footer, below w-2 at 1440x757, light.
- Observed: the link reads "Show 1 more of 1". "of 1" repeats the count in a form that reads as a fraction, and the singular is not handled.
- Expected: "Show the last run" or "Show 1 more run", with the same noun as the Runs heading.
- Suggestion: pluralise and name the noun: "Show 1 more run" for a count of 1, "Show 25 more runs" otherwise.
- Status: fixed 30f6d1d
- Fix: Reads Show N more runs, or Show 1 more run for the last one.

## ISS-6D-03 No favicon on the workspace page
- Severity: Low
- Kind: craft
- Where: Home at http://127.0.0.1:8766/, browser tab, 1440x757, light.
- Observed: the document has no link rel=icon (checked in the DOM: none). The tab shows the browser default page icon. The tab title is correct ("lab · workspace · Streams Simulator").
- Expected: the product mark (the pulse icon in the top bar) as the tab icon, on every screen.
- Suggestion: add a favicon link built from the top-bar logo mark.
- Status: fixed 664896b
- Fix: Inline SVG favicon (the top-bar mark on the accent colour) in index.html, no external asset.

## ISS-6D-04 The "Not checked" chip repeats on all 21 run rows and carries no information
- Severity: Low
- Kind: component
- Where: Home, Runs card, every row (w-2 .. w-21), 1440x757, light.
- Observed: all 21 rows show the same grey "Not checked" chip beside an identical "Review" button. The column is the same on every row, so it adds noise and looks like a disabled state.
- Expected: an unchecked run says so once (or the chip is shown only when a verdict exists), so the eye lands on the runs that have a verdict (W-3 asks for the verdict once checked).
- Suggestion: hide the chip when there is no verdict, or show a single column header "Verdict" with the value only when it is known.
- Status at end of review: not present on the later build (10 plants, 25 runs, dark). The rows show no chip, and the list carries the note "Runs are checked when you open their review." Treated as resolved on that build.
- Status: fixed 2fb8826
- Fix: The Not checked chip is gone from rows; one muted line under the Runs heading says runs are checked when their review opens. Verified, Not verified and Did not finish keep chips.

## ISS-6D-08 "New plant" stays a filled primary button while its form is open
- Severity: Low
- Kind: component
- Where: Home, Plants card header, "+ New plant" with the form open, 1440x757, light.
- Observed: the header button keeps the solid blue primary style and the same label while the form it opened is showing. Clicking it again does nothing visible.
- Expected: while the form is open the header shows a pressed or secondary state, or the button is hidden, so there is one primary action at a time (brief point 4, states).
- Suggestion: style the header button as pressed (or as Cancel) while the form is open and make the second click close the form.
- Status: fixed d3927a5
- Fix: The header New plant button takes the pressed (secondary) style while its form is open (verified in Chrome).

## ISS-6D-14 Recent-run rows break the date in the middle
- Severity: Low
- Kind: typography
- Where: Plant view inspector, RECENT RUNS rows, 1440x757, light, e.g. "w-9 · Explore · 30 min · seed 1 · 1 fault · 15 Oct / 2026, 02:05".
- Observed: each row wraps to two lines and the wrap falls inside the date ("15 Oct" on one line, "2026, 02:05" on the next). The Review link sits at the right on the first line, so the two lines read as two unrelated facts.
- Expected: the date keeps together (non-breaking space), and the metadata wraps at a separator, with the Review link aligned to the title row.
- Suggestion: use a non-breaking space inside the date and put the date on its own line beneath the run's mode and faults.
- Status: fixed 2b30851
- Fix: The run date keeps its spaces non-breaking, so it no longer breaks between day and year.

## ISS-6D-15 Outline kind text is cut in the rail
- Severity: Low
- Kind: typography
- Where: Plant view, left outline, row production-line-a with kind "production ..." under PRODUCTION HALL A, 1440x757, light.
- Observed: the kind after the id is cut to "production ..." while the rail has empty room to its right; the id and kind share one line and the kind is the part that is lost.
- Expected: P-5, the kind in words is visible in the outline; when the rail is too narrow, the kind wraps under the id, not cut.
- Suggestion: put the kind on a second line in the outline row, in the same grey as the other secondary text.
- Status: fixed 64830c4
- Fix: Outline kind wraps under the id instead of being cut.

## ISS-6D-17 "Browse incoming and outgoing links" repeats the list directly above it
- Severity: Low
- Kind: component
- Where: Plant view, inspector for compressor-1 and air-dryer-1, LINKS IN and LINKS OUT sections, 1440x757, light.
- Observed: LINKS OUT lists "air out → air-dryer-1 air in", "air out → receiver-header-1 air in", "power out → power-meter-air loads in"; expanding "Browse outgoing links (3)" lists the same three again, with the instance name in front and a "Site" heading. The two forms use different wording for the same link. The disclosure therefore shows nothing new.
- Expected: one list, with each row selecting its other end (P-7), and the disclosure removed or used for the extra detail only.
- Suggestion: make the rows in LINKS IN and LINKS OUT the clickable list and drop the second disclosure.
- Status: fixed f669179
- Fix: The inspector lists the links once, as the clickable rows; the Browse disclosure that repeated them is removed.

## ISS-6D-19 The Water and Pressure wire colours look alike in the legend
- Severity: Low
- Kind: colour
- Where: Plant view, Legend popover ("Wire colour and pin shape by port type"), 1440x757, light.
- Observed: the Pressure swatch (teal diamond) and the Water swatch (teal circle) are close in hue, and Condenser water (green) and Power (olive) are close as well. Two of the nine wire families differ mainly by pin shape, which the legend states only for the pin end.
- Expected: wire families have distinct hues in both themes, and the legend says which mark is the wire and which is the pin (colour never the only carrier, but here the colours barely separate).
- Suggestion: move Pressure to a hue away from Water (for example a violet) and keep the shapes as the second cue.
- Status: fixed 79561fe
- Fix: Pressure is violet and power is amber, in both themes, so no two wire families share a hue.

## ISS-6D-23 Units differ between the Size preset and the inspector
- Severity: Low
- Kind: content
- Where: Editor, Build panel, Size select after choosing Centrifugal pump: "Medium (about 165 m3/h)"; plant inspector for compressor-1 shows "air flow · m³/h". Light theme, 1440x757.
- Observed: the preset says "m3/h" and the channel says "m³/h". The same unit is written two ways.
- Expected: one unit form everywhere (G-7 and G-3: units set as "12.9 °C" with one notation).
- Suggestion: take the preset label from the same unit formatter the inspector uses.
- Status: needs-binary
- Fix: The Size preset title comes from the binary's built-in library (for example 'Medium (about 165 m3/h)'). Needed: the library writes the preset title with the display unit m³/h, the same notation the vocabulary uses for its channels.

## ISS-6D-24 The equipment search keeps a full-height result box for one match
- Severity: Low
- Kind: layout
- Where: Editor, Build panel, "Find equipment" search typed "pump", 1440x757, light.
- Observed: the result list keeps its full height (about 135 px) with one row, "Centrifugal pump", and empty space beneath it, so the list looks empty and the description block that follows is pushed down.
- Expected: the result list sizes to its matches (or to a few rows), with an empty state when nothing matches (G-4).
- Suggestion: size the list to its content up to a maximum height, and show "No equipment matches 'xyz'" when there are none.
- Status: fixed b830f94
- Fix: The equipment result list is as tall as its matches, up to six rows; the empty state line already existed.

## ISS-6D-25 Build panel copy uses internal words and a broken sentence
- Severity: Low
- Kind: content
- Where: Editor, Build panel, empty plant, 1440x757, light: "Used in files and links. Leave empty to use the suggestion." under Name, and "The areas are this plant's own, the site first. Add more under Add area." under Area.
- Observed: "files" is an internal term for ids. "the site first" is an uncertain sentence fragment. The Area select's hint sits inside the Build panel as a paragraph and reads as a rule.
- Expected: plain sentences that say what the user does: "Leave empty to use the suggested name, CP-1." and "Areas are listed in order, with the site first. Add more under Add area."
- Suggestion: rewrite both hints in plain verbs and show the actual suggested value in the field help.
- Status: fixed 78a2c36
- Fix: Name hint reads 'Leave empty to use the suggested name, CP-1.'; the area hint reads 'Areas are listed in order, with the site first.'

## ISS-6D-26 The Keyboard dialog's title and Close scroll out of view at 1440x757
- Severity: Low
- Kind: layout
- Where: Editor (dir-review-empty), Keyboard dialog opened from the top bar, 1440x757, light; after scrolling the dialog's list.
- Observed: the dialog is taller than the window, so its list scrolls inside it. The "Keyboard" title and the list's first rows leave the view, and only the Close button remains at the bottom. The scroll position also rests on a partly cut row ("Tab" row under the top edge).
- Expected: the dialog title and Close stay in place while the key list scrolls, or the dialog fits the window at 1280x800 and 1440x757 (G-15, no clipping).
- Suggestion: make the header sticky and let only the list scroll, or tighten row padding so the list fits 757 px.
- Status: fixed aafed09
- Fix: The Keyboard title and Close stay in place; only the key list scrolls.

## ISS-6D-28 The Run panel's Mode section moves 60 px when the duration error appears
- Severity: Low
- Kind: motion
- Where: Run panel of Dir Review Empty, Duration set to 0 then Start pressed, 1440x757, light. MODE heading moves from y 264 to y 324; Explore and Evaluate move with it.
- Observed: the Start button and the footer stay where they were, which is right, but the controls above the footer jump by the height of the three-line error.
- Expected: the error takes its space without moving the mode choice (R-10 asks that Start does not move; the same reasoning covers the choices above it).
- Suggestion: put the duration error on one line beside the field or reserve its height.
- Status: fixed ba02b47
- Fix: The duration problem takes one line in the field column and its slot is always reserved, so the Mode choice no longer moves. Not re-measured in Chrome.

## ISS-6D-29 The Run panel footer says "1 entry needs attention" instead of naming the field
- Severity: Low
- Kind: content
- Where: Run panel, footer above "Start the run", after Duration is set to 0, 1440x757, light.
- Observed: "1 entry needs attention before starting." The footer does not name the field, and "entry" is a schema word.
- Expected: "Fix the duration to start: it must be at least one minute." or "Duration needs attention before starting." (R-6, N-7).
- Suggestion: name the field in the footer summary and use the field's label.
- Status: fixed 11f7858
- Fix: The footer names the field: 'Duration needs attention before starting.', or the list of fields when there are several.

## ISS-6D-34 The Inspect tab repeats the same timestamp on every channel and the same explanation for each node
- Severity: Low
- Kind: content
- Where: Run review of w-22, Inspect tab, selected CP-1 (http://127.0.0.1:8766/#/run?world=w-22), 1440x757, light and dark. Rows "flow 165 m³/h 00:00:15" and "power 16.4 kW 00:00:15", then "TRUE RESULTS: What really happened in the plant. An outside system could not see it while the run was live; it is shown here once the run has ended."
- Observed: each observed row carries its own timestamp, and during playback the channels show different times in the same panel (flow at 00:02:20, power at 00:02:15, panel header at 00:02:21). The true-results paragraph repeats the Report's paragraph for every instance. Inspect has no trend line, no fault or alarm marks and no channel kind or area (V-5 still partly open).
- Expected: one timestamp in the panel header for the playback time, a per-channel last-reported time only when it differs, a short trend, and the paragraph shown once in the Report (V-5).
- Suggestion: show the playback time once, add a compact trend with the cursor, and drop the repeated paragraph from Inspect.
- Status: fixed f762224 (partly)
- Fix: The true-results paragraph is shown once, in the Report, not for every instance in Inspect. Not done: the per-channel time column and a trend line (a new series view, not a fix).

## ISS-6D-38 The link rows in the editor's inspector use icon-only delete buttons
- Severity: Low
- Kind: component
- Where: Editor of Compressed-air works, compressor-1 inspector, LINKS (3) rows "air out → air-dryer-1 air in", "air out → receiver-header-1 air in", "power out → power-meter-air loads in", 1440x757, dark.
- Observed: each row ends in a trash icon with no visible label, and the three icons look alike. The row text names the link, but the delete action is not named where it sits.
- Expected: a labelled control ("Remove link") or a visible tooltip and an accessible name on each row (E-13, C-1; brief point 4, icon plus label).
- Suggestion: label each row's delete control with its text or give it a visible tooltip with the same wording.
- Status: fixed 3f1996d
- Fix: Each link row has a visible Remove button (accessible name unchanged).

## ISS-6D-40 KPI units sit in a different font and above the number's baseline
- Severity: Low
- Kind: typography
- Where: Run review of w-13 and w-22, Report, TRUE RESULTS KPI table, and the Inspect tab KPI table, 1440x757, dark.
- Observed: the Mean value "0.10" is set in mono, and its unit "kWh/m³" is a separate small sans run that sits visibly higher than the number's baseline. "Mean" starts at the number's left edge while "Range" is right-aligned, so the three columns do not share one alignment edge.
- Expected: numeric columns share one alignment edge (V-11), and the unit sits on the number's baseline in the same font as the label beside it.
- Suggestion: right-align Mean and Range values and headers to one edge, and set the unit inline with the number at the same baseline.
- Status: fixed 7250ae3
- Fix: Mean and Range values share one right edge; units sit on the number's baseline.

## Score table (round 6, 1440x757 only; light and dark)

| # | Dimension | Score | Evidence (measured, with finding IDs) |
|---|---|---|---|
| 1 | Layout | 5 | P-3 still open: the Compressed-air plant opens with about 270 px of empty canvas under the frames, and Fit changes nothing (ISS-6D-13, ISS-6D-18). Navigator is cut at the window bottom (ISS-6D-12). The review graph opens clipped at the inspector edge (ISS-6D-35). Good: no horizontal page scroll on Home, plant, editor or review at 1440. |
| 2 | Typography | 5 | Node ids and port names cut with ellipses at the opening zoom (ISS-6D-10, ISS-6D-21). Inspector run dates break inside the date (ISS-6D-14). KPI units in a different font above the number's baseline (ISS-6D-40). Good: ids in mono, times as hh:mm:ss, thousands separators (23 040, 8 553) and dates "15 Oct 2026, 02:05" used consistently. |
| 3 | Colour and contrast | 5 | Dark review text measured 6.5:1 to 14.5:1 for 11 to 13 px text, so text contrast passes. Three field-error colours (ISS-6D-27). Water and Pressure wire families look alike in the legend (ISS-6D-19). Good: run mode chips pair colour with an icon (eye for Explore, lock for Evaluate); Verified uses a tick. |
| 4 | Components | 5 | Icon-only delete in link rows (ISS-6D-38). Browse-links disclosure repeats the list above it (ISS-6D-17). Field errors look like help text (ISS-6D-22). Good: a visible blue focus ring on keyboard focus (checked on Edit, dark); Keyboard dialog has a scrim, a Close button and Escape closes it, returning focus to its button. |
| 5 | Content design | 4 | A raw schema message ("missing required property \"equipment\"") shown twice for an invalid id (ISS-6D-05, High). ISA-18.2 codes and "floods, chattering, standing" left in the Alarms section (ISS-6D-31). "Site" heading over links between Compressor-house instances (ISS-6D-16). "Show 1 more of 1" (ISS-6D-02). "1 entry needs attention" (ISS-6D-29). Good: the run plan and effector copy is plain sentences. |
| 6 | Feedback and flow | 4 | In the editor a click on a node at fitted zoom sometimes selects nothing and jumps the view to 100 % on site-calendar, 3 of 5 clicks (ISS-6D-36, High). Verified badge does not explain itself (ISS-6D-30). Recent runs shows 5 of the 8 runs with no path to the rest (ISS-6D-09). Running panel has no speed and the graph no running indicator (ISS-6D-39). Good: Save states read "Unsaved changes" and "Ready to run. Saved."; "Stop now" says what is kept; a top-bar pill follows the run. |
| 7 | Information design of graph and review | 4 | Wires run along node header edges and through headers (ISS-6D-11, ISS-6D-32), and behind node bodies (ISS-6D-20). Node titles touch the top border (ISS-6D-33). Review graph clipped (ISS-6D-35). Good: the timeline legend (fault starts, effector call), mark shapes differ, and mark labels centre under their marks (V-3, V-10 met). Live readouts show in the nodes. Selecting a node dims the other wires. |
| 8 | Motion and responsiveness | 5 | The Create button moves about 160 px when the id error appears (ISS-6D-06). Run-panel controls move 60 px on a duration error (ISS-6D-28). Good: playback and selection respond at once, and the run-panel footer stays fixed. |
| 9 | Craft | 5 | No favicon (ISS-6D-03). Truncated plant and run names have no full-name tooltip (ISS-6D-01). Keyboard dialog title and first rows scroll away (ISS-6D-26). Good: tab titles name the screen ("Run w-22 · Dir Review Empty"), and the Keyboard button shows a focus ring after Escape. |
| 10 | Whole-product coherence | 5 | Editor outline shows eight areas (four empty) while the plant view shows four (ISS-6D-37). Three error treatments (ISS-6D-27). "m3/h" in the Size preset and "m³/h" in channels (ISS-6D-23). Good: Open, Edit, Run and Review keep the same verbs on Home, plant and run. |

**Overall: 4** (the lowest dimensions are 6 Feedback and flow, 5 Content design and 7 Information design of graph and review; all three are 4). Not rounded up.

Finding counts: High 2 (ISS-6D-05, ISS-6D-36); Medium 21; Low 17; total 40.

Improvements seen against UI-EXPECTATIONS (no previous evaluation file exists, so this is against the bar):
- Met and verified: G-13 (theme shows its label), G-15 (scrim, real keys, Escape closes), P-11 (legend on plant and editor), V-3 (timeline legend), V-10 (labels centred under marks), N-3 (full copy and empty plant offered as choices), E-2 (save state wording), E-3 (library search with description and a suggested name), R-7 (focus lands on Seed), RUN-3 (stop explains what is kept), RUN-6 (top-bar pill keeps the run visible).
- Still open: P-3, P-10, P-14, E-14, V-9, V-14, V-15, V-13 (not re-checked in this round), E-13 (icon-only delete), E-6 (no issues list seen; not reached in this round).

## Not covered in this round

- 1280 and 1920 widths: not covered (see the header). The window stayed at 1440x757 for every resize request.
- Hall view (Enter opens a hall): no plant in this set has a hall node, so not tested.
- Failed or interrupted run review, and the "Run finished" status toast: not reached. The 24 h scratch run finished within about half a minute and opened its review before any toast was captured.
- Empty workspace (no plants): not reachable, the lab workspace has plants.
- Compressed-air editor "Saved but not composed" and "Not saved" states: not reached.
