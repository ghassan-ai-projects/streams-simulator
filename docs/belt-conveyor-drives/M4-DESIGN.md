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
| 0 | all | Throughput first: a 16 h story on the barrier-paced twin takes about an hour (S1-A row 1: 63 min); find where the time goes and bring a 16 h story under 15 min, or the 10-seed evidence takes days (O32) |
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
- O32 Throughput before evidence: the 10-seed evidence waits for item 0.

## 5. Exit

PLAN.md section 7: every story's pass rule over 10 seeds; the A arm's report against the baselines; `easy` flags
explain any story a baseline passes; `verify-run` and zero quarantines on every run; `make phase-gate` in streamsim,
`make ci-check` in agentic-stream and tamoz's own gate; the fan guards. Left for the owner: the live approval of the
playbook proposal, arms B and C (T6), and the FR+ fan rung. Then the full end-to-end test of all seven stories (O13).
