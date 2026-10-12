# The plant wire in the simulator

- **Date**: 2026-10-12 (phase P7a)
- **Contract**: [contracts/plant-wire/SPEC.md](../../contracts/plant-wire/SPEC.md). This document holds what is specific to
  this repository: where the node lives, how it joins the run, and where each review finding lands.

## 1. Placement (review item M5)

- **Module**: `internal/plantwire` (core with edges). The domain layer builds births from the domain's kinds through an
  allow-list of fields, sequences records at transmission, maps commands and refusals, derives the digests; the app layer runs a
  node session (handshake, lanes, rebirth, barrier, liveness); a `uds` edge dials the host's owner-only socket. The schemas are
  embedded from a byte-identical copy of `contracts/plant-wire/schemas`.
- **One effector path**: a `DCMD` becomes a call to the run's `InvokeEffector`, so wire commands, MCP `sim.effector.invoke` and
  scenario files share one command log, one ledger and one replay. World refusals map to wire classes: interlock →
  `interlocked`, failed precondition → `permissive_not_met`, random reject → `not_ready`.
- **Pipeline amendment**: world → wire records → perturbation → transmission, with `seq` assigned at transmission. A
  perturbation acting on the wire therefore drops, duplicates or reorders sequenced records, which the host must handle.
- **Determinism**: births, data and deaths come from the world in deterministic order and enter the delivery ledger; host
  records enter the command log at the world time of the slice end that processed them; replay reproduces the whole exchange.
- **Coexistence with the device path (R0-05)**: the device emulator, `deviceworld`, the vendored device-wire-v1 contract and the
  thermal LED/fan experiment stay unchanged and green in this goal. A run uses one host link: the device socket or the plant
  wire. Agentic Stream moves to the plant wire in the next goal, together with the gateway.
- **Truth canary**: P7b adds a test that injects a fault with a recognisable id and scans every wire record of the run for it.
- **Commands (P7c)**: a `DCMD` is checked in the SPEC §9 order at receipt and, when queued for `not_before_ns`, again at
  execution; it runs through `run.InvokeEffector` at the world time of the slice end that processed it, with its
  `idempotency_key` as the world command id. The node's idempotency ledger lives for the whole run (longer than SPEC §9's
  24 h minimum for any run the simulator allows). World outcomes map to the wire: accepted (including `partial` and
  `silent_no_effect`, which the device reports as `succeeded`) → `accepted` and `DRESULT succeeded`; `ack_lost` → no receipt
  and no result; interlock → `interlocked`; failed precondition → `permissive_not_met`; random reject → `not_ready`; invalid
  arguments, including a narrowed bound → `out_of_range`. Safe-stop commands are the node's `SafeStop` option; they skip the
  node's deadline, epoch and policy checks, bypass the ledger and are never queued; a node refuses to start when a safe-stop
  effector carries preconditions or an interlock (R6c-02). A command whose arguments pass the kind's bounds but not an
  instance's narrowed bound is refused `out_of_range` at execution, a second one-bit fact about the instance.
- **Lease fallback on host loss** (R7c-01): a simulated node ends its session, and the run, on host loss; the domain's own
  effector leases lapse in the world. Acting on each command's `fallback_on_lease_loss` comes with reconnection (P7d).
- **Wire faults (P7c)**: an outbound fault plan acts on sequenced records: `drop`, `duplicate`, `swap` (with the next record)
  and `garble` (a schema-invalid record keeping its `seq`).

## 2. Disposition of the review findings (REVIEW-2026-10-12 §2–§4)

| Item | Where it lands |
|---|---|
| D1 Sparkplug names | SPEC §1, §4. **Deferred**: `NDATA` and node-control metrics (run state changes are announced by rebirth for now) |
| D2 node dials host | SPEC §1 |
| D3 supervisory actions, `fallback_on_lease_loss`, heartbeat | SPEC §8, §9; domain-spec v0.2 §18 |
| D4 node-side gate | SPEC §9 (digest allow-list for the bench; plant limits envelope named). **Deferred**: the envelope's format (next goal, gateway) |
| D5 versioning | SPEC §3, §6 |
| M5 placement | Section 1 |
| P1 births | SPEC §4, §5; node P7b |
| P2 historical, current values, `bd_seq`, synthesised death | SPEC §4, §7; node P7b, reference host P7d |
| P3 sequences and lanes | SPEC §3, §7; node P7b. **Deferred**: a data-age watchdog (reference host P7d) |
| P4 freshness, re-check, flush, dedup | SPEC §9; node P7c |
| P5 `authority_epoch` | SPEC §9, §13 (gateway and host only); node P7c |
| P6 board | SPEC §1; domain-spec v0.2 §12; node P7b |
| P7 refusal mapping | SPEC §9, §13; node P7c |
| P8 commissioned safety facts | SPEC §5; reference host P7d; Agentic Stream next goal |
| P9 commissioning | SPEC §5; reference host P7d; Agentic Stream next goal |
| P10 quality mapping | SPEC §10; node P7b. **Deferred**: the host's situation-facts column (Agentic Stream, next goal) |
| P11 strings | SPEC §4, §10; fixtures; reference host P7d |
| P12 barrier and clock mode | SPEC §8; node P7d |
| P13 edge realism | SPEC §4, §7, §8. **Deferred**: NTP/PTP and a drift metric (gateway, next goal) |
| P14 host liveness | SPEC §8; node P7b |
| P15 capability digest | SPEC §5; node P7b |
| P16 alerts through `notify` | Next goal (Agentic Stream) |
| P17 contract supply chain | SPEC §11; this phase |
| P18 lying results, reference host | Fixtures; reference host P7d |
| P19 wire fault layer | Node P7c: split and corrupt frames, reset mid-command, reconnect storm are new work |
| P20 gateway as a safety point | Next goal (gateway) |
| P21 self-reported state | SPEC §4 |
| P22 death reasons | SPEC §4 |
| P23 batching and limits | SPEC §4, §12 |
| P24 experiment naming | The acceptance test is "assessment Experiment 02"; round-2 experiments are "round-2 #n" |
| Host liveness of a node over TCP | Next goal, with authentication (no node heartbeat in v1) |
| Reference-host inventory file | P7d defines its own commissioned-inventory file; not part of the contract |
