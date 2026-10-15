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

Off-loop pipeline repair (the PR's last CI failure): **a47a98b** deletes the review node text-clash
check ("a faulted node's id, kind, cause and effector outcome never overlap at any zoom step").
The check compared unclipped text runs, so a readout label that
`overflow: hidden; text-overflow: ellipsis` cuts a few pixels short of its value was reported as an
overlap on CI's Linux Chrome, whose font metrics are wider than the review machine's; the pair never
paints over each other. The owner directed deleting the false-positive check rather than teaching the
checker to discount clipped text (`Owner-approved-gate-change` in the commit body). CI is green on
[38797824680](https://github.com/ghassan-ai-projects/streams-simulator/actions/runs/38797824680);
the compact-tier label-intersection guard and the whole-cause wrap check still hold the same class.

The loop's work rides on branch `ui-score-loop`, PR #42 (`canvas: the dimension score loop — side panels and the graph round (overall 3 → 7)`); PR #41 was merged earlier on 2026-10-15 from the same line's predecessor.

Still needs an owner/binary decision: P-11-4 Size preset units ("m³/h" from the component library — Go-side data change, may touch composed-domain bytes and the behaviour pin).

## Graph-round decision framing (fit-vs-floor, P-11-1 / ISS-11D-02)

`openingViewport` (canvas/src/model/viewport.ts) clamps the opening zoom to the 0.4 floor; `tierOf` puts compact at ≥ 0.4 — below it the tier drops/shrinks labels to the 8 px floor (2dda0b2), which is the legibility hole the floor was built to close. The campus at 1440 × 900 naturally fits at ≈ 0.30, so floor ⇒ overflow ⇒ cut (now cued in the review only).

- **Option A (recommended): keep the 0.4 floor everywhere and extend the pan cue** (939406a's `cutSide`/`PanCue`) to the plant view and editor. Kind lines always readable; widest plants open overflowing with an honest affordance; Navigator shows the whole. Cheapest, consistent with the fix that lifted round 10's cap.
- **Option B: fit-first** — reopens sub-0.4 tiers at the opening zoom on wide plants (labels at the 8 px floor again), the exact regression the floor closed in round 10. Needs the overview tier to get genuinely better first.
- Hybrid (fit when overflow ≤ ~15 %, floor otherwise) rejected for now: two rules to test at every tier, against the simplicity priority.

The wire-routing pass (stash `wire hug avoidance`) and the full-tier header hug (P-11-2) are decision-free and proceed either way.

## Infrastructure (this machine)

- Owner's live instance: `127.0.0.1:8766` — DO NOT touch; it serves the original scratch lab in `/private/tmp`.
- Lab backup (durable copy of the ten authored simulations + runs w-1…w-43):
  `/Users/ghassan/my-projects/streams-simulator-review-lab/` (838 MB).
- Review instance for the orchestrator's interactive Chrome tab: `127.0.0.1:8767`, same `--ui-dir canvas/dist`, paired.
- Review instance for Luna's Playwright profile: `127.0.0.1:8768`, paired in `/tmp/canvas-review/pw-profile`.
- Both review instances serve `canvas/dist`; after each fix round run `cd canvas && npm run build` so the
  review sees the fix.
- Screenshots and review scripts live under `/tmp/canvas-review/` (never committed).
- `governance/CANVAS_UI_ACCEPTANCE.md` §3 is the review rubric; `docs/simulation-canvas/UI-EXPECTATIONS.md`
  the numbered expectations.

## Parked work

- **Stashed wire-routing work** (`git stash`): "wire hug avoidance (for the graph dimension round)" —
  grid hug margins (wireGrid), frame margin lanes + crowd occupancy (wireRoute/wireSearch), routeView wiring
  (wireView). Model unit tests passed with it, but it had no test of its own and no committed coverage; the
  graph round must land it with tests or redo it.
- Round-10 open graph items for that round: ISS-10D-02/04 (overview-tier kind lines, opening zoom floor),
  ISS-10D-03 (review readout kind collisions), ISS-10D-06 (wire through area title), ISS-6D-35 (frames cut
  at inspector edge — partly side panel), ISS-6D-20 (wires under node bodies, known exclusion).
