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
  | S1-B | +2.12 (B1 wins) | -3.07 | -4.63 | -3.07 | yes |
  | S2-A | - | -5.51 | -0.19 | -5.52 | no |
  | S2-B | - | -5.85 | -0.24 | -5.85 | no |
  | S3-A | - | - | -4.59 | -3.35 | no |
  | S4-A | - | -1.89 | -3.02 | - | no (not valid) |
  | S4-B | - | -2.30 | -5.24 | - | no (not valid) |
  | S5-A | -4.05 | -4.88 | -2.98 | -4.89 | no |

  Departures from PLAN.md section 2.3: B1 is not late on S1-B (it alarms 47 min before D-3 would trip, empties the belt
  and stops D-3 in time, and beats the reference policy, which limps at 18 % from 4 h after the onset: S1-B is easy by
  cost; its pass rule's 2 h lead is not part of J); B1 fires on S5-A, because a policy watches the vibration state, not
  the frozen reading; B2 does not fire on S0 (the feed never passes about 1.33x) nor on S3-A (the currents pass 1.1x only
  in the seconds before the seizure); B3 beats nothing but costs least on S2 (0.2 h behind the reference).
- `streamsim outcome` on a 16 h story replays six runs plus the reproduction check: 321 s with `GOMAXPROCS=4` on a
  loaded machine (S1-B).
- A label is keyed by entity and fault (O25): with several labels on one entity a detection is judged with the label its
  `label` names and a call with the labels that expect it; what names none is judged with each. One label per entity
  scores bit for bit as before (digest-pinned run-score documents).
