# UI polish through simulation authoring — 2026-10-15

Goal: ten reasonable, useful, multi-level simulations authored through the canvas UI in Chrome; every comprehension or
functional problem met on the way is reported here, fixed one commit per fix, and the interface re-evaluated until it
scores 10/10.

Roles: an author (Haiku) only builds the simulations through the UI and reports blockers; a hunter (Haiku) separately
hunts for comprehension, functional and improvement issues, including real-world failure probes. Both append to
`issues/round-N-*.md` as they go; a Sonnet fixer fixes them one commit each while they work; the orchestrator reviews every
diff; a detail-oriented UI/UX director (Sonnet, Chrome screenshots at 1280/1440/1920, light and dark) scores ten dimensions
in `evaluation-round-N.md` and owns the 10/10 verdict, which is the lowest dimension score. The loop ends when ten simulations exist and the evaluation is 10/10.

The authored plants live in a scratch workspace outside the repository; each round's notes list their ids, areas, equipment,
instruments and the actions exercised.

## Status note (round 10 reached; stopped by the goal's stop rule)

Outcome, as measured. The goal is not met: the interface is not at 10/10.

| Director evaluation | Overall | Lowest dimension |
|---|---|---|
| Round 6 | 4 | content, feedback, graph (all 4) |
| Round 8 | 4 | graph and review information design (4) |
| Round 10 | 3 | graph and review information design (3) |

Round 10 also found a regression (kind line printed over the first port row at the opening zoom), fixed afterwards in
`f76dff2`; the director has not re-scored since. Only 1440 px wide could be reviewed in every round (the Chrome window did not
resize), so 1280 and 1920 are not covered.

Done:
- Ten simulations built only through the canvas and run (Explore and Evaluate, two faults, effector calls): compressed-air-works,
  office-tower-hvac, water-pumping-station, data-centre-cooling, packaging-factory, boiler-steam-loop, greenhouse-ventilation,
  food-cold-chain, power-distribution-standby, multi-hall-campus. They live in the scratch workspace outside the repository.
- About 170 commits on the branch: authoring fixes (blank plant, repeated kinds, areas, name rules, delete confirmation, leave
  confirmation), effector calls in the run plan, run listing with seed and completion, abandoned runs recorded incomplete, bridge
  catalog reload after runs, empty plants that compose, wording passes over the report, and graph legibility work.
- Gates: `make ci-check` passed from one continuous run after the last Go change; `cd canvas && npm run check` passed after the
  last canvas change (324 unit tests, 44 e2e).

Open, and why the score is capped:
- Graph and review information design. Wires still run behind node bodies and between same-column nodes (needs a routing pass,
  which the owner scoped out earlier); per-node height in the review; the overview tier cuts long kinds (ISS-10D-02 remainder,
  ISS-10D-04 opening-zoom floor); wires crossing area titles regressed in round 10 (ISS-8D-02).
- Decisions for the owner: (1) approve a wire routing pass or accept the cap; (2) refresh `test/pin/BEHAVIOUR_PIN.txt` (one stale
  entry, the CLI usage text, needs an `Owner-approved-gate-change:` line); (3) port descriptions and the missing electrical and
  load links in the component library (see `simulator-backlog.md` B-9, B-19, B-21); (4) a sampled-loop connection option (B-20);
  (5) whether sealed truth must survive a director restart after a catalog change.
- Simulator behaviours found by the failure probes are in `simulator-backlog.md` (alarms never raised, frozen reading without a
  quality flag, no feeder trip, no late or duplicated readings, unconnected equipment running as healthy).

