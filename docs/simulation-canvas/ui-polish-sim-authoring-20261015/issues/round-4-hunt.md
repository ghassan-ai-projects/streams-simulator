# Round 4 hunt: interface and simulator behaviour (hunter)

Date: 2026-10-15. App: http://127.0.0.1:8766 (Chrome, own tab, 1440 wide).
Bar: docs/simulation-canvas/UI-EXPECTATIONS.md. Hunt plan: brief-hunter.md, sections 1-10.
Reading flows: compressed-air-works and runs w-1..w-4. Edits: plants named hunt-*.
Not re-reported: issues round-1..3 with Status fixed or needs-binary.

Findings are appended below as verified, one `## ISS-4H-nn` each.

## ISS-4H-01 Plant-graph nodes hide the kind word when the id is long (packaging line, production line)
- Status: fixed 5df7bdb
- Fix: The node header is two rows: the id, then the kind in words on its own line; the kind is no longer hidden when a status tag shows.
- Severity: Medium
- Kind: unclear
- Where: Plant screen "Compressed-air works" at the opening zoom (61 %), nodes `production-line-a` (Production hall A) and `packaging-line-b` (Production hall B). Compare `compressor-1` ("air c…") and `air-dryer-1` ("air dry…").
- Steps:
  1. Home > Open "Compressed-air works".
  2. Look at the node headers in Production hall A and Production hall B without zooming.
  3. Zoom the screenshot on each node header.
- Observed: The header shows only the id. The kind "production line" and "packaging line" is not visible at all (the header is clipped, and packaging-line-b shows a stray bar). Other nodes show the kind cut to "air c…", "air dry…", "f…" (power-meter-air, feeder meter). The DOM text contains the full kind, so the kind is clipped by the node, not missing.
- Expected: At the opening zoom every node shows its id and kind in words (P-2), and the kind is not cut to a fragment; a user can tell a production line from a packaging line without zooming.
- Suggestion: Put the kind on its own line under the id, or let the header wrap, instead of clipping it on the same line.

## ISS-4H-02 Save confirmation is cut off: "Ready to run. Composed 4 channels, 24 faults, 3 effectors. Saved." shows only "Ready to run. Composed 4 chann..."
- Status: fixed 264ddc8
- Fix: The save state reads "Ready to run. Saved." and the composition counts sit in its tooltip; the status wraps instead of clipping.
- Severity: Medium
- Kind: not-working
- Where: Editor top bar, status text left of "Open plant" (plant `hunt-empty`, editor at 1440 px).
- Steps:
  1. Home > New plant: Name "hunt-empty", Start from "Chiller pilot", "An empty plant", Create and open the editor.
  2. Build > Add to the plant (Water-cooled chiller, name WCC-1).
  3. Click Save.
  4. Read the status at the top of the graph area.
- Observed: The status shows "Ready to run. Composed 4 chann…". The element is 198 px wide with text-overflow ellipsis, and its full text is 375 px. The word "Saved." and the counts after "channels" are hidden. The "Saved" part is the only confirmation that the save reached the file (E-2).
- Expected: The save state is visible and unambiguous (E-2) and nothing important clips at 1440 px (G-10). Either the status wraps, or the "Saved" state moves to its own element that cannot be cut.
- Suggestion: Give the save state its own short label ("Saved") and move the composition counts to the tooltip or a details line.

## ISS-4H-03 A long instance name overflows its node header and hides the kind word
- Status: fixed 88bbbed
- Fix: The node id truncates with an ellipsis and shows the full name on hover; the idle Delete label truncates with the full name on hover; the confirmation question still wraps.
- Severity: Low
- Kind: not-working
- Where: Plant editor, graph node header of a pump renamed in the inspector (Name field, then Rename). Also the inspector Delete button.
- Steps:
  1. Editor `hunt-empty`: select CP-1 (Centrifugal pump).
  2. In Name type `pump-house-secondary-circuit-feed-north-long-name-07`, click Rename.
  3. Look at the node in the Pilot plant area.
- Observed: The id runs past the right edge of the node (the header box ends near x 780, the text ends near x 933 at 100 %). The kind "pump" disappears from the header. The inspector Delete button wraps its label onto two lines ("Delete pump-house-secondary-circuit-feed-" / "north-long-name-07") and the first line sits on the icon.
- Expected: A long id is truncated or wraps inside the node frame, and the kind word stays visible (P-2, P-14 no overlap). The Delete button keeps one line or a short label ("Delete CP-1" plus the name in a tooltip).
- Suggestion: Truncate the id with an ellipsis inside the header and keep the kind on a separate line; use a short label on the Delete button.

## ISS-4H-04 Undo and Redo buttons move by about 230 px whenever the save status text changes width
- Status: fixed 264ddc8 (same commit as ISS-4H-02)
- Fix: Open plant, Show issues and Reload moved left of the status, so Undo, Redo, Revert and Save keep their place.
- Severity: Low
- Kind: unclear
- Where: Editor toolbar, Undo (Cmd+Z) and Redo (Cmd+Shift+Z) next to the save status.
- Steps:
  1. Editor `hunt-empty`, add a chiller (Build > Add to the plant). The status reads "Unsaved changes" and Undo is at about x 605, Redo at about x 636.
  2. Click Undo until the status reads "No changes". Undo and Redo are now at about x 835 and 865.
  3. Click Save. Status reads "Ready to run. Composed 4 channels…" and Undo is back at about x 605.
- Observed: The two history buttons jump between x 605 and x 835 as the status label changes. A click aimed at the old position after a save or undo lands on the status label or on nothing. (I missed Redo twice for this reason before the button ref was used.)
- Expected: Undo and Redo keep one fixed position in the toolbar whatever the status says (G-2 consistent layout; keyboard and mouse equal, principle 6).
- Suggestion: Anchor Undo and Redo to a fixed slot (left of the status or right of Save) and let only the status label flex.

## ISS-4H-05 Leaving the editor with unsaved changes does not ask; the Home card still shows the old saved state
- Status: fixed d8a15b3
- Fix: Leaving with unsaved changes (breadcrumb, Home, Open plant, Edit on another plant, or Back) asks: Save and leave, Discard and leave, or Stay; Escape stays. The beforeunload warning on reload is unchanged.
- Severity: Medium
- Kind: not-working
- Where: Editor `hunt-empty` top bar, "Workspace" breadcrumb (and Home after it). Status "Unsaved changes".
- Steps:
  1. Editor `hunt-empty` (saved with WCC-1). Add Centrifugal pump CP-1, Rename it to `pump-house-secondary-circuit-feed-north-long-name-07`. Status reads "Unsaved changes", Save is enabled.
  2. Click the "Workspace" breadcrumb.
  3. Read the Home card for hunt-empty.
  4. Click Edit on hunt-empty again.
- Observed: No prompt at step 2; Home opens at once. The Home card for hunt-empty still reads "1 instance · no runs yet" (the saved state). Edit reopens the editor with the unsaved pump and "Unsaved changes" still shown. A reload attempt (location.reload) showed no prompt either; I could not confirm that the reload happened.
- Expected: Leaving with unsaved changes asks first, with Save, Discard, and Stay (E-9; G-11). A user who closes the tab must be warned.
- Suggestion: Ask on breadcrumb, Home and Open plant clicks, and on beforeunload when the draft is dirty; keep the draft in memory while it is shown as unsaved.

## ISS-4H-06 New plant id errors are shown only after Create, away from the field, and stay after the id is corrected
- Status: fixed 3e7ffc8
- Fix: The Id is checked when its field is left, with the binary's own id rule (component.new, no file written); the message sits under the field and clears when the value changes.
- Severity: Medium
- Kind: unclear
- Where: Home > New plant form, Id field and the message line above "Create and open the editor".
- Steps:
  1. Home > New plant. Name `hunt-bad-id`, Id `Bad Id!`. Nothing is shown; "Create and open the editor" is enabled.
  2. Click Create and open the editor. The message `"Bad Id!" is not a plant id: 3 to 64 lower-case letters, digits and hyphens, starting with a letter` appears about 190 px below the Id field, after an empty gap.
  3. Select the Id field, type `hunt-empty`. The message still reads `"Bad Id!" is not a plant id…`.
  4. Click Create and open the editor: `A plant with the id hunt-empty already exists.` appears in the same place.
- Observed: The id is checked only on Create, so the user gets no feedback while typing. The first message stays on screen after the value is changed and is valid, so it reads as if the new id were still invalid until the next Create. The message sits in a gap below the radio buttons, far from the field. The duplicate check works (nothing was overwritten).
- Expected: The Id is checked as the binary would while the user types (W-5), its message sits under the Id field, and the message clears when the value changes.
- Suggestion: Validate on blur or on input with the binary's rule, show it under Id, and clear it on the next edit.

## ISS-4H-07 The review's Scoring section uses internal words: "consumer" and "operator endpoint"
- Status: fixed 0a3e31f
- Fix: Scoring reads "Not scored: no outside system is connected to this run."; the developer wording is in a "For developers" disclosure.
- Severity: Low
- Kind: unclear
- Where: Review (run w-5, plant `hunt-empty`), Report tab, last section "SCORING".
- Steps:
  1. Home > Run hunt-empty > Start the run (Explore, no faults). The review opens.
  2. Report tab, scroll to the bottom: "SCORING".
- Observed: "Nothing was scored: no consumer verdict was recorded for this run. A consumer attaches through the operator endpoint; the canvas does not attach one." A stranger does not know what a consumer or an operator endpoint is. The sentence explains a developer setup, not what the user can do.
- Expected: Plain words for the state and the next step, for example "Not scored. Scoring needs an outside system connected to this run; this canvas does not connect one." (G-6, V-13; the same thing carries one name.)
- Suggestion: Keep the developer detail in a details disclosure; headline "Not scored: no outside system is connected to this run".

## ISS-4H-08 Clicking a fault dot on the scrubber seeks to 00:09:59 and selects nothing; clicking its label selects the instance
- Status: fixed c0b98ac
- Fix: Each fault mark on the track is a control that selects its event at its own time and its instance, the same as its label. Checked in Chrome: the 00:10:00 dot lands on 00:10:00 and selects compressor-1.
- Severity: Medium
- Kind: not-working
- Where: Review run w-1 (plant Compressed-air works), scrubber track under the playback buttons, marks "00:10:00 compressor-1" and "00:20:00 power-meter-air".
- Steps:
  1. Open http://127.0.0.1:8766/#/run?world=w-1 (direct link).
  2. Click the red dot on the track at the 00:10:00 tick (about x 406, y 190).
  3. Read the time and look at the graph and the right panel.
  4. Click the label chip "00:10:00 compressor-1" under the track (about x 402, y 213).
- Observed: Step 2 moves the playhead to 00:09:59 (one second before the fault), selects nothing, and the right panel stays on Report. Step 4 moves to 00:10:00, selects compressor-1 (it gets the selection ring and the outline row is highlighted), and switches the panel to "Inspect · compressor-1" at 00:10:00. The two marks of one event behave differently.
- Expected: Clicking a mark seeks to the event and selects its instance (V-3), the same from the dot or the label, landing on the event time itself.
- Suggestion: Make the whole mark (dot and label) one target that seeks to the event time and selects its instance.

## ISS-4H-09 "Set up another run" keeps the first fault's id but shows it blank, and Start sends it anyway, so the run fails and a failed run is saved
- Status: fixed 78405ea
- Fix: A prefilled fault the plant does not have is cleared when the plan is built and named in the "Not copied" note; Start is blocked with "Fault N needs a fault."
- Severity: High
- Kind: not-working
- Where: Review run w-1 (Compressed-air works) > "Set up another run" > Run panel, "Faults to inject", row 1 ("At minute 10, compressor-1 · air compressor, has a Choose a fault"), then "Start the run".
- Steps:
  1. Open http://127.0.0.1:8766/#/run?world=w-1.
  2. Click "Set up another run". The panel says "Prefilled from w-1: seed, duration, mode and faults. Change them if needed."
  3. Row 1 shows minute 10 and compressor-1, but the fault select reads "Choose a fault". Row 2 is filled ("power-meter-air · load surge"). (Reproduced twice: once from the review button, once from a fresh load of the same review URL.)
  4. Do not choose a fault. Click "Start the run".
- Observed: The form does not stop Start, and no "needs a fault" message appears. The run starts and fails: the panel says "The run did not finish. run: inject fault: world: unknown fault "air_compressor__fails_to_load" What ran before the failure is kept as run w-6." The failed run w-6 is now in the plant's runs list. The select's blank value and the submitted id disagree.
- Expected: The prefilled fault is either filled in or shown as needing a choice; Start is blocked with "Fault 1 needs a fault" as it is for a new row (no failed run is saved). A fault id the current plant does not know is never sent. Note: the plant file changed after w-1 (review banner), so the old id may no longer exist; the form should say so.
- Suggestion: Map the prefilled fault to the current plant's fault list; if it no longer exists, mark the row "This fault is not in the plant any more: choose one" and keep Start blocked.
- Side effect: My probe created failed run w-6 in Compressed-air works (kept by the binary; it cannot be removed from the canvas). Nothing else was changed.

## ISS-4H-10 A failed run shows the binary's raw error: "run: inject fault: world: unknown fault …"
- Status: fixed e3a8346
- Fix: The failure is one sentence for the step that failed (for a refused injection, the fault and the instance); the binary message is in a "Binary message" disclosure.
- Severity: Medium
- Kind: unclear
- Where: Run panel banner after a failed Start, "The run did not finish." (Compressed-air works, from ISS-4H-09).
- Steps: Reproduce ISS-4H-09 (Set up another run, then Start the run with the stale fault).
- Observed: Banner text: `The run did not finish. run: inject fault: world: unknown fault "air_compressor__fails_to_load" What ran before the failure is kept as run w-6.` The words "world", "run: inject fault:" and a schema-style fault id are internal. The top bar toast says only "The run of Compressed-air works did not finish". The sentence "What ran before the failure is kept as run w-6" runs into the error without a full stop.
- Expected: One plain sentence that names the problem and the fix: "Fault 1 is no longer in this plant (compressor-1 · fails to load). Choose a fault and start again. The partial run w-6 is kept." (G-5, G-6, N-7.)
- Suggestion: Map binary errors to the affected row where possible; show the raw message only in a details disclosure.

## ISS-4H-11 A run that did not finish is listed and reviewed as a normal run: "Verified", "10 min", "None" disturbances
- Status: fixed fa99419 (canvas: the Home row reads Did not finish with the reason, 4a5e88c for the review header; e2e in e2e/abandoned-run.spec.ts). Binary part fixed b333d38 and 56a910d (run listing carries complete and incomplete_reason).
- Fix: A run that did not finish reads "Did not finish" on its Home row and in its review header, with the recorded reason as tooltip, detail line and header reason. Root cause (binary, fixed in b333d38): a world destroyed before run.end was recorded complete; the canvas now reads the listing's complete field.
- Severity: High
- Kind: not-working
- Where: Workspace "Runs" row for w-6 (Compressed-air works, 15 Oct 2026, 02:01, 10 min, no faults) and its review "Run w-6" (http://127.0.0.1:8766/#/run?world=w-6). Failed run from ISS-4H-09.
- Steps:
  1. Workspace: find the w-6 row in Runs. Read its chips.
  2. Click Review on w-6.
  3. Read the Report tab: verdict chip, the header, DISTURBANCES AND EFFECTOR CALLS.
- Observed: The Runs row shows "Not checked" and the same layout as a finished run, with no failure mark. The review header reads "Explore · seed 1 · 10 min" with a green "Verified" chip; the report says "Verified: the files match and the replay reproduces" and "Explore run: the simulated truth below was visible while it ran"; "DISTURBANCES AND EFFECTOR CALLS: None in this run." The plan had two faults (run w-6 was planned as the one-hour plan with 2 faults). The only sign is the top-bar toast "The run of Compressed-air works did not finish", which is not repeated in the row and disappears.
- Expected: A run that did not finish has its own state (RUN-5: complete, cancelled, interrupted, failed each say what evidence was kept). The row and the review say "Did not finish after 10 of 60 min; kept evidence: …" and do not show the Verified chip as if the whole plan ran. Disturbances show the planned faults and which were not injected.
- Suggestion: Add a "Did not finish" state with its own chip in Runs and the review header; show the planned duration next to the run's duration.

## ISS-4H-12 Effector outcome is cut to "applie…" in the review, and the run panel calls the same call "accepted"
- Status: fixed 7c8f079 (part)
- Fix: The Outcome column wraps instead of being cut. Not changed: the run panel says "accepted" (the binary's acknowledgement) and the review says "applied" (the outcome); one vocabulary is a binary-words decision.
- Severity: Low
- Kind: unclear
- Where: Review run w-7 (Compressed-air works), Report tab, "EFFECTOR CALLS" table, column "Outcome"; and the run panel completion message "Run w-7 finished. Every effector call was accepted."
- Steps:
  1. Plant > Run plant. Add effector call: at minute 5, compressor-2 · air compressor, calls "set running", running 1. Start the run.
  2. The run panel says "Run w-7 finished. Every effector call was accepted." with "00:05:00 · compressor-2 · set running: accepted".
  3. Click "Open the review". In EFFECTOR CALLS the Outcome cell reads "applie…".
- Observed: The cell text is "applied · 312 ms" in a 74 px cell with an ellipsis, so the word is cut and the time is hidden. The same call is "accepted" in the run panel and "applied" in the review, with no explanation of the difference.
- Expected: Tables show their outcome in full (V-12). One word for one state across screens (G-6), or a sentence that says both: "accepted by the simulator, applied at 312 ms".
- Suggestion: Widen the Outcome column or wrap it; use one outcome vocabulary and show the milliseconds in the details.

## ISS-4H-13 A run that does not exist opens a stuck screen: "Opening the plant this run used…" with a raw file error
- Status: fixed 7844cc9
- Fix: A run whose files are not found reads "Run w-99 is not in this workspace." with "Back to the workspace", instead of the opening message. The raw file error still shows in the Report side panel.
- Severity: Medium
- Kind: not-working
- Where: Deep link http://127.0.0.1:8766/#/run?world=w-99 (no such run). Top bar "No report" chip, Report tab.
- Steps:
  1. Type the URL above (or open it from a bookmark).
  2. Wait 5 seconds.
- Observed: The top bar reads "Run w-99 · Plant" and the centre stays on "Opening the plant this run used…" indefinitely. The only message is in the Report panel: "report: runs/w-99 holds neither run.json nor journal.jsonl: file does not exist (file_not_found)". The top bar has no link back to the workspace except the logo, and the header shows no breadcrumb.
- Expected: A missing run gets a designed state: "Run w-99 is not in this workspace." with a link "Back to the workspace" (G-4 empty/error states, N-7 recovery that works without finding a file).
- Suggestion: Resolve or refuse the run before the graph opens; show the message in the centre, with the file path hidden in details.

## ISS-4H-14 A plant that cannot be opened shows a raw error and no way back
- Status: fixed 91b320e
- Fix: A plant that is not in the workspace shows the breadcrumb, a plain sentence and "Back to the workspace"; the binary's message is in a Details disclosure.
- Severity: Medium
- Kind: unclear
- Where: Deep link http://127.0.0.1:8766/#/plant?domain=domains%2Fdoes-not-exist.domain.json. Top bar: logo, Outline, Inspector, Keyboard, Theme.
- Steps:
  1. Open the URL above.
  2. Read the centre of the screen. Look for a way back.
- Observed: The centre shows, in small red text at the top left: "The plant could not be opened: canvasbridge: file not found: domains/does-not-exist.domain.json (not_found)". The breadcrumb ("Workspace") disappears, so there is no link back to the workspace; the only route is the browser back button. "canvasbridge", "not_found" and the file path are internal.
- Expected: A plain sentence with a button: "This plant is not in the workspace any more. Back to the workspace." (G-5, G-6, N-7.)
- Suggestion: Show an empty-state panel with "Back to the workspace"; keep the file path and code in a details line.

## ISS-4H-15 "Run plant" after a failed run reopens the old failure banner, not a fresh plan
- Status: fixed fc37de9
- Fix: Run plant clears an earlier failed launch before the panel opens, so it opens a fresh plan with Start. Verified by the code path and the launch tests; not reproduced in Chrome (needs a refused run).
- Severity: Low
- Kind: unclear
- Where: Plant screen "Compressed-air works", "Run plant" button in the top bar, then the run panel.
- Steps:
  1. Reproduce ISS-4H-09 (run w-6 did not finish).
  2. Go to the plant (Workspace > Compressed-air works, or the plant link), and click "Run plant".
- Observed: The run panel opens on the old failure: "The run did not finish. run: inject fault: … What ran before the failure is kept as run w-6." with only "Change the plan". There is no Start button until you press "Change the plan". The plant top bar still shows the toast about w-6.
- Expected: "Run plant" opens a fresh plan (30 min, no faults, Start). The failure is shown on the review of w-6, not as the plan's state.
- Suggestion: Keep the last failure with its run; reset the run panel when it is opened from "Run plant".

## Simulator behaviour notes

Probes run on compressed-air-works (runs w-7, w-8, w-9) and hunt-empty (run w-5). Verdict: supported | partly | missing.

1. Stale (stuck) sensor: fault "power · stale" on power-meter-air at 00:15 (w-8). The panel shows the fault, but the power channel keeps moving (254 kW at 15:00, 256 kW at 25:00; timestamps advance each 5 min). A stuck reading should freeze, so a consumer cannot tell it from a live one. Verdict: partly.
2. Drifting sensor: fault "discharge temp · drift" on compressor-1 from 00:08 (w-8). Readings: 89.0 °C at 07:51 (before), 89.0 °C at 20:07, 89.1 °C at 25:10. The drift is about 0.1 °C over 17 min, with no alarm. Verdict: partly (applied, but too small to matter in 30 min).
3. Command not executed, reported as accepted: "set loaded" on compressor-2 at 05:00 while compressor-2 "fails to load" (w-9). Run panel: "Every effector call was accepted. 00:05:00 · compressor-2 · set loaded: accepted". Load status stays "on" (no change), no refusal; the review reports the call as "applied · 335 ms". Verdict: missing (silent no-effect; the result says nothing about the failed device).
4. Command refused: "shelve alarm" on compressor-2 at 20:00 with no alarm active (w-8). Accepted, review outcome "applied · 281 ms"; ALARM SYSTEM shows 0 activations. Verdict: missing (a refusal or "no alarm to shelve" was expected: "the simulator may refuse it, and the result says why").
5. Device failed and stays failed: "fails to load" on compressor-2 from 02:00 (w-9). It stays failed to 25:11, and air flow turns negative (−2.20 m³/h at 25:11, shown as a plain value; this is the known V-8 gap). Run status and load status stay "on", alarm state stays "normal", and no alarm is raised. Verdict: partly (persists, but the status channels say it runs and nothing alarms).
6. Alarm that never clears: no alarm fired in any probe run (activations 0, standing at end 0), even with a failed compressor and a drifting sensor. Cannot be tested until alarms fire. Verdict: missing (no alarm produced).
7. Feeder trip / load surge: power-meter-air "load surge" at 20 min in w-1 (earlier run, not compared at the same minute). Not verified in detail. Verdict: not probed.
8. Comms gap, duplicated or late readings: not probed (the canvas offers "dropout" faults such as "power · dropout", "run status · dropout", which are the nearest; not run for time). Verdict: not probed.
9. Unknown fault id after the plant changed: the prefilled fault "air_compressor__fails_to_load" from w-1 is rejected by the binary as unknown (run w-6 kept as partial, reviewed as "Verified", see ISS-4H-09 and ISS-4H-11). Verdict: partly (refusal is explicit, but the partial run is presented as a complete verified run).
10. Unconnected equipment runs as healthy: hunt-empty (chiller WCC-1 and pump CP-1 with no links, the pump linked to nothing) shows steady readings for 30 min: pump flow 166 m³/h, power 16.4 kW, chiller COP 7.43, supply 7.00 °C. A real plant with no load, no flow and no water would not produce these. Verdict: missing (model runs unconnected equipment as healthy).
11. Condenser water: the pump output "cw flow out" (condenser water) has no input on the water-cooled chiller, whose "cond water in" takes a cooling-water temperature (link form: "No instance takes condenser water… The inputs in this plant take thermal load, chilled water, cooling water temp and tariff"). A condenser loop cannot be wired with a pump. Verdict: missing (library lacks a condenser-water input; the form message is clear).
12. Units: "Specific power" for the chiller (w-5) and the compressors (w-8) shows 0.14 and 0.11 with no unit on the chiller; "production-line-a" shows "kW/%". Verdict: partly (unit missing or odd for some KPIs).

## Checked and not reproduced

- Empty plant from "Chiller pilot" (orchestrator check). On `hunt-empty`, the "An empty plant" radio was ref-clicked and read back as checked (`new-plant-copy`, second radio `checked: true`, first `false`). Create and open the editor gave 0 instances and one area, "Pilot plant" (the site area). The choice was honoured, so this is not an ISS. The author's report (4 instances, Utilities / Chiller hall / Office block) did not reproduce on this route; if it persists, check which route the author used.
- The first ref-click on "New plant" (Home) did nothing; the second click at coordinates opened the form. Not repeated, so not reported.
- Duplicate plant id: "hunt-empty" with Create was refused with "A plant with the id hunt-empty already exists." Nothing was overwritten. Works.
- Graph click selects the instance in the Run panel (R-3): "compressor-2" selected by clicking its node; the fault row picked it up. Works.
- Fault list is filtered to the chosen instance (R-2): air-dryer-1 offers "desiccant exhaustion, regen valve stuck, …" and compressor-2 offers "airend wear, fails to load, …". Works.
- Undo, redo and Save (E-7): undo steps through move, area add, rename, add; redo restores them; Save reports "Nothing to save" afterwards. Works (layout note in ISS-4H-04).
- Link form: a pump's "cw flow out" to the chiller shows the binary's reason ("No instance takes condenser water…"). Fixed ISS-3-09 holds. Linking "flow out" to "flow in" works and the link list shows "flow out → WCC-1 flow in" with a labelled remove button.
- Keyboard help (?/Keyboard button): dims the page, real keys, Escape closes it. Fixed G-15 holds.
- Run panel validation: an empty fault row and an empty effector argument are flagged next to the field ("Fault 1 needs a fault.", "Effector call 1: duration s needs a value.") and Start stays put. Works (R-6).
- Run completion: "Run w-7 finished. Every effector call was accepted." with the Open the review button; toast in the top bar. Works.
- Playback: Play/Pause, speeds 1×/10×/60×/600×, scrubber, and fault chips seek to the event (label). Works except ISS-4H-08.
- Unknown plant and missing run show errors rather than a blank page (but see ISS-4H-13 and ISS-4H-14 for the wording and the stuck state).

## Worked well

- The Keyboard help and the theme control are labelled (G-13, G-15 fixed).
- Run completion and the review are quick (30 min plan in a few seconds).
- The fault list reads as sentences and is per instance; the run panel explains the plan in one line before Start.
- Verified chip and "The plant file changed after this run" banner on w-1 make it clear the review uses the recorded plant.
- Duplicate plant ids and invalid ids are refused with the binary's rule in words.
- Undo/redo and Revert-state are consistent across add, rename, move and link.

## Cleanup and side effects

- Plants I created: `hunt-empty` (saved, 2 instances, 1 link removed, 1 run w-5). `hunt-bad-id` was never created (Create refused). Nothing else of the author's was changed.
- Runs I created on compressed-air-works: w-6 (failed, partial, from ISS-4H-09), w-7 (effector probe), w-8 and w-9 (fault and effector probes). w-6 cannot be removed from the canvas; it is listed in the workspace Runs.
- The author's tab (770933647, Office tower HVAC / Water pumping station) and a second tab (770933650, Compressed-air works plant) were not touched by me.

## Summary

Verified findings: 15 (ISS-4H-01 to ISS-4H-15). Stopped well below the 25 cap; the remaining plan items (comms gap, duplicated readings, dark-mode contrast, keyboard tab order on the run panel) were not reached.

Counts by severity: High 2, Medium 8, Low 5.
