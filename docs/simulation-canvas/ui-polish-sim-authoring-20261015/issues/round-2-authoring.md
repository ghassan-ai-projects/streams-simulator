# Round 2 - simulation authoring (UI comprehension test)

Result: neither simulation met the brief. `office-tower-hvac` is saved but cannot be composed, so it has no run. `compressed-air-works` composes and was run, but with one compressor instead of two, and without the area tree. The main blocker is one rule: two instances of the same equipment kind cannot be composed in one plant (ISS-2-01), which also breaks the brief's "two or more" requirements.

Path: `http://127.0.0.1:8765`, single tab `seed`, viewport 1440x900 (reset to desktop at the end). Round 1 issues ISS-1-01 to ISS-1-05 were checked as fixed (New plant button, UI-based empty-state steps, Theme label, workspace name) and are not re-reported. Not paired (no pairing screen).

Issue counts: High 3, Medium 9, Low 6 (18 issues total).

## ISS-2-01 Two instances of the same equipment kind cannot be composed
- Severity: High
- Kind: not-working
- Where: Editor (`Edit plant` on the Home card) > Build panel. Status line at top of the graph: "Saved, not composed: 1 issue". Build panel block "ISSUES FROM THE BINARY".
- Steps:
  1. Home > New plant > Name "Pump test", keep Id, Start from "Factory utilities", "An empty plant...", Create and open the editor.
  2. Build: Equipment "Centrifugal pump", Name "pump-a", Area "Site" (any), Add to the plant.
  3. Build again: Equipment "Centrifugal pump", Name "pump-b", Add to the plant.
  4. Click "Save and compose (Ctrl+S)".
  5. Read the status line and the Build panel.
  6. Repeat with two "Air-handling unit", two "Water-cooled chiller" (sizes Medium and Large), two "Oil-injected screw air compressor" (in Compressed-air works), two "Production line".
- Observed: The status reads "Saved, not composed: 1 issue". The Build panel shows `kind name "pump" is used by another entry` (for compressors: `kind name "air_compressor" is used by another entry`). Only one issue is shown per save; after fixing it, another duplicate kind is named. Choosing a different Size (Medium vs Large chiller) does not avoid it. Deleting the second instance removes the clash for that kind only. In Compressed-air works, one compressor composed (28 channels, 143 faults, 13 effectors); two compressors did not. Office tower HVAC, with one of each kind, does not compose either (see ISS-2-08 and the simulation notes).
- Expected: Several instances of one kind (for example 2 AHUs, 2 chillers, 3 pumps) can be placed and composed. If a plant needs one definition for several instances, the UI should offer a quantity or "copies" field.
- Suggestion: Allow repeated instances of one library kind; if not, say in the Build panel that each kind can appear once, before the user adds the second one.
- Status: fixed e7f0a7e
- Fix: Adding a second instance of a component joins the entry with the same size and area, or gets its own kind name (`as`, e.g. `chiller_2`) when size or area differ; moving one of several instances to another area does the same. Unit and e2e (`repeated-equipment.spec.ts`: two pumps, two chillers of different sizes, two AHUs compose) tests added.

## ISS-2-03 Spaces in instance names are accepted, then the save fails with 42 regex errors
- Severity: High
- Kind: not-working
- Where: Build panel "Name" field; inspector "Name" textbox with "Rename" button; status "Saved, not composed: 42 issues"; per-instance "Issues on <name>" box.
- Steps:
  1. Build: Equipment "Water-cooled chiller", Name "Chiller CH-1" (with a space), Area "Chiller hall", Add to the plant. No error appears.
  2. Add more instances with names such as "Cooling tower CT-1" and "AHU-1 Floors 1-3".
  3. Click "Save and compose (Ctrl+S)".
  4. Read the status and the issue text.
- Observed: The add succeeds with no message. After save the status shows "Saved, not composed: 42 issues". The text on each instance reads `string does not match pattern ^[A-Za-z0-9][A-Za-z0-9._:/-]{0,255}\.[a-z][a-z0-9_]{0,47}$ (×8)` and `"Chiller CH-1" is not an instance name: a letter or digit, then letters, digits and . _ : / -, with no spaces`. Renaming each instance to a hyphenated id ("chiller-ch1") fixed it, one instance at a time.
- Expected: Names with spaces are refused at Add or Rename, with a plain message such as "Use letters, digits and hyphens, no spaces", and the field says what a valid name looks like.
- Suggestion: Validate the name in the field as the user types and show the allowed pattern in plain words.
- Status: fixed f7f75bb
- Fix: The Name field shows the rule text read from the composition contract (`x-expected` of the names schema, the one the binary uses), flags a name that rule rejects with that text and refuses Add and Rename; the pattern is read from the contract, not hand-coded.

## ISS-2-05 Areas cannot be created, and there is no area tree
- Severity: High
- Kind: not-working
- Where: Build panel "Area" select; inspector "Area" select; Outline group headers.
- Steps:
  1. Home > New plant, Start from "Office tower HVAC", "An empty plant...", create.
  2. Look for a way to add an area, rename one, or nest one under another (no button or menu on the Area control, the Outline, or the Build panel).
  3. Open the Area select on any instance.
  4. Repeat with Start from "Factory utilities".
- Observed: Area is a single-level select with fixed choices. Office tower HVAC: "Utilities", "Chiller hall", "Office block". Compressed-air works: "Site", "Production hall", "Compressor room", "Chiller plant", "Substation". Nothing is nested, and no area can be added or renamed. The brief's "Site > Plant room", "Site > Floors (two groups)", "Site > Compressor house" and "Site > Production hall A and B" could not be built.
- Expected: A place to add areas under a site, with a visible tree (Site > child areas), so that several halls or floors can be named.
- Suggestion: Add "New area" under a parent in the Outline and in the Area select; show the tree in the Outline.
- Status: fixed ea73d4c
- Fix: New "Add area" form (name and parent select) in the Build panel; the Outline shows areas nested under their parent, including empty ones in the editor; the Area selects list areas as paths (Site › Plant room). Unit and e2e (`areas.spec.ts`) tests added. Names are validated by the binary on save.

## ISS-2-02 "Add to the plant" has no quantity, but the Home text describes several instances per definition
- Severity: Medium
- Kind: unclear
- Where: Home card text "Instance definitions are entries in the plant file. One definition can expand into several instances when you open the plant." Build panel "Add to the plant".
- Steps:
  1. Read the Home line above the Plants list.
  2. Open the editor and look in the Build panel for a count, quantity or copies field.
- Observed: There is no count control. Each "Add to the plant" creates a new definition, and a second one of the same kind triggers ISS-2-01. The sentence implies an expansion step that the UI never shows, and the plant view does not say where it happens.
- Expected: Either a count field where the user says "3 pumps", or no sentence about expansion.
- Suggestion: Add a quantity field to the Build panel, or remove the sentence.
- Status: fixed 0f6b4ad
- Fix: Removed the Home sentence about expansion; each Add is one instance, and a repeated kind no longer needs any extra step.

## ISS-2-04 The "Name" field is the instance id, with a hidden default value that typing appends to
- Severity: Medium
- Kind: unclear
- Where: Build panel "Name" textbox (placeholder text is not shown), inspector "Name" textbox with "Rename".
- Steps:
  1. Build: Equipment "Water-cooled chiller". Look at the Name field before typing.
  2. Click the Name field and type "Chiller CH-1" (no select-all).
  3. Read the field value (or the instance name after Add).
  4. Repeat with Equipment "Centrifugal pump" after another pump was added.
- Observed: The field is not empty. It holds a generated id such as "WCC-1" (chiller) or "chw-pump-p-1" (pump). Typing appends to it, so the instance became "WCC-1Chiller CH-1" and "chw-pump-p-1chw-pump-p2". Nothing says the field is an id or that the value is a suggestion.
- Expected: An empty field with a placeholder, or a visible label "Id (used in files)", as the New plant form already does for its Id.
- Suggestion: Show the generated id as placeholder text, not as a value; label the field as the instance id.
- Status: fixed f7f75bb
- Fix: The Build Name field starts empty with the suggested id as placeholder, and says it is used in files and links and may be left empty.

## ISS-2-08 Composition messages are jargon, show one at a time, do not name the instance, and stay after the cause is removed
- Severity: Medium
- Kind: unclear
- Where: Build panel "ISSUES FROM THE BINARY" block; status "Saved, not composed: N issue(s)"; per-instance "Issues on <name>".
- Steps:
  1. Office tower HVAC: with one Air-handling unit, one chiller and one pump (any of these), save.
  2. Read the Build panel block: "array shorter than minItems 3" and "no array item matches contains".
  3. Add a second pump, save; the block now shows one kind message.
  4. Delete the second pump, save; the block shows a different kind message.
- Observed: Messages use schema words ("minItems", "contains", "kind name ... used by another entry"). The block lists only the first problem each time, and does not say which instance or what to add. The "array shorter than minItems 3" and "no array item matches contains" messages gave no hint which items to add. Clicking the kind message selects one instance, but the names can be stale (the message still named "pump" after the second pump was deleted, until the next save).
- Expected: One plain sentence per problem, naming the instance and what to change ("Add at least 3 ...", "The plant needs a ... to run"), and all problems listed.
- Suggestion: List every problem with the instance name and a plain fix; refresh the list on every save.
- Status: fixed ed6153d, c0a0445 (the kind-name clash is reported one at a time: still the binary's, see ISS-2-01)
- Fix: Identical issues are listed once with a count, each names its instance or, for plant-wide ones, the file location; a stale list says "from the last save". In the binary (c0a0445): `x-expected` now also words a minItems or contains failure, and the domain-spec schema (`contracts/` and `internal/schemas`, equal) gives the profiles array and each mandatory profile a sentence, so the two messages read "needs at least 3 profiles: add nominal, correlated_cascade and sensor_pathology" and "needs a profile named nominal" (and likewise for correlated_cascade and sensor_pathology).

## ISS-2-09 "Saved" is shown while the plant is "Not composed", which reads as done
- Severity: Medium
- Kind: unclear
- Where: Editor status line ("Saved, not composed: 1 issue"); Home plant card ("Not composed yet: edit and save it to run it").
- Steps:
  1. Edit a plant so that it has a problem (see ISS-2-01), click "Save and compose".
  2. Read the status line and then go Home.
  3. Read the Home card.
- Observed: The word "Saved" comes first, so the user thinks the plant is ready. The Home card then says "Not composed yet: edit and save it to run it", which tells the user to save again, although they just did. The Run button is not offered anywhere until the plant composes.
- Expected: "Not runnable: 1 issue to fix" as the state, and the Home card should say the same thing, with the issues count.
- Suggestion: Use one state word ("Needs fixing", "Ready to run") with the saved state as a separate note.
- Status: fixed 15aecca
- Fix: The status reads "Needs fixing: N issues. Your changes are saved." or "Ready to run. Composed ... Saved."; the Home card for a plant with no domain reads "Not runnable yet: open Edit, save, and fix any issue it lists."

## ISS-2-10 Link form resets the target when the output changes, and offers no target without a reason
- Severity: Medium
- Kind: unclear
- Where: Inspector "Link an output" form: "Output", "To", "Input" selects, "Link" button.
- Steps:
  1. Select Chilled water pump P-1. Set Output "flow out", To "Chiller CH-1", Input "flow in". Link.
  2. Select Plant power meter (feeder meter). Set Output "power out".
  3. Read the To list. Then set To "VFD for P-1" and Input "loads in" (to link meter power out to the drive).
  4. Separately: set To first, then change Output: the To value reverts to "Choose…".
- Observed: In step 2 the To list is empty (the drive "VFD for P-1", which has an unconnected Power input, is not offered), with no explanation. Changing Output after To clears To without a message. Other outputs (pump flow, AHU load) do show targets. The "Input" list appears only after To is set, which is not stated.
- Expected: The To list shows every instance with a compatible input, or says why none is available ("no instance takes power here"). Changing Output keeps To if it is still valid.
- Suggestion: Keep To when still valid; show "No compatible inputs" in the To list when empty.
- Status: fixed 4d34920
- Fix: Changing Output keeps To (and Input) while still valid; an empty To or Input list says why ("Choose an output first", "No instance takes heat yet", "Choose an instance first"). The VFD was correctly not offered: its input takes load torque, not electrical power.

## ISS-2-11 Deleting an instance removes its links without a warning or a count
- Severity: Medium
- Kind: unclear
- Where: Inspector "Delete <name>" button.
- Steps:
  1. Select an instance with several links (for example Chiller CH-2 with 3 links).
  2. Click "Delete chiller-ch2".
  3. Read the page; check the Outline and Links counts on the other instances.
- Observed: The instance is removed at once. No confirmation, no count of links lost, no undo message. Its links disappeared from the neighbours' link lists. The header Undo button was not tested.
- Expected: A confirmation that says "Delete chiller-ch2 and its 3 links?", or an undo toast.
- Suggestion: Confirm deletes of instances with links; show the link count.
- Status: fixed faa6557
- Fix: Delete on an instance with links turns into "Delete <id> and its N links?" and needs a second press (blur cancels); an instance without links is deleted at once; Undo is available.

## ISS-2-12 Heat-reactivated air dryer has no compressed-air outlet, so air cannot pass through it
- Severity: Medium
- Kind: unclear
- Where: Plant graph node "air-dryer-1 · air dryer" (ports "air in", "dew point out", "power out"); Link form "Output" list for the dryer.
- Steps:
  1. Compressed-air works: add "Heat-reactivated desiccant air dryer", select it, read the Output list.
  2. Try to link the dryer's air to the receiver.
- Observed: Outputs are "dew point out" and "power out" only. The dryer takes compressed air but does not give it on. The air path therefore ends at the dryer, and I had to connect the compressor directly to the receiver.
- Expected: Either an air outlet, or a note on the component ("this component ends the air path") so the user knows the topology is intended.
- Suggestion: Add an air outlet to the dryer, or show a short note on the component about what it does not pass on.
- Status: fixed 641343b
- Fix: The built-in air dryer gains an `outlet_flow` state (inlet flow less purge flow, an F1 sum declared in data) and an `air_out` provides port of type `compressed_air`, so compressor -> dryer -> receiver composes. The factory-utilities example domain is composed again (its dryer digest changes). Existing plants that ended the path at the dryer still compose unchanged.

## ISS-2-13 No sensors can be added; "channels", "instruments" and "observations" are used for the same thing
- Severity: Medium
- Kind: unclear
- Where: Plant view inspector "CHANNELS 4"; Run page "11 224 observations · opened once every instrument reported"; Editor status "Saved. Composed 28 channels, 143 faults, 13 effectors."
- Steps:
  1. Open Compressed-air works in the plant view and select receiver-header-1.
  2. Look for an "Add sensor" or "Instruments" control in the Build panel or the inspector.
  3. Open a run and read the observation counts and the "Inspect" tab.
- Observed: No sensor or instrument can be added. Sensors are the channels generated from each component's ports (for example "header pressure · kPa", "outlet flow · m³/h"). The words "channel", "instrument" and "observation" are all used, and the brief's "sensors on temperatures, flows, power and run status" cannot be set up as such.
- Expected: One term for measured values, with a place to see and choose them, or a sentence saying that every port is measured.
- Suggestion: Pick one term (channel or sensor) and use it on every screen; show where channels come from.
- Status: fixed b822b52
- Fix: "Channel" is the one term: the run header says "opened once every channel reported", the plant hint explains channels are the values measured on each port and state, and the Build panel says sensors are not added by hand because every port is measured.

## ISS-2-14 Effectors are listed but cannot be called from the UI
- Severity: Medium
- Kind: improvement
- Where: Plant view inspector "EFFECTORS 2" (receiver-header-1: "drain", "set drain isolation"); Run page "EFFECTOR CALLS: None".
- Steps:
  1. Plant view > select receiver-header-1; read the EFFECTORS section.
  2. Click "drain" and "set drain isolation"; look for a control, form or "Call" button.
  3. In Run > Run plant, look for a way to schedule an effector call.
- Observed: The effectors are plain text with no action. The Run form has faults only. The run report says "EFFECTOR CALLS: None" and gives no way to add one. Start/stop of pumps, fans or compressors is therefore not testable from the UI.
- Expected: A way to call an effector (a button in the inspector, and a "Schedule effector call" row in the run form), or a note that effectors are only callable by a consumer.
- Suggestion: Add an "Effector calls" section to the run form, or label effectors as "consumer only".
- Status: binary done 1bc4389 (canvas side open)
- Fix: The bridge now registers `effector.list` {world_id} and `effector.invoke` {world_id, entity_id, effector, args?, command_id?}. They call the director's new tools `sim.world.effectors` and `sim.world.invoke`, which resolve the world's own operator view and use its existing `ListEffectors` and `Invoke` (one implementation; the capability token never leaves the director). `effector.list` returns `{world_id, effectors: [{name, args_schema, risk_class}], simulated}`; names are `<kind>__<effector>` (for example `chw_pump__start`) and `args_schema` is the JSON Schema of the arguments (usually `<kind>_id` plus the set-point). `effector.invoke` returns the operator acknowledgement `{accepted, simulated, world_id, command_id, effect_eta_ns?, reason?}`: `accepted: false` with `reason` (`effector_refused`, `unknown_effector`, ...) is a refusal, `effect_eta_ns` is when the effect lands on the world clock. The call applies at the world's current clock, so the canvas advances the clock to the scheduled minute first; an omitted `command_id` is numbered `director-N` per world.

## ISS-2-20 Empty "What this component models" heading in the plant inspector
- Severity: Low
- Kind: unclear
- Where: Plant view inspector for any component (receiver-header-1): "What this component models" disclosure.
- Steps:
  1. Plant view > select receiver-header-1.
  2. Click "What this component models".
- Observed: The section is shown but no text appears under it. It is not clear whether the text is still missing or the section is broken.
- Expected: A short description of the component, or no disclosure at all.
- Suggestion: Hide the disclosure when the component has no description.
- Status: won't-fix not reproducible
- Fix: In the paired Chrome tab at 1440 wide the disclosure for receiver-header-1 opens and shows the library description (checked: paragraph visible, 87 px high, readable colour); the disclosure is already hidden when a component has no description.

## ISS-2-07 "Details panel" button hides the panel titled "Build"
- Severity: Low
- Kind: unclear
- Where: Editor header button "Details" (accessible name "Details panel"); right-hand panel heading "Build".
- Steps:
  1. Open a plant in the editor with nothing selected (Build panel shown).
  2. Click "Details" in the header.
  3. Click it again.
- Observed: The first click hides the Build panel (the right side is empty). The second click shows Build again. The label "Details" does not say it toggles Build, and when an instance is selected the same panel shows the instance's details.
- Expected: The label says what it toggles ("Build panel" or "Inspector").
- Suggestion: Rename the header buttons to name the panels they show.
- Status: fixed 37e64ed
- Fix: The right-panel header button is named "Build" in the editor and "Inspector" on the plant and run screens, with its title and keyboard hint unchanged.

## ISS-2-06 Default Area and the Area list differ between plants with no explanation
- Severity: Low
- Kind: unclear
- Where: Build panel "Area" select (default value); inspector "Area" select.
- Steps:
  1. Office tower HVAC: open the Build panel; note the default Area "Utilities" and the choices "Utilities", "Chiller hall", "Office block".
  2. Compressed-air works (Start from Factory utilities, empty): note the default "Site" and the choices "Site", "Production hall", "Compressor room", "Chiller plant", "Substation".
- Observed: The default and the choices come from the plant's starting data, with no label saying so. After adding one instance the default changes to the last area used in some cases.
- Expected: The default is the first area on the site, and the list says where the areas come from.
- Suggestion: Show the area list's source ("from Factory utilities") or show a site root that always exists.
- Status: fixed fa58ee9
- Fix: The default Area is the first area of the tree (the site), and a line under the select says the areas are the plant's own and where to add more.

## ISS-2-15 Escape does not close the Legend popover; it covers the Outline until the Legend button is clicked
- Severity: Low
- Kind: not-working
- Where: Editor header "Legend" button; popover "Wire colour and pin shape by port type".
- Steps:
  1. Click "Legend" in the editor header.
  2. Press Escape.
  3. Click on the Outline row for another instance.
- Observed: Escape does not close the popover. The first click on the Outline or the graph closes it without selecting, so the next click is needed.
- Expected: Escape closes the popover, as it clears the selection elsewhere in the editor.
- Suggestion: Close the legend on Escape and on outside click, and pass the click through.
- Status: fixed fcbe8a5
- Fix: Escape closes the Legend wherever focus is, and a press outside closes it and still reaches its target. Unit test added.

## ISS-2-16 Evaluate-mode report says "The report has no scoring section" without saying why
- Severity: Low
- Kind: unclear
- Where: Run page, Report tab, bottom section "SCORING"; Run form "Evaluate" radio ("The truth stays sealed while it runs, so a consumer can be scored").
- Steps:
  1. Run > Run plant > choose "Evaluate", add a fault, Start the run.
  2. Read the report's SCORING section.
- Observed: "The report has no scoring section." No consumer is attached to the plant, so there is nothing to score, and the report does not say so. The user is told the truth was sealed during the run, but cannot see what it was scored against.
- Expected: "No consumer attached: nothing to score" as the reason.
- Suggestion: Say why there is no score (no consumer attached) and where to attach one.
- Status: fixed 1ea87f6
- Fix: The scoring section says "Nothing was scored: no consumer verdict was recorded for this run. A consumer attaches through the operator endpoint; the canvas does not attach one."

## ISS-2-17 Plant card says "instance definitions" while the plant view says "instances"
- Severity: Low
- Kind: unclear
- Where: Home plant card "7 instance definitions · 3 runs"; plant view header "7 instances · 11 links".
- Steps:
  1. Home: read the card under "Compressed-air works".
  2. Open the plant; read the "7 instances" line.
- Observed: The same seven items are called "definitions" on Home and "instances" in the plant view. The word "definition" is not used anywhere else.
- Expected: One word for the same thing.
- Suggestion: Use "instances" on both screens.
- Status: binary done b4c6b24 (canvas label open)
- Fix: The listing field `instances` now sums the names of each set (or `count.default` for an id_template set) in a domain's `instances` and in a composition's `equipment`, so it matches the plant view; the canvas can label it "instances". The expansion sentence was removed (ISS-2-02).

## Worked well
- The New plant form (round 1 fix) now asks "Start from" and "What to keep" in plain words; the Id is filled from the name as you type.
- The Link form filters the Input list to compatible inputs ("heat in", "flow in", "air in") and shows "Linked heat out to ... heat in" after saving a link.
- The Outline groups instances by area and clicking an issue in the Build panel selects the instance it names.
- The run form shows "The same seed, plant, faults and simulator version give the same run" and "Run again" reopens the form with the same seed and faults. Two runs with the same seed and fault gave identical true KPIs (w-2 and w-3).
- The run view has a clear scrubber: Play, speed buttons, "Next event", fault markers on the timeline, and "Verified: the files match and the replay reproduces."
- Inspect shows each instance's observed values at the playback time with a time stamp per channel.

## Simulation notes

### office-tower-hvac (title "Office tower HVAC") - not completed, saved but not composed
- Area tree (flat, from the plant's starting data): Utilities; Chiller hall; Office block. No nesting.
- Equipment saved at the end (kind: count):
  - Water-cooled chiller: 1 (chiller-ch1, medium). Planned 2; chiller-ch2 (large) and a re-added medium one both failed (ISS-2-01).
  - Cooling tower: 1 (cooling-tower-ct1).
  - Centrifugal pump: 1 (chw-pump-p1). Planned 3 (P-1 chilled, P-2 chilled, P-3 condenser); the second and third failed (ISS-2-01).
  - Variable-frequency motor drive: 1 (vfd-p1).
  - Feeder meter: 1 (plant-power-meter).
  - Air-handling unit: 1 (ahu-floors-1-3). Planned 2; the second (ahu-floors-4-6) failed (ISS-2-01).
  - Weather: 1 (site-weather).
  - Site calendar: 1 (site-calendar).
- Instruments: none can be added (ISS-2-13). Channels come from ports.
- Links at the last save: 15 (for example chiller-ch1 supply out -> ahu-floors-1-3 chw in; weather dry bulb out -> ahu outdoor in; site calendar tariff out -> chiller tariff in; pump flow out -> chiller flow in; ahu load out -> chiller load in; cooling tower basin out is not linked).
- Compose history: (a) 42 issues from names with spaces (ISS-2-03); (b) "kind name chiller is used by another entry" with the large chiller; (c) "kind name pump" with the second and third pumps; (d) "kind name ahu" with two AHUs; (e) with one AHU and one of each other kind: "array shorter than minItems 3" and "no array item matches contains" (ISS-2-08). Not resolved.
- Actions exercised: none (the plant does not compose, so there is no run). Home card: "Not composed yet: edit and save it to run it."

### compressed-air-works (title "Compressed-air works") - composed and run, short of the brief
- Start from "Factory utilities", "An empty plant" (site, areas, profiles and axes).
- Area tree (flat): Site (site-calendar); Production hall (production-line-a, packaging-line-b); Compressor room (compressor-1, air-dryer-1, receiver-header-1, power-meter-air). "Production hall A and B" and "Compressor house" could not be created (ISS-2-05).
- Equipment: Oil-injected screw air compressor 1 (compressor-1, medium 75 kW). Planned 2; the second failed (ISS-2-01). Heat-reactivated desiccant air dryer 1. Compressed-air receiver and header 1. Feeder meter 1 (power-meter-air). Production line 1 (production-line-a). Packaging line 1 (packaging-line-b). Site calendar 1.
- Links at the last save: 11 (compressor air out -> air-dryer-1 air in; compressor air out -> receiver air in; receiver pressure out -> production-line-a air in; receiver pressure out -> packaging-line-b air in; production-line-a air out and packaging-line-b air out -> receiver demand in; compressor, dryer, production-line-a and packaging-line-b power out -> power-meter-air loads in; site calendar shift out -> production-line-a shift in; the calendar shift out -> packaging-line-b shift in is not linked, and the calendar tariff out is not linked).
- Compose: "Saved. Composed 28 channels, 143 faults, 13 effectors." With a second compressor: "kind name air_compressor is used by another entry" (ISS-2-01); I removed the second compressor before the run.
- Runs (all Verified):
  - w-1, Explore, seed 1, 1 h, faults: compressor-1 "fails to load" at 00:10:00; power-meter-air "load surge" at 00:20:00. Played at 10x, used Next event, Pause, opened Inspect for compressor-1 (channels: air flow, run status, load status, alarm state, discharge temp, power, vibration).
  - w-2, Evaluate, seed 59776 (picked with the dice button), 30 min, fault receiver-header-1 "header leak" at 00:15:00. Truth sealed during run, revealed at end. "EFFECTOR CALLS: None".
  - w-3, "Run again" from w-2: same seed, same fault, same mode; true KPIs identical to w-2.
- Effectors listed on receiver-header-1: "drain", "set drain isolation" (not callable, ISS-2-14).
- Home card afterwards: "2 plants · 3 runs"; the plant opens from the Home card ("Open"), and the three runs are listed under Runs.

## Could not do
- Two or more instances of the same equipment kind in one plant (two AHUs, two chillers, three pumps, two compressors, two production lines of one kind). Composition refuses them (ISS-2-01). Office tower HVAC and Compressed-air works both fall short of the brief because of this.
- Create areas, nest areas under "Site", or name "Site > Plant room", "Site > Floors" (two floor groups), "Site > Compressor house" and "Site > Production hall A and B" (ISS-2-05).
- Add explicit sensors or instruments (pressure, flow, temperature, power, run status). Only the channels generated from ports exist (ISS-2-13).
- Call effectors (start/stop of pumps, fans, compressors, valves) from the UI. Effectors are listed but there is no action control (ISS-2-14).
- Set a quantity or count for an equipment definition (ISS-2-02).
- Pass compressed air through the air dryer (no air outlet, ISS-2-12).
- Score an Evaluate run: no consumer is attached from the UI, so there is no scoring (ISS-2-16).
