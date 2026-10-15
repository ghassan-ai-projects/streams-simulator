# UI/UX evaluation, round 11

- Director: LUNA, director of design. Chrome (Playwright, channel "chrome") at http://127.0.0.1:8768, own review instance, paired with a single-use code; a persistent profile kept the pairing and the theme across runs.
- Date: 2026-10-15.
- Bar: `governance/CANVAS_UI_ACCEPTANCE.md` §3.2 and `docs/simulation-canvas/UI-EXPECTATIONS.md`.
- Previous evaluations: round 6 (overall 4), round 8 (overall 4), round 10 (overall 3, lowest Information design of graph and review).
- **Viewports actually covered (all three, measured with `window.innerWidth` × `window.innerHeight` after `setViewportSize`):**
  - 1440 × 900 — the depth pass (Home, six plant views, three editors, the run panel, reviews w-44/w-45/w-33/w-13/w-40/w-35/w-43, legend, Keyboard dialog, dark pass).
  - 1280 × 800 — Home, water pumping plant, campus plant, reviews w-33 and w-13 (measured `{"w":1280,"h":800,"sw":1280}`; no horizontal page scroll).
  - 1920 × 1080 — Home, campus plant, water plant, review w-33 (measured `{"w":1920,"h":1080,"sw":1920}`).
- **Themes covered:** System (light) and Dark, both applied through the "Theme:" button and screenshotted on Home, the plant + inspector, the review w-33, the legend, the Keyboard dialog and the run panel. The theme was restored to System at the end.
- Plants opened in the plant view: water-pumping-station (62 %), food-cold-chain (62 %), compressed-air-works (62 %), multi-hall-campus (40 %), power-distribution-standby (44 %), office-tower-hvac (45 %), hunt-special-names (62 %), dir-review-empty (100 %). Editors opened (read only, nothing saved): compressed-air-works, boiler-steam-loop, water-pumping-station (Build panel only).
- Runs opened in full: w-33 (campus, Evaluate, 2 faults), w-13 (water pumping, Evaluate), w-40 (boiler, Explore, 1 fault), w-43 (boiler, Explore), w-35 and w-22 state checked, w-44 and w-45 (this round's scratch runs). All other runs were checked as rows on Home and in RECENT RUNS.
- Scratch artefacts created in this round (recorded for cleanup, nothing deleted or edited):
  - **Run w-44** (Water pumping station, Explore, seed 0, recorded Incomplete) — created accidentally: a probe meant for the Duration field filled the Seed field and pressed Start; the browser session then closed and the run was recorded as abandoned. Its review was used to audit the "Did not finish" state.
  - **Run w-45** (Dir Review Empty, Explore, 30 min, seed 1, Verified) — started deliberately to reach the finished-run panel and the fresh-review state; finished in about 30 s at the "Watch" speed.
  - No plant was created, edited or saved. The two-run scratch budget was used exactly.
- Console: **zero errors or warnings** on every normally reached surface in both themes at all three widths. The only failures were the expected 404 (run.json) and 422 (report) fetches when reviewing the abandoned run w-44 (ISS-11D-09).

## Findings — side panels first

Ordered most severe first.

## ISS-11D-01 The integrity detail still leads with file names, the binary's command and folder words
- Severity: Medium
- Kind: content
- Where: Run review, Report tab, the Verified block expanded — w-33, w-45 and w-40, light and dark, 1440 × 900. Also the grey banner of an incomplete run: w-44 review, light.
- Observed: the expanded block reads "REPLAY CHECK" with the binary's command **"streamsim verify"** as the label beside it; "FILE CHECK … files match" with the note "The report compares the files in the run folder with the digests the run recorded."; the seven rows "trace, journal, sealed journal, journal file, sealed journal file, ledger, history"; "DELIVERIES: Records/Count — emitted, delivered, dropped". On the abandoned run the banner reads "Not checked: streamsim verify could not replay this run." over "Details". Round 10 found this (ISS-10D-08); nothing changed.
- Expected: plain words for each check ("Replaying the run gives the same output", "The run's files match its record"), with file names, digests and the command under For developers (G-6, N-7, W-7; V-6 keeps the one-line summary).
- Suggestion: keep the collapsed line "Verified: the files match and the replay reproduces." as the only plain-words layer; move the REPLAY/FILE CHECK tables, the "streamsim verify" label and the digest note into the For developers disclosure; reword the incomplete banner without the binary's name ("Not checked: this run could not be replayed for its verdict.").

## ISS-11D-02 The opening view of wide plants and their reviews is cut at the inspector edge at 1280 and 1440
- Severity: Medium
- Kind: layout (review half; the plant-view half is parked for the graph round)
- Where: Review w-33 at the opening zoom 40 %, light, 1440 × 900: the rightmost area frame ("Hall 1") is cut at the inspector edge — the frame's right edge sits at x 1302 while the canvas ends at x 1066 (DOM-measured), so the frame title's "5 nodes" count is not visible and pk-h1/pl-h1 are half-hidden. The same at 1280 × 800 (canvas ends at x 944, 358 px hidden). Review w-13 at its opening 41 %: the "Reservoir" frame title is cut and wraps mid-word ("Rese / rvoir" visible, screenshot 17); at 1280 the Reservoir/Intake frames lose 96 px. At 1920 × 1080 both reviews fit (w-33 frames end at 1373, canvas at 1460). Plant view of Multi-hall campus behaves the same (see the parked section).
- Observed: the 40 % opening floor (introduced for kind-line legibility) is above the width-set fit on the two smaller desktop widths, so the fit overflows right and the user must pan or zoom out to see the last area. Nothing in the canvas indicates hidden content to the right.
- Expected: the review graph opens fitted with all frame labels visible and nothing clipped (V-9, ISS-6D-35); at 1280 nothing important is clipped (G-10). A frame title never breaks mid-word.
- Suggestion: when the legibility floor exceeds the fit zoom, open at the floor **and** show an affordance (a right-edge fade or a "pan right for Hall 1" hint), or fit first and let the tier rule drop the kind lines. At minimum, give frame titles `text-wrap: balance/nowrap` so a cut never splits a word.

## ISS-11D-03 Inspector run rows leave a dangling "·" at the wrap point
- Severity: Low
- Kind: typography
- Where: Plant view inspector, RECENT RUNS — water pumping at 1440 × 900 light ("w-13 · Evaluate · 30 min · seed 1 · 2 faults ·" line 1, "15 Oct 2026, 02:35" line 2); the same rows at 1280 × 800 wrap the other way, leaving "· 15 Oct 2026, 02:35" with a leading "·" (screenshot 41). Runs rows on Home wrap the same way at 1280.
- Observed: the metadata row's separator dots are not bound to their neighbours, so the wrap leaves a stray separator at the end or the start of a line on the panel's most-read list.
- Expected: the row wraps at a separator, not leaving the separator alone (typography bar: numbers and dates kept whole; N-7 no stray fragments).
- Suggestion: wrap the metadata in spans separated by `·` with `white-space: nowrap` on "label value" pairs, or drop the final "·" before the date and let the date start the second line.

## ISS-11D-04 Legend marks still collide: three families share the diamond and three share the square
- Severity: Low
- Kind: colour
- Where: Legend popover on the plant view and on the review, light and dark, 1440 × 900 (screenshots 19, 21, 34).
- Observed: Pressure (violet), Temperature (magenta) and Weather (grey) all use the diamond; Heat (pink), Power (amber) and Signal (grey) all use the square. Within each mark group only hue separates the families. Round 10 recorded the diamond triple (ISS-10D-11); the square triple is now equally loaded.
- Expected: each family separable by hue **and** mark, so colour is never the only carrier (P-4; the ISS-8D-08 fix gave Signal its square for exactly this reason).
- Suggestion: give two of the three families in each group a distinct mark (for example Temperature a hexagon, Heat a hollow circle), keeping the total set simple; verify the marks at legend size in both themes.

## ISS-11D-05 The plant overview's AREAS order contradicts the outline's order on the same screen
- Severity: Low
- Kind: coherence
- Where: Plant view of Multi-hall campus, 1440/1920, light: the Outline rail reads Site, Hall 1, Hall 2, Hall 3, Utilities (plant order); the inspector's AREAS block reads Utilities 9, Hall 1 5, Hall 2 5, Hall 3 5, Site 2 (count-descending, then alphabetical). Power distribution shows the same divergence (outline: Site, Line A, Line B, Substation, Generator yard; AREAS: Line A, Line B, Generator yard, Site, Substation).
- Observed: the same five areas are listed in two different orders in the two side panels of one screen, so the eye cannot confirm they are the same list.
- Expected: the same thing looks the same everywhere (consistency item of §3.2; VIS-5).
- Suggestion: order AREAS in the plant overview by plant order, like the outline; keep counts right-aligned.

## ISS-11D-06 The editor outline still interleaves empty areas between populated ones
- Severity: Low
- Kind: layout
- Where: Editor of Boiler steam loop, Outline rail, light, 1440 × 900: SITE 2, BOILER HOUSE 6, **GAS TRAIN 0 · EMPTY**, DISTRIBUTION 3, PROCESS HALL 3, **HEAT USERS 0 · EMPTY**. Compressed-air works happens to sort its empties last (PRODUCTION HALL A/B first, then four "0 · EMPTY" lines).
- Observed: round 10's two-part suggestion was applied only in half: each empty area is now a single heading line ("GAS TRAIN 0 · EMPTY" — the two-line "No equipment yet" is gone), but populated-first ordering is not applied, so an empty group still sits between two populated ones on the boiler editor.
- Expected: areas with equipment first, plant order kept within each group (ISS-10D-05's own suggestion; G-4).
- Suggestion: sort the outline groups by equipment count (populated first, stable within groups) or move all "0 · EMPTY" groups below the populated ones.

## ISS-11D-07 The plant bar does not singularise: "1 instances · 0 links"
- Severity: Low
- Kind: content
- Where: Plant view of Dir Review Empty, graph bar, light and dark, 1440 × 900 (screenshot 13).
- Observed: "1 instances · 0 links". The plural is wrong on the one plant built to exercise this state.
- Expected: plain, correct copy (copy item of §3.2; W-1's row wording handles counts).
- Suggestion: pluralise the nouns ("1 instance · 0 links"), the same rule "Show N more runs / 1 more run" already applies on Home.

## ISS-11D-08 The instance inspector states the kind twice in its first two lines
- Severity: Low
- Kind: content
- Where: Plant view inspector with an instance selected (CP-1 on water pumping, light and dark, 1440 × 900, screenshot 03): line 1 "Centrifugal pump · Pump house", line 2 "Kind: pump · streamsim/pump".
- Observed: the kind appears twice in words ("Centrifugal pump", then "Kind: pump") before any channel; only the raw vocabulary id is new information on line 2.
- Expected: the kind in words leads and raw ids sit in one secondary detail line (P-7); nothing repeats in two lines.
- Suggestion: reduce line 2 to the raw id in mono ("streamsim/pump"), or fold it into the "What this component models" disclosure.

## ISS-11D-09 Reviewing an incomplete run issues failed requests (404, 422) in the console
- Severity: Low
- Kind: function
- Where: Review of the abandoned run w-44, light, 1440 × 900. Console: "Failed to load resource: 404" (run.json) and "422" (report) against the page URL.
- Observed: the page state itself is well designed — "Did not finish" chip with an icon, "run.json is absent: the run is still open or did not end", a collapsed "Incomplete" block, TRUE RESULTS "Not available yet…", SCORING "Not scored…" — but the failed fetches contradict the no-failed-request checklist item for a reachable, normal state.
- Expected: no failed network request in normal use (§3.1 function); the absence of the file should be a handled response, not an error (A5's crashed-run case).
- Suggestion: catch the 404/422 in the bridge's run/report loaders and resolve them into the existing "Did not finish" state without logging failed resources (or fetch with a HEAD/existence check first).

## ISS-11D-10 The truncated breadcrumb plant name has no full-name tooltip
- Severity: Low
- Kind: craft
- Where: Top bar breadcrumb on the hunt plant (Hunt & "Ålborg" pumps / 冷却 test…), all three widths, light and dark. DOM: the crumb element's `title` is null.
- Observed: the long name is cut with an ellipsis in the breadcrumb, and the only place the full name appears is the inspector heading. ISS-6D-01 gave Home rows their hover titles; the breadcrumb crumb was left out.
- Expected: a truncated name shows its full text on hover (craft bar; G-1's breadcrumb).
- Suggestion: set `title` (and `aria-label`) to the full plant or run name on the breadcrumb's truncated crumb.

## ISS-11D-11 The outline rail clips its last row mid-glyph with no scroll cue at 800-high windows
- Severity: Low
- Kind: layout
- Where: Plant view of Multi-hall campus at 1280 × 800, light (screenshot 42): the "PRODUCTION HALL B / packaging-line-b" row is cut through the text at the NAVIGATOR heading's top edge; no fade or partial-row cue marks the scroll point.
- Observed: the outline scrolls under the Navigator, but the boundary is a hard cut of half a row, which reads as a rendering fault rather than a scroll affordance (P-9's scroll cue).
- Expected: a scroll cue (fade or mask) at the outline's bottom edge, as the run footer already has (ISS-8D-20's shadow).
- Suggestion: add a bottom fade/mask on the outline's scroll region above the Navigator frame, in both themes.

## Parked for the graph round

Recorded and scored honestly; the owner runs a dedicated graph round next. Not counted in the side-panel action list.

- **P-11-1 Campus/power opening fit remainder of ISS-10D-02/ISS-10D-04.** The kind-line half is **fixed**: the campus opens at 40 % (was 30 %) and every node shows its full kind ("packaging line", "production line", "cooling tower" complete); power distribution opens at 44 %, fully fitted, all kinds complete. The fit half is the live remainder: at 1440 the campus plant spans to x 1302 against a canvas that ends at 1066 (Hall 1 cut), at 1280 to 944; at 1920 it fits (1444 ≤ 1460). Same root cause as ISS-11D-02 — decide the fit-vs-floor rule once for plant, editor and review.
- **P-11-2 Full-tier node header is stretched.** Plant view of Dir Review Empty at 100 % (light, screenshots 13, 14) and the w-45 review at 100 %: the CP-1 node is ~250 px tall with ~50 px of empty space between the id and the kind row and more between the kind row and the readout rows. Same class as the won't-fixed ISS-8D-17 (fixed 132 px readout + header). A header that hugs its two rows would remove the "unfinished" look at 100 %.
- **P-11-3 Known wire-routing exclusions stand.** The campus's dashed site wire runs through the Utilities frame's row gap at 1920 (close above ch-u1, no crossing of text); wires under node bodies remain possible (ISS-6D-20, won't-fix, needs the routing pass). No area-title crossing found this round: on hunt-special-names at 62 % the dashed wire now passes **below** the "Production hall A" title (clip screenshot 24) — ISS-10D-06 is fixed.
- **P-11-4 Size preset units (ISS-6D-23, needs-binary) still open.** The Build panel's Size select reads "Medium (about 165 m3/h)" while channels read "m³/h". Confirmed 2026-10-15 on the water-pumping editor.

## Regression pass on rounds 6–10 (the items in the brief)

"Holds" = the fix was seen on screen or in the DOM this round.

- **ISS-10D-01 kind line overprint — fixed, holds.** DOM intersection test of every kind row against every pin label on six plants at their opening zooms (water 62 %, food 62 %, air 62 %, campus 40 %, power 44 %, office 45 %): **0 overlaps on all six**. Kind rows sit between the header and the first pin row on every node checked.
- **ISS-10D-02 + ISS-10D-04 overview-tier kind lines — fixed (with a parked remainder).** Campus at 40 %: every kind line complete, none cut, none missing (cal-site, wx-site included). Power at 44 %: complete and fully fitted. The remainder is the fit overflow at 1280/1440 (parked, P-11-1 / ISS-11D-02).
- **ISS-10D-03 review readout labels touching the kind line — fixed, holds.** At 58 % on w-33 and w-13: 0 overlaps (DOM). Kind rows ("pump", "control valve", "motor drive", "feeder meter") sit clear of "suction level", "flow out" and friends.
- **ISS-10D-05 editor outline empty areas — partly fixed.** Empty areas take one line ("PRODUCTION HALL 0 · EMPTY") — holds. Populated-first ordering not applied — ISS-11D-06.
- **ISS-10D-06 wire through "Production hall A" title — fixed, holds** (clip screenshot 24; the wire passes below the title band).
- **ISS-10D-07 integrity opens by default on a fresh review — fixed, holds.** w-45 (fresh, opened from the finished panel): `open: false`. w-33, w-35, w-43: `open: false`. One collapsed line everywhere.
- **ISS-10D-08 file names and "streamsim verify" in the integrity detail — not holding** (ISS-11D-01).
- **ISS-10D-09 legend heading overflows the popover — fixed, holds.** Popover padding 12 px; the heading wraps to two lines inside the box; heading right edge at x 1038 against the popover edge 1051 (13 px clearance), light and dark.
- **ISS-10D-10 keyboard dialog lists no previous/next event keys — fixed, holds.** RUN REVIEW lists Space ("play or pause") and `,` `.` ("go to the previous event, the next event"), plus the slider keys; no "(graph)"/"(playback)" in key caps.
- **ISS-10D-11 Pressure/Temperature/Weather share the diamond — not holding** (ISS-11D-04; Heat/Power/Signal now share the square too).
- **ISS-10D-12 a finished run names itself twice — fixed, holds.** The finished state shows one toast-like pill "Run w-45 finished · Open review" (dismissible), and the panel shows the plan summary ("30 min of simulated time · Explore · no faults.") with "Open the review". No duplicated sentence.
- **ISS-6D-35 review clipped at the inspector edge — not holding** (ISS-11D-02), now at the 40 % floor instead of the old 50 % floor; fits at 1920.
- **ISS-8D-17 tall readout node — still visible** (parked P-11-2; the won't-fix stands).
- **ISS-8D-19 node title inset — holds** (visually consistent 6 px inset under the category rule on every node checked at 62 %).
- **ISS-8D-20 run footer scroll cue — holds** (the shadow above the sticky footer marks the cut "EFFECTOR CALLS" heading).
- **ISS-6D-23 Size preset units — still open** (needs-binary; P-11-4).
- **ISS-8D-05 campus editor "Halls" group — not re-checked** (the campus editor was not opened this round).
- Round-10 items confirmed still holding in passing: loading wording ("Reading the run" family), the review's placeholder time, "Show 20 more runs", the area-click focus rule, Verified chip hover titles, alarm rows in plain words, one observed time in Inspect, KPI column alignment (Mean/Range share right edges at 1295/1422 CSS on w-33), link rows with labelled Remove buttons.

## Score table (round 11; 1440 × 900 depth, 1280 × 800 and 1920 × 1080 spot checks; light and dark)

| # | Dimension | Score | Evidence |
|---|---|---|---|
| 1 | Layout | 7 | Good: the shell is stable on every surface; no horizontal page scroll at any width (scrollWidth 1280/1440/1920); the plant bar fits at 1280 with no separator reflow (G-10 holds); panels toggle via buttons and `[`/`]`; opening views centre and fit every mid-size plant (water fits 52 % at 1280, 93 % at 1920). Against: the opening view of the two widest surfaces (campus plant, reviews w-33/w-13) is cut at the inspector edge at 1280/1440 with no affordance (ISS-11D-02); the outline rail clips its last row at 800-height windows with no cue (ISS-11D-11). |
| 2 | Typography | 8 | Good: the round-10 overprint class is gone — DOM-measured 0 kind/pin overlaps on six plants at opening zooms and 0 kind/readout overlaps on both reviews at 58 %; ids mono, dates whole, numbers grouped (19 089), units lighter and on the baseline. Against: dangling "·" at the run-row wrap in RECENT RUNS and on Home at 1280 (ISS-11D-03); "Rese / rvoir" mid-word frame-title wrap (ISS-11D-02); "1 instances" (ISS-11D-07). |
| 3 | Colour and contrast | 8 | Good: dark secondary text measures 7.8:1 and primary 14.3:1 on the panel; a faulted node is marked by dashed border + icon + text, never colour alone (w-40 "Fault · scaling"); mode/verdict chips pair icons with hue; wire families have distinct hues. Against: the legend's mark set does not separate three diamonds (Pressure/Temperature/Weather) or three squares (Heat/Power/Signal) (ISS-11D-04). |
| 4 | Components | 8 | Good: one field pattern everywhere (Seed/Duration both 28 px, 13 px, one font); the Seed dice button exists; error slots are reserved so Mode and Start do not move; the Keyboard dialog is per screen with the step keys and a scrim; every button has an accessible name (DOM check: none unnamed); the legend popover is padded with a wrapping heading. Against: the incomplete-run review logs 404/422 fetch failures (ISS-11D-09). |
| 5 | Content design | 7 | Good: the report leads with disturbances, effector calls and true results in words; "Explore run: the true results below were hidden… and shown when it ended" states the truth boundary plainly (V-13); the run panel explains what Start does and what repeatability needs ("The same seed, plant, faults and simulator version…"); "Not runnable yet" and "None: the plant runs healthy." are designed states. Against: the integrity detail still speaks in files, digests and "streamsim verify" (ISS-11D-01); the inspector says the kind twice (ISS-11D-08); "1 instances" (ISS-11D-07). |
| 6 | Feedback and flow | 8 | Good: a finished run is announced once with a direct "Open review" action (ISS-10D-12 fixed); the integrity line is collapsed on every review, fresh or old (ISS-10D-07 fixed); clicking a timeline chip seeks, selects the instance, scrolls the outline to it and switches to Inspect (w-40 boiler-1); reviews start at the first event; the abandoned run renders a complete, explained "Did not finish" state. Against: failed console requests on that state (ISS-11D-09); the truncated breadcrumb name has no tooltip (ISS-11D-10). |
| 7 | Information design of graph and review | 6 | Lowest. Good: kind lines are complete at every opening zoom (the round-10 cap is lifted); review readouts align labels left and values right with "+N more in Inspect" disclosure; timeline chips centre under their marks with a legend and a first-fault caption; the review of w-13/w-33 at 58 % and w-40 at 41 % is dense but calm. Against: the opening fit still cuts the rightmost area on the widest plants at 1280/1440 (ISS-11D-02, P-11-1); the full-tier node header is stretched with empty bands (P-11-2); wires under bodies remain the known exclusion. |
| 8 | Motion and responsiveness | 8 | Good: no layout shift on errors (reserved slots), the finished toast replaces the running pill in place, zoom steps respond at once, theme switches instantly, the run panel's elapsed clock ticks while the review placeholder time stays stable until the run is read; screens opened well inside the 1–3 s budgets on every visit. Nothing flashy, nothing jumpy. |
| 9 | Craft | 8 | Good: the legend heading wraps inside its 12 px padding (ISS-10D-09 fixed); the hunt plant's long title wraps cleanly to three lines in the inspector; the run footer's scroll shadow marks the cut; focus rings visible on keyboard paths; favicon and per-screen tab titles in place. Against: the mid-word frame-title wrap and the clipped outline row (ISS-11D-02, ISS-11D-11); no tooltip on the truncated breadcrumb crumb (ISS-11D-10). |
| 10 | Whole-product coherence | 8 | Good: Open / Edit / Run / Review keep one verb set across Home, plant, editor and review; one top bar with one breadcrumb pattern at all three widths; dark and light both designed (measured contrast); status chips, error patterns and field styles are shared components. Against: AREAS order contradicts the outline order on the same screen (ISS-11D-05); the integrity block's file words contradict the product's plain-word register (ISS-11D-01). |

**Overall: 6** (the lowest dimension is 7, Information design of graph and review, at 6; not rounded up). Dimensions 1 and 5 are 7; dimensions 2, 3, 4, 6, 8, 9 and 10 are 8.

Finding counts: Medium 2 (ISS-11D-01, ISS-11D-02); Low 9 (ISS-11D-03 … ISS-11D-11); total 11, plus 4 parked graph notes.

## Compared with round 10

Round 10 overall 3 (lowest: Information design of graph and review, 3). Round 11 overall 6 (same dimension, now 6). Per dimension: Layout 5 → 7; Typography 4 → 8; Colour 6 → 8; Components 7 → 8; Content 5 → 7; Feedback 5 → 8; Information design 3 → 6; Motion 7 → 8; Craft 5 → 8; Coherence 5 → 8.

What improved and was verified: the entire label-overprint class (ISS-10D-01/02/03/04) is fixed and DOM-verified at three widths; the review defaults are consistent (integrity collapsed, honest placeholder time); the finished-run flow states the run once; the legend and keyboard dialog defects are closed; the run panel gained the dice, the unit choice, the speed choice and the "what Start does" footer; the outline's empty areas take one line; the hunt title is clear of wires.

What still holds the score down: the opening fit of the widest plants and reviews at 1280/1440 (the 40 % floor's cost, parked for the graph round but scored); the integrity detail's file-level copy; and a tail of single-line defects (dangling separator, singular counts, mark collisions, panel order divergence) that are each cheap but together keep five dimensions from 9–10.

## Concrete actions, ranked by impact

### Side-panel actions

1. **Rewrite the integrity detail in plain words** (ISS-11D-01, review Report): keep the one-line "Verified: the files match and the replay reproduces."; move the REPLAY/FILE CHECK tables, the "streamsim verify" label and the digest note under For developers; reword the incomplete-run banner without the binary's name. Raises Content design and Coherence.
2. **Make the opening fit honest on wide plants and reviews** (ISS-11D-02 with P-11-1, plant + review): when the legibility floor exceeds the fit zoom, show a right-edge affordance (fade or "pan right" hint) or fit first and let the tier rule drop kind lines; never let a frame title break mid-word. Raises Information design (the current lowest) and Layout.
3. **Separate the legend's marks** (ISS-11D-04, legend popover): one distinct mark per family group — give Temperature (or Pressure) a third shape and split Heat/Power/Signal's square pair. Raises Colour and contrast.
4. **Fix the run-row wrap and counts** (ISS-11D-03 + ISS-11D-07, inspector RECENT RUNS + Home + graph bar): bind separators so wraps never strand a "·", and pluralise "1 instance". Raises Typography and Content design.
5. **One area order and one kind line in the inspector** (ISS-11D-05 + ISS-11D-08, plant overview + instance header): order AREAS by plant order like the outline; reduce the instance header's second line to the raw id. Raises Coherence and Content design.
6. **Sort the editor outline populated-first** (ISS-11D-06, editor Outline rail): move "0 · EMPTY" groups below populated ones. Raises Layout.
7. **Handle the incomplete-run fetches quietly and name the truncated crumb** (ISS-11D-09 + ISS-11D-10 + ISS-11D-11): catch the 404/422 into the existing Did-not-finish state, add the crumb's hover title, and fade the outline's bottom edge. Raises Feedback and flow and Craft.

### Graph-round parked actions

1. **Decide the fit-vs-floor rule once** for plant, editor and review (P-11-1 / ISS-11D-02): fit-first with a legibility tier, or floor with a pan affordance — one rule, tested at 1280/1440/1920 on the campus, power distribution, w-33 and w-13.
2. **Tighten the full-tier node header** (P-11-2): at 100 %, hug the id and kind rows so a one-channel node is not half empty.
3. **Wire routing pass** (known exclusions): wires under node bodies and through frame gaps (ISS-6D-20 stands); re-verify every area title at every zoom tier after it.
4. **Size preset units** (P-11-4, needs-binary): the component library should write "m³/h" so the Build panel's Size select matches the channels.
