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
