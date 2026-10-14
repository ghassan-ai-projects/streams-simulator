# P9 handoff — playback and review (2026-10-14)

State at the head of `canvas-foundation` on branch `canvas-foundation` (86 commits ahead of `main`, not pushed; no PR yet).

## Where it stands

- The canvas runs on real workspace files through the bridge (`streamsim canvas`): pairing, workspace home, plant view
  (factory and a 26-hall campus, halls open by Enter, double-click or the outline), and run review (playback, scrubber
  with ticks and event chips, verdict badge with verify's reason, file checks, fault intervals and declared alarm
  states on nodes, Inspect with "Active now", KPI truth, hall rollups).
- The canvas computes nothing the binary owns: fault intervals, channel state vocabularies and the verify reason come
  from the simulator (R9-15–R9-17, R9-20, R9-21).
- Gates: `make ci-check`, `make quick`, and `cd canvas && STREAMSIM_BIN=$(cd .. && pwd)/bin/streams-simulator npm run
  check` (124 unit and component tests, 15 Playwright tests against the real binary) and `make phase-gate` (including the Enola check) all pass.
- Review loop: five rounds with three reviewers (design director, accessibility, engineer); record in
  [ui.md](ui.md). Every reviewer is at 3/5 with the round 5 findings fixed but not yet re-reviewed. The bar is all
  three at 4 or more with no High.

## Next session, in order

1. **E5-1, E5-2, E5-4 are fixed** (`cf24c70`, R9-30–R9-32): an incomplete run never verifies as a match (reasons
   `incomplete`, `incomplete_stamp`), report loads run.json as verify does and reports a missing digest as a failing
   check, and an open evaluate report says truth is sealed until the run ends. Check in round 6 that the canvas never
   reads "Reproduces" for an incomplete run.
2. **Round 6**: freeze a build (copy `bin/streams-simulator` and `canvas/dist`), record a fresh review workspace (recipe
   below), send the three reviewers with the previous findings and the round 5 dispositions. Loop until approved.
3. **Simulator items that block P9 sign-off** (reviewers' judgement): R9-24 alarm activations in the report (lane
   chips, "jump to alarm"), R9-13 Explore true state vs reported (G6, S5), R9-14 report/verify memory and time on the
   campus (8 GB, minutes), R9-22 negative air flow from unfloored measurement noise.
4. Then P10 (live run and control) and P11 (builder), and the acceptance tests A1–A7 with complex simulations.

## Review workspace recipe

Compose `examples/compositions/factory-utilities.composition.json` and a 26-hall campus composition (the e2e fixture
`canvas/e2e/fixtures/campus-3.composition.json` shows the shape) into `domains/`, then record with the current binary:

- `w-factory-evaluate`: `-domain factory-utilities -seed 11 -duration 28800 -fault AC-1=compressor_lead__oil_cooler_fouling@21600`
- `w-factory-explore`: the same plus `-fault CT-1=cooling_tower__fan_failure@10800 -explore`
- `w-factory-edited`: `-seed 5 -duration 7200`, then append `{}` to its `trace.jsonl` (must read Not verified)
- `w-campus`: `-domain factory-utilities-campus-26 -seed 3 -duration 3600 -explore` with faults on `h07/M-1`, `h07/M-2`, `h19/M-1`
- Edit the factory domain's description afterwards (the review must show the "plant file changed" note).

Runs recorded by an older binary do not verify (evidence digests became required in R9-15); re-record after a change
to the run artifact.

## Known open findings

P8b is built but R0-15 (bridge threat model, one test per threat), R0-16 (stress fixture) and R2-01 (size bound on
every input read) are open, so it is not marked Done (the findings gate enforces this). R9-01 world ids collide; R9-04
8 MiB file cap; R9-05 playback streams the whole journal; R9-13, R9-14, R9-22–R9-26
(see `governance/FINDINGS.md`); D5-12 hall summary figure; wires over frame titles (layout round).
