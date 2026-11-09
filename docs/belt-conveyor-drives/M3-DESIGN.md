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

## 4. As built, row 1 (O24, streamsim)

- Action fidelity matches a verdict action to the world's call by command id, effector and entity (and belief against
  acceptance); `issued_at` must parse but may precede the call. A plant-wire host reports `'sha256:' + hex(key)`, which is
  the DCMD's `idempotency_key`, the id the node logs the call under (`plantwire` `execute` passes `c.Key`).
- Domain-spec v0.2: a fault may list `expected_actions: [{kind, effector}]`; the domain checks the kind is declared and
  declares the effector (`/expected_actions/<i>/kind|effector`). The label resolves them when it is built
  (`truth.BuildRecord`): the fault's own kind on the faulted instance, another kind on every instance of that kind in the
  run's population; ground-truth v0.1 carries `expected_actions: [{entity_id, effector}]`. Both members are omitempty, so
  every domain and label without them is byte-identical (behaviour pin unchanged).
- Composition v0.1: an equipment entry may give `expected_actions` keyed by its component's fault id, each
  `{kind, effector}` with the composed kind name and the component's effector name. Compose writes them onto the qualified
  fault with the effector qualified (`bulk_feeder__set_feed`) and refuses a fault the equipment does not declare. They
  belong to the composition, not the component, because a component does not know the other kinds of a plant (and the
  kind names can be changed with `as`).
- Scoring: a right call is the expected effector on the label's entity or one of the label's expected actions (with
  effect); it counts for `action_appropriate` and the deadline. With several labels a scorecard's judgment sees only its
  entity, its loop its entity and its expected instances; `unexpected` holds detections on unlabeled entities and
  actions and calls on entities neither labeled nor expected.
- `grading.Grade` (one label only) and `ErrNotOneLabel` are gone; the report grades with `GradeRun` whenever `label.json`
  or `labels.json` is present and stores `grading.Document` (the scorecard for one label, the run document otherwise);
  the Markdown shows each scorecard and the unexpected counts.
- Conveyor compositions: `lube_loss` expects `set_feed` on the feeder and `stop_drive` on the drive, `gear_oil_loss` the
  same plus `stop_belt` on the belt, `vfd_fault` `set_feed` on the feeder; composed domains regenerated (only these
  members change; the world, the stories and the outcome numbers are untouched).
- Contract test `test/contract/belt_score_test.go`: plant B, D-3's inverter trips at 60 s, a host decides a feed cut at
  65 s that the world applies at 70 s under the DCMD key; the label expects `bulk_feeder__set_feed` on F-1, and
  `GradeRun` gives action fidelity, the right action, D-3 detected with the right label and nothing unexpected.

## 5. As built, row 2 (O21, streamsim and agentic-stream)

- No conveyor command carries an id argument: the call (or the plant wire device) addresses the instance. Nothing generic
  needs one: the only `_id` rules are a baseline policy refusing an entity argument its target supplies
  (`validation_policies`, kept with a test-made `feeder_id`) and the reference consumer's and test support's guesses.
- `start_drive`/`start_belt` take `run_command` (number, minimum = maximum = 1), `stop_drive`/`stop_belt` take
  `run_command` 0, and each assigns `run_cmd` from it (`assign_from_arg`). The `requires` permits are unchanged, and under
  them the assignment equals the former ±1 delta, so the world is unchanged (stories and outcomes green); a value of the
  other verb is an invalid invocation. `reset_trip` takes no argument.
- Read-back: drive and belt gain the point `run_command` on `run_cmd` with a new instrument
  `streamsim/command-readback` (state family, unit `on`, 0..1, number, resolution 0.1, periodic 2 s, no noise, `dropout`).
  The setpoint read-back is a `%` of the fraction family and cannot take a 0/1 state; the run-status contact reports what
  the drive does (`enabled`, report by exception) and a boolean cannot be verified within a resolution (a host's
  tolerance is the metric's resolution, 1 for a boolean). `run_status` stays as it was.
- Kind files: `conveyor_drive` (`current`, `winding`, `vibration`, `run_status`, `run_command` writable; `start_drive`,
  `stop_drive`, `reset_trip`) and `belt_conveyor` (`tail_speed`, `run_command` writable; `start_belt`, `stop_belt`) are
  semantic files mapped to the raw qualified names by `drive-controller.board.json` and `belt-controller.board.json`; the
  compositions name the boards; `feeder-controller` loses its fixed `F-1`. The component module's embedded example kinds
  carry six files.
- Regenerated: compositions and embedded copies, composed domains, kind files and embedded copies, story scenario files
  (sessions hold no commands and did not move). `stories.json` stop calls carry `run_command` 0.
- Tests: a 180 s plant-B session against the reference host: every kind is born equal to its file; `set_feed`, belt stop
  and start, and D-3 stop and start (once the belt ran empty) are each verified by read-back, and D-3's stop on the loaded
  belt is refused by the plant permit (`permissive_not_met`); a world test refuses a run command of the other verb.
- agentic-stream: kind copies updated, the site's drive readings name `current`, the fake node births `current` and
  `tail_speed`.

## 6. As built, row 3 (U7, streamsim)

- F1 form `ratio` (algebraic, no `dt_ms`): `x = Π_num cᵢ·sᵢ / max(Π_den cⱼ·sⱼ, floor)`; an input with `denominator: true` is a
  factor of the denominator; 1 to 4 factors on each side; `floor` > 0, so a denominator at or below it (zero or negative)
  divides by the floor. Model (`FormRatio`, `F1Dyn.Floor`, `F1Input.Denominator`), domain validation
  (`/inputs`, `/floor`, `/inputs/i/denominator` outside a ratio), world (`ratio` in `algebraic.go`), domain-spec v0.2 schema
  and copy, TECHNICAL_DESIGN §6.2 and DOMAIN_SPEC_V0_2 §23.
- Oracle (`internal/world/internal/domain/ratio_oracle_test.go`): 60 random plants of 1-5 drives, weights in [0.5, 1.5],
  random enabled subsets, against the plain closed form `wᵢ·eᵢ·D / Σ w·e` (1e-12 relative); the shares add to the demand;
  with nothing enabled every share is 0 (a zero numerator over the floor); the floor below and above the denominator; equal
  weights give a factor of exactly 1 for 1 to 6 drives.
- Conveyor: each drive has `weight` (initial 1) and provides `weighted = weight·enabled`; the belt sums them like the
  running count (sampled 1 s, new port type `share_weight`, fraction) and hands `drives` and `weights` back; each drive
  computes `share_factor = ratio(weight·drives_seen / weights_seen)` and its load is `share·share_factor·enabled/F_rated`,
  where `share` is the belt's equal share as before. The instance fault `conveyor_drive__share_skew` (step, negative class,
  `weight += severity`) skews one drive from t = 0; seed rows use it later.
- Deviation from "belt provides demand, each drive computes w·demand/Σw": the belt's `inv_n` is a 7-digit table
  (0.3333333), so `w·demand/Σw` would differ from today's runs in the last digits. Scaling today's equal share by
  `w·n/Σw`, which is exactly 1 when the weights are equal (n/n), keeps every equal-weight run bit for bit and still gives
  `w·demand/Σw` to the table's precision. Guard (long tier, `belt_equal_share_test.go`): the trace digest and the as-run
  outcome digest of S1-B, S2-B, S3-A references and S4-A do-nothing, taken before the change, are unchanged.
- Plant tests: D-3 skewed by 0.2 carries 1.2/3.2 of the demand and D-1, D-2 1/3.2 each, together the whole demand;
  stopping the skewed D-3 leaves the survivors equal shares.
- Stop permit under unequal weights (known limit, stated not fixed): the belt's permit checks the survivors' total
  rating and restart capacity (`(n−1)/n` of the running totals), so it assumes they share equally. Stopping the skewed drive
  is exact (the survivors' weights are equal). Stopping another drive leaves the skewed survivor `(1+s)/(2+s)` of the demand
  instead of 1/2: 9 % more than the permit assumes for s = 0.2. A conservative permit would scale the survivors' rating by
  `Σ_survivors w / ((n−1)·w_max)`; not needed while seed rows skew only the drive that the plan stops (D-3).
- Open point (coordinator, 2026-11-09): `share_skew` is not added to the committed story seed rows in M3. A negative-class
  label on D-3 beside D-3's real fault label would put two labels on one entity, which scoring cannot yet separate (its
  judgment and loop scope by entity). Skewed rows wait until scoring can tell two labels on one entity apart.

## 7. Calibration, row 4 (streamsim, equal weights)

Method (`test/contract/belt_calibration_test.go`, `belt_calibration_table_test.go`, long tier): every seed row of S1-A,
S1-B, S3-A (do nothing), S0-A and S4-A/B, run to the horizon with noise, read as a host receives the readings (every
drive's winding every 10 s and vibration every 60 s, with their noise, quantization and sensor outages: availability
0.999, 10 min repairs). Per drive, at the end of every minute: the 15 min mean of vibration and of winding and the
least-squares winding slope over 30 min (K/h); a window counts only when it holds at least 90 % of its readings (81 of
90 winding, 14 of 15 vibration, 162 of 180 for the slope). A residual is the drive's value minus the largest of its two
peers'. Lead = D-3's do-nothing trip (S1: the first protective trip, overtemperature; S3: the seizure, `free_ok`) minus
the first minute the D-3 residual reaches the threshold; "early" counts rows that crossed before the fault's onset (none).
The noiseless runs gave the same leads within 0.1 h (measured once, not kept in the test).

D-3 residual leads, hours before the do-nothing trip (min / median / max over 10 seed rows):

| Residual | Threshold | S1-A | S1-B | S3-A (to seizure) |
|---|---|---|---|---|
| vibration 15 min mean | 0.05 mm/s | 4.54 / 4.70 / 5.51 | 4.33 / 4.89 / 5.48 | never |
| | 0.1 mm/s | 4.25 / 4.30 / 5.03 | 3.92 / 4.37 / 4.99 | never |
| | 0.2 mm/s | 3.85 / 3.91 / 4.56 | 3.49 / 3.81 / 4.40 | never |
| | 0.3 mm/s | 3.60 / 3.63 / 4.21 | 3.17 / 3.44 / 3.95 | never |
| | 0.5 mm/s | 3.23 / 3.25 / 3.81 | 2.67 / 2.96 / 3.33 | never |
| | 1 mm/s | 2.64 / 2.64 / 3.09 | 1.96 / 2.14 / 2.42 | never |
| winding 15 min mean | 0.5 K | 4.95 / 4.96 / 5.86 | 4.84 / 5.36 / 6.10 | 1.31 / 1.54 / 1.81 |
| | 1 K | 4.61 / 4.63 / 5.46 | 4.44 / 4.89 / 5.58 | 1.21 / 1.44 / 1.70 |
| | 2 K | 4.20 / 4.21 / 4.96 | 3.92 / 4.32 / 4.94 | 1.06 / 1.28 / 1.51 |
| | 3 K | 3.91 / 3.92 / 4.62 | 3.57 / 3.94 / 4.49 | 0.92 / 1.13 / 1.34 |
| | 5 K | 3.48 / 3.49 / 4.12 | 3.04 / 3.36 / 3.84 | 0.67 / 0.84 / 1.03 |
| winding slope 30 min | 0.5 K/h | 5.22 / 5.29 / 6.14 | 5.12 / 5.51 / 6.42 | 1.44 / 1.69 / 1.96 |
| | 1 K/h | 4.87 / 4.92 / 5.72 | 4.66 / 5.09 / 5.64 | 1.39 / 1.63 / 1.90 |
| | 2 K/h | 4.41 / 4.43 / 5.08 | 4.01 / 4.34 / 4.81 | 1.31 / 1.56 / 1.81 |
| | 4 K/h | 3.74 / 3.76 / 4.22 | 3.09 / 3.26 / 3.48 | 1.19 / 1.41 / 1.64 |

Onset / do-nothing trip, hours, per seed row (sorted): S1-A 1.14/8.53 1.14/9.31 1.23/8.19 1.32/8.28 1.47/8.46 1.50/8.55
1.57/8.56 1.66/8.61 1.68/8.69 1.97/8.96; S1-B 0.82/9.68 0.87/8.01 0.96/8.11 1.05/8.49 1.09/8.97 1.32/9.80 1.37/9.18
1.69/10.35 1.73/10.59 1.74/9.39; S3-A (seizure) 0.50/2.23 0.55/2.71 0.71/2.59 0.80/2.58 0.94/2.56 1.15/2.95 1.21/3.00
1.39/3.55 1.43/3.35 1.47/3.43.

Largest residual of any drive against its peers over every seed row (noise, quantization, outages):

| Story | Vibration mm/s | Winding K | Slope K/h |
|---|---|---|---|
| S0-A (shift pattern, 24 h) | 0.023 | 0.038 | 0.832 |
| S4-A (one surge) | 0.024 | 0.023 | 0.113 |
| S4-B (three surges) | 0.020 | 0.056 | 0.132 |

Without the 90 % coverage rule an outage during a load change biases one drive's view: S0's slope residual then reached
6.9 K/h and S4-B's winding residual 0.32 K. A spec must skip a window that misses readings.

Recommended starting thresholds (guarded by the long-tier test: the S1 leads ≥ 2 h, S3 ≥ 0.5 h, no row early, S0/S4
below half of each threshold):

| Use | Rule | S1 least lead | S3 least lead | S0/S4 margin |
|---|---|---|---|---|
| watch (R0) | winding residual ≥ 1 K or vibration residual ≥ 0.1 mm/s | 4.4 h | 1.2 h | 18× / 4× |
| bearing (S1) | vibration residual ≥ 0.3 mm/s | 3.2 h | never | 12× |
| degrading | winding residual ≥ 2 K | 3.9 h | 1.1 h | 36× |
| gearbox (S3) | winding slope residual ≥ 2 K/h while the vibration residual is < 0.05 mm/s (same minute) | never fires (0 of 20 rows) | 1.3 h (1.31-1.81) | 2.4× |

In S1 the winding residual crosses 1 K about 20 min before the vibration residual crosses 0.1 mm/s, so a label taken at
the first winding crossing would say gearbox: label at the vibration crossing (bearing) or with the gearbox rule above,
which in S1 never holds (the vibration residual is past 0.05 mm/s before the slope residual reaches 2 K/h in every row)
and in S3 holds 1.31-1.81 h before the seizure. The slope has the smallest margin on S0 (0.83 of 2 K/h). The test also
asserts the gearbox rule.

## 8. As built, rows 5-7 (agentic-stream) and exit evidence (2026-11-09)

agentic-stream commits on `belt-m3-predict`: `abdc2ae6` (kind copies and site, O21), `22f6f78c` (data, the
`drive-degradation` spec, fake-node tests, the barrier-paced twin), `699a89e9` (seed rows from `stories.json`, plant B).

- Data: per-drive winding and vibration and the belt's tail speed per device; routes `reduce_feed` (R1, presets `limp_a`
  54, `limp_b` 18, `restore`), `stop_feed` (R2), `isolate_drive` (R2, D-3), `stop_belt` (R2); a test approver.
- Spec: `on_close` windows of 1, 15 and 30 min with coverage guards; residuals against the larger peer; phases normal →
  watch → degrading → shedding → belt_empty → isolated or belt_stopped; one trigger per action phase; ceiling R2. The
  bearing/gearbox split is the reasoner's (the calibrated rules are in the diagnosis catalog); in the tests a scripted
  stand-in plays the reasoner.
- Found on the way: a barrier-paced node reads a command only at its next 60 s slice, so the host's 20 s route expiry
  refused every command; conveyor routes expire after 180 s and serve runs with a 5 min dispatch lease (the host times
  expiry against node time; a product-level fix is open). `export-run` needs `--device-id` when a run commanded several
  devices (the artifact manifest holds one device; open).

Exit evidence, barrier-paced twin (`--slice-s 60`, story `horizon_s`), seed rows 1-3 of each story, 12/12 pass: routes
in order, every command read back, R2 only on settled features and approved, no drive trip, never stranded, zero
quarantines, `verify-run` passes.

| Story | Isolation or belt stop, h after onset | J − B0 | J − B4 | Score |
|---|---|---|---|---|
| S1-A | 2.93, 3.36, 3.08 (do-nothing trips 8.2-8.6 h) | −4.17, −3.74, −4.10 | −5.07, −4.49, −4.97 | D-3 `lube_loss` detected (latency 0.38-0.87 h), action appropriate, nothing unexpected |
| S1-B | 4.02, 4.13, 3.74 | −2.96, −2.82, −3.61 | −2.95, −2.81, −3.61 | as S1-A (latency 0.79-1.16 h) |
| S3-A | belt stopped 0.84, 0.77, 0.83 (before the seizure) | −3.06, −3.38, −3.06 | same | `gear_oil_loss` labelled gearbox, action appropriate |
| S4-A | no action (1, 1 and 3 surges) | - | - | no detection, no action |

Decision on A5 (PLAN.md section 7): not needed for M3; S1's lead holds without a time-to-threshold feature (O24); A5
is built in M4 for U9 (O31).

Wall time per row: about 10.5 min for a 16 h S1 row (session 4.3 min, `streamsim outcome` 3.3 min, `score` 0.8 min,
the rest wind-down), about 2.5 min for S3 and S4. The 63 min of the first run came from machine load, not the loop.
