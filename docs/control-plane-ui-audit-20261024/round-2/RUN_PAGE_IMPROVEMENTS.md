# Run-page improvements

The walkthrough used a factory cooling exercise with sensor loss, a fan failure, a stop/restart sequence, and a setpoint change. The question was: **what would help a person plan, follow, decide, and explain this run without reconstructing its meaning across screens?**

All items below are proposals. The evidence records the current workflow that motivated them. “First” identifies the core operating experience; “Next” adds decision and investigation depth; “Later” adds repeatable team workflows. This is not a defect register or a backend implementation assessment.

## Prepare the run

### R01 · First · Give every run a purpose and a readable name

**Opportunity.** Seed, duration, faults and effectors describe inputs. The user also needs to capture the question: “How does cooling respond when the tower fails and its temperature sensor is unavailable?” The completed records are identified mainly by IDs such as w-6. See [the plan](evidence/09-run-plan.txt).

**Improve it.** Start with a name, a short objective, and the evidence needed to answer it. Offer domain-declared objectives where available; allow a plain-language investigation goal without converting it into a fabricated pass/fail rule. Keep the stable run ID beside the name for audit.

**Example.** Name: `Cooling loss — staged response`. Investigation goal: `Observe chiller response and production impact with tower temperature unavailable`. Quantitative limits must be explicitly configured or supplied by the domain; do not invent acceptable temperatures.

**Acceptance.** The operator can identify the run's question before launch and recover it from Live, Review, and history. An unset quantitative objective is shown as an investigation, not a passed experiment.

### R02 · First · Build the plan in chronological order

**Opportunity.** Configuring two faults and three calls produced a long right panel. The first listed fault was at minute 5 and the next at minute 3; calls occupied a separate section. The user has to mentally merge the sequence. See [the plan screenshot](evidence/09-run-plan.jpg).

**Improve it.** Show one time-ordered plan with lanes for process disturbances, observation/delivery changes, and commands. Keep an editable table as an alternative to dragging. A selected event opens its editor, with exact arguments and units.

**Example sequence.** 03:00 sensor dropout → 05:00 fan failure → 08:00 stop → 12:00 restart → 15:00 setpoint 9°C. Give repeated or simultaneous events an explicit order supplied by the simulator.

**Acceptance.** The operator can read the whole sequence without scrolling through multiple categories, move an event, and explain its order. Available event types and validation come from the binary.

### R03 · First · Let the operator define phases

**Opportunity.** The exercise naturally divides into nominal operation, observation loss, process disturbance, shutdown, restart, and adjustment. The current run and report expose events and whole-run aggregates. See [review overview](evidence/12-review-overview.txt).

**Improve it.** Allow named intervals anchored to events: `Before disturbance`, `Sensor unavailable`, `After fan failure`, `Stopped`, `Restarted`, `Adjusted`. Phases become navigation, comparison windows, and report sections. Keep user-defined phases distinct from detected conditions.

**Acceptance.** Selecting a phase aligns charts, relevant events, and summary statistics to that exact interval. The UI states whether an aggregate covers the selected interval or the whole run. Simulator-owned calculations remain in the binary.

### R04 · First · Make readiness a concise review of intent and capability

**Opportunity.** Start summarizes counts and mode. A person operating a complex scenario benefits from seeing targets, what will happen, what will be measured, and which capabilities are ready. The existing time-bound validation usefully focuses the field requiring attention. See [schedule guidance](evidence/08-schedule-bounds.txt).

**Improve it.** Add a compact preflight section: exact model/version, environment, truth visibility, duration, watched evidence, schedule, output/consumer readiness, and any intentional limitations. Link each item directly to its editor. Keep endpoint and hardware readiness separate from plan validity in connected mode.

**Acceptance.** The operator can confirm the intended environment and full schedule in one view. Each blocking item has a specific next action. An unscored investigation can still run with its status explicit.

### R05 · First · Choose a watch list while planning

**Opportunity.** The operator knows before launch that tower temperature, fan power, chiller temperature/power, cooling-meter power, and production-quality evidence matter. These are currently encountered through individual asset selection. See [live view](evidence/10-run-live.txt) and [intervention view](evidence/14-intervention-selected.txt).

**Improve it.** Add `Watch` beside channels and offer a suggested signal set for the chosen assets. Save the watch list with the run plan. Distinguish available measured channels from model-only quantities and post-seal truth. Unsupported signals should explain their availability.

**Acceptance.** Launch opens the selected signals together. The same set persists into Review and comparison. The observer receives only evidence allowed by its mode and role.

## Follow the run

### R06 · First · Promote the run to its own workspace

**Opportunity.** During execution the primary interaction remains a plant plus a run panel, with progress and Stop now. The operator's work has expanded beyond configuring that form. See [live screenshot](evidence/10-run-live.jpg).

**Improve it.** On Start, enter a dedicated run workspace. Keep name/ID, plant version, origin, Explore/Evaluate, truth access, execution status, simulation time, wall time, and recording status consistently available. Use Plan / Live / Review as state-aware views of the same run.

**Acceptance.** Selecting an asset, a signal, an event, or a plan step preserves the run context. The user can inspect evidence while retaining access to run controls and upcoming work. Historical playback has an unmistakable Replay label.

### R07 · First · Make trends the primary live diagnostic surface

**Opportunity.** At a chiller stop event, readings at 08:00 and 09:00 give different parts of the action's response. Inspecting isolated values takes deliberate time navigation. See [08:00](evidence/14-intervention-selected.txt) and [09:00](evidence/15-after-stop-observation.txt).

**Improve it.** Show synchronized small-multiple trends for the watch list, with shared time, units on each plot, event overlays, quality gaps, and a common cursor. Let the plant remain a compact adjacent context or an alternate main view. Keep the time range and selection when switching views.

**Acceptance.** The operator can follow tower evidence, chiller response and downstream effect at the same time, then inspect exact values at one cursor. Signals with different units retain meaningful independent scales. Missing samples are not silently interpolated across a loss interval.

### R08 · First · Explain freshness directly beside the value

**Opportunity.** At replay time 30:00, the tower's basin value was 23.5°C with Observed at 03:00. The existing timestamp is useful and should be retained. See [sensor-loss evidence](evidence/13-sensor-loss-review.txt).

**Improve it.** Add age and quality beside the latest value: `Last observation 03:00 · 27 min old at cursor`. Use the backend's quality/cadence rules for labels such as stale. Distinguish measurement time, receipt time, missing observation, instrument-range information, and a process-limit alarm. Provide a last-good-sample action and show the gap on the trend.

**Acceptance.** The operator can tell which value is current, held, missing or outside the instrument's declared range without interpreting a small icon or subtracting times. A missing reading is never treated as a normal zero. Evaluate can show permitted quality evidence without revealing the sealed cause.

### R09 · First · Show Now and Next from the plan

**Opportunity.** The live run panel summarizes two faults and three calls, while the operator needs the current phase, what already occurred, and what will happen next. See [live panel](evidence/10-run-live.txt).

**Improve it.** Include a compact phase band and a Now / Next list with event time, countdown, target and state. In Explore or an explicitly privileged director view, show scheduled disturbances. In blind observer mode, show only permitted actions and observations; future hidden disturbances stay hidden.

**Acceptance.** During the exercise the authorized user can identify the next scheduled chiller action and the evidence being collected in the current phase. Completed, pending, skipped, and unconfirmed events are distinct when the binary reports them.

### R10 · Next · Focus the plant on operational relationships

**Opportunity.** Selecting CH-1 highlights its links and event navigation focuses the correct asset. The full 23-instance graph still occupies much of the review. See [selected intervention](evidence/14-intervention-selected.jpg).

**Improve it.** Add an `Affected system` focus that groups the cooling tower, chiller, relevant pumps/meter, and connected production assets. Use physical/typed relationships from the model. Offer overview, focused process, and wiring detail without requiring the user to relearn navigation.

**Acceptance.** A user can follow the tower-to-chiller-to-production relationship, expand a neighbor, and return to the full plant while preserving evidence time. Graph reachability should be labeled as a relationship, not automatically asserted as proven causal impact.

### R11 · Next · Give each command an action story

**Opportunity.** Selecting a command correctly moves to the asset and time. The stop shows applied with 338 ms; the setpoint event shows applied with 285 ms. The operator also needs the requested setpoint, the expected response, and the relevant observations after application. See [stop](evidence/14-intervention-selected.txt) and [setpoint](evidence/16-setpoint-event.txt).

**Improve it.** Open a command detail panel with target, source/actor, reason, exact arguments/units, requested and applied times, reported outcome, and linked evidence. Distinguish simulation application, hardware receipt, board report, and independently observed effect according to actual capability. Add `Show before and after` to open an aligned evidence window.

**Acceptance.** The reviewer can reconstruct `CH-1 setpoint = 9°C at 15:00`, see what the simulator reported, and inspect subsequent measurements without rebuilding context. If the intended effect cannot be assessed, say which evidence or criterion is missing.

### R12 · Next · Put approvals in the run's decision context

**Opportunity.** The user's intended control plane includes approving actions. This walkthrough used scheduled simulator calls; it did not exercise a live host-backed approval queue.

**Improve it.** Give the run a visible Decisions area, with a count and a selected-request panel linked to the asset, evidence window, and purpose. Show exact target, environment, arguments, risk, policy scope, expiry and preconditions. Approve once, reject with a reason, or defer. Keep enabling a policy, releasing a timing hold, and approving a specific action distinct.

**Acceptance.** A proposed action can be understood and decided on without leaving its evidence. The result remains linked to the command and its outcome. Request binding, expiry, authority checks and dispatch are owned by the backend; the UI displays their results.

### R13 · Next · Make time controls match the operating task

**Opportunity.** The interface provides watch pacing during execution and several replay speeds afterward. An operator's task changes between watching a transient, preparing a response, and investigating recorded evidence. See [live panel](evidence/10-run-live.txt) and [review](evidence/12-review-overview.txt).

**Improve it.** Explain simulation time, elapsed wall time, selected cursor time and execution rate separately. Provide deliberate simulator pause/step/rate controls through supported binary operations, alongside suitable run-end controls. In connected operation, explain which controls affect the simulated world, recording lifecycle, or physical equipment.

**Acceptance.** The user can slow a simulation around an intervention and later replay it without ambiguity about what is advancing. Pausing a chart or replay never implies a physical stop. Control labels reflect the actual backend operation.

## Investigate and explain

### R14 · First · Turn the timeline into an investigation tool

**Opportunity.** Event markers, next/previous stepping, keyboard seeking and automatic asset focus provide a strong starting point. The operator can use them to investigate the staged sequence. See [the five-event timeline](evidence/14-intervention-selected.jpg).

**Improve it.** Add a direct time field, a selectable interval, zoom to selected event/phase, and filters for the selected asset or event category. Provide separate lanes for commands, quality/availability, conditions and annotations. Keep one shared cursor across charts, asset values and event details. Show both the event and its subsequent evidence window.

**Acceptance.** The user can select 08:00–12:00, inspect the shutdown period, jump to the next action, and return to the previous window. Keyboard and pointer paths lead to the same investigation state. Hidden Evaluate events remain hidden.

### R15 · First · Label the evidence type and its time scope

**Opportunity.** Inspect shows observations at the selected cursor alongside true KPI mean/range results. The intended reading is easier when each section states its source and aggregation interval. See [selected chiller evidence](evidence/14-intervention-selected.txt).

**Improve it.** Label sections `Observed at cursor`, `Whole-run truth`, `Selected-phase summary`, and `Consumer-reported result` where applicable. Add evidence provenance and completeness in Details. Preserve the existing distinction between observations and truth, with explicit time scope.

**Acceptance.** The user can tell which number is a momentary sample, a phase aggregate, a whole-run statistic, a model truth value, or a consumer assertion. Changing the interval updates only results computed for that interval and clearly labels unchanged whole-run results.

### R16 · First · Lead Review with the run's question and outcome

**Opportunity.** The completion screen reports that every call was accepted. The report supplies counts and KPI tables. The user needs to know what those facts mean for the original investigation. See [completion](evidence/11-run-completion.txt) and [review](evidence/12-review-overview.txt).

**Improve it.** Begin with purpose, execution outcome, evidence coverage, configured objective results, and unresolved questions. Keep recording/replay integrity and scoring as separate assessments. Then provide phase summaries, action stories, and detailed tables. Use `Not assessed` when no criterion was declared.

**Acceptance.** A reviewer can explain what ran, which actions were applied, which observations are incomplete, and which objectives were assessable. The page supports a concise handover without requiring a tour of every KPI row.

### R17 · Next · Compare reference runs in the same workspace

**Opportunity.** This round created nominal, disturbance-only, and planned-response runs with seed 314 and the same duration. Their visible chiller mean COPs were 5.09, 5.02 and 4.40 respectively. Those figures need phase, downtime, production and coverage context before an operator can evaluate the response. See [nominal](evidence/19-nominal-reference.txt), [disturbance reference](evidence/17-fault-only-reference.txt), and [planned response](evidence/12-review-overview.txt).

**Improve it.** Let users choose reference runs, see the input differences, align by run time or event onset, and compare the same signals/objectives over the same interval. Show missing evidence, changed model versions, and other comparability limits. Allow named reference roles such as nominal, disturbance, and response.

**Acceptance.** The operator can identify what differs between these three plans and compare a chosen post-event window. The interface does not rank responses from one aggregate or present a single-seed exercise as broad proof of effectiveness.

### R18 · Next · Keep a complete, editable experiment package

**Opportunity.** The existing rerun prefill is useful for preparing a related experiment. In this walkthrough its omitted calls were deliberately used to create the disturbance-only reference. See [baseline plan](evidence/18-baseline-plan.txt).

**Improve it.** Offer explicit operations: repeat complete plan, create a variant, create a nominal reference, or remove responses for comparison. Save name, objective, inputs, event schedule, arguments, watch list, phases and notes together. Show a plan diff before launch.

**Acceptance.** A user can intentionally create all three references without guessing which fields were retained. A later reviewer can recover exactly which inputs and interventions define each run.

### R19 · Next · Capture observations and decisions as part of the investigation

**Opportunity.** Reconstructing why a person stopped, restarted or changed a setpoint is part of understanding the experiment. The recorded command names and times provide a useful skeleton. See [command sequence](evidence/12-review-overview.txt).

**Improve it.** Add notes and bookmarks tied to a time, interval, asset, action or approval. Distinguish authored notes from generated events and sealed truth. Include `Question to investigate`, `Decision and reason`, and `Evidence for handover`. Preserve selection when switching between notes and evidence.

**Acceptance.** A second person can open a note, land on the same evidence, and understand the reason for an action. Authored explanations remain visibly attributed and do not become simulator ground truth.

### R20 · Later · Make completion produce a useful handover

**Opportunity.** Review is an engineering record and can also be the handover point for the user controlling the system. The retained timeline, results, and verification provide its foundation.

**Improve it.** Provide a concise export with purpose, exact model/plan identities, execution outcome, data coverage, selected plots/windows, decisions, command evidence, objective results and open questions. Offer contextual next actions: compare, repeat, change plan, or continue a connected operating workflow where supported. Distinguish ending a simulated experiment from ending observation of physical equipment.

**Acceptance.** A person who did not attend the run can understand it from the exported record and navigate back to supporting evidence. Any unresolved physical state, unavailable assessment or incomplete recording remains explicit.

## What to preserve

- Graph selection already fills a run target and changes Add fault/Add effector call to the selected asset. Extend this into Watch and Inspect actions. See [graph-assisted planning](evidence/21-graph-assisted-plan.txt).
- Event selection already focuses the related asset and cursor. Keep this as the common investigation mechanism.
- The held tower reading includes its observation time. Add age/quality emphasis while retaining that provenance.
- Instrument-range information is exposed on observations. Keep it distinct from a declared operating limit or an alarm condition.
- The replay-integrity explanation is clear. Keep it available alongside the other independent run assessments.
- The plan catches out-of-duration events and identifies the relevant field. Preserve focused guidance as the planning surface grows.

## Recommended sequence

| Slice | User outcome | Improvements |
|---|---|---|
| One coherent run | Purpose, readable identity, chronological plan, watch list, persistent stage/context | R01–R06, R09 |
| Understand the response | Shared trends, quality/freshness, focused plant, timeline windows, evidence scope | R07–R10, R14–R15 |
| Decide and act | Action story, approval context, deliberate execution controls | R11–R13 |
| Explain and reuse | Objective-led review, reference comparison, complete variants, notes and handover | R16–R20 |

Each slice should be reviewed through the factory exercise, as a complete operator task. Keep rules, validation, derived assessment and orchestration in the owning simulator modules. The proposal does not require a frontend architecture rewrite.
