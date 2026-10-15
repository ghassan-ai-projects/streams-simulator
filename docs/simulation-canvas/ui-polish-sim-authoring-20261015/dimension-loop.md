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
| 12 | Side-panel fix batch 1 (fixer, in flight): 11D-01/02(review half)/04/06 + regression tests for already-fixed 10D-07/09/10/12 | in progress | — |
| 13 | Side-panel fix batch 2 (planned): 11D-03 (run-row separator wrap), 11D-05 (AREAS in plant order), 11D-07 ("1 instance"), 11D-08 (instance header second line = raw id), 11D-09 (quiet 404/422 on incomplete runs), 11D-10 (breadcrumb crumb tooltip), 11D-11 (outline bottom fade) | pending | — |
| 14 | Graph round (planned, dedicated): fit-vs-floor rule once for plant/editor/review (P-11-1/11D-02), full-tier header hug (P-11-2), wire routing pass with the stashed hug work, area-title re-verify at every tier | pending | — |

Still needs an owner/binary decision: P-11-4 Size preset units ("m³/h" from the component library — Go-side data change, may touch composed-domain bytes and the behaviour pin).

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
