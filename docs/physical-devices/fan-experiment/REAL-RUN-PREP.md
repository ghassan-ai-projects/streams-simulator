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

## Real rung: what remains (needs the owner at the board)

No board was attached (`/dev/cu.usb*` empty). A real plan with the real policy digest checks as
`environment: hardware`, three `real` instances (`examples/devices/thermal-board/fan-experiment/real.plan.json` with
the sockets under a short owner-only folder and the policy list replaced by the digest `serve` logs).
Follow `agent-research-lab/real-world-sensor/assessment/RUNBOOK-PLANTWIRE-GATEWAY.md`, the fan-experiment README and
PLAN section 5 (O1 to O5). Start order: worker, `serve` (same flags as above), gateway on the Mega's serial port,
`plantwire connect --plan real.plan.json`. Open: the Mega may reset when the serial port opens (UNVERIFIED), W15.
