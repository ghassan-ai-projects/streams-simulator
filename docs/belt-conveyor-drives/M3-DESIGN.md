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
