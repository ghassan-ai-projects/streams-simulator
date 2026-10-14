# Canvas product review — end state, guidance and functional coverage

- Date: 2026-10-14.
- Reviewed: [UI expectations](../UI-EXPECTATIONS.md), the existing canvas design, goals, source and tests.
- Purpose: work backward from the intended finished product and identify what the current expectations still need.
- Scope: review document only. The owner explicitly requested no implementation.
- Evidence: document and source inspection, not a fresh Chrome audit or an executed simulation.

## 1. Assessment

The expectations establish a useful plant-centred model and cover the reduced build → run → review scope.
They do not yet fully explain what the simulator does, how a beginner creates an experiment, what running and stopping
mean, or how the broader live and consumer-testing workflow fits together.

The visual direction is sound but needs a clearer finished-product reference. A checklist can pass while the screen
still feels confusing: correct controls do not establish a clear mental model by themselves.

This review builds on the product already started. It does not replace its design or silently bring every deferred
capability into the current implementation scope. Requirements below are proposed additions; checks are unrun.

## 2. Start with the finished product

One person can open the product, understand its purpose, run a supplied example, change a plant, introduce a disturbance,
stop the experiment and explain the result without reading JSON or opening a terminal after setup.

The opening explanation should communicate:

> Build a simulated plant, generate its readings, test how a connected system responds, and review what happened.

The product distinguishes three objects in plain language:

- **Plant:** the equipment, connections and modelled behaviour.
- **Run plan:** the duration, repeatability settings and disturbances for one experiment.
- **Run:** the recorded execution and its results.

“New simulation” is an entry into these choices, not a fourth object duplicating plants or runs.

### What the finished screens look like

| Screen | Main content | Primary next action |
|---|---|---|
| Workspace | Plants and recent runs; a prominent example entry; helpful first-use guidance | Try an example or New simulation |
| Creation | Ready-to-run example, copy existing plant, or empty plant; name and short explanation | Create and open |
| Plant/editor | Readable equipment graph, searchable outline, useful inspector, clear save/readiness state | Run or Save changes |
| Running | Plant name, simulated time, run status, planned disturbances and clearly labelled stop control | Cancel run; live controls only when supported |
| Review | What happened, aligned event timeline, playback, readings/trends, consumer results when available | Run again or Edit plant |

Conceptual layout, not a screenshot or an implemented mockup:

```text
+------------------------------------------------------------------------+
| Streams Simulator   Workspace > Fan experiment       Theme   Help      |
+------------------------------------------------------------------------+
| Fan experiment   Simulated device   Ready             Edit   Run       |
+----------------+-----------------------------------+-------------------+
| Outline      < | Equipment and connections          | > Details         |
| Search         |                                   | Selected equipment|
| Sensor         | Sensor → control path → fan        | What it does      |
| Fan            |                                   | Receives / supplies|
| Indicator      | Readings and visible status        | Values and origin |
+----------------+-----------------------------------+-------------------+
| Review playback: Play   speed   time cursor   event lanes              |
+------------------------------------------------------------------------+
```

The bottom playback bar belongs to review. A running experiment has its own controls and status. Replaying recorded
events must never look like starting a new simulation or sending a command to hardware.

## 3. Beginner guidance and simple creation

| ID | Proposed requirement | Acceptance evidence |
|---|---|---|
| N-1 | Offer a ready-to-run example as the shortest first-use path; also allow copying and empty creation | A first-time user reaches a completed example run without editing files |
| N-2 | Ask for a human name; suggest the technical ID and validate it through the owning binary capability | No schema knowledge required; conflicts produce an actionable field message |
| N-3 | Explain what will be copied, retained or omitted before creation | Copying a complete plant and creating an empty plant are visibly different choices |
| N-4 | Guide the whole journey: create, add equipment, connect, check, run, review | Each unfinished step offers a specific next action; guidance can be dismissed and reopened |
| N-5 | Use sensible visible defaults; put seed and specialist settings under Advanced | An example runs with default settings; mode and truth exposure remain clear |
| N-6 | Equipment choices explain purpose, preset size and important parameters | A user can choose equipment without guessing what a component ID means |
| N-7 | Errors identify the affected item, explain the problem and offer recovery | A deliberately invalid connection can be fixed from its error without finding a file |

The inspected creation form requires an ID and an existing base. Its explanation says it copies site, areas, profiles
and axes but starts without equipment. This is an empty-builder workflow, not a complete example-copy workflow.
See [NewPlant](../../../canvas/src/ui/home/NewPlant.tsx) and
[newComposition](../../../canvas/src/model/compositionEdits.ts).

An actually empty workspace needs a documented source of starter examples; it cannot depend solely on existing
workspace compositions. Starter content must remain ordinary validated data, not a special simulator code branch.

## 4. Explain running, stopping and results

**Run** means creating a new execution of the selected plant, advancing simulated time, applying the planned
disturbances and recording evidence. It does not automatically mean real-time playback, connecting a consumer,
running the supervisor, or operating physical equipment.

The inspected launch flow calls `world.create`, advances the clock and injects scheduled faults, then calls `run.end`.
It does not itself launch the external supervisory stack. See [LaunchStore](../../../canvas/src/app/launchStore.ts).

Before Start, show the plant, mode, simulated duration, disturbances, consumer connection if applicable and where the
result will appear. During execution, show what stage is active and distinguish simulated time from elapsed wall time.

| ID | Proposed requirement |
|---|---|
| RUN-1 | Explain Start: “Simulate this plant for the selected duration and save a run for review” |
| RUN-2 | Explain whether execution runs to completion or supports interactive live control; expose only working controls |
| RUN-3 | Keep the stop control visible and state whether it cancels, pauses or ends and saves the run |
| RUN-4 | Explain cancellation latency and acknowledge the request immediately; do not imply an in-flight command stops instantly |
| RUN-5 | Distinguish complete, cancelled, interrupted, failed and checking states; explain what evidence was retained |
| RUN-6 | Closing/collapsing a panel must not implicitly cancel a run; navigation behaviour must be deliberate and visible |
| RUN-7 | Review Play reads recorded events; Run again creates a new execution with an explicit plan and plant version |
| RUN-8 | Explain repeatability accurately: seed alone is insufficient; model, simulator version, inputs and commands also matter |

Source inspection found cancellation sets a flag checked before subsequent plan steps and then attempts `world.destroy`.
It does not establish that partial evidence is finalized or retained. A user-facing “End and save” promise therefore
needs separate binary support and tests. `LaunchPanel` also requests cancellation on unmount: collapsing panels must
not accidentally trigger this lifecycle. See [LaunchPanel](../../../canvas/src/ui/launch/LaunchPanel.tsx).

## 5. Explain connections, inputs and outputs

A connection makes a value supplied by one simulated component available to a receiving component according to the
declared model. It is a meaningful model dependency, not merely a line on the screen.

- **Output / Supplies:** a value or resource representation this component provides.
- **Input / Receives:** the modelled quantity this component uses.
- These terms describe equipment ports. They are separate from the simulator's event stream to a consumer.
- An equipment connection may represent temperature, demand, power, flow or a schedule. It is not always a literal pipe.

Examples already present in the factory composition:

- Calendar shift output → production-line shift input: the line receives the operating schedule.
- Chiller heat output → cooling-tower heat input: the tower receives the modelled heat-rejection load.

See [factory composition](../../../examples/compositions/factory-utilities.composition.json).

| ID | Proposed requirement |
|---|---|
| C-1 | Ports show plain names, direction and quantity; raw identifiers are secondary |
| C-2 | Selecting a wire explains “A supplies X to B” and what the receiving model uses it for |
| C-3 | A guided first connection previews both endpoints and explains its effect before adding it |
| C-4 | Show compatible targets and explain rejected links using binary-owned rules; colour family alone is not compatibility |
| C-5 | Explain required, optional, unconnected and default-fed inputs from component metadata; do not invent this information |
| C-6 | Distinguish a model connection from physical wiring and from communication with the consumer |

The proposed explanation requires metadata when the current contract does not supply meaning, units or default
behaviour. It must not be a second hand-written implementation of component semantics in the UI.

## 6. What is fixed, inherited or hard-coded?

The problem is not every constant. It is that users cannot tell a model assumption from a chosen setting, a default,
a presentation convention or a product limitation. “Why” below describes the source's role, not verified author intent.

| Observed item | Source and role | Required treatment |
|---|---|---|
| Available areas | Composition data; creation inherits them from the base and equipment forms list them | Explain inheritance. Area names are not a fixed Go list. Area editing remains deferred under the current goal |
| Presets and preferred preset | Binary component response, parsed by `library.ts` | Show the selected preset, what it sets and which values are overrides |
| Equipment parameters | Composition/library data; factory compressors have explicit pressure-band overrides | Explain origin and editability. The inspected instance editor lacks general parameter editing |
| Seed 1, 30-minute duration, Explore | `firstDraft` in the run form; editable UI starting values | Label as defaults, not model requirements. Record the actual selected plan |
| Minimum one-minute duration | `planProblems` in the canvas model | Review ownership and whether this is a deliberate product restriction; do not present it as a physics limit |
| Clock advance chunk count of 20 | `launchSteps`; canvas command orchestration | Explain stop/progress granularity and test determinism. Review ownership before changing it |
| Output file locations | `domainOf` and editor creation follow naming conventions | Keep paths secondary; the existing goal already defers binary-owned path resolution |
| Hall grouping | `hallOf` derives a group from the instance-ID prefix before `/` | Document the naming convention; do not imply it is universally the declared area hierarchy |
| Layout dimensions, zoom tiers and unit labels | UI presentation constants | Keep in the visual system; these are not simulated physical values |

Evidence: [composition](../../../canvas/src/model/composition.ts), [library](../../../canvas/src/model/library.ts),
[instance editor](../../../canvas/src/ui/editor/InstanceEditor.tsx), [run defaults](../../../canvas/src/ui/launch/draft.ts),
[launch rules](../../../canvas/src/model/launch.ts), [plant view](../../../canvas/src/model/plantView.ts),
[layout](../../../canvas/src/model/layout.ts) and [units](../../../canvas/src/model/units.ts).

The finished inspector should answer: What is this value? Where did it come from? Can I change it? What will it affect?
Show origins such as “Preset”, “Plant override”, “Run setting”, “Observed” and “Simulated truth” where meaningful.
Do not mix measured and simulated values, or silently substitute missing readings with plausible numbers.

## 7. Visual improvements and collapsible panels

| ID | Proposed visual requirement |
|---|---|
| VIS-1 | One stable shell; plant/run title first, short summary second, actions aligned consistently |
| VIS-2 | Equipment is recognizable by plain kind and icon; zoom switches detail rather than shrinking important text into illegibility |
| VIS-3 | Quiet grid, subtle area frames and flatter normal nodes; selection and abnormal states carry stronger emphasis |
| VIS-4 | Distinct treatment for category, selection, keyboard focus, faults and alarms, including combined states |
| VIS-5 | Aligned workspace rows and inspector values; predictable action locations; fewer competing chips and headings |
| VIS-6 | Campus overview uses readable hall summaries; opening a hall reveals readable equipment. Fit and universal node detail are not simultaneous requirements |
| VIS-7 | Review has aligned fault/alarm/effector lanes, one time cursor and compact labelled trends |
| VIS-8 | Left and right panels collapse independently through visible buttons and keyboard paths; reopening remains discoverable |
| VIS-9 | Panel state is remembered; collapsing preserves selection, zoom, unsaved work and run lifecycle; focus moves predictably |
| VIS-10 | Side panels have consistent starting widths and can resize within bounds; the centre takes remaining space |
| VIS-11 | Provide reference screens and a state sheet for both themes, including empty, loading, error and selected states |

P-13's close control and collapsible inspector sections do not specify collapse of both entire side panels.
G-2 should explicitly describe bounded side widths and a flexible centre. Review at 1280, 1440 and 1920 px, with all
four panel combinations, using the real binary. Contrast and readability need measurement and rendered inspection.

## 8. Functional coverage and omissions

“Covered” means specified in the expectations, not proven implemented or passing.

| Capability | Coverage | Scope / next decision |
|---|---|---|
| Workspace, plant opening and run discovery | Covered by W/P items | Immediate core |
| Add equipment, connect, save, compose and show issues | Covered by E items | Immediate core; guidance and compatibility semantics need additions |
| Run duration, seed, mode and planned faults | Covered by R items | Immediate core; explain execution and cancellation |
| Playback, event navigation, report and rerun | Covered by V items | Immediate core; verify against the changing implementation |
| Beginner example, complete-copy creation and guided journey | Partial or absent | Immediate experience requirement proposed here |
| Collapsible side panels | Absent as a complete interaction | Immediate experience requirement requested by owner |
| Configure existing equipment's presets/parameters | Inadequately specified | Proposed core extension; binary/library metadata required |
| Validate and plausibility-lint readiness | Validation present; lint/readiness workflow thin | Distinguish saved, composed and plausibility-checked |
| Interactive live monitoring, pause, step, speed and reset | Mostly absent | Explicitly deferred in the reduced production goal; part of broader P10 |
| Inject/clear faults, environment changes and delivery perturbations mid-run | Mostly absent | Broader P10; distinct from scheduling faults before Start |
| Save, edit, run and export reusable scenarios | Mostly absent | Scenario listing is not scenario authoring/execution; broader P10/P11 |
| Connect a consumer, show its activity and assess its response | No complete workflow | Broader goal; use reference-host scope and preserve external integration boundaries |
| Full kinds, dynamics, sensors, actions and fault authoring | Absent from reduced editor expectations | Explicitly deferred P11 work |
| Conflicts, interruption and partial evidence recovery | Partial | Core reliability requirements need explicit acceptance cases |
| True versus observed values | Specified but partly deferred | Preserve mode/sealing restrictions; do not silently expand current scope |

Source of scope decisions: [reduced production goal](../GOAL-UI-PRODUCTION.md), [UX goal](../GOAL-UI-UX.md),
[original goal](../GOAL.md), [design](../canvas-design.md), [phase plan](../../../governance/CANVAS_PLAN.md) and
[acceptance](../../../governance/CANVAS_UI_ACCEPTANCE.md).

Integrity verification and the consumer's verdict must remain distinct. A reproduced run can contain a poor consumer
response. Show checking, not checked, check failed and incomplete separately from verified/not verified.
V-13 now addresses revealed truth wording; the checklist also needs the actual payload and lifecycle boundary from
the acceptance document. Do not remove Explore's truth-exposure notice merely to eliminate jargon.

## 9. Flagship example: simulated version of the real fan experiment

The owner requests a simulated version of the existing DHT11 + fan + LED experiment as the concrete example of moving
from simulation to the real world. This means an executable, data-defined experiment with a reviewed transfer boundary,
not only a link to the bench report or a decorative animation. Implementation is pending.

### Intended example experience

“Try the fan experiment” opens a small, understandable setup. It explains temperature and humidity readings, the bounded
fan action, indicator meaning and feedback. Start generates a recorded simulation; the user follows observation →
decision → admitted command → reported result → separate effect evidence → safe stop.

The graph must distinguish physical/modelled relationships from event and command paths. Do not label every arrow
“connection” when some represent supervision or protocol exchanges rather than equipment ports.

| Scenario | Required simulated outcome |
|---|---|
| Normal temperature | Monitoring; no cooling command |
| One high sample | No action solely from a spike |
| Persistently hot and rising | The reviewed supervisory path requests bounded cooling; commands respect declared limits |
| Temperature below lower band or lease expiry | Declared stop behaviour; final state and its evidence shown |
| Stale/disconnected sensor | No fresh actuation based on invalid evidence; reason visible |
| Accepted command with no effect / stalled fan | Receipt is visible but effect stays unverified; no false cooling success |
| Reboot, lost receipt or uncertain outcome | Reconciliation before retry; identity and uncertainty preserved |

Temperature is the primary control signal; humidity remains a distinct correctly typed measurement. The experiment
proposes hysteresis, persistence and trend, rather than a single threshold. Its 28 °C / 27 °C simulator thresholds are
example settings, not established hardware operating limits. Any simulated cooling dynamics need declared assumptions
and simulator-owned oracles; copying the command protocol alone does not reproduce a physical thermal process.

### What constitutes transfer proof

1. Define one reviewed observation, intent, device-command and feedback contract, including units and uncertainty.
2. Run the simulated device through the intended supervisory/consumer path, with recorded commands and effect evidence.
3. Keep the reviewed control logic and contracts fixed while replacing the simulated endpoint with the hardware gateway.
4. Declare the differences: timing, noise, sensor resolution, model parameters, capabilities and evidence source.
5. Compare scenario outcomes and contract behaviour; do not demand bit-identical physical sensor traces.
6. Require physical feedback before claiming real actuation, and a separate measured thermal outcome before claiming cooling.

A simulation pass proves the software/model contract for its tested scenarios. Transfer is established by a paired
hardware run, not inferred from the simulation alone. Real hardware setup/control remains in its existing gateway
boundary; this proposal does not give the canvas direct equipment-effector authority or add consumer-specific code
to the simulator. A missing control, adapter or model capability needs a documented binary/module design first.

### Source experiment and evidence boundary

Read-only source references assume the existing sibling repository layout; they are documentation references, not
runtime or test dependencies:

- [Experiment 02](../../../../agent-research-lab/real-world-sensor/assessment/EXPERIMENT-02-DHT11-FAN-LED.md).
- [Execution report](../../../../agent-research-lab/real-world-sensor/assessment/runs/round-008-fan-experiment-20260908-01/EXECUTION-REPORT.md).
- [System diagrams](../../../../agent-research-lab/real-world-sensor/diagrams/README.md).

The later 2026-09-09 report records corrected L293D input wiring, operator-confirmed rotation under a governed direct
serial command, DHT11 telemetry and safe-stop. It explicitly leaves the live Tamoz → Agentic Stream → gateway loop and
conversion seam open. Rotation was operator-observed, not independently instrumented. Earlier failure entries and the
older status summary must not override that later scoped result, nor should the later result become a full-HIL claim.

The conversion from Tamoz's `request_bounded_cooling` to Agentic Stream's `select_thermal_mode` is an explicit contract
gap to resolve before claiming the joined simulated and physical workflow. A completed simulated example must show
which integrations actually ran and must label unavailable stages rather than synthesizing their success.

## 10. Work backward into milestones

| Milestone | Outcome to prove | Work implied |
|---|---|---|
| M1 — Understand | A beginner explains plant, run, connection, input/output and stop behaviour | Plain wording, example explanations, value origins and next-step guidance |
| M2 — Create and recover | A beginner opens an example or creates a plant, fixes an issue and saves without losing work | Simple creation, useful defaults, compatibility feedback, panel controls and recovery |
| M3 — Run and review | The user starts, cancels and reviews with accurate lifecycle and evidence states | Explicit run semantics, retained-result policy, playback/report clarity and meaningful checks |
| M4 — Simulated fan | The same experiment executes its reviewed contracts and fault cases in simulation | Validated data/model, supervisory integration and reproducible evidence; implementation pending |
| M5 — Transfer | The simulated endpoint is replaced by hardware with declared differences and paired evidence | Reviewed adapter seam and separately authorized physical qualification; outside this document's execution |
| M6 — Broader instrument | Interactive disturbances, reusable scenarios and consumer testing work end to end | Explicitly scheduled P10/P11 work, not an accidental expansion of polish rounds |

Protect truth, evidence and user work before visual polish. Keep existing workspace/wording priority for the experience
rounds, but do not classify unsaved navigation, failed recovery or misleading measurements as cosmetic leftovers.

## 11. Review evidence and completion

Future implementation rounds should identify a frozen build, exact widths/themes, test results and screenshots.
Record each proposed ID as met, partial, failed, deferred or not tested, with an evidence pointer. Split compound
expectations when their parts can pass independently. Define performance timing boundaries and workloads explicitly.

Source and tests were inspected around branch `claude/canvas-ux`, HEAD `edd8ecc`; concurrent source edits were present
during this review, including emerging playback controls. The dated “Now” column is therefore historical evidence,
not a fresh verdict on the latest working tree. Existing Playwright tests show intended build/run coverage; they were
not executed for this review. No implementation, simulation, hardware action, gate change, commit or push is part of it.

Document validation: check relative links, whitespace and repository diff. Full CI and browser acceptance remain
unrun because this change is a review document only.
