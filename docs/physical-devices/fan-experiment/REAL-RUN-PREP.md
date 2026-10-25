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
