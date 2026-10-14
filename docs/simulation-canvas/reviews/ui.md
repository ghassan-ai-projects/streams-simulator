# Canvas UI — review record (real app on real files)

Benchmark: the frozen mockups (`docs/simulation-canvas/mockups/`, BRIEF C1–C13) and `governance/CANVAS_UI_ACCEPTANCE.md` §3.
Reviewers: design director, accessibility and keyboard, process and test engineer. Order: function first, then detail.

## Round 1 (P9: pairing, home, plant view, review)

Scores: director 2/5, accessibility 2/5, engineer 3/5.

| Area | Findings | Plan |
|---|---|---|
| Graph readability | Director F1, F2, F10, F12, F14; a11y U4; engineer P9-10 | Done in this round: fit to the plant on load, semantic zoom (full ≥ 0.85, compact ≥ 0.45, overview below; text keeps its on-screen size), zoom bar in the header, Navigator docked in the side panel, category stripes, invisible anchors, selection pans into view |
| Graph keyboard | A11y U1–U3, U5–U8, U11, U14 | Batch 1: the graph is one tab stop with arrow-key travel, Enter selects or opens a hall, Esc backs out, a help overlay, focus and title on route change, a focusable inspector with a live announcement |
| Review honesty | Engineer P9-1–P9-4, P9-6–P9-8; director F3–F5, F9; a11y U9, U10 | Batch 2: ended / complete / truth-unavailable states; badge only from verify; an inspector with every channel at the scrub time and decoded annunciator codes; report columns that never wrap; injected and onset times with the day; KPI units and entity filter; staged progress; fault and alarm states on the scrubbed graph; event navigation; the review draws the domain the run used (`run.json` `domain_spec`) |
| Home, campus, pairing | Engineer P9-9, P9-11; director F6–F8, F13, F15–F18; a11y U12, U13 | Batch 3 |
| Simulator | R9-03 raised to Medium by the engineer | Compose writes `from_port` and `to_port` on every link (in progress) |

### Round 1 dispositions (commits 6d6e26d, c111b98, bad3f51, 711044d, 726abd4)

| Findings | Disposition |
|---|---|
| Director F1, F2, F10, F11, F12, F14; a11y U4; engineer P9-10 | Fixed: fit on load, semantic zoom, docked Navigator, zoom bar in the header, selection pans into view, category stripes, a one-line header |
| A11y U1–U3, U5–U8, U11, U12, U14 | Fixed: the graph is one tab stop with arrows, Enter, Escape and announcements; edges leave the tab order; ? lists the keys; screens focus their heading and set the title; the inspector is focusable and announces; outline areas collapse |
| Engineer P9-1, P9-3; director F3 | Fixed: badge from the report and verify only (Verifying, Not checked, Verified, Not verified, Incomplete); unclosed runs say so, counts n/a; unavailable truth named; scoring only as reported; elapsed seconds while the binary works. Accepted for the rest of F3: report and verify take 25–36 s on an 18 h run because the binary replays it; a faster check belongs in the simulator |
| Engineer P9-2, P9-6; director F5, F9 | Fixed: Inspect tab with every channel at the scrub time; annunciator states in words; units apart; the review draws the domain in run.json; alarm, fault and warning states on the scrubbed graph; event chips, previous and next event, 10-minute steps. Accepted for now: a play/pause transport (scrub and step cover the review task) |
| Engineer P9-4, P9-7, P9-8; director F4; a11y U9, U10 | Fixed: columns never wrap, ids read as words with the raw id in the title, onset and set times with the day, KPI units and the selection filter, banners for files that differ and incomplete playback, Shift+arrow and 24 px targets |
| Engineer P9-5, R9-03; director F7 | Fixed in the simulator: compose writes from_port and to_port on every link; links without them draw dashed "port not recorded" pins |
| Director F6, F13 | Fixed: square hall grid, a shared instance that feeds every hall says so instead of 52 wires, view figures in the header, hall prefix dropped inside a hall |
| Engineer P9-9, P9-11; director F8, F15–F18; a11y U13 | Fixed: home lists every run with ended or not closed and the session badge, unclosed runs open in review, names no longer clip, pairing error under focus and one announcement, 401 returns to pairing, port legend, neutral logo, kinds as words. Accepted for now: per-run span and counts on home (each needs the 25 s report) |

## Round 2 (after round 1 fixes)

Scores: director 3/5 (five High: overlapping nodes at the compact zoom, cropped and wrapping report columns, an unreadable review at fit, raw ids and outcomes, campus link lists), accessibility 3/5 (N1 zoom-scaled focus ring, N2 review graph at zero height under a banner), engineer 3/5 (R2-1 zoom tiers, R2-2 a real divergence shown as Not checked because the bridge drops `error.result`, R2-3 the edited plant breaks every earlier verify — R9-06, R2-4 report wording).

### Round 2 dispositions (commit 6b4ff9a)

| Findings | Disposition |
|---|---|
| Director H1, M1; engineer R2-1 | Fixed: geometric zoom tiers (full, compact, overview, map) so counter-scaled headers never outgrow their node; edges go to anchors in overview and map; ids hide on the map |
| A11y N1 | Fixed: focus and selection rings divide by the zoom and stay 2 px on screen |
| A11y N2; director H3 | Fixed: the review grid has a row for the banners, the graph keeps at least 360 px and opens at no less than 50 % zoom; an end-to-end test asserts the height under a banner |
| A11y N3, N4 | Fixed: Esc on the key list returns focus where it was; `+`, `-`, `0` zoom while the graph has focus; the list names them and Shift+Arrow |
| A11y N5–N8 | Fixed: one live message per selection, the hall announcement no longer repeats, Report and Inspect are toggle buttons with `aria-pressed`, a 24 px playback range, the pairing title |
| Director H2, M2; engineer R2-4 | Fixed: a fixed table layout with column widths; Disturbances in three columns with the set time on hover; effector outcomes and fault names in words with the raw id in the title; a Replay check section with the verify verdict and a File check section |
| Director H4, M3 | Fixed: banners say what happened in plain words with the binary's message under Details |
| Director H5, M4–M6, L1–L3 | Fixed: link lists over 12 group by the instance's own port with link and instance counts; the event lane keeps two rows and drops the overflow (events stay reachable with previous and next); home paths truncate from the start with the full path in the title |
| Engineer R2-2, R2-3 (R9-06) | In progress in the simulator: verify replays the domain the run recorded and reports a changed domain as information; the bridge keeps a CLI failure's result so a real divergence reads Not verified |

## Round 3 (after round 2 fixes; reviewed at b412223)

Scores: director 3/5 (five High: D3-1 report cells cut, D3-2 readout values and units cut, D3-3 node text below its size at the full tier, D3-4 review opens with the top row cut, D3-5 no hall crumb or hall KPI in review), accessibility 3/5 (A3-1 keyboard ring still scaled by zoom, A3-2 keys dead inside a hall opened by Enter, A3-3 review status by colour only), engineer 3/5 (E3-1 Replay check showed the combined badge as verify's verdict, E3-2 a tampered recorded spec read Not checked).

### Round 3 dispositions (commits 1bae764, 883063a, 65a178a, 1c848d3, 24cc7a2, 30f48ca)

| Findings | Disposition |
|---|---|
| Engineer E3-1, E3-4, E3-9 | Fixed: the Replay check states verify's own verdict and what differs; the File check the report's; truth under a run that did not verify carries a caution |
| Engineer E3-2, E3-3, E3-6–E3-8, E3-10 (R9-07–R9-12) | Fixed in the simulator: a broken recorded digest is a verdict with a reason; verify checks the trace file; export and score use the recorded specs; docs; domain_changed omitted means unknown; port-specific session cookie |
| Engineer E3-5, E3-11 | Fixed: store, component and end-to-end tests for a divergent replay, a tampered recorded plant and an edited trace; a geometric check that no node overlaps or spills at any tier |
| A11y A3-1, A3-11 | Fixed: rings are 2.5 px divided by the zoom (device-pixel rounding keeps 2 px on screen), drawn only while the graph has focus; a click moves the keyboard position; an end-to-end test measures the ring at fit and on the map tier |
| A11y A3-2, A3-8; director D3-5 | Fixed: entering or leaving a hall keeps focus in the graph with a valid position; halls open in the review with a hall crumb; KPI truth for a hall lists its instances |
| A11y A3-3; director D3-7, D3-9 | Fixed: status tag with icon and word (and the fault's name) on nodes, in the name and the announcement; alarmed channels first; halls carry their worst member status with a count |
| A11y A3-4–A3-7, A3-10, A3-12, A3-13 | Fixed: one announcement per verdict, counters and zoom percentage out of live regions, settled zoom announced, step buttons stay focusable, legend a disclosure that Esc closes, reduced motion honoured, folder out of heading names |
| A11y A3-9; director D3-1 | Fixed: number, time and unit columns size to their content and never clip; only labels truncate with a title; the panel widens to 520 px at large widths |
| Director D3-2, D3-3, D3-6, D3-14, D3-18 | Fixed: one-column readouts with value and unit never clipped; full tier from 100 % so text never shrinks; compact tier shows the kind; right diamond pins clear their label; overview labels top-align |
| Director D3-4 | Fixed: the graph opens on the whole plant, or on its top-left corner at the minimum zoom; an event chip reveals its instance |
| Director D3-8 | Fixed: channels, states, effectors and faults read as words with display units everywhere |
| Director D3-10, D3-13 | Fixed: chips size to their text and stay in the lane with a count of the rest; the flat density strip is gone; transport errors are amber. Open: alarm activations on the lane need alarm events in the report |
| Director D3-11, D3-12 | Logged as R9-14 and R9-13: report and verify memory on the campus, playback indexing off the main thread; an Explore truth read for the true state |
| Director D3-15, D3-16, D3-17 | D3-16 and D3-17 fixed with E3-1 and A3-11. Accepted for now: wires over frame titles and frame gutters (layout round) |
