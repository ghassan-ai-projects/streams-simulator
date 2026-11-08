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
