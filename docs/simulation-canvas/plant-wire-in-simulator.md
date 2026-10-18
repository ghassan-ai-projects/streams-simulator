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
  A run is a pure function of the host records as the node processed them, and so of the host's timing: a record that
  arrives after a barrier's wall cap is processed at the next slice end. The journal's `wire` and `pacing` records show
  which slice that was; the artifact's command log replays it exactly.
- **Typed records**: `domain.Decode` checks a line against its schema and its size (64 KiB, 1 MiB for births) once and
  reads it into a typed `Message`; an invalid record keeps only its `seq`, `type` and `command_id`, and a record without a
  usable `seq` consumes none. After the handshake, a record from another session or version is invalid.
- **Coexistence with the device path (R0-05)**: the device emulator, `deviceworld`, the vendored device-wire-v1 contract and the
  thermal LED/fan experiment stay unchanged and green in this goal. A run uses one host link: the device socket or the plant
  wire. Agentic Stream moves to the plant wire in the next goal, together with the gateway.
- **Truth canary**: P7b adds a test that injects a fault with a recognisable id and scans every wire record of the run for it.
- **Commands (P7c)**: a `DCMD` is checked in the SPEC §9 order at receipt (every check, then the idempotency ledger) and,
  when queued for `not_before_ns`, again at execution; due queued commands run at every slice end until host loss (three
  missed barriers), and commands that arrive with a lost connection are not run; it runs through `run.InvokeEffector` at the world time of the slice end that processed it, with its
  `idempotency_key` as the world command id. The node's idempotency ledger lives for the whole run (longer than SPEC §9's
  24 h minimum for any run the simulator allows). World outcomes map to the wire: accepted (including `partial` and
  `silent_no_effect`, which the device reports as `succeeded`) → `accepted` and `DRESULT succeeded`; `ack_lost` → no receipt
  and no result; interlock → `interlocked`; failed precondition → `permissive_not_met`; random reject → `not_ready`; invalid
  arguments, including a narrowed bound → `out_of_range`. Safe-stop commands are the node's `SafeStop` option, declared
  `safe_stop: true` in the `NBIRTH`; they are checked only for device, command and boot (one reason), skip the node's
  deadline, epoch and policy checks, bypass the ledger, are never queued, and act every time they arrive; a node refuses to start when a safe-stop
  effector carries preconditions or an interlock (R6c-02). A command whose arguments pass the kind's bounds but not an
  instance's narrowed bound is refused `out_of_range` at execution, a second one-bit fact about the instance. A retry of a
  command whose acknowledgement was lost learns `accepted` and `succeeded` (SPEC §9 idempotency), a deliberate third one:
  the MCP path answers `acknowledgement_timeout`. The node keeps at most 1024 queued commands, 65 536 ledger entries and 64
  waiting repeats of one key (beyond them `not_ready`), and reads at most 4096 host records per barrier (beyond them the
  host is lost). Argument schemas reach the `NBIRTH` with validation keywords only.
- **Lease fallback on host loss** (R7c-01): a simulated node ends its session, and the run, on host loss; the domain's own
  effector leases lapse in the world. Acting on each command's `fallback_on_lease_loss` comes with reconnection (P13).
- **Journal (P7d)**: the node writes every record it sends and receives as an observer `wire` session record, and one
  `pacing` session record per barrier (`through_seq`, the wall time waited, whether the cap expired). Both are outside the
  stream digest: the world is reproduced by the command log, the wire by replaying the host's records.
- **Reference host (P7d)**: `plantwire host` (module layer `host`, socket edge `uds/listen.go`): a socket published owner-only
  by rename into an owner-only directory (a stale socket is replaced, a live host or a file is not), one node that must send
  its `NBIRTH` within 10 s, schema-invalid records ignored (their `seq` consumed), nothing before an accepted `NBIRTH`,
  handshake checks (protocol range, environment, contract version and digest, version, every commissioned device present in
  the birth's boards) with refusal and then a close, a commissioned inventory (device to kind; others
  quarantined), one rebirth request per gap, historical data kept out of live values, scripted commands issued at the barrier
  whose world time reaches them (with the content-derived key, the epoch and its policy; denied for a device not accepted),
  verification of each succeeded command by an observation that arrives after its result, within a window (only `good`,
  current data of the device's born boot counts; a `DDEATH`, a current `NDEATH` or a rebirth as another kind removes the
  device), and a configurable delay before each `HSYNC`. It trusts no birth for
  permissions: it only issues what its script says.
- **Deferred to P13**: reconnection and the faults that need it (split frames, a reset mid-command, a reconnect storm; P19),
  and acting on `fallback_on_lease_loss` (R7c-01).
- **Wire faults (P7c)**: an outbound fault plan acts on sequenced records: `drop`, `duplicate`, `swap` (with the next record)
  and `garble` (a schema-invalid record keeping its `seq`).

- **Relay (R7b)**: `plantwire relay` is the simulator between an edge gateway and the real host (hardware only; the design is
  [PHYSICAL_DEVICES §2](../design/PHYSICAL_DEVICES.md)). It is the host of one gateway link and the node of one host link, and
  it is a different use case from the simulated node: no run, no world, no domain. The rules are the `plantwire` domain
  layer's (`relay*.go`, a pure state machine fed one record or event at a time that returns the records to send, the lines to
  journal and the links to close); the app layer runs the loop over a gateway link, a host link and a tick channel; the `uds`
  edge adds the gateway's listener (`gateways.go`, one owner-only socket, one link per connection) and the tick (`ticks.go`).
  The handshake checks are the reference host's, one predicate (`domain.Admission`) used by both. Records are carried with the
  content they arrived with; the relay owns only the envelope (`session_id`, `seq` assigned at transmission, the `NBIRTH`'s
  `bd_seq` and `prior_last_seq`). The barrier is mapped through a table of (gateway `seq`, relay `seq`) pairs: `NSYNC.through_seq`
  up, `HSYNC.through_seq` down, each to the last carried record at or below it. The journal is written through
  `run.WireTrace` (the run module owns the run folder), with the directions `gateway`, `node`, `host` and `relay`, so a field
  trace (R8) is the `gateway` lines. Nothing here acts on a hardware record: the relay's files may not name perturbation or
  wire fault code (`test/architecture`), and it has no option for either.

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
| P10 quality mapping | SPEC §10; node P7b sends `good` and `warming_up` only (R7d-01). **Deferred**: the host's situation-facts column (Agentic Stream, next goal) |
| P11 strings | SPEC §4, §10; fixtures; reference host P7d |
| P12 barrier and clock mode | SPEC §8; node P7d |
| P13 edge realism | SPEC §4, §7, §8. **Deferred**: NTP/PTP and a drift metric (gateway, next goal) |
| P14 host liveness | SPEC §8; node P7b |
| P15 capability digest | SPEC §5; node P7b |
| P16 alerts through `notify` | Next goal (Agentic Stream) |
| P17 contract supply chain | SPEC §11; this phase |
| P18 lying results, reference host | Fixtures; reference host P7d |
| P19 wire fault layer | Node P7c (drop, duplicate, swap, garble); split and corrupt frames, reset mid-command, reconnect storm: P13 |
| P20 gateway as a safety point | Next goal (gateway) |
| P21 self-reported state | SPEC §4 |
| P22 death reasons | SPEC §4; the node sends `NDEATH shutdown` only, no `DDEATH` yet (R7d-01) |
| P23 batching and limits | SPEC §4, §12 |
| P24 experiment naming | The acceptance test is "assessment Experiment 02"; round-2 experiments are "round-2 #n" |
| Host liveness of a node over TCP | Next goal, with authentication (no node heartbeat in v1) |
| Reference-host inventory file | P7d defines its own commissioned-inventory file; not part of the contract |
