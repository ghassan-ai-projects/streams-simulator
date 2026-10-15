# Simulator backlog from real-world failure probes

Collected while authoring simulations through the canvas. These are simulator behaviours, not UI defects; they are not
built in the UI polish loop. Verdicts are the tester's reading of the UI and need confirming against the domain data
before any is implemented.

| # | Real-world situation | Observed in the simulator (round 3, compressed-air-works run w-4) | Verdict |
|---|---|---|---|
| B-1 | Compressor "fails to load" and stays unloaded | The fault persists to the end of the run and air flow falls to about zero, but run status and load status keep their pre-fault values and no alarm is raised | partly |
| B-2 | Sensor drifts high | Discharge temperature drifts to 89 C; alarm state stays normal and the ISA-18.2 summary shows no activation | partly |
| B-3 | Reading outside its physical range | Negative air flow (-0.50, -3.50 m3/h) is shown with a warning icon only; no alarm, no reason | partly |
| B-4 | Command sent but not executed, executes late, or is refused | Effectors now exist in the bridge (`effector.list`, `effector.invoke`); a refusal reports `accepted:false` with a reason. Behaviour with delayed or dropped execution still to be probed | to probe |
| B-5 | Feeder or utility trip and the response of dependants | Not yet probed | to probe |
| B-6 | Communication gap, duplicated or out-of-order readings | Not yet probed (the perturbation layer exists; the canvas does not expose it) | to probe |
| B-7 | Stuck or dead sensor (flat-line, frozen timestamp) | Not yet probed | to probe |
| B-8 | Flow and pressure meters as standalone instruments | The library has no standalone flow or pressure meter equipment; measurements are generated per port | missing |
| B-9 | A pump driven by a variable-frequency drive | `streamsim/motor-drive` consumes a `load_torque` input, but no library equipment offers a `load_torque` output: the pump exposes flow, level and electrical power only, so a pump with its drive cannot be linked (author blocker BLK-4-02) | missing (library data) |
| B-10 | Command reported applied on a failed device | On a compressor with "fails to load", effector `set loaded` is reported `applied` and nothing changes (the run keeps its status "on"); the operator view cannot tell | partly (a sealed-truth silent no-effect; the consumer still needs a visible signal) |
| B-11 | Shelving an alarm that is not active | "shelve alarm" is accepted with 0 activations and no refusal | missing (refusal with a reason) |
| B-12 | Stale or drifting sensor | A stale sensor still moves; a drift fault changes about 0.1 C over 17 minutes, too small to see in a short run | partly |
| B-13 | Unconnected equipment | A chiller and a pump with no links run as healthy for 30 minutes; nothing says the plant is incompletely wired | missing (a lint finding) |
| B-14 | Condenser-water feed | A condenser-water pump cannot feed the water-cooled chiller (port types differ) | missing (library data) |
| B-15 | Communication gap | The `dropout` fault freezes the value (21.9 C from 00:05 to past 00:16) and it reads as a normal observation; no staleness or quality flag, no alarm | partly |
| B-16 | Duplicated, late or out-of-order readings | No fault kind produces them in the canvas fault list (the delivery perturbation layer exists in the binary) | missing in the UI path |
| B-17 | Feeder or utility trip | No trip fault and no supply model: a dependent load keeps running | missing |
| B-18 | An alarm that raises | No round-5 run raised an alarm, including with a frozen or drifting sensor | missing (alarm rules per instrument) |
| B-19 | Air consumers read backwards | Production and packaging lines declare `air_out` (compressed_air, provides) that feeds the receiver's `demand_in`; they have no air input. A user expects "air from the receiver to the line" and finds no inlet (author BLK-5, packaging-factory). The model is right (demand flows to the supply); the port names and descriptions do not say so | unclear (port naming/description in the library; same family as ISS-3-10) |

Source: `issues/round-3-authoring.md`, `issues/round-4-hunt.md`, `issues/round-5-hunt.md`, section "Simulator behaviour notes". Later rounds append rows.
