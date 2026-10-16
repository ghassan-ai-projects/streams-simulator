# Canvas comprehension, round 2: meaning, appearance, placement and space

Date: 2026-10-15. Status: review findings and design recommendations; no implementation or new quality score.

## Owner direction

This round follows [the first 25 improvements](COMPREHENSION-IMPROVEMENTS-2026-10-15.md). The owner asked for another round, emphasizing the meaning of things, how they look, where they are on screen, annoying boxes, and wasted desktop space. Validation testing, further dark-mode review, and zoom work were excluded. A brief dark-mode observation preceded that instruction; it is not a deliverable or recommended workstream here. Subsequent inspection used light mode. Nothing below requires zoom changes.

These are 25 further recommendations. Some add new findings; others make earlier proposals more specific. They should be consolidated with round 1, not implemented as 50 unrelated features. The goal is that a person understands the product without memorizing its internal vocabulary or discovering critical meaning by accident.

## Main finding

The interface gives many different concepts similar visual weight: equipment cards, area containers, control groups, capability lists and live conditions. Users must infer distinctions that the screen could state directly. Repeated borders, shadows and header bands make that inference harder by making everything look like a separate box.

Desktop space is allocated responsively, but the composition does not adapt enough to the current task. The main element filled the tested viewport. In the Food cold chain view, the outline stayed about 255 px wide; the inspector grew from 336 px at a 1280 px window to 460 px at 1920 px. The graph region grew from 688 px to 1204 px. These measurements do not support a claim that the whole page has a fixed width or fails to fill the browser. They support a narrower recommendation: improve how the available area is apportioned and how useful content occupies it. Blank space is useful when it establishes hierarchy; it is wasteful when it crowds the active task elsewhere or leaves users unsure what belongs there.

## Walkthrough evidence

Same temporary review workspace and served binary/UI as round 1; no rebuild or source-to-artifact parity check. This round primarily inspected 1440 × 900 light mode, plus desktop allocation at 1280 × 900 and 1920 × 1080. It is an expert walkthrough, not a participant study. No production code or original workspace data was edited in this round.

| ID | What was observed |
| --- | --- |
| S1 | Multi-hall campus: 26 instances, 20 links, five populated areas. Chiller ch-u1 has an existing expandable model explanation, Channels 4, Faults 24, Effectors 3 and States 24. The model explanation is substantial and useful. |
| S2 | Expanded ch-u1 Faults mixes condenser fouling, refrigerant undercharge and compressor wear with measurement offset, drift, dropout and noise faults. Expanded States is a flat list including running, cooling, setpoint, COP, cost rate and measurement biases. Effectors start/stop are plain catalogue text in this view. |
| S3 | After scrolling down the expanded chiller inspector, selecting pmp-u1 updates the content but retains a lower scroll position. The new selected object's heading and description remain off-screen. Expanded section choices also carry across. |
| S4 | Workspace w-44 says Did not finish using a green badge and radio/broadcast icon. DOM inspection confirmed a live-style class, text rgb(15,107,63), and translucent green background. This is an appearance/meaning mismatch, not a new runtime diagnosis. |
| S5 | Food cold chain: 13 instances and 11 links. Graph and outline use abbreviated ahu, chw and cond labels, while the selected ahu-a inspector expands its kind to Air-handling unit. Equipment uses bordered, shadowed cards with tinted header bands, inside bordered area containers. Hiding both side panels leaves the box-heavy graph structure intact. |
| S6 | In Food cold chain editing, Site and Cold rooms show 0 · EMPTY, although Room A and Room B exist beneath the Cold rooms hierarchy and each contains an air-handling unit. The Area selector makes the full path explicit: Site › Cold rooms › Room A. The read-only outline omits these unpopulated parent areas. |
| S7 | Selected ahu-a in view has Links in: chiller-1 supply out → ahu-a chw in. In edit, the same relationship becomes chw in ← chiller-1 supply out under a combined Links section. The creation form offers Link an output. The panel close control is called Close the instance. |
| S8 | Food cold chain Evaluate recording w-29: faults at 00:06:00 and 00:18:00, and a command at 00:22:00. Visible event chips show time and equipment; full fault/command names are in accessible names/tooltips and the report. Clicking the 00:18:00 seal-leak event seeks, selects pump-cond, and highlights the matching report row. Report remains selected; Inspect becomes Inspect · pump-cond. |
| S9 | At 00:18:00, Inspect combines Observed · pump-cond at 00:18:00, an active Fault: seal leak line, observed flow/power and a True results mean/range table. Some Observed at cells contain no visible timestamp. The table's aggregate time scope is not stated beside its heading. |
| S10 | Food cold chain setup says None: the plant runs healthy when no fault is scheduled. This wording is stronger than the setup evidence alone establishes. |
| S11 | At 1920 × 1080, selected-equipment editing still uses a single right-hand form column; the short ahu-a form leaves much of its lower panel unused. The app provides panel show/hide buttons. No obvious resize handle was visible; pointer resizing was not exhaustively tested. The wide editor toolbar has a long gap between status and save/revert actions. |

Existing strengths: precise endpoint labels, current-kind descriptions, a useful model explanation, typed connections, explicit editing state, and coordinated event/outline/graph/report selection. Preserve them. A low-detail graph representation also already exists; round 1's detail-level proposal should be read as refinement of existing behavior, not an absent capability. Zoom is outside this round.

## 25 recommendations

P1 means substantial risk of confusion in a core task; P2 means a useful reduction in reading or navigation effort. “New” means newly articulated in this round, not necessarily a new product feature. Recommendations are not verified defects unless a reproduction is described in the evidence above.

### R2-01. Adapt the desktop composition to the task

**P1 · Refines round 1: 02, 13 · S5, S11**

**Meaning:** Viewing relationships, editing a connection and reading a report have different space needs. **Appearance:** Use a stable shell with task-specific proportions, not a separate design system per page. **Placement:** Give the active form or report enough width while retaining plant context; avoid automatically filling empty space with extra panels.

**Acceptance:** At 1280, 1440 and 1920 px, the current task reads comfortably and nearby context remains visible. Wider screens improve task visibility rather than merely increasing gaps. No zoom work is required.

### R2-02. Make desktop panel widths adjustable and recoverable

**P2 · New spatial proposal · S11**

**Meaning:** Width is a user-controlled reading preference. **Appearance:** Provide a quiet but discoverable separator handle with a keyboard equivalent. **Placement:** Put it on actual panel boundaries; remember reasonable per-task preferences and offer Reset layout. Check existing resizing support before adding anything.

**Acceptance:** Users can make a long connection label or report readable without hiding all context, and can recover the default layout. Resizing does not change selection or silently switch tasks.

### R2-03. Give short inspector content a deliberate layout

**P2 · Refines round 1: 13 · S11**

**Meaning:** A short form should feel complete, not unfinished. **Appearance:** Group identity and editable properties compactly; use only relevant supplementary help. **Placement:** Keep contextual help near the field it explains and reserve a consistent action location. Let users collapse the panel to reclaim width rather than adding filler to its blank lower half.

**Acceptance:** Users recognize that the selected-equipment form is complete and know where to act next. Extra width is available to the task when they dismiss the panel.

### R2-04. Keep task actions together and visually anchored

**P2 · Refines round 1: 15 · S11**

**Meaning:** No changes, pending changes and Save all concern the same editing transaction. **Appearance:** Use one compact change/action group with a clear primary action. **Placement:** Keep change status and Save/Revert close enough to read together on a wide desktop; keep view navigation separate.

**Acceptance:** Users can identify what is pending and where to save without scanning the entire toolbar. Actions stay in predictable positions when status wording changes.

### R2-05. Replace most area boxes with lighter grouping

**P1 · New visual proposal · S1, S5**

**Meaning:** Areas group equipment; they are not equipment themselves. **Appearance:** Prefer a clear area label, whitespace and a subtle separator or background cue over a complete rounded rectangle around every area. **Placement:** Put labels consistently beside their group, with enough separation to distinguish neighboring areas.

**Acceptance:** People can identify Room A, Room B and the refrigeration plant with fewer enclosing borders. Connections remain traceable, and selecting or editing an area still provides an explicit boundary when needed.

### R2-06. Flatten ordinary equipment nodes

**P1 · New visual proposal · S5**

**Meaning:** Equipment is a selectable model object, not a collection of floating dashboard cards. **Appearance:** Remove default drop shadows and reduce routine border weight; preserve strong selected/fault/focus treatments. **Placement:** Keep identity and endpoints aligned so flatter nodes remain recognizable and distinct.

**Acceptance:** Unselected equipment is visually quiet; users still reliably distinguish separate objects, selected equipment and abnormal conditions. Prototype one representative plant before applying this globally.

### R2-07. Reduce repeated header bands and internal framing

**P2 · New visual proposal · S5**

**Meaning:** A component title introduces the component; it does not need another container. **Appearance:** Replace large tinted header bands and repeated internal dividers with a clear type hierarchy and spacing. **Placement:** Put ID and full equipment kind together, then align the secondary port information below.

**Acceptance:** Titles are easy to scan without making every component look like a miniature window. The selected object and status remain clearer than its ordinary header.

### R2-08. Stop boxing every toolbar action

**P2 · New visual proposal · S5, S11**

**Meaning:** Primary actions, navigation, toggles and secondary utilities have different roles. **Appearance:** Reserve filled treatment for primary actions, use quiet text/icon buttons for secondary operations, and use visible toggle states where appropriate. **Placement:** Group controls by purpose with spacing instead of additional enclosing borders.

**Acceptance:** Run or Save is immediately identifiable; users can distinguish navigation from state toggles. Removing frames does not remove hover, keyboard focus, disabled or selected feedback.

### R2-09. Give every persistent colour accent one explainable meaning

**P1 · Refines round 1: 09, 10 · S1, S5**

**Meaning:** Component accents, connection types, selection and abnormal conditions must not imply the same thing. **Appearance:** Use semantic colours consistently and pair state with text or an icon. **Placement:** Explain category accents where they are used; the wire-type legend should not leave users guessing whether equipment stripes mean health.

**Acceptance:** Users can explain a green equipment accent without assuming the equipment is healthy or running. Status remains distinguishable in light mode without colour alone.

### R2-10. Rename the fault catalogue so it cannot look like an incident count

**P1 · New semantic finding · S1–S2**

**Meaning:** Faults 24 in a model inspector counts supported fault definitions. **Appearance:** Label it Available fault types, with a neutral catalogue count. Reserve alarm styling and active counts for actual runtime conditions. **Placement:** Put it under model capabilities, away from current-condition summaries.

**Acceptance:** On opening an untouched chiller, users do not conclude that 24 faults are active. The catalogue remains easy to find when planning an experiment.

### R2-11. Separate equipment failures from measurement disturbances

**P1 · New semantic proposal · S2**

**Meaning:** Compressor wear changes the plant model; sensor offset changes the measurement path. **Appearance:** Group the catalogue as equipment faults and measurement faults, with one plain-language example for each. **Placement:** Keep groups under available faults and show the affected equipment or channel next to the entry.

**Acceptance:** Users can choose between a physical-model failure and a faulty reading intentionally. Group membership comes from authoritative metadata, not guessed keyword classification in the client.

### R2-12. Call model states variables, not operating conditions

**P1 · New semantic finding · S2**

**Meaning:** States 24 mixes running, cooling, COP, cost and biases; it is not a list of current conditions. **Appearance:** Use Model variables, with declared units/descriptions where available. Distinguish flags, inputs and derived quantities only when the model supplies that classification. **Placement:** Put these in advanced model details, below the operational explanation.

**Acceptance:** Users do not infer that running and cooling are both currently true merely because the words appear in the list. Catalogue entries are distinguishable from measured values.

### R2-13. Make available commands look different from executable controls

**P1 · New semantic finding · S2, S5**

**Meaning:** The view inspector's start/stop entries describe supported effectors. **Appearance:** Label them Available commands; render them as descriptions, not ambiguous unlabeled verbs. **Placement:** Provide an explicit route to schedule a command in run setup when supported; keep catalogue reading separate from execution.

**Acceptance:** Users can tell whether clicking or reading start will do anything. An equipment overview never suggests that a plain catalogue entry has already commanded the plant.

### R2-14. Put a one-sentence functional explanation before technical detail

**P2 · Refines round 1: 01, 04, 16 · S1, S5**

**Meaning:** Explain what the selected equipment contributes before describing its equations or variables. **Appearance:** Expand ahu/chw/cond in human-facing labels; retain exact IDs as secondary information. **Placement:** Show a concise role sentence beneath the selected name; retain the useful full model explanation under disclosure.

**Acceptance:** Users unfamiliar with the abbreviations can explain what an air-handling unit or chiller does in this model without reading the entire technical paragraph.

### R2-15. Distinguish no direct equipment from an empty area

**P1 · New semantic finding · S6**

**Meaning:** Cold rooms has no directly assigned equipment but contains Room A and Room B. **Appearance:** Replace 0 · EMPTY with an explicit direct/descendant distinction, or a child-area count. **Placement:** Show it beside the relevant area heading, not as a global warning.

**Acceptance:** Users do not think their rooms or equipment have vanished when a parent has zero direct instances. Count labels have the same scope across views.

### R2-16. Keep area hierarchy recognizable between view and edit

**P1 · New spatial finding · S6**

**Meaning:** Room A belongs to Cold rooms, which belongs to Site. **Appearance:** Preserve a consistent hierarchy/path treatment, with intentional handling of unpopulated parent areas. **Placement:** Use the outline and selected-object location line; avoid making the Area dropdown the only place the full relationship is visible.

**Acceptance:** Moving from view to edit does not make the same plant look as though it acquired new areas. Users can identify an object's parent area without opening a form.

### R2-17. Make incomplete status look incomplete

**P1 · New reproduced visual finding · S4**

**Meaning:** Did not finish is distinct from live, healthy or verified. **Appearance:** Replace the green live/broadcast badge with an appropriate neutral or warning treatment and an incomplete symbol. **Placement:** Use the same semantic status in workspace rows and review headers.

**Acceptance:** People can distinguish running, incomplete and verified recordings at a glance and explain each state correctly from its wording and icon.

### R2-18. Separate run mode from current evidence availability

**P1 · Refines round 1: 17 · S8–S9**

**Meaning:** Evaluate describes a run mode; it does not necessarily mean its now-completed recording is still sealed. **Appearance:** Separate the mode label from visible/sealed availability and give lock imagery a precise scope. **Placement:** Put current evidence availability near the data it governs and mode beside the run identity.

**Acceptance:** A completed Evaluate review showing true results does not appear contradictory because a generic lock remains beside its mode name.

### R2-19. Keep selected-object identity visible while reading details

**P1 · New reproduced orientation finding · S3**

**Meaning:** All inspector details belong to one selected object. **Appearance:** Keep a compact sticky name/kind/location header. **Placement:** On changing objects, either return the detail pane to its start or preserve the selected subsection with that heading visible. Preserve focus deliberately rather than allowing scroll position to define the context.

**Acceptance:** Reproduce S3: expand chiller details, scroll down, select a pump. The new pump's identity remains visible immediately; users never have to infer it from a remote outline highlight.

### R2-20. Read every connection in the same source-to-receiver order

**P1 · New cross-screen inconsistency · S7**

**Meaning:** The model relationship is unchanged by entering edit mode. **Appearance:** Keep source output → receiver input order in both screens, with Inputs and Outputs groups where useful. **Placement:** When starting from a receiver, provide a clear Connect an input route or a direct route to the source's connection tool.

**Acceptance:** Users can match a viewed relationship to its editable representation without mentally reversing the arrow. Keyboard users can begin a connection from either conceptual endpoint.

### R2-21. Name panel-back actions by their destination

**P2 · Refines round 1: 03, 13 · S7**

**Meaning:** Closing selected-instance details returns to build tools; hiding the panel is a different action. **Appearance:** Use an explicit back label or understandable tooltip instead of relying on the same X shape for every operation. **Placement:** Put Back to build tools beside the selected-object heading, separate from the shell's show/hide control.

**Acceptance:** Users can predict whether the action clears selection, returns to the library, or hides the entire panel before activating it.

### R2-22. Put the event's meaning in its visible timeline label

**P1 · Refines round 1: 22 · S8**

**Meaning:** 00:18:00 pump-cond is not enough to explain what happened. **Appearance:** Include a short event description such as Seal leak, with full detail available on focus. Distinguish the selected event from an event currently active at the playhead. **Placement:** Show the label with its marker or in a nearby selected-event summary; preserve existing linked report-row highlighting.

**Acceptance:** Users understand the event without hovering or searching the report. They can distinguish a future event, a past event and a condition still active now.

### R2-23. Label snapshot readings and whole-run summaries separately

**P1 · Refines round 1: 24 · S9**

**Meaning:** Observed at 00:18:00 and a whole-run mean answer different questions. **Appearance:** Give each a concise time-scope label; use a visible timestamp or explicit same-as-playhead indicator where cells are currently blank. **Placement:** Put scope beside the relevant table heading, not only in a distant playback clock.

**Acceptance:** Users do not interpret a run-average KPI as the value at the current playhead, or a blank timestamp as evidence that a reading has no time.

### R2-24. Separate scenario facts from observations

**P1 · New semantic concern · S9**

**Meaning:** A scheduled/known fault and a sensor observation are different sources of evidence. **Appearance:** Label the provenance of Fault: seal leak where it appears beneath Observed; do not imply that the external consumer detected it unless that is recorded. **Placement:** Put scenario conditions and measured readings in clearly named adjacent subsections, respecting the existing truth boundary.

**Acceptance:** Users can state whether a fault label comes from the scenario, observed alarm data, or an outside system's verdict. This is a presentation proposal, not a claim of live truth leakage.

### R2-25. Say no scheduled faults without promising plant health

**P1 · New wording finding · S10**

**Meaning:** No injected fault is planned; it is not a health verdict. **Appearance:** Replace None: the plant runs healthy with No faults scheduled, plus a short explanation if needed. **Placement:** Keep it inside the fault schedule; show actual outcomes only in evidence-backed running/review sections.

**Acceptance:** Users understand that an unperturbed plan can still have interesting operating behavior and that health must be interpreted from the relevant evidence.

## First design direction to review

Make one light-mode representative plant visibly flatter: lighter area grouping, quiet equipment surfaces, fewer header bands and fewer boxed toolbar controls. Keep equipment identity, source-to-receiver direction, keyboard focus, selection and abnormal status explicit. Alongside it, prototype a compact inspector header and an adjustable task panel. Do not add dashboards or decorative cards to occupy blank areas. Do not alter zoom as part of this work.

Start with R2-10, 12, 13, 15, 17, 19, 24 and 25 for semantic errors and orientation; prototype R2-01–09 together on one plant for spatial/visual clarity. Then refine hierarchy, connection editing and recorded-event interpretation. These are suggestions, not authorization to change contracts or simulation semantics. Runtime classifications and truth provenance must come from their owning module rather than guessed rules in the UI.

Future comprehension review should ask people to explain: what Faults 24 counts; whether start is a live command; why Cold rooms says zero; which object owns the details after switching selection; what the green incomplete badge means; and whether a KPI is instantaneous or aggregated. Compare the flatter layout against the current one using those tasks. No new 9–10 rating is earned until people can answer accurately and complete the flows with fewer interruptions.
