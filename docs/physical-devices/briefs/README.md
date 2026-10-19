# Handoff briefs for the work outside this repository

The physical-devices goal (M1 to M5) built everything streamsim owns. These briefs hand over the rest: the gateway's plant wire side, the plant wire host, the independent effect instrument, and the owner's bench checks. They are working documents; they name the other repositories and tools because their readers are those teams. They contain no secrets or customer data. Written in R13, against plant wire contract 1.3.0.

| Brief | For | Milestones | What it asks |
| --- | --- | --- | --- |
| [gateway-plant-wire.md](gateway-plant-wire.md) | The Arduino gateway (`agent-research-lab`) | M2 | A plant wire node that announces the same kind files, reports readings with quality, executes commands with the epoch and policy checks, handles the relay's node-stop commands, and recovers by rebirth |
| [agentic-stream-host.md](agentic-stream-host.md) | Agentic Stream | M2 to M5 | Handle `local_control` refusals, `queued` receipts and `expired`; know that node safe stops are not visible as commands; expose the policy digest it signs with; optional `command_evidence`; epoch behaviour on host loss |
| [bench-instrument.md](bench-instrument.md) | The owner | M5 | An instrument that confirms a physical effect independently of `DRESULT` and of the board, with timestamps aligned to the run, and the acceptance for it |
| [bench-checks-m2-m5.md](bench-checks-m2-m5.md) | The owner | M2 to M5 | The exits of `docs/design/PHYSICAL_DEVICES.md` section 7 as exact commands, expected outputs and pass or fail criteria, plus the evidence to send back |

Read order for the owner: the bench checks, the instrument brief, then the other two as the work reaches them. Read order for a team: its own brief, `contracts/plant-wire/SPEC.md`, and `docs/design/PHYSICAL_DEVICES.md` sections 1, 2 and 4.2.

## Facts all four briefs rely on

- The plant wire contract is `contracts/plant-wire/` (vendored byte for byte, pinned `DIGEST`). streamsim accepts only the contract version (1.3.0) and digest it was built with.
- The device model is the kind file (`examples/devices/thermal-board/*.kind.json` for the thermal bench). A gateway announces the same kinds and `kind_digest`s; a simulated twin implements them.
- The policy digest in a command is checked by the node: the session plan's list for simulated devices, the gateway's own list for the real board. device-wire-v1 has no policy refusal code, so the gateway is the only place a real board can refuse a policy.
- streamsim only refuses or stops at its node. It never originates an ordinary command, never perturbs a real device's records, and has no path to hardware from a replay.
- A `DRESULT` is self-reported. Only an independent instrument confirms a physical effect.

## Open items that need an owner decision

Listed in the R13 report to the owner and in `governance/DEFERRED.md` (D-111 to D-115): how an independent instrument should feed the live effect stage, how a hardware-only session carries `command_evidence`, and whether the contract's minimum-version rule should replace the exact-match check.
