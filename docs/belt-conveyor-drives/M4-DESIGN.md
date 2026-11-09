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
