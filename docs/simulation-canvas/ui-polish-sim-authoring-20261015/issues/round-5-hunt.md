# Round 5 hunt: interface and simulator behaviour (hunter)

Date: 2026-10-15. App: http://127.0.0.1:8766 (Chrome, own tab 770933653, 1440x900 viewport).
Bar: docs/simulation-canvas/UI-EXPECTATIONS.md. Hunt plan: brief-hunter.md, sections 1-10.
Focus: (a) regression of round-4 fixes, (b) new areas (Runs filter, compare runs, Report wording, Inspect at different times,
keyboard help, themes, widths, long and special names, history navigation, deep links), (c) section 10 probes not covered in round 4.
Reading flows: compressed-air-works and runs w-1, w-3. Edits: plants named hunt-* only (hunt-empty).
Not re-reported: round-1..4 issues with Status fixed (regressions of them would be noted as such).

Findings are appended below as verified, one `## ISS-5H-nn` each.

## ISS-5H-01 The name suggestion for a new pump copies an existing pump's long name, not the next free id
- Severity: Medium
- Kind: unclear
- Where: Plant editor `hunt-empty` > Build panel > Add equipment > "Centrifugal pump" (library list), Name field placeholder.
- Steps:
  1. Editor `hunt-empty` with a pump already renamed to `pump-house-secondary-circuit-feed-north-long-name-07` (from round 4 probe).
  2. Build panel: click "Centrifugal pump" in the library list. Read the Name placeholder (it is the suggestion).
  3. Click "Cooling tower" in the list, then "Centrifugal pump" again by a ref click. Read the placeholder with a DOM read.
- Observed: Placeholder reads `pump-house-secondary-circuit-feed-north-long-name-2`. Cooling tower suggests `CT-1` (correct: a short id). The pump suggestion carries the other pump's long name with a `-2` suffix, so accepting the suggestion adds a long, unrelated id.
- Expected: The suggestion is the next free short id for the kind (for a pump, `CP-2`, as the pump list is `CP-n`), as UI-EXPECTATIONS E-3 says ("suggests the next free id").
- Suggestion: Derive the suggestion from the kind's id prefix and the next free number in the plant, never from another instance's name.
- Status: fixed aef37b0
- Fix: suggestedName uses a sibling prefix only when it is a short id (1-4 letters); otherwise the kind's initials, so a new pump is CP-n.

## ISS-5H-02 A long plant name pushes the top bar wider than the window: Keyboard and Theme are cut off, the page scrolls sideways
- Severity: Medium
- Kind: not-working
- Where: Top bar of the editor for plant `hunt-special-names` (name `Hunt & "Ålborg" pumps / 冷却 test: 01 ünïcode ✓ with a very long name that keeps going to test wrapping in the header`), at 1440 px. Breadcrumb "Wo…" > plant title; right side "Keyboard" and "Theme".
- Steps:
  1. Home > New plant. Name (as above), Id `hunt-special-names`, Start from "Compressed-air works", "A full copy". Create and open the editor.
  2. Read the top bar. Read the page width with a DOM read (document.documentElement.scrollWidth).
  3. Reload the editor URL (`#/edit?composition=compositions%2Fhunt-special-names.composition.json`) and zoom the top bar.
- Observed: The top bar and page are 1517 px wide in a 1440 px window (scrollWidth 1517 against innerWidth 1440). The "Theme: system" button and the "Keyboard" button are cut off at the right edge, and the page scrolls sideways. The plant title is shown in full and bold, while the "Workspace" crumb is cut to "Wo…" first, so the crumb that tells the user where they are is the one that disappears. The same happens after a reload.
- Expected: The plant title truncates with an ellipsis inside a fixed budget, the top bar never widens the page (G-10: no horizontal scroll at 1280, nothing important clipped), and the Workspace crumb keeps at least its first word (G-1 breadcrumb).
- Suggestion: Cap the title's width in the top bar (flex: 1, min-width: 0, ellipsis) and give the crumb and the right-hand controls a fixed minimum; show the full name in the title tooltip.
- Status: fixed 1dbb9c9
- Fix: the shell grid column is minmax(0, 1fr) and the Workspace crumb is not shrunk; a long plant name is cut with ellipsis and the page does not scroll sideways (e2e at 1440 px).

## ISS-5H-03 A new plant made by copy is "Not runnable yet: open Edit, save…", but Save is disabled and nothing is listed to fix
- Severity: High
- Kind: not-working
- Where: Home row for `hunt-special-names` ("Not runnable yet: open Edit, save, and fix any issue it lists."), then its editor: status "No changes", Save "Nothing to save" (disabled), no issue list or Show issues button in the top bar.
- Steps:
  1. Home > New plant. Name "Hunt & "Ålborg" pumps …" (any name), Id `hunt-special-names`, Start from "Compressed-air works", "A full copy of its equipment…". Create and open the editor. Do not change anything. Go to Workspace.
  2. Read the Home row for hunt-special-names: "Not runnable yet: open Edit, save, and fix any issue it lists." Only an Edit button; no Run.
  3. Click Edit. Read the status ("No changes") and the Save button (disabled, title "Nothing to save"). Look for the issue list.
- Observed: The message tells the user to save, but Save is disabled because nothing changed, and the editor shows no issue to fix. The plant cannot become runnable from here until the user invents an edit: after adding one Cooling tower (CT-1) and clicking Save, the status read "Ready to run. Saved." and Run appeared. A user who follows the Home instruction is stuck; the plant shows "Not runnable" with no reason.
- Expected: A new full copy is runnable at once (as Compressed-air works is), or the editor says why it is not runnable and offers Save (E-2 "Saved but not composed (with the issue count)"), so Save is enabled exactly when the message asks for it.
- Suggestion: Compose the copy when it is created; if a plant is not composed, make the status "Saved but not composed: N issues" and enable Save or the issue list.
- Status: fixed 6bc57a5
- Fix: a full copy is composed when created, so it is runnable at once; an empty plant is still created without a domain.

## ISS-5H-04 The Run panel still says "a consumer" in the Evaluate explanation (the word fixed in the Report is left here)
- Severity: Low
- Kind: unclear
- Where: Run panel (plant screen > Run plant, or "Set up another run"), MODE, the Evaluate option description. Same word in the Report on Evaluate runs is fixed (ISS-4H-07); here it is not.
- Steps:
  1. Plant `hunt-special-names` > Run plant (any plan).
  2. Read the MODE section: Evaluate option.
  3. Read it again through "Set up another run" from review w-14.
- Observed: "The truth stays sealed while it runs, so a consumer can be scored; it is revealed when the run ends." A stranger does not know what a consumer is (the same word the round-4 fix removed from the Report).
- Expected: Plain words, one name for the outside system across the product (G-6): e.g. "…so an outside system can be scored; it is revealed when the run ends." The same wording as the Report's "Not scored: no outside system is connected to this run."
- Suggestion: Replace "a consumer" with "an outside system" here and in any Evaluate help text.
- Status: fixed e342a6f
- Fix: the Evaluate explanation in the Run panel and the run status says "outside system" and "true results ... hidden", not "consumer" or "sealed".

## ISS-5H-05 Report wording still uses sealed-truth and alarm-engineering jargon: "TRUE KPIS", "sealed", "revealed", "ALARM SYSTEM ISA-18.2", "flood intervals", "chattering", "standing at end"
- Severity: Medium
- Kind: unclear
- Where: Review, Report tab, sections "TRUE KPIS" and "ALARM SYSTEM", and the first line for Explore/Evaluate runs. Checked on w-1 (Explore), w-3 (Evaluate), w-14 (Explore), and the Inspect tab of a run ("No true KPIs for receiver-header-1.").
- Steps:
  1. Open http://127.0.0.1:8766/#/run?world=w-3 (Evaluate). Read the Report tab (get_page_text).
  2. Open w-1 and read the Report tab. Then select receiver-header-1 and read Inspect.
- Observed: Words a stranger does not know: "TRUE KPIS" (heading), "the simulated truth below stayed sealed while it ran and was revealed when it ended", "ALARM SYSTEM" with "ISA-18.2", and the rows "activations", "peak per 10 min", "flood intervals", "chattering", "stale", "standing at end". Inspect says "No true KPIs for receiver-header-1."
- Expected: Plain sentences with the alarm measures named in words and a one-line explanation each (G-6, V-12: real column headers and words, plain row names). For example: "Alarms: 0 raised; 0 repeated; standing at end of run: 0." The standard name sits in a details line.
- Suggestion: Rename "TRUE KPIS" to "Results of the simulation" (or "Results"), keep "sealed/revealed" as "Hidden while running, shown at the end"; put "ISA-18.2" and the measure definitions behind a disclosure.
- Status: fixed 7c3f7c1
- Fix: "True KPIs" is "True results" with a plain explanation; the alarm table is "Alarms" with a plain label and meaning per measure; ISA-18.2 stays as the section note.

## ISS-5H-06 Revert looks enabled when there are no changes (Save looks disabled), though Revert is aria-disabled and does nothing
- Severity: Low
- Kind: accessibility
- Where: Editor top bar, "Revert" and "Save", state "No changes" (plant `hunt-empty` after Undo to the saved state; also `hunt-special-names` before any edit).
- Steps:
  1. Open the editor of hunt-empty with no changes. Zoom the top bar (x 620–1066).
  2. Read the DOM: buttons "Revert" and "Save".
  3. Click Revert.
- Observed: "Save" is greyed (title "Nothing to save"). "Revert" has dark text and a full border, so it reads as an active button, but its aria-disabled is "true". The click does nothing and shows no message. The two controls look different in state from each other while both are disabled (UI-EXPECTATIONS E-11 asks for both to be visibly disabled).
- Expected: Revert is greyed like Save when there are no changes; a disabled control is not visually active (G-9: disabled controls state contrast; E-11).
- Suggestion: Use the same disabled style for Revert and Save from the same state flag.
- Status: won't-fix: not reproduced on this build; in a clean editor Revert and Save compute the same disabled colours and borders, and Revert is aria-disabled
- Fix: none.

## ISS-5H-07 The Runs filter ("All plants") is wider than its card and is cut off at the window edge when a long plant name is listed
- Severity: Medium
- Kind: not-working
- Where: Home (Workspace), Runs card header, the "All plants" select, after a plant with a long name ("Hunt & "Ålborg" pumps / 冷却 test: 01 ünïcode ✓ …") has runs.
- Steps:
  1. Open http://127.0.0.1:8766/ at 1440 px with runs of that plant listed.
  2. Read the position of the "All plants" select and the Runs card edge (DOM: select right edge vs card right edge).
- Observed: The Runs card ends at x 1407, but the select's right edge is at x 1577. The select runs past the card border and the window (page scrollWidth stays at 1440 but the control's right side is not visible). Runs rows show the long plant name truncated, so the header grows with the row content.
- Expected: The filter keeps its place inside the card header at every width; long names truncate in rows and never widen the header (G-10; W-4 filter by plant).
- Suggestion: Give the Runs header a fixed width with the select at a fixed size; let the row title truncate with min-width: 0 in its flex row.
- Status: fixed 203b581
- Fix: the Runs plant filter has a 12rem basis, min-width 0 and an ellipsis, so it stays inside its card (probe: right edge 1395, card 1408, before the fix 1487).

## ISS-5H-08 Playback opens at 00:01:00 with no explanation; the handle is one minute in, not at the start
- Severity: Low
- Kind: unclear
- Where: Review, playback bar, on runs w-1 (1 h), w-3 (30 min) and w-19 (30 min with a fault). Readout "00:01:00 of 01:00:00" next to the handle, at the first tick.
- Steps:
  1. Open http://127.0.0.1:8766/#/run?world=w-1 (also w-3 and w-19). Wait 3 s.
  2. Zoom the playback bar. Read the readout.
- Observed: The readout starts at 00:01:00 and the handle sits one minute in, not at 00:00:00. The only explanation is the note "11 224 observations · opened once every channel reported" below the graph. The first tick (00:00:00) is never the playhead. A stranger cannot tell whether the run began at 00:01:00.
- Expected: Playback opens at 00:00:00 (or the first event), as V-2 says "starts at the beginning, or at the first event"; if the first complete reading is at 00:01:00, the readout says so next to the time ("first full reading at 00:01:00").
- Suggestion: Start at 00:00:00 with the first sample held, or state the first full reading next to the readout.
- Status: fixed 47e49d3
- Fix: the playback readout says when it opened and whether every channel had reported by then ("opened at 00:01:00, once every channel had reported", or "before" when the 10% cap applies).

## ISS-5H-09 Run rows and the plant's recent runs do not show the seed, so two runs of the same plan look identical
- Severity: Low
- Kind: improvement
- Where: Home Runs list (rows w-14, w-15, w-17 of `hunt-special-names`, all "Explore · 30 min · no faults"), and the plant overview "Recent runs" list.
- Steps:
  1. Run `hunt-special-names` with seed 1 (w-14) and with seed 2 (w-15): "Set up another run", change Seed, Start the run.
  2. Read the Home Runs rows and the plant's "Recent runs".
- Observed: The rows read "w-14 · 15 Oct 2026, 02:35 · 30 min · no faults" and "w-15 · 15 Oct 2026, 02:35 · 30 min · no faults". The seed appears only inside the review header ("seed 1 · 30 min"). Comparing two runs of one plan (the reason to "Run again with a changed seed") needs opening each review.
- Expected: Each run row shows the seed next to the mode (W-3 lists what a run shows; the seed is what tells runs apart). A compare view is not offered; the rows are enough for a first step.
- Suggestion: Add "seed 2" after the mode in the run row and in Recent runs.
- Status: fixed fa99419
- Fix: Home Runs rows and the plant's Recent runs show "seed N" after the duration (the seed comes from the run listing, 56a910d; null shows nothing).

## ISS-5H-10 Review effector table: the Effector name is cut to "set lo…" and the Outcome runs past the window edge (partly a regression of ISS-4H-12)
- Severity: Medium
- Kind: regression
- Where: Review run w-9 (Compressed-air works, plant on 1440 px, the review opened from the plant's Recent runs), Report tab, "EFFECTOR CALLS" table, row "00:05:00 · compressor-2 · set loaded · applied · 335 ms".
- Steps:
  1. Open http://127.0.0.1:8766/#/run?world=w-9 at 1440 px.
  2. Zoom the Report panel's EFFECTOR CALLS table.
  3. Read the table cells with a DOM read (client width and right edge of each cell).
- Observed: The Effector cell is 68 px wide and its text "set loaded" (87 px) is cut to "set lo…". The Outcome cell's right edge is at x 1469, past the 1440 px window, so "applied · 335 ms" is cut at "applied · 335 m…" on screen. ISS-4H-12 fixed the cut Outcome word ("applie…") but the table is now wider than the panel, so the outcome is still clipped at 1440. The row with "fails to load" (no outcome) is fine.
- Expected: Both cells show their full text at 1440 px (V-12: real column headers and words; G-10 nothing important clipped). The table fits the panel: the outcome wraps or the effector column takes the free width.
- Suggestion: Let the table use the panel width (table-layout: fixed with a percentage for each column) and wrap the outcome onto two lines before it overflows.
- Status: fixed 2ed24e5
- Fix: the effector name wraps in its own column (new "effector" column class) instead of an ellipsis; a long name wraps and is not clipped at 1440 px (probe).

## ISS-5H-11 Inspector "None" under LINKS IN is flush to the left edge, not inset like the rows around it
- Severity: Low
- Kind: unclear
- Where: Plant screen, Inspector for instance compressor-1 (Compressed-air works), section "LINKS IN" (count 0), text "None".
- Steps:
  1. Open http://127.0.0.1:8766/#/plant?domain=domains%2Fcompressed-air-works.domain.json.
  2. Click compressor-1 in the graph (or the outline row).
  3. Zoom the Inspector between "alarm state" and "LINKS OUT".
- Observed: The heading "LINKS IN" and its count "0" are inset about 20 px; the word "None" under it starts at the panel's left edge (x about 1066 at 1440 px). Other section rows (LINKS OUT items, FAULTS) are inset.
- Expected: The empty text sits in the same inset as the section rows (W-10: controls and text inside cards are inset like the rows around them).
- Suggestion: Give the empty-state line the same padding as the section body.
- Status: fixed 4cd9478
- Fix: the empty list text ("None") uses the same inset as the list rows (new .empty class in the Inspector).


## Simulator behaviour notes

Probes run on hunt-special-names (runs w-14, w-15, w-17, w-19) and on compressed-air-works (w-1, w-3, w-9 read only). Verdict: supported | partly | missing.

1. Comms gap (sensor dropout): "basin temp · dropout" on CT-1 from 00:05 (w-19, seed 1, Explore). The basin temp reads 21.9 °C at 00:05, 00:12 and 00:16:54 (held, not blank); the node and Inspect show it as an ordinary value next to a red "Fault" tag. No gap, no age indicator, no quality flag. ALARM SYSTEM: activations 0, stale 0. A real comms gap leaves no new reading and, after a timeout, a stale or bad-quality flag. Verdict: partly (the fault is applied, but the held value looks like a live reading and nothing alarms).
2. Duplicated or late readings: no fault kind offers delay, duplicate, out-of-order or timestamp jitter. The options per instance are only: offset, drift, ct ratio, stale, dropout, noise burst, clamp open, harmonics (feeder meter) and the equipment-specific faults. Verdict: missing.
3. Feeder trip (fault on the feeder meter or transformer): the feeder meter's fault list is sustained overload, capacitor bank failure, load surge and meter errors. No trip or loss of supply exists, and the plant has no supply-side model to show what depends on it. Verdict: missing (load surge exists as a load step; it was not probed in detail in round 4 either).
4. What depends on a dropped sensor: CT-1 has no links in this plant, so nothing downstream uses its basin temp; the fan power keeps its normal values (4.60 kW at 00:05, 4.50 kW at 00:16:54). The Report does not say that the dropped value feeds nothing. Verdict: missing (a dependent consumer cannot be tested here; the Report does not say what is affected).
5. Alarm that never clears / alarm at all: no run in this round produced an alarm (w-1, w-3, w-14, w-15, w-17, w-19: activations 0 and standing at end 0), including a 12-minute frozen sensor. Verdict: missing (no alarm produced; see round 4 note 6).
6. Repeatability: the same plan with seed 1 gives the same values (w-14 and w-17: CT-1 basin 21.8 °C and fan 4.60 kW; production-line-a 179 m³/h and 165 kW; packaging-line-b 514 m³/h and 84.4 kW). A new seed changes them (w-15 seed 2: packaging-line-b 506 m³/h, compressors 694 and 697 m³/h). Verdict: supported.
7. Stuck sensor, drift, command refused or not executed, device failed and stays failed: not re-probed this round; round 4 notes 1–5 stand.
8. Implausible values: at 00:16:54 of w-19, production-line-a air flow reads −0.30 m³/h with a small marker beside it; the Report does not name the marker. Round 4 V-8 gap stands. Verdict: partly (a marker exists; its meaning is not stated in words).

## Round 4 regression check

- Leave-confirm with unsaved changes (ISS-4H-05): holds.
- Undo and Redo stability (ISS-4H-04): holds.
- Save status text (ISS-4H-02): holds ("Ready to run. Saved.").
- Long names in nodes (ISS-4H-01, ISS-4H-03): the node header keeps the kind visible. The top bar is new breakage: ISS-5H-02.
- Fresh run plan after failure (ISS-4H-15): holds.
- Stuck run and plant screens (ISS-4H-13, ISS-4H-14): hold.
- Scrubber fault mark (ISS-4H-08): holds.
- Scoring wording (ISS-4H-07): holds in the Report; the same word "consumer" remains in the Run panel: ISS-5H-04.
- Effector outcome wording (ISS-4H-12): still partly failing: the Outcome cell runs past the window edge at 1440: ISS-5H-10. The run-panel "accepted" and review "applied" split was not re-checked.
- Not re-checked this round: ISS-4H-06 (id check on blur), ISS-4H-09 (prefilled stale fault), ISS-4H-11 (needs binary).
- Widths 1280 and 1920: not tested. resize_window returned success, but the Chrome viewport stayed 1440 x 757.

## Cleanup and side effects

- Plants I created: `hunt-special-names` (saved, 9 instances, runs w-14, w-15, w-17, w-19). Its composition was not runnable at creation (ISS-5H-03); after my CT-1 was added and saved it is runnable.
- Plant `hunt-empty` (mine from round 4): I added and removed a pump (undo) and added CT-1 and saved; 3 instances now; run w-5 only.
- Runs created on hunt-special-names only. No author plant was edited or run by me this round. The author's tab (770933647) and the workspace were not touched by me, though the workspace showed the author's new runs (w-10 to w-13, w-16, w-18) while I worked.
- Browser: theme restored to System; my tab 770933653 closed at the end.

## Worked well

- Run again with the same seed reproduces the same values (w-14 and w-17), so the "same plan, same run" promise holds.
- Keyboard list: dims the page, real keys, Escape closes it and focus returns to the Keyboard button.
- Leaving an editor with unsaved changes asks first: "Save and leave", "Discard and leave", "Stay". Verified on hunt-empty; Stay keeps the draft.
- Save: "Ready to run. Saved." after a save; Undo and Redo keep their place.
- Runs filter by plant works (hunt-empty shows its one run), and a plant without runs is not listed in the filter.
- Scrubber: the fault label "00:15:00 receiver-header-1" seeks to 00:15:00 and selects the instance; the Inspect tab then shows "Fault: header leak" with the observed value.
- Evaluate review (w-3): "Evaluate run: the simulated truth below stayed sealed while it ran and was revealed when it ended." matches the run panel.
- Invalid instance name is refused in words: "'CP 1 & pump/A' is not a valid name. An instance name: a letter or digit, then letters, digits and . _ : / -, with no spaces."
- Unknown run and unknown plant now have a designed state with "Back to the workspace" (ISS-4H-13 and ISS-4H-14 hold).
- Back and forward work: Home > Plant > Review > Back returns to the plant with its zoom; Forward returns to the review.
- Run plant after an earlier failure opens a fresh plan (ISS-4H-15 holds): seed 1, 30 min, no faults, Start the run.
- Dark theme review (Compressed-air works, w-1): the labelled theme control reads "Theme: Dark", and node text and wires stay readable.

Counts by severity: High 1, Medium 5, Low 5 (11 verified findings, ISS-5H-01 to ISS-5H-11). Not reached: widths 1280 and 1920 (viewport did not change), dark theme on every screen (review only), keyboard tab order on the Run panel, the run-panel effector wording check.

## Fixer follow-up: ISS-4H-11 (abandoned run)

- Status: fixed fa99419 (Home row and e2e added after the listing carried complete and incomplete_reason, 56a910d). Earlier: 4a5e88c for the review header.
- Fix: the review header reads "Did not finish" with the report's incomplete reason beside it (the Report keeps "Incomplete: ..."). The Home run row now shows it from the run listing (complete and incomplete_reason, 56a910d; canvas fa99419).
