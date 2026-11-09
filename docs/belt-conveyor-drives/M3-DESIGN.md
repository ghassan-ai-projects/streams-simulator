# M3 design: prediction and planned isolation

Status: design for M3 (2026-11-09), from [PLAN.md](PLAN.md) section 7 (M3), sections 2.3 (S1, S3, S4) and 5 (the S1-A
walk-through), and decisions O21-O24 (section 9.6), taken while the owner was away. Branch `belt-m3-predict` in
streams-simulator and agentic-stream, cut from M2's tips.

## 1. Target

Barrier-paced closed loops (O23) on plant A and B: S1 (slow bearing on D-3) and S3-A (gear oil loss, then seizure) are
caught early from per-drive vibration and winding residuals against the peers; the line sheds load (R1), runs the belt
empty (R2, approved), stops D-3 by plan (R2, approved; plant permit as defense in depth) or, for S3, stops the empty belt
(R2), then restores the limp feed; S4 (feed surge) never triggers a throughput-reducing action. Every command is read back
(O21). Exit (PLAN.md section 7): 3 seeds each; S1: detection lead, planned stop before the do-nothing trip, no survivor
trip, never stranded, beats B0 and B4; S3: belt empty before the seizure; S4: zero false actions; verify-run passes;
zero quarantines; `streamsim score` credits the right actions (O24).

## 2. Work, in order

| # | Repo | Work |
|---|---|---|
| 1 | streamsim | O24 scoring: command-id matching; `expected_actions` on faults (model, domain-spec, validation, label resolution, scoring scope and unexpected); `Grade`/report for multi-label runs; conveyor faults declare their expected actions |
| 2 | streamsim | O21 read-back: conveyor commands without id arguments; writable run command on drive and belt; boards for drive and belt; feeder board without a fixed id; kind files and stories regenerated; the M2 loop test data in agentic-stream updated to match |
| 3 | streamsim | U7 unequal sharing: a closed algebraic `ratio` form with an independent oracle (equal weights reproduce today's results bit for bit); per-drive weight state set by an instance fault at t=0 (skew in the seed rows later); the stop permit's limit under unequal weights stated or bounded conservatively |
| 4 | streamsim | Calibration: noiseless and noisy runs of S1-A/B, S3-A, S4-A/B rows to fix the detection thresholds and windows the spec uses |
| 5 | agentic-stream | Data: per-device winding and vibration readings, event schemas, catalog routes (`reduce_feed`, `stop_feed` R2, `isolate_drive` R2 to D-3, `stop_belt` R2), principals with a test approver |
| 6 | agentic-stream | Spec `drive-degradation`: `on_close` windows (O22), residuals and slopes against peers, phases normal → watch → degrading → shedding → belt_empty → isolated, one trigger per phase (O23), risk ceiling R2 |
| 7 | agentic-stream | Fast tests on the fake node (hours of backdated readings): the command order, approvals, S4 silent, R2 not denied for completeness; then streamsim-driven barrier-paced tests: S1-A/B and S3-A over 3 seeds, S4 |

## 3. Rules

As in M2: each repository's AGENTS.md, generic code only, no gate loosening, long tests per package with `-p 1` and
`GOMAXPROCS=4`, nothing left running; the fan experiment guards stay green.
