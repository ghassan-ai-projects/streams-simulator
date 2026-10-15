# Dimension score loop — 2026-10-15

The owner directs a dimension-by-dimension loop on the director's score table: take the dimensions from
lowest to highest score, ask the director (Luna) for the concrete actions that raise the score, fix them,
re-score, and loop on one dimension until it reaches 9–10, then move to the next. One commit per fix round.
The owner's added priority for this loop: **side panels first** (inspector, editor outline rail, run panel,
report column, legend popover, keyboard dialog) — canvas-graph findings are parked for a dedicated graph
round at the graph dimension's turn.

Round numbering continues the evaluation series (`evaluation-round-N.md`).

## Baseline (round 10)

| Dimension | Round 10 |
|---|---|
| Information design of graph and review | 3 |
| Typography | 4 |
| Layout | 5 |
| Content design | 5 |
| Feedback and flow | 5 |
| Craft | 5 |
| Whole-product coherence | 5 |
| Colour and contrast | 6 |
| Components | 7 |
| Motion and responsiveness | 7 |

Loop order: graph information design (3) is lowest but is canvas-graph work — it runs **last of the sub-5
set or as its own dedicated round** with the parked routing work; the side-panel dimensions run in score
order first per the owner's focus: Typography (4) → Layout (5) → Content (5) → Feedback (5) → Craft (5) →
Coherence (5) → Colour (6) → Components (7) → Motion (7) → Graph (3, dedicated round).

## Round log

| Round | Dimension focus | Result | Commits |
|---|---|---|---|
| 11 | Side panels across all ten dimensions (review) | overall 6 (was 3); lowest now Information design of graph and review 6; Layout 7, Content 7, rest 8; 11 findings (ISS-11D-01…11) + 4 parked graph notes; evaluation-round-11.md | — |
| 12 | Side-panel fix batch 1: 11D-01/02(review half)/04/06 + regression tests for already-fixed 10D-07/09/10/12 | done — 5 commits (6611972 outline order, 939406a pan cue + whole-word titles, 1c9089c plain-words integrity, 3bfdbf9 legend marks, 37a7c17 gap tests); `npm run check` green (411 unit, 57 e2e); visually verified in Chrome | 6611972, 939406a, 1c9089c, 3bfdbf9, 37a7c17 |
| 13 | Side-panel fix batch 2: 11D-03/05/07/08/09/10/11 | done — 7 commits (920ab1d MetaParts wraps, 1a1fbe1 areas in plant order, b3abb86 singular counts, 25ae10c inspector kind once, 17f2274 no doomed fetches on incomplete runs, 158a7b9 breadcrumb aria-label, 2397b3a outline bottom fade); gates running at ledger-update time | 920ab1d, 1a1fbe1, b3abb86, 25ae10c, 17f2274, 158a7b9, 2397b3a |
| 14 | Round-12 re-score (owner: ONE viewport only — 1440×900, light and dark) | overall 7 (was 6); Typography/Colour/Content/Motion/Craft 9; Layout/Components/Feedback/Coherence 8; graph 7; 3 findings (ISS-12D-01…03); all 12 fixes hold except 12D-01's gate width; evaluation-round-12.md | — |
| 15 | Side-panel fix batch 3: 12D-01 gate widened to every listing-known `!complete` run (listing carries no abandonment signal — investigated), 12D-02 one `areaOrder` for rail + AREAS, 12D-03 one state name "Did not finish" | done — 43fb7ab, 3dc9570, 031d0c8; `npm run check` green | 43fb7ab, 3dc9570, 031d0c8 |
| 16 | Graph round: pan cue into plant+editor (Option A), full-tier header hug, wire-routing pass (stash) | done — 7c06d48 pan cue (plant+editor), 0348a3e wire routing with the stashed hug-avoidance and 6 new tests (grid hug marking, crowd, path choice; the popped code refactored under the complexity gates), d472050 full-tier header hug (CSS, rows hug under the category band). The failed fixer session (usage limit) left a half-threaded tier-aware header height that would have desynced wire ends from rendered rows — reverted in favour of the contained CSS hug; a true tier-aware layout stays deferred with ISS-8D-17's won't-fix rationale. | 7c06d48, 0348a3e, d472050 |
| 17 | Round-13 re-score (LUNA, 8768, one viewport 1440×900, light + dark) | done — overall 8 (was 7); every dimension ≥ 8, lowest Craft 8 and Whole-product coherence 8; all six fix commits hold (43fb7ab surface answered from the listing, 3dc9570 one area order, 031d0c8 one unfinished name, 7c06d48 pan cue on plant/editor/review, 0348a3e wire routing, d472050 header hug); 2 Low findings (ISS-13D-01 legend heading overflows its popover, ISS-13D-02 editor opens below the plant view's floor); P-12-1 closed as a rendering question, P-12-2 closed, P-12-3 not reproducible in that workspace, P-12-4 not re-checked; evaluation-round-13.md, 68 screenshots in `/tmp/canvas-review/round-13/` | — |
| 18 | Side-panel fix batch 4: ISS-13D-01 legend heading overflow + ISS-13D-02 one opening rule | done — 6b52e29: `.legendTitle` wraps inside its popover on every screen, and the heading check measures the painted text runs (the element-box blind spot that let the overflow through); the editor drops its `minFitZoom={0.25}` so plant view, review and editor open on the one `openingFloor` (0.4) rule, with `overviewFloor` named, a model rule test and a cross-screen e2e check (factory-utilities opens at 0.4 on both screens). UI-EXPECTATIONS E-15 now states the implemented rule. | 6b52e29 |
| 19 | Round-14 re-score (LUNA, 8768, one viewport 1440×900, light + dark) | done — overall 9 (was 8); **all ten dimensions 9** (Craft 8→9, Whole-product coherence 8→9); **no findings** (Critical/High/Medium/Low 0); both round-13 findings closed by independent measurement — the legend heading's painted runs end 9.75 px inside its popover content box on two plant screens and on a review (was 4.7 px past the border), and the campus plant view and editor are identical to the pixel (0.4, tier-compact, 236.8 px cut, same right cue) while a fitted plant opens fitted with no cue on both; the ceiling is scope, not defect (1440 px only by owner direction; P-12-2/03/04 unverified in that workspace); evaluation-round-14.md, 42 screenshots in `/tmp/canvas-review/round-14/` | — |

**Loop status after round 19: the target band is reached — every dimension is 9 and no finding is open.** The
owner's loop was to take each dimension to 9–10; the remaining distance to 10 is scope rather than defect:
the two widths the owner restricted (1280 and 1920 were not tested), P-12-2's tier-aware full-tier header
(deferred deliberately with ISS-8D-17's rationale), P-12-3's wire-under-body exclusion (its plant is absent
from the review workspace; the e2e fixture has it), and P-12-4's size-unit normalisation, which waits on the
owner's choice between the three recorded options.

Off-loop pipeline repair (the PR's last CI failure): **a47a98b** deletes the review node text-clash
check ("a faulted node's id, kind, cause and effector outcome never overlap at any zoom step").
The check compared unclipped text runs, so a readout label that
`overflow: hidden; text-overflow: ellipsis` cuts a few pixels short of its value was reported as an
overlap on CI's Linux Chrome, whose font metrics are wider than the review machine's; the pair never
paints over each other. The owner directed deleting the false-positive check rather than teaching the
checker to discount clipped text (`Owner-approved-gate-change` in the commit body). CI is green on
[38797824680](https://github.com/ghassan-ai-projects/streams-simulator/actions/runs/38797824680);
the compact-tier label-intersection guard and the whole-cause wrap check still hold the same class.

The loop's work rides on branch `ui-score-loop`, PR #42 (`canvas: the dimension score loop — side panels, the graph round and the round-13 fixes (overall 3 → 8)`); PR #41 was merged earlier on 2026-10-15 from the same line's predecessor.

Still needs an owner/binary decision: **P-12-4 (was P-11-4) Size preset units.** Scope verified 2026-10-15 (read-only): the canvas renders channel units typographically — `canvas/src/model/units.ts` maps `m3/h` → `m³/h` — so a plant readout shows "m³/h", while the Build panel's Size select shows the component library's preset *title* verbatim ("Small (22 kW, 200 m3/h)"), which is ASCII `m3/h`. The library holds 51 such `m3/h` strings in 15 files; nothing writes `m³/h` in Go data today, so this is a normalisation, not a mismatch between two live formats. Preset titles are display-only — neither `examples/factory-utilities.domain.json` nor its composition carries "Small (22 kW" — so composed-domain bytes are unaffected; anything that *prints* component titles (the component listing the behaviour pin hashes) would change, so the pin needs a re-run and a reviewed diff. Options: (a) normalise the 15 library files to `m³/h` and re-pin, (b) leave data-authored prose as the library writes it, (c) have the binary return structured preset facts and let the canvas format them — the largest of the three and the one the architecture rule prefers if the canvas is ever to show a size in the reader's units.

## Polish pass (after the scored loop) — 2026-10-15/16

The scored loop closed at round 19 with every dimension at 9 and no open finding. The owner then asked for a UI/UX
improvement pass, run as three independent reviews and one fix batch on this same branch. Its records are:

- **Three review streams.** `improvement-hunt-20261015.md` — an unconstrained hunt against the then-served bundle:
  22 `IMP-NN` improvements plus 7 bigger bets, the ten dimensions swept one by one, 33 axe runs with zero
  violations, 64 screenshots. `COMPREHENSION-IMPROVEMENTS-2026-10-15.md` and
  `COMPREHENSION-ROUND-2-2026-10-15.md` — 25 + 25 comprehension proposals from a separate session against a
  temporary lab copy, explicitly proposals rather than approval.
- **One consolidated backlog** (`improvement-backlog-20261015.md`, `e30a50e`): 79 source items → **68 `CB-NN`
  entries**, a 79-row equivalence map (one row per source item, script-verified, 11 merges absorbing 18 items),
  a coherence check that settles contradictions against the source — the Link form *does* filter by port type
  (`ConnectForm.tsx:47-51`), so the hunt's "unfiltered" claim is stale; the keyboard dialog is dimmed by a
  sibling `.scrim` div, so its `::backdrop` finding is rejected rather than fixed; the run-panel footer was
  already fixed. Entries opposing deliberate decisions (editor rail order, the resolved fit-vs-floor rule,
  ISS-8D-17's header band, the six-row library list) are surfaced with the tradeoff named, and the document
  states what it did not reach.
- **Presentation fixes landed**, each with a test that fails without it: `bfabd05` the run plan's fields scroll
  above a fixed footer (the hunt's top item); `f26b0d4` the zoom controls carry the tooltip their name already
  had; `044b569` the plant view and a review promise selection, not deletion; `ee34c84` the equipment list says
  how many of the library it is showing; `71b71ee` the right panel has one name per screen, in the help and the
  toggle; `113e937` the Link form's target list names each instance's kind; `5431282` the rail writes an area
  name in the row's own style rather than borrowing the section-heading style, with the disclosure affordance
  split out so headings keep their uppercase; `dd755ee` a run with no recorded end wears the warn chip
  instead of the success tokens, with the listing's missing abandonment signal recorded rather than guessed.
  `6c168b8` the empty fault list says only what the plan knows instead of claiming plant health.
  `IMP-02` was dropped rather than implemented, because the backlog proved the premise wrong.

Two deliberate divergences from the backlog: the panel naming uses one truthful name per screen
(`rightPanelName` — Inspector on home and plant, Run details on a review, Build in the editor) with the toggle,
its title and the help pinned to it by a per-screen coherence test, instead of "Inspector" everywhere; and
`CB-41` (the editor's doomed domain probe) stays in the backlog because its fix changes how the app resolves a
domain. `CB-16`'s open question — the workspace listing carries no abandonment signal, so a recorded run that
did not finish cannot be told from one still open — is binary-owned and recorded, not guessed in the canvas.

Every item accepted for this batch has now landed; `IMP-02` was dropped because the backlog proved its
premise wrong, not deferred. The backlog's remaining entries — including its 12-item recommended batch and the
owner decisions it names — are the next pass's work list.

## Graph-round decision framing (fit-vs-floor, P-11-1 / ISS-11D-02) — resolved

**Resolved in favour of Option A across rounds 16 and 18.** The cue is on the plant view, the editor and the review (7c06d48), and every screen now opens on the one 0.4 `openingFloor` rule (6b52e29 removed the editor's `minFitZoom={0.25}`), so the framing below stands as the record of the decision, not as open work. What the floor does *not* constrain: the Fit control and manual zoom still reach the overview tier (measured on the campus editor: Fit → 0.302 with every node inside and no cue), which is what the Navigator is for.

`openingViewport` (canvas/src/model/viewport.ts) clamps the opening zoom to the 0.4 floor; `tierOf` puts compact at ≥ 0.4 — below it the tier drops/shrinks labels to the 8 px floor (2dda0b2), which is the legibility hole the floor was built to close. The campus at 1440 × 900 naturally fits at ≈ 0.30, so floor ⇒ overflow ⇒ cut (cued where it happens on all three screens).

- **Option A (recommended): keep the 0.4 floor everywhere and extend the pan cue** (939406a's `cutSide`/`PanCue`) to the plant view and editor. Kind lines always readable; widest plants open overflowing with an honest affordance; Navigator shows the whole. Cheapest, consistent with the fix that lifted round 10's cap.
- **Option B: fit-first** — reopens sub-0.4 tiers at the opening zoom on wide plants (labels at the 8 px floor again), the exact regression the floor closed in round 10. Needs the overview tier to get genuinely better first.
- Hybrid (fit when overflow ≤ ~15 %, floor otherwise) rejected for now: two rules to test at every tier, against the simplicity priority.

The wire-routing pass (the stashed `wire hug avoidance`) and the full-tier header hug (P-11-2) were decision-free and both landed in round 16 (`0348a3e`, `d472050`).

## Infrastructure (this machine)

- Owner's live instance: `127.0.0.1:8766` — DO NOT touch; it serves the original scratch lab in `/private/tmp`.
- Lab backup (durable copy of the ten authored simulations + runs w-1…w-43):
  `/Users/ghassan/my-projects/streams-simulator-review-lab/` (838 MB).
- Review instance for the orchestrator's interactive Chrome tab: `127.0.0.1:8767`, same `--ui-dir canvas/dist`, paired.
- Review instance for Luna's Playwright profile: `127.0.0.1:8768`, paired in `/tmp/canvas-review/pw-profile`.
- Both review instances serve `canvas/dist`; after each fix round run `cd canvas && npm run build` so the
  review sees the fix.
- Screenshots and review scripts live under `/tmp/canvas-review/` (never committed).
- The full local `npm run check` is load-sensitive on this machine: three wall-clock specs can flake
  while the whole suite runs in parallel — the alarm-playback "Verified" wait, launch's 3 s
  effector-mark hold, and plant's "a selection made outside the graph comes into view" — and each
  passes with `--workers=1` and in CI. Run the suite on a quiet machine, or re-run a failure alone
  before treating it as a regression.
- `governance/CANVAS_UI_ACCEPTANCE.md` §3 is the review rubric; `docs/simulation-canvas/UI-EXPECTATIONS.md`
  the numbered expectations.

## Parked work

- **Stashed wire-routing work** — **landed and consumed.** Round 16 (`0348a3e`) took the stashed work
  (grid hug margins in `wireGrid`, frame margin lanes + crowd occupancy in `wireRoute`/`wireSearch`,
  `wireView` wiring) and gave it the tests it lacked (grid hug marking, crowd, path choice), so
  `git stash list` is empty now and the redo instruction no longer applies.
- Round-10 open graph items, status after rounds 16–18: ISS-10D-02/04 (overview-tier kind lines, opening
  zoom floor) — closed; the floor decision is Option A and every screen opens on it (round 18), kind
  lines are complete at every opening zoom. ISS-10D-03 (review readout kind collisions) — closed by the
  40 % / 8 px readout floor and the label-intersection guards. ISS-10D-06 (wire through area title) —
  closed by the routing pass; round 13 measured 0 title crossings across eleven plant/zoom combinations.
  ISS-6D-35 (frames cut at the inspector edge) — now the accepted, cued floor behaviour. ISS-6D-20
  (wires under node bodies) — still recorded: round 13 could not reproduce it, because the
  factory-utilities plant that showed it is not in the review instance's workspace (the verifier's
  factory-utilities routes answer 404 there). The plant is reachable in the e2e fixture — `canvas/e2e/support/workspace.ts`
  composes `examples/compositions/factory-utilities.composition.json` into the fixture's workspace — so a
  Playwright check over that fixture is the practical place to settle it before calling it closed.
