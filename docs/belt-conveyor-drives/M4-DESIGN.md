# M4 design: hard cases and evidence

Status: design for M4 (2026-11-09), from [PLAN.md](PLAN.md) section 7 (M4), sections 2.2-2.3 (B1-B3, S5, S6), 4.1-4.3
(U8, U9, A3, A5, T1-T3, T6), 6.2 (prognostic metrics) and decisions O25-O32 (section 9.6), taken while the owner was
away. Branch `belt-m4-evidence` in streams-simulator, agentic-stream and tamoz, each cut from its M3 tip (tamoz from
`belt-conveyor-drives`, it had no M3 branch).

## 1. Target

Every story's pass rule holds over 10 seed rows, against the baselines B0-B4, with the loop closed through
agentic-stream; the hard cases are in: a frozen sensor that hides S1 (S5), a dropped-out sensor that must not deny the
line's R2 steps (S5b, A3), and the loss of the torque master (S6). Predictions of the failure time are graded (U9). The
S2 reaction becomes a playbook rule that a human approves. Arms that need the live model (T6, the FR+ fan rung) are
built and wait for the owner.

## 2. Work, in order

| # | Repo | Work |
|---|---|---|
| 0 | all | Throughput: met without change; a 16 h story on the barrier-paced twin takes about 10.5 min unloaded (M3-DESIGN §8; the first 63 min came from machine load) |
| 1 | streamsim | B1-B3 as closed baseline records (O27) on both plants; the outcome document says a story is `easy` when a baseline ties or beats the reference; story tables checked |
| 2 | streamsim | Two labels on one entity (O25): scoring keys a label by entity and fault; one label per entity scores bit for bit as before |
| 3 | agentic-stream | A6 per-(entity, diagnosis) detections in the verdict writer; spec `sensor_stuck` (flat D-3 vibration while the peers move, from existing aggregates); S5-A on the twin, 3 seeds |
| 4 | agentic-stream | A3: input-level `completenessRole: required | advisory`; an advisory input's loss is the fact `input_stale.<name>`, not line uncertainty; existing spec and policy digests unchanged (pinned); S5b story row in streamsim (dropout instead of freeze); S5b on the twin |
| 5 | streamsim | S6: drive `role` (`master`, `follower`) and `master_loss` (`stop_all`, `promote`) as component data with an independent oracle; two drive kinds in the compositions; kind files and boards; S6 story rows (both parameters) |
| 6 | agentic-stream | S6 spec phase `master_lost`: no restart while loaded; an advisory unload request (R0) and an empty-restart plan; `promote` runs S2-A's reaction |
| 7 | streamsim + agentic-stream | U9 (O31): consumer-verdict v0.2 optional `predicted_failure_at`; score computes α-λ (α 0.2, β 0.5, λ 0.5 and 0.75), relative accuracy and prognostic horizon, truth from the do-nothing outcome; A5 `extrapolate` operator publishes a time to threshold; the verdict writer passes it |
| 8 | streamsim | U8: `severity_range` per fault weight in suite campaign phases, fixed draw order; the committed story tables stay |
| 9 | tamoz + agentic-stream | T1 conveyor domain fixture (codes unified with the specs: `bearing`, `gearbox_seizing`, `vfd_trip`, `feed_surge`, `sensor_stuck`, `idler_drag`); T2 verified (already built in tamoz `c4a758f1`), plus the allowed intents in the repair message; T3 (O29) |
| 10 | agentic-stream | Playbook rule for the S2 reaction (O30): `promotion` section in `line-overload`, learning runs with the scripted approver through the signed relay, a pending proposal left for the owner's live approval |
| 11 | all | 10-seed acceptance: every story through the twin, background, capped; the A arm (rules only) report against B0-B4; the T6 runner for arms B and C built and dry-run without a model |

## 3. Rules

As in M3: each repository's AGENTS.md, generic code only (no conveyor names in any binary), no gate loosening, long
tests per package with `-p 1` and `GOMAXPROCS` at most 4, at most two long test processes at once, nothing left running;
the fan experiment guards stay green; new world behaviour comes with an independent oracle; one commit per improvement.

## 4. Decisions (PLAN.md section 9.6)

- O25 Two labels on one entity: a label is keyed by entity and fault. A detection on an entity that carries several
  labels is matched to the label its diagnosis maps to (the verdict's subject map), so S5's `sensor_stuck` and `bearing`
  on D-3 are judged separately; with one label per entity nothing changes. This also unblocks `share_skew` in seed rows,
  which stays out of the committed tables (it would move every outcome digest).
- O26 A frozen sensor's lead is counted from its injection: a freeze is unobservable by design, so it has no
  `first_observable`; S5's order rule (stuck reported before the bearing) needs none.
  Amended 2026-11-10 (consumer run of S5-A): a real session's label took the freeze's first observable from the
  single-channel detector, when the wiggle had moved the true reading 3σ off the frozen one (row 1: 1.62 h, after the
  bearing onset), so a report from the flatness 15 min after the freeze was graded too early. Truth now counts a held
  reading (`stuck` or `freeze`) of a channel with noise as first observable at its injection, as a dropout is: the
  noise stops at once. The lead is counted from the injection as decided, and the score's too-early rule is unchanged.
- O27 Baselines: B1 = vibration > 4.5 mm/s for 60 s, then feed 0 and stop the matched drive 60 s later. B2 departs from
  the plan: the closed record has no rolling mean, so B2 is an absolute current threshold at 1.1 times the
  nominal-feed drive current, held 60 s, then feed 67 %. B3 = a condition true from the first sample (feeder setpoint
  ≥ 0), then the plant's safe limp feed (54 % A, 18 % B, from M0 rather than the plan's 55/30).
- O28 Arms: A = rules only (the specs and the scripted stand-in, no model), B = Tamoz (live model reasons, specs act),
  C = both (an approved playbook rule plus Tamoz). B and C need the live model, its cost and the owner.
- O29 T3 is fixed on the agentic-stream side: a retryable failed terminal (`checkpoint_conflict`) counts as an attempt
  and the next dispatch waits at least the worker's lease; the terminal's failure reason is stored and shown. Tamoz
  changes only if it sends no reason.
- O30 The playbook rule is drafted from verified S2-B runs approved by the scripted approver (labelled as such); the
  proposal then waits for one live approval by the owner (a human approves every rule).
- O31 U9's consumer side uses A5 (`extrapolate`: (limit − level)/slope over a declared window with a validity flag), a
  new operator kind, designed in agentic-stream's design notes before it is written.
- O32 Throughput before evidence: met (about 10.5 min per 16 h story); the 10-seed evidence needs no speed-up.

## 5. Exit

PLAN.md section 7: every story's pass rule over 10 seeds; the A arm's report against the baselines; `easy` flags
explain any story a baseline passes; `verify-run` and zero quarantines on every run; `make phase-gate` in streamsim,
`make ci-check` in agentic-stream and tamoz's own gate; the fan guards. Left for the owner: the live approval of the
playbook proposal, arms B and C (T6), and the FR+ fan rung. Then the full end-to-end test of all seven stories (O13).

## 6. As built

Rows 1 and 2 (2026-11-09):

- B1-B3 are closed `baseline_policies` on both plants beside B4. B2's threshold is 1.1 times the drive current at full
  feed on three drives, from the model in closed form: 3.0768 A on plant A and 4.2614 A on plant B, written as 3.385 A
  and 4.688 A (a contract test checks both against the closed form and the steady world).
- `streamsim outcome` gains `easy`: valid, and some baseline policy ties or beats the run. Doing nothing never makes a
  story easy.
- What each baseline does on each story's first row, through the binary (`test/acceptance`), as the delta of the run's
  cost J against it (equivalent production hours; negative: the run beats the baseline; `-`: the baseline never calls):

  | Story | B1 vibration | B2 current | B3 safe limp | B4 drive lost | easy |
  |---|---|---|---|---|---|
  | S0-A | - | - | -9.63 | - | no (not valid) |
  | S1-A | -4.14 | -5.67 | -2.92 | -5.67 | no |
  | S1-B (row 1, M1 reference) | +2.12 (B1 wins) | -3.07 | -4.63 | -3.07 | yes |
  | S1-B (row 1, M4 reference, below) | -0.42 | -5.61 | -7.17 | -5.60 | no |
  | S2-A | - | -5.51 | -0.19 | -5.52 | no |
  | S2-B | - | -5.85 | -0.24 | -5.85 | no |
  | S3-A | - | - | -4.59 | -3.35 | no |
  | S4-A | - | -1.89 | -3.02 | - | no (not valid) |
  | S4-B | - | -2.30 | -5.24 | - | no (not valid) |
  | S5-A | -4.05 | -4.88 | -2.98 | -4.89 | no |

  Departures from PLAN.md section 2.3: B1 is not late on S1-B (it alarms 47 min before D-3 would trip, empties the belt
  and stops D-3 in time, and beat the M1 reference policy, which limped at 18 % from 4 h after the onset; the M4
  reference below beats it on every row); B1 fires on S5-A, because a policy watches the vibration state, not
  the frozen reading; B2 does not fire on S0 (the feed never passes about 1.33x) nor on S3-A (the currents pass 1.1x only
  in the seconds before the seizure); B3 beats nothing but costs least on S2 (0.2 h behind the reference).
- `streamsim outcome` on a 16 h story replays six runs plus the reproduction check: 321 s with `GOMAXPROCS=4` on a
  loaded machine (S1-B).
- A label is keyed by entity and fault (O25): with several labels on one entity a detection is judged with the label its
  `label` names and a call with the labels that expect it; what names none is judged with each. One label per entity
  scores bit for bit as before (digest-pinned run-score documents).

S1-B was easy; why, and the plant-B rule (2026-11-09):

- Per-term deltas of the outcome document, row 1, M1 reference against B1: neither run trips, strands or pays for a call;
  `delivered` −7.451 against −9.569 (+2.118), `life_used` 0.017 against 0.019, `gear_used` 0.010 against 0.009: J +2.12
  is throughput alone, about 318 t (150 t per equivalent hour). The reference empties the belt and stops D-3 at 5.8 h and
  limps at 18 % for the remaining 10.2 h; B1 runs full feed until its alarm at 9.56 h (D-3 would trip at 10.35 h), then
  delivers nothing for 6.4 h. Every hour the isolation comes earlier costs 0.82 equivalent hours of full feed (1 − 0.18);
  every hour of 18 % after B1's alarm gains 0.18. On plant B the limp feed is too small to pay for an early isolation,
  and the agentic-stream loop (18 % limp from about onset + 3.9 h) loses to B1 for the same reason.
- What beats B1: full feed until a planned stop, then feed 0 for 180 s (belt empty), stop D-3, then the safe limp feed
  18 %. With the stop at the do-nothing trip minus a lead L, measured on all ten rows (exploration runs with the outcome
  replays, J delta against B1; beats means below −0.1):

  | L | Rows that beat B1 | Delta against B1 |
  |---|---|---|
  | 2 h (stop at the pass rule's lead) | 8 of 10 | −0.39 .. −0.13; rows 1 and 2 tie (−0.02, +0.01) |
  | 1.5 h | 10 of 10 | −0.80 .. −0.40 |

  So no plan of this shape beats B1 on every row with the stop 2 h ahead. PLAN.md section 2.3 asks the 2 h lead of the
  bearing *detection* and only "D-3 stopped by command before t_f (planned, not tripped)" of the stop, so the new
  reference stops 1.5 h ahead: `planned_stop_lead_s` 5400 in the S1-B entry of `stories.json`, and each row's
  `planned_stop_s` is the do-nothing trip less 1.5 h, rounded down to 5 min (the reference knows the trip as a perfect
  prognosis would; the long tier checks the lead on every row). Row seeds and draws are unchanged; only
  `s1-b.reference.scenario.json` is regenerated (session plans carry no policy).
- New S1-B reference, every row (equivalent hours; delta of the reference against each baseline; all `beats`, never
  `easy`; reference J −10.16 .. −8.05):

  | Row | J | B0 | B1 | B2 | B3 | B4 |
  |---|---|---|---|---|---|---|
  | 1 | −9.96 | −5.61 | −0.42 | −5.61 | −7.17 | −5.60 |
  | 2 | −10.16 | −5.57 | −0.39 | −5.57 | −7.37 | −5.57 |
  | 3 | −8.80 | −5.82 | −0.61 | −5.82 | −6.01 | −5.82 |
  | 4 | −9.41 | −5.73 | −0.58 | −5.73 | −6.62 | −5.72 |
  | 5 | −9.48 | −5.68 | −0.49 | −5.68 | −6.69 | −5.67 |
  | 6 | −8.12 | −6.00 | −0.76 | −6.00 | −5.32 | −6.00 |
  | 7 | −8.39 | −5.89 | −0.66 | −5.89 | −5.60 | −5.89 |
  | 8 | −8.05 | −6.03 | −0.79 | −6.03 | −5.25 | −6.03 |
  | 9 | −9.14 | −5.75 | −0.51 | −5.74 | −6.35 | −5.74 |
  | 10 | −9.00 | −5.83 | −0.61 | −5.82 | −6.21 | −5.82 |

- Rule for the agentic-stream S1 spec on plant B (survivors without margin, limp feed 18 %): do not cut the feed when the
  bearing is detected; detect at least 2 h before the trip (the pass rule) and keep full feed; take the planned isolation
  (feed 0, belt empty, stop D-3, feed 18 %) from the predicted trip time (A5 `extrapolate`, the time published as
  `predicted_failure_at`), with the stop 1.5 h before it, no earlier than the prediction's error needs; never limp at
  18 % before isolating. Each hour of isolation earlier than needed costs 0.82 equivalent hours: a prediction up to about
  0.5 h early still beats B1 on every row (at a 2 h lead two rows tie), and a late one must stay well inside the 1.5 h or
  D-3 trips. Plant A keeps its rule (limp at 54 % costs 0.46 per hour and the S1-A reference beats B1 by 4.14).

Rows 7 (streamsim side) and 8 (2026-11-09):

- U9: a detection's optional `predicted_failure_at` is added to consumer-verdict **v0.1** in place, departing from
  "v0.2": the repository adds optional members without a version bump (session-plan `horizon_s`, domain-spec
  `outcome_weights`), and an unchanged version keeps agentic-stream's verdict writer valid. `streamsim score` replays the
  run with every effector call removed only when the verdict carries a prediction (grading app), and the score domain
  grades predictions on each label's entity: t_f its first protection trip there; graded are predictions at or after
  first observable and before t_f and before the run's first call that took effect (any entity: once anything acts,
  the do-nothing truth is gone, L6). e = w·|p − t_f| / (t_f − t) with w 2 when late; inside ±α when e ≤ 0.2 (a point
  prediction meets β 0.5 exactly when inside); PH = t_f − the earliest prediction from which every later one is inside;
  α-λ and RA = 1 − e at t_P + λ(t_f − t_P) for λ 0.5 and 0.75, evaluating the prediction in force there, ungraded when
  that point is not before the window's end. Output: `judgment.prognosis` (absent without predictions, so pinned
  scorecards are unchanged). Oracles: a hand-computed case in the score domain, and S1-A row 1 with hand-written
  predictions placed against the D-3 trip of a direct world run (do-nothing run: λ points graded; reference run: the
  feed cut at onset + 4 h ends the window before λ 0.5, so a consumer that acts at its detection is graded on PH only).
- U8: a campaign phase's fault weight is a number or `{weight, severity_range}`; severities come from a per-run stream
  `campaign/run/<n>/severity`, so the first-slice campaign is byte for byte as before (digest-pinned).

Row 4, streamsim side: the S5b story rows (2026-11-10):

- No new world or perturbation behaviour: the composed plants already carry the accelerometer's instrument fault
  `conveyor_drive__vibration__dropout` (observation form `dropout`), under which the world schedules no reading of the
  channel while the fault acts (`emitChannelReading`, `processEventCheck`), and truth's single-channel detector already
  treats a dropped reading as an infinite deviation. S5b is data only: a new `stories.json` entry `s5b-a` (story `S5b`).
- Rows: S5-A's ten rows (seed, onset, severity, downstream stop) with `frozen_s` renamed `dropout_s`; disturbances and
  reference policy are S5-A's with the freeze replaced by the dropout. So S5-A and S5b are a pair: a consumer's results
  on the same row differ only by freeze against dropout. Generated: `s5b-a.do-nothing` and `.reference` scenarios and the
  session plan; every other story file is byte for byte unchanged.
- Labels: two on D-3, `conveyor_drive__lube_loss` (the bearing) and `conveyor_drive__vibration__dropout`. The dropout is a
  label, not left unlabelled: every planned disturbance yields a record (`labels.json` of a plant-wire session), and
  marking it negative-class would claim a look-alike, which it is not. O25 keys the two labels by fault, so a consumer
  that reports the lost input is credited on the dropout and its bearing detection is judged on its own; one that does
  not misses only the dropout label. O26 does not apply: unlike the freeze, a dropout is observable, at its injection
  (first observable = injection, checked). The dropout expects no action: its only right response is to keep the R2
  steps open (agentic-stream A3), which the outcome grades through the bearing's consequences.
- S5b story table (equivalent production hours):

  | | S5b-A |
  |---|---|
  | Do nothing | D-3 overtemperature trip 6.99-8.54 h after onset (8.12-9.67 h after the start); the belt strands at the downstream stop; J −8.07 .. −5.71 |
  | Reference (S5-A's: limp 54 % at onset + 4 h, empty, stop D-3, limp) | no survivor trip, never stranded; beats doing nothing on every row by 2.72-4.95 (J −11.07 .. −10.64) |
  | Row 1 against B0-B4 (binary) | B0 −3.65, B1 −4.05, B2 −4.88, B3 −2.98, B4 −4.89: beats all, every baseline fires |
  | Labels | `lube_loss` on D-3 and `vibration__dropout` on D-3 (first observable at injection, no expected action) |
  | Easy | no |

  B1 fires as on S5-A: a baseline policy watches the vibration state, not the delivered reading. The long tier's
  outcome lines (D-3 trip time, J and delta for both policies) are identical row for row to S5-A's, as they must be: a
  dropout changes what the observer gets, not what happens.
- Evidence: the long tier (`TestEverySeedRowOfEveryStoryTellsItsStory/s5b-a`, 20 runs, 320 plant hours) in 3 min 15 s at
  `GOMAXPROCS=4 -p 1`; row 1 through the binary in `test/acceptance` (`run --scenario`, 2.3 min; `outcome`, 5.1 min),
  which also reads the delivered `trace.jsonl`: D-3's vibration 41 readings before the dropout at 2485 s and none after;
  in CI, `test/contract/belt_stories_dropout_test.go` runs the committed scenario to 10 min past the dropout (D-3 none,
  D-1 and D-2 one a minute) in about 5 s, and builds the dropout's label (positive, observable at injection, no action).
- Over the plant wire the dropout is a metric that is no longer refreshed (no `DATA` for it), not a quality code; the
  host's staleness by scan class is what agentic-stream's A3 turns into `input_stale.<name>`.

## 7. S6 design (row 5)

Loss of the torque master, PLAN.md section 2.3 S6, decided as O33 (section 9.6). Only existing F1 forms (`sum`, `trip`,
`product`) and an F0 state; no new form, no code.

**Component.** `streamsim/conveyor-drive@2` is `@1` plus a speed-reference link (a contract test strips the link from
`@2` and finds `@1`). Plants A and B stay on `@1`, byte for byte. Added to the drive:

| State | Form | Meaning |
|---|---|---|
| `ref_seen` | F0 1; port `reference_in` (`run_state`, consumes, `sample_s` 1) | the speed reference from the drive before it in the line; 1 when nothing is connected |
| `ref_in` | sum: master `unit`, follower `ref_seen` | a master ignores any reference; a follower obeys it |
| `ref_held` | trip: `ref_in` < 0.5 for `promote_s`, latched, reset by `reset_trip` | the reference has been missing for `promote_s` (a trip state, so the outcome facts carry its time) |
| `promoted` | sum: `stop_all` 0; `promote` 1 − `ref_held` | the follower closed the speed loop itself |
| `speed_lead` | sum: master 1, follower `promoted` | the drive runs the belt's speed loop |
| `ref_ok` | sum: `stop_all` `ref_in`; `promote` 1 | the drive may run on the reference it has |
| `enabled` | product `run_cmd · available · ref_ok` (was `run_cmd · available`) | |

Port `reference_out` (`run_state`, provides, state `enabled`): a running drive passes the reference on. Parameters, enum
cases writing coefficients only: `role` (`master` default, `follower`), `master_loss` (`stop_all` default, the usual
fieldbus-loss action of an inverter, or `promote`: keep the last torque and let the next drive take the speed loop) and
`promote_s` (2 s, 1-60 s, the trip's hold). A follower that lost its reference stays `available` (its protections and
inverter are fine), so the belt's restart margin still counts it: two healthy followers cannot break the loaded belt
loose, which is why the stopped belt is stranded.

**Line.** D-1 `reference_out` -> D-2 `reference_in`, D-2 `reference_out` -> D-3 `reference_in` (a chain, so D-2 is the
successor). `stop_all`: D-2 stops one sample after D-1, D-3 one after D-2. `promote`: D-2 keeps its torque through the
transfer, so D-3 never loses its reference, and D-2 leads `1 + promote_s` s after D-1. The hold changes nothing the belt
shows (O9: no slowdown under a torque deficit); after it the survivors carry S2-A's share. Known limit: in the chain a
follower's own trip under `stop_all` also stops the drives after it (no story trips a follower; a star fieldbus would not).

**Plants.** `belt-conveyor-drives-s6-stop` and `-s6-promote` are plant A with two drive entries: D-1 (`as: drive_master`,
`role: master`) and D-2, D-3 (`as: drive_follower`, `role: follower`, `master_loss`), and the two reference connections.
Outcome weights name both drive kinds; `b1` and `b2` watch the followers, `b4` the master (the event of the story),
`b3` is plant A's. They are not shipped as built-in examples (like M0's 2-drive belt), so the canvas and its example list
are unchanged. Both lint clean.

**Kind files and boards.** A kind file's `kind` is the domain kind's name (the wire announces it), so the role kinds need
their own: `drive_master.kind.json` and `drive_follower.kind.json` (the content of `conveyor_drive.kind.json` under the new
name, so two new kind digests) and boards `drive-master-controller` and `drive-follower-controller` (raw names
`drive_master.*`, `drive_follower__*`). The conveyor kind file, its board and plants A and B do not change.

**Oracle** (`test/contract/belt_master_loss_test.go`, closed forms in the test, noiseless plant, master lost at 600 s):

| Check | Closed form | World |
|---|---|---|
| `stop_all`: followers stop | k reference samples after D-1 (k = 1, 2) | 1 s, 2 s |
| shortfall | the belt counts its drives one sample late; one drive's run limit 1380 N < demand 1866 N: 2 s | - |
| stall trip | shortfall + `boost_s` = 4 s | 4 s |
| stopped (v < 0.1 m/s) | shortfall + 2 s · ln(2 / 0.1) = 7.99 s | 8 s |
| stranded | stopped + 10 s = 17.99 s | 18 s |
| restart margin | 2 · 1.5 · 920 − (1630 + 1.424 · m) − 6 · 27.6 < 0 | to 1e-6 N |
| a restart (`stop_belt`, `start_belt`) | stalls again within `boost_s` + 1 s, a second stall counted | 1 s, 2 stalls |
| `promote`: D-2 leads | 1 + `promote_s`: 3 s (and 6 s with `promote_s` 5); D-3 never | 3 s, 6 s; D-3 0 |
| survivors' load | (F_empty + k_m · m + v · feed + drag) / 2 / F_rated = 0.98437 | 0.98437 each |
| until the master is lost | every state of plant A, bit for bit, through a D-3 trip | equal |

**Stories.** `s6-a-stop` (plant `s6-stop`) and `s6-a-promote` (plant `s6-promote`), 3 h, wall pace, ten rows drawn once
with Python's `random.Random(20261109)`: seed U{1, 2³¹ − 1}, `trip_s` U{600, 2400}, and for `promote` the downstream stop
`stop_s` = `trip_s` + U{600, 2400} for 30 min (S2-A's). Reference policies: `stop` feed 0 ten seconds after the trip and
nothing else (no restart: stranded is unavoidable, harm is the criterion); `promote` S2-A's reaction (limp feed 54 % at
trip + 30 s). `s6-a-stop` is marked `unavoidable`: no reaction prevents the stop, so its reference only does no harm and
ties doing nothing exactly (calls are priced only on a story whose do-nothing run is clean); a futile restart costs a
second stall.

**What agentic-stream needs for S6** (not edited here): the two kind files and boards in its S6 site; births of D-1 as
`drive_master` and D-2, D-3 as `drive_follower`, with per-device reading rules (A1) under the new kind names; catalog
routes for `start_drive`, `stop_drive`, `reset_trip` per role kind; a spec phase `master_lost` (D-1's current near 0 and,
under `stop_all`, every drive's run status 0 and the tail pulley stopped with a loaded belt) that refuses `start_belt` while
loaded and raises an R0 advisory (manual unloading, an empty-restart plan); under `promote` S2-A's reaction on D-1's loss;
the verdict subject map for `drive_master__vfd_fault` on D-1 (diagnosis `vfd_trip`, expected `set_feed` on F-1); the S2
playbook rule pins `conveyor_drive`'s kind digest, so it does not apply to the role kinds, and an S6 rule needs its own pin.

**Evidence** (every seed row on noiseless worlds, `belt_stories_*_test.go`; first row through the binary,
`test/acceptance`):

| Story | Do nothing (10 rows) | Reference (10 rows) | J reference − J do nothing | B1 vibration | B2 current | B3 safe limp | B4 drive lost | easy |
|---|---|---|---|---|---|---|---|---|
| S6-A-stop | stall at trip + 4 s, stranded at trip + 18 s and to the horizon, one stall, no protective trip | the same; no restart | 0 exactly (ties) | - (ties) | - (ties) | ties (−0.09) | ties (0.00) | yes, by design |
| S6-A-promote | D-2 leads at trip + 3 s; survivors carry 0.98x, no trip; stranded at the downstream stop | D-2 leads at trip + 3 s; never stranded, no survivor trip | −4.94 to −5.47 h (beats) | - (−5.13) | −5.30 | −0.16 | −5.31 | no |

B0 is not clean on either (the stall and the stranded belt are priced), so both stories are valid. On S6-A-stop nothing
prevents the stop: the reference policy (feed 0 ten seconds after the trip) ties doing nothing exactly because calls are
free on a valid story, and every baseline ties it (B3's lost throughput before the trip, 0.09 h on the first row, is inside
the 0.1 h tie), so the story is `easy`: harm is its criterion, and J sees harm as a second stall (0.5 h) for a futile
restart. On S6-A-promote the story is S2-A with the master's drive lost: the reference beats every baseline, B3 by the
smallest margin (0.16 h), as on S2-A. The 40 long-tier runs of S6 (3 h each) take about 3 min at `GOMAXPROCS=4`;
`streamsim outcome` on an S6 run about 2.5 min.

Open point (world, not fixed here): each drive-to-drive link costs the world's lazy coupling about as much as the rest of
the plant, so an S6 plant steps about 2.6 times slower than plant A (3.5 s against 1.3 s per plant hour; `@2` alone 1.7 s);
its 24 h domain lint takes about 87 s against 33 s.

## 8. Live evidence, arm B (GLM 5.3 Flash, first three seed rows)

Four rounds on the barrier-paced twin through Tamoz, 2026-11-09 (agentic-stream `belt-m4-evidence`, evidence kept
outside the repository):

| Story | Round 1 | Round 2 | Round 3 | Round 4 |
|---|---|---|---|---|
| S2-B | 3/3 (198 quarantines each) | 3/3, 0 quarantines | - | - |
| S3-A | 2/3 | 1/3 (feed level missing: the objective named a plant) | rows 1, 3 pass | - |
| S1-A | 0/3 (`unknown` and a watch only) | 0/3 (Tamoz timeout, rejection, watch only) | 0/3 (steps refused as `action_without_diagnosis`; rows 2, 3 isolated D-3 but never restored the feed) | 3/3 |

Round 4 on S1-A: every row plays the whole plan (limp, belt empty, isolate D-3, limp), no trip, no boundary violation,
J - B0 -3.11, -2.40, -3.20; best baseline B3 by -1.27, -0.60, -1.53; detection 0.38-0.87 h after onset with the label
correct; 4 model calls and about 15 k tokens per run. What changed: O34 (act before the cause is known), O35 (a watch
wakes reasoning; failed reasoning is retried), O36 (`drive_degrading`), the per-plant limp feed in the device catalog.

Finding: the model alone does not take a cause-free step under an unknown cause; the loop around it must name the
symptom and keep asking when evidence changes.

## 9. S6 consumer side (agentic-stream `belt-m4-s6`, 2026-11-10)

Spec `master-loss`: phases `running`, `master_lost`, then `stranded` (stop_all: `stop_feed`, R2) or `followers_lead`
(promote: limp 54 %, R1), diagnosis `drive_tripped`. Scripted twin row 1: s6-a-promote beats every baseline (B0 −5.14,
B3 −0.18, B4 −5.32); s6-a-stop ties every baseline and is `easy`, with no restart while loaded. Gaps for the simulator
backlog, not built:

- An R0 advisory (an unload request to a person) has no route: the plant wire carries no operator-message record, and
  every declared intent must route to a device. The request is text in the decision's summary.
- An empty restart is not expressible: no reading says the belt is empty (unloading is manual), and the S6 plants have
  no `start_belt` route. "No restart while loaded" holds because no start intent is declared.
- The outcome counts D-2's `ref_held` promotion latch as a trip; the promote story is judged against doing nothing's
  trips.
