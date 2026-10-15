# UI/UX evaluation, round 12

- Director: LUNA, director of design. Chrome (Playwright, channel "chrome") at http://127.0.0.1:8768, own review instance; the persistent profile (`/tmp/canvas-review/pw-profile`) still held its pairing — no code was entered, no re-pair attempted. Ports 8766/8767 were never touched.
- Date: 2026-10-15.
- Bar: `governance/CANVAS_UI_ACCEPTANCE.md` §3.2 and `docs/simulation-canvas/UI-EXPECTATIONS.md`.
- Previous evaluations: round 6 (overall 4), round 8 (overall 4), round 10 (overall 3), round 11 (overall 6, lowest Information design of graph and review).
- **Viewport: one only, by owner direction.** 1440 × 900, set once and measured as proof: `{"w":1440,"h":900}` (`window.innerWidth` × `window.innerHeight` after `setViewportSize`). 1280 and 1920 were not tested.
- **Themes covered:** System (light) and Dark, both applied through the "Theme:" button and screenshotted on Home, the plant + inspector + legend, the review w-33 (pan cue), the review w-44, the campus outline fade and the boiler editor. The theme was restored to System at the end (final button label "Theme: System").
- Round 12 is a **regression pass** over the two fix batches committed after round 11, plus a fresh score. Surfaces opened: Home, the boiler-steam-loop editor (read only, nothing saved), plant views of water-pumping-station, multi-hall-campus, power-distribution-standby, dir-review-empty, hunt-special-names (and boiler-steam-loop for the pin marks), reviews w-33, w-13 and w-44. The Keyboard dialog was reopened on the plant view.
- Scratch artefacts created in this round: **none.** No run was started (budget: at most one; used zero). No plant created, edited or saved. Screenshots (21) live under `/tmp/canvas-review/round-12/`, scripts under `/tmp/canvas-review/round-12/scripts/` — recorded for cleanup, nothing in the repository touched.
- Console: **zero errors or warnings** on every normally reached surface in both themes. The only failures anywhere are the w-44 review's own fetches — 404 `run.json` and 422 `run.verify` — which this round's batch was supposed to remove (ISS-12D-01). One `ERR_ABORTED` on the w-44 event stream appears when navigating away from that review; an abort on navigation, not a failed request, and not counted.

## Regression pass over the two fix batches

"Holds" = the fix was seen on screen or in the DOM this round, at 1440 × 900, in both themes where the item names a theme.

### Batch 1

- **6611972 editor outline: empty areas below every populated area — holds.** Boiler-steam-loop editor outline, light and dark: Site 2, Boiler house 6, Distribution 3, Process hall 3, then **Gas train 0 · empty** and **Heat users 0 · empty** last. No empty group sits between populated ones.
- **939406a review pan cue + frame titles — holds (review half; the plant-view half is not implemented, parked as directed).** Review w-33 at its opening 40 %: `div[data-pan-cue="right"]` present, 32 px wide, full canvas height, exactly where the plant runs past the canvas (frames end at x 1302 against the canvas right edge 1066); it renders in dark too (screenshot 18). Review w-13 at 41 %: "Reservoir" renders whole — `white-space: nowrap`, no ellipsis, head and tail on one line; "Pump house" splits only at the word boundary (head "Pump", tail " house"). No cue appears on w-13, which fits at 1440 — no false positive. The plant screens pass no `panCue` prop (DOM and source agree), so the campus plant still cuts Hall 1 at the inspector edge with no affordance — the parked graph-round item, scored as-is.
- **1c9089c integrity block speaks in checks, details under For developers — holds.** Review w-33, integrity expanded: the collapsed line is "Verified: the files match and the replay reproduces."; REPLAY CHECK leads with the plain sentence **"Reproduces: replaying the run gave the same output."** and FILE CHECK with **"Files match the record."**; two collapsed "For developers" disclosures hold, respectively, "Checked with the streamsim verify command over the run folder." (plus the replay detail) and the digest note plus the seven-row file table (trace/journal/sealed journal/journal file/sealed journal file/ledger/history — all "match"). On the abandoned run w-44 the banner reads **"Not checked: this run could not be replayed for its verdict."** — the binary's name is gone.
- **3bfdbf9 legend marks — holds, both themes at legend size.** Legend popover on the plant view and on the review, light (screenshot 05) and dark (screenshot 16): Temperature and Signal are hexagons, Heat and Weather are bars, **Pressure is the only diamond, Power the only square**; the swatch classes (`_hexagon_`, `_bar_`, `_diamond_`, `_square_`) match on every row. On the node pins the marks carry through at 61 % on the boiler plant (computed styles): temp/signal pins take the hexagon clip-path, pressure the rotated diamond, weather the 6 × 10 bar, water/condenser circles (radius 50 %), power squares (radius 2 px); hollow (white fill) vs filled (family hue) still encodes connection, with "Filled: connected / Hollow: not connected" rows in the legend.

### Batch 2

- **920ab1d run metadata rows wrap at whole label-value pairs — holds.** Home and the plant inspector's RECENT RUNS now render each "label value ·" as one span (`MetaParts`): across 97 measured spans on Home, **zero lines start with a lone "·"**; in the campus inspector the w-33 row wraps as "w-33 · Evaluate · 30 min · seed 1 · 2 faults ·" / "15 Oct 2026, 04:13" — the separator stays glued to "faults", the date opens line 2 whole.
- **1a1fbe1 inspector AREAS: plant order + right-aligned counts — partly holding.** The counts right-align on campus and power (every count's left edge at x 1420, names ragged-left), and the list no longer sorts count-descending-then-alphabetical: it now follows the domain's declared area order (campus: Site 2, Utilities 9, Hall 1 5, Hall 2 5, Hall 3 5 — exactly the `areas` array's order). **But the acceptance "same order the outline rail shows" fails on both checked plants**: the outline rail sorts populated-first (Site, Hall 1, Hall 2, Hall 3, Utilities on campus; Site, Line A, Line B, Substation, Generator yard on power) while AREAS keeps declared order (Utilities second on campus; Substation and Generator yard before the Lines on power). The same five names, two orders, one screen — ISS-12D-02.
- **b3abb86 singular counts — holds.** Dir-review-empty: plant bar **"1 instance · 0 links"**, overview sentence **"1 instance in 1 area · 0 links"**. Water reads "11 instances · 11 links", campus "26 instances in 5 areas · 20 links" — plurals intact.
- **25ae10c instance inspector states the kind once — holds.** CP-1 on water pumping, light and dark: line 1 "Centrifugal pump · Pump house", line 2 only the raw id **"streamsim/pump"** in mono; "Kind:" appears **nowhere** in the inspector (DOM count 0).
- **17f2274 reviewing w-44 with no failed requests — NOT holding on w-44.** The console on the w-44 review still records **404 `/api/file?path=runs/w-44/run.json`** and **422 `/api/actions/run.verify`**, in light and in dark. The fix gates the fetches on a listing entry that is ended-and-incomplete (its e2e covers a destroyed world, whose entry is `ended=true, complete=false`); w-44 is listed as **"not finished"** (`ended=false`), so the gate does not apply and both fetches fire. The state and copy half holds: "Did not finish" chip with icon, the reason "run.json is absent: the run is still open or did not end", the collapsed "Incomplete" line, TRUE RESULTS "Not available yet…", SCORING "Not scored…" — all unchanged from round 11 (ISS-12D-01).
- **158a7b9 truncated breadcrumb keeps its full name — holds.** Hunt-special-names: the truncated plant crumb (`h1`, visible text cut at "…✓ with") carries `title` **and** `aria-label` with the full name ("…with a very long name that keeps going to test wrapping in the header"); the "Workspace" crumb button likewise has both.
- **2397b3a outline rail fades its cut instead of slicing a glyph — holds, both themes.** Campus plant at 1440 × 900: the rail carries `data-cut="true"` while rows continue below, and the bottom ~32 px alpha-masks the cut row ("rcv-u1 air receiver" fades out — visible in light, screenshot 08, and in dark, screenshot 21; the mask blends to the panel in both). Scrolled to the bottom, the attribute is removed (`data-cut` null) and the last row ("mtr-u1 feeder meter") renders whole — no false cue. Water and power rails fit at this height and show no cue.

Riding along in the batches: 37a7c17 (tests for integrity-closed, step keys, legend heading — all three observed holding on screen), 875a1e1 and 69c0a3e (docs only, no UI).

## Carried round-11 items, seen this round

- ISS-11D-01 (integrity detail) — fixed by 1c9089c, holds. ISS-11D-02 review half — addressed by the 939406a cue; plant half parked. ISS-11D-03 (dangling "·") — fixed by 920ab1d, holds. ISS-11D-04 (mark collisions) — fixed by 3bfdbf9, holds. ISS-11D-05 (AREAS order) — partly fixed by 1a1fbe1 (ISS-12D-02). ISS-11D-06 (editor empties interleaved) — fixed by 6611972, holds. ISS-11D-07 ("1 instances") — fixed by b3abb86, holds. ISS-11D-08 (kind twice) — fixed by 25ae10c, holds. ISS-11D-09 (w-44 failed fetches) — not fixed for w-44 (ISS-12D-01). ISS-11D-10 (breadcrumb tooltip) — fixed by 158a7b9, holds. ISS-11D-11 (outline hard clip) — fixed by 2397b3a, holds.
- Still holding in passing: integrity opens collapsed (w-33 `open:false`), one observed time in Inspect, KPI Mean/Range share right edges, timeline chips centre under their marks with the first-fault caption ("First fault · condenser fouling on ch-u2 at 00:10:00 · 1 effector call"), "Show 20 more runs", the Keyboard dialog's per-screen keys, every button named on Home (0 unnamed), legend heading 13 px clear of the popover edge.

## Parked for the graph round (status only)

- **P-12-1 (was P-11-1) fit-vs-floor.** The review half now opens with a right-edge pan cue at the 40 % floor (w-33, light and dark). The plant-view half is unchanged and cueless: campus at 40 % still ends at x 1302 against the canvas edge 1066, Hall 1 cut under the inspector. The `PanCue` component exists and is review-only; wiring it to plant screens is the remaining move. Scored as-is per owner direction.
- **P-12-2 (was P-11-2) full-tier header still does not hug.** Dir-review-empty CP-1 at 100 %: node 176 px tall (was ~250 px in round 11), header band 64 px, with a visible empty band between the id "CP-1" and the kind row "pump". Improved, not closed.
- **P-12-3 (was P-11-3) wire-routing exclusions stand.** No new title or text crossing was observed on the surfaces visited this round; wires under node bodies remain the known exclusion for the routing pass.
- **P-12-4 (was P-11-4) Size preset units (needs-binary).** Not re-checked this round (Build panel not opened); stands as parked.

## Findings

Ordered most severe first.

## ISS-12D-01 Reviewing the abandoned run w-44 still issues its failed requests (404, 422)
- Severity: Medium
- Kind: function
- Where: Review of the abandoned run w-44, light and dark, 1440 × 900. Console: `404 /api/file?path=runs/w-44/run.json` and `422 /api/actions/run.verify`, each with a "Failed to load resource" entry.
- Observed: the batch's fix (17f2274) skips the verify and run-domain fetches only when the listing's entry is **ended and incomplete** — the destroyed-world case its end-to-end test covers. w-44 is listed as "not finished" (`ended=false`), so the gate does not apply, the page still asks for run.json and the verdict, and both fail. The listing plainly knows the state (the Home row and the review header both say "not finished"). Round 10 found this class (ISS-10D-family), round 11 recorded it (ISS-11D-09); the fix misses the exact run it was aimed at.
- Expected: no failed network request in normal use (§3.1 function); a run the listing knows is unfinished — whatever its `ended` flag — is answered from the listing, not the network (A5's crashed-run case).
- Suggestion: widen the gate in `reviewStore.open`/`absentLoads` to every entry the listing knows is `!complete` (synthesize the same failed loads for `ended=false`), and add an abandoned-run (browser-closed) case to the end-to-end test so the covered path is the review-lab path.

## ISS-12D-02 The inspector's AREAS order and the outline rail's order still disagree on the same screen
- Severity: Low
- Kind: coherence
- Where: Plant views of multi-hall-campus and power-distribution-standby, inspector AREAS block vs the outline rail, light and dark, 1440 × 900.
- Observed: campus AREAS reads Site 2, Utilities 9, Hall 1 5, Hall 2 5, Hall 3 5 (the domain's declared `areas` order) while the outline rail on the same screen reads Site, Hall 1, Hall 2, Hall 3, Utilities (populated-first tree order). Power: AREAS Site, Substation, Generator yard, Line A, Line B; rail Site, Line A, Line B, Substation, Generator yard. The commit (1a1fbe1) delivered deterministic plant order and right-aligned counts (both verified), but its stated target — "the plant order the outline shows" — does not hold on either plant, because the rail itself sorts populated-first (6611972's rule).
- Expected: the same thing looks the same everywhere (consistency item of §3.2); the two panels of one screen show one order.
- Suggestion: pick one order function for both panels — either feed AREAS from the outline's populated-first tree order, or drop the rail's re-sort — and assert equality in a component test on campus and power.

## ISS-12D-03 One unfinished-run state carries two names in one header
- Severity: Low
- Kind: content
- Where: Review of w-44, the header chip row, light and dark, 1440 × 900; also the Home row for w-44.
- Observed: the review header shows the listing chip **"not finished"** (lowercase, live style) and, beside it, the state chip **"Did not finish"** (with icon and the reason) — two labels for one state in one bar. The Home row uses "not finished" while the review's state vocabulary says "Did not finish" (and the end-to-end copy for destroyed worlds uses "Did not finish" on the row too).
- Expected: one state, one name (W-1's word rules; the product's one-verb/one-noun register).
- Suggestion: name the state "Did not finish" everywhere (row chip, header chip, listing), and drop the duplicate chip when the state chip is already shown.

Finding counts: Medium 1 (ISS-12D-01); Low 2 (ISS-12D-02, ISS-12D-03); total 3, plus 4 parked graph notes (two updated, one unchanged, one not re-checked).

## Score table (round 12; 1440 × 900 only; light and dark)

| # | Dimension | Score | Evidence |
|---|---|---|---|
| 1 | Layout | 8 | Good: the shell is stable on every surface visited; no horizontal page scroll (scrollWidth 1440); the outline rail fades its cut instead of slicing a glyph in both themes and clears the cue at the bottom (ISS-11D-11 fixed); the review's opening cut now carries a right-edge pan cue where the plant runs past the canvas (w-33, both themes). Against: the plant-view half of the opening fit is unchanged — campus at 40 % ends at x 1302 against the canvas edge 1066 with no affordance (parked P-12-1, scored as-is). |
| 2 | Typography | 9 | Good: the mid-word frame-title wrap is gone — "Reservoir" renders whole on w-13, and the head/tail split only ever falls at a word boundary; run rows wrap at whole "label value" pairs with zero lone "·" lines across 97 measured spans; ids mono, dates whole, numbers grouped (19 089), counts right-aligned in the tallies. No new typographic defect observed in either theme. |
| 3 | Colour and contrast | 9 | Good: the legend's look-alike families now separate by mark as well as hue — Pressure the only diamond, Power the only square, Temperature/Signal hexagons, Heat/Weather bars — verified at legend size in light and dark, and carried by the pins themselves (hexagon clip-path, rotated diamond, 6 × 10 bar, radius 50 % vs 2 px) with hollow/filled still encoding connection; dark primary text measures 15.5:1 on the panel. No contrast defect found this round. |
| 4 | Components | 8 | Good: one field pattern everywhere; the Seed dice and reserved error slots stand; the Keyboard dialog is per screen with the step keys; every button named on Home (DOM check: 0 unnamed); the legend popover keeps its 12 px padding with the heading 13 px clear of the edge. Against: the w-44 review still logs its 404/422 fetch failures (ISS-12D-01). |
| 5 | Content design | 9 | Good: the integrity block now leads with plain sentences ("Reproduces: replaying the run gave the same output.", "Files match the record.") with the command, digest note and file table under For developers; the incomplete banner no longer names the binary; the instance inspector states the kind once and leaves the raw id on its own line; counts singularise ("1 instance in 1 area · 0 links"). Against: the unfinished state carries two names in one header (ISS-12D-03). |
| 6 | Feedback and flow | 8 | Good: the integrity line is collapsed on every review; the abandoned run renders the complete, explained "Did not finish" state with unchanged copy (reason line, TRUE RESULTS "Not available yet…", SCORING "Not scored…"); a finished run is announced once with a direct "Open review" action; clicking a timeline chip still seeks, selects and switches to Inspect. Against: that state's fetches still fail in the console (ISS-12D-01). |
| 7 | Information design of graph and review | 7 | Good: the review's opening overflow is now marked (pan cue at the cut, both themes); frame titles never break mid-word; kind lines complete at every opening zoom (standing from round 11); the review of w-33 at 40 % and w-13 at 41 % stays dense but calm; the legend explains mark and fill. Against: the plant-view opening still cuts the rightmost area with no cue (parked P-12-1); the full-tier node header still does not hug (176 px node, 64 px header band, id-to-kind gap; P-12-2); wires under node bodies remain the known exclusion (P-12-3). |
| 8 | Motion and responsiveness | 9 | Good: zoom steps respond at once; the theme switches instantly in both directions; no layout shift on error slots; the run panel's elapsed clock ticks while the review placeholder time stays stable; the outline fade appears and clears without animation cost; screens opened well inside the 1–3 s budgets on every visit this round. No defect observed. |
| 9 | Craft | 9 | Good: the outline rail's bottom edge fades the row it cuts in both themes and shows no false cue at the bottom (DOM: `data-cut` true at the top, absent at the bottom); the truncated breadcrumb crumb keeps its full text in `title` and `aria-label`; the legend heading sits inside its popover; favicon and per-screen tab titles in place; focus rings visible on keyboard paths. No new craft defect observed. |
| 10 | Whole-product coherence | 8 | Good: Open / Edit / Run / Review keep one verb set; one top bar and breadcrumb pattern with the full name preserved on cut crumbs; dark and light both designed; the integrity block now speaks the product's plain-word register; status chips, error patterns and field styles remain shared components. Against: AREAS order contradicts the outline rail on the same screen (ISS-12D-02); one state, two names ("not finished" vs "Did not finish", ISS-12D-03). |

**Overall: 7** (the lowest dimension is 7, Information design of graph and review; not rounded up). Dimensions 2, 3, 5, 8 and 9 are 9; dimensions 1, 4, 6 and 10 are 8.

## Compared with round 11

Round 11 overall 6 (lowest: Information design of graph and review, 6). Round 12 overall 7 (same dimension, now 7). Per dimension: Layout 7 → 8; Typography 8 → 9; Colour 8 → 9; Components 8 → 8; Content 7 → 9; Feedback 8 → 8; Information design 6 → 7; Motion 8 → 9; Craft 8 → 9; Coherence 8 → 8.

What moved and was verified on screen: all eleven round-11 side-panel findings were addressed; ten of the eleven fixes hold under regression (the editor outline order, the review pan cue, the frame titles, the integrity rewrite, the legend marks, the row wraps, the singular counts, the kind-once header, the breadcrumb title, the outline fade). The two that do not fully hold are the w-44 fetch gate (the fix misses the abandoned-run path it was aimed at) and the AREAS order (deterministic and right-aligned, but still not the outline's order).

What still holds the score down: the parked graph items (plant-half opening fit without a cue, the stretched full-tier header, wires under bodies) keep Information design at 7; the w-44 fetch failures keep Components and Feedback at 8; the two remaining coherence defects (AREAS order, doubled state name) keep Coherence at 8.

## Concrete actions, ranked by impact

### Side-panel actions

1. **Make the no-fetch gate cover every run the listing knows is unfinished** (ISS-12D-01, review store): gate the verify and run-domain requests on `!complete` rather than `ended && !complete`, so an abandoned ("not finished", `ended=false`) run gets the same synthetic loads as a destroyed one; extend the end-to-end test with the abandoned-run path. Takes Components and Feedback to 9–10.
2. **One order for areas, in both panels** (ISS-12D-02, plant overview + outline): feed the inspector's AREAS block from the same order function the outline rail uses (or drop the rail's populated-first re-sort), and pin it with a campus/power component test. Takes Coherence to 9–10 (with action 3) and removes Layout's last non-parked nit.
3. **One name for the unfinished state** (ISS-12D-03, Home row + review header): "Did not finish" everywhere; drop the duplicate "not finished" chip when the state chip is present. Takes Content design to 9–10 and helps Coherence.

### Graph-round actions (parked, recorded — they gate Information design)

1. **Decide the fit-vs-floor rule once** for plant, editor and review (P-12-1): the review now shows the cue at the floor; wiring the existing `PanCue` into the plant screens (campus, power) closes the plant half of the same rule. Verify at 1280/1440/1920 when the graph round lifts the single-viewport restriction.
2. **Tighten the full-tier node header** (P-12-2): hug the id and kind rows at 100 % so a one-channel node is not half empty (today: 176 px node, 64 px header band with an id-to-kind gap).
3. **Wire-routing pass** (P-12-3): wires under node bodies remain; re-verify every area title at every zoom tier after it.
4. **Size preset units** (P-12-4, needs-binary): the component library writes "m³/h" so the Build panel's Size select matches the channels; not re-checked this round.
