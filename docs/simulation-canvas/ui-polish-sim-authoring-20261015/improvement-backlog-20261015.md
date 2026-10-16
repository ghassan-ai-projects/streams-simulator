# UI/UX improvement backlog — the consolidated work list

**Date:** 2026-10-15. **Single consolidated work list** for UI/UX improvement of the canvas. The three source
documents remain the detailed evidence and are not replaced by this one; every entry below points back to them.

- **Sources consolidated here**
  1. `improvement-hunt-20261015.md` — the improvement hunt, items `IMP-01`…`IMP-22` under `## Improvements`, plus its
     seven numbered `## Bigger bets` (cited below as `BB-1`…`BB-7`; the labels are this document's, for traceability —
     the hunt numbers them 1–7).
  2. `docs/simulation-canvas/COMPREHENSION-IMPROVEMENTS-2026-10-15.md` — comprehension round 1, 25 recommendations
     (cited as `r1-01`…`r1-25`, matching its `### 01.`…`### 25.`).
  3. `docs/simulation-canvas/COMPREHENSION-ROUND-2-2026-10-15.md` — comprehension round 2, 25 recommendations
     (`R2-01`…`R2-25`).
- **Provenance, and its limit.** All three streams are dated 2026-10-15 and all three were produced against the same
  served bundle: `canvas/dist` **as built at 22:16**, `assets/index-CS8g2xLX.js` (660,743 B) and
  `assets/index-Di5IAvBm.css` (82,977 B).
  - The hunt measured the **served bundle**, at HEAD `0a477da` when it started and `3520ff3` at hand-off; it declares
    that the served bundle never changed while it ran, and that `a13c028`'s empty-editor step was therefore **in source
    but not on screen**.
  - The two comprehension documents used an **unrebuilt binary plus `canvas/dist`** while four local commits
    (`a13c028`, `969531d`, `6921832`, `3520ff3`) sat in the tree; round 1 records source HEAD `3520ff3` with a clean
    worktree, and round 2 records "no rebuild or source-to-artifact parity check".
  - **Therefore no claim in any stream is proof about the current HEAD.** HEAD is now
    `bfabd05` ("the run plan scrolls its fields above a fixed Start footer"), the commits above it are committed but
    were never in the 22:16 bundle, and `canvas/dist` has since been rebuilt (23:47, different asset hashes), so the
    bundle every stream measured is no longer in the tree. Where a proposal's premise could be settled by reading the
    source, this document says so and names the file and line; where it could not, the entry attributes the observation
    to the stream that measured it and marks it as that stream's claim.
- **Rule.** This is a work list, not a mandate. **Nothing here changes behaviour without the owner's word.** Entries
  marked `accepted` are coherent with the recorded decisions and inside the standing bar, so they need no new decision
  to start; they are not owner-approved. Entries marked `owner decision` cannot be taken up without the owner choosing.
- **Standing bar:** `governance/CANVAS_UI_ACCEPTANCE.md` §3 (and §4 for budgets); the numbered expectations and their
  status marks are `docs/simulation-canvas/UI-EXPECTATIONS.md`.

## How to read it

**IDs.** `CB-NN` = consolidated backlog item. The id is stable; numbering runs in dimension order and is this
document's own. `IMP-NN`/`BB-N` are the hunt's items, `r1-NN` comprehension round 1, `R2-NN` comprehension round 2.

**Class** (what kind of change the entry is):

| Class | Means |
|---|---|
| `ui-presentation` | Look, copy, markup or CSS. Same facts, same behaviour. |
| `ux-polish` | A small interaction or feedback change that does not add a capability. |
| `behaviour-change` | The app would start or stop doing something it does today. |
| `feature-or-binary` | New surface or interaction, or a fact the canvas does not have (binary-owned). |
| `owner-design-decision` | A product/taste decision the owner must make before anyone implements. |
| `already-fixed` | The recorded defect no longer describes the tree; the entry is the record that it is done. |
| `rejected` | The premise does not hold on the source; recorded here so it is not re-proposed. |

**Status:**

| Status | Means |
|---|---|
| `in flight` | In another agent's hands now: uncommitted in the working tree, or declared in flight by the delegating agent. |
| `accepted` | No new owner decision is needed to start; inside the standing bar and consistent with the recorded decisions. |
| `owner decision` | Blocked on an owner choice (it opposes a recorded decision, or the trade-off is a product call). |
| `done — <commit>` | Landed at that commit. |
| `stale` | The recorded observation no longer describes the tree (already satisfied, or the premise is false). |

**Effort** S (one file, one sitting) · M (a few files or a component plus tests) · L (a workstream).
**Impact** low (a scan or reading cost) · medium (a wrong reading or a lost action) · high (blocks a core flow).
**Needs binary**: `yes` when a fact the canvas cannot compute is required; the parenthetical says which fact.

**Evidence rule.** `Observed` carries only what a stream measured or what the source states, attributed to the stream
or to the file and line. No claim is upgraded here.

## The consolidated list

### Components

#### CB-01 The equipment chooser hides most of the library with no cue
- **Class** `ui-presentation` · **Dimension** Components · **Effort** S · **Impact** medium · **Needs binary** no · **Status** `accepted`
- **Sources** IMP-01. (Hunt expectation rows E-3 and N-6 cover the same ground as evidence.)
- **Where** Editor → Build panel → "Add equipment", the `select[aria-label="Equipment"]` listbox.
- **Observed** Hunt: a native multi-row listbox, rect `1079, 167, 349 × 142`, `size="6"`, so 6 of 13 campus components (6 of 20 on the empty editor) paint at once, `scrollHeight 448` against `clientHeight 140` (≈ 8 rows hidden); no border of its own and macOS paints its overlay scrollbar only while scrolling. Source today: `AddEquipment.tsx:132` `size={Math.min(6, Math.max(2, found.length))}`, `.library select[size]` at `Editor.module.css:257` sets `height: auto; padding` and nothing else.
- **Suggestion** Give the listbox a visible 1 px boundary and label it with its size ("13 components") so a partial view reads as partial. IMP-01 also offers raising `size` to fit the library; **do not** — that is the smaller-looking change but it opposes `d8a8549`.
- **Conflicts** `d8a8549` ("its library list shows six rows", deliberate).

#### CB-02 (rejected) The keyboard dialog does not dim the page
- **Class** `rejected` · **Dimension** Components · **Effort** — · **Impact** — · **Needs binary** no · **Status** `stale`
- **Sources** IMP-02.
- **Where** The Keyboard dialog, every screen.
- **Observed** Hunt: `getComputedStyle(dialog, "::backdrop").backgroundColor` is `rgba(0, 0, 0, 0)` on a plant and on a review. Source today: the dim is a **sibling** `<div className={styles.scrim}>` (`KeyHelp.tsx`, rendered only while open) with `background: var(--scrim)` (`KeyHelp.module.css:1-6`); `--scrim` is `rgb(15 23 42 / 32%)` in light and `rgb(0 0 0 / 55%)` in dark (`tokens.css:75, 132`), added by `4a3d258`. The dialog is opened with the `open` attribute, not `showModal()`, so the native `::backdrop` paints nothing.
- **Suggestion** None. The hunt measured the wrong element; G-15's "no scrim" clause is already satisfied. If the dim is thought too weak, that is a token change (`--scrim`), not a backdrop rule.
- **Conflicts** Rounds 13 and 14 both describe the dialog as having "a scrim" — they are right about the page and the hunt is right about the pseudo-element (see Coherence check).

#### CB-03 The keyboard dialog's first focus is invisible
- **Class** `ui-presentation` · **Dimension** Components · **Effort** S · **Impact** low-medium · **Needs binary** no · **Status** `accepted`
- **Sources** IMP-03.
- **Where** The Keyboard dialog on open, plant and review.
- **Observed** Hunt: `document.activeElement` on open is the H2 "Keyboard" and its computed outline is `rgb(20, 26, 34) none 3px`, no box shadow; every top-bar tab stop shows a 2 px `rgb(37, 99, 235)` ring. Source today: `KeyHelp.tsx` focuses the title ref on open and `KeyHelp.module.css` sets `.title:focus { outline: none }` explicitly. G-15 names focus on Close; G-8 and §3.3 require visible focus.
- **Suggestion** Move the initial focus to the dialog's Close button (the first actionable control and the one G-15 names), or ring the focused heading like every other stop. Keep Escape returning focus to the help button.
- **Conflicts** none.

#### CB-04 The three graph zoom controls are the only icon-only controls without a tooltip
- **Class** `ui-presentation` · **Dimension** Components · **Effort** S · **Impact** low · **Needs binary** no · **Status** `in flight`
- **Sources** IMP-04.
- **Where** Graph bar on the plant view, the editor and the review (Zoom out, Zoom in, Fit the plant in view).
- **Observed** Hunt: all three carry `aria-label` and a 14 px lucide icon and `title === null`, while every other icon-only control on the same rows (playback steps, alarm jump) carries both. Source at HEAD before the working-tree edit: `GraphControls.tsx:154-164`, `aria-label` only.
- **Suggestion** Set `title` from the existing `aria-label` on those three buttons. **This is already in the working tree uncommitted**: `GraphControls.tsx` now passes `title="Zoom out" / "Zoom in" / "Fit the plant in view"`, with `GraphControls.test.tsx` updated. Do not duplicate it.
- **Conflicts** none.

#### CB-05 The New plant flow does not show what a starting plant contains
- **Class** `feature-or-binary` · **Dimension** Components · **Effort** M · **Impact** medium · **Needs binary** yes (the purpose of a shipped starting choice is authored data the canvas does not load) · **Status** `accepted`
- **Sources** r1-06. (Hunt expectation rows N-2 and N-3 cover the same ground as evidence.)
- **Where** Home → New plant → the "Start from" selector.
- **Observed** r1-06: the selector mixes existing workspace plants with shipped choices with no preview; W-5/N-2 record that the flow navigates straight into the editor with focus in a field rather than showing a named creation form, and N-3 records that the form says it copies site, areas, profiles and axes with no full-copy choice. The hunt reached New plant but did not submit it.
- **Suggestion** A compact preview before creation: source, purpose, area and equipment counts, and what a full copy keeps versus an empty plant. The counts are derivable from the composition the canvas already loads; the purpose sentence is authored data, so leave it explicitly unavailable rather than inventing it.
- **Conflicts** none.

#### CB-06 The equipment chooser gives no purpose or size meaning for a choice
- **Class** `ui-presentation` · **Dimension** Components · **Effort** M · **Impact** medium · **Needs binary** yes (what a preset size changes about the model is not in the canvas payload) · **Status** `accepted`
- **Sources** r1-16. (Hunt expectation rows E-3 and N-6 cover the same ground as evidence.)
- **Where** Editor → Build panel → Add equipment (library choice, Size select).
- **Observed** r1-16: the library already provides descriptions and size choices. Source today: `AddEquipment.tsx` renders `presets.description` in a `.describe` block once a component is chosen, and the Size select is `[["", "Default"], ...presets.options]` — titles only, so "Default" is all that reads until a component is picked (E-3, hunt). The hunt measured no description beside a library *choice*.
- **Suggestion** Extend the existing `.describe` block into a compact selected-equipment preview: the description already shown, plus important inputs/outputs from the specs already loaded and what the chosen size changes. Keep engineering detail behind disclosure; do not replace the description with decorative cards. The ports half is available now; the size-effect half is binary-owned.
- **Conflicts** none.

#### CB-07 The fault catalogue reads like an incident count
- **Class** `ui-presentation` · **Dimension** Components · **Effort** S · **Impact** medium · **Needs binary** no · **Status** `accepted`
- **Sources** R2-10.
- **Where** Plant view → instance inspector → the "Faults 24" block (round 2's S1–S2, chiller `ch-u1`).
- **Observed** R2-10: "Faults 24" in a model inspector counts supported fault definitions, and on an untouched chiller it reads as 24 active faults; the expanded list mixes catalogue entries with no incident styling distinction.
- **Suggestion** Label it "Available fault types" with a neutral catalogue count, under model capabilities and away from current-condition summaries. Presentation only; the count and list come from the binary's component facts.
- **Conflicts** none.

#### CB-08 Equipment failures and measurement disturbances are mixed
- **Class** `feature-or-binary` · **Dimension** Components · **Effort** M · **Impact** medium · **Needs binary** yes (group membership must come from authoritative metadata) · **Status** `accepted`
- **Sources** R2-11.
- **Where** Plant view → instance inspector → expanded Faults (S2: condenser fouling, refrigerant undercharge and compressor wear sit with measurement offset, drift, dropout and noise).
- **Observed** R2-11: a compressor-wear fault changes the plant model, a sensor offset changes the measurement path, and the list does not say which is which.
- **Suggestion** Group the catalogue as equipment faults and measurement faults with one plain-language example each, and show the affected equipment or channel next to the entry. R2-11's own acceptance forbids keyword classification in the client, so the grouping needs the binary's classification.
- **Conflicts** none (an architecture constraint, `AGENTS.md` "the canvas computes nothing the binary owns").

#### CB-09 "States" mixes running, cooling, COP, cost and biases
- **Class** `ui-presentation` · **Dimension** Components · **Effort** S · **Impact** medium · **Needs binary** no (the rename); yes for the flags/inputs/derived classification the entry asks for as optional · **Status** `accepted`
- **Sources** R2-12.
- **Where** Plant view → instance inspector → the "States 24" block (S2).
- **Observed** R2-12: the block mixes running, cooling, setpoint, COP, cost rate and measurement biases, so a reader can take "running" and "cooling" as both currently true.
- **Suggestion** Rename to "Model variables" and put the block in advanced model details, below the operational explanation. Add declared units/descriptions where the payload already has them; do not invent a flags/inputs/derived split — that classification is binary-owned.
- **Conflicts** none.

#### CB-10 The view inspector's start/stop entries look like controls
- **Class** `ui-presentation` · **Dimension** Components · **Effort** S · **Impact** medium · **Needs binary** no (the label and the rendering); the "route to schedule a command" half is a behaviour change · **Status** `accepted`
- **Sources** R2-13.
- **Where** Plant view → instance inspector → the effector entries.
- **Observed** R2-13: "Effectors start/stop are plain catalogue text in this view" (S2) that reads as if it could command the plant; an equipment overview must not suggest a catalogue entry has already commanded anything.
- **Suggestion** Label them "Available commands" and render them as descriptions rather than bare verbs. If a route from a catalogue entry into run setup is wanted, that is a separate behaviour change and needs the owner's word.
- **Conflicts** none. Vocabulary split against CB-29 is fixed in the Coherence check: a supported effector shown in a catalogue is an "available command"; one placed in a plan is a "scheduled command".

#### CB-11 Ordinary equipment nodes carry card shadows and borders
- **Class** `ui-presentation` · **Dimension** Components · **Effort** M · **Impact** medium · **Needs binary** no · **Status** `accepted`
- **Sources** R2-06.
- **Where** The graph on the plant view, the editor and the review.
- **Observed** R2-06 and S5: equipment is drawn as bordered, shadowed cards inside bordered area containers. Source today: `nodes.module.css:1-8` `.node` has `border: 1px solid var(--node-border)`, `border-radius: 8px` and `box-shadow: var(--shadow-panel)`.
- **Suggestion** Remove the default drop shadow and reduce the routine border weight, keeping selection, fault, focus and alarm treatments strong. R2-06's own acceptance says prototype one representative plant first — see "Do not do / not now".
- **Conflicts** none (VIS-3, ◐, already asks for flat normal nodes).

#### CB-12 Every area is a full rounded box
- **Class** `ui-presentation` · **Dimension** Components · **Effort** M · **Impact** medium · **Needs binary** no · **Status** `accepted`
- **Sources** R2-05.
- **Where** Area frames on the graph, all screens.
- **Observed** R2-05, S1 and S5: every area is a complete rounded rectangle; hiding both side panels leaves the box-heavy structure intact.
- **Suggestion** Prefer a clear area label, whitespace and a subtle separator or background cue over a full rectangle around every area, with labels consistently beside their group and enough separation to tell neighbours apart. Keep an explicit boundary when an area is selected or edited, and keep connection traceability and the frame-title crossing rules.
- **Conflicts** none.

#### CB-13 Repeated tinted header bands and internal dividers
- **Class** `ui-presentation` · **Dimension** Components · **Effort** M · **Impact** medium · **Needs binary** no · **Status** `accepted`
- **Sources** R2-07.
- **Where** Node headers and internal framing, all screens.
- **Observed** R2-07 and S5: every component reads as a miniature window. Source today: `nodes.module.css:19-29` `.header` has `background: var(--surface-hover)` and `box-shadow: inset 0 3px 0 var(--cat, …)`; `.note` adds a bordered band (`:486-495`).
- **Suggestion** Replace the tinted band and repeated dividers with type hierarchy and spacing: id and full kind together, then the secondary port information aligned below. **Keep the header's reserved height**: the full-tier band is deliberately tier-stable (P-12-2 / ISS-8D-17) so a node does not grow when a fault line appears mid-run, and CB-58 is that same decision. Change the paint, not the box.
- **Conflicts** P-12-2 / ISS-8D-17 (the tier-stable 64 px header band) — the constraint, not a reason to drop the entry.

#### CB-14 Every toolbar action is boxed
- **Class** `ui-presentation` · **Dimension** Components · **Effort** M · **Impact** low-medium · **Needs binary** no · **Status** `accepted`
- **Sources** R2-08.
- **Where** The action bars on the plant view, the editor and the review.
- **Observed** R2-08 and S11: primary actions, navigation, toggles and secondary utilities all carry a border, so the eye cannot rank them. The hunt confirms one primary action per screen exists (`Edit plant` / `Run plant`), which is what this entry protects.
- **Suggestion** Reserve the filled treatment for primary actions, use quiet text/icon buttons for secondary operations and visible toggle states for toggles; group by purpose with spacing instead of borders. Removing frames must not remove hover, focus, disabled or selected feedback.
- **Conflicts** none.

### Motion and responsiveness

#### CB-15 The bundle is served uncompressed
- **Class** `feature-or-binary` · **Dimension** Motion and responsiveness · **Effort** S · **Impact** medium · **Needs binary** yes (it is the `streamsim canvas` response, a Go edge) · **Status** `accepted`
- **Sources** BB-4.
- **Where** The transport that serves `canvas/dist` to the browser.
- **Observed** BB-4: the §4 gzip budget is met with room — 660,743 B raw JS / 203,494 B gzip and 82,977 B raw CSS / 15,164 B gzip, 218,658 B against the 600 kB budget — but the review instance served both **uncompressed** (`encodedBodySize == decodedBodySize == 660,743`), so a cold load transfers ≈ 726 kB, 3.3× the budget.
- **Suggestion** Serve the two assets with content encoding. One transport change; the appearance of a budget breach on every cold load goes with it.
- **Conflicts** none. Note the hunt's Motion dimension verdict was "nothing to improve" on measured motion and performance; this is the one responsiveness item, filed here because the loop has no transport dimension.

### Colour and contrast

#### CB-16 "Did not finish" wears the success green
- **Class** `ui-presentation` · **Dimension** Colour and contrast · **Effort** S · **Impact** medium · **Needs binary** no · **Status** `in flight`
- **Sources** IMP-05, R2-17. (Round 2's evidence row S4 is the same finding, measured independently.)
- **Where** The run state chips: the Home run row and the review header.
- **Observed** Hunt: the review's Verified badge is `color rgb(15, 107, 63)` on `rgba(15, 107, 63, 0.12)` with a matching border, and Home's "Did not finish" chip measures **exactly the same**. Round 2 (S4): the badge uses a live-style class, text `rgb(15,107,63)` and a translucent green background, with a radio/broadcast icon. Source today: `RunList.tsx:35-38` renders `styles.live` with a `Radio` icon and the words "Did not finish" when `!row.run.ended`; `HomeScreen.module.css:211-214` `.live { color: var(--status-ok); background: var(--status-ok-bg); }`. "Explore" uses purple, "Evaluate" neutral.
- **Suggestion** Give unfinished/incomplete its own neutral-or-attention token, leaving the ok green to Verified. The delegating agent declares this in flight for another agent. **Watch the live case**: the chip is keyed on `!run.ended`, and round 15's investigation recorded that the workspace listing carries no abandonment signal, so the canvas cannot tell an abandoned run from a genuinely still-open one — see Coherence check.
- **Conflicts** `031d0c8` (one name, "Did not finish") is about the wording and is not opposed; the still-open case is the open question.

#### CB-17 The node's category stripe can read as health
- **Class** `ui-presentation` · **Dimension** Colour and contrast · **Effort** M · **Impact** medium · **Needs binary** no · **Status** `accepted`
- **Sources** R2-09.
- **Where** Node category accents, all screens; the wire-type legend.
- **Observed** R2-09: a green equipment accent must not read as "healthy" or "running"; the legend should not leave a reader guessing whether equipment stripes encode health. Source today: the stripe is `.header { box-shadow: inset 0 3px 0 var(--cat, …) }` (`nodes.module.css:29`), a category colour.
- **Suggestion** One semantic meaning per persistent accent, each paired with text or an icon, and say what the category accent is where it is used. Implements §3.2's "colour never the only signal".
- **Conflicts** none; overlaps CB-16 (one green for two ideas) and CB-60 (the legend's rows).

### Layout

#### CB-18 The run plan's sticky footer covered the last section
- **Class** `ui-presentation` · **Dimension** Layout · **Effort** S · **Impact** medium-high · **Needs binary** no · **Status** `done — bfabd05`
- **Sources** IMP-06.
- **Where** Run plan panel (right panel), after "Faults to inject"; measured at 1440 × 900 and 1280 × 720.
- **Observed** Hunt: footer rect y 785 → 900, sticky, with the last fieldset's children underneath it — "None: nothing is commanded…" at y 786 → 843 and the "Add effector call" button at y 859 → 889 both `coveredByFooter: true`, and the footer's own "Start simulates…" sentence reading as Effector-calls content.
- **Suggestion** None; done. `bfabd05` cites IMP-06 by name: the footer now sits outside the scroll area (the form is a flex column, the fields scroll in their own region, the footer is a fixed row with a full-bleed top border), measured with no field covered at 1440 or 1280 and a new e2e guard. **The fix IMP-06 suggested was measured and rejected**: reserving the footer height as `padding-bottom` changed nothing at the default position and added 115 px of dead scroll.
- **Conflicts** IMP-06's own suggestion (superseded by the measured fix).

#### CB-19 The right panel scrolls with no cue that anything is below
- **Class** `ui-presentation` · **Dimension** Layout · **Effort** S · **Impact** low-medium · **Needs binary** no · **Status** `accepted`
- **Sources** IMP-07.
- **Where** Every scrolling right panel: the plant inspector, the editor's Build panel, the run plan, the review report.
- **Observed** Hunt: `aside._panel` has `overflow: auto`, `scrollHeight 1423` against `clientHeight 856`, `mask-image: none`, no `data-cut` and `padding-bottom: 0`, with a cut row at the window edge; the **outline rail in the same shell already has the cue** (`data-cut="true"` plus `linear-gradient`). Source today: only `PlantScreen.module.css:29-31` defines the `[data-cut="true"]` mask.
- **Suggestion** Apply the rail's existing cut fade (or a 1 px inset shadow) to the right panel's bottom edge when it scrolls, driven by the same scroll listener the rail already has.
- **Conflicts** none.

#### CB-20 The rail and the right panel size themselves by different rules
- **Class** `ui-presentation` · **Dimension** Layout · **Effort** S · **Impact** low · **Needs binary** no · **Status** `accepted`
- **Sources** IMP-08. (Hunt expectation rows G-2 and VIS-10 cover the same ground as evidence. R2-02 is the same
  element with a different fix and a different class, so it is **not** merged here: CB-68.)
- **Where** The shell at 1280 / 1440 / 1920.
- **Observed** Hunt: the left rail is a fixed 255 px at all three widths while the right panel is 336 px at 1280,
  374.4 px at 1440 and 460 px at 1920 — 37 % growth for the same content. R2-15/S11: the graph region grows from
  688 px to 1204 px. Source today: `--sidebar-width: 256px` and `--inspector-width: clamp(336px, 26vw, 460px)`
  (`tokens.css:21-22`).
- **Suggestion** Pick one rule for both panels — a fixed width, or a proportional one applied to both — and let the
  centre take the extra space. The fixed rule is the smaller change and matches the rail that already does it. Panel
  **resizing** is a different proposal with a different class (CB-68); do not fold it in here as presentation.
- **Conflicts** none recorded. R2-01 (CB-24) is the broader version of the same ground.

#### CB-68 Make the panel widths adjustable and recoverable
*Recorded in the Layout section and out of id order: it was split out of CB-20 after the consolidation, because it is a
new capability and not a presentation change.*
- **Class** `behaviour-change` · **Dimension** Layout · **Effort** L · **Impact** low-medium · **Needs binary** no · **Status** `accepted`
- **Sources** R2-02.
- **Where** The shell's panel boundaries, on the plant view and in the editor.
- **Observed** R2-02/S11: no obvious resize handle was visible at 1920 and pointer resizing was not tested. Verified in
  code by the coordinating agent: there is no drag-resize anywhere — `Shell`, `PlantScreen` and `EditorScreen` hold no
  resize logic and no CSS `resize:` — and §3.2's "panels resize" is met today only by the collapse/expand toggles
  (`PanelToggles`), which do change the layout (collapsing the outline measured `pastRight 237 → 0` in round 13).
- **Suggestion** A quiet but discoverable separator handle on the actual panel boundaries with a keyboard equivalent,
  reasonable per-task preferences remembered, and a Reset layout. Check existing resizing support before adding
  anything — R2-02's own instruction — and note there is none. Resizing must not change the selection or silently
  switch tasks. This is a new capability: it needs the persistence and focus story stated before it is scheduled.
- **Conflicts** none recorded. R2-01 (CB-24) is the umbrella decision over this and CB-20.

#### CB-21 A short inspector form leaves its lower half empty
- **Class** `ui-presentation` · **Dimension** Layout · **Effort** S · **Impact** low-medium · **Needs binary** no · **Status** `accepted`
- **Sources** R2-03.
- **Where** Editor → the selected-equipment form at 1920 (S11, the short `ahu-a` form).
- **Observed** R2-03/S11: a short form leaves much of the lower panel unused and the form reads as unfinished; the plan show/hide buttons exist, so the empty half is avoidable.
- **Suggestion** Group identity and editable properties compactly, keep contextual help near the field it explains, and reserve a consistent action location; let the user collapse the panel to reclaim the width instead of filling the space. Do not add filler or decorative cards.
- **Conflicts** none.

#### CB-22 Change status and Save/Revert are not anchored together
- **Class** `ui-presentation` · **Dimension** Layout · **Effort** S · **Impact** medium · **Needs binary** no · **Status** `accepted`
- **Sources** R2-04, r1-15 (one entry: R2-04 is placement, r1-15 is content).
- **Where** The editor bar ("No changes" / "Unsaved changes" / Save / Revert) at wide widths; the run panel's footer sentence.
- **Observed** R2-04/S11: "No changes", pending changes and Save all concern one editing transaction but sit far apart on a wide desktop (a long gap between status and save/revert actions); on narrow widths the positions shift when the status wording changes. r1-15: the save state, undo/revert scope and what Save does to existing recordings are not explained.
- **Suggestion** One compact change/action group with a clear primary action, change status and Save/Revert close enough to read together, view navigation kept separate, and positions stable as the status words change. Add a compact change summary ("Renamed 1 instance; added 1 connection") and say that Save updates the composition and the runnable plant while existing recordings keep their recorded plant.
- **Conflicts** none.

#### CB-23 The selected object's identity scrolls out of view
- **Class** `ux-polish` · **Dimension** Layout · **Effort** M · **Impact** medium · **Needs binary** no · **Status** `accepted`
- **Sources** R2-19.
- **Where** Plant view → instance inspector (reproduced as S3).
- **Observed** R2-19/S3: after scrolling down an expanded chiller inspector, selecting `pmp-u1` swaps the content but keeps the scroll position, so the new object's heading and description stay off-screen; expanded section choices also carry across.
- **Suggestion** A compact sticky name/kind/location header, and on changing objects either return the detail pane to its start or preserve the selected subsection with its heading visible. R2-19 gives the acceptance test: expand chiller details, scroll down, select a pump, and the pump's identity is visible immediately.
- **Conflicts** none.

#### CB-24 The desktop composition does not adapt to the task
- **Class** `owner-design-decision` · **Dimension** Layout · **Effort** L · **Impact** medium · **Needs binary** no · **Status** `owner decision`
- **Sources** R2-01. (Round 1's 02 and 13 are the proposals it refines; they are recorded separately as CB-28 and CB-26.)
- **Where** The shell on the plant view, the editor and the review, 1280/1440/1920.
- **Observed** R2-01/S5, S11 and the round-2 main finding: the composition does not adapt to the current task — a viewing, a connection-editing and a report-reading task have different space needs, and wider screens mostly grow the gaps. Round 2 is explicit that its measurements do **not** support a "fixed width" or "does not fill the browser" claim.
- **Suggestion** A stable shell with task-specific proportions rather than a separate design system per page: give the active form or report enough width while retaining plant context, and avoid filling empty space with extra panels. This is the umbrella over CB-20 (width rule), CB-21 (short form) and CB-26 (panel sections); take it as one product decision, not as four independent tweaks.
- **Conflicts** none recorded; it would supersede the current `clamp()` panel rule if the owner chooses.

#### CB-25 An unfamiliar plant has no short introduction
- **Class** `feature-or-binary` · **Dimension** Layout · **Effort** M · **Impact** medium · **Needs binary** yes (an authored purpose; the canvas must not invent a process story) · **Status** `accepted`
- **Sources** r1-01.
- **Where** The default plant inspector (the no-selection overview).
- **Observed** r1-01: the overview leads with inventory counts (areas, kinds, recent runs — P-6, ✅). Nothing states what the plant models or where to start.
- **Suggestion** An optional, concise overview: what is modelled, the main equipment groups, important inputs/outputs and a suggested first component. Use authored domain information and leave missing explanations explicitly unavailable. Dismissible and reopenable.
- **Conflicts** none. Overlaps CB-40 (first-use path) — the two should be decided together.

#### CB-26 The Build panel is one long mixed column
- **Class** `ux-polish` · **Dimension** Layout · **Effort** M · **Impact** medium · **Needs binary** no · **Status** `accepted`
- **Sources** r1-13.
- **Where** Editor → Build panel (r1-13, observed as "long mixed panel" in E4–E5).
- **Observed** r1-13: Add equipment, Edit selected equipment, Connections and Areas run together in one column, so adding a pump means scrolling past area-management controls; the hunt measured the Build panel at `scrollHeight 1423` against `clientHeight 856`.
- **Suggestion** Separate Add equipment, Edit selected equipment, Connections and Areas into clearly named sections in the existing panel, keep the current task and selected item visible, and move repeated syntax instructions into short help disclosures while keeping actionable hints next to their controls. Closing equipment details should clearly return to the build choices.
- **Conflicts** none.

### Content design

#### CB-27 The right panel has two names
- **Class** `ux-polish` · **Dimension** Content design · **Effort** S · **Impact** medium · **Needs binary** no · **Status** `accepted`
- **Sources** IMP-09.
- **Where** The Keyboard dialog's "Everywhere" group against the top-bar button and the review's panel label.
- **Observed** Hunt: the dialog says "hide or show the outline panel, the **details panel**"; the button is "Inspector" with `title="Hide the inspector panel (])"`; the review's region is `aria-label="Run details"`; the plant inspector is `aria-label="Inspector"`. Source today: `keyList.ts:23-24` still says "the details panel" twice.
- **Suggestion** Use "Inspector" in the keyboard help, the button title and the review's region label. G-6 asks one action to have one name.
- **Conflicts** none.

#### CB-28 The current activity is not named
- **Class** `ui-presentation` · **Dimension** Content design · **Effort** M · **Impact** medium · **Needs binary** no · **Status** `accepted`
- **Sources** r1-02.
- **Where** All screens: the breadcrumb and the screen's own bar.
- **Observed** r1-02 (grounded in E2, E4, E6–E7): the breadcrumb says where you are but not what you are doing; the hunt confirms each screen still carries its own second bar with its own control set (G-1, ◐, "partly").
- **Suggestion** Make the activity explicit — View plant, Edit plant, Run setup, Running, Review recording — and keep it distinct from Explore/Evaluate, which describes the run's evidence visibility. Reuse one small status component across screens rather than adding another navigation bar; G-1's two-rows-per-screen question is the same ground.
- **Conflicts** G-1 (◐; the bar pattern is unrecorded as a decision either way).

#### CB-29 Domain terms are not explained where they are used
- **Class** `ui-presentation` · **Dimension** Content design · **Effort** M · **Impact** medium · **Needs binary** no · **Status** `accepted`
- **Sources** r1-04.
- **Where** All screens: instance, port, channel, effector, seed, truth, verification.
- **Observed** r1-04: the terms are used throughout and defined nowhere; the hunt found no developer jargon in the default state (G-6, closed), which is the different problem.
- **Suggestion** Short contextual definitions where each term becomes relevant, one consistent meaning across view, edit and run, exact technical identifiers kept in inspectable detail. **One vocabulary with CB-10**: a supported effector in a catalogue is an "available command"; one placed in a plan is a "scheduled command"; the technical object stays "effector call" in details. Do not rename the same string twice.
- **Conflicts** none.

#### CB-30 Equipment ids carry no human-readable role
- **Class** `ui-presentation` · **Dimension** Content design · **Effort** M · **Impact** medium · **Needs binary** yes (an authored purpose; a new persisted description field needs a contract) · **Status** `accepted`
- **Sources** r1-12, R2-14 (one entry).
- **Where** Plant view and editor: the selected object's name and kind line; abbreviated kinds (`ahu`, `chw`, `cond`).
- **Observed** r1-12: stable ids are the only handle. R2-14/S1, S5: the graph and outline abbreviate to `ahu`/`chw`/`cond` while the inspector expands the selected one to "Air-handling unit", and the useful model explanation sits below the technical text. The hunt measured the instance inspector already stating the kind once against the raw id (round 13's regression pass, held).
- **Suggestion** Expand abbreviations in human-facing labels and show a concise role sentence beside the selected name before the equations, with the exact id secondary and the full model explanation kept under disclosure. Use existing metadata first; adding a persisted purpose field is its own contract decision.
- **Conflicts** none.

#### CB-31 An area with no direct equipment reads "0 · empty"
- **Class** `ui-presentation` · **Dimension** Content design · **Effort** S · **Impact** low-medium · **Needs binary** no · **Status** `accepted`
- **Sources** IMP-13, R2-15 (one entry, two different fixes for one element).
- **Where** The outline rail: the editor's Halls/area groups and the read-only outline's parent areas.
- **Observed** Hunt (IMP-13): the group header's count element reads "0 · empty" while every sibling shows a bare count ("2", "9", "5", "12"); the empty editor's area group reads "Factory 0 · empty". Round 2 (R2-15/S6): in Food cold chain editing, **Site** and **Cold rooms** show `0 · EMPTY` although Room A and Room B exist beneath Cold rooms and each holds an air-handling unit; the read-only outline omits these unpopulated parents entirely. Source today: `Outline.tsx:158-160` `{items.length === 0 ? "0 · empty" : items.length}`.
- **Suggestion** Both streams agree the state word must go. IMP-13 (smaller): a bare "0" matching the column, with the empty case said in words where the Add-area action is. R2-15 (the one that fixes the S6 confusion): a direct/descendant distinction or a child-area count beside the heading. Take R2-15's scope: show the direct count, and where the area has child areas with equipment, show the child count too — the outline already builds that tree (`groups`/`branchesOf`). Count scope must then match the inspector's AREAS list.
- **Conflicts** none. The editor's populated-first ordering is CB-53 and must not be changed by this entry.

#### CB-32 Simulated time and waiting time are not distinguished
- **Class** `ui-presentation` · **Dimension** Content design · **Effort** S · **Impact** medium · **Needs binary** no · **Status** `accepted`
- **Sources** r1-18.
- **Where** Run setup and the running panel: Pace/speed, duration, fault minutes.
- **Observed** r1-18 (grounded in E6, E10): compressed "s per s" language, and no separation between plant duration, elapsed wall time, execution pacing and playback speed. Source today: `launch.ts:220` still renders `${Math.round(rate)} s per s`, with `watchWords` giving "about 30 s".
- **Suggestion** Keep plant duration, elapsed wall time, execution pacing and playback speed distinctly labelled; replace "s per s" with a plain equivalent plus optional technical detail; label every scheduled event in plant time. r1-18's acceptance: a fault at minute 5 is five **simulated** minutes after run start.
- **Conflicts** none. R-4/R-9 (the running panel's speed) were never reached inside the hunt's budget.

#### CB-33 Command parameters are mechanically humanised
- **Class** `ui-presentation` · **Dimension** Content design · **Effort** M · **Impact** medium · **Needs binary** yes (the command's declared purpose and prerequisites are domain metadata) · **Status** `accepted`
- **Sources** r1-20.
- **Where** Run setup → Effector calls → the parameter field (E10: label "pressure kpa", range placeholder 500 to 1300).
- **Observed** r1-20/E10: a label produced by splitting the key; the allowed range reads only as a placeholder and disappears once a value is typed; nothing says what 1100 means or that a setpoint is a request, not an outcome. Source today: `humanise()` in `report.ts:274-278` splits on `__` and replaces `_` with a space, so `pressure_kpa` becomes "pressure kpa".
- **Suggestion** "Pressure setpoint (kPa)", the allowed range visible after typing, and one plain sentence on what the command requests and its prerequisites from domain metadata — distinguishing a requested setpoint from a guaranteed measured pressure. The label and unit half is presentational; the purpose text is binary-owned.
- **Conflicts** none.

#### CB-34 Snapshot readings and whole-run summaries share one scope
- **Class** `ui-presentation` · **Dimension** Content design · **Effort** M · **Impact** medium · **Needs binary** no (the run artifact's `command_log` items already carry `args`) · **Status** `accepted`
- **Sources** r1-24, R2-23 (one entry).
- **Where** Review → Inspect and the report's KPI tables.
- **Observed** R2-23/S9: "Observed at 00:18:00" and a whole-run mean answer different questions; some "Observed at" cells carry no visible timestamp and the table's aggregate time scope is not stated beside its heading. r1-24/E7, E10: the outcome table shows "applied · 251 ms" without the requested 1100, does not explain accepted/applied/refused or what the latency measures, and the KPI tables do not say whether a range is observed min–max or an acceptable band. The report's `outcome()` (`report.ts:129-130`) prints `applied · <ms>` and ignores the command's arguments, but the arguments are already in the canvas's payload: `contracts/run-artifact-v0.1.schema.json` `/properties/command_log/items` carries `seq`, `at_ns`, `op`, `args` and `result_digest` (verified in code by the coordinating agent).
- **Suggestion** A concise time-scope label on each table (snapshot at the playhead vs whole run) with a visible timestamp or an explicit same-as-playhead indicator; carry the requested arguments next to accepted/applied/refused and say what the latency measures; explain mean and range, the aggregation interval, and direction or target only when declared. State plainly that a simulator's applied command is not proof of a physical effect. The command half needs **no** binary fact — only `args` rendered beside the outcome; a KPI's declared direction/target is the one part that would need metadata if it is ever wanted.
- **Conflicts** none. Keeps V-11/V-12 (numbers aligned, real headers — both closed by the hunt) untouched.

#### CB-35 A fault label does not say where it comes from
- **Class** `ui-presentation` · **Dimension** Content design · **Effort** S · **Impact** medium · **Needs binary** no · **Status** `accepted`
- **Sources** R2-24.
- **Where** Review → Inspect, where a `Fault: <cause>` line sits beneath Observed (S9).
- **Observed** R2-24/S9: a scheduled, known fault and a sensor observation are different sources of evidence and the line does not say which it is. Source today: `Readout.tsx:49-59` `Note` renders `Fault · <cause>` above the observed rows.
- **Suggestion** Label the provenance of the fault line (scenario / observed alarm data / an outside system's verdict) and keep scenario conditions and measured readings in clearly named adjacent subsections. Respect the truth boundary: never imply an external consumer detected it unless that is recorded. Presentation only — not a claim of live truth leakage.
- **Conflicts** none; a constraint from UI-EXPECTATIONS principle 7 and §3.1's truth-boundary check.

#### CB-36 "None: the plant runs healthy." promises health
- **Class** `ui-presentation` · **Dimension** Content design · **Effort** S · **Impact** low-medium · **Needs binary** no · **Status** `in flight`
- **Sources** R2-25. (Round 2's evidence row S10 is the same finding.)
- **Where** Run panel → Faults to inject, empty state.
- **Observed** R2-25/S10: no injected fault is planned; that is not a health verdict. The hunt's expectation table records this string as a met empty state (G-4, "closed") — the two streams disagree about whether it is a defect. Source today: `FaultRows.tsx:117` `<p className={panel.hint}>None: the plant runs healthy.</p>`. The delegating agent declares this in flight for another agent.
- **Suggestion** "No faults scheduled", with a short explanation if needed, kept inside the fault schedule and leaving health to evidence-backed running/review sections.
- **Conflicts** the hunt's G-4 reading (which treats it as a designed empty state). Round 2's reading is the stricter one and the owner has asked for comprehension over a feature checklist; the change is copy-only.

#### CB-37 The plant row's not-runnable sentence is a run-on
- **Class** `ui-presentation` · **Dimension** Content design · **Effort** S · **Impact** low · **Needs binary** no · **Status** `accepted`
- **Sources** IMP-14.
- **Where** Home → the plant row whose composition does not compose.
- **Observed** Hunt: "Not runnable yet: open Edit, save, and fix any issue it lists." sits in the meta slot where every other row reads "14 instances · 7 runs · last 15 Oct 2026, 16:15". Source today: `workspace.ts:70` returns that exact sentence.
- **Suggestion** State first, action second ("Not runnable yet. Edit it, then save."), or move the remedy to the Edit button's tooltip and leave the row with the state.
- **Conflicts** none.

#### CB-38 "Verified" appears twice in adjacent lines
- **Class** `ui-presentation` · **Dimension** Content design · **Effort** S · **Impact** low · **Needs binary** no · **Status** `accepted`
- **Sources** IMP-15.
- **Where** Review header chip and the first line of the report.
- **Observed** Hunt: the chip reads "Verified" (green, tick icon) and directly under it the collapsed integrity line reads "Verified: the files match and the replay reproduces." Source today: `reportViews.ts:104` still returns that sentence, and it is pinned by `ReportPanel.test.tsx:22`.
- **Suggestion** Let the badge own the word and the sentence own the explanation ("The files match and the replay reproduces."), which is also shorter inside a 374 px panel. Update the test string.
- **Conflicts** none. `1c9089c` (plain-words integrity) is not opposed.

#### CB-39 The About-this-run sentence does not parse
- **Class** `ui-presentation` · **Dimension** Content design · **Effort** S · **Impact** medium · **Needs binary** yes (the wall-clock run date is G-7's recorded `B` gap) · **Status** `accepted`
- **Sources** IMP-10, BB-5 (one entry: BB-5 is the binary half of the same popover).
- **Where** The ⓘ "About this run" popover in the review header (440 × 69 px).
- **Observed** Hunt: it renders "… **run time 00:00:00 is 00:00:00 UTC simulated**", two facts stated as one equation with no labels, so neither the wall-clock time nor the simulated span can be read out of it, although both are known. Identical shape on w-33 and w-45.
- **Suggestion** Split it into labelled facts ("Simulated 30 min of plant time", and the wall-clock line in G-7's format) and drop the duplicated "00:00:00 is". The sentence half can be corrected now for the facts the bridge already provides; a run **date** needs the binary to report when the run ran (BB-5).
- **Conflicts** G-7's `B` item (the run date) — the reason the entry is only half doable now.

#### CB-40 A ready-to-run example and a guided journey are missing
- **Class** `feature-or-binary` · **Dimension** Content design · **Effort** L · **Impact** medium · **Needs binary** no · **Status** `owner decision`
- **Sources** BB-6. (Hunt expectation rows N-1 and N-4 are the same ground as evidence.)
- **Where** Home (no example entry among 14 plants) and the editor (the canvas hint is the only guidance).
- **Observed** BB-6: N-1 and N-4 both stand (✗). The hunt notes the editor's canvas hint as the closest thing that exists.
- **Suggestion** Add a ready-to-run example as the shortest first-use path and offer a journey that names the next unfinished step. This is new content and new interaction, not polish, and it overlaps CB-05 (creation preview) and CB-25 (plant introduction) — the owner should take the three as one first-use decision. Trade-off recorded in BB-6: shortest first-use path versus more surface to keep honest.
- **Conflicts** none; UI-EXPECTATIONS' "Order of work" already lists a starter example and creation second.

### Feedback and flow

#### CB-41 A doomed domain probe leaves a 404 and a console error on the editor path
- **Class** `ux-polish` · **Dimension** Feedback and flow · **Effort** S · **Impact** medium · **Needs binary** no · **Status** `accepted`
- **Sources** IMP-16.
- **Where** Home → the "Dir Empty Check 8" row (labelled "Not runnable yet") → Edit.
- **Observed** Hunt: the click issues `200 …compositions%2Fdir-empty-check-8.composition.json`, then **`404 …domains%2Fdir-empty-check-8.domain.json`**, then `200 component.list`, and logs `error: Failed to load resource: … 404`. The editor still renders correctly and offers no "Open plant". Source today: `editorStore.load()` calls `versionOf(domainOf(doc))` → `readFile("domains/<id>.domain.json")`, which swallows the failure but still issues the request. §3.1 requires no failed request and no console error in normal use. Round 13's `17f2274` gated the analogous **review** fetches; this editor path is a different surface and is not covered by it.
- **Suggestion** Answer "is there a composed domain?" from the workspace listing the editor already holds — the same fact the Home row used — instead of re-probing the file. This removes one request rather than adding any, and needs no new capability or endpoint.
- **Conflicts** §3.1 ("no failed network request in normal use"). Not a behaviour change: the rendered states are unchanged.

#### CB-42 The workspace cannot answer "which plant or run do I need?"
- **Class** `feature-or-binary` · **Dimension** Feedback and flow · **Effort** M · **Impact** medium · **Needs binary** no · **Status** `accepted`
- **Sources** r1-05.
- **Where** Home: the Plants and Runs lists.
- **Observed** r1-05/E1: no search across plant names/ids and run ids, and the run filters are limited. What exists today: W-4 (filter by plant and paging) is ✅, W-3/W-9/W-11 are ✅, and the hunt measured 14 plants / 45 runs with W-4 scaling to hundreds.
- **Suggestion** Add a focused search across plant names/ids and run ids plus useful run filters (mode, completion state), with filters kept visible, result counts shown and one clear reset. Show plant identity, run time and status in a stable order. Say why a filtered list is empty.
- **Conflicts** none.

#### CB-43 Outline search does not show on the graph
- **Class** `behaviour-change` · **Dimension** Feedback and flow · **Effort** M · **Impact** medium · **Needs binary** no · **Status** `accepted`
- **Sources** r1-10.
- **Where** Plant view, editor, review: the outline filter against the graph.
- **Observed** r1-10/E3: the outline correctly reports three pump matches and announces the count (source today: `Outline.tsx:190-196` `matchWords`), but the graph gives no corresponding emphasis.
- **Suggestion** Highlight matching nodes and optionally fit those results while retaining surrounding context; keep search matches visually distinct from the selected item and from alarms; clearing the query restores the previous overview, and an empty result explains itself with a reset. New visual state on the graph — the owner's word covers whether a fit-on-search is wanted.
- **Conflicts** none. It must not reopen the opening-zoom floor (CB-57): highlight is not a zoom change.

#### CB-44 The running screen does not explain what has happened
- **Class** `feature-or-binary` · **Dimension** Feedback and flow · **Effort** M · **Impact** medium · **Needs binary** no · **Status** `accepted`
- **Sources** r1-21.
- **Where** While a run runs: beside the progress bar.
- **Observed** r1-21/E6: progress and graph fault indication exist (R-4, ◐: "no speed, no fault ticks"), but nothing says what has happened so far. The hunt never entered a running state (R-4/R-5/R-9/RUN-5 not reached, budget spent on zero runs).
- **Suggestion** A small recent-event strip beside progress: what occurred, which equipment, and where to inspect it; distinguish active from ended disturbances; brief transitions with a stable layout and reduced motion respected; never expose forbidden truth in Evaluate mode. Events must stay readable without animation or colour alone.
- **Conflicts** none. Because the hunt could not measure the running screen, this entry rests on r1-21's walkthrough (E6) alone.

#### CB-45 The whole experiment cannot be previewed before Start
- **Class** `ui-presentation` · **Dimension** Feedback and flow · **Effort** M · **Impact** medium-high · **Needs binary** no · **Status** `accepted`
- **Sources** r1-19.
- **Where** Run panel, before Start.
- **Observed** r1-19/E6, E10: the plan is a set of form rows plus the footer sentence ("Start simulates 30 min of plant time in Explore mode with no faults, paced to take about 30 s…", RUN-1 ✅); nothing lists the faults and commands chronologically with names, times and arguments. The planned marks are already computed for the graph: `planMarks` (`launch.ts`) and the plan's own `mark.entity`/`mark.what`/`mark.atS`.
- **Suggestion** A compact chronological preview of faults and commands with equipment names, times and the entered arguments, keeping the existing start summary and an easy route back to edit a scheduled item. r1-19's acceptance: notice an unintended time or target by reading it.
- **Conflicts** none. Not the same element as CB-18 (the footer) — this is a new list in the panel body.

#### CB-46 The graph's screen-reader instructions promise deletion on read-only screens
- **Class** `ux-polish` · **Dimension** Feedback and flow · **Effort** S · **Impact** low · **Needs binary** no · **Status** `in flight`
- **Sources** IMP-11.
- **Where** The plant graph on the plant view, in the editor and in a review.
- **Observed** Hunt: a `display: none` node inside `[role="application"]` reads "Press enter or space to select a node. Press delete to remove it and escape to cancel." on all three screens, where only the editor can delete. Source at HEAD before the working-tree edit: `PlantGraph.tsx` passed one `ariaLabelConfig`. **This is already in the working tree uncommitted**: `PlantGraph.tsx` now defines `graphWords.select` and `graphWords.edit` and passes `ariaLabelConfig={onConnect === undefined ? graphWords.select : graphWords.edit}`, with an untracked `PlantGraph.test.tsx`. Do not duplicate it. **The impact is lower than the hunt's wording suggests**: the fixer measured that the description element is `display: none` and nothing references it by id, so it was latent DOM text — invisible and unannounced — rather than a description a screen reader read out.
- **Suggestion** None beyond the landed edit: keep the select-only words off the editor and the delete words only where `onConnect` exists. Worth a one-line comment in the component that the element is inert today, so the next reader does not re-rate it as an announcement defect.
- **Conflicts** none. Same principle as G-15's "describes no Play that does not exist".

#### CB-47 Explore/Evaluate and current evidence availability share one label
- **Class** `ux-polish` · **Dimension** Feedback and flow · **Effort** M · **Impact** high · **Needs binary** yes (before a run, whether an outside system is connected and whether a score can be produced) · **Status** `accepted`
- **Sources** r1-17, R2-18 (one entry; R2-18 refines r1-17).
- **Where** Run setup, the running screen and the review header/report.
- **Observed** r1-17/E8, E10: Explore setup says the truth is visible after the run; the Explore report says true results were visible while it ran; the same report's generic true-results explanation says an outside system could not see them while live — one screen's three statements disagree. R2-18/S8, S9: "Evaluate" describes a run mode but a completed Evaluate review showing true results still carries a generic lock beside its mode name. V-13's recorded ✗ ("KPI TRUTH shown while the run panel says the truth stays sealed") is the same ground. After the run the report does state availability separately: "Not scored: no outside system is connected to this run."
- **Suggestion** One mode-specific explanation covering who can see what, when and why; separate the mode label from current visible/sealed availability and give lock imagery a precise scope; show separately whether an outside system is connected and whether this run will produce a score, so choosing Evaluate does not imply a scored experiment is configured. Repeat the same meaning in the review. Present it as status/help, not as a new validation exercise, and do not touch the truth boundary.
- **Conflicts** V-13 (✗) and E8's wording ambiguity — this entry is the fix for both.

#### CB-48 An incomplete run has no explained next step
- **Class** `already-fixed` · **Dimension** Feedback and flow · **Effort** — · **Impact** — · **Needs binary** yes for the "still executing vs did not finish" distinction only · **Status** `stale`
- **Sources** r1-25. (Hunt expectation rows E-9 and RUN-5 are the same ground as evidence.)
- **Where** The Home run row and the review of an unfinished run (w-44).
- **Observed** r1-25/E9: "Did not finish", "Not checked", no explanation and unavailable true results. Rounds 13 and 14 measured the current state independently: the w-44 review answers entirely from the listing with **no doomed fetch**, and renders "Not checked: this run could not be replayed for its verdict.", "Not available yet: the run has not ended, so its true results are still hidden.", "Not scored: no outside system is connected to this run.", plus "Open plant" and "Set up another run" in the header band. `031d0c8` gave the state one name.
- **Suggestion** No change for the explanation and the next steps; they exist. The one clause of r1-25 that is still open — "distinguish a run still executing from one that did not finish when the runtime can establish that distinction" — needs a binary signal; round 15 investigated and recorded that the listing carries no abandonment signal.
- **Conflicts** none. Related: CB-16's live-vs-abandoned question is the same missing signal.

### Craft

#### CB-49 Keeping graph meaning readable as zoom changes
- **Class** `already-fixed` · **Dimension** Craft · **Effort** — · **Impact** — · **Needs binary** no · **Status** `stale`
- **Sources** r1-07.
- **Where** The graph at every zoom, all screens.
- **Observed** r1-07/E2, E7: small labels at overview. What the tree does today: a tier system `tierOf(zoom)` puts compact at ≥ 0.4 and full at 1, drops/shrinks labels below it to an 8 px floor, and shows id and kind in words at the opening zoom (P-2 ✅, `d8a8549`, `98368b3`); selection details sit outside the scaled graph in the inspector. Round 2 itself says "a low-detail graph representation already exists; round 1's detail-level proposal should be read as refinement of existing behavior, not an absent capability".
- **Suggestion** No change as written. If something is still wanted, it is the parked overview-tier improvement, which the ledger records as closed by the 0.4 floor (ISS-10D-02/04) and by the readout floor (ISS-10D-03) — that is CB-57's trade-off, not a legibility defect.
- **Conflicts** reading r1-07 as "fit every label at overview" reopens the fit-vs-floor decision (CB-57).

#### CB-50 Reference screens and a state sheet do not exist
- **Class** `owner-design-decision` · **Dimension** Craft · **Effort** M · **Impact** low · **Needs binary** no · **Status** `owner decision`
- **Sources** BB-7. (Hunt expectation row VIS-11 is the same ground.)
- **Where** Documentation, not the product.
- **Observed** BB-7/VIS-11 (✗): no reference screens or state sheet exist; every state in the hunt was measured off the built product.
- **Suggestion** A documentation artifact — both themes, covering empty, loading, error and selected states — that would make the next review cheaper without changing the UI. The owner must decide whether it is worth funding; it is not a UI change and must not be smuggled into one.
- **Conflicts** none.

### Typography

#### CB-51 A channel with no reading is a bare dot whose meaning is only in a tooltip
- **Class** `ui-presentation` · **Dimension** Typography · **Effort** S · **Impact** medium-high · **Needs binary** no · **Status** `accepted`
- **Sources** IMP-21, r1-11 (one entry).
- **Where** Review readout nodes (all runs) and the same markup in the plant graph's readouts; the Inspect table's value cells.
- **Observed** Hunt: the value slot paints a single middle dot (`·`, at x 760.7 → 764 in a 237 px node) with the words only in a `title` and a screen-reader span, dozens of times per campus screen, so a row of dots cannot be scanned and G-10 ("a tooltip must never be the only copy of a fact") is breached at the one fact that decides missing versus zero. r1-11: the graph uses different compact representations for "not observed" and "no reading yet" and they should be discoverable and distinguished from zero, stale data, disconnected ports and unavailable truth, with the same wording and treatment in the graph and the inspector. Source today: `Readout.tsx:9-17` paints `·` for not observed and `–` for no reading yet.
- **Suggestion** Paint the state in the value slot — a dimmed em dash, or "not observed", at the same 11 px unit tone — keeping the `title` for the longer phrasing and the `sr-only` words. Do the same for "no reading yet" so the two states are distinguishable without hovering, and use the same words in Inspect.
- **Conflicts** none. V-5's trend line is CB-64; the two do not touch the same cell.

### Whole-product coherence

#### CB-52 One area is typed three ways on one screen
- **Class** `ui-presentation` · **Dimension** Whole-product coherence · **Effort** S · **Impact** medium · **Needs binary** no · **Status** `accepted`
- **Sources** IMP-18, R2-16 (one entry, two streams asking for different treatments of the same label).
- **Where** Editor and plant view: the outline rail, the canvas frame, the Area select and the AREAS list.
- **Observed** Hunt: the rail shows "HALL 1" (uppercase via CSS, 11 px), the canvas frame shows "Hall 1", and the Area select and AREAS list show "Site › Halls › Hall 1"; on the plant view the rail shows "HALL 1" against the frame's "Hall 1"; the rail also types "HALL 1" against its own AREAS block. Source today: the outline group title is the area name (`Outline.tsx:155`), styled by `shared.module.css:16-27` `.sectionTitle` with `text-transform: uppercase`; the editor's Area pick renders `area.path` (full path, `InstanceEditor.tsx:89-95`). R2-16/S6: the Area selector is the only place the full relationship is visible, and moving from view to edit makes the plant look as if it acquired areas (unpopulated parents are omitted in the read-only outline and shown in the editor).
- **Suggestion** The two streams disagree about the rail. IMP-18: the leaf name in the rail and on the frames, the full path only in the Area select and the AREAS list — which is almost what the tree already does, so the change reduces to **not uppercasing the rail's area name** (the section-title transform is a shared convention; override it for area names and keep it for the other section headings), and to showing the unpopulated parents in the read-only outline as R2-16 asks. R2-16: put the hierarchy in the outline itself. The smaller fix, and the one that keeps the rail's scan rhythm, is IMP-18's leaf + the path in the selected-object location line and the select; R2-16's path-in-outline is the larger one and is a taste call. Take the leaf + the parent/child treatment from CB-31 so view and edit agree about which areas exist.
- **Conflicts** none recorded. R2-16 wants the path where IMP-18 wants the leaf — the choice is named above, not smoothed over.

#### CB-53 The same plant's outline is ordered differently on the plant screen and in the editor
- **Class** `owner-design-decision` · **Dimension** Whole-product coherence · **Effort** M · **Impact** medium · **Needs binary** no · **Status** `owner decision`
- **Sources** IMP-19.
- **Where** The outline rail on the plant view and in the editor, same plant, same session.
- **Observed** Hunt: the plant rail reads `wx-site`, `cal-site`, `comp-u1 … mtr-u1` (Utilities), then `pl-h1 … mtr-h1` (Hall 1) and the rest; the editor rail reads `wx-site`, `cal-site`, then the halls, with **Utilities last** and an extra "Halls" group heading the hall block. Source today: both screens use the same `Outline` component and the same `areaOrder(plant)` (`Outline.tsx:61-67`), but the editor passes `showEmptyAreas`, which both shows the empty groups **and** reorders them with `populatedFirst` (empty areas sink below populated ones); the "Halls" section is rendered above the area groups for both screens when a hall can be opened.
- **Suggestion** IMP-19 wants one order: the plant's declared area order on both screens, keeping the "Halls" affordance as a heading at Hall 1's position rather than a reordering. That reverses a deliberate rule, so the owner decides.
- **Conflicts** `6611972` (an empty area in the outline sits below every populated area, never between them) and `3dc9570` (the plant rail and the inspector's AREAS block read one declared area order). **The trade-off**: one order shared by both rails and the AREAS block, versus an editor rule that keeps an author's still-empty areas out of the way — which is also what round 2's S6/R2-15 (CB-31) wants visible. Fixing CB-31 without touching this rule is the smaller step.

#### CB-54 The same link is written in two directions
- **Class** `ui-presentation` · **Dimension** Whole-product coherence · **Effort** S · **Impact** low · **Needs binary** no · **Status** `accepted`
- **Sources** IMP-20, R2-20 (one entry; round 2's evidence row S7 is the same finding).
- **Where** The plant inspector's links row against the editor's Links row.
- **Observed** Hunt: the plant inspector renders "comp-u1 air out → rcv-u1 air in" (source first, arrow right; its `aria-label` reads "Select source comp-u1; …"); the editor renders "air in ← comp-u1 air out" (destination first, arrow left). Round 2 (S7): the same relationship becomes "chw in ← chiller-1 supply out" in edit; the creation form offers "Link an output", so the form starts from the source while the list starts from the receiver. Source today: `InspectorLinks.tsx:40` builds `${from} → ${to}`; `InstanceEditor.tsx:104-108` renders `{end.port}` then `{end.outgoing ? "→" : "←"}` then the other end, i.e. the receiving port first for an incoming link.
- **Suggestion** Read every connection in source-output → receiver-input order in both places, with Inputs and Outputs groups where useful. R2-20 adds: when starting from a receiver, provide a clear "Connect an input" route or a direct route to the source's connection tool, and keyboard users should be able to begin from either endpoint. The order swap is the smaller change; the connect-an-input route is a behaviour change and its own decision.
- **Conflicts** none. The legend's own sentence already says values run outputs → inputs (round 2's E2), so one direction is enough.

#### CB-55 Panel-back actions are not named by their destination
- **Class** `ux-polish` · **Dimension** Whole-product coherence · **Effort** S · **Impact** low-medium · **Needs binary** no · **Status** `accepted`
- **Sources** R2-21. (Round 1's 03 is the proposal it refines; it is recorded separately as CB-56.)
- **Where** Editor → the selected-instance panel's close control ("Close the instance", S7) against the shell's panel show/hide button.
- **Observed** R2-21/S7: closing selected-instance details returns to the build tools, while hiding the panel is a different action, and both rely on the same X shape; a user cannot predict whether the action clears the selection, returns to the library or hides the whole panel.
- **Suggestion** An explicit back label or understandable tooltip ("Back to build tools") beside the selected-object heading, separate from the shell's show/hide control. The shell's show/hide already has correct `aria-pressed` and titles (`d9c2b8c`-era work, VIS-8/VIS-9 ✅) — keep those distinct.
- **Conflicts** none.

#### CB-56 Moving between plant, editor and review loses the user's context
- **Class** `behaviour-change` · **Dimension** Whole-product coherence · **Effort** M · **Impact** medium · **Needs binary** no · **Status** `accepted`
- **Sources** r1-03.
- **Where** Plant ↔ editor ↔ review transitions, and following a link into another area.
- **Observed** r1-03/E2, E4, E7: moving between screens drops compatible selection and camera context, and opening the historical plant recorded with a run is not labelled as a context change. Existing pieces: G-11 (◐ — the hunt found panel collapse is remembered across screens and a browser restart, VIS-9 ✅, but the unsaved-leave half was not exercised), and historical-plant notices exist where available (r1-03 says reuse them).
- **Suggestion** Retain compatible selection and camera context when moving between plant, editor and review, label transitions that intentionally change context (especially opening a run's recorded plant), and offer a clear back action after following a connection into another area. r1-03's acceptance: after following two linked components, return to the original selection without hunting.
- **Conflicts** none. Distinct from CB-55 (a label) and from VIS-9 (panel state, done).

#### CB-57 Fit-first opening on wide plants
- **Class** `owner-design-decision` · **Dimension** Whole-product coherence · **Effort** M · **Impact** medium · **Needs binary** no · **Status** `owner decision`
- **Sources** BB-2. (Hunt expectation rows P-3 and V-9 are the same ground as evidence.)
- **Where** The opening viewport on every screen that shows a plant.
- **Observed** BB-2/P-3, V-9: every screen opens at the 0.4 floor and gets a pan cue when the plant is cut; P-3 measures the campus at 1440 as 353 of 856 px (41 %) empty, split 178 above / 135 below. Round 13/14 measured the same rule as intended and identical on the plant view and in the editor: campus 0.4 / tier-compact / 236.8 px cut / cued on both, `dir-review-empty` fitted at 1.0 with no cue. The ledger records the rule as the resolved Option A ("keep the 0.4 floor everywhere and extend the pan cue", `7c06d48`, `6b52e29`).
- **Suggestion** Also on record: let wide plants open fitted. It re-introduces sub-0.4 tiers with 8 px kind lines — the regression round 10 closed — so it needs the overview tier to get genuinely better first.
- **Conflicts** the resolved fit-vs-floor decision (Option A; `6b52e29`, `7c06d48`, ledger "Graph-round decision framing"). **The trade-off**: no cut to explain versus illegible labels at the opening view. Recorded here rather than dropped.

### Information design of graph and review

#### CB-58 The readout node reserves an empty band between its kind and its channels
- **Class** `owner-design-decision` · **Dimension** Information design of graph and review · **Effort** M · **Impact** medium · **Needs binary** no · **Status** `owner decision`
- **Sources** IMP-17, BB-1 (one entry: BB-1 is the parked decision IMP-17 belongs to).
- **Where** Every review readout node (w-45's `CP-1`, w-33's `ch-u2`…), and the same markup in the plant graph's readouts.
- **Observed** Hunt: in the 2× crop `93-w45-node-2x.png` the header (id + kind) ends and ~30 CSS px of bare node surface runs before the first channel row, with no content and no separator; the ledger records this as P-12-2's 23 px reserved below the kind row for fault/effector/outcome lines, present whether or not the node has one. V-14 records the symptom as "tall nodes with empty bottoms". Source today: `Readout.tsx:80` always renders `<Note>` and `nodes.module.css:486-495` `.note` has `min-height: calc(var(--note-lines, 1) * 24px)` with `justify-content: center`.
- **Suggestion** IMP-17: collapse the band when there is no fault/effector/outcome line, or paint it as part of the header so it reads as structure. BB-1 is the reason it is not a fix: those lines appear **while a run plays**, so collapsing the band makes the node grow when a fault lands — the layout jump the reserved band exists to prevent. The real fix is a layout that measures the lines it will show, deferred deliberately with ISS-8D-17.
- **Conflicts** P-12-2 / ISS-8D-17 (tier-stable node layout, won't-fix rationale). **The trade-off**: a stable node that reserves space versus a compact node that moves. CB-13 is the same construct and is constrained by the same decision.

#### CB-59 The Link form offers bare instance ids
- **Class** `ui-presentation` · **Dimension** Information design of graph and review · **Effort** S · **Impact** medium · **Needs binary** no · **Status** `accepted`
- **Sources** IMP-12. (Hunt expectation row E-4 is the same ground, and its "lists every input" clause is stale — see Coherence check.)
- **Where** Editor → select an instance → "Link an output" → the "To" select.
- **Observed** Hunt: its 21 options are `["comp-u1","ch-u1",…,"mtr-h3"]` — ids only, no kind — and every other place the app names an instance it appends the kind (graph accessible names "comp-u1, air compressor, 1 input, 2 outputs", outline rows "comp-u1 air compressor", the run panel's fault select "STW-1 · storage tank"). Source today: `ConnectForm.tsx:47-51,118-126` filters the options to instances with a matching input port and renders `option.replaceAll("_", " ")` of `entry.id` — the kind is still missing. **The hunt's "unfiltered" half does not hold**: the filter is present and landed in `e9c8974` (2026-10-14 19:07), before the stream's own bundle was built.
- **Suggestion** Append the kind to each option ("ch-u1 · chiller"), matching the run panel's fault select. Full compatibility filtering is CB-67; the canvas already does the presentational half.
- **Conflicts** the hunt's E-4 row ("the form lists every input", ✗, `B`) — stale; and round 1's E4/14 ("compatible choices", "retain the filtered output/receiver/input choices") is right about the tree.

#### CB-60 The legend lists every port family in the product, not the plant's
- **Class** `ui-presentation` · **Dimension** Information design of graph and review · **Effort** S · **Impact** medium · **Needs binary** no · **Status** `accepted`
- **Sources** IMP-22.
- **Where** The Legend popover on the plant view, the editor and the review.
- **Observed** Hunt: the same 12 rows render everywhere reached — nine port families plus "Port not recorded / Filled: connected / Hollow: not connected" — on the 26-instance campus, on the 1-instance `dir-review-empty` plant and on a review, while water-pumping-station's wires measure exactly two families (`rgb(7,120,102)` Water and `rgb(143,98,0)` Power). So seven of the nine rows describe colours that are not on the screen. Source today: `GraphControls.tsx` `Legend` iterates a static `families` list, unfiltered.
- **Suggestion** Keep the vocabulary but mark the plant's own families — full contrast for the families present, dimmed for the rest — or lead with "In this plant: Water, Power" and keep the full list under it. The canvas already computes each wire's family to colour it, so no new fact is needed. Do not drop rows: a plant can gain a family on a composition edit.
- **Conflicts** none. Round 13/14's regression pass pins the marks (shape as well as hue) and the 12/14 rows; keep them.

#### CB-61 A selected connection is not explained as a model relationship
- **Class** `ui-presentation` · **Dimension** Information design of graph and review · **Effort** S · **Impact** medium · **Needs binary** no · **Status** `accepted`
- **Sources** r1-08. (Hunt expectation rows C-2 and C-6 are the same ground.)
- **Where** A selected wire on the graph, plant view and editor.
- **Observed** r1-08/E2, E4: arrowheads and a typed legend exist and the legend already explains direction. C-2 (✗): selecting a wire does not state "A supplies X to B" nor what the receiving model uses it for; C-6 (✗): model connections, physical wiring and consumer communication are not distinguished. The hunt did not exercise wire selection (C-2 "not reached").
- **Suggestion** On selecting a connection, state the exact source, value/unit, receiver and declared role ("a pressure output supplies a valve's pressure input"), and say that a model dependency is not necessarily a physical pipe and that graph placement does not imply topology. Reuse the inspector's existing link data rather than adding a panel.
- **Conflicts** none. r1-09's follow-path mode (CB-63) must not label a connected path as proven causal influence.

#### CB-62 A link's meaning cannot be previewed before it is created
- **Class** `ui-presentation` · **Dimension** Information design of graph and review · **Effort** M · **Impact** medium · **Needs binary** no · **Status** `accepted`
- **Sources** r1-14. (Hunt expectation row C-3 is the same ground.)
- **Where** Editor → "Link an output", before Link is pressed.
- **Observed** r1-14/E4: the filtered output/receiver/input choices work and creation reports success and a link count. What is missing is a sentence and a temporary graph preview before committing. Part of the "why" already exists: `ConnectForm.tsx` `TargetNote` renders `noTargetReason` when no target matches, and the empty options say "No instance takes it" / "No matching input". C-3 (✗) records no preview of both endpoints.
- **Suggestion** Show "source output → receiver input" with quantity and unit where known before Link, explain why a receiver is selectable or why no compatible destination exists, and keep the new relationship visibly selected long enough to inspect it after linking. Keep it adjacent to the existing form; do not add a second panel.
- **Conflicts** none. Round 1's C-4/C-5 (compatibility and required/optional vocabulary) are binary-owned and recorded as CB-67.

#### CB-63 No focused path through connected equipment
- **Class** `feature-or-binary` · **Dimension** Information design of graph and review · **Effort** L · **Impact** medium · **Needs binary** no · **Status** `accepted`
- **Sources** r1-09.
- **Where** Plant view and review.
- **Observed** r1-09/E2, E7: incident-edge highlighting works (P-4 ◐, selection highlights; the hunt measured 3 px family-coloured wires for the selection and 1 px grey for the rest), but there is no way to keep a source-to-receiver path prominent.
- **Suggestion** An optional "Follow inputs/outputs" mode on top of the existing emphasis: keep the selected component and its immediate neighbours prominent, show a short named path, and offer one return to overview. Follow documented links only; never label a connected path as proven causal influence. Unrelated equipment stays available but visually secondary.
- **Conflicts** none.

#### CB-64 No channel trends tied to events
- **Class** `feature-or-binary` · **Dimension** Information design of graph and review · **Effort** L · **Impact** medium · **Needs binary** no · **Status** `accepted`
- **Sources** r1-23. (Hunt expectation row V-5, "◐ partly", and VIS-7, "◐ partly", are the same ground.)
- **Where** Review → Inspect for a selected channel; the timeline.
- **Observed** r1-23/E7: inspection is snapshot-only, and V-5 records that the small trend line across the run and the fault/alarm marks are still missing; VIS-7 records no separate fault/alarm/effector lanes and no compact trends. The observed series is already read for the review (the readouts and the KPI tables render from it), so no new binary fact is obviously needed — but V-5's `B` marks true-state-beside-observed as binary, so keep the observed and true series distinct and expose truth only when allowed.
- **Suggestion** An explicit route from a selected observed channel to its time trend, synchronized with playback and the event marks; label axes and units, preserve missing intervals, mark observation time separately from the playhead, and keep observed and true series distinct.
- **Conflicts** none. Do not let it re-open the opening-zoom rules; it is a panel, not a graph change.

#### CB-65 The review does not open with an evidence-linked result summary
- **Class** `ux-polish` · **Dimension** Information design of graph and review · **Effort** M · **Impact** high · **Needs binary** no · **Status** `accepted`
- **Sources** r1-22.
- **Where** Review → the report's first screenful.
- **Observed** r1-22/E7, E10: the report is table-first; the reader must assemble the sequence. V-6 (✅) records that the report leads with what happened and integrity is one line, so the complaint is about the summary's shape, not its existence. Round 2's S8 shows the pieces that exist: clicking a timeline event seeks, selects the equipment and highlights the matching report row.
- **Suggestion** Lead with completion, verification scope, scoring availability, the important recorded events and any remaining active alarms, with each item able to seek/select its evidence — for w-46, the flame failure at 00:05:00, its end at 00:05:10 and the alarm at 00:08:35 as a sequence. Do not assert causation merely because events are ordered. r1-22's acceptance also asks that a verified recording be distinguishable from a successful controller or a healthy plant.
- **Conflicts** none. V-6 and V-13 are the expectation rows it serves.

#### CB-66 The timeline label does not say what an event was
- **Class** `ui-presentation` · **Dimension** Information design of graph and review · **Effort** S · **Impact** medium · **Needs binary** no · **Status** `accepted`
- **Sources** R2-22. (Round 1's 22 is the proposal it refines; that is CB-65.)
- **Where** Review → the timeline's event marks and any selected-event summary.
- **Observed** R2-22/S8: visible chips show time and equipment ("00:18:00 pump-cond") while the full fault or command name lives in the accessible name/tooltip and the report; a reader cannot tell what happened without hovering or searching. V-3's marks and legend are met (two labelled fault marks, one effector mark, legend "Fault starts" / "Effector call"), and V-10's mark chips are centred on their marks.
- **Suggestion** Put a short event description such as "Seal leak" in the visible label or a nearby selected-event summary, with full detail on focus, and distinguish the selected event from one currently active at the playhead. Preserve the existing linked report-row highlighting.
- **Conflicts** none.

#### CB-67 Compatibility-aware linking needs the binary's rule
- **Class** `feature-or-binary` · **Dimension** Information design of graph and review · **Effort** L · **Impact** medium · **Needs binary** yes · **Status** `owner decision`
- **Sources** BB-3. (Hunt expectation rows E-4, C-4 and C-5 are the same ground as evidence.)
- **Where** Editor linking (E-4, C-4) and the inspector's port vocabulary (C-5).
- **Observed** BB-3: the canvas filters presentational compatibility by port type (`ConnectForm.tsx:47-51`, from the component specs it loads), but the real rule — which inputs are required, optional or default-fed, and why a link is refused — belongs to the binary. C-4 (✗, `B`) and C-5 (✗, `B`) both stand.
- **Suggestion** Have the binary expose the compatibility rule and the required/optional/default-fed vocabulary, then list only compatible inputs and explain a refusal in its words. CB-59 (kind on the options) is the presentational half available now; this entry is the capability.
- **Conflicts** none. It corrects the hunt's E-4 row in one direction (the canvas already filters by type) and confirms its `B` mark in the other (the rule itself is binary-owned).

## Equivalence map

Every source item appears exactly once. Four tables, one per source document/section.

### Hunt — `## Improvements`

| Source | Item | CB |
|---|---|---|
| IMP-01 | The equipment chooser hides most of the library with no cue | CB-01 |
| IMP-02 | The keyboard dialog does not dim the page | CB-02 (rejected) |
| IMP-03 | The keyboard dialog's first focus is invisible | CB-03 |
| IMP-04 | The graph's zoom controls are the only icon-only controls without a tooltip | CB-04 |
| IMP-05 | "Did not finish" shares the success green with "Verified" | CB-16 |
| IMP-06 | The run panel's sticky footer covers the last section's text and its button | CB-18 (done) |
| IMP-07 | The right panel scrolls with no cue that anything is below | CB-19 |
| IMP-08 | The outline rail and the right panel size themselves by different rules | CB-20 |
| IMP-09 | One panel is called two things | CB-27 |
| IMP-10 | The About-this-run sentence does not parse | CB-39 |
| IMP-11 | The graph's screen-reader instructions promise deletion on read-only screens | CB-46 |
| IMP-12 | The Link form offers bare instance ids | CB-59 |
| IMP-13 | The empty Halls group reads "0 · empty" | CB-31 |
| IMP-14 | The plant row's not-runnable sentence is a run-on | CB-37 |
| IMP-15 | The review says "Verified" twice in adjacent lines | CB-38 |
| IMP-16 | A doomed domain probe leaves a 404 and a console error on the normal editor path | CB-41 |
| IMP-17 | The readout node carries an empty band between its kind and its channels | CB-58 |
| IMP-18 | One area is typed three ways on one screen | CB-52 |
| IMP-19 | The same plant's outline is ordered differently on the plant screen and in the editor | CB-53 |
| IMP-20 | The same link is written in two directions | CB-54 |
| IMP-21 | A channel with no reading is a bare dot whose meaning is only in a tooltip | CB-51 |
| IMP-22 | The legend lists every port family in the product, not the plant's | CB-60 |

### Hunt — `## Bigger bets` (labelled BB-1…BB-7 here)

| Source | Item | CB |
|---|---|---|
| BB-1 | A tier-aware node header, or no reserved band at all | CB-58 |
| BB-2 | Fit-first opening on wide plants | CB-57 |
| BB-3 | Compatibility-aware linking | CB-67 |
| BB-4 | Serve the bundle compressed | CB-15 |
| BB-5 | A run date in the About-this-run popover | CB-39 |
| BB-6 | A ready-to-run example and a guided journey | CB-40 |
| BB-7 | Reference screens and a state sheet | CB-50 |

The hunt's `## Quick wins` list (20 checkboxes) is a re-ordering of IMP-01…IMP-22 and adds no item. Its
`## Expectations still not met` table (85 `UI-EXPECTATIONS.md` rows) and `## Checked and holding` are evidence and
status, not proposals; the rows in them that produced a proposal are absorbed above, and the rows that did not are
listed under Coverage.

### Comprehension round 1 — `### 01.`…`### 25.`

| Source | Item | CB |
|---|---|---|
| r1-01 | Give an unfamiliar plant a short introduction | CB-25 |
| r1-02 | Name the current activity consistently | CB-28 |
| r1-03 | Preserve a visible route back to the user's context | CB-56 |
| r1-04 | Explain domain terms exactly where they become relevant | CB-29 |
| r1-05 | Make the workspace answer "which plant or run do I need?" | CB-42 |
| r1-06 | Show what a starting plant contains before copying it | CB-05 |
| r1-07 | Keep graph meaning readable as zoom changes | CB-49 (already fixed) |
| r1-08 | Explain a connection as a model relationship | CB-61 |
| r1-09 | Offer a focused path through connected equipment | CB-63 |
| r1-10 | Connect outline search to what users see on the graph | CB-43 |
| r1-11 | Explain missing readings instead of relying on punctuation | CB-51 |
| r1-12 | Pair equipment IDs with understandable roles | CB-30 |
| r1-13 | Organize editing around the user's current task | CB-26 |
| r1-14 | Preview the meaning of a link before creating it | CB-62 |
| r1-15 | Explain the pending changes and what Save does | CB-22 |
| r1-16 | Let users compare equipment choices by meaning | CB-06 |
| r1-17 | Explain Explore/Evaluate and whether scoring is actually available | CB-47 |
| r1-18 | Make simulated time and waiting time unmistakable | CB-32 |
| r1-19 | Preview the whole planned experiment before starting | CB-45 |
| r1-20 | Describe command parameters and intended effects plainly | CB-33 |
| r1-21 | Make the running screen explain meaningful changes | CB-44 |
| r1-22 | Open review with an evidence-linked result summary | CB-65 |
| r1-23 | Connect events to synchronized channel trends | CB-64 |
| r1-24 | Explain outcomes and KPIs without overstating evidence | CB-34 |
| r1-25 | Give incomplete or interrupted work a clear next step | CB-48 (already fixed) |

### Comprehension round 2 — `R2-01`…`R2-25`

| Source | Item | CB |
|---|---|---|
| R2-01 | Adapt the desktop composition to the task | CB-24 |
| R2-02 | Make desktop panel widths adjustable and recoverable | CB-68 |
| R2-03 | Give short inspector content a deliberate layout | CB-21 |
| R2-04 | Keep task actions together and visually anchored | CB-22 |
| R2-05 | Replace most area boxes with lighter grouping | CB-12 |
| R2-06 | Flatten ordinary equipment nodes | CB-11 |
| R2-07 | Reduce repeated header bands and internal framing | CB-13 |
| R2-08 | Stop boxing every toolbar action | CB-14 |
| R2-09 | Give every persistent colour accent one explainable meaning | CB-17 |
| R2-10 | Rename the fault catalogue so it cannot look like an incident count | CB-07 |
| R2-11 | Separate equipment failures from measurement disturbances | CB-08 |
| R2-12 | Call model states variables, not operating conditions | CB-09 |
| R2-13 | Make available commands look different from executable controls | CB-10 |
| R2-14 | Put a one-sentence functional explanation before technical detail | CB-30 |
| R2-15 | Distinguish no direct equipment from an empty area | CB-31 |
| R2-16 | Keep area hierarchy recognizable between view and edit | CB-52 |
| R2-17 | Make incomplete status look incomplete | CB-16 |
| R2-18 | Separate run mode from current evidence availability | CB-47 |
| R2-19 | Keep selected-object identity visible while reading details | CB-23 |
| R2-20 | Read every connection in the same source-to-receiver order | CB-54 |
| R2-21 | Name panel-back actions by their destination | CB-55 |
| R2-22 | Put the event's meaning in its visible timeline label | CB-66 |
| R2-23 | Label snapshot readings and whole-run summaries separately | CB-34 |
| R2-24 | Separate scenario facts from observations | CB-35 |
| R2-25 | Say no scheduled faults without promising plant health | CB-36 |

## Coherence check

Contradictions and duplications between the streams, and how each was resolved. Nothing here was dropped; where a
stream opposes a recorded decision, the entry exists and the trade-off is named.

1. **The Link form filters — two streams say it does not.** The hunt (IMP-12, E-4) reports the "To" select listing all
   21 other instances as bare ids, "unfiltered by compatibility"; round 1 (E4, 14) reports compatible choices and asks
   to retain them. Resolved for the tree: `ConnectForm.tsx:47-51` filters targets to instances with an input port
   matching the chosen output's type, and that landed in `e9c8974` (2026-10-14 19:07) — before either stream's 22:16
   bundle. With **no output chosen** the type is empty and `portsMatching` (`ports.ts:77-81`) legitimately returns every
   input, which is the state the hunt read. So round 1 is right, the hunt's "unfiltered" clause and E-4's recorded ✗
   ("the form lists every input") are stale, and IMP-12's surviving half is the bare id with no kind. CB-59 keeps that
   half; CB-67 is the binary rule behind it.
2. **One link, written in two directions.** All three streams see it (hunt IMP-20, round 1's C-2/08, round 2's
   S7/R2-20) and all three want the same thing; merged into CB-54. The editor's arrow is keyed on direction
   (`InstanceEditor.tsx:108`) while the inspector always writes source → destination (`InspectorLinks.tsx:40`).
3. **The area label: leaf or full path?** The hunt (IMP-18) wants the leaf in the rail and on the frames with the path
   only in the select and the AREAS list; round 2 (R2-16) wants the hierarchy visible in the outline itself, "avoid
   making the Area dropdown the only place the full relationship is visible". Merged into CB-52 with the choice named:
   leaf plus the path in the selected-object location line and the select (smaller, keeps the rail's scan rhythm) —
   round 2 proposed path-in-outline. The tree already shows the leaf in the rail, so the real change is the uppercase
   section-title transform plus showing unpopulated parents.
4. **Editor rail order versus a recorded decision.** IMP-19 asks for the plant's area order in the editor rail;
   `6611972` deliberately sinks empty areas below populated ones there, and `3dc9570` gives the plant rail and the
   inspector's AREAS block one declared order. Surfaced as CB-53, `owner decision`. **Trade-off**: one order shared by
   both rails and the AREAS block, versus keeping an author's still-empty areas out of the way — which CB-31/R2-15 also
   wants visible. CB-31 is the smaller fix and does not need this decision.
5. **The unfinished chip: green, and what is it keyed on?** The hunt files it as Colour (IMP-05); round 2 files the same
   chip as a reproduced visual finding (S4/R2-17). One entry (CB-16), in flight. The nuance both streams elide: the
   chip renders on `!row.run.ended` (`RunList.tsx:34-38`) and is classed `.live`, so the listing may mean a **still
   running** run; round 15 investigated and recorded that the listing carries no abandonment signal. A warn treatment
   must therefore not mislabel a genuinely live run, and that distinction is binary-owned (shared with CB-48).
6. **Does the Keyboard dialog dim the page?** The hunt measures `dialog::backdrop` at `rgba(0,0,0,0)` and concludes the
   page is not dimmed (IMP-02); rounds 13 and 14 describe "a real `<dialog>` with a scrim". Both are right about their
   own element: the dialog is opened with the `open` attribute, so `::backdrop` paints nothing, and the dim comes from
   the sibling `.scrim` div (`KeyHelp.module.css:1-6`, `--scrim` at `tokens.css:75, 132`). Resolved in favour of the
   source: CB-02 is `rejected`/`stale`, and G-15's "no scrim" clause is already met. (Its focus-ring clause is CB-03.)
7. **Reserve the footer or move it?** IMP-06's suggestion is to reserve the sticky footer's height as `padding-bottom`;
   `bfabd05` measured that first and found it changed nothing at the default position while adding 115 px of dead
   scroll, so it moved the footer outside the scroll region instead. CB-18 records the landed fix and the rejected
   alternative.
8. **One width rule, or adjustable panels?** IMP-08 (one rule for both panels, the rail's rule being smaller) and
   R2-02 (a resize handle, remembered, with Reset) propose different fixes for one element, and they are **different
   classes of change**: IMP-08 is presentation (CB-20), R2-02 is a new capability (CB-68, `behaviour-change`). The
   coordinating agent verified in code that there is no drag-resize anywhere (`Shell`/`PlantScreen`/`EditorScreen` hold
   no resize logic and no CSS `resize:`), and that §3.2's "panels resize" is met today only by the collapse/expand
   toggles (`PanelToggles`), which do change the layout (collapsing the outline measured `pastRight 237 → 0` in round
   13). They were therefore **kept separate rather than merged**: putting a new capability in the presentation batch is
   exactly how a batch stops being presentation-only. R2-01 (CB-24) is the umbrella decision over both.
9. **Motion: "0 transitions" versus the tree.** The hunt's Motion verdict states "0 elements carry a non-zero
   transition or animation anywhere reached"; the tree gives every node `transition: border-color var(--motion-fast)`
   with `--motion-fast: 120ms` (`nodes.module.css:7`, `tokens.css:23`), zeroed under `prefers-reduced-motion`
   (`tokens.css:139-141`). No proposal follows, because 120 ms is inside §3.2's ≤ 150 ms bar, so the verdict ("nothing
   to improve") stands and only the absolute claim does not. Recorded so a later round does not measure the same
   surprise.
10. **The opening viewport.** The hunt's P-3 (41 % empty above and below the campus frames) and V-9 ("clipped at
    1280") and round 2's S11 sit on the same 0.4 opening floor; rounds 13/14 measured that floor plus its pan cue as
    intended and identical on the plant view and in the editor. The ledger records the rule as the resolved Option A
    (`7c06d48`, `6b52e29`). Surfaced as CB-57, `owner decision`, rather than dropped or smoothed.
11. **The equipment listbox: two fixes, one of them opposes a decision.** IMP-01 offers a boundary + count **and**
    raising `size` past six rows. The second opposes `d8a8549`. CB-01 keeps the first and names the conflict.
12. **The reserved note band collides with a deliberate rationale.** IMP-17 and BB-1 ask to collapse the empty band;
    the same hunt records that the lines using it appear while a run plays and that round 16 deferred a measuring
    layout with ISS-8D-17. CB-58 keeps both halves and marks the entry `owner decision`; CB-13 (R2-07) touches the same
    construct and is constrained to change paint, not the reserved height.
13. **One vocabulary for the command words.** Round 1 (r1-04) prefers "Scheduled commands" for effector calls; round 2
    (R2-13) wants the view inspector's catalogue entries labelled "Available commands". They are not two renames of one
    string, so one rule is stated once and used by both: a supported effector in a catalogue is an **available
    command** (CB-10), one placed in a plan is a **scheduled command** (CB-29), and the technical object stays an
    **effector call** in details and reports.
14. **Two fixes for one count element.** IMP-13 wants the state word dropped from the count; R2-15 wants a
    direct/descendant distinction or a child-area count. Merged into CB-31; the merged suggestion drops the word and
    takes R2-15's scope, because that is the fix for the confusion round 2 reproduced (a parent with children reading
    as empty).
15. **The hunt's G-4 and round 2's S10 disagree about one sentence.** The hunt records "None: the plant runs healthy."
    as a designed empty state that is met (G-4 "closed"); round 2 says it promises health (S10/R2-25). CB-36 keeps
    round 2's reading because it is the stricter one and the owner has asked for comprehension over a checklist; the
    change is copy-only and in flight.
16. **Already done or stale, with the commit that did it.**
    - `bfabd05` — the run plan's footer (CB-18); it cites IMP-06 by name.
    - Uncommitted working-tree edits by another agent — the zoom control tooltips (CB-04) and the graph's select-only
      screen-reader words (CB-46); `git status` shows `GraphControls.tsx`, `GraphControls.test.tsx`
      and `PlantGraph.tsx` modified with an untracked `PlantGraph.test.tsx`. The CB-46 text was latent DOM text
      (`display: none`, nothing referencing it by id), so that entry's impact is `low`, not the hunt's low-medium.
    - Declared in flight by the delegating agent — the unfinished chip's token (CB-16) and the run panel's fault copy
      (CB-36).
    - `a13c028` (the editor's empty first step, E-10 ✅) and `969531d` (the links empty-state sentence, E-12 ✅) were
      committed after the 22:16 bundle, so any empty-editor or links-empty-state observation in the three streams
      describes a superseded build. `6921832` (the wire-under-node-bodies e2e check) and `3520ff3` (expectation rows)
      are likewise later than the bundle.
    - r1-25 (CB-48) and r1-07 (CB-49) are already satisfied by the shipped behaviour; both are recorded as
      `already-fixed`/`stale`.
    - The hunt's IMP-02 (CB-02) is rejected on the source.
17. **One surface, two fetch-gate scopes.** Round 13's `17f2274` closed the review's doomed fetches (no `run.json`, no
    `run.verify` for a listing-known unfinished run) and round 14 re-verified it. The hunt's IMP-16 is a **different**
    surface — the editor's `versionOf(domainOf(doc))` probe — so the round-13 claim does not cover it. Not a
    contradiction between streams, but the reason CB-41 is still open while the same class of defect is closed
    elsewhere.

## Recommended next batch

The smallest coherent set inside the owner's current band: **UI/UX presentation, no behaviour change, no binary fact**,
excluding everything already in flight and everything needing a design decision. All Effort S. Order is by
(benefit × safety), not by source order.

| # | CB | Change | Why now |
|---|---|---|---|
| 1 | CB-41 | Answer the composed-domain question from the workspace listing so no doomed probe is issued | Removes a measured 404 and console error on a normal path; §3.1 is a standing bar and this is the only request the hunt found; no user-visible behaviour changes |
| 2 | CB-03 | Move the Keyboard dialog's first focus to Close, or ring the focused heading | §3.3 / G-8 / G-15; one CSS or one `ref` change; the last focus-visibility gap the hunt found |
| 3 | CB-51 | Paint "not observed" (or a dimmed em dash) instead of the bare `·`, and keep "no reading yet" distinct | G-10's "a tooltip must never be the only copy of a fact" at the fact that decides missing versus zero; medium-high impact for one glyph |
| 4 | CB-19 | Reuse the outline rail's cut fade on the right panel's bottom edge | The shell already owns the pattern and the listener; 567 px of the Build panel and 120 px of the run plan are below the fold today |
| 5 | CB-27 | One name, "Inspector", in the keyboard help, the button title and the review's region label | G-6's one action, one name; the strings already exist; `keyList.ts:23-24` still says "details panel" |
| 6 | CB-38 | Let the integrity sentence drop the repeated "Verified:" | One string plus its test; the badge already carries the word |
| 7 | CB-54 | Write links source-output → receiver-input in both places | One coherence rule the legend already states; the two rows render one relationship two ways |
| 8 | CB-20 | One width rule for the rail and the right panel (the fixed rule; not the resize handle) | G-2 / VIS-10; removes a 37 % panel-width swing for identical content |
| 9 | CB-31 | Drop "empty" from the count and show the direct/child distinction | Fixes the "Cold rooms is empty but holds Room A and Room B" reading; the outline already builds the tree |
| 10 | CB-59 | Append the kind to each instance option in the Link form | The form is already filtered; the missing kind makes one choice between 21 look-alike ids |
| 11 | CB-37 | Make the not-runnable plant row one state and one action | One string in `workspace.ts:70`; removes a three-job clause from a row of single facts |
| 12 | CB-39 | Split the About-this-run time clause into labelled facts (the **sentence** half only) | The popover is the one place a run's provenance can be checked; the run **date** is the binary half (BB-5) and stays out |

**Already in flight — do not duplicate:** CB-04 (zoom tooltips, uncommitted), CB-46 (graph screen-reader words,
uncommitted), CB-16 (unfinished chip token), CB-36 (run panel fault copy). **Already done:** CB-18 (`bfabd05`).
**Rejected:** CB-02.

If the batch is taken in two passes, do 1–4 first: they are the only ones that touch a standing-bar clause (§3.1 no
failed request, §3.3 focus visible, G-10 no tooltip-only fact) rather than a polish call.

## Do not do / not now

| Entry | Reason |
|---|---|
| CB-02 | Rejected: the premise does not hold. The page is dimmed by the `.scrim` div; `dialog::backdrop` is transparent because `showModal()` is not used. |
| CB-53 (editor rail order) | Opposes a deliberate decision (`6611972`, `3dc9570`). One order for both rails versus keeping an author's empty areas out of the way is the owner's call. |
| CB-57 (fit-first opening) | Opposes the resolved Option A fit-vs-floor decision (`6b52e29`, `7c06d48`); re-introduces sub-0.4 tiers with 8 px labels unless the overview tier improves first. Needs the owner's word. |
| CB-58 (collapse the readout band) | Opposes ISS-8D-17's won't-fix rationale: the band's lines appear while a run plays, so collapsing it makes the node grow when a fault lands. The real fix is a measuring layout, deferred deliberately. |
| CB-11, CB-12, CB-13, CB-14 as four separate fixes | They are one visual direction (round 2's R2-05–R2-08, which says to prototype R2-01–09 together on one plant). Take them as one prototype, settle CB-13 against P-12-2 first, and do not apply globally before that. R2-06's own acceptance says prototype one representative plant. |
| CB-24 (task-adapted composition) | A product decision (L), and it subsumes CB-20/CB-21/CB-26. Do those three as the cheap path or take the umbrella as one decision — not both by accident. |
| CB-68 (adjustable panel widths) | A new capability (`behaviour-change`), not presentation: there is no drag-resize in the shell today, so it needs a handle, a keyboard equivalent, remembered preferences and a Reset, all stated first. Keep it out of a presentation batch. |
| CB-08, CB-67 | Need the binary's authoritative facts (fault classification; the compatibility rule and required/optional/default-fed vocabulary). R2-11 and BB-3 both forbid guessing them in the client. |
| CB-40 (example + guided journey) | New content and interaction (L), and it overlaps CB-05 and CB-25. Take the three as one first-use decision. |
| CB-25 (plant introduction) | Needs an authored purpose the canvas does not have; without it the canvas would have to invent a process story, which r1-01 explicitly forbids. |
| CB-42, CB-43, CB-44, CB-45, CB-56, CB-63, CB-64 | Behaviour changes or new features (search, search-on-graph, recent-event strip, experiment preview, context retention, follow-path, trends). Each is worth doing, none belongs in a presentation batch, and CB-44 and CB-45 rest on walkthrough evidence the hunt could not measure (zero runs started). |
| CB-15 (serve compressed) | A Go transport change; the gzip budget is already met, so it is a correctness-of-appearance fix, not a UI one. |
| CB-47 | Presentation, but it must not be done separately from the truth-boundary rules; it needs the binary's "is an outside system connected" fact for the pre-run half. |
| CB-48, CB-49 | Already fixed. Do not re-open. The one open clause inside r1-25 (running vs did-not-finish) needs a binary signal, not UI work. |
| CB-50 (reference screens) | A documentation artifact, not UI. The owner should decide whether it is funded; it must not arrive disguised as a UI change. |
| CB-01's "raise `size` past six rows" variant | Opposes `d8a8549` (six visible rows, deliberate). The boundary plus the count is the change to make. |
| CB-16's "just make it orange" shortcut | The chip is keyed on `!run.ended`, which may be a genuinely running run; a warn token without the abandoned/running distinction would mislabel it. Take the token change with that caveat, not instead of it. |
| CB-59's "filter the To list" as if it were missing | The filter exists (`e9c8974`); rebuilding it would duplicate working code and re-open a closed expectation row. |

## Coverage

**Source counts**

| Source | Items |
|---|---|
| `improvement-hunt-20261015.md` → `## Improvements` | 22 (`IMP-01`…`IMP-22`) |
| `improvement-hunt-20261015.md` → `## Bigger bets` | 7 (its numbers 1–7; `BB-1`…`BB-7` here) |
| `COMPREHENSION-IMPROVEMENTS-2026-10-15.md` | 25 (`### 01.`…`### 25.`; `r1-01`…`r1-25` here) |
| `COMPREHENSION-ROUND-2-2026-10-15.md` | 25 (`R2-01`…`R2-25`) |
| **Total source items** | **79** |

Not counted as items: the hunt's `## Quick wins` (a re-ordering of the same IMPs), its `## Expectations still not met`
table (85 re-measured `UI-EXPECTATIONS.md` rows), `## Per-dimension coverage`, `## Checked and holding` and the closing
"one measurement that could not be taken cleanly" note (status and evidence); round 1's and round 2's evidence tables
(E1–E10, S1–S11), their dimension-coverage tables and their delivery-order paragraphs (evidence and sequencing).

**Entries produced: 68 distinct `CB-NN`.**

Eighteen sources were absorbed by eleven many-to-one entries:

| Entry | Absorbs |
|---|---|
| CB-16 | IMP-05 + R2-17 |
| CB-22 | R2-04 + r1-15 |
| CB-30 | r1-12 + R2-14 |
| CB-31 | IMP-13 + R2-15 |
| CB-34 | r1-24 + R2-23 |
| CB-39 | IMP-10 + BB-5 |
| CB-47 | r1-17 + R2-18 |
| CB-51 | IMP-21 + r1-11 |
| CB-52 | IMP-18 + R2-16 |
| CB-54 | IMP-20 + R2-20 |
| CB-58 | IMP-17 + BB-1 |

79 − 11 = **68**, which is the `CB-NN` count (CB-20 and CB-68 are separate entries for one element, so no merge is
counted there). **Every one of the 79 source items appears in exactly one row of the equivalence map above**
(22 + 7 + 25 + 25 = 79 rows, and no source id appears twice).

**Entries by dimension** (the loop's ten): Components 14 (CB-01…CB-14) · Motion and responsiveness 1 (CB-15) ·
Colour and contrast 2 (CB-16, CB-17) · Layout 10 (CB-18…CB-26 plus CB-68, out of id order — see CB-68's note) ·
Content design 14 (CB-27…CB-40) · Feedback and flow 8 (CB-41…CB-48) · Craft 2 (CB-49, CB-50) · Typography 1 (CB-51) ·
Whole-product coherence 6 (CB-52…CB-57) · Information design of graph and review 10 (CB-58…CB-67).

Two dimensions carry only the entries this document placed there, and the placement is worth stating plainly:
**Motion and responsiveness** — the hunt measured nothing to improve (with the transition caveat in the coherence
check), and no round-1 or round-2 item asks for motion as its primary change, so its one entry (CB-15) is the
transport item the hunt filed as performance. **Typography** — the hunt measured the type scale clean and
"nothing to improve" (sizes at 11/12/13 px, ids and values mono, units 11 px in a lighter tone, numbers tabular and
right-aligned), so its one entry (CB-51) is the missing-reading glyph; the remaining typographic asks inside round 1
(casing, labels, units) sit in the Content design and Components entries by their dominant change.

**`UI-EXPECTATIONS.md` rows the hunt re-measured as still open, with no proposal in any stream.** The hunt's
`## Expectations still not met` table marks 85 rows; most are "closed" on today's build. The rows below are the ones it
left as "holds", "partly", "not reached" or "not reproducible" **and** that none of the three streams turned into a
proposal. They are recorded here so the re-measurement is not lost.

| Rows | State, and why there is no entry |
|---|---|
| G-1, VIS-1 | Shell/breadcrumb pattern questions (each screen's second bar; the editor drops the summary where the save state sits). No stream proposed a fix; CB-28 touches the same ground from the activity-naming side. |
| G-3, G-5, G-17 | Raw ids still appear in the editor's instance lists and area paths; no in-place progress was observed (the slowest reads were 7 ms and 99 ms); value origins were not exercised. CB-52/CB-59 touch the id cases; the rest have no proposal and G-17 was not reached. |
| E-5, E-6, E-9, E-14 | Deleting a wire from the graph was not exercised; no composition in the workspace carries a validation issue; no unsaved state was created; the "wires under node bodies" exclusion is not reproducible in the review workspace (the ledger's P-12-3); `6921832` added the e2e check over the fixture plant that does have it. No proposals. |
| C-2, C-3, C-4, C-5, C-6 | Wire selection, first-connection preview and the compatibility/vocabulary rows were not exercised in the hunt; C-2/C-3/C-6 are CB-61/CB-62, and C-4/C-5 are CB-67. |
| R-3, R-4, R-5, R-6, R-9, R-10 | Running-state rows: no run was started in the hunt's budget (0 runs), so picking a fault instance from the graph, progress/speed/ticks, the finish status, field-level problems and the problem banner were not seen. CB-44 (running screen) is the only proposal that touches them, and it rests on round 1's walkthrough instead. |
| V-5, V-8, V-9, V-14 | The trend line and fault/alarm marks in Inspect (CB-64); implausible values flagged (needs `factory-utilities`, absent from the review workspace, and binary-owned, V-8 `B`); the review opening cut (CB-57); full kinds and even node heights (CB-58). |
| A-2 | The bridge never stopped during the hunt; no proposal in any stream. |
| N-1, N-2, N-3, N-4, N-6 | First-use and creation rows: CB-40 (example, journey), CB-05 (creation preview, human name first), CB-06 (equipment choice meaning). |
| P-3, P-12, W-5, W-8, VIS-6, VIS-7, VIS-10, VIS-11 | The opening band and review clipping (CB-57); hall nodes' open affordance and name case (CB-52); New plant prominence and an empty workspace (not reached; no proposal); readable hall summaries (no proposal; CB-52 covers the names); review lanes and compact trends (CB-64); resizable panels (CB-68); reference screens (CB-50). |

**What this document did not reach.** Stated plainly rather than left thin:

- No browser was opened and nothing was re-measured; `canvas/dist` was not touched. Every observation is a stream's own
  measurement or a file-and-line reading, and the streams' bundle (the 22:16 build) is no longer in the tree.
- Three entries rest on a single stream's walkthrough rather than two independent measurements: CB-44 (the running
  screen, r1-21/E6 — the hunt reached no running state), CB-45 (the experiment preview, r1-19/E6, E10) and CB-63
  (r1-09). They are the weakest evidence in the list and are marked as such in their entries.
- I did not trace every round-1 and round-2 claim to source: r1-02 (activity naming), r1-05 (workspace search),
  r1-09, r1-19, r1-21, r1-22 and R2-04's toolbar geometry are recorded from the streams as they wrote them. The
  coherence check states which contradictions I could settle in code (the Link-form filter, the dialog scrim, the link
  direction, the area count, the reserved note band) and which I could not (the running states, the hunt's FCP
  measurement).
- The hunt's own untaken measurement ("shell first render", FCP 1,448 ms / 1,972 ms versus a §4 budget of 1.5 s) is not
  an entry: it is reported as measured and attributed to the harness, and nothing was added for it.

**Nothing in this document is a measurement.** Every number in `Observed` is attributed to the stream that took it, or
to a file and line in the tree; where the two disagree the Coherence check says so. The streams' own limits carry over:
the hunt started no runs, created no plants and opened every editor read-only, so no running state, no plan problem and
no unsaved-leave state was observed, and `factory-utilities` was absent from its workspace; both comprehension rounds
were expert walkthroughs against a served build that no longer exists in the tree, with validation testing excluded and
no participant study.
