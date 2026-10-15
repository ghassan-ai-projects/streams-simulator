# Round 3 authoring issues (ui-polish-sim-authoring-20261015)

Tester: simulation author (Haiku 5.5), Chrome tab on http://127.0.0.1:8766, viewport 1440x900.
Assignments: office-tower-hvac (complete + compose + run), compressed-air-works (complete + run), water-pumping-station (new, compose + run).
Round-1 and round-2 issues marked fixed were not re-reported; only regressions and new problems are listed.

## ISS-3-01 Composition still shows raw schema text for profiles, and the two plants cannot compose
- Severity: High
- Kind: not-working
- Where: Build panel block "ISSUES FROM THE BINARY"; status line "Needs fixing: 2 issues. Your changes are saved." (editor top bar); Home card "Not runnable yet: open Edit, save, and fix any issue it lists." (Office tower HVAC, Water pumping station)
- Steps:
  1. Home > Edit on "Office tower HVAC" (already has weather, site calendar, two AHUs, chillers, pumps, cooling tower, meter).
  2. Save (button "Save and compose (Ctrl+S)").
  3. Open the Build panel (button "Build"), read "ISSUES FROM THE BINARY".
  4. Repeat with a new plant: Home > "New plant", Name "Water pumping station", Id "water-pumping-station", Start from "Chiller pilot", "An empty plant", Create; add equipment and areas; Save.
- Observed: Both plants show exactly two issues: "array shorter than minItems 3" and "no array item matches contains", each followed by "in /profiles". The status reads "Needs fixing: 2 issues. Your changes are saved." Home still says "Not runnable yet". No Run button is offered for either plant. Compressed-air works (also has a site calendar, no weather) saves and reads "Ready to run. Composed 28 channels..." so the problem is specific to these plants' profiles. Nothing in the UI lets the author add or edit profiles.
- Expected: Plain-language text such as "needs at least 3 profiles: add nominal, correlated_cascade and sensor_pathology" (the wording recorded as the fix for ISS-2-08), and a place in the Build panel to add them. A plant should be able to reach a run.
- Suggestion: Confirm the ISS-2-08 wording reaches the composed plant (this looks like a regression of that fix), and let the author add the missing profiles from the Build panel.
- Status: binary fixed (see commit log: fix(component) empty plant carries the required profiles); canvas profile editor open
- Fix: Root cause was `component new --empty` (`internal/component/internal/domain/new_plant.go`): it dropped the base plant's `not_applicable` profiles, so an empty plant from the chiller pilot kept only 2 of the 3 profiles the domain spec requires. The names of the required profiles are now read from the domain-spec schema (its `profiles` allOf/contains clauses); an empty plant keeps the applicable base profiles without their fault weights and adds any required one that is missing as a plain profile (`duration_s` 86400). Regression: the facade test composes an empty plant from every shipped example once one pump is added. The ISS-2-08 wording did reach the plant (it is the "needs ..." text); the cause was the missing profiles, not the wording.

## ISS-3-02 "Not runnable yet" points to a list that only appears inside the Build panel after Save
- Severity: Medium
- Kind: unclear
- Where: Home plant card "Not runnable yet: open Edit, save, and fix any issue it lists."; editor status "Needs fixing: 2 issues. Your changes are saved."; Build panel "ISSUES FROM THE BINARY"
- Steps:
  1. Home: read the card under "Office tower HVAC" and "Water pumping station".
  2. Click "Edit"; the status shows "Needs fixing: 2 issues. Your changes are saved."
  3. Find the issues: they are not shown under the status; I had to open "Build" (it is hidden when an instance is selected).
- Observed: The status and the card say "fix any issue it lists" but the list is not visible on the editor's first screen, and the status line does not say where the list is. After closing the inspector, the list appeared only in the Build panel, and it disappears again after a reload.
- Expected: The status line names where the issues are ("see Build > Issues") or the list stays visible next to the status.
- Suggestion: Make the status a link or button that opens the issue list.
- Status: fixed 3ce3e35
- Fix: The editor status reads "Needs fixing: N issues, listed in Build. Your changes are saved." and a "Show issues" button in the top bar deselects the instance and opens the Build panel's issue list. Tests: EditorStatus.test.tsx, EditorPanel.test.tsx.

## ISS-3-03 Areas cannot be renamed or deleted, so starting from an example or empty copy leaves unwanted areas behind
- Severity: Medium
- Kind: improvement
- Where: Build panel "ADD AREA" form ("Area name", "Inside", "Add area"); left Outline area headers; area select "Area" in the inspector.
- Steps:
  1. Compressed-air works: add areas "Compressor house", "Production hall A", "Production hall B" under "Site", move every instance into them.
  2. Look at the Outline: "PRODUCTION HALL" (0), "COMPRESSOR ROOM" (0), "CHILLER PLANT" (0), "SUBSTATION" (0) remain.
  3. Click an area header in the Outline: it only collapses the group. No rename or delete control appears in the inspector or the Build panel.
  4. Office tower HVAC: same result for "UTILITIES", "CHILLER HALL", "OFFICE BLOCK" after moving all instances.
- Observed: Empty areas stay in the Outline and in every "Area" select. There is no way to rename or remove an area, so the brief's tree (Site > ...) sits next to leftover areas.
- Expected: An area can be renamed, and an empty area can be removed (or the start copy does not bring areas the user did not ask for).
- Suggestion: Add "Rename" and "Delete (empty areas only)" for areas, like instances have.
- Status: fixed fc8cef3
- Fix: The Build panel has an "Areas" list: each area can be renamed (its child areas, its instances' area and every Area select follow) and deleted only when no instance is in it and no area is inside it; a row that cannot be deleted says why ("Move the 2 instances out of it first."). The rules are in compositionEdits.ts. Tests: AreaEditor.test.tsx, compositionEdits.test.ts, areas.spec.ts. The Outline headers still only collapse.

## ISS-3-04 Empty plant from "Chiller pilot" starts with Utilities / Chiller hall / Office block, not a site
- Severity: Low
- Kind: unclear
- Where: New plant form ("Start from", "An empty plant with only its site, areas, profiles and axes"); editor Build panel "Area" select and Outline
- Steps:
  1. Home > "New plant", Name "Water pumping station", Start from "Chiller pilot", "An empty plant", "Create and open the editor".
  2. Read the Outline and the Build panel "Area" select.
- Observed: The new plant already has top-level "UTILITIES" (with "CHILLER HALL" and "OFFICE BLOCK" under it). Default Area is "Utilities". The text "its site, areas, profiles and axes" does not say what an "axis" is or that the areas come from the chosen start plant. To get "Site" I had to add it as a new top-level area.
- Expected: The start point's areas are described, or the empty plant starts with a single "Site" area.
- Suggestion: Name the source of the areas on the form, or start an empty plant with one "Site" area.
- Status: fixed a0e7b1f (one site area), with 88ebe63 (the form wording)
- Fix: The New plant option now reads "An empty plant with the site, areas, profiles and axes of the start plant and no equipment". The binary: `emptyPlant` in internal/component/internal/domain/new_plant.go keeps the site and starts an empty plant with one top-level area named after the site ("Site" when unnamed). Tests: internal/component/new_plant_test.go and new_plant_test.go in the domain package. The canvas fake (`component.new`) and plantMaker.test.ts follow it (02d7f75). Test: HomeScreen.test.tsx.

## ISS-3-05 Typed instance Name is dropped when the form re-renders; the suggestion placeholder looks like a typed value
- Severity: Medium
- Kind: not-working
- Where: Build panel "ADD EQUIPMENT" > "Name" textbox (placeholder shows the suggested id, for example "chiller-ch-1", "PCV-1", "FM-1", "SC-1", "STW-1")
- Steps:
  1. Build panel: choose Equipment "Water-cooled chiller" (or "Pneumatic control valve with positioner").
  2. Immediately click the Name field and type "chiller-ch2" (or "valve-1").
  3. Screenshot before adding.
- Observed: For the chiller, the Name field showed the grey suggestion "chiller-ch-1" and the typed text was gone; the instance was not created with my name. It only worked after waiting for the form to settle and clicking again. The grey suggestion is styled like a value, so it looks entered.
- Expected: Typed text stays in the field. The suggestion should be visibly a hint, not a value.
- Suggestion: Do not reset the Name field after choosing the equipment, or make the suggestion a distinct hint style.
- Status: fixed 464453c
- Fix: Typed name is kept when equipment is chosen or the plant re-renders (the choice no longer clears it); the suggestion is an italic placeholder. Tests: EditorPanel.test.tsx.

## ISS-3-06 New plant "Id" field appends typed text to the id derived from the Name
- Severity: Medium
- Kind: unclear
- Where: Home > "New plant" form, fields "Name" and "Id" (the same pattern as ISS-2-04, in a different field)
- Steps:
  1. Home > "New plant". Click "Name" and type "Water pumping station". Click "Id" and type "water-pumping-station".
  2. Read the Id field.
- Observed: Id had already been filled as "water-pumping-station" from the Name, and typing appended: the field read "water-pumping-stationwater-pumping-station". The form still accepted it until I triple-clicked and retyped.
- Expected: Typing in Id replaces the derived value, or the Id stays derived until the user edits it.
- Suggestion: Select the derived Id text when the field is focused, or stop deriving it once the user has typed in Id.
- Status: fixed 04dfd17
- Fix: Id field selects its derived text on focus, so typing replaces it; deriving stops once Id is typed. Tests: NewPlant.test.tsx (shared harness in src/test/homeHarness.tsx).

## ISS-3-07 Instance subtitle shows an internal kind number: "chiller 2", "air compressor 2", "storage tank 2" on the first instances too
- Severity: Medium
- Kind: unclear
- Where: Outline list and inspector subtitle (for example "chiller-ch1 chiller 2" and "chiller-ch2 chiller 2"); Run page Outline "compressor-1 air compressor 2" and "compressor-2 air compressor 2"; Run Inspect fault dropdown "compressor-1 · air compressor 2"; Build panel after adding "reservoir-tank storage tank 2"
- Steps:
  1. Office tower HVAC: open the Outline; both "chiller-ch1" and "chiller-ch2" show "chiller 2".
  2. Compressed-air works: both "compressor-1" and "compressor-2" show "air compressor 2".
  3. Water pumping station: after adding "reservoir-tank" it shows "storage tank 2" while "intake-tank" shows "storage tank".
- Observed: The second word after the kind appears on some instances but not others, and it is the same for every instance of that kind in a plant. It is not an instance number and is not explained anywhere.
- Expected: A plain kind name ("chiller", "air compressor") or a visible instance number that differs per instance.
- Suggestion: Show the kind's display name only; drop the internal kind suffix from the subtitle.
- Status: fixed 2983767
- Fix: Kind labels drop the internal numeric suffix (kindLabel in model/report.ts) in the outline, inspector, editor subtitle, run fault and call lists, and plant kind tally. Component titles still win where the library has one.

## ISS-3-08 Clicking Build while an instance is selected shows the instance inspector, not the Build panel
- Severity: Low
- Kind: unclear
- Where: Top bar "Build" button; inspector panel header (instance name) with the close "X"
- Steps:
  1. Editor: click an instance on the graph (for example "site-calendar").
  2. Click "Build" in the top bar.
  3. Read the right panel heading.
- Observed: The panel shows the instance inspector ("Link an output", "Delete ..."), not the "Build" heading and "ADD EQUIPMENT". Clicking "Build" again toggles the whole panel closed. The only way to get Add equipment back was to close the inspector with the "X".
- Expected: "Build" opens the Build panel (switching from the inspector), and the Build button does not toggle the panel closed unexpectedly.
- Suggestion: Make "Build" always show the Build panel, and let the inspector close on its own.
- Status: fixed cc5dc44
- Fix: Build with an instance selected deselects it and shows the Build panel; with nothing selected it toggles as before. Tests: showBuild.test.ts, PanelToggles.test.tsx.

## ISS-3-09 A pump's flow output has no target; the link form says "No instance takes water yet" without saying why
- Severity: Medium
- Kind: unclear
- Where: Inspector "Link an output" form, "Output" = "flow out" on "pump-1" (water Pump house); "To" select
- Steps:
  1. Water pumping station: select "pump-1". Output "flow out".
  2. Read the "To" list: "No instance takes water yet".
  3. Compare "valve-1" (Output "flow out" is offered and reaches "reservoir-tank").
- Observed: The valves' only inlet is "pressure in", so pump "flow out" cannot be linked to any valve. The message does not say that the library has no water inlet that accepts a pump's flow. I could not show pump-to-valve flow and wired the valves to the reservoir only.
- Expected: The message names the reason (for example "control valves take pressure, not water") or the library offers a compatible inlet.
- Suggestion: Explain why the To list is empty, or add a water inlet to the control valve.
- Status: fixed efdaf79
- Fix: The empty To list says why, from the output's type and the plant's inputs: "No instance takes chilled water: pump-1's flow out is chilled water. The inputs in this plant take liquid flow and pressure." The pump's "flow out" is chilled water and the tank takes liquid flow, so both read as water but do not match. Tests: ports.test.ts, ConnectForm.test.tsx.

## ISS-3-10 Reservoir port "outflow in" used for a valve feed; the port names do not say which side is feed
- Severity: Low
- Kind: unclear
- Where: Inspector "Link an output" > "Input" select for "reservoir-tank" (options "inflow in", "outflow in", "level out")
- Steps:
  1. Water pumping station: select "valve-2". Output "flow out". To "reservoir-tank". Read the Input list.
  2. Choose "outflow in" (I chose it; "inflow in" was already used by valve-1 and I was not sure whether one inlet takes two feeds).
- Observed: The link "flow out -> reservoir-tank outflow in" was created. Nothing says whether "outflow in" is a valid feed for a valve or what the tank does with it. The guess was mine.
- Expected: Port names or a short hint make the feed side clear, or a second feed can be added to "inflow in".
- Suggestion: Add a one-line hint on each port (for example "feeds the tank" / "draws from the tank").
- Status: needs-binary: a port's declared description is not in the component contract
- Fix: No canvas change. The port spec carries only name, type, direction (provides/consumes) and state (`contracts/component-v0.1.schema.json`), and the canvas already shows direction. Proposal: an optional per-port `description` in the component contract and the builtin library (for example storage-tank `outflow_in`: "draws from the tank"), which the canvas would show under the Input choice.

## ISS-3-11 Enter in the Name field does not add the equipment
- Severity: Low
- Kind: improvement
- Where: Build panel "ADD EQUIPMENT" > "Name" textbox and "Add to the plant" button
- Steps:
  1. Choose Equipment "Centrifugal pump", Area "Site > Pump house", type Name "pump-2".
  2. Press Return.
- Observed: Nothing happened; the form still shows "pump-2" and the button "Add to the plant" must be clicked.
- Expected: Return adds the equipment, as in most forms, or the hint says "click Add to the plant".
- Suggestion: Submit on Return in the Name field.
- Status: fixed f023c8a
- Fix: Not reproduced on the current build: Return in Name already submits and adds the equipment in Chrome. Added a Chrome e2e guard (editor.spec.ts) so it stays that way; no product code changed.

## ISS-3-12 "Run again" does not start a run; it opens the run setup with the old faults filled in
- Severity: Low
- Kind: unclear
- Where: Run page (run w-4) top right button "Run again"
- Steps:
  1. Open run w-4 for Compressed-air works (Home > Review).
  2. Click "Run again".
- Observed: The page goes to "Opening the plant..." and then the plant's run panel "Run Compressed-air works" with seed 1, duration 30 minutes, and both faults (compressor-2 at 10, compressor-1 at 20) already filled in. Nothing is started until "Start the run" is pressed. The label suggests the run starts immediately.
- Expected: Either a run starts, or the label says "Set up another run" / the panel says it is prefilled from w-4.
- Suggestion: Rename the button or add a line "Starts with the faults of w-4; change them if needed."
- Status: fixed 63be44c
- Fix: Button is now "Set up another run"; the panel note reads "Prefilled from w-4: seed, duration, mode and faults. Change them if needed." (mode added because it is copied too). Tests: RunHeader.test.tsx, LaunchPanel.test.tsx.

## ISS-3-13 Home card counts 7 instance definitions for Compressed-air works after it gained an 8th instance
- Severity: Low
- Kind: not-working
- Where: Home plant card "Compressed-air works": "7 instance definitions · 4 runs · last 15 Oct 2026, 01:06"; plant view header "8 instances · 13 links"
- Steps:
  1. Edit Compressed-air works: add "compressor-2", link it, Save and compose ("Ready to run").
  2. Go Home: read the card. Then Run > plant view: read "8 instances · 13 links".
- Observed: Home says 7 instance definitions; the plant view says 8 instances. The card's "definitions" wording is still in place (ISS-2-17 canvas label open).
- Expected: The card shows the same count as the plant view.
- Suggestion: Refresh the card count on every save.
- Status: not a binary defect; canvas servers run the old binary
- Fix: `DefinitionOf` counts expanded instances since b4c6b24 (confirmed by `TestADefinitionCountsTheInstancesItExpandsTo`: named sets by name, id_template sets by `count.default`, composition `equipment` the same way). The card shows 7 because the running bridge is the old build, which counts `len(instances)` (definitions); restart the servers on a binary built after b4c6b24. The "definitions" label in the canvas is still open (ISS-2-17).

## ISS-3-14 Return in the "Find equipment" search submits the add form
- Severity: Medium (found in review of the add form, not in the round-3 file)
- Kind: not-working
- Where: Build panel "ADD EQUIPMENT" > "Find equipment" search box
- Steps: choose an Equipment, type in "Find equipment", press Return.
- Observed (code review): the search box sits inside the add form, so Return is an implicit submit and adds the chosen equipment with the typed name, while the user only meant to filter the list.
- Expected: Return in the search box only filters.
- Status: fixed 405b12a (fixed together with ISS-3-15)
- Fix: Return in the search box is consumed (no implicit submit). Equipment is added by the Add button or by Return in the Name field. Tests: EditorPanel.test.tsx (keyDown is default-prevented, nothing added); editor.spec.ts "Enter in the equipment search only filters...".

## ISS-3-15 A name another instance already uses is refused only after Save
- Severity: Medium (found in review of the add form, not in the round-3 file)
- Kind: unclear
- Where: Build panel "ADD EQUIPMENT" > "Name"; the editor status after Save
- Steps: add an instance named "CT-1" when "CT-1" already exists; press "Add to the plant".
- Observed (code review): the form accepts the name, and the binary refuses the duplicate only at Save, so the error appears later on the instance and in the status.
- Expected: The form says it at once: "<name> is already used by another instance." and does not add.
- Status: fixed 405b12a (fixed together with ISS-3-14)
- Fix: The add form checks the name against the plant's instance list (the plant shown in the editor, including unsaved adds) and shows "<name> is already used by another instance." in place of the Add button's reason; the add is refused. Tests: EditorPanel.test.tsx, editor.spec.ts "Enter in the equipment search only filters; a name another instance uses is refused in the add form".

## Simulator behaviour notes

Probes were run on Compressed-air works run w-4 (30 min, Explore, seed 1; faults: compressor-2 "fails to load" at 00:10, compressor-1 "discharge temp · drift" at 00:20). Office tower HVAC and Water pumping station could not be run (see ISS-3-01), so no probes were possible there.

### Probe 1: device fails and stays failed (compressor-2, fails to load)
- Tried: Run w-4 Inspect at 00:14:27 and at 00:24:59 on compressor-2 (after the fault at 00:10).
- UI offered: Inspect panel with "Fault: fails to load", channels air flow, run status, load status, alarm state, discharge temp, power, vibration, and TRUE KPIs.
- Simulator: Fault shown at both times. Air flow reads -0.50 m3/h at 14:25 and 6.00 m3/h at 24:59 (near zero, noisy). Run status "on" and load status "on" stayed with timestamp 00:00:05, i.e. they never changed after the fault. Alarm state "normal".
- Expected in a real plant: a compressor that fails to load goes unloaded, its load status changes, and an alarm is raised or the device is flagged as tripped.
- Verdict: partly (failure persists to the end of the run; load status, run status and alarm do not reflect it).

### Probe 2: drifting sensor (compressor-1, discharge temp drift)
- Tried: Inspect compressor-1 at 00:24:59.
- UI offered: "Fault: discharge temp · drift" with the discharge temp channel.
- Simulator: discharge temp reads 89.0 C at 00:24:50 (observed); values drift upward after 00:20.
- Expected: the drift shows in the reading and a drift alarm or high-temperature alarm eventually asserts.
- Verdict: supported for the drift in the value; the alarm part is missing (see probe 3).

### Probe 3: alarm never raised for an out-of-range reading
- Tried: Same as probe 2; Report tab "ALARM SYSTEM ISA-18.2".
- Simulator: discharge temp 89.0 C with alarm state "normal"; ISA-18.2 summary shows activations 0, peak per 10 min 0, standing at end 0.
- Expected: a high discharge temperature alarm activates and then clears or stays standing.
- Verdict: missing (no alarm was raised for this reading in this run).

### Probe 4: out-of-range or impossible reading
- Tried: Inspect production-line-a at 00:24:59 and compressor-2 at 00:14:27.
- Simulator: air flow shown as "-3.50 m3/h" on production-line-a and "-0.50 m3/h" on compressor-2, with a warning-circle icon next to the value, alarm state "normal".
- Expected: a negative air flow is either rejected or flagged as a stuck/dead sensor with an alarm.
- Verdict: partly (value flagged by an icon only; no alarm, no reason shown).

### Probe 5: command sent but not executed (start/stop, set-point)
- Tried: Looked for any start/stop or set-point control in the plant inspector, the Run form and the review page.
- UI offered: The run review shows "EFFECTOR CALLS: None". No control to call an effector exists in the UI (consistent with the open canvas side of ISS-2-14).
- Verdict: missing (cannot test from the UI).

### Probe 6: feeder trip, communication gap, duplicated or out-of-order readings
- Not attempted: I did not add a feeder or utility fault, a communication-gap fault or a duplicate/out-of-order fault in this round. Not tested.

## Could not do
- Flow meter and pressure meter (brief for the water station): the equipment library has no flow or pressure meter. Only sensors per port are generated automatically; the Feeder meter is power only.
- Pump flow into a control valve (water station): no compatible inlet on the library's control valve (see ISS-3-09).
- Compose and run of office-tower-hvac and water-pumping-station: blocked by ISS-3-01 (profiles). No way to add profiles in the UI.
- Effector calls (start/stop of pumps, drives, compressors): not offered in the UI (ISS-2-14 canvas side still open).
- Deleting or renaming leftover areas (see ISS-3-03).

## Worked well
- Adding an area with "Inside" set to a parent works and the Outline shows the nesting at once.
- Moving an instance by the inspector's "Area" select works and the graph regroups immediately.
- The "Link an output" form, once the output is chosen, lists only compatible targets, and the inspector's LINKS list shows each link in both directions.
- Compressed-air works composes after two compressors are added (two of the same kind now compose; ISS-2-01 verified).
- Run page: the timeline's fault markers, the scrubber, the Inspect tab per instance, and "Verified: the files match and the replay reproduces." all read clearly.

## Simulation notes
- office-tower-hvac (not runnable): areas Site > Plant room; Site > Floors > Floors 1-3, Floors 4-6; leftover empty Utilities, Utilities > Chiller hall, Utilities > Office block (cannot delete). Equipment (12): Water-cooled chiller x2 (chiller-ch1, chiller-ch2), Cooling tower x1 (cooling-tower-ct1), Centrifugal pump x3 (chw-pump-p1, chw-pump-p2, cw-pump-p3), Variable-frequency motor drive x1 (vfd-p1), Feeder meter x1 (plant-power-meter), Air-handling unit x2 (ahu-floors-1-3, ahu-floors-4-6), Weather x1 (site-weather), Site calendar x1 (site-calendar). Links made: chw-pump-p2 flow out -> chiller-ch2 flow in; ahu-floors-4-6 load out -> chiller-ch2 load in; chiller-ch2 supply out -> ahu-floors-4-6 chw in; cw-pump-p3 cw flow out -> cooling-tower-ct1 cw flow in; cooling-tower-ct1 basin out -> chiller-ch2 cond water in; site-weather dry bulb out -> ahu-floors-4-6 outdoor in; site-calendar shift out -> ahu-floors-4-6 shift in; site-calendar tariff out -> chiller-ch2 tariff in; chiller-ch2 power out -> plant-power-meter loads in. Actions exercised: none (not runnable). Save result: "Needs fixing: 2 issues" (ISS-3-01).
- compressed-air-works (composed, run w-4): areas Site > Compressor house (compressor-1, compressor-2, air-dryer-1, receiver-header-1, power-meter-air); Site > Production hall A (production-line-a); Site > Production hall B (packaging-line-b); leftover empty Site > Production hall, Site > Compressor room, top-level Chiller plant and Substation. Equipment (8): Oil-injected screw air compressor x2, Heat-reactivated desiccant air dryer x1, Compressed-air receiver and header x1, Feeder meter x1, Production line x1, Packaging line x1, Site calendar x1. Links: 13 (Home card and plant view). Instruments: channels generated per port (28 channels reported at compose). Faults run: compressor-2 "fails to load" at 10 min, compressor-1 "discharge temp · drift" at 20 min. Actions exercised: scrubber, Inspect per instance, Run again (opened the setup prefilled; not started), Start the run. Effectors: none offered. Scoring: none (Explore mode).
- water-pumping-station (not composed): id water-pumping-station, title "Water pumping station", started from Chiller pilot (empty). Areas: Site > Intake, Site > Pump house, Site > Reservoir; leftover Utilities, Utilities > Chiller hall, Utilities > Office block. Equipment (11): Storage tank with level control x2 (intake-tank, reservoir-tank), Centrifugal pump x2 (pump-1, pump-2), Variable-frequency motor drive x2 (drive-1, drive-2), Pneumatic control valve with positioner x3 (valve-1, valve-2, valve-3), Feeder meter x1 (pump-power-meter), Site calendar x1 (site-calendar). Links: intake-tank level out -> pump-1 level in and pump-2 level in; pump-1 and pump-2 power out -> pump-power-meter loads in; drive-1 and drive-2 power out -> pump-power-meter loads in; valve-1 flow out -> reservoir-tank inflow in; valve-2 flow out -> reservoir-tank outflow in. valve-3 is not linked. Flow and pressure meters: not in the library. Save result: "Needs fixing: 2 issues" (ISS-3-01). No run.

## Summary
- Issues: 13 (High 1, Medium 6, Low 6).
- Completed simulations: compressed-air-works (composed and run with two faults). Not completed: office-tower-hvac and water-pumping-station (both blocked from composing by ISS-3-01).
