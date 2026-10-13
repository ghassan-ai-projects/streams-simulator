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
