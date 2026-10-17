# Two end-to-end simulations in real Chrome (after the migration)

Setup: the release binary served `canvas/dist` over a realistic workspace (the e2e fixture builder: five plants, eleven recorded runs including tampered, incomplete and diverged ones). Driven through Claude in Chrome against a fresh server; no mocks.

## Simulation A — Factory utilities, Evaluate, two faults, two effector calls

1. Pairing: a wrong code is rejected with an inline error; the printed code pairs and opens the workspace (5 plants, 11 runs).
2. Run panel: seed 42, 20 minutes, Evaluate, as fast as possible; faults CT-1 fan failure at minute 5 and D-1 prefilter clogging at minute 8; effector calls CH-1 set setpoint at 3 and D-1 set running at 12.
3. Validation: a setpoint of 99 is refused ("setpoint c is a number from 4 to 12", focus returns to the field, "Effector call 1 needs attention before starting"); 7 starts the run.
4. Result: "20 min of simulated time · Evaluate · 2 faults · 2 effector calls. Every effector call was accepted." The top bar shows "Run w-1 finished · Open review".
5. Review (`#/review?world=w-1`): Verified ("The files match and the replay reproduces"); disturbances and effector calls tables with outcomes (applied 251 ms, 323 ms); KPI table; timeline with four marks (3:00, 5:00, 8:00, 12:00). Seeking to the fan failure selects CT-1 (fault tag on the node, inspector tab "Inspect · CT-1", disturbance row highlighted); 600x playback runs to 00:20:00 with both faults shown on the graph.
6. "Set up another run": navigates to the plant with the run panel prefilled (seed, duration, mode, faults; "Not copied: 2 effector calls"). Seed 43, Watch pace: live run with the top-bar indicator (`00:10:40 of 00:20:00 · 60 x per s`), faults appearing on the nodes at their times, progress bar with fault ticks; after the 2 s hold the app opened the review of w-2, Verified.

## Simulation B — a plant built in the editor, then run

1. New plant "North annex" (id derived: `north-annex`), an empty plant from the Boiler template.
2. Editor: areas Pump hall and Cooling yard; Cooling tower CT-A (Cooling yard) and Water-cooled chiller CH-A (Pump hall) from the library (search filter and Enter-to-add); two links created from the instance inspector (CH-A heat out to CT-A heat in, CT-A basin out to CH-A cond water in); undo and redo (2 to 1 to 2 links); Save composes the domain ("Ready to run. Saved."); Open plant.
3. Run: seed 7, 15 minutes, Explore, Watch pace; fault CT-A fan failure at minute 4; effector call CH-A stop at minute 2. Live: CT-A shows "Fault · fan failure" and the top bar counts simulated time. "Stop now" saved the run at 00:11:30 of 00:15:00 ("Every effector call was accepted"); the panel stayed open with "Open the review" because the run has a call.
4. Review (w-3): Explore, 12 min, Verified; CT-A approach and CH-A COP/specific power KPIs; the alarm table. Reload restores the same review from the address; `?` opens the keyboard help with focus on its heading; Space plays; CH-A power drops to 0.40 kW after the stop call; the theme button cycles.
5. Unknown run (`#/review?world=w-does-not-exist`): "Run w-does-not-exist is not in this workspace" with a Back to the workspace button, no crash.

## Findings

- No console errors in the tracked session. The network log listed `/api/runs/w-1/events` as 503, but the Go bridge has no code path that answers 503, and a direct request for the same stream with a paired cookie returns 200 with `text/event-stream` (checked with curl for w-1 and w-evaluate); the 503 is how the Chrome extension's log reports a long-lived stream that the page closed when it navigated, not an application error. The other non-2xx entries are the deliberate unknown-run probe (403 file, 404 stream, 422 report/verify) which the screen turns into 'Run … is not in this workspace'.
- Artifacts persisted by the simulations: `runs/w-1..w-3` (command logs of 24, 33 and 25 commands), `compositions/north-annex.composition.json` and the composed `domains/north-annex.domain.json`.
- Dragging from an output pin to an input pin with scripted mouse coordinates panned the canvas instead of linking (the pin hit area is small; the supported link paths — the inspector form — worked). Not investigated further.
