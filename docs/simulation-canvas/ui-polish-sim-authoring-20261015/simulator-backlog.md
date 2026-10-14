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

Source: `issues/round-3-authoring.md`, section "Simulator behaviour notes". Later rounds append rows.
