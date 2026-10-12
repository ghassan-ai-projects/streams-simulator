# P7a review record — plant wire spec v2 and contract

- **Reviewed**: the P7a working-tree changes against `c3015fa` (2026-10-12)
- **Reviewer**: one independent Sonnet subagent, as an industrial protocol and safety reviewer, with validation probes on mutated
  records.
- **Outcome**: 9 High, 17 Medium, 1 Low. All High and most Medium fixed in the spec, schemas, fixtures and tests; the rest are
  explicit deferrals in [plant-wire-in-simulator.md](../plant-wire-in-simulator.md).

| # | Sev | Finding | Disposition |
|---|---|---|---|
| 1 | High | Queued commands were re-checked only for time | Fixed: every check runs again at execution; host loss and reboot flush the queue (SPEC §8, §9); `queued-across-reboot` exchange |
| 2 | High | Receipt lifecycle undefined; refused without reason accepted | Fixed: queued → one terminal receipt → one result (SPEC §9); `DRECEIPT` if/then/else; result mapping (SPEC §13); invalid fixtures for both directions |
| 3 | High | Optional safety members plus must-ignore let typos through; negative deadline | Fixed: `DCMD` is strict and requires `not_before_ns` and `fallback_on_lease_loss`; deadlines unsigned and at most 24 h ahead; misspelt-member and negative-deadline fixtures |
| 4 | High | Epoch floor could be lowered | Fixed: floor = highest epoch announced, lower `STATE` ignored, `DCMD` must equal it, gateway persists it; `stale-epoch` exchange covers a lower `STATE` |
| 5 | High | Gap rule against lane priority caused rebirth storms | Fixed: `seq` assigned at transmission, a bounded reorder window, duplicates dropped, one outstanding rebirth (SPEC §3, §6, §7) |
| 6 | High | Handshake, negotiation, refusal, rebirth undefined | Fixed: SPEC §6 (version 1 handshake envelope, `STATE.refusal`, no reconnect after refusal, session vs rebirth rules); `version-negotiation` and `environment-refusal` exchanges |
| 7 | High | Truth boundary unguarded (`origin`, quality from fault state, refusal classes) | Fixed: `origin` removed; quality and death only from observable models; refusal classes documented as one bit; truth canary test planned for P7b |
| 8 | High | Clocks unspecified | Fixed: SPEC §8 (time base, barrier wall cap, host-record world time, liveness per mode, deterministic ids); `now_ns` fixture step |
| 9 | High | Idempotency identity undefined | Fixed: content and key defined (SPEC §9); every fixture key is the real digest (tested); `duplicate-different-content` exchange |
| 10 | Med | Safe-stop semantics | Fixed: SPEC §9; `safe-stop-from-stale-epoch` exchange |
| 11 | Med | Authentication deferral and private socket missing | Fixed: SPEC §2 |
| 12 | Med | Narrow safety floor | Fixed: floor covers commands, births, data, state, receipts, results and `DCMD` strictness |
| 13 | Med | Closed enums vs must-ignore | Fixed: SPEC §3 (a new value is a new version) |
| 14 | Med | Folder not self-contained; drift partial | Fixed: simulator parts moved to `plant-wire-in-simulator.md`; drift compares the whole folder |
| 15 | Med | Inconsistent limits | Fixed: SPEC §12; boards bounded |
| 16 | Med | Integers above 2^53 | Fixed: SPEC §3 requires exact 64-bit parsing |
| 17 | Med | Births can widen; no birth time | Fixed: host rules in SPEC §5; `birth_time_ns` required |
| 18 | Med | Thin injection coverage; free-text ids | Fixed: identifier pattern (no spaces or prose), an id-with-prose invalid fixture, an argument-schema injection fixture |
| 19 | Med | Exchange gaps, no clock, negative-only verdicts, wrong alias | Fixed: nine new exchanges, `now_ns`, positive `verified-result`, setpoint alias corrected |
| 20 | Med | Capability digest underspecified; placeholder digests | Fixed: SPEC §5 definition; fixtures carry real contract and capability digests (tested against Go canonical JSON) |
| 21 | Med | device-wire-v1 field mapping missing | Fixed: SPEC §13 with the deadline conversion rule |
| 22 | Med | Reboot wording | Fixed: accepted iff `expected_boot_id` equals the current boot |
| 23 | Med | Partial closures | Fixed: explicit deferrals (data-age watchdog, NTP/PTP, envelope format, NDATA, situation facts) |
| 24 | Med | Lenient sequencing test | Fixed: contiguity from 0 per direction unless the exchange declares a gap |
| 25 | Med | `contract-sync` could wipe the folder | Fixed: the master must hold `SPEC.md`, `DIGEST`, `schemas`, `fixtures`; the drift test runs with the master |
| 26 | Low | Cosmetic | Fixed: fixture names, stricter id pattern, `NDEATH` reasons, refusal `protocol_invalid` |
| 27 | Low | No node heartbeat over TCP; reference-host inventory format | Deferred (next goal; P7d) |
