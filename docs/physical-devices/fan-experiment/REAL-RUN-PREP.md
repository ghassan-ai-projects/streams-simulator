# Real run: preparation record (2026-10-25)

Branch `fan-real-prep` in streams-simulator, agentic-stream (with `fan-a` merged in), agent-research-lab and tamoz,
all cut from `main`. Nothing pushed.

## Simulated rung (P4 repeat on the new branches): passed

Run folder `~/my-projects/.e2e-run/fan-p4-20261025T090533` (kit `fan-p4-kit`, `SIM_REF=AS_REF=fan-real-prep`).

- Preflight tests (`TestThePlantWireEffectorCommandsAPacedStreamsimSession`, `TestTheFanExperimentCloses...`,
  `TestServeHostsTwoSessions...`) passed against the fresh binaries.
- One `serve` with `--effect-profile physical --live-actuation --owner-authorized --plant-wire-environment mixed
  --hardware-boards mega-01` announced `environment=mixed hardware_boards=[mega-01]`; policy digest matched the plan.
- Zone 34 C to 42 C; zero quarantines; readings every 2.03 s mean.
- 3 commands: `show_state alert` twice, `set_cooling 45` once; each DRECEIPT accepted, DRESULT succeeded, verified by
  the twin's reading (fan `level_percent` 0, 45, 0).
- `agentic-stream verify-run`: `pass`; `streams-simulator verify`: `matches: true`.
- Observation: the model proposed `set_indicator alert` twice (08:10:44 and 08:11:14 UTC), the second redundant.
- The run was stopped at about T+16 min (plateau reached, evidence complete), not at T+35.
- Gateway: contract `in sync`; 259 gateway tests (against a PTY stand-in board and the new binary) pass.

## Real rung (P6) result, 2026-10-25: the loop closed on the real board

Run folder `~/my-projects/.e2e-run/fan-p4-20261025T145716` (canvas run `c-0277c88f`; an earlier session `c-09a46393` was stopped
to change the spec). The owner was at the board, E-stop within reach, Approve each on for every device, and warmed the DHT11 by hand.

- **Deviation (bench-scale spec):** open at mean > 23 C with slope > 6 C/h (was 30), fan escalation at mean > 26 C (was 38), close at
  mean <= 22 C (was 28). Spec digest `sha256:8b57e684...`, policy digest `sha256:3f34cef3...05a1`. Kept in the kit
  (`fan-p4-kit/agentic-stream/zone-thermal-bench-fan.situation.yaml`; the committed bench spec is `...committed-bench`).
- **Commands:** `show_state alert` twice, `set_cooling 45` once; each queued at the node, approved by the owner, DRECEIPT accepted,
  DRESULT succeeded, verified by the board's reading (`state alert`; fan `level_percent` 0, 45 at 15:09:15, 0 at 15:09:21 local).
  Zero quarantines; temperature readings 2.01 s mean gap, 0 gaps over 4 s; 3 episodes, 3 decisions, 3 verifications.
- **Verification:** `agentic-stream verify-run` pass; `streams-simulator verify` matches. Gateway stopped with LED and fan safe stops.
- **Not exercised:** deny, hold, expiry (criteria 4 denial and 6), unplug/replug, host loss, emergency stop.

Findings:

1. The first LED command was approved 0.4 s before its deadline (`expires_after_ms` 387) and still succeeded.
2. The model sent `show_state alert` twice, 26 s apart (the second was redundant).
3. The situation stayed in `over_ceiling` when the zone cooled to 25 C: the spec has no exit from it except closing (mean <= 22 C for
   2 min), so no wind-down decision was made.
4. The gateway logged 10 receipt/result pairs as `firmware_answer_unmatched` with `command_id` unknown. They look like answers to the
   gateway's own polls (`plantwire_board.py` polls the board after each command and every interval); UNVERIFIED, no harm seen.
5. The canvas boxes for the fan and LED stay blank (D-120); the gateway announces the board as `board-01`, so the plan binds `board-01` and `serve` takes `--hardware-boards board-01` (fixed: `real.plan.json` and its README now say `board-01`).
6. `check-run.sh` C5 reads `session.plan.json`, which was the committed plan with the old digest, not the real plan; the digests the
   commands carried match the real plan and the gateway's list.
7. The Mega resets when the port opens (new `boot_id` each gateway start: boot-69, 70, 71): confirmed.
8. A fan command is a 5 s burst by design (firmware lease) and at most 6 per hour: the run proves the loop, not sustained cooling.

## Real rung with the fan unplugged (2026-10-25, second session)

Run folder `~/my-projects/.e2e-run/fan-p4-20261025T152716` (canvas run `c-34ea80d9`), same bench-scale spec and policy digest, owner
approving each command, motor supply disconnected with power off, LED and sensor attached. Archived outside git:
`~/my-projects/.e2e-run/archive/fan-real-run-unplugged-20261025.tar.gz` (checksum beside it); the first session is
`fan-real-run-20261025.tar.gz`.

- **Silent fan not detected:** `set_cooling 45` was queued, approved, accepted, `DRESULT succeeded`, the board read back `level_percent`
  45 then 0 (about 5.5 s), and Agentic Stream verified it `succeeded`. The board reports the commanded duty, so a disconnected fan is
  invisible to the chain; only the zone temperature failing to fall could show it.
- **Silent board:** the first `show_state alert` was forwarded at 15:31:12 and the board gave no answer in 2 s (`firmware_silent`); the
  gateway sent the LED its safe stop and answered `not_ready`; the LED never changed; Agentic Stream marked it `failed`. Cause unknown
  (it did not happen in the first session).
- **Deny:** the second `show_state alert` was refused at the node as `local_control`; Agentic Stream marked it `failed` and did not retry.
- **`verify-run` says `fail`** (`unresolved_action_outcomes=2`): the two failed LED commands leave their verifications unresolved, so a
  run with any refused or failed command fails verification (W15). The simulator's `verify` matches. Zero quarantines, readings 2.01 s mean.
- Correction to an earlier remark in the session: the first LED command was not an approval-window expiry; it was approved and reached the board.

## Real rung, third session: fan connected, with the fixes (2026-10-25)

Run folder `~/my-projects/.e2e-run/fan-p4-20261025T160108` (canvas run `c-0e26955c`), Agentic Stream `fan-w15` (W15 fix), gateway with the
quiet-window hold, canvas with D-120 fixed, owner approving each command. Archived outside git: `~/my-projects/.e2e-run/archive/fan-real-run3-20261025.tar.gz`.

- `show_state alert` (15:04:53) and `set_cooling 45` (15:05:42) each queued, approved, accepted, `DRESULT succeeded`; the fan read
  `level_percent` 45 at 15:05:43 and 0 at 15:05:48; verified by the board's reading. **No silent-board event, no `not_ready`.**
- **`verify-run` pass** and `streams-simulator verify` matches (the first session's two failed commands in session 2 had made it fail: W15).
  Zero quarantines, readings 2.01 s mean, max gap 2.0 s.
- The gateway still logged 18 `firmware_answer_unmatched` pairs (`command_id` unknown) in 9 minutes: the 30-byte state polls are
  also damaged when they land in the firmware's busy window; they are not held by the quiet window and lose nothing. A follow-up is to send
  polls in the window too.
- **Twin vs board** (the review page): `fan-01.level_percent` agrees (184 readings, 0 beyond ±3 %); `zone-01.temperature` diverges (the twin
  rests at 34 C, the bench at 22 to 29 C: the twin's ambient is not this bench's); `zone-01.humidity` diverges (the twin holds 42 %, the
  board read 47 to 72 % while the sensor was warmed by hand); `led-01.state` cannot be judged (an enum, no numeric tolerance).

## Twin calibrated to the bench (2026-10-25)

`examples/devices/thermal-board/fan-experiment/arduino-thermal.bench.domain.json`: rest 22 degC (was 34: the twin assumed 12 degC of board
heat the DHT11 does not see), humidity intercept 97 (53 percent at rest, was 42), time constant kept at 180 s (the real decays fitted
126 to 354 s). `domain lint`: 0 findings. **Compare with twin on the third session's run, before and after:**

| Metric | Before | After |
|---|---|---|
| `zone-01.temperature` beyond tolerance | 188 of 188 | 115 of 188 (all while the sensor was warmed by hand, 00:02:51 to 00:06:39; the rest and the cool-down are within +-2.5 degC) |
| `zone-01.humidity` beyond tolerance | 136 of 189 | 108 of 189 (the board read 39 to 72 percent while warmed and breathed on; the twin holds 53) |
| `fan-01.level_percent` | agrees | agrees |
| all readings beyond tolerance | 324 of 561 | 223 of 561 |

What remains is the hand: a heat and moisture source the twin has no input for. The comparison cannot inject a disturbance; a stated
`zone_heat_load` disturbance in the twin would be the next step.

