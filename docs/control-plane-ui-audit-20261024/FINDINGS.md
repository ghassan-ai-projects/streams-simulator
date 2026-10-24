# Findings

The operator's job is to answer: What is happening? Can I trust the readings? Who can act? What needs my decision? Did the action work? What happens if the connection or browser fails? The current UI makes several of these answers depend on which page is open.

## Reading the findings

- **P1:** resolve before relying on the UI for the affected evaluation or critical operating workflow.
- **P2:** substantial operational friction or a missing workflow in the reviewed surfaces.
- **P3:** clarity or consistency improvement.
- **Observed:** directly demonstrated on screen. **Not found:** no usable entry point in the reviewed screens; not proof that the backend lacks the capability.

Evidence references below point to numbered files in [evidence/](evidence/). The [review log](REVIEW_LOG.md) records reproduction and limits. None of the findings asserts a code-level root cause.

## Trust, lifecycle, and recovery

### F01 · P1 · Evaluate reveals the active injected fault

**Observed.** During Evaluate run `w-2`, closing the run panel and selecting the thermal sensor showed `Fault · temperature · dropout` on its node. A separate Evaluate run `w-5` showed `Fault · temperature · drift` while the run panel still said `Evaluate`. Evidence: [19](evidence/19-evaluate-running.jpg), [29](evidence/29-evaluate-repro-setup.txt), [30](evidence/30-evaluate-fault-disclosure.jpg).

This lets a person using the live plant identify the injected cause before sealing. It conflicts with the existing canvas acceptance requirement that Evaluate pages carry no hidden fault label before sealing. The operator/director distinction is not visible in this screen. The test does **not** establish that an external consumer received hidden data.

**Change:** enforce a clear observer view for Evaluate, with no active injected-fault identity before sealing. Any intentionally privileged director workspace needs explicit role and unblinded labeling and must not become the scored observer's view. Show observation-quality evidence without identifying the sealed cause. Test both the visible label and accessible name.

### F02 · P1 · Reload loses the active evaluation and leaves no recovery action

**Observed.** Started `w-2` for 30 minutes in Watch mode, with a fan call at minute 1 and dropout at minute 5. Reloaded around minute 6. The running banner and stop control disappeared. The plant offered Run plant and Review w-2. Review remained `Did not finish`, `run.json is absent`, and six minutes of recording on a later visit after the expected completion time. Evidence: [19–21](evidence/21-evaluate-incomplete-after-reload.txt), [23](evidence/23-reload-still-incomplete.txt).

An ordinary browser refresh breaks the operator's continuity. There is no visible resume, reconnect, recover recording, or explicitly close abandoned run. UI evidence supports an interrupted run; it does not identify the server cancellation mechanism.

**Change:** make active-job state recoverable independently of the page. On reconnect, restore run identity, mode, progress, control availability, and lifecycle. If execution cannot continue, present the interruption reason and supported recovery/finalization options. Never offer an ordinary new-run state while an existing job's status is unknown.

### F03 · P1 · A failed session is presented as Verified without its failed outcome

**Observed.** A connected-session start failed with `file_not_found`. The retained run `c-6dc5acbb` subsequently opened as `Verified`, `00:00:00 of 00:00:00`, and `0 observations`. It showed true KPI values and zero alarm counts. The earlier launch error was not the report's primary state. Evidence: [09](evidence/09-host-unavailable.txt), [11](evidence/11-failed-session-verified.jpg), [11 text](evidence/11-failed-session-verified.txt).

The integrity explanation correctly says the files match and replay reproduces. The problem is that integrity dominates the operating outcome. An operator can mistake a faithfully recorded failure to start for a successful connected test. Initial-state KPIs are also easy to mistake for measured session results.

**Change:** display independent fields for execution outcome, recording integrity, data coverage, and evaluation result. Lead with `Failed to start — host unavailable`, preserve the error, and label initial-state-only results. Keep the integrity result as useful secondary evidence.

### F04 · P1 · The workspace calls an actively running run Did not finish

**Observed.** While `w-4` was advancing, the workspace's top bar said Running and its run row said Did not finish. Evidence: [26](evidence/26-running-labelled-did-not-finish.jpg).

This contradicts the operator's main status signals and makes an active recording resemble a crashed one.

**Change:** distinguish queued, starting, running, stopping, completed, stopped by operator, interrupted, and failed to start. Absence of a final artifact while running must not imply a terminal failure. Unknown status should remain Unknown until reconciled.

### F05 · P2 · Deliberate stop loses original intent in the review

**Observed.** Requested a 60-minute Explore run `w-4`; Stop now finalized it at 28 minutes. Review showed Verified and 28 minutes, including About this run, without a visible `Stopped by operator` outcome or the original 60-minute target. Evidence: [27](evidence/27-stopped-run.txt).

Keeping the shorter run is good. Presenting it like an ordinary 28-minute completed experiment weakens handover and comparison.

**Change:** retain requested and actual duration, termination reason and timestamp, and initiating actor where supported. Show the planned events that were never reached. Preserve integrity verification separately.

## Monitoring and active operation

### F06 · P1 · The simulation inspector is a catalogue while the plant runs

**Observed.** During `w-1`, selecting CT-1 showed basin temperature and fan power channel names with units, but no current measured values, age, or trend. CH-1 likewise listed channels and effectors without an operating form. Fault badges updated on the graph. Evidence: [01](evidence/01-factory-running.jpg), [02](evidence/02-run-controls.txt). The same pattern occurred on the authored twin in [19](evidence/19-evaluate-running.txt).

An operator can see that a known fault was injected but cannot diagnose its changing consequences from the live inspector.

**Change:** present observed value, unit, quality, acquisition/arrival time, age, and mini-trend for selected signals. Make the asset's observed operating state prominent. Keep model metadata available below it. Evaluate must remain limited to permitted observations.

### F07 · P1 · No usable multi-signal trend investigation was found

**Not found** in the tested simulation and review screens. CT-1 and the thermal sensor Inspect views show cursor-time values and KPI aggregates, without an add-to-trend action, shared multi-asset chart, or saved watch list. Evidence: [04](evidence/04-fault-inspect.jpg), [25](evidence/25-thermal-response-end.jpg).

To understand tower → condenser temperature → chiller power → production quality, the operator must select assets one at a time and remember readings. The successful fan/temperature experiment required manual time seeking to establish the response.

**Change:** allow pinning channels from several assets to synchronized charts, with explicit units and separate scales where needed. Overlay command, fault, quality, and alarm events appropriate to the current role. Add pre/post-event windows and baseline comparison.

### F08 · P1 · Watch execution cannot be operated interactively

**Observed / not found.** The active to-completion run panel offered Stop now. It did not expose pause, step, speed selection, an immediate effector call, or inject/clear disturbance controls. Calls and faults were scheduled before launch. Replay has speed and seek controls, but that is a different task. Evidence: [02](evidence/02-run-controls.txt), [30](evidence/30-evaluate-fault-disclosure.txt).

The operator cannot freeze a simulation to investigate, test a mitigation based on current evidence, or recover from a mistake without ending the run. This falls short of the documented run-and-control and disturb goals for a simulation sandbox.

**Change:** expose simulator-owned pause/step/rate and immediate commands/disturbances in an explicit simulation mode. Record actions in the run history. Do not make a simulation pause imply that physical devices pause; connected operation needs separate control semantics.

### F09 · P2 · The workspace is an inventory, not an operations overview

**Observed.** Home contains plants, device definitions, and runs. It has no persistent overview of active sessions, degraded assets, stale data, pending decisions, or failed operations. A transient global running banner helps with navigation but does not replace that overview. No plant search or status-based run filter was visible; run filtering was by plant. Evidence: [10](evidence/10-workspace-after-failures.txt), [26](evidence/26-running-labelled-did-not-finish.txt).

The opening task becomes “choose a file or plant” rather than “understand the system and act on what matters.”

**Change:** provide an operations landing view with active context, exceptions, pending actions, last contact, and clear links to the relevant asset/session. Keep authoring inventory as a separate destination. Establish scope and data freshness before showing a green summary.

## Experiment design and repeatability

### F10 · P1 · Rerun setup omits the intervention

**Observed.** Set up another run from `w-1` copied seed, duration, mode, and faults, but displayed `Not copied: 1 effector call`. The same happened for `w-3`. Evidence: [05](evidence/05-repeat-drops-call.txt).

The warning is honest, but an otherwise identical fault experiment no longer includes its mitigation. A repeat or comparison can therefore test a different causal question.

**Change:** offer `Repeat exact plan` and `Create variant` with a complete manifest of retained and changed inputs. Include command arguments, times, disturbances, mode, and exact input versions. Clearly mark anything that cannot be reproduced.

### F11 · P1 · Rerun from an incomplete Evaluate run silently resets to Explore

**Observed.** Set up another run from incomplete `w-2` opened defaults: Explore, seed 1, 30 minutes, no faults, no calls. It did not provide the completed-run prefill notice. Evidence: [24](evidence/24-incomplete-rerun-resets-mode.txt).

Recovery changes the experiment's truth policy and removes its intervention schedule. The user may think they are retrying the failed evaluation.

**Change:** recover the original draft when available; otherwise explicitly say the old plan could not be recovered and require deliberate mode selection. Never imply that defaults reproduce an incomplete run.

### F12 · P2 · No scenario library or save/export plan workflow was found

**Not found** in workspace, launch, or review. The UI supports a launch form and partial rerun prefill, but no named scenario, save plan, import/export scenario, or experiment comparison entry point was visible. Evidence: [05](evidence/05-repeat-drops-call.txt), [28](evidence/28-new-plant.txt), [03](evidence/03-factory-report.txt).

Complex fault sequences are work products. Rebuilding them in a transient panel is error-prone and prevents reliable team handover.

**Change:** make a saved, versioned run plan a first-class workspace object. Show its parameters before launch, permit variants, and export the complete session through the binary's supported contract. Display provenance in the resulting review.

### F13 · P2 · Disturbance authoring does not express recovery or delivery conditions

**Observed / not found.** Fault rows expose instance, fault, and start minute. No end/clear time, severity controls, environment-event panel, or delivery-perturbation panel was visible in the tested launch flow. Physical faults and sensor faults appear together in long selectors. Evidence: [05](evidence/05-repeat-drops-call.txt), [18](evidence/18-command-validation.txt).

It is difficult to author “fan fails, sensor becomes stale, transport drops messages, operator intervenes, fan recovers” or explain which layer caused a symptom.

**Change:** provide a time-based scenario editor with explicit disturbance category, target, start, end/clear, and supported parameters. Distinguish world events, measurement faults, and delivery perturbations. The binary must remain the authority for available operations and validation.

## Connected operation and recovery

### F14 · P1 · Plan validity can be mistaken for connection readiness

**Observed.** Check the plan returned `The plan holds: 23 simulated instances` for an absent host path and an empty policy list. The policy warning correctly said commands would be refused except safe stops. Launch then failed. Evidence: [07](evidence/07-plan-check-absent-host.txt), [08](evidence/08-session-failure.txt), [09](evidence/09-host-unavailable.txt).

Structural validation is useful but insufficient for an operator deciding whether the system can run. The check does not establish host availability, device readiness, or permission to command.

**Change:** show separate results: plan valid, endpoint reachable/unknown, expected devices present/unknown, command authority available/refused, and recording ready. Explain which checks were actually performed. Do not weaken the current policy refusal to make readiness appear green.

### F15 · P2 · Connecting equipment requires integration knowledge and manual identity mapping

**Observed.** Session setup asks for host socket, gateway socket, policy digests, and free-text board/device identifiers. Add a real board initially selects the first plant instance. No discovery, endpoint selection, live identity preview, or compatibility check appeared before launch. Evidence: [06](evidence/06-connected-plan.txt).

The form is usable by the person who assembled the integration. It is hard for an on-call operator, and selecting the wrong simulated instance for a physical device could mislead the operator about control scope.

**Change:** provide named saved connection profiles, explicit hardware discovery/import and identity confirmation, and a preflight mapping table. Ask the user to choose an instance rather than silently selecting the first one. Show which identities and policies remain unverified. Keep raw connection details in an expandable section.

### F16 · P2 · Failure handling preserves data but does not guide recovery

**Observed.** Session failures show nested `plantwire connect: plantwire: plantwire:` messages, an error code, `What ran is kept in run ...`, and Back to the session plan. There is no direct Review retained run action on that failure state. The initial `/tmp` attempt correctly refused a non-owner-only socket location, but the UI did not explain the remedy in operator language. Evidence: [08](evidence/08-session-failure.txt), [09](evidence/09-host-unavailable.jpg).

Preserving the failed attempt is valuable. Requiring users to interpret a raw filesystem/security error and then find the run on Home makes incident handling slower.

**Change:** lead with the failed operation, target, reason, and next step. Retain the exact binary error and code under Details. Provide Retry with preserved draft and Review attempt. Explain the owner-only directory requirement without suggesting that security checks be bypassed.

## Plant comprehension and evidence

### F17 · P2 · The factory graph emphasizes wiring over operating state

**Observed.** The 23-instance factory fits at roughly 49% zoom in plant view and 40% in review. Nodes are small, wires cross large portions of the graph, and a fault badge overlaps the node heading. Selecting a node usefully highlights its links, but there is no visible process-level view of the affected subsystem or consequence. Evidence: [01](evidence/01-factory-running.jpg), [04](evidence/04-fault-inspect.jpg).

Even this modest plant is hard to scan in an incident. Navigating technical ports competes with monitoring production impact.

**Change:** provide area-level operating summaries, focus on affected assets and neighbors, selectable connection categories, and progressive detail. Preserve the wiring view for engineering. Make faults and asset names independently readable at fit-to-view zoom.

### F18 · P2 · Missing observations do not explain why data is absent

**Observed.** Factory replay contains many `not observed` port values alongside measured channels. The newly generated twin says `no port data`. The UI gives insufficient local explanation of whether this is an unmeasured model quantity, not-yet-arrived reading, stale reading, disconnected sensor, or unavailable metadata. Evidence: [03](evidence/03-factory-report.txt), [17](evidence/17-generated-twin-topology.txt), [25](evidence/25-thermal-response-end.txt).

These conditions require very different operator responses. A missing model observation must not look like a broken field instrument.

**Change:** distinguish never instrumented, awaiting first reading, missing metadata, bad quality, stale, offline, and access-restricted. Show observation/arrival times, expected cadence where known, and the source of any quality judgment. Do not infer plant failure from absent model metadata.

### F19 · P2 · The report lists results but does not answer the operational question

**Observed.** `w-1` reports disturbances, one applied effector, many mean/range KPIs, zero alarm counts, and no outside-system score. It does not summarize whether cooling degraded, which production assets were affected, or whether the setpoint intervention improved the outcome. Evidence: [03](evidence/03-factory-report.txt), [04](evidence/04-fault-inspect.jpg).

An applied command proves a simulator outcome, not that the mitigation achieved its objective. A run-wide average can hide the period after an intervention. Zero alarms also does not establish that fault detection was correctly configured or effective.

**Change:** add an evidence-backed incident summary with event-relative periods, objective/threshold results, affected assets, and before/after comparison. Clearly separate measured evidence, sealed truth, configured alarm coverage, consumer verdict, and inferred causal explanation. Say when a conclusion is not assessable.

### F20 · P2 · Explore truth explanations contradict each other

**Observed.** Launch says truth is visible after the run. The report says `the true results below were visible while it ran`; the next explanation says an outside system could not see them while live. Evidence: [05](evidence/05-repeat-drops-call.txt), [03](evidence/03-factory-report.txt), [27](evidence/27-stopped-run.txt).

The reader cannot determine who had access to what at which point. This is central to the simulator's credibility as a test instrument.

**Change:** use a single explicit matrix of role, mode, lifecycle, and information visibility. Put the active mode and truth-access status in the persistent shell, including when the launch panel is closed.

### F21 · P2 · Evaluate can be launched without a visible route to a scorer

**Observed / not found.** Evaluate's description promises an outside system can be scored. The to-completion form contains no consumer/sink/scorer connection step. The completed evaluation's report says no outside system is connected. Evidence: [29](evidence/29-evaluate-repro-setup.txt), [31](evidence/31-completed-evaluate-report.txt).

The operator can complete an evaluation-shaped run without having configured the thing to be evaluated.

**Change:** before launch, show consumer connection, verdict contract, scoring readiness, and objective. If an unscored blind run is intentional, label it explicitly and explain what must happen next. Do not equate Evaluate mode with a scored result.

## Device and model authoring

### F22 · P2 · Twin defaults are structurally convenient but semantically surprising

**Observed.** A fan's `level_percent` reading defaults to a valve position transmitter. Adding a thermal sensor offers sensible RTD/humidity instruments, while BLE readings stay plain. The delayed-response form explains gain and time constant usefully, but the initial model is not accompanied by a concise statement of what physical behavior it does and does not reproduce. Evidence: [15](evidence/15-twin-authoring.txt), [16](evidence/16-twin-check.txt).

Compatible units alone do not establish a faithful fan model. A user may accept a default because the twin check later passes.

**Change:** label defaults as suggestions and require review of measurement semantics, ranges, noise, dynamics, saturation, and safe-stop behavior. Show a compact model-assumptions summary and a preview response experiment. Keep interface conformance separate from physical calibration.

### F23 · P2 · Unsupported twin settings are not identified

**Observed.** Adding indicator failed with `A twin does not reproduce this setting`, Open indicator, and Retry. The message did not identify the unsupported field or setting. The rest of the twin could still be Created and Check Passed while the indicator remained outside it. Evidence: [15](evidence/15-twin-authoring.txt), [16](evidence/16-twin-check.txt).

The user cannot tell what to fix or confidently distinguish a complete intended twin from a valid partial one.

**Change:** name the unsupported setting and its impact; link to the exact field. Before creation, show included, excluded, and blocked devices. Make any partial result explicit in the check summary.

### F24 · P2 · Twin provenance is fragmented across device and plant screens

**Observed.** The created three-device twin appears as a read-only plant. Home shows fan as Twin ready, but thermal_sensor and ble_sensor_node still as Needs a twin although included in that twin. The plant has no obvious return-to-behaviour action. It also says some links record no ports. Evidence: [16](evidence/16-twin-check.txt), [17](evidence/17-generated-twin-topology.txt), [26](evidence/26-running-labelled-did-not-finish.txt).

The operator cannot easily answer which device definition, model choices, and plant belong together, or whether a label means “not included anywhere” or “has no separately owned twin.”

**Change:** show the device → twin → plant → physical instance → run relationships with explicit versions and edit origins. Distinguish “included in another twin” from “has its own twin.” Explain why generated models have no port metadata and preserve causal links with meaningful names.

### F25 · P2 · Existing equipment has no visible parameter-editing workflow

**Observed / not found.** CH-1's plant editor exposes rename, area, links, and delete. Add equipment offers size presets. No existing-instance parameter, preset review, or operating-envelope editor was found. Saving an added area successfully returned `Ready to run. Saved.` and composition counts. Evidence: [12](evidence/12-editor-asset.txt), [13](evidence/13-editor-save.txt).

An IoT specialist needs to inspect and adjust capacity, response, limits, and model assumptions, not only connect boxes. Save success and structural composition do not explain the changed operating envelope.

**Change:** expose supported model parameters through the owning binary, with units, bounds, defaults, provenance, and change impact. Show validation and plausibility-check status separately. Let users see the planned version diff before launching a new experiment.

### F26 · P2 · Read-only plants have no obvious safe revision path

**Observed / not found.** Many workspace plants, including the generated twin, are marked read-only. New plant's start selector offers the composition-backed chiller/factory and the same examples; the reviewed UI does not provide a clone-to-edit path for the other domains. Evidence: [28](evidence/28-new-plant.txt), [17](evidence/17-generated-twin-topology.txt).

The operator can run a model but cannot discover how to revise it or which editor owns it. Read-only describes a limitation without a next step.

**Change:** explain the reason and offer the correct action: edit source behaviour, create a supported variant, import an editable composition, or open documented advanced editing instructions. Never silently lose unsupported domain semantics when converting to an editor format.

## Discoverability and clarity

### F27 · P3 · Board capture offers runs with no board announcement

**Observed.** Capture from a board initially selected the newest incomplete simulation `w-2`; the dropdown also offered ordinary simulations and failed sessions. Read this run correctly reported No announcement found, said nothing was saved, and offered Retry. Evidence: [22](evidence/22-capture-no-announcement.txt).

The recovery message is good. The selection makes the user perform an avoidable failed operation before learning which runs are usable.

**Change:** identify or filter runs with board announcements, show unsupported/unknown availability, and explain how to record a suitable source. Retain paste/import as an advanced route. Do not imply this page discovers a physically connected board.

### F28 · P3 · Raw unit codes and technical names interrupt operational reading

**Observed.** The thermal twin displays `Cel`; BLE integer readings display values such as `128 1`; other plants display `°C`. Device and policy terminology includes kind names, digests, and internal-style identifiers. Evidence: [25](evidence/25-thermal-response-end.jpg), [14](evidence/14-device-command.txt), [06](evidence/06-connected-plan.txt).

Raw codes are useful for integration but slow the main task of interpreting a value and its meaning.

**Change:** present human-readable unit symbols and names, with canonical codes in details/copy actions. Render dimensionless readings without a misleading visible `1` suffix unless the distinction matters. Allow engineering labels while retaining exact identifiers for audit and integration.

## Strengths to preserve

- The factory is a useful integrated plant, not a decorative demo. Areas, outline search, typed connections, selected-neighbor links, legend, and minimap support engineering inspection.
- Invalid fan level 70 was blocked before launch with an exact whole-number range message and field focus. Evidence: [18](evidence/18-command-validation.txt).
- Replay supports event markers, event stepping, speeds, and keyboard seeking. A valid thermal intervention produced a visible recorded response.
- Verification explains file matching and replay. Preserve this evidence while separating it from execution outcome.
- Failure states retain run data. Deliberate Stop now finalizes a shorter recording. Normal in-app navigation retains a global run banner and return path.
- Device commands show risk, argument bounds, and the special safe-stop semantics. Connected command copy explicitly distinguishes acceptance/report from effect. Evidence: [14](evidence/14-device-command.txt), [11](evidence/11-failed-session-verified.txt).
- The twin response form explains negative gain and time constant, and generates a readable response sentence once filled.
- Capture correctly says nothing was saved when no board announcement was found.
