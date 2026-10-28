# Rule promotion: first-slice result

- **Date**: 2026-10-28
- **Status**: campaign 2 ran twice. The first run (§2, §3) showed learning working end to end but the learned rule
  never acting, because its envelope was too narrow and the rate limit counted wall-clock hours. After two design
  changes in Agentic Stream (§0) the rerun, **campaign 2b**, shows the rule doing the work it was learned for:
  P2, P3, P4, P6 and S2 pass; **P1 fails** (only one of three fault families got a rule) and **P5 fails** (the
  scheduler's debounce counts wall-clock time and rule episodes queue behind model episodes).
- **Builds** (campaign 2b): Agentic Stream `62ccc278`, Tamoz `d9fcbdb7` (model through `ZAI_API_BASE`), Streams
  Simulator `5baed2eb` (code), all on `rule-promotion`. Plant wire contract 1.5.0.
- **Evidence**: [evidence/campaign-2b/](evidence/campaign-2b/) and [evidence/campaign-2/](evidence/campaign-2/) hold
  the campaign spec and, per arm, every recorded model answer (`responses.jsonl`), every request sent to Tamoz with
  its streamed answer (`worker-exchanges.jsonl`), the campaign score, an extract of the consumer's ledgers
  (`ledgers.json`: cases, drafts, proposals with envelope and cap, rule evaluations, rule fires, policy results,
  commands) and each run's labels and consumer verdict; arm B also has the playbook it learned.
  [evidence/arm-a/](evidence/arm-a/) is campaign 1. Full run folders and databases are kept outside the repositories
  in `~/my-projects/rule-promotion-runs/` (`cap-B`, `cap-A`, `settled-B`, `settled-A`, `slice3-A`; `cap-B-ratelimited`,
  `cap-B-slept` and `cap-A-slow` are interrupted attempts).

## 0. Campaign 2b: after the envelope and rate-limit changes

Same campaign spec and seed as campaign 2 (§1), with two design changes in Agentic Stream:

- **Envelope** (`adbad905`): a rule's cap widens its source cases' range by `envelopeMargin` (0.1) of each feature's
  magnitude instead of the range's width, an active rule acts anywhere inside its approved cap (the envelope stays
  the record of what the source cases covered), and only shadow matches inside the cap count toward the shadow exit.
  The misalignment rule's cap on `axial_mean_15m` went from 2.746 to 2.765 (campaign 2) to 2.47 to 3.04.
- **Rate limit** (`5fe03bb1`): the policy plane's hourly limit counts the event time of the situation version an
  intent answers, not the wall clock. The playbook's per-rule rate cap already did.

Arm B learned for 3 runs (learning settled), then played the 3 holdout runs in 56 minutes; arm A played the same
holdout from arm B's learned database in 25 minutes. No booking was denied by the rate limit in either arm.

### Learning (arm B)

- **Misalignment rule**: `axial_mean_15m / axial_mean_24h ≥ 1.3`, radial vibration and current not risen, heartbeat
  present. Replay clean, approved, 3 shadow agreements inside the cap, promoted to active during learn run 2.
- **Bearing wear**: proposed three times and rejected three times by the careful approver, each replay showing one
  disagreement; the key is closed until new evidence.
- **Sensor drift**: never diagnosed (always called bearing wear), so never a key.

### Holdout (3 runs, 18 faults, 3 look-alikes)

| | Arm B | Arm A |
|---|---|---|
| Model episodes | 13 | 15 |
| Rule decisions | 5 | 0 |
| Faults detected (right pump) | 17 of 18 | 15 of 18 |
| Right diagnosis | 11 of 18 | 11 of 18 |
| Right repair by the deadline | 7 of 18 | 6 of 18 |
| Misalignment: right repair by the deadline | 5 of 6 | 5 of 6 |
| Bearing wear: right repair by the deadline | 2 of 6 | 1 of 6 |
| Sensor drift: right repair by the deadline | 0 of 6 | 0 of 6 |
| Action on a look-alike | 0 of 3 | 0 of 3 |
| Bookings denied `rate_limited` | 0 | 0 |

- Every rule decision was a true misalignment (pumps 2 and 8 in holdout run 1, 5 and 8 in run 2, 3 in run 3), each
  booked with the right work order before its deadline. The sixth misalignment (pump 4, run 1) got no repair in
  either arm: its pump was blocked by an unknown outcome earlier (§5.3).
- The bearing-wear difference between the arms comes from the model, not from a rule: no bearing-wear rule exists.

### Pass conditions (campaign 2b)

| | Condition | Result |
|---|---|---|
| P1 | Model calls per handled fault: B at most 50% of A | **Fail.** 13 of 17 detected faults (0.76) against 15 of 15 (1.00): 76%. The misalignments the rule handled took no model call; bearing wear and sensor drift got no rule. |
| P2 | Damage: B no more than A | **Pass.** 11 faults unrepaired by the deadline in B, 12 in A. |
| P3 | False actions on healthy pumps: B no more than A, none by a rule | **Pass.** 0 and 0. |
| P4 | Every rule-handled case right (action and deadline) | **Pass.** 5 of 5. |
| P5 | Situation to recommendation for rule cases ≤ 1 s p95, wall time | **Fail.** 147 s p95. The rule itself answers in about 1 ms; each case waits the trigger's `debounce: 1m`, counted in wall-clock time (§5.5), then up to 87 s behind model episodes, because episodes of a tenant run one at a time. |
| P6 | Every pair with 3 verified, uncontradicted cases got a proposal | **Pass.** Misalignment and bearing wear. |
| P7 | Masking trap | **Not covered.** |
| S2 | Cavitation look-alike | **Pass.** 0 of 3 in both arms. |
| S7 | | **Not covered.** |

## 1. What ran

**Campaign 1** (arm A only, 14 runs, §4): the first-slice campaign of [SCENARIOS.md](SCENARIOS.md) as first
written, one fault per 24 h run. It was stopped early because three consumer defects made it uninformative.

**Campaign 2** (both arms), reshaped to cost about 1.5 hours instead of about 30:

- One pump station of 8 pumps. Each run is **30 simulated hours**; up to 6 faults start on different pumps in the
  first 2 hours (`bearing_wear`, `misalignment`, `sensor_drift` in equal weights) beside one `transient_none`
  (cavitation) look-alike. 30 h leaves every case its full 24 h settle window inside its own run (last fault plus
  about 4 h to detect and act plus 24 h; checked for every run before launch).
- Arm B: 1 warm-up run (no faults), up to 7 learn runs, 3 holdout runs. Learning ends early once it settles (no
  draft or proposal in flight, every playbook rule active); the database it ends with is kept.
- Arm A: the same 3 holdout runs, Tamoz alone, started from a copy of arm B's learned database without its playbook,
  so both arms enter the holdout with the same history.
- The world is paced by the consumer: the node waits at each 60 s slice until the consumer has ingested the slice
  and has no reasoning in flight (up to 240 s, an episode's budget), so the world never outruns the model.
- The approver is the campaign's careful approver: it approves a proposal only when the replay over past data shows
  no disagreement.

Wall time: arm B 57 min (7 runs), arm A 26 min (3 runs). Learning settled after 3 learn runs, so learn runs 4 to 7
were skipped.

## 2. Results (campaign 2, before the changes)

### 2.1 Learning (arm B)

| Step | What happened |
|---|---|
| Cases | 15 settled: misalignment 6 verified; bearing wear 6 verified, 3 not verified (two because a new situation opened on the pump inside the window, one because its command was never acknowledged, §4.6). |
| Drafts | 3 drafts, all answered by Tamoz with a proposal (1,133 to 1,365 input tokens and 999 to 1,669 output tokens each). |
| Misalignment proposal | `(axial_mean_15m / axial_mean_24h) >= 1.25 && !heartbeat_missing_5m`. Replay: would have fired 3 times, agreed 3, disagreed 0. Approved, written in shadow mode. |
| Bearing-wear proposals | `vibration_mean_15m / vibration_mean_24h >= 1.25` (replay 5 agree, 1 disagree) and `> 1.2×` (6 agree, 1 disagree). Both rejected by the approver because of the disagreement. |
| Shadow | 3 agreeing shadow matches, then promoted to active. Only 1 of the 3 was inside the rule's envelope (§5.1), and two were the same situation reasoned twice. |
| End | Learning settled: one active rule, bearing wear closed until new evidence, no other key with 3 verified cases. |

The proposed conditions are sensible: each compares a reading with the same pump's own 24 h baseline, so it would
apply to any pump, which is what the draft instructions ask for.

### 2.2 Holdout (3 runs, 18 faults, 3 look-alikes)

| | Arm B | Arm A |
|---|---|---|
| Model episodes | 16 | 16 |
| Rule decisions | 0 | 0 |
| Faults detected (right pump) | 16 of 18 | 16 of 18 |
| Right diagnosis | 11 of 18 (all 6 sensor drifts called bearing wear) | 11 of 18 (same) |
| Right repair by the deadline | 4 of 18 | 5 of 18 |
| Action on a look-alike | 0 of 3 | 0 of 3 |
| Holdout bookings denied `rate_limited` | 7 of 16 | 7 of 15 |

Per fault family, right repair by the deadline: misalignment B 3 of 6, A 4 of 6; bearing wear 1 of 6 in both;
sensor drift 0 of 6 in both. Most faults with a right diagnosis and no repair were denied by the policy plane's rate
limit (§5.2): in arm A's last holdout run all 5 maintenance bookings were denied. The one-fault difference between
the arms is which bookings the rate limit happened to deny (pump 8's misalignment in the second holdout run was
denied in B and booked in A), not anything a rule did.

### 2.3 Throughput

Barrier answers held at about 0.17 s per 60 s slice for the whole campaign (§4.5). A run without faults took about
7 minutes of wall time; a run with 7 disturbances took 8 to 10 minutes, the difference being model time (episodes 20
to 90 s each). No episode failed; 33 produced a decision in arm B and 33 in arm A (including the learn phase arm A
inherited).

## 3. Pass conditions (campaign 2, before the changes)

| | Condition | Result |
|---|---|---|
| P1 | Model calls per handled fault: B at most 50% of A | **Fail.** 16 against 16 (100%). The active rule matched each misalignment but never fell inside its envelope (§5.1), so every case went to Tamoz. |
| P2 | Damage: B no more than A | **Not readable as a rule effect.** B left 14 faults unrepaired by the deadline, A 13; no rule acted, and the difference is which bookings the wall-clock rate limit denied (§5.2). |
| P3 | False actions on healthy pumps: B no more than A, none by a rule | **Pass.** 0 and 0; no rule acted. |
| P4 | Every rule-handled case right: 100% | **Nothing to measure**: no rule-handled case. |
| P5 | Rule decision p95 ≤ 1 s | **Nothing to measure**: no rule decision. |
| P6 | Every fault and action pair with 3 verified, uncontradicted cases got a proposal | **Pass.** Misalignment (6 verified) and bearing wear (6 verified) both got proposals; no other pair reached 3. |
| P7 | Masking trap never passes silently | **Not covered**: no `reduce_load` rule was proposed. |
| S2 | Cavitation look-alike: no false maintenance, none by a rule | **Pass** in both arms (0 of 3). |
| S7 | | **Not covered** by this campaign. |

Also not covered: arm 0, arm C, repeated seeds, and replay identity of a campaign.

## 4. Defects found and fixed

Campaign 1 and the attempts at campaign 2 found these; all are fixed on `rule-promotion` and the fix of each is
covered by a regression test.

1. **No command reached the plant.** `serve` fell back to its in-process simulated effector without a device
   catalog. It now refuses a plant wire socket without one, and the campaign passes the pump station catalog
   (Agentic Stream `2bef09f9`).
2. **Watches never fired** (`no such key: vibration_mean_15m`): a watch was evaluated against raw event data instead
   of its situation's features (`495959b2`).
3. **Verdict times were wrong**: detections and actions were stamped with the situation's first event, often days
   earlier (`84557ea5`).
4. **The world outran the model, and the node then dropped the host.** A barrier-paced world raced ahead of
   episodes (`52057ed3`: the host holds a barrier while reasoning is in flight). That hold then lasted up to 4.5
   minutes when 6 faults arrived together, and the node counted three 30 s missed barriers as host loss. Fixed in
   the contract: plant wire 1.5.0 judges liveness by heartbeats only, and the host keeps heartbeating while it holds
   (Streams Simulator `5baed2eb`, Agentic Stream `80bcfaad`, `bf9d13c9`). Commands were also refused as `out_of_range`
   and `expired` across the simulated clock (`2fd87a28`, `7f5a8a9c`), and the campaign driver now fails a run whose
   session ends early instead of waiting forever (`d36807f5`).
5. **Barrier answers slowed from 0.2 s to 1.6 s** over the first simulated day: each reading reloaded and rewrote
   every 24 h window as one JSON blob of up to 1.3 MB. Window samples are now rows (`a3596d7e`).
6. **No rule draft could succeed**: each case carried the situation's `evidence` list (256 event ids, 10.6 KB),
   over Tamoz's 4 KB per case, and `evidence` was offered as a feature. A draft now sends only the features a rule
   may read (`3b80a31f`). Tamoz: a failed model call now names its cause (`4557f24e`), and a rate-limited call is
   retried (`d9fcbdb7`).
7. **The campaign shape was wrong** (mine): back-to-back 16 h runs refaulted pumps inside their 24 h settle window,
   so cases could not be verified. Runs are now 30 h (§1).
8. **The campaign driver scanned the whole event log every half second** to read a run's world time (1.9 s per
   query on a 1 GB database); it now reads by event log position (6 ms, `62ccc278`). An attempt at campaign 2b also
   stopped because the computer slept with its lid closed: the runtime's one-minute owner lease expired during the
   sleep and it stopped, which is the intended fencing; the rerun kept the computer awake.

## 5. Open questions (design decisions needed)

1. **The envelope stopped every rule from generalising** (resolved in campaign 2b, §0). A rule acts only when each of its features lies inside the
   range seen in its source cases, widened by `envelopeMargin` (0.1) *of that range*. The simulator's faults have
   nearly identical sizes, so the misalignment rule's envelope is `axial_mean_15m` 2.747 to 2.763 and the margin
   adds 0.0016. The envelope pins absolute levels while the condition is a ratio to the pump's own baseline, so
   it defeats the generalisation the condition was written for. The design defers envelope growth until after the
   first slice. Options: an envelope over the condition's own terms (the ratio), a margin relative to the values
   rather than the range, or envelope growth now. Shadow promotion also counted matches outside the envelope,
   where the rule can never act.
2. **The rate limit counted wall-clock hours** (resolved in campaign 2b, §0). `schedule_maintenance` is limited to 8 per hour, bucketed by the
   policy plane's wall-clock time. A paced simulated world puts about 30 simulated hours into 8 minutes, so 13
   bookings per arm were denied. Should a rate limit count the world's time (event time)?
3. **An unknown outcome on a command with nothing to read back blocks its pump.** The simulator drew `ack_lost`
   for one booking (by design: the effect happened, the acknowledgement was lost). The host recorded an unknown
   outcome and opened a reconciliation; a maintenance booking leaves nothing to read back, so only an operator can
   clear it, and every later command to that pump fails. Bookings that ended in manual review also count as
   unresolved and would block clearing too. Options: an operator resolves it (no change), or a reconciliation
   counts only commands whose device effect is uncertain (a design change).
4. **Sensor drift is always diagnosed as bearing wear**, and such a case is still verified when nothing reopens
   within 24 h, so a wrong diagnosis feeds the bearing-wear key. Verification by "no recurrence" cannot tell a
   repair from a fault that a maintenance visit happened to clear.

5. **The scheduler's debounce and cooldown count wall-clock time**, like the rate limit did. In a paced world each
   rule case waits a real minute before admission, which alone breaks P5. Counting them in event time would make a
   simulated world and a replay meet the same timing as the live plant.
6. **Rule episodes queue behind model episodes**: a tenant's episodes run one at a time, so a 1 ms rule decision can
   wait for a 30 to 90 s model episode on another pump. Rule episodes could bypass that queue.

## 6. Campaign 1 (arm A only)

14 runs of 24 h, one fault each, before the defects of §4.1 to §4.3 were fixed: every faulted run got an episode on
the right pump (10 of 10) and no negative got an action (4 of 4); 14 of 18 decisions named the true fault, the four
wrong ones being the two `sensor_drift` runs called `bearing_wear`; Tamoz booked a watch instead of maintenance at
confidence 0.45; about 2 episodes per fault. No repair reached the plant (§4.1), so damage could not be judged.
Run-by-run detail is in [evidence/arm-a/](evidence/arm-a/).

## 7. Next

1. Decide §5.5 and §5.6, then rerun campaign 2b to read P5.
2. Decide §5.3 (a pump blocked after an unknown outcome on a command with nothing to read back).
3. P1 needs rules for more than one family: bearing wear was rejected for a disagreement in its replay each time,
   and sensor drift is never diagnosed (§5.4).
