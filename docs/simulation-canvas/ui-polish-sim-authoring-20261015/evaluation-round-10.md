# UI/UX evaluation, round 10

- Director: UI/UX director (Haiku 5.5 subagent), Chrome at http://127.0.0.1:8766, own tab 770933667 (tab 770933647 was not used).
- Date: 2026-10-15.
- Bar: `docs/simulation-canvas/UI-EXPECTATIONS.md`.
- Previous evaluations: `evaluation-round-6.md` (overall 4, 40 findings) and `evaluation-round-8.md` (overall 4, lowest Information design of graph and review at 4, 20 findings).
- Viewport tested: 1440 x 614 only. Proof: `javascript_tool` read `window.innerWidth` 1440, `innerHeight` 614, `devicePixelRatio` 2, and `scrollWidth` 1440 on Home and on the campus plant. No resize was requested (brief). 1280 and 1920 were NOT covered. The round-8 height of 757 px was not reproduced; the opening zoom of several plants differs at 614 px high, and that is a finding (ISS-10D-04).
- Theme: light (the machine's "System") for the ten plant views and the Home list; dark for the Home New plant form, the hunt plant, the campus legend, the review w-33, w-13, w-22 and w-35, the run panel and the Keyboard dialog. The theme was switched to Dark for the dark pass and restored to System at the end.
- Plants reviewed (all ten, opened in the plant view at their opening zoom): compressed-air-works 62 %, office-tower-hvac 41 %, water-pumping-station 60 %, data-centre-cooling 47 %, packaging-factory 49 %, boiler-steam-loop 46 %, greenhouse-ventilation 49 %, food-cold-chain 55 %, power-distribution-standby 33 %, multi-hall-campus 30 %. Also hunt-special-names (62 %, dark) and dir-review-empty. Editor opened for compressed-air-works (populated, one instance selected) and boiler-steam-loop (populated); the editor was not saved or edited.
- Runs opened in full: w-13 (Evaluate, water pumping), w-22 (Explore, dir review empty), w-33 (Evaluate, campus), w-35 (Explore, dir review empty, created in this round). Runs w-1 to w-12, w-14 to w-21, w-23 to w-32 and w-34 were checked only as rows on Home.
- Scratch artefacts created in this round: run w-35 (30 min Explore on `dir-review-empty`, started from the run panel). No plant was created or saved. The New plant form was filled with the name "Scratch" and the id "Bad Id_1" and closed with Cancel (no plant made). Plant `dir-empty-check-8` is a round-8 scratch plant that was already on Home; it was not created or changed in this round.
- Not reached: the Running state of a run (w-35 finished before the first screenshot, about 1 s after Start, so the speed and progress could not be watched); failed, cancelled and interrupted runs (none exist in the lab); the editor issues list (not in this build); the hall nodes (no plant has one).

## Findings

Ordered most severe first. Each finding was appended to this file as it was found; the numbering is the ID.

## ISS-10D-01 Kind lines print on top of the first port row on many nodes at the opening zoom
- Severity: High
- Kind: graph
- Where: Plant view of Food cold chain, 55 %, light, 1440x614 (zoom of x 320-1010, y 140-300): pump-cond "pump" over "level in"; tank-chw "storage tank" over "inflow in"; drive-chw "motor drive" over "load in"; chiller-1 and chiller-2 "chiller" over "load in"; receiver "air receiver" over "air in"; cooling-tower "cooling tower" over "heat in"; ahu-a "ahu" over "chw in"; gen-1 "generator" over "load in"; meter-main "feeder meter" over "loads in". Same pattern on Water pumping station at 60 % ("pump" over "level in" on CP-1, CP-2, CP-3; "storage tank" over "inflow in" on STW-1 and STW-2), Compressed-air works at 62 % ("air compressor" over "pressure in" on compressor-1 and compressor-2; "air dryer" over "air in"), and Hunt & "Ålborg" pumps at 62 % in dark ("cooling tower" over "heat in", "air dryer" over "air in").
- Observed: the kind line and the first port row sit at the same height, so two texts print over each other and neither reads. The kind line also sits on the header band's lower edge on several nodes. Round 8 did not record this on these plants, and round-8 ISS-8D-01 / 8D-02 are about wires, not labels.
- Expected: no two labels overlap at any zoom. The kind is in words (P-2), and the port names are legible (C-1, E-14). When a label does not fit, it is left out at that zoom, as ISS-6D-21 already does for port names.
- Suggestion: reserve a row for the kind below the header and start the first port row under it; if the row does not fit at the zoom, drop the kind line there instead of overprinting.

## ISS-10D-02 Some nodes show no kind line at all, while others show the kind cut
- Severity: Medium
- Kind: graph
- Where: Plant view of Multi-hall campus, opening zoom 30 %, light, 1440x614. No kind line on cv-h1 (conveyor), vfd-h1 (motor drive), mtr-h1 (feeder meter), comp-u1 (air compressor), rcv-u1 (air receiver), mtr-u1 (feeder meter), cal-site (site calendar) and wx-site (weather). On the same screen ct-u1 shows "cooling to…", pk-h1 "packaging …" and pl-h1 "production …".
- Observed: a kind that fits (conveyor, weather) is missing while a longer one is cut. The eye gets no rule; the node heights also differ by one line. The outline lists the kinds of these instances, so the data is there.
- Expected: every node shows its kind in words (P-2), by one rule at every node. A kind that does not fit wraps onto a second line, as the ids do (ISS-6D-10), or the omission is the same rule for all nodes.
- Suggestion: one rule for the kind line (always shown, wrapped to a second line if needed), so every node in a plant has the same header height.

## ISS-10D-03 In the live review, the kind line overprints the first readout label on most nodes
- Severity: Medium
- Kind: graph
- Where: Run review of w-33 (Multi-hall campus, Evaluate), 50 %, dark and light, 1440x614: cal-site "site calendar" over "site load"; ch-u2 "chiller" over "chw supply temp"; rcv-u1 "air receiver" over "header pressure"; ct-u1 "cooling tower" over "basin temp"; wx-site "weather" over "dry bulb temp". Run w-13 (Water pumping station, Evaluate), 50 %, dark: CP-2 "pump" over "flow", PCV-1 "control valve" over "position", VFM-2 "motor drive" over "speed", STW-2 "storage tank" over "makeup flow".
- Observed: the kind line and the readout label are set one or two pixels apart and touch. During playback the readout labels are what the user reads, so the collision is on the screen that matters most.
- Expected: the kind line and each readout label have their own row and a gap (V-4, V-14 readout nodes show their full kind and are legible).
- Suggestion: give the kind line the same row gap the readout rows have, and check the collision at the 50 % and 100 % tiers.

## ISS-10D-04 Round-8 ISS-8D-03 reproduces: the campus opens at 30 % here and cuts its kind lines; power distribution opens at 33 %
- Severity: Medium
- Kind: graph
- Where: Plant view of Multi-hall campus, opening zoom 30 %, light, 1440x614: "packaging …" (pk-h1, pk-h2, pk-h3), "production …" (pl-h1, pl-h2, pl-h3), "cooling to…" (ct-u1). Plant view of Power distribution with standby, opening zoom 33 %, light: "production li…" on line-a and line-b. Node kind text is about 8 px in the screenshot at these zooms.
- Observed: round 8 recorded the campus at 52 % on its window and marked the finding as not reproduced. On this window the same plant opens at 30 %, the graph spans x 272-1050 (the width sets the zoom), and the kind lines are cut and very small. The opening zoom depends on the window, so it is not a fixed legibility floor.
- Expected: at the opening zoom every node shows its id and its full kind in words (P-2, G-12, VIS-6).
- Suggestion: set the opening zoom from the legibility floor (the longest kind must fit at that zoom) and let the user pan, or wrap the kind line as the id is wrapped (ISS-6D-10).

## ISS-10D-05 Round-8 ISS-8D-18 does not hold: empty areas still lead the editor outline, two lines each
- Severity: Medium
- Kind: layout
- Where: Editor of Compressed-air works, Outline rail, light and dark, 1440x614: PRODUCTION HALL 0, COMPRESSOR ROOM 0, CHILLER PLANT 0 and SUBSTATION 0 come first, each with "No equipment yet" on a line of its own under the heading (about 240 px together); COMPRESSOR HOUSE 5 starts below them. Editor of Boiler steam loop, light: SITE 2, BOILER HOUSE 6, then GAS TRAIN 0 (empty) between the populated groups, then DISTRIBUTION 3.
- Observed: the populated group the user works in sits below four empty groups, and an empty group sits between two populated ones. The round-8 fix said populated areas come first and an empty area takes one line.
- Expected: areas with equipment first; an empty area on one line with the heading (G-4, ISS-6D-37).
- Suggestion: order the outline by equipment count (populated first, plant order kept within each group), and put "No equipment yet" on the heading's row.

## ISS-10D-06 Round-8 ISS-8D-02 does not hold: a dashed wire runs through an area title
- Severity: Medium
- Kind: graph
- Where: Plant view of Hunt & "Ålborg" pumps (hunt-special-names), 62 %, dark, 1440x614: the dashed wire from site-calendar crosses the frame title "Production hall A" between "hall" and "A" (zoom of x 270-470, y 380-430).
- Observed: the wire is drawn over the title, so the title reads as part of the wire. Round 8 reported that titles were moved to an opaque layer above the wires; this title is not in that layer in this build.
- Expected: area titles are never crossed by wires (P-14).
- Suggestion: draw every area title above the wires with an opaque backing, and check each title at each zoom tier.

## ISS-10D-07 The integrity detail opens by default on a new run and is closed on older runs
- Severity: Medium
- Kind: feedback
- Where: Run review of w-35 (Dir Review Empty, Explore), dark, 1440x614, opened from the finished run panel. The Report starts with "▼ Verified: the files match and the replay reproduces." and expands to REPLAY CHECK, FILE CHECK (seven rows), and DELIVERIES before DISTURBANCES and TRUE RESULTS. The same line on w-22 and w-33 is collapsed ("▸") on first open.
- Observed: the same review opens in two states. After a fresh run the whole integrity detail leads the report; on older runs it is one line.
- Expected: integrity is one collapsed line on every review (V-6), and the report leads with what happened. The default does not depend on how the review was reached.
- Suggestion: collapse the integrity block on every review by default; remember an open state only for the session.

## ISS-10D-08 The integrity detail shows file-level words and the binary's name as its main content
- Severity: Medium
- Kind: content
- Where: Report, Verified block of w-35, expanded, dark, 1440x614: the sub-label "streamsim verify" beside REPLAY CHECK; FILE CHECK rows "trace", "journal", "sealed journal", "journal file", "sealed journal file", "ledger", "history"; the note "The report compares the files in the run folder with the digests the run recorded."; DELIVERIES "emitted, delivered, dropped".
- Observed: file names, a product command and a folder word lead the detail of a verified run. This is the same class as ISS-8D-04, which the fixer changed for the loading text only.
- Expected: plain words for each check ("Replay gives the same output", "Files match the record"), with file names, digests and the command under For developers (G-6, N-7, W-7).
- Suggestion: move the file table, the command label and the digest note under For developers, and keep the one-line summary in the Verified line.

## ISS-10D-09 The Legend heading runs past the popover's right edge
- Severity: Low
- Kind: craft
- Where: Legend popover on the plant view, Hunt & "Ålborg" pumps, 62 %, dark, 1440x614: the heading "Wire colour and pin shape by port type" ends at x of about 1055, and the popover's right edge is at x 1050, so the "e" of "type" touches the border (zoom of x 830-1060, y 85-200). The body lines keep their 12 px inner padding.
- Observed: the heading has no right inset and the text meets the border.
- Expected: heading and body share the popover's padding, as ISS-8D-12 asked (N-7 and craft).
- Suggestion: let the heading wrap inside the padded box, the same as the body text.

## ISS-10D-10 The review's Keyboard list gives no key for previous or next event
- Severity: Low
- Kind: content
- Where: Keyboard dialog on run review w-35 and w-33, dark, 1440x614. The RUN REVIEW group lists only "Space: play or pause". The timeline has previous and next controls (« ‹ › ») that the list does not mention.
- Observed: the list does not name a key for stepping to the previous or next event (G-15 asks for the step keys). I did not test whether any such key works, so the finding is that the list is silent, not that the key is missing.
- Expected: the dialog names the key for previous and next event, or the control says it has none (G-15, principle 6: mouse and keyboard are equals).
- Suggestion: add the step keys to the RUN REVIEW group, or make the step controls' keyboard path visible in their tooltips.

## ISS-10D-11 Pressure, Temperature and Weather share one diamond mark in the legend
- Severity: Low
- Kind: colour
- Where: Legend popover, Hunt & "Ålborg" pumps, 62 %, dark: Pressure is a violet diamond, Temperature a magenta diamond, Weather a grey diamond. Signal now has a square (ISS-8D-08 holds).
- Observed: three families share one mark and differ only by hue, so the legend's mark does not separate them.
- Expected: each family separable by hue and by mark (P-4, colour never the only carrier).
- Suggestion: give Temperature or Pressure a second mark (for example a hexagon), as Signal was given a square.

## ISS-10D-12 A finished run states its name twice
- Severity: Low
- Kind: content
- Where: Run panel after Start on Dir Review Empty, dark, 1440x614, about 1 s after the run: the top status reads "Run w-35 of Dir Review Empty finished · Review", and the panel below reads "Run w-35 finished." with "Open the review".
- Observed: the same status appears twice in two wordings, and the name of the plant is in one and not the other.
- Expected: one status line that names the run once (R-5, "Run w-4 finished").
- Suggestion: make the toast "Run w-35 finished · Review" and leave the panel's line as the plan summary.

## Regression pass on round 8 (the 20 ISS-8D findings)

Checked against the build on 2026-10-15 at 1440x614. "Holds" means I saw the fix on the screen or in the DOM; "not holding" means the defect is still visible; "not re-checked" means the surface was not reached.

Holds:
- ISS-8D-04 loading wording: the review reads "Reading the run · 0 s", the panel "Reading the run...", the scrubber "Events appear once the run is read." (w-33).
- ISS-8D-07 area click: a click on the "Hall 1" frame title on the campus leaves no focus ring on cal-site (first instance).
- ISS-8D-08 legend: Signal has a square mark; Weather keeps the grey diamond and dashed wire (remaining mark overlap is ISS-10D-11).
- ISS-8D-10 review time on load: "--:--:-- of --:--:--" until the run is read (w-33 loading).
- ISS-8D-11 "Show all 8 runs" sits on the same inset as the rows (x 1079).
- ISS-8D-12 legend popover has inner padding (the heading overflow is ISS-10D-09).
- ISS-8D-13 Keyboard dialog: the groups are per screen and no empty heading shows (the review lists the plant-graph and review groups; see ISS-10D-10 for the missing step keys).
- ISS-8D-14 the punctuation list in the name rule stays on one line (Build panel, "letters, digits and . _ : / -, with").
- ISS-8D-15 the id hint reads "Lower case letters, digits and hyphens, starting with a letter."; the id error ends "starting with / a letter" on two lines, not one word.
- ISS-8D-16 Inspect shows one "OBSERVED · ch-u2 at 00:01:00" time (the caption is gone); the TRUE RESULTS block keeps the entity name.
- ISS-8D-01 wire ends, plant view: wires land on pin rows (packaging-line-b enters production-line-a on its "air in" row; water pumping wires land on pin rows). The loops between same-column nodes are the known exclusion. The review's wire ends were not measured this round.
- ISS-8D-09 run panel header: the plant name in the panel header is not cut for a short name ("Run Dir Review Empty"). The speed could not be observed, because the run finished before the first screenshot.
- ISS-8D-06 new plant save state: not re-checked (no plant was created; the Home row for dir-empty-check-8 reads "Not runnable yet: open Edit, save, and fix any issue it lists.").

Not holding:
- ISS-8D-02 wires cross area titles: still present on Production hall A in the hunt plant (ISS-10D-06).
- ISS-8D-03 campus kind lines cut: reproduced at its 30 % opening zoom on this window (ISS-10D-04).
- ISS-8D-18 editor empty areas first: not holding (ISS-10D-05).

Not re-checked:
- ISS-8D-05 the Halls group on the campus editor; ISS-8D-17 is still present (w-22: CP-1 about 230 px tall with two rows; won't-fix in round 8, still visible); ISS-8D-19 node title inset (not measured); ISS-8D-20 run footer cue (run panel not opened in a long plan).

Round-6 items still open: ISS-6D-35 (review at 50 %: the rightmost frame is cut at the inspector edge: "9 nodes" on w-33 and "1 nod" on w-13); ISS-6D-20 wires that pass under node bodies (food cold chain, office tower, data centre; the known exclusion); ISS-6D-23 (not re-checked: the Size preset was not opened this round).

## Score table (round 10, 1440x614 only; light and dark)

| # | Dimension | Score | Evidence |
|---|---|---|---|
| 1 | Layout | 5 | Good: no horizontal scroll (scrollWidth 1440); the Show-all link and the Navigator frame sit cleanly; the Run panel footer and Start stay in place when the duration error shows. Against: the editor outline still leads with four empty groups (ISS-10D-05); the campus opening zoom is set by width and leaves the plant small (ISS-10D-04); the review frames are cut at the inspector edge at 50 % (ISS-6D-35, open). |
| 2 | Typography | 4 | Overprinted labels on four plants at the opening zoom: "pump" over "level in", "storage tank" over "inflow in", "air receiver" over "air in" (ISS-10D-01); review readout labels touch their kind line (ISS-10D-03); campus kind lines cut at 30 % (ISS-10D-04). Good: ids in mono, dates kept whole ("15 Oct 2026, 04:13"), thousands separators, the name rule's list kept on one line. |
| 3 | Colour and contrast | 6 | Good: the legend's Signal square separates it from Weather (ISS-8D-08 holds); Pressure is violet and Water teal; error red is one token across the id, name and duration fields; status chips pair colour with a tick or an icon. Against: Pressure, Temperature and Weather share the diamond (ISS-10D-11). |
| 4 | Components | 7 | Good: the field error pattern (red border, mark, text) is shared by the New plant id and the run duration; link rows have labelled Remove buttons; the legend popover is padded; the Keyboard dialog is per screen; the Create and Start controls do not move. Against: the review's keyboard list gives no step keys (ISS-10D-10); the legend heading overflows its box (ISS-10D-09). |
| 5 | Content design | 5 | Good: "Reading the run" replaces the binary name while loading; the id hint is plain; "Show 15 more runs"; the review's disturbances and effector calls are sentences; "Not scored: no outside system is connected to this run." Against: the integrity block shows file names, a folder word and "streamsim verify" as its detail (ISS-10D-08); a finished run names itself twice (ISS-10D-12). |
| 6 | Feedback and flow | 5 | Good: the area click no longer moves the focus ring to the first instance (ISS-8D-07 holds); a refused duration shows the error and the footer names the field ("Duration needs attention before starting."); the run finish status with "Open the review"; the theme button shows its state. Against: the integrity block opens on a new run and is closed on older ones (ISS-10D-07); the step keys are not listed (ISS-10D-10). Running state and speed not observable (run too short). |
| 7 | Information design of graph and review | 3 | Lowest. The kind and port labels overprint on food cold chain, water pumping, compressed air and the hunt plant (ISS-10D-01); kind lines are missing on eight campus nodes and cut on three (ISS-10D-02); the review readouts collide with their kind line (ISS-10D-03); a wire crosses the Production hall A title (ISS-10D-06); the campus opens too small to read (ISS-10D-04). Good: plant wires land on pin rows and the same-column loops are the known exclusion; the timeline marks are centred under their labels (w-13: 5 min, 12 min, 18 min marks and chips align); the fault and effector legend shows on the timeline; selection dims the other wires. |
| 8 | Motion and responsiveness | 7 | Verified: the id error appears without moving Create (y 405 before and after); the duration error leaves the Mode choice in place (y 314); the review's time placeholder keeps the cursor still until the run is read; the run's finished status arrives in about 1 s. Not measured: the running state's progress. |
| 9 | Craft | 5 | Good: the Legend popover is padded (ISS-8D-12 holds); the inspector title of a long plant name wraps onto three lines (hunt plant). Against: the legend heading overflows its box (ISS-10D-09); a wire crosses an area-frame title (ISS-10D-06). Not checked this round: favicon, tab titles, node title inset. |
| 10 | Whole-product coherence | 5 | Good: Open, Edit, Run and Review keep their verbs on Home, plant and review; "Run again" became "Set up another run" consistently; the New plant header button shows pressed while its form is open. Against: the editor outline's empty groups contradict the plant view's area list (ISS-10D-05, and round-8 ISS-8D-05 not re-checked); the report's "streamsim verify" and file words contradict the product's plain words (ISS-10D-08). |

**Overall: 3** (the lowest dimension is 7, Information design of graph and review, at 3; not rounded up). Dimensions 1, 5, 6, 9 and 10 are 5; dimension 2 is 4; dimensions 3, 4 and 8 are 6, 7 and 7.

Finding counts: High 1 (ISS-10D-01); Medium 7 (ISS-10D-02 to ISS-10D-08); Low 4 (ISS-10D-09 to ISS-10D-12); total 12.

## Compared with round 8

Round 8 overall 4 (lowest: Information design of graph and review, at 4). Round 10 overall 3 (lowest: the same dimension, now at 3). The dimension scores moved as follows: Layout 6 to 5; Typography 5 to 4; Colour and contrast 5 to 6; Components 6 to 7; Content design 5 to 5; Feedback and flow 5 to 5; Information design 4 to 3; Motion 6 to 7; Craft 5 to 5; Whole-product coherence 5 to 5.

What improved and was verified:
- The loading text, the area-click focus, the legend's Signal mark and padding, the Show-all inset, the id hint and error, the name-rule punctuation, the Inspect caption and the review placeholder time all hold (see the regression pass).
- The run panel's error treatment and Create/Mode stability hold (ISS-6D-05, 6D-06, 6D-28, 6D-29 remain fixed).
- Plant wires land on pin rows in the plant view (ISS-8D-01 holds for the plant view).

What did not improve or regressed:
- Label overprint is the largest new problem: kind lines print on the first port row on four plants and on most review readouts. Round 8 did not measure these plants at this window, so it is recorded as found now, not as a regression of one round-8 item.
- ISS-8D-02 (titles crossed by wires) and ISS-8D-18 (empty groups first) are recorded as not holding.
- ISS-8D-03 (campus kinds) is reproduced at 30 % on this window; round 8 had recorded 52 % on its window.
- The integrity block is the new review-level problem: it opens by default on a new run and shows file names and the binary's name as its detail (ISS-10D-07, 10D-08).
- Still open from round 6 and 8: ISS-6D-35 (review clipped at the inspector edge at 50 %), ISS-8D-17 (tall readout node, won't-fix), wires under node bodies and same-column loops (known exclusions).

UI-EXPECTATIONS status changes seen in this round: P-2 not met at the opening zoom on several plants (ISS-10D-01, ISS-10D-02, ISS-10D-04); P-14 not met (ISS-10D-06); E-14 partly met (plant wires end on pins; same-column loops excluded); V-4 and V-14 partly met (ISS-10D-03); V-6 not met (ISS-10D-07); G-15 partly met (ISS-10D-10); G-16 not re-checked; E-15 met (the editor opens fitted: compressed-air at 62 %); R-5 partly met (the finished status shows, but the review waits for "Open the review" and does not open by itself); R-6 and R-10 met (duration error next to its field, Start and Mode do not move); W-5 partly met (the id error shows once, under the field; creation was not tested).
