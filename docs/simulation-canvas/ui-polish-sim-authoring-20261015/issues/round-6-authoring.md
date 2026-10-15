# Round 6 authoring report

Author: simulation builder (brief-author). Browser: Chrome tab 770933647 at http://127.0.0.1:8766 (window resized to 1440x900 at the start; the pane later reported 1280x657, which I did not change).
Plants completed: `boiler-steam-loop`, `greenhouse-ventilation`, `food-cold-chain` (in that order).
Plants named `hunt-*` and the five existing plants were not touched. The director's tab was not used.

## Simulation notes

### boiler-steam-loop - "Boiler steam loop"

- Area tree: Site > Boiler house > Gas train (empty); Site > Distribution; Site > Process hall > Heat users (empty). The Gas train and Heat users areas exist only to give a second area level.
- Equipment (14 instances): gas-fired steam boiler x1 (boiler-1); water-to-water heat exchanger x2 (hx-1, hx-2); centrifugal pump x3 (pump-feed, pump-circ, pump-return); variable-frequency motor drive x2 (drive-feed, drive-circ); storage tank with level control x1 (tank-cond); pneumatic control valve x2 (valve-steam, valve-process); feeder meter x1 (meter-gas); weather x1; site calendar x1.
- Channels: one per instance port, as the plant view lists them.
- Links: 10 (pressure to both valves, pump-feed liquid to tank-cond, tank level to pump-circ, three power links to meter-gas, pump-circ flow to hx-2 hot flow, hx-1 hot out to hx-2 hot in, pump-feed condenser-water to hx-1 cold flow). One pump-feed/tank-cond loop link was removed (see BLK-6-02).
- Site was renamed from the start plant's "Factory" to "Site" through the Areas list.
- Run w-23 (Explore, seed 1, 30 min): faults 00:10 pump-feed strainer blockage, 00:25 drive-feed winding overheat; effector 00:15 pump-circ stop (accepted). Review opened, scrubber moved to 00:15, Verified.
- Run w-24 (Evaluate, same faults, effector re-added by hand because "Set up another run" does not copy it): Verified.

### greenhouse-ventilation - "Greenhouse ventilation"

- Area tree: Site > Plant room; Site > Greenhouse > Bay 1, Bay 2 (Greenhouse itself holds no equipment).
- Equipment (12 instances): air-handling unit x2 (ahu-1 in Bay 1, ahu-2 in Bay 2); water-cooled chiller x1 (chiller-1) and water-to-water heat exchanger x1 (hx-cool) for cooling; centrifugal pump x2 (pump-chw, pump-cw); variable-frequency motor drive x2 (drive-chw, drive-cw); feeder meter x2 (meter-power in Plant room, meter-bay1 in Bay 1); weather x1; site calendar x1.
- Links: 12 (chilled-water supply to both AHUs, weather dry bulb and calendar shift to both AHUs, pump-chw flow to chiller, chiller and both drives and meter-bay1 power to meter-power, pump-cw flow to hx-cool hot flow). No cycle was created.
- Run w-26 (Explore, seed 1, 30 min): faults 00:08 ahu-1 coil fouling, 00:20 pump-chw strainer blockage; effector 00:14 drive-cw set running 0 (accepted). Review opened, scrubber moved, Verified.
- Run w-27 (Evaluate, same faults, effector re-added by hand): Verified.

### food-cold-chain - "Food cold chain"

- Area tree: Site > Refrigeration plant; Site > Cold rooms > Room A, Room B; Site > Power.
- Equipment (13 instances): water-cooled chiller x2 (chiller-1, chiller-2); cooling tower x1; centrifugal pump x2 (pump-cond, pump-chw); variable-frequency motor drive x1 (drive-chw); storage tank with level control x1 (tank-chw); oil-injected screw air compressor x1 (compressor); compressed-air receiver and header x1 (receiver); air-handling unit x2 (ahu-a in Room A, ahu-b in Room B); feeder meter x1 (meter-main); standby generator x1 (gen-1).
- Links: 11 (pump-chw flow to both chillers, chiller-1 and chiller-2 supply to ahu-a and ahu-b, chiller-1 heat to cooling tower, pump-cond condenser-water to cooling tower, chiller-1, chiller-2, drive-chw, compressor power to meter-main, compressor air out to receiver). The standby generator has no link.
- Run w-28 (Explore, seed 1, 30 min): faults 00:06 chiller-2 compressor wear, 00:18 pump-cond seal leak; effector 00:22 compressor set loaded 0 (accepted). Review opened, scrubber moved, Verified.
- Run w-29 (Evaluate, same faults and effector re-added): Verified.

After the runs, the Workspace list shows all three plants with 2 runs each; each plant reopens from its Open button.

## Blockers

### BLK-6-01 Pressing Return in the area Inside select submits the Add area form

- Where: Build panel > ADD AREA > Inside (select), next to Area name and the Add area button.
- Steps: type "Boiler house" in Area name, click Inside, press Down, press Return.
- Observed: Return submitted the Add area form with Inside still "Top level". "Boiler house" was created as a top-level area beside Site, not inside it. I deleted it and re-added it.
- Expected: Return on the select only commits the select's value; it does not submit the form.
- Workaround: click Inside and type "s" to pick "Site" (the single-letter type-ahead worked), then click Add area.
- Status: fixed ac60a3b
- Fix: Return on the Inside select picks the option and no longer submits the Add area form (the same guard the run form uses).


### BLK-6-02 A pump and storage-tank loop blocks "Ready to run" with a binary issue

- Where: Build panel > Issues from the binary, after Save and compose.
- Steps: link pump-feed liquid out to tank-cond inflow in; link tank-cond level out to pump-feed level in; click Save.
- Observed: the issue reads "kinds read each other over links in a cycle (path [pump storage_tank pump]); a cycle needs a sampled read (sample_s) in /kinds/2". The plant is not "Ready to run".
- Expected: either the editor offers a way to supply the sampled read where the loop is made, or the loop is refused at link time with the same explanation.
- Workaround: deleted the tank-cond level out to pump-feed level in link; the plant then saved as "Ready to run". I did not change the domain kinds.
- Status: needs-binary
- Fix: none in the canvas. Missing in the binary: a way to declare the sampled read (sample_s) for a kinds cycle, or refusing the link with that explanation at link time. The canvas cannot supply it.


### BLK-6-03 Typing into the Inside select landed in the Area name field

- Where: Build panel > ADD AREA > Area name and Inside.
- Steps: after a previous add, click the Inside select and type "Site › Cold rooms" (or click it and type "Site" after the Home key).
- Observed: the typed text was appended to Area name ("Room B› Cold rooms"). Pressing Home scrolled the whole Build panel to the top, and the Inside select then showed "Site › Cold rooms › Room A" instead of the intended parent.
- Expected: with the Inside select focused, typed text selects an option; it does not reach the name field.
- Workaround: cleared the name field with cmd+a and Delete, retyped the name, then moved focus with Tab from the name to Inside and typed "Site › Cold rooms". Afterwards every Inside choice was checked with a zoom before Add area.
- Status: not reproduced (no change)
- Fix: on this build, typing with the Inside select focused selects 'Site' and leaves Area name unchanged; the fixer could not reproduce the append. Needs a re-check by the author on a fresh load.

Blockers count: 3 (BLK-6-01 to BLK-6-03). None stopped a plant from being completed; each needed a workaround.

## Could not do

- "Set up another run" does not copy effector calls ("Not copied: 1 effector call"). For each Evaluate run I re-added the effector call by hand.
- A pump and tank cycle could not be modelled with a sampled read: the editor has no control for the sampled read that the binary's cycle issue asks for (see BLK-6-02).
- The New plant form has no site name field. The site is inherited from the start plant (here "Factory") and has to be renamed in the Areas list.
