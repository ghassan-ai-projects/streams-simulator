# Canvas UI expectations

This is the bar the canvas is held to. Every review round audits the build against it, item by item, and cites the IDs.
The scope is desktop Chrome only (1280–1920 px, light and dark) and the four flows: workspace, build a plant, run a
plant, review a run.

How to read the tables:

- **Now** is the status found in the 2026-10-14 walkthrough in Chrome:
  - ✅ met;
  - ◐ partly met;
  - ✗ not met.
- **B** in the binary column marks an item that needs a binary or bridge capability first, because the canvas
  computes nothing the binary owns.

## Owner decisions (2026-10-14)

- **Plant model:** a plant is one object on the workspace. Its composition and composed domain sit behind one row
  with Open, Edit and Run, and files are secondary detail (W-1, W-2, W-7).
- **Review:** a review opens at 0 (or the first event) with Play. The report leads with what happened, and integrity
  collapses to one line (V-2, V-6).
- **Order:** workspace and wording first.
- **Out of scope for this goal:**
  - area editing (E-8, second sentence);
  - true state beside observed in Inspect (V-5, last sentence);
  - wire routing around nodes (P-4, first sentence; highlighting stays in scope).
- **Fan experiment:** The simulated fan experiment (review §9, milestones M4–M5) is a separate goal; the canvas keeps an example entry ready and labels unavailable stages.

## 0. Principles

**Product model.** Build a simulated plant, generate its readings, test how a connected system responds, and review what happened. Three objects carry it:

- **Plant:** the equipment, connections and modelled behaviour.
- **Run plan:** the duration, repeatability settings and disturbances for one experiment.
- **Run:** the recorded execution and its results.

"New simulation" is an entry into these choices, not a fourth object.

1. **The user is an engineer, not a developer.** They think in plants, equipment, faults and runs; never in files, JSON,
   world ids, digests or schema versions. Developer facts may be shown, but only in a secondary place (a details
   disclosure or a tooltip), never as the headline.
2. **One object, one name.** A *plant* is one thing the user can open, edit, run and review. Its composition and its
   composed domain are two files behind one plant, not two items in two lists.
3. **Every screen answers three questions at a glance:** where am I, what am I looking at, what can I do next.
4. **Every state is designed:** loading, empty, ready, unsaved, busy, refused, failed, done. A blank area is never a
   state.
5. **Calm and dense like a desktop tool.**
   - Neutral surfaces, colour only for meaning (state, port family, selection).
   - One accent for the primary action per screen.
   - Nothing bigger than it needs to be; nothing truncated that matters.
6. **Mouse and keyboard are equals.** Every action has a visible control and a keyboard path; focus is never lost.
7. **The binary is the authority.** Validation, composition, truth, scores and verdicts come from `streamsim`; the
   canvas shows them clearly and never invents them.

## 1. Whole application

| ID | Expectation | Now | B |
|---|---|---|---|
| G-1 | A persistent top bar shows the product, the workspace name, and a breadcrumb of where I am (Workspace › Plant › Run or Editor). It also holds the theme switch. Every screen uses the same bar. | ◐ the breadcrumb sits inside each screen's own bar and differs per screen | |
| G-2 | Screen layout: left rail (outline), centre (graph or content), right panel (context), each with a fixed, consistent width. At 1920 px the centre takes the extra space and nothing floats in empty margins. | ◐ rail and panel widths are one token on plant, editor and review; not resizable | |
| G-3 | Typography: one scale. Headings name things in title text. Ids use mono and the instance id is the only mono identifier in prose. Units are set as "12.9 °C". | ◐ raw snake_case ids leak (`cooling_tower`, `streamsim/cooling-tower`) | |
| G-4 | Every list, panel and graph has a designed empty state that says what goes here and how to add it. | ◐ empty Scenarios card and the blank Inspector | |
| G-5 | Every action that waits on the binary shows progress in place within 100 ms and its outcome when done. Failures say what happened, why (the binary's message), and what to do next. | ◐ | |
| G-6 | Wording is plain, consistent and specific. The same action has the same verb everywhere (Open, Edit, Run, Review). No jargon in headlines (no "unblinded", "native-jsonl", "sim 0.3.2", "digest"). | ◐ review jargon removed; some raw ids remain in the editor | |
| G-7 | Times show as `hh:mm:ss` of simulated time from the run start; wall-clock dates as "14 Oct 2026, 13:05". Numbers use thin grouping, fixed decimals per unit, and a minus sign (−). | ◐ review shows "00:30:00 of 00:00:00 – 00:30:00 UTC"; axis ticks mm:ss | B (run date) |
| G-8 | Keyboard: Tab order follows the layout (left → centre → right). Escape closes the innermost panel or dialog. `?` lists every key. Focus is always visible and never lands on the page body after an action. | ◐ | |
| G-9 | Light and dark are designed separately: contrast is at least 4.5:1 for text and 3:1 for controls and state colours. Selection and focus are distinct from state colours. | ◐ disabled buttons about 2:1; olive and dimmed grey wires faint | |
| G-10 | No horizontal scroll at 1280 px; nothing important is clipped at 1280. Tooltips carry only extra detail, never the only copy of a fact. | ◐ at 1280 the plant bar clips to "23…"; separator becomes a dot | |
| G-11 | Navigation keeps context: going back to a plant restores its selection and zoom; an unsaved editor asks before it is left. | ◐ | |
| G-12 | Performance feel: screens open in under 1 s for the factory, under 3 s for a 26-hall campus. Playback stays smooth at 60 frames per second on the factory. | ◐ | |
| G-13 | The theme control shows a visible label and its current state, not a bare icon. | ✗ bare monitor icon on every screen | |
| G-14 | The breadcrumb uses one separator and one type size on every screen and at every width. | ✗ review crumbs larger than plant; separator becomes a dot at 1280 | |
| G-15 | A visible Keyboard button opens a help dialog that dims the page and shows real keys, with no "(graph)" or "(playback)" inside key caps. It names the key for previous and next event and describes no Play that does not exist. | ✗ no scrim, focus ring on Close, describes missing Play | |
| G-16 | The zoom readout uses one font on every screen. | ✗ "36 %" small sans on plant, "100 %" mono in editor | |
| G-17 | Every shown value says where it came from where it matters (preset, plant override, run setting, observed, simulated truth), and defaults read as defaults, not model rules. | ✗ see product review §6 | |

## 2. Workspace (home)

| ID | Expectation | Now | B |
|---|---|---|---|
| W-1 | The home screen leads with **Plants**: one row or card per plant (title, short id, number of instances, number of runs, last run). Composition and domain are one plant. A plant without a composition shows "read-only". | ✅ one plant row with Open, Edit and Run; composition and domain are one object (edd8ecc) | |
| W-2 | Each plant offers its next actions directly: **Open**, **Edit**, **Run**. | ✅ Open, Edit and Run on every plant row (edd8ecc) | |
| W-3 | **Runs** are listed newest first. Each run shows: plant title; when it ran; duration; mode (Explore or Evaluate); number of faults injected; and the verdict once checked. The world id is secondary. | ✅ runs show plant, time, duration and mode, newest first (edd8ecc) | B (plant, start time, duration, fault count per run) |
| W-4 | Runs can be filtered by plant, and the list scales to hundreds of runs (paged or virtual) without slowing the screen. | ✅ filter by plant and paging | |
| W-5 | **New plant** is prominent in the Plants header. The form explains the base plant, validates the id as the binary would before it writes, and opens the editor. | ◐ New plant in Compositions card, flush to card border | B (id check) |
| W-6 | Scenarios appear only when there are any. Otherwise one line under the runs says what a scenario is, with no empty card. | ✅ no empty Scenarios card (edd8ecc) | |
| W-7 | File paths and folder names appear only as secondary detail (a tooltip or a "Files" disclosure). | ✅ no paths in rows; files are secondary detail (edd8ecc) | |
| W-8 | First run: an empty workspace explains the three steps (create a plant, run it, review it) and offers New plant. | ◐ | |
| W-9 | The page uses the width: two columns at ≥ 1440 px (Plants, Runs) that grow to fill 1920 px; one column at 1280. | ✅ two columns at wide widths, one at 1280 | |
| W-10 | Controls inside cards are inset like the rows around them. | ✅ controls inset like the rows around them (edd8ecc) | |
| W-11 | Run status chips carry meaning: the colour and mark differ by state and mode. | ✅ status chips differ by state and mode (edd8ecc) | |

## 3. Plant screen

| ID | Expectation | Now | B |
|---|---|---|---|
| P-1 | The header shows the plant title, a one-line summary (instances, areas, links), and its actions: Edit and Run. Nothing important clips at 1280. | ◐ the summary clips | |
| P-2 | At the opening zoom every node is legible: id, kind in words, and a category colour or icon. A user can tell a chiller from a pump without zooming. | ✅ id and kind in words at the opening zoom, with a category band (fa21952) | |
| P-3 | The opening view fits the plant to the space with a small even margin. There are no large empty bands above or below the area frames. | ✗ | |
| P-4 | Wires are readable: they are routed between areas without crossing node text, and coloured by port family. Hovering or selecting a node highlights only its wires and dims the rest. | ◐ selection highlights; the overall view is spaghetti | |
| P-5 | The outline groups instances by area and shows each instance's kind in words. It filters as you type. | ✅ kinds and a filter in the outline | |
| P-6 | The inspector with nothing selected shows the plant overview: areas with counts, kinds with counts, recent runs of this plant with Review links. | ✅ plant overview with recent runs, areas and kinds | B (runs per plant) |
| P-7 | The inspector for an instance leads with its kind in words and its area, then channels with units, links in and out (clickable to select the other end), faults and effectors in words. Raw ids sit in a details line. | ◐ kind and area lead; links are not clickable yet | |
| P-8 | Halls (large plants) open with one click or Enter, and the breadcrumb shows the hall. | ✅ | |
| P-9 | The outline scrolls in its own region with a scroll cue, and the selected row is scrolled into view. | ◐ the selected row scrolls into view | |
| P-10 | The Navigator minimap is fully visible at a fixed size at 1280 and 1440. | ✗ cut off at the window bottom | |
| P-11 | A legend explains wire colours (port families) and pin shapes on the plant and editor screens. | ✗ no legend | |
| P-12 | Hall nodes show that they open (a chevron or "Open"), the hall layout is regular, site nodes are not blank, and hall names match the outline case. | ✗ campus ragged; CAL-1 truncates "feeds all 3 h…"; lowercase h01 | |
| P-13 | The inspector has a close control, its kind line never breaks mid-token, and long lists (13 faults, 17 states) are grouped or collapsible. | ✅ close control, folding lists | |
| P-14 | Area frame titles are never crossed by wires, and wires do not cross node headers. | ✗ yellow wire through "Substation" at 1920 | |

## 4. Editor (build and edit a plant)

| ID | Expectation | Now | B |
|---|---|---|---|
| E-1 | The editor looks like the plant screen it edits (same graph, outline, panel) plus an edit bar: the save state, Save, Revert, and Open plant. | ✅ Open plant whenever a composed domain exists | |
| E-2 | The save state is always visible and unambiguous: No changes, Unsaved changes, Saving…, Saved and composed, Saved but not composed (with the issue count), Not saved (with the reason and a recovery). | ✅ | |
| E-3 | Adding equipment: search the library by name or tag; see each component's title, short description and ports before adding; preset choices name their size. The name field suggests the next free id ("P-5"). Added equipment appears selected and in view. | ◐ plain selects, no search, no description, no name suggestion | |
| E-4 | Linking: dragging from a pin shows which pins accept it (compatible family highlighted, incompatible dimmed). The keyboard form lists only compatible inputs. A new link animates into place. | ✗ any pin accepts; the form lists every input | |
| E-5 | Removing a link from the graph: select a wire, then Delete or the × on it. The links list highlights the wire on hover. | ✗ | |
| E-6 | Issues from the binary: each flagged element is visible at any zoom, the issue list is ordered by element, and clicking an issue selects the element and scrolls it into view. Messages that name a pattern are rephrased by the binary in words. | ✗ no issue list; issue only in inspector, raw regex, double outline | B (message wording) |
| E-7 | Undo and redo (Ctrl+Z, Ctrl+Shift+Z) for every edit until save. | ✅ undo and redo with buttons and Ctrl/Cmd+Z (4446aef) | |
| E-8 | Moving an instance between areas shows the area frames as targets. Areas can be added and renamed. | ◐ select only; no area editing | |
| E-9 | Leaving with unsaved changes, by navigation or reload, asks first. | ◐ reload only | |
| E-10 | The empty editor shows a clear first step ("Add the first piece of equipment") next to the add form. | ✅ the step leads the Build panel, above the add form; the canvas empty state keeps the state and where equipment appears (a13c028) | |
| E-11 | Save and Revert are visibly disabled when there are no changes. | ✅ Save is disabled when clean | |
| E-12 | Help copy reads as sentences, with no stray hint between fields. | ✅ the two named hints are gone (d41c032); the links empty state, the last fragment, is a sentence (969531d) | |
| E-13 | The links list shows port names in words, each row selects its other end, and its count style matches the other lists. | ✅ port names in words and a labelled button that selects the other end (ba90274); a visible Remove control on every row (3f1996d); the count reads like the other lists' counts (43c79a6) | |
| E-14 | Every wire visibly ends on its two pins and never passes behind a node body. | ✗ wire hidden behind CH-2; loops at M-1 and M-2 inputs | |
| E-15 | The editor opens the plant on the same opening rule as the plant screen: fitted when it fits above the 40 % floor, otherwise at the floor, with the fade where to pan. | ✅ one opening rule on plant, editor and review; the cut is cued and the Navigator shows the whole | |

## 5. Run panel (run a plant)

| ID | Expectation | Now | B |
|---|---|---|---|
| R-1 | Opening Run shows a compact plan: seed (with a dice button for a fresh one), duration with unit, mode (Explore or Evaluate explained in one line each), and faults. | ◐ no dice | |
| R-2 | A fault row reads as a sentence: "At minute [10], [CT-1 · cooling tower] has a [fan failure]". The fault list shows only faults of that instance, in words. | ◐ | |
| R-3 | Picking a fault instance can also be done by clicking a node in the graph while the panel is open. | ✗ | |
| R-4 | While running: the progress bar, the simulated time, the speed (simulated seconds per second), the faults already injected ticked off, and Cancel. The graph shows a running indicator. | ◐ no speed, no fault ticks | |
| R-5 | When done, a toast-like status says "Run w-4 finished" and the review opens. The new run appears at the top of Runs. | ◐ | |
| R-6 | Problems in the plan are shown when Start is pressed, next to the field they concern. | ◐ shown in a list, not next to the fields | |
| R-7 | First focus goes to Seed, labels have a gap before their input, and the instance select shows the full instance and kind. | ✗ minute input focused; label touches it; "CT-1 · cooling t" | |
| R-8 | Seed, Duration and minute inputs share one height and font. | ✗ minute input taller with smaller text | |
| R-9 | The progress panel names the plant and run in words and shows the planned duration; the bar is thicker than 3 px; the graph shows a running indicator; Edit plant is disabled while running. | ✗ 3 px bar; no names; no indicator | |
| R-10 | A problem banner has an icon, and Start does not move when it appears. | ✗ no icon; Start jumps down 51 px | |
| R-11 | The mode labels "Explore" and "Evaluate" are set apart from their one-line explanation. | ✗ label and explanation run together | |

## 6. Review (review a run)

| ID | Expectation | Now | B |
|---|---|---|---|
| V-1 | The header names the run in words: "Factory utilities · run w-1 · Explore · seed 7 · 30 min". The verdict badge (Verified, Not verified) explains itself on hover or click. Version and adapter details sit in a details disclosure. | ✅ header in words; details behind a disclosure | |
| V-2 | **Playback**: Play and Pause (Space), speed 1×, 10×, 60× and 600×, a scrubber with event and alarm marks, and step to the previous or next event. Playback starts at the beginning, or at the first event, not the end. | ✅ Play, Space, 1×/10×/60×/600×; starts once every channel has reported | |
| V-3 | The timeline shows injected faults, alarms and effector calls as distinct marks with a legend. Hovering a mark shows what it is; clicking seeks to it and selects its instance. | ◐ only a red label chip, 50 px off its tick; no legend | |
| V-4 | The graph shows live readouts at the playback time. A node in fault or alarm is marked by shape and colour, not colour alone. | ✅ | |
| V-5 | Inspect (instance selected): every channel with its value at the playback time, a small trend line across the whole run with the playback cursor, and its fault and alarm marks. Explore runs show the true state beside the observed state. | ◐ two values, timestamp on every row; no trend, kind or area | B (true state, R9-13) |
| V-6 | The report leads with what happened: the disturbances, the alarms, effector calls and the KPI summary in words. Integrity (file check, replay check, digests) is one line ("Verified: files match and replay reproduces") that opens to its detail. | ✅ report leads with what happened; integrity is one line | |
| V-7 | From a review: **Run again** (the same plan in the run panel), **Open plant**, **Back to runs**. | ✅ Open plant and Run again; Back is the breadcrumb | |
| V-8 | Implausible values are flagged, not shown as normal (for example negative air flow). | ✗ negative air flow shown as normal (M-4 −12.8 m³/h, AC-1 −4.60) | B |
| V-9 | The review graph opens fitted, with the top frame labels uncut and no clipping at 1280. | ✗ "Site" and "Production hall" clipped at 50 %; clipped right at 1280 | |
| V-10 | The scrubber has a tick or dot on the track for each mark, with its label centred on it. | ✗ label chip only, offset about 50 px | |
| V-11 | Panel values right-align to one edge. | ✗ "match", counts and "0" end at different x | |
| V-12 | Tables have real column headers and words, with units and plain row names. | ◐ visible column headers | |
| V-13 | An Evaluate run says whether the truth is revealed, and the report agrees with the run panel. | ✗ KPI TRUTH shown while the run panel says the truth stays sealed | |
| V-14 | Readout nodes show their full kind and have even heights. | ✗ "AC-1 compressor l…"; tall M-1 and M-2 with empty bottoms | |
| V-15 | Header and playback controls carry text or visible tooltips. | ✗ « ‹ › » and bell are icon-only; bell purpose unclear | |

## 7. Pairing

| ID | Expectation | Now | B |
|---|---|---|---|
| A-1 | The pairing screen explains in one sentence why a code is needed, takes focus, accepts the code with or without dashes, and pairs on Enter. | ✅ | |
| A-2 | When the bridge stops, every screen says so in place, with the command to restart, and recovers on its own when the bridge is back. | ◐ | |

## 8. First use and creation

| ID | Expectation | Now | B |
|---|---|---|---|
| N-1 | A ready-to-run example is the shortest first-use path from the workspace, beside copy and empty creation. | ✗ no example entry |  |
| N-2 | Creation asks for a human name first, suggests the technical id, and validates it through the binary with an actionable field message. | ✗ an id is asked first; no human name |  |
| N-3 | Creation shows what a copy keeps and omits before it is made, and offers a full copy and an empty plant as distinct choices. | ◐ the form says it copies site, areas, profiles and axes; there is no full-copy choice |  |
| N-4 | A guide walks the journey (create, add equipment, connect, check, run, review), offers a specific next action at each unfinished step, and can be dismissed and reopened. | ✗ |  |
| N-5 | Visible defaults let an example run unchanged; seed and specialist settings sit under Advanced, while mode and truth exposure stay clear. | ◐ defaults are visible; the seed is not under Advanced |  |
| N-6 | Each equipment choice explains its purpose, preset size and key parameters, so no component id has to be guessed. | ◐ presets are listed by title; there is no description |  |
| N-7 | Each error names the affected item, explains the problem and offers a recovery that works without finding a file. | ◐ issues sit on the instance, but the wording is the binary's |  |

## 9. Running and stopping

| ID | Expectation | Now | B |
|---|---|---|---|
| RUN-1 | Before Start, the panel says what Start does: "Simulate this plant for the selected duration and save a run for review". | ◐ |  |
| RUN-2 | The panel says whether a run goes to completion or supports live control, and shows only controls that work. | ◐ "To completion" or "Connected session" (R9: linked to a plant wire host and a gateway until Stop and save); live control is R10 to R11 |  |
| RUN-3 | The stop control stays visible and says whether it cancels, pauses or ends and saves the run. | ✅ "Cancel and discard"; the panel explains what cancel keeps (d9c2b8c) |  |
| RUN-4 | Cancel is acknowledged at once, and its latency is explained without implying an in-flight command stops instantly. | ✅ "Canceling at the next step…" acknowledges the request at once |  |
| RUN-5 | Complete, cancelled, interrupted, failed and checking are distinct states, each saying what evidence was kept. | ◐ |  |
| RUN-6 | Collapsing a panel or leaving the plant never cancels a run, and the run stays visible from elsewhere. | ✅ leaving the plant keeps the run; a top-bar indicator follows it (d9c2b8c) |  |
| RUN-7 | Review Play reads recorded events only; Run again creates a new execution from an explicit plan and plant version. | ✅ review Play reads recorded events; Run again proposes the same plan (6790ec5) |  |
| RUN-8 | Repeatability is explained accurately: the seed alone is not enough, and the model, simulator version, inputs and commands also matter. | ✗ |  |

## 10. Connections, inputs and outputs

| ID | Expectation | Now | B |
|---|---|---|---|
| C-1 | Ports show a plain name, direction and quantity, with raw ids secondary. | ◐ |  |
| C-2 | Selecting a wire states "A supplies X to B" and what the receiving model uses it for. | ✗ |  |
| C-3 | A first connection previews both endpoints and its effect before it is added. | ✗ |  |
| C-4 | Compatible targets are shown and a rejected link is explained by binary-owned rules, not by colour family alone. | ✗ | B |
| C-5 | Required, optional, unconnected and default-fed inputs are explained from component metadata, never invented by the UI. | ✗ | B |
| C-6 | A model connection reads as distinct from physical wiring and from communication with a consumer. | ✗ |  |

## 11. Panels and visual system

| ID | Expectation | Now | B |
|---|---|---|---|
| VIS-1 | One stable shell puts the plant or run title first, a short summary second, and actions in the same place on every screen. | ◐ |  |
| VIS-2 | Equipment shows its kind in words and an icon, and zoom changes the detail level instead of shrinking text to illegibility. | ◐ kind in words at the opening zoom; no icon |  |
| VIS-3 | The grid and area frames are quiet and normal nodes are flat, so selection and abnormal states carry the stronger emphasis. | ◐ |  |
| VIS-4 | Category, selection, keyboard focus, faults and alarms each look different, including when combined. | ◐ |  |
| VIS-5 | Workspace rows and inspector values align, actions stay in predictable places, and fewer chips and headings compete. | ◐ |  |
| VIS-6 | The campus overview shows readable hall summaries, and opening a hall reveals readable equipment. | ◐ |  |
| VIS-7 | Review shows aligned fault, alarm and effector lanes under one time cursor, with compact labelled trends. | ◐ |  |
| VIS-8 | The left and right panels collapse independently through visible buttons and keyboard paths, and reopening is discoverable. | ✗ |  |
| VIS-9 | Panel state is remembered, and collapsing keeps selection, zoom, unsaved work and the run, with focus moving predictably. | ✗ |  |
| VIS-10 | Side panels start at consistent widths, resize within bounds, and leave the remaining space to the centre. | ◐ consistent widths; not resizable |  |
| VIS-11 | Reference screens and a state sheet exist for both themes, covering empty, loading, error and selected states. | ✗ |  |

## Audit 2026-10-14 (design director, screenshots round 0)

The director's ranked top ten:

1. Rebuild the home as Plants and Runs: no paths, one plant object, rich run rows with verdict (W-1, W-2, W-3, W-6, W-7, W-10, W-11).
2. Add playback: Play and Space, speeds, start at the beginning, step to events, a scrubber with marks and a legend (V-2, V-3, V-10, V-15).
3. Rewrite the review header and report in words, leading with disturbances, alarms and KPIs; fold integrity into one line; add Run again, Open plant and Back (V-1, V-6, V-7, V-11, V-12, V-13, G-6).
4. Make the plant graph legible on open: fit, kind and icon on nodes, routed and legible wires, a legend, and the same fit in the editor and review (P-2, P-3, P-4, P-11, P-14, V-9, E-15).
5. Fix the left rail and Navigator: scrolling outline, area groups, kind in words, filter, a fixed-size minimap (P-5, P-9, P-10, G-10).
6. Rebuild the plant inspector: overview, kind and area first, clickable links, close, grouped faults (P-6, P-7, P-13, G-3).
7. Make editing state honest: Save and Revert disabled when clean, Open plant always present, a global issue list with markers, binary-worded messages (E-1, E-6, E-11, E-15).
8. Redesign Add equipment (search, description, suggested name), the Link form (compatible only) and the help copy (E-3, E-4, E-5, E-12, E-13, E-14).
9. Finish the run panel: dice, sentence rows, graph pick, progress with speed and ticks, field-level problems; fix focus, truncation and size defects (R-1 to R-11).
10. Unify the shell and bars: one top bar with breadcrumb and a labelled theme switch, consistent rail and panel widths, visible keyboard help, 1920 content that uses the width (G-1, G-2, G-13, G-14, G-15, G-16, W-9); repair G-9 for disabled controls.

## Order of work

1. RUN-6 (done).
2. Starter example and creation (N-1 to N-3, N-6).
3. Collapsible panels (VIS-8 to VIS-10).
4. Editing comfort (E-3 to E-5, E-12 to E-14, C-1 to C-3).
5. Run panel finish (R-items, RUN-1, RUN-5, RUN-8).
6. Value origins (G-17).
7. Polish.

Items marked B start with the smallest binary or bridge change that provides the fact. Each is listed in the round
record where it lands.
