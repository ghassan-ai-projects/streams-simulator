# Canvas comprehension: 25 improvements toward a 9–10 experience

Date: 2026-10-15. Status: design proposals informed by direct Chrome review; not implementation approval or completed work.

## Outcome

A person opening an unfamiliar plant should be able to answer five questions without a guide:

- Where am I, and am I viewing, editing, running, or replaying?
- What does this equipment do, and how is it connected to the rest of the model?
- What am I about to change, and when will that change take effect?
- What happened, what evidence supports it, and what remains unknown?
- What is the next useful action, and how do I get back?

The owner clarified that comprehension and avoiding confusion are the priority. These 25 suggestions therefore emphasize orientation, meaning, and continuity. Colour, typography, layout, components, and motion should reinforce those goals. Completing a feature checklist does not guarantee a 9–10 score: the resulting experience needs fresh review and observation of users doing real tasks.

## Evidence and boundaries

The review used the installed Chrome browser and a fresh local bridge on port 55231, running the existing `bin/streams-simulator` and `canvas/dist`. It used a temporary copy of `streams-simulator-review-lab`; authoring and new runs affected that copy. Source HEAD at the deeper pass was `3520ff3`, with a clean worktree. The binary and UI were not rebuilt, so this is evidence about the served build, not proof that those artifacts match that commit.

The initial pass sampled light/dark at 1280, 1440, and 1920 desktop widths; it was not a complete screen-by-theme matrix. The deeper comprehension pass concentrated on 1440 × 900. Validation testing was explicitly excluded at the owner's request. No claim is made about conflict recovery, full keyboard coverage, accessibility compliance, large-campus performance, or consumer integration.

| Evidence | Directly inspected or exercised |
| --- | --- |
| E1 | Workspace with 14 plants and 45 existing runs; Open/Edit/Run entry points, run filter, new-plant form, and incomplete-run row. |
| E2 | Boiler steam loop: 14 instances, 10 links; outline selection, inspector, connected-edge emphasis, zoom, and legend. The legend already explains output-to-input direction and uses both colour and pin shapes. |
| E3 | Outline search for `pump`: three matching entries and a match-count announcement; the graph continued to show the full plant without obvious search-result emphasis in the capture. |
| E4 | Editor: equipment library, add pump and undo; selected equipment rename/area/link controls; compatible output/receiver/input choices; creation of `pump-feed flow out → hx-1 hot flow in`, success text and link count, followed by undo. Drag-to-connect was not tested. |
| E5 | Created a full-copy plant named Comprehension review in the temporary workspace; renamed pump-feed to feed-pump-review; Save showed “Saving and composing…” then “Ready to run. Saved.” Reload retained the renamed equipment. |
| E6 | Explore run w-46: 30 minutes, seed 1, boiler flame failure scheduled at 00:05:00; watched progress and graph fault indication, then opened the recorded review. |
| E7 | w-46 playback: keyboard seek from 00:08:35 to 00:08:36, play/pause, Jump to alarm selecting boiler-1 at 00:08:35, observed readings, alarm status, report and true KPI tables. Fault end was listed at 00:05:10. |
| E8 | Explore setup says truth is visible after the run; Explore report says true results were visible while it ran; the same report's generic true-results explanation says an outside system could not see them while live. This is a wording ambiguity, not a demonstrated truth-boundary failure. |
| E9 | Existing incomplete w-44: Did not finish, “Not checked,” absent run.json explanation, and unavailable true results. No recovery action was exercised. |
| E10 | Evaluate setup and completed w-47: set pressure setpoint to 1100 at minute 5, fastest execution. Parameter label was “pressure kpa,” range placeholder 500 to 1300. Completion said accepted; report said applied · 251 ms without showing 1100 in that outcome table. Report also said no outside system was connected and the run was not scored. Evaluate live blindness was not tested. |

Existing strengths to preserve: the stable shell, explicit editing badge, useful component descriptions, direction arrows, typed legend, exact endpoint lists, compatible connection choices, undo, save feedback, and linked alarm navigation. Extend them instead of replacing them with a new navigation system. Prior scoped comprehension work is recorded in [the connection comprehension review](desktop-polish-20261014/round2-comprehension.md).

Priority: **P1** removes substantial ambiguity in a core flow; **P2** reduces effort or improves explanation. “Observed” identifies a concrete friction in this walkthrough. “Proposal” is an extension worth testing, not a claim that a feature is absent everywhere.

## The 25 improvements

### 01. Give an unfamiliar plant a short introduction

**P1 · View/workspace · Proposal grounded in E1–E2 · Content, layout, comprehension**

The default plant inspector leads with inventory counts. Add an optional, concise plant overview: what is modeled, the main equipment groups, important inputs/outputs, and a suggested first component to inspect. Use authored domain information; leave missing explanations explicitly unavailable rather than inventing a process story.

**Acceptance:** A first-time user can describe the plant's purpose and identify a sensible starting point in their own words within one minute. They can dismiss the introduction and reopen it later.

### 02. Name the current activity consistently

**P1 · All screens · Proposal grounded in E2, E4, E6–E7 · Coherence, feedback, layout**

Keep breadcrumbs, but make the current activity explicit: View plant, Edit plant, Run setup, Running, or Review recording. Distinguish the activity from Explore/Evaluate, which describes the run's evidence visibility. Reuse a small status component across screens rather than adding another navigation bar.

**Acceptance:** At any sampled screen, users can identify whether a control edits the plant, schedules a future run, or seeks an existing recording. They do not confuse playback with a live run.

### 03. Preserve a visible route back to the user's context

**P2 · View/edit/review · Proposal grounded in E2, E4, E7 · Coherence, motion, task efficiency**

When moving between plant, editor, and review, retain compatible selection and camera context. Label transitions that intentionally change context, especially opening the historical plant recorded with a run. Offer a clear back action after following a connection into another area. Reuse existing historical-plant notices where available.

**Acceptance:** After following two linked components, a user can return to the original selection without hunting. Historical and current plants are distinguishable without comparing their equipment manually.

### 04. Explain domain terms exactly where they become relevant

**P1 · All screens · Observed terminology burden in E2, E4, E10 · Content, typography, coherence**

Provide short contextual definitions for instance, port, channel, effector, seed, truth, and verification. Prefer familiar labels such as “Scheduled commands” with “effector calls” as secondary vocabulary. Preserve exact technical identifiers in inspectable details for engineers.

**Acceptance:** A novice can distinguish an equipment instance, its port, a sensor reading, and a command without leaving the task. Definitions use one consistent meaning across view, edit, and run.

### 05. Make the workspace answer “which plant or run do I need?”

**P2 · Workspace · Proposal grounded in E1 · Layout, components, task efficiency**

Add a focused search across plant names/IDs and run IDs, and useful run filters such as mode and completion state. Show plant identity, run time, and status in a stable order. Keep filters visible, show result counts, and provide one clear reset.

**Acceptance:** Users can find an unfamiliar run from its plant and approximate date without scanning every row, and can explain why a filtered list is empty.

### 06. Show what a starting plant contains before copying it

**P2 · New plant · Observed choice burden in E1, E5 · Components, content, comprehension**

The Start from selector mixes existing workspace plants with shipped choices. Add a compact preview with source, purpose, area/equipment counts, and what a full copy versus empty plant retains. Explain profiles and axes in the empty-copy option instead of assuming prior knowledge.

**Acceptance:** Before creating a plant, users can predict what equipment, structure, and settings will appear. They can distinguish a shipped starting example from a colleague's workspace plant.

### 07. Keep graph meaning readable as zoom changes

**P1 · View/review · Observed small labels in E2, E7 · Typography, layout, craft**

Use progressive detail: at overview show equipment identity, kind, and significant status; at closer zoom reveal ports and selected readings. Keep selection details readable outside the scaled graph. Avoid fitting every small label onto an overview at the expense of legibility.

**Acceptance:** At 1280 and 1440 widths, users can locate named equipment and identify its status without reading miniature port labels. Zooming in exposes detail without losing their place.

### 08. Explain a connection as a model relationship

**P1 · View/edit · Proposal extending existing legend in E2, E4 · Graph information design, content**

Keep the current arrowheads and typed legend. On selecting a connection, state the exact source, value/unit, receiver, and declared role: for example, a pressure output supplies a valve's pressure input. Explain that a model dependency is not necessarily a physical pipe. Do not infer physical topology from graph placement.

**Acceptance:** Users can tell which endpoint supplies the value and which receives it, and can explain the relationship without relying only on wire colour.

### 09. Offer a focused path through connected equipment

**P2 · View/review · Proposal extending selected-edge emphasis in E2, E7 · Graph information design, colour, motion**

Build on existing incident-edge highlighting with an optional “Follow inputs/outputs” mode. Keep the selected component and immediate neighbors prominent, show a short named path, and provide a single return to overview. Follow documented links; do not label a connected path as proven causal influence.

**Acceptance:** Users can trace a requested source-to-receiver path and return to the full plant without remembering intermediate IDs. Unrelated equipment stays available but visually secondary.

### 10. Connect outline search to what users see on the graph

**P1 · View/edit/review · Observed in E3 · Feedback, colour, coherence**

The outline correctly reports three pump matches, but the captured graph gives no obvious corresponding search emphasis. Highlight matching nodes and optionally fit those results while retaining surrounding context. Distinguish search matches from the one selected item and from alarms.

**Acceptance:** Typing pump immediately makes the matching graph equipment recognizable. Clearing the query restores the previous overview; an empty result gives an explicit explanation and reset.

### 11. Explain missing readings instead of relying on punctuation

**P1 · Review · Observed in E7, E9 · Content, feedback, typography**

The graph uses different compact representations for not observed and no reading yet. Make those states discoverable and distinguish them from zero, stale data, disconnected ports, and unavailable truth. Use the available evidence to name the reason; never fabricate a missing measurement or silently substitute true values.

**Acceptance:** Users can explain why a value is absent and distinguish that absence from a valid zero. The same state has the same wording and visual treatment in the graph and inspector.

### 12. Pair equipment IDs with understandable roles

**P2 · View/edit · Proposal grounded in E2, E5 · Typography, content, craft**

Keep stable IDs, but provide an optional human-readable role or description beside them: equipment kind plus authored purpose. Expose full names on focus/selection when truncated. Use existing metadata first; any new persisted description field needs a deliberate contract design.

**Acceptance:** Someone unfamiliar with the naming scheme can tell two pumps apart by purpose while still locating their exact IDs in links and reports.

### 13. Organize editing around the user's current task

**P1 · Edit · Observed long mixed panel in E4–E5 · Layout, components, coherence**

Separate Add equipment, Edit selected equipment, Connections, and Areas into clearly named sections using the existing side panel. Keep the current task and selected item visible. Move repeated syntax instructions into short help disclosures while retaining actionable hints next to relevant controls.

**Acceptance:** Users can find how to add a pump, rename it, or connect it without scrolling through unrelated area-management controls. Closing equipment details clearly returns to the build choices.

### 14. Preview the meaning of a link before creating it

**P1 · Edit · Proposal extending working compatible choices in E4 · Feedback, graph information design, craft**

Retain the filtered output/receiver/input choices. Show a sentence and temporary graph preview before Link: source output → receiver input, with quantity and unit when known. Explain why a receiver remains selectable or why no compatible destination exists. After linking, keep the new relationship visibly selected long enough to inspect it.

**Acceptance:** Users can predict the exact endpoints and value type before committing. After creating a link, they can identify it on the graph and find the same relationship in the selected equipment's link list.

### 15. Explain the pending changes and what Save does

**P1 · Edit · Proposal extending successful E4–E5 feedback · Feedback, content, trust**

Keep Unsaved changes, Undo, Revert, and the existing saved confirmation. Add a compact change summary such as “Renamed 1 instance; added 1 connection.” Explain that Save updates the composition and runnable plant, while existing recordings retain their recorded plant. Make undo/revert scope clear.

**Acceptance:** Before Save, users can state what will change and whether old run reviews will change. After Save, they can distinguish success from work that is still pending.

### 16. Let users compare equipment choices by meaning

**P2 · Edit/library · Proposal extending descriptions and sizes in E4 · Components, content, comprehension**

The library already provides descriptions and size choices. Present a compact selected-equipment preview with purpose, important inputs/outputs, and what the chosen size changes. Keep engineering details behind disclosure; avoid replacing useful descriptions with decorative cards.

**Acceptance:** Users can choose between plausible equipment or sizes based on the intended model behavior, rather than memorizing names or trying each one.

### 17. Explain Explore/Evaluate and whether scoring is actually available

**P1 · Run/review · Observed ambiguity in E8, E10 · Content, coherence, trust**

Use one mode-specific explanation covering who can see what, when, and why. Resolve the conflicting Explore text. Show separately whether an outside system is connected and whether a score can be produced; choosing Evaluate alone should not imply that a scored experiment is configured. Present this as status/help, not a new validation exercise.

**Acceptance:** Before starting, users can explain what will be hidden, from whom, when it becomes visible, and whether this particular run will produce a score. The review repeats the same meaning.

### 18. Make simulated time and waiting time unmistakable

**P1 · Run/review · Proposal grounded in E6, E10 · Typography, feedback, content**

Keep plant duration, elapsed wall time, execution pacing, and playback speed distinctly labeled. Replace compressed “s per s” language with an understandable equivalent plus optional technical detail. Label every scheduled event in plant time.

**Acceptance:** Users know that a fault at minute 5 is five simulated minutes after run start, not five minutes of waiting. They can distinguish run execution speed from replay speed.

### 19. Preview the whole planned experiment before starting

**P1 · Run setup · Proposal grounded in E6, E10 · Layout, feedback, graph information design**

Show a compact chronological preview of faults and commands, with equipment names, times, and entered arguments. Keep the existing start summary, but make it specific enough to confirm the intended experiment without reopening every form row. Preserve an easy route back to edit a scheduled item.

**Acceptance:** Users can accurately describe the run's sequence before starting and notice an unintended time or target by reading the preview.

### 20. Describe command parameters and intended effects plainly

**P1 · Run setup · Observed in E10 · Content, typography, components**

Replace mechanically humanized labels such as pressure kpa with “Pressure setpoint (kPa).” Keep the allowed range visible after typing, not only as a placeholder. Explain the command's declared purpose and prerequisites from domain metadata; distinguish requesting a setpoint from guaranteeing a measured pressure.

**Acceptance:** Users can state what 1100 means, its unit, and what the command requests. They do not assume entering a setpoint guarantees that the process will attain it.

### 21. Make the running screen explain meaningful changes

**P2 · Running · Proposal extending progress and fault badges in E6 · Motion, feedback, colour**

Add a small recent-event strip beside progress: the scheduled fault or command occurred, which equipment it concerns, and where to inspect it. Distinguish active from ended disturbances. Use brief transitions to draw attention while keeping layout stable; respect reduced motion. Never expose forbidden truth in Evaluate mode.

**Acceptance:** A user watching a run can tell what has happened so far and what remains scheduled without memorizing setup. Events remain understandable without animation or colour alone.

### 22. Open review with an evidence-linked result summary

**P1 · Review · Observed table-first explanation burden in E7, E10 · Graph/review information design, layout, content**

Lead with completion, verification scope, scoring availability, important recorded events, and remaining active alarms. Each item should seek/select its evidence. In w-46, show the flame failure at 00:05:00, its end at 00:05:10, and the alarm at 00:08:35 as a sequence; do not assert causation merely because events are ordered.

**Acceptance:** Within one minute, users can describe the main recorded sequence and open supporting evidence. They can distinguish a verified recording from a successful controller or a healthy plant.

### 23. Connect events to synchronized channel trends

**P1 · Review · Proposal grounded in snapshot-only inspection in E7 · Graph/review information design, motion, typography**

Offer an explicit route from a selected observed channel to its time trend, synchronized with playback and event markers. Label axes and units, preserve missing intervals, and mark observation time separately from the playhead. Keep observed and true series distinct, exposing truth only when allowed.

**Acceptance:** Users can inspect how steam pressure changed around the flame failure and alarm without seeking repeatedly and remembering individual numbers. The plotted points retain their evidence timestamps.

### 24. Explain outcomes and KPIs without overstating evidence

**P1 · Review · Observed in E7, E10 · Content, coherence, trust**

Carry the requested command arguments into the review next to accepted/applied/refused outcomes and explain those stages. State what the displayed latency measures. For KPIs, explain mean and range, the aggregation interval, and direction or target only when declared. Explain whether a displayed range is observed min–max or an acceptable operating band. A simulator's applied command is not proof of a physical effect.

**Acceptance:** Users can retrieve “requested 1100 kPa at 00:05:00” from the outcome and distinguish it from measured pressure. They do not interpret recording verification, command acceptance, or a KPI range as an automatic pass verdict.

### 25. Give incomplete or interrupted work a clear next step

**P1 · Run/review/edit · Observed incomplete explanation burden in E9; extension proposal elsewhere · Feedback, components, content**

Explain what is known, what is unavailable, and what the user can do next: inspect partial evidence, return to the plant, or set up another run. Distinguish a run still executing from one that did not finish when the runtime can establish that distinction. Keep technical file details available under disclosure. Preserve existing safeguards around unsaved work; their exhaustive behavior was not tested here.

**Acceptance:** After encountering an incomplete run, users can explain why true results or scoring are unavailable and take a useful next action without a terminal or guessing whether the product is still working.

## Dimension coverage

The original sampled scores remain a baseline, not a new score for unimplemented proposals.

| Dimension | Baseline /10 | Main suggestions | What a 9–10 experience must demonstrate |
| --- | ---: | --- | --- |
| Components | 8 | 05, 06, 13, 16, 20, 25 | Controls expose their purpose, current state, and next step consistently. |
| Motion | 7 | 03, 09, 21, 23 | Transitions preserve orientation; temporal changes are understandable with reduced motion. |
| Colour | 7.5 | 09, 10, 21 | Search, selection, alarms, and evidence states remain distinct without colour alone. |
| Layout | 7 | 01, 02, 05, 07, 13, 19, 22 | The current task and its relevant explanation are visible without competing panels. |
| Content | 6.5 | 01, 04, 08, 11, 17, 18, 20, 24 | Users explain the model and evidence correctly using the UI's words. |
| Feedback | 8 | 02, 10, 11, 14, 15, 18, 19, 21, 25 | Users can tell what changed, when, whether it completed, and what happens next. |
| Craft | 7 | 07, 12, 14 | Fine detail supports reading and locating information at every supported desktop width. |
| Coherence | 8 | 02, 03, 04, 10, 13, 17, 24 | The same object, mode, and outcome retain the same meaning across the whole workflow. |
| Typography | 6.5 | 04, 07, 11, 12, 18, 20, 23 | IDs, values, units, and time remain readable and distinguishable in context. |
| Graph and review information design | 6.5 | 08, 09, 14, 19, 22, 23 | Users trace declared relationships and explain recorded sequences with direct evidence. |

## Suggested delivery order and comprehension review

Start with 02, 04, 11, 17, 18, 20, and 24: they reduce misleading or ambiguous interpretation across existing screens. Then address graph orientation and editing clarity with 07–10 and 13–15. Add experiment/review explanations with 19, 21–23, and 25. Refine discovery and onboarding with 01, 03, 05, 06, 12, and 16. This is a sequence of small changes, not a mandate for a new framework or simultaneous redesign.

For each round, observe representative first-time and returning users at the supported desktop widths and both themes. Ask them to explain what they see rather than merely click a prescribed button. Record wrong interpretations, requests for help, time spent finding the next step, and whether users can cite supporting evidence. Proposed initial targets: at least four of five representative participants complete each core scenario without moderator direction, and none confuse playback with live execution, observed with true values, or verification with system success. These are future review targets, not measurements from this agent walkthrough.

Keep the existing [canvas acceptance bar](../../governance/CANVAS_UI_ACCEPTANCE.md). Any new runtime-derived summary, type rule, aggregation, truth access, or outcome interpretation belongs in the binary's owning module and is exposed through its existing boundaries. Canvas should render and navigate evidence, not invent simulation results. No implementation, validation test suite, or production release qualification is claimed by this document.
