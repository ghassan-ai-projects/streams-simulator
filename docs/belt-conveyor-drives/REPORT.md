# Belt conveyor drives: project report

Status: M0-M4 done (2026-11-10). Pull requests open, not merged: streams-simulator
[#60](https://github.com/ghassan-ai-projects/streams-simulator/pull/60), agentic-stream
[#75](https://github.com/ghassan-ai-projects/agentic-stream/pull/75), tamoz
[#86](https://github.com/ghassan-ai-projects/tamoz/pull/86). Design and decisions: [PLAN.md](PLAN.md) (§9.6, O1-O41) and
[M0](M0-DESIGN.md)-[M4](M4-DESIGN.md) design files. A public-facing version (what, why, data, evidence only) is
[WEBSITE.md](WEBSITE.md).

## 1. The problem

A 60 m belt conveyor is pulled by three drives that share the belt pull by torque. When one drive weakens or stops, its
share moves at once onto the other two. With a large rating margin (regime A, plant A) the survivors run near rating and
age faster; with a small margin (regime B, plant B) they overload, trip within about two minutes, and the belt stops
loaded. Two drives cannot restart a loaded belt (the restart trap). The goal: turn the failing drive's precursors into an
early, correct, governed reaction (shed load, run the belt empty, stop the weak drive at a chosen moment, keep the belt
restartable) and never be fooled by a feed surge or a dead sensor.

The loop: streams-simulator (the plant, sealed truth, grading) → agentic-stream (situations, policy, actuation) → Tamoz
(judgment by a language model) → back to the plant over the plant wire.

## 2. What was built

| Milestone | Repos | Built |
|---|---|---|
| M0 scenarios as data | streamsim | Conveyor component library (drive, belt, feeder, instruments), plants A and B as compositions, inverter overload trip, starting boost, seized drag, downstream run permit, 2-6 drives; seven stories (S0-S6) × 10 seed rows; a hand-written reference policy per story |
| M1 truth and grading | streamsim | Outcome recorder in the world; `streamsim outcome`: the run replayed as run, with nobody acting (B0) and under fixed baseline policies; cost J; session `horizon_s`; report `--outcome` |
| M2 first closed loop | streamsim, agentic-stream | S2-B through the plant wire: a trip seen, the feed cut to the safe limp feed, read back through kind files |
| M3 prediction and planned isolation | streamsim, agentic-stream, tamoz | Degradation caught by trend against peers; planned isolation with human-approved (R2) steps; unequal load sharing (U7); 12/12 scripted rows |
| M4 hard cases and evidence | all three | Baselines B1-B4; two labels per entity; prognostic grading; severity ranges; S5 (frozen sensor), S5b (dropout), S6 (lead drive lost); act before the cause is known; watches that wake reasoning; retry of failed reasoning; advisory inputs; failure-time prediction; plant-B planned stop; live Tamoz test for eight stories; report readability |

M4 in detail, by repository:

- **streams-simulator:** B1-B3 closed baseline records and the `easy` flag (O27); labels keyed by entity and fault
  (O25); `predicted_failure_at` graded with α-λ, relative accuracy and prognostic horizon (U9); campaign
  `severity_range` (U8); S6 plants on `conveyor-drive@2` (master/follower roles, O33); S5b story; a held reading first
  observable at its injection (O26); concurrent outcome replays (346 s → 191 s on a 16 h story); a cancelled replay stops;
  run report states the winner, the cost unit, baseline names, control stories, KPI counters by change; CLI failure prefix.
- **agentic-stream:** act on what is known before the cause is (O34); `drive_degrading` symptom code (O36); a watch
  fires on a rising edge and wakes its Situation, failed reasoning retried as a new episode (O35); a version is material
  only through a trigger that applies to it (O41); advisory inputs, loss as a fact `input_stale.<name>` (A3);
  `extrapolate` operator, quadratic fit (A5, O37); verdicts carry `predicted_failure_at`, a first decision per (entity,
  diagnosis), symptom folded into the causes that refine it; specs: drive-degradation (S1, S3, S5, S5b), plant-B planned
  stop (S1-B), line-overload (S2), master-loss (S6); per-plant device catalogs; twin and live tests.
- **tamoz:** the episode frame shows the situation's phase; the repair message lists the allowed intents (T2); evidence
  references accept only the prefixes the frame offers (T3); owner rule "Tamoz carries no caller's domain" (ADR-058).

## 3. Results

### 3.1 Simulator reference policy (all 10 seed rows, noiseless)

| Story | Reference J | Do nothing J | Delta (negative = reference better) |
|---|---|---|---|
| S1-A | -11.09 .. -10.70 | -8.43 .. -6.02 | -4.77 .. -2.52 |
| S1-B | -10.16 .. -8.05 | -4.59 .. -2.01 | -6.03 .. -5.57 |
| S2-A | -1.64 .. -1.48 | 3.48 .. 4.02 | -5.51 .. -5.05 |
| S2-B | -0.84 .. -0.59 | 4.99 .. 5.29 | -5.88 .. -5.83 |
| S3-A | -1.98 .. -1.01 | 1.45 .. 2.77 | -3.90 .. -3.35 |
| S5-A, S5b-A | -11.07 .. -10.64 | -8.07 .. -5.71 | -4.95 .. -2.72 |
| S6-A-promote | | | -5.47 .. -4.94 |
| S0-A, S4-A, S4-B, S6-A-stop | | | 0 (ties: control or unavoidable) |

### 3.2 Arm A: rules only, through agentic-stream, 10 seeds (2026-11-10)

113/113 runs pass on the barrier-paced twin (streamsim `ebae07b9`, agentic-stream `1269adf6`), 0 quarantined events,
`verify-run` on every run, 3 h 23 min wall at 4 processes.

| Story | Rows | Result | J - B0 | J - B1 | Detection latency |
|---|---|---|---|---|---|
| S1-A bearing, plant A | 10 | beats B0-B4 every row, label correct | -3.17 .. -1.01 | -2.62 .. -1.51 | 0.34-0.87 h |
| S1-B bearing, plant B | 10 | beats B0-B4 every row | -6.11 .. -5.62 | -0.87 .. -0.44 | 0.49-1.17 h |
| S3-A gearbox | 10 | beats B0-B4 every row | -3.51 .. -3.06 | -3.51 .. -3.06 | 0.46-0.55 h |
| S5-A frozen sensor | 10 | beats B0-B4, both labels | -4.22 .. -1.97 | -3.42 .. -2.46 | 0.24-0.27 h |
| S5b-A dropped sensor | 10 | beats B0-B4, both labels | as S5-A | as S5-A | 0.24-0.25 h |
| S6-A-promote | 10 | beats B0-B4 | -5.48 .. -4.95 | -5.48 .. -4.95 | 10-13 s |
| S6-A-stop | 10 | ties every baseline (`easy` by design) | 0 | 0 | 12-16 s |
| S4-A, S4-B, S4-A and S0-A on the planned-stop spec | 40 | no step, no false detection | - | - | - |
| S2-B | 3 fixed seeds | beats B0, B1, B2, B4; ties B3 (`easy`) | -5.49 | -5.49 | 8 s |

### 3.3 Arm B: Tamoz with GLM 5.3 Flash, seed rows 1-3

| Story | Result | Round | J - B0 | Calls per run |
|---|---|---|---|---|
| S2-B | 3/3 | 2 | -5.49 | 1 |
| S3-A | 3/3 | 2 (row 2), 3 (rows 1, 3) | -3.06 .. -3.38 | 3 |
| S1-A | 3/3 | 4 | -2.40 .. -3.20 | 4 |
| S1-B | 3/3 | 6 | -5.63 .. -5.94 | 5 |
| S5-A | 3/3 | 6 | -2.19 .. -3.93 | 5-10 |
| S5b-A | 3/3 | 6 | -2.19 .. -3.86 | 5-6 |
| S6-A-promote | 3/3 | 5 | -5.14 .. -5.35 | 1 |
| S6-A-stop | 3/3 | 5 | 0 (ties, `easy`) | 1 |

S1-A over the rounds: 0/3 (the model answered `unknown` and only watched), 0/3 (Tamoz timeout, rejection), 0/3 (Tamoz
refused steps under `unknown`), 3/3 once the symptom code `drive_degrading` existed. Round 6 cost 52 calls, 116 k input
and 190 k output tokens for 9 runs; no Tamoz failure.

### 3.4 Prediction (S1-B, A5)

Calibration on the 10 S1-B do-nothing rows: a quadratic fit over 60 min of D-3's winding toward its 150 °C trip
predicts the trip 1.42-1.47 h ahead (S1-A 1.40-1.44 h), late by 0.03-0.10 h; a linear fit leads 1.0-1.2 h and never
falls inside α = 0.2. On the twin and live, D-3 is stopped 1.34-1.37 h before its do-nothing trip.

### 3.5 End-to-end through the CLI

Every story run, verified, graded and reported through the documented commands: every outcome matches the design
tables, repeat runs byte-identical, `make phase-gate` green (172 canvas browser tests, Enola, behaviour pin).

## 4. Outcome

The plan's acceptance holds for arm A: every positive story's pass rule over 10 seeds, every control story clean, every
`easy` flag explained. The live model reaches the same decisions on three seeds per story once the loop around it names
the symptom and keeps asking. Finding: the model alone holds back when the cause is uncertain.

## 5. Missing, known limits

- **Evidence gaps:** live runs are 3 seeds per story (T6 at 10 fresh seeds not run); S1-A, S2-B, S3-A live results come
  from rounds before the last engine changes; arm C (approved playbook rule plus Tamoz) not run; the agentic-stream S2
  twin plays S2-B on three fixed seeds and no S2-A; the S1-B prediction was calibrated on the same rows it is judged on.
- **Needs the owner:** live approval of the S2 playbook rule (O30); T6 at scale; the fan experiment FR+ rung.
- **Simulator gaps:** no operator-message route and no empty restart (S6); window operators close windows per input, so
  features flip at minute boundaries; the world spends about 80 % of a step recomputing algebraic states.
- **Consumer gaps:** watch fires in a phase whose trigger names `watch_fired` can still withdraw an approval during a
  long episode; a prediction is published only on a version change (2-3 per run); the plant-B spec is a near copy of
  plant A's.
- **Tamoz:** an in-flight model call is not stopped at the wall budget; no level between one cause and `unknown`
  (future plan `tamoz/docs/ACT_UNDER_UNCERTAINTY_PLAN.md`); the pump-station evaluation fixture copies a caller's spec.

## 6. Next

1. Review and merge the three PRs.
2. T6: 10 fresh seeds per story on the final setup, arm B (about 80 runs, about 10 h at 2 at a time).
3. Approve the S2 playbook proposal live; run arm C.
4. Tamoz future plan: evaluation first, then hierarchical diagnosis and an expected step.
5. Engine: close windows on the watermark; a cumulative marker for KPIs.
6. Then the Retrofit Designer.
