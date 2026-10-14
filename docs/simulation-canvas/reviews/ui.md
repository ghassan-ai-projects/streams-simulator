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

## Round 4 (after round 3 fixes)

Scores: director 3/5 (High: D4-1 fault and alarm text past the node edge, D4-2 row headers in browser default style), accessibility 4/5 (High: A4-1 status word cut at the opening zoom), engineer 3/5 (High: E4-1 verify passed a missing evidence file or digest, E4-2 the canvas worked out fault state itself).

### Round 4 dispositions (commits 1a1db02, c2c9f3b, c204052, 5b862c5, 7e4ff4c, 47080d9, e3c708d and the design commit after them)

| Findings | Disposition |
|---|---|
| Engineer E4-1, E4-3, E4-8 (R9-15–R9-17) | Fixed in the simulator: every evidence file and digest of a closed run is required, checked against what was recorded and what the replay gives; verify names a `reason` and `file`; the canvas headline is phrased from the reason |
| Engineer E4-2, E4-4 (R9-20, R9-21) | Fixed in the simulator: the report gives each fault as an interval with how it ended; instruments declare their state vocabulary and compose copies it onto the channel; the canvas only presents both |
| Engineer E4-5–E4-7 | Fixed: the status word never truncates; an end-to-end run with two faults in one hall proves the rollup, the fault names and the way back; e2e specs have their own tsconfig |
| A11y A4-1–A4-10 | Fixed: tags at 11 px on screen with the kind giving way; fault name in the readout and the node name; chip and hall-entry announcements; headings hold only titles and every table has headers; set times as text; the key list closes when focus leaves; review breadcrumb nav; one application role |
| A11y A4-11 | Accepted: a run that does not verify is a 422 from the bridge, its contract for a failed command; the browser logs the status |
| Director D4-1, D4-2 | Fixed: word values in the label face with an ellipsis and a title, labels keep room; row headers styled as cells |
| Director D4-3–D4-7, D4-10, D4-11 | Fixed: double-click opens a hall; selected wires stand out; hall summaries; Inspect lists what is active now; scrubber time ticks and an empty-lane message; compact labels at 11 px; separators; legend for filled and hollow pins |
| Director D4-8, D4-9; alarm events on the lane | Logged as R9-22, R9-23, R9-24 for the simulator |
| Watch test flake | Fixed: the test replaces the file atomically; os.WriteFile showed two real changes |

## Round 5 (after round 4 fixes; reviewed on a frozen build at 6a653c8)

Scores: director 3/5 (High: D5-1 ellipses at the opening zoom, D5-2 dimmed wires under 3:1, D5-3 "no events" while the report loads), accessibility 3/5 (High: A5-1 dimmed wires under 3:1; regressions A5-3 outside selection not followed, A5-6 arrow travel off screen), engineer 3/5 (High: E5-1 `incomplete: true` in run.json bypasses every evidence check).

### Round 5 dispositions (commits 58729f9, 596dc11, 8726dd8)

| Findings | Disposition |
|---|---|
| A11y A5-1; director D5-2 | Fixed: dimmed wires are the tertiary text colour (5.3:1 light, 7:1 dark), measured by an end-to-end test |
| A11y A5-2; director D5-4 | Fixed: opening a hall clears the selection; nothing is emphasised for a node outside the view |
| A11y A5-3, A5-6 | Fixed: outline and chip selections move the keyboard position and come into view; the follow is instant and uses layout sizes; end-to-end tests |
| A11y A5-4, A5-5, A5-7–A5-9 | Fixed: tag text in the primary colour with a coloured icon; Space selects and Enter opens a hall; every active fault named, also under an alarm; Observed table headers; consistent hall counts; singular lane note |
| Director D5-1, D5-8 | Fixed: the compact tier wires to anchors, drops the pin band and stacks label over value with wrapping, so the opening zoom shows no ellipsis |
| Director D5-3, D5-5, D5-13 | Fixed: the lane says events appear once the report is read; short chips in three rows; red only while the fault is active |
| Director D5-6, D5-7, D5-9–D5-11 | Fixed: words for members, checks and pins; the header names plant, version, seed, adapter and simulator; overview ids fit; copy |
| Director D5-12 | Open: one summary figure per hall tile |
| Engineer E5-3, E5-6, E5-7 | Fixed: every active fault per instance; an alarm state by its declared name from a composed domain, end to end |
| Engineer E5-1, E5-2, E5-4 | In progress in the simulator (see the P9 handoff) |
| Engineer E5-5, E5-8 | Logged as R9-25 and R9-26 |

## Round 6 (frozen merged build, 2026-10-14)

Scores: design director 4/5, desktop keyboard/input 3/5, process/test engineer 2/5. **Not approved**.
The owner requests functionality first and visual optimization afterwards; the second reviewer now covers desktop
keyboard/input usability. Full findings, build digests, actual width/theme coverage and prior-disposition checks:
[P9-round-6.md](P9-round-6.md).

| Findings | Disposition |
|---|---|
| E6-1 | Fixed: rejected verify results warn above truth even when the overall badge is Incomplete, including hall KPI-only view. Component regression and real-binary Chrome test. Awaiting re-review. |
| E5-1, E5-2 | Chrome regressions prove an incomplete stamp never says Reproduces and a missing ledger digest fails file/replay checks. |
| E6-2–E6-5 | R9-14, R9-22, R9-24 and R9-13 remain functional P9 blockers. |
| U6-1 | Open: reproduce hall clipping at the reported widths and assert Fit geometrically in Chrome; reviewers disagree on different views. |
| U6-2 | Open: direct keyboard route to the graph. |
| D6-01, D5-12, frame-title wire crossings | Deferred visual detail after functionality per owner direction. |
| E6-boundary | Candidate: strict response validation at the adapters boundary needs verification and correction. |

## Round 7 (functional alarm-playback batch)

The owner requests functionality before visual optimization. This round reviews R9-24 and integrity propagation;
it does **not** approve all of P9 or repeat the campus review. A freshly frozen simulator 0.3.2 binary and UI
serve a synthetic scheduled-alarm run. Its reported activation is 140 seconds after start.

| Lens | Scoped score | Review record |
|---|---|---|
| Design director | 5/5 | [P9-alarm-design-review.md](P9-alarm-design-review.md) |
| Desktop input | 5/5 | [P9-alarm-input-review.md](P9-alarm-input-review.md) |
| Process/test engineer | 4/5 | [P9-alarm-engineering-review.md](P9-alarm-engineering-review.md) |

Dispositions: E7-1 fixed by propagating verify state into Inspect; R9-24 now supplies alarm instants, lane chips
and Jump to alarm; event reveal opens Inspect. Independent scan, owned-copy, initially standing alarm and
open-Evaluate truth tests pass, and Chrome proves the 02:20 chip/jump, source selection and 85 °C reading.
E7-3/E8-1 becomes R9-37, a precision requirement on the next bounded query contract.
All remaining P9 functional blockers stay open. The full phase gate passed;
[P9-alarm-evidence.md](P9-alarm-evidence.md) records the command, toolchain and exact pin differences.
