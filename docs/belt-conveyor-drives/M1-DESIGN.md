# M1 design: outcome truth and grading

Status: design for M1 (2026-11-08), from [PLAN.md](PLAN.md) section 4.1 (U2-U5), section 6 and the owner decisions O15
and O16 (section 9.6). It places every piece in an existing module; no new package, no new import edge between modules.

## 1. What M1 answers

For any finished run of a domain that declares outcome data: what happened (trips, boundary violations, KPI changes),
what would have happened with nobody acting (B0), what a fixed policy would have done (B4), and a cost J for each, so a
controller is graded on consequences, not only on detections.

## 2. Pieces and where they live

| Piece | Module / layer | What |
|---|---|---|
| Records | `model` | `OutcomeFacts` (one run's outcome), `OutcomeWeights` and `BaselinePolicy` (domain data under `ground_truth`) |
| Domain data | `domain` (schema, validation) | `ground_truth.outcome_weights` and `ground_truth.baseline_policies` in domain-spec v0.2 (optional members), checked against the domain's kinds, states, boundaries, KPIs and effectors |
| Outcome recorder | `world` domain + facade | Exact, while stepping: per entity and trip state the first trip time and the number of trips; per boundary the violations, first time and total duration; per KPI state its start and end value |
| Outcome runs | `run` app + facade | `ReplayOutcome(artifact, mode)`: the run's command log replayed with the run's own world settings; mode `as_run` (unchanged), `do_nothing` (every `effector.invoke` removed; O15) or `policy:<name>` (removed, and the policy applied; O16). Returns `OutcomeFacts` and, for a policy, the calls it made |
| Outcome score | `score` domain + facade | Compares as-run with do-nothing and each policy: per-term deltas, cost J of each, `beats` and `valid` (do nothing must not be clean) |
| Use case | `grading` app + facade | `GradeOutcome(runDir, dirs)`: loads the artifact and domain, runs the three outcome replays, scores. Needs no verdict and no label, so scripted runs are gradable |
| Surface | `cli` | `streamsim outcome --run <dir>`: one facade call, the `outcome` document (cli-output contract def), failure mapping |
| Report | `report` | An outcome section when the domain declares outcome data (later round) |
| Session horizon (U5) | `model`, `plantwire` | Optional `horizon_s` in session-plan v0.1: the session ends itself at that world time from the world start |

Nothing is sealed or stored: the outcome is recomputed from the artifact whenever it is asked for, so it is reproducible
by construction and never visible while a consumer runs.

## 3. Records

`OutcomeFacts`: `horizon_ns`, and per entity (sorted): `trips` `[{state, first_ns, count}]` for every trip state that
tripped at least once; `boundaries` `[{name, severity, violations, first_ns, duration_s}]` for every violated boundary;
`kpis` `[{state, start, end}]`; plus `calls` (effector, entity, at, accepted) for the run the facts describe.

`OutcomeWeights` (closed, generic; every name is checked against the domain):

| Term | Shape | Cost |
|---|---|---|
| `boundaries` | `{kind, name, per_violation, per_hour}` | per violation and per hour of violation |
| `trips` | `{kind, state, per_trip}` | per trip of a trip state |
| `kpis` | `{kind, state, per_unit}` | per unit of (end − start), as-run minus do-nothing; a negative weight rewards more (tonnes delivered) |
| `calls` | `{effector, per_call}` | per accepted call, only in runs whose do-nothing outcome is clean (no boundary violation, no trip): an unneeded action |

J = Σ terms, in the domain's own unit (for the conveyor: equivalent production hours). Departure from PLAN.md section
6.3: cascade and unplanned trips are both "a trip of a protection state" with one weight per state, and a stranded belt
carries the cascade's consequence; futile restarts are not a term (they would need a call-followed-by-trip rule).

`BaselinePolicy`: `{name, when: {kind, state, operator, threshold}, for_s, then: [{effector, target, args, delay_s}]}`.
`target` is `matched` (the instance whose condition fired) or an instance name. The policy fires on each instance's
rising edge of the condition held for `for_s` (default 0), and each firing issues the call list in order at its delays.
Calls refused by an effector's `requires` are recorded as refused. Nothing else: no and/or, no expressions, no chaining.

Conveyor data (both plants): B4 = `{when: {conveyor_drive, available, lt, 0.5}, then: [{set_feed, F-1, {percent: 67}}]}`;
weights from PLAN.md section 6.3 mapped onto the terms above (stranded 4 per violation; protection trips 0.5;
`life_used` 16/20000 per h; `gear_used` 16/50000 per h; `delivered` −1/150 per t; `set_feed`, `stop_drive`, `stop_belt`
0.25 per unneeded call).

## 4. Rounds (one commit each, tests at the layer they prove)

1. Records, domain-spec members and their validation; conveyor plants gain B4 and weights (recomposed); session-plan
   `horizon_s` with plantwire ending the session itself; the story session plans carry it.
2. World outcome recorder with an independent oracle (trip and boundary times against closed forms, as M0's tests do).
3. `run.ReplayOutcome` (three modes) with the oracle "a hand-written command log equals the policy run".
4. `score` outcome and J; `grading.GradeOutcome`; `streamsim outcome`; cli-output contract and docs.
5. Exit evidence on the stories (long tier): every row, the reference policy beats B0 and B4 on S1-S3 and ties B0 on
   S0 and S4; B0 is invalid-clean on no positive story; replay determinism of the outcome; then the report section.

## 5. Exit (PLAN.md section 7, as changed by O15)

Outcome computed for scripted and plant-wire runs; grading the reference-policy runs beats B0 and B4 on S1-S3 and ties
B0 on S0 and S4; B0 fails every positive story; the outcome is identical across replays; `make phase-gate`; fan re-check.

## 6. Known limits

- Outcome runs cost one world run each at scoring time (30-60 s for a 16 h conveyor story).
- Director runs and the canvas get the outcome later, through the same facade, when a surface needs it.
- B1-B3 and prognostic metrics (U9) are M4.

## 7. As built (2026-11-08)

Five rounds on `belt-m1-truth`: `10dbd423` (session `horizon_s`), `b651ad17` (outcome weights and baseline policies),
`c0b300ec` (world outcome recorder), `04c151f8` (`run.ReplayOutcome`), `0b4c2c82` (`score` outcome, `grading.GradeOutcome`,
`streamsim outcome`), then the exit evidence and the report section.

What differs from sections 2-3:

- `horizon_s` counts from the world start (in a wall-paced session that is the session start, while disturbances count
  from the host's attach) and is refused with a `gateway_socket`: a session with real boards still runs until stopped.
- `OutcomeFacts` carries each entity's `kind` (weights are per kind); `horizon_ns` is the absolute world time the facts
  are taken at, the run's last world time.
- A policy samples its condition once a second from the world start on every instance of its kind. A condition already
  true at the first sample fires; a fault injected at an instant is seen at the next sample (the sample at that instant
  precedes the injection). The call's entity argument (a required string argument named `*_id`) is filled with the
  target; command ids are `policy-<name>-<n>`.
- Terms: boundary violations and boundary hours are two terms; protection trips are `relay_ok`, `inverter_ok`,
  `overtemp_ok`, `free_ok` (seizure) and the belt's `stall_ok` (it stands in for futile restarts), 0.5 each; `tie` is a
  member of `outcome_weights` (0.1 on both plants).
- `grading.GradeOutcome` first replays the run with emission and refuses one that does not reproduce or is unblinded, as
  `score` does; the outcome replays themselves run silent. The command is `streamsim outcome --run <dir>`.
- The run report gains an `outcome_score` member (status, reason, result = the `outcome` document), opt-in with
  `streamsim report --outcome` (`report.Request.Outcome`) and then present only when the artifact's embedded domain
  declares outcome weights; a default report is unchanged and replays nothing.
- The story long tier's do-nothing runs now run to the horizon (the cut 120 s after stranding is gone) so that their
  outcome covers the same window as the reference run.

Exit evidence:

- Every seed row of every story (`test/contract`, long tier, the tier's own noiseless runs with the recorder on): the
  reference policy beats doing nothing on S1-A, S1-B, S2-A, S2-B, S3-A and S5-A in all 10 rows; doing nothing is valid
  (not clean) on every positive story and clean on S0-A, S4-A, S4-B, where the run ties.

  | Story | Reference J | Do nothing J | Delta |
  |---|---|---|---|
  | S1-A | -11.09 .. -10.70 | -8.43 .. -6.02 | -4.77 .. -2.52 |
  | S1-B (M4 reference) | -10.16 .. -8.05 | -4.59 .. -2.01 | -6.03 .. -5.57 |
  | S2-A | -1.64 .. -1.48 | 3.48 .. 4.02 | -5.51 .. -5.05 |
  | S2-B | -0.84 .. -0.59 | 4.99 .. 5.29 | -5.88 .. -5.83 |
  | S3-A | -1.98 .. -1.01 | 1.45 .. 2.77 | -3.90 .. -3.35 |
  | S5-A | -11.07 .. -10.64 | -8.07 .. -5.71 | -4.95 .. -2.72 |
  | S0-A, S4-A, S4-B | - | -20.40; -6.12 .. -6.03; -6.11 .. -6.02 | 0 (ties) |

  S1-B's reference changed in M4 (M4-DESIGN section 6): it keeps full feed and stops D-3 1.5 h before the do-nothing
  trip; the M1 reference (18 % limp from 4 h after the onset) had J -7.47 .. -6.71, delta -4.74 .. -2.87.

- Against B4, row 1 of every story through the binary (`streamsim run --scenario`, then `streamsim outcome`,
  `test/acceptance`): the reference beats B4 on S1-A (-5.67), S1-B (-3.07; the M4 reference -5.60), S2-A (-5.52), S2-B (-5.85), S3-A (-3.35) and
  S5-A (-4.89); on S0-A, S4-A and S4-B B4 never fires and everything ties.
- S2 costs (row 1, equivalent production hours):

  | Story | Reference | B0 | B4 |
  |---|---|---|---|
  | S2-A | -1.54 | 3.92 | 3.98 |
  | S2-B | -0.69 | 5.16 | 5.16 |

- Determinism: two gradings of one run folder give byte-identical documents; `as_run` equals a direct world run of the
  same commands; a hand-written command log equals the B4 policy run on S2-B.
- Plant-wire: a simulated session that ended at its horizon is graded from its run folder.

Timings (10 cores): outcome replays of a 16 h story take 26 s (`as_run`), 22 s (`do_nothing`) and 25 s (`b4`); `streamsim
outcome` on it about 100 s (one emission replay plus three silent ones); the story long tier 258 s (205 s before the cut was
removed); `test/contract` without `-short` 280 s on 10 cores and 613 s with `GOMAXPROCS=4 -p 1` (over go's default
10 minute test timeout at that cap); `test/acceptance` without `-short` 271 s on 10 cores, 379 s with `GOMAXPROCS=4`.

Known limits found on the way: B4 is defined only on drive unavailability, so on plant A it lowers throughput and still
strands the belt (it costs about as much as doing nothing); drive wear (`life_used`, `gear_used`) weighs under 0.002 per
run against 4 for a stranded belt; a policy samples retired entities too (the conveyor has none); `make phase-gate` and the
fan re-check are run by the coordinator, not by the round.
