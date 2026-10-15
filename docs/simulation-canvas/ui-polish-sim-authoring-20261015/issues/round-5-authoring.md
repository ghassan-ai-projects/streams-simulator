# Round 5 authoring report

Author agent: simulation author (builder), round 5. Plants built and run through the canvas UI in Chrome (tab 770933647). Other agents' tabs and `hunt-*` plants and `compressed-air-works` were not touched.

Assignment, in order:
1. `office-tower-hvac` - open Edit, save, fix the Build issues until "Ready to run", then run with 2 faults and 1 effector call.
2. `water-pumping-station` and `data-centre-cooling` - each run in Explore with 2 faults at different minutes and at least one effector call; open each review; run each once more in Evaluate.
3. `packaging-factory` ("Packaging factory") - build, compose, run with 2 faults and an effector call.

## Blockers

### BLK-5-01 Save is disabled on an unchanged plant, so its saved issues cannot be seen
- Where: Editor top bar, "Save" button (text "Nothing to save") and the workspace Plants row for "Office tower HVAC" ("Not runnable yet: open Edit, save, and fix any issue it lists.").
- Steps: Workspace -> Edit "Office tower HVAC" (the plant is not changed); click "Save".
- Observed: the Save button is disabled with "Nothing to save"; the top bar shows "No changes". The Build panel lists no issue, although the plant is "Not runnable yet". The issue ("Needs fixing: 1 issue, listed in Build") appears only after an edit is saved.
- Expected: opening the editor of a not-runnable plant shows its issues without a dummy edit, or Save is enabled when the plant is not ready.
- Workaround: renamed the area "Floors 4-6" to "Floors 4 to 6" and saved. The save then reported "Needs fixing: 1 issue, listed in Build" and "Show issues" listed it. Note: the area is still named "Floors 4 to 6" in this plant (not renamed back).
- Status: fixed b66485a
- Fix: a saved plant with no domain yet shows "Not checked yet" and Save is enabled; Save composes it and lists its issues. Ctrl+S uses the same rule.

### BLK-5-02 Enter in the run form's fault select starts the run
- Where: Plant view, "Run Office tower HVAC" panel, "FAULTS TO INJECT" > fault 1 "Choose a fault" select (and the instance select above it).
- Steps: Plant view -> Run plant; Add fault; pick the instance (typed "P-1" then Return); click the fault select; type "impeller" then Return.
- Observed: the second Return submitted the whole run form. Run w-10 started at once with one fault (P-1 impeller wear at 00:00, minute 0), no effector call and 30 min Explore. The run panel closed and the run review opened ("Verifying: replaying the run...").
- Expected: Return in a select only chooses the option; the run starts only from "Start the run".
- Workaround: none yet for the run itself; for picking options I type the option text and do not press Return (I continue with a click on the next field). Run w-10 is an accidental run and stays in the catalog; it is not part of the assignment.
- Status: fixed c499421
- Fix: Enter on a select in the run form no longer submits the form; only "Start the run" starts a run (unit and e2e; the e2e fails on the old build).

### BLK-5-03 Picking an area in the equipment or instance Area select is unreliable when the label has a "›" separator
- Where: Build panel, ADD EQUIPMENT > "Area" select (also the instance inspector "Area" select after placement, and ADD AREA > "Inside").
- Steps: pick "Factory › Production hall › Line 1" in the Area select: click the select, then type "Factory › Production hall › Line 1" and press Tab.
- Observed: the ASCII part reaches the select, the "›" characters land in the Name text field (e.g. Name became "CV-1› Production hall › Line 1"), and the Area stays "Factory › Utilities" or "Factory › Production hall". Clicking an option is not possible because the native select popup is not visible in the screenshot, and Down/Up keys scroll the page.
- Expected: choosing an area by its full label works by typing or by clicking an option, and typed text never reaches the Name field.
- Workaround: with the select focused, type the letter F as many times as there are options before the target (the options all start with "Factory"; Chrome cycles through the options that start with the typed letter). Confirm with find on the combobox value before Add. Used for CV-1 (Line 1). Also: Name fields must be retyped after the area change, because the panel relayouts.
- Status: won't-fix: not reproduced; canvas e2e (areas.spec) selects 'Site › Pump house' by label and Name stays unchanged; the '›' text came from keyboard input into a focused text field, not from the select
- Fix: none; the select already takes the full label.

### BLK-5-04 Choosing a fault can scroll the run panel, so a click on "Add effector call" starts the run
- Where: Plant view, Run panel, "EFFECTOR CALLS" > "Add effector call" (Packaging factory, after two faults were set).
- Steps: Run plant; Add fault; minute 20, instance PL-1, fault "product jam" (Tab after each); then click where "Add effector call" was (about y 680 in a 1568x700 window).
- Observed: the panel had scrolled after the fault choice, and the click hit "Start the run". Run w-20 started at once: Explore, 30 min, 2 faults, no effector call. The review opened.
- Expected: "Start the run" is the only control that starts a run, and the panel keeps its position when a fault is chosen.
- Workaround: a screenshot before each click in the run panel. Run w-20 stays in the catalog as an accidental run; the planned run is w-21.
- Status: fixed bfc994d
- Fix: Start the run is a sticky footer below the add rows, apart from them. An e2e checks the panel keeps its scroll position and the add button keeps its place across a fault pick. The shift itself did not reproduce in Chromium.

## Simulation notes

Each plant below was built and run in the canvas (tab 770933647). Run ids are the run folders shown in the run list.

**office-tower-hvac** (Office tower HVAC, 15 instances, 23 links, "Ready to run" after the stale fault weight was removed)
- Areas: Site > Plant room; Site > Floors > Floors 1-3 and Floors 4 to 6 (renamed from "Floors 4-6" during the BLK-5-01 workaround; not renamed back).
- Equipment: chiller large x2 (CH-1, CH-2), chiller x1 (CH-3), cooling tower x1, pump x3, motor drive x3, feeder meter x1, ahu x2, weather x1, site calendar x1.
- Fix: the Build panel listed profile "sensor_pathology" weight for chiller_small__cond_temp__drift; "Remove this weight", then Save: "Ready to run. Saved."
- w-11 (Explore, seed 1, 30 min): faults P-1 impeller wear at 00:10, CH-1 condenser fouling at 00:20; effector VFM-1 set running 0 at 00:15, applied (372 ms). Review: scrubber to the effector minute, KPIs, outcome read.
- w-10 (Explore, accidental, see BLK-5-02): P-1 impeller wear at 00:00 only; no effector call.

**water-pumping-station** (11 instances, 11 links)
- Areas: Site; Pump house; Reservoir; Intake. Equipment: pump x3 (CP-1..3), motor drive x3, control valve x2 (PCV-1, PCV-2), storage tank x2 (STW-1, STW-2), feeder meter x1.
- w-12 (Explore, 30 min): faults CP-1 impeller wear at 00:05, CP-2 strainer blockage at 00:12; effector CP-3 stop at 00:18, applied (371 ms). Review: scrubber, KPIs, outcome read.
- w-13 (Evaluate, same faults and effector re-added): every effector call accepted; review says the truth stayed sealed while it ran and was revealed when it ended.

**data-centre-cooling** (11 instances, 10 links)
- Areas: Site; Plant; Hall A; Hall B; Power. Equipment: weather x1, chiller x2 (WCC-1, WCC-2), cooling tower x1 (CT-1), pump x2 (CP-1, CP-2), ahu x2, transformer x1, feeder meter x1, standby generator x1.
- w-16 (Explore, 30 min): faults WCC-1 condenser fouling at 00:08, CT-1 fan failure at 00:16; effector CP-2 stop at 00:12, applied (403 ms). Review: scrubber, KPIs, outcome read.
- w-18 (Evaluate, same faults, effector CP-2 stop at 00:12): accepted; review shows the sealed-then-revealed note, effector applied (403 ms) and KPIs.

**packaging-factory** (Packaging factory, 13 instances, 11 links; created from "Factory utilities" as an empty plant)
- Areas: Factory (site) > Utilities; Factory > Production hall > Line 1 and Line 2; Factory > Warehouse. The site is named "Factory", inherited from the start plant (the brief's "Site" label).
- Equipment: production line x2 (PL-1 in Line 1, PL-2 in Line 2), packaging line x1 (PK-1, Warehouse), belt conveyor x2 (CV-1, CV-2; each has its motor, so the belt conveyor kind is the conveyor with its drive), motor drive x2 (VFM-1, VFM-2, one per conveyor, torque out to load in), oil-injected screw air compressor x1 (CMP-1), heat-reactivated desiccant air dryer x1 (DRY-1), compressed-air receiver and header x1 (AR-1), feeder meter x1 (FM-1), weather x1 (W-1), site calendar x1 (SC-1).
- Links (11): CMP-1 air out to DRY-1 air in; DRY-1 air out to AR-1 air in; DRY-1 dew point out to PL-1 dew point in; CMP-1 power out, PL-1 power out, PL-2 power out to FM-1 loads in; SC-1 shift out to PL-1, PL-2 and PK-1 shift in; CV-1 torque out to VFM-1 load in; CV-2 torque out to VFM-2 load in.
- w-20 (Explore, accidental, see BLK-5-04): CMP-1 fails to load at 00:10, PL-1 product jam at 00:20; no effector call.
- w-21 (Explore, 30 min, seed 1): the same two faults plus effector VFM-1 set running 0 at 00:15, accepted; review shows "applied · 297 ms", disturbances, KPIs.
- Compressed air is only partly connected: AR-1 air out only accepts DRY-1 (the air-in pins of PL-1, PL-2 and PK-1 are not offered as targets for AR-1 air out), so the lines have no air link (see Could not do).
- Opened from the workspace after the build: the plant opens, and its runs are listed (w-20, w-21).

Runs made in this round, in order: w-10, w-11 (office-tower-hvac); w-12, w-13 (water-pumping-station); w-16, w-18 (data-centre-cooling); w-20, w-21 (packaging-factory). That is 8 runs, including the two accidental runs w-10 and w-20. Runs w-14, w-15, w-17 and w-19 belong to the hunter (hunt-special-names).

Plants I did not touch: compressed-air-works and the hunt-* plants.


## Could not do

- Setting an area on an equipment or instance Area select was not possible by typing the full label or by clicking an option (BLK-5-03). I used the letter-cycling workaround; there is no UI control I could use that is reliable through this tool.
- No "Run again" control was found in the review; "Set up another run" prefills the seed, duration, mode and faults but not the effector calls, so each Evaluate run needed its effector call added again.
- Areas cannot be moved: an area created under "Top level" that should have been inside Factory had to be deleted and re-added (Utilities in Packaging factory).
- Air from the receiver (AR-1 air out) can only be linked to DRY-1 air in; the production lines and packaging line air inputs were not offered as targets, so the compressed-air supply to the lines is not wired. I did not find the pin kind that would allow it.
- The Floors 4-6 area in Office tower HVAC is named "Floors 4 to 6" (the BLK-5-01 workaround); I did not rename it back.
- I did not look for, or use, a way to delete the accidental runs w-10 and w-20 (not in the brief, and deletion needs your approval).

