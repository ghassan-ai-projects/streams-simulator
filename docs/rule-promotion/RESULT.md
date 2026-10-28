# Rule promotion: first-slice result

- **Date**: 2026-10-28
- **Status**: arm A done (14 runs); arm B not run. The campaign is being reshaped (several faults per run, shorter
  runs) and three consumer defects found here block arm B (see §4).
- **Builds**: Agentic Stream `41f23b3d`, Tamoz `c4a758f1` (model through `ZAI_API_BASE`), Streams Simulator
  `7cb0e02c`, all on `rule-promotion`.
- **Evidence**: [evidence/arm-a/](evidence/arm-a/) holds the campaign spec, every recorded model answer
  (`responses.jsonl`), every request sent to Tamoz with its full streamed answer (`worker-exchanges.jsonl`), one
  line per episode (`episodes.jsonl`) and each run's label, consumer verdict and simulator scorecard. The full run
  folders and the consumer database (5.5 GB) are kept outside the repositories in
  `~/my-projects/rule-promotion-runs/slice3-A`.

## 1. What ran

The first-slice campaign of [SCENARIOS.md](SCENARIOS.md): one pump station of 8 pumps, 24 h per run, one fault per
run, faults `bearing_wear`, `misalignment` and `sensor_drift` in equal shares and `transient_none` (cavitation) as
the negative class. Arm A is Tamoz alone: every situation the scheduler admits becomes a real Tamoz episode, and no
rule is learned.

It was planned as 33 runs (24 learn, 9 holdout) and stopped after 14 finished runs (all in the learn phase),
because the findings below make further arm A runs uninformative until they are fixed. Run 1 builds the 24 h
baseline. Runs 1, 5, 10 and 14 are negatives; the other 10 carry a fault.

## 2. Results

Times are hours after the fault's injection, in simulated time. *Observable* and *deadline* come from the sealed
label (bearing wear: 1.43 h and 13.43 h; misalignment: 0.73 h and 12.73 h; sensor drift: 0.67 h and 2.67 h).

| Run | Truth | Pump | Episodes (h) | Tamoz said (confidence) | Action booked | Right repair by deadline |
|---|---|---|---|---|---|---|
| 1 | transient_none | 5 | none | | none | n/a: no false action |
| 2 | sensor_drift | 5 | 1.62, 2.92 | bearing_wear (0.55, 0.62) | maintenance: bearing_wear ×2 | no: wrong work order |
| 3 | misalignment | 1 | 1.77, 3.35 | misalignment (0.70, 0.80) | maintenance: misalignment ×2 | yes |
| 4 | bearing_wear | 8 | 3.31, 6.61 | bearing_wear (0.45, 0.45) | watch ×2 | no: watch only |
| 5 | transient_none | 3 | none | | none | n/a: no false action |
| 6 | bearing_wear | 5 | 3.09, 6.97 | bearing_wear (0.55, 0.45) | maintenance, then watch | yes |
| 7 | misalignment | 4 | 1.85, 3.43 | misalignment (0.70, 0.85) | maintenance ×2 | yes |
| 8 | misalignment | 4 | 9.75 | misalignment (0.80) | maintenance | yes, late |
| 9 | sensor_drift | 7 | 1.63, 3.06 | bearing_wear (0.62, 0.62) | maintenance: bearing_wear ×2 | no: wrong work order |
| 10 | transient_none | 7 | none | | none | n/a: no false action |
| 11 | bearing_wear | 1 | 3.02, 6.45 | bearing_wear (0.45, 0.55) | watch, then maintenance | yes |
| 12 | bearing_wear | 1 | 12.92 | bearing_wear (0.45) | watch | no: watch only |
| 13 | misalignment | 2 | 1.84, 3.34 | misalignment (0.70, 0.80) | maintenance ×2 | yes |
| 14 | transient_none | 2 | none | | none | n/a: no false action |

- **Detection.** Every faulted run got an episode on the right pump (10 of 10). No negative run got an episode or an
  action (4 of 4). In 8 of 10 the first episode came between the fault becoming observable and the label's
  "unavoidable" time (the signal-to-noise ratio at which detection is trivial). The two late ones, runs 8 and 12,
  hit a pump whose situation from the run before was still open.
- **Diagnosis.** 14 of 18 decisions named the true fault. All four wrong ones are the two `sensor_drift` runs,
  both diagnosed `bearing_wear` with confidence 0.55 to 0.62. Bearing wear and misalignment were never confused.
- **Action.** The right repair was booked before the deadline in 6 of 10 faulted runs (misalignment 4 of 4,
  bearing wear 2 of 4, sensor drift 0 of 2). Tamoz picks a watch rather than maintenance whenever its confidence is
  0.45, which happened in 5 of the 7 decisions on bearing-wear runs; in runs 4 and 12 the watch was all it booked.
- **Model calls.** 18 episodes for 10 faults: about 2 per fault, because each fault's situation got a new version
  1.3 to 3.9 h after the first decision and was reasoned about again. All 18 produced a valid decision; none failed or
  timed out. Tamoz took 19 to 55 s per episode (median
  38 s). Tamoz reports no token usage on its terminal events, so tokens per decision are unknown.
- **Throughput.** About 37 min of wall time per 24 h run, almost all of it the consumer processing the station's
  127,000 readings; model time was about 10 min over the whole arm.

## 3. Pass conditions

None of P1 to P7, S2 or S7 can be read yet: they compare arm B with arm A in the holdout phase, and neither arm
reached it. What arm A gives the comparison:

- **P1 (model calls per handled fault).** The baseline is about 2 calls per fault, not 1. Part of that is likely
  caused by defect 4.1: the fault was never repaired, so it kept moving the situation.
- **P2, P3 (damage, false actions).** No false action in 4 negatives. Damage cannot be judged: no repair reached
  the plant (4.1), so every fault ran to its failure in the simulated world whatever the consumer decided.
- **P4 (rule-handled cases correct).** The risk is now concrete: `sensor_drift` cases arrive labelled
  `bearing_wear`. Whether a rule learns from them depends on case verification, which needs real repairs (4.1).

## 4. Defects found

1. **No command reached the plant (blocker).** The campaign starts `serve` with `--plant-wire-socket` but without
   `--device-catalog`. With no catalog, `serve` falls back to its in-process simulated effector, which reports
   every command `succeeded` (provider `agentic-stream-simulator`) and sends nothing. The run journals show only
   heartbeats and state from the host, the simulator scores every run `effect_calls: 0` and `resolved: false`,
   and every ledger outcome says the command succeeded. Fix: pass the pump station's catalog, and make `serve`
   refuse a plant wire socket without a catalog instead of falling back silently.
2. **Watches never fire.** The `install_watch_condition` preset's expression is
   `features.vibration_mean_15m >= 4.0`, but the feature is not found under that name: each reading on a watched
   pump logged
   "watch expression evaluation skipped: no such key: vibration_mean_15m" (513,000 warnings in 14 runs). A
   low-confidence diagnosis that books a watch therefore never escalates; this is why runs 4 and 12 got no repair.
3. **Verdicts report the situation's first event, not the detection.** The consumer verdict's `detected_at` is the
   situation's `first_event_time`, which is often days before the run (2026-01-01 00:00:10 for pumps 2 and 7: the
   first 10 s of the warm-up run), so the simulator marks detections suspicious. Runs 6, 8, 11 and 12, whose situation
   opened in an earlier run, report no detection at all, and action times are not issue times (run 4's commands,
   issued at 08:09 and 11:27 on 2026-01-04, are stamped 2026-01-05 00:00:57, after the run ended).

## 5. Next

1. Fix 4.1 to 4.3 in Agentic Stream.
2. Reshape the campaign so it costs a few hours instead of about 30: several faults per run on different pumps
   (with a cavitation look-alike beside them), runs of about 16 h with faults starting in the first 2 h, one shared
   warm-up database for both arms, a learn phase that ends as soon as every key has an active rule or cannot get
   one, and arm A on the holdout only.
3. Run arm B, then arm A's holdout, and fill §3.
