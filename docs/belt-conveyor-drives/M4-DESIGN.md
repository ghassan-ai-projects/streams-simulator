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
