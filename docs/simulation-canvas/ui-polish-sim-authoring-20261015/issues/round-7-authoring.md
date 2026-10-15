# Round 7 authoring report

Author: simulation builder (brief-author). Browser: Chrome tab 770933647 at http://127.0.0.1:8766 (window resized to 1440x900; the tab was reloaded once at the start). One extra tab (770933659) belongs to another agent and was not used.
Plants planned for this round, in order: `power-distribution-standby`, then `multi-hall-campus`.
Plants named `hunt-*`, `dir-*` and the eight existing plants were not touched.

## Blockers

### BLK-7-01 Options of the editor selects cannot be picked by clicking
- Where: editor Build panel, ADD AREA > Inside; Build panel > Area (equipment); Link an output > Output, To, Input.
- Steps: click the select, then click the option in the list.
- Observed: the native option list does not appear in the screenshot and a click on the option text or on the place where it should be does nothing; the select keeps "Top level" (or the previous value). Pressing Return while a select is focused submits the Add area form with the old value (Substation was added at top level twice and had to be deleted).
- Expected: clicking an option selects it, as with the radio buttons.
- Workaround used: focus the select with Tab from the preceding field, then type the option text (for example "S" for Site, then "S" again to move to the next option that starts with S). Confirm the value with a read of the select before pressing Add. In the Link panel, click the select, type the instance name and press Return; the Return there does not submit a form.
- Status: won't-fix: Return not reproduced; option click not verifiable in automation
- Fix: Return on the focused Inside select does not submit Add area (guard from ac60a3b): checked in Chrome, "Probe" stayed in Area name and no area was added. Typing a letter on the focused select selects its option. The native option list is not visible to the automation, so the click path was not verified; the author should re-check the click on a fresh load.

### BLK-7-02 A motor drive's load input does not take a feeder meter's power
- Where: Link an output, Output "power out" of meter-a (feeder meter), To list.
- Steps: select meter-a, Output = power out, To = vfd-a1.
- Observed: the To list offers tx-sub-1, meter-sub, meter-gen, gen-1, meter-b and does not offer vfd-a1 or vfd-b1 (the motor drive "load in" is not a valid target for the feeder's power output).
- Expected: a feeder can supply a drive, or the list says why a target is missing.
- Workaround used: the drive's "power out" links into the feeder's "loads in" (vfd-a1 power out -> meter-a loads in). Pumps, belt conveyors and production lines also link their power out into the line feeder. No cycle was created.
- Status: needs-binary
- Fix: Missing in the library data, not the canvas: the motor drive (internal/component/internal/files/builtin/equipment/motor-drive.component.json) has no electrical input; its only input is load_in, typed load_torque. So no electrical output can link to it, and the To list is right for that data. Needed: an electrical supply input on the drive in the binary library, or a message in the Link form that says why the target is missing (the second is canvas text, not done here).

## Simulation notes

### power-distribution-standby - "Power distribution with standby"

- Created from "An empty plant" (start-from copy of boiler-steam-loop, no equipment kept); the copied "Factory" area was renamed to "Site".
- Area tree: Site (wx-site, cal-site); Site > Substation (tx-sub-1, meter-sub); Site > Generator yard (gen-1, meter-gen); Site > Production hall (no equipment) > Line A (vfd-a1, pump-a1, belt-a1, line-a, meter-a); Site > Production hall > Line B (vfd-b1, pump-b1, belt-b1, line-b, meter-b).
- Equipment: 16 instances, 9 kinds: distribution transformer x1, standby generator x1, feeder meter x4, variable-frequency motor drive x2, centrifugal pump x2, belt conveyor x2, production line x2, weather x1, site calendar x1. The plant view shows 16 instances in 5 areas, 13 links.
- Links (13, no cycle): gen-1 -> meter-gen; meter-gen -> meter-sub; meter-sub -> tx-sub-1; meter-sub -> meter-a and meter-b; in each line, vfd, pump, belt and line power out -> the line's feeder meter "loads in" (4 per line).
- Runs (seed 1, 30 min, faults at 00:08 tx-sub-1 overload and 00:20 gen-1 fails to start, effector 00:12 vfd-a1 set running 0):
  - w-30, Explore: Verified. Review opened, scrubber moved to 00:14:58, fault markers and the effector marker visible.
  - w-31, Evaluate: same faults and effector, re-added by hand (Run plant opens an empty setup). Verified.

### multi-hall-campus - "Multi-hall campus"

- Created from "An empty plant" as above; "Factory" renamed to "Site".
- Area tree: Site (wx-site, cal-site); Site > Utilities (comp-u1, rcv-u1, ch-u1, ch-u2, ct-u1, pmp-u1, pmp-u2, pmp-u3, mtr-u1); Site > Halls (no equipment) > Hall 1, Hall 2, Hall 3 (each: pl-hN production line, pk-hN packaging line, cv-hN belt conveyor, vfd-hN motor drive, mtr-hN feeder meter).
- Equipment: 26 instances, 12 kinds: oil-injected screw air compressor x1, compressed-air receiver x1, water-cooled chiller x2, cooling tower x1, centrifugal pump x3, feeder meter x4, production line x3, packaging line x3, belt conveyor x3, motor drive x3, weather x1, site calendar x1. The plant view shows 26 instances in 5 areas, 20 links.
- Links (20, no cycle): comp-u1 air out -> rcv-u1 air in; comp-u1, ch-u1 and pmp-u1 power out -> mtr-u1 loads in; pmp-u1 flow out -> ch-u1 flow in; ch-u1 heat out -> ct-u1 heat in; wx-site wet bulb out -> ct-u1 wet bulb in; cal-site shift out -> pl-h1 shift in; per hall, pl, pk, cv and vfd power out -> mtr-hN loads in (12).
- Runs (seed 1, 30 min, faults at 00:10 ch-u2 condenser fouling and 00:22 pmp-u2 seal leak, effector 00:15 pmp-u3 set speed 0.6):
  - w-32, Explore: Verified; scrubber moved to 00:16:32, the ch-u2 fault shown on the graph.
  - w-33, Evaluate: same faults and effector, re-added by hand. Verified.
- Large-graph legibility (from the build, not a hunt list): at the default 30 % zoom the node labels are cut ("cal-si...", "packaging ...", "production ...", "cooling to..."), so the kind and the instance name cannot both be read. The two Site nodes sit in a small box in the top-left corner, and the dotted links that cross the Utilities box into the Hall 1 column are hard to follow. The outline lists every instance, which is the only way to check a link's end at this size.

Workspace list after the runs: both new plants are listed with 2 runs each (power-distribution-standby 16 instances, multi-hall-campus 26 instances). Both reopen from their Open buttons. Hunt plants, dir-review-empty and the eight existing plants were not touched.

## Could not do

- Pick an option in an editor select by clicking it (the native option list is not reachable): areas were nested and instances linked through keyboard typeahead after focusing the select, see BLK-7-01.
- Feed a motor drive's load input from a feeder meter (the library's To list does not offer it), so the drive reports its power into the feeder instead, see BLK-7-02.
- Build a second distribution transformer for plant 1: the brief's "transformer(s)" was met with one (tx-sub-1), which is enough for the fault and the links.
- Test whether "Set up another run" copies faults and effector calls: I used Run plant instead, which opens an empty setup, so each Evaluate run was set up by hand.

Blockers count: 2 (BLK-7-01, BLK-7-02). Simulations completed: power-distribution-standby, multi-hall-campus.
