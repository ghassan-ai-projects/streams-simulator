# P7c review record — plant wire commands

- **Reviewed**: the P7c working-tree changes against `a28bd63` (2026-10-12)
- **Reviewer**: one independent Sonnet subagent, as a safety-minded protocol reviewer.
- **Outcome**: 2 High, 7 Medium, 7 Low. Fixed, documented, or carried.

| # | Sev | Finding | Disposition |
|---|---|---|---|
| 1 | High | Queued commands ran during missed barriers and at host loss | Fixed: the queue runs only after a met barrier; host loss drops it without receipts; `TestAQueuedCommandNeverRunsAfterTheHostIsLost` |
| 2 | High | A repeat of a queued key queued again (two executions, two terminal answers) and could overwrite the ledger | Fixed: a pending ledger entry; repeats wait for the one execution and all receive its answer; entries are settled or released, never overwritten; `TestARepeatOfAQueuedCommandWaitsForItsOneExecution` |
| 3 | Med | `fallback_on_lease_loss` not acted on | Carried (R7c-01, P7d): a simulated node's session ends with the run on host loss |
| 4 | Med | The world read lease arguments as seconds; `lease_ms` lapsed 1000× late | Fixed: a lease declares its `unit` (`s` or `ms`); `TestALeaseInMillisecondsLapsesOnTime` |
| 5 | Med | A safe stop's device preconditions were only documented | Fixed: a node refuses to start when a safe-stop effector has preconditions or an interlock; `TestASafeStopWithPreconditionsIsRefusedAtStart` |
| 6 | Med | A safe stop could be refused for idempotency; `not_before` unclear | Fixed: safe stops bypass the ledger (world id = `command_id`) and are never queued (SPEC §9) |
| 7 | Med | The key is not checked against the content | Accepted and documented (SPEC §9 "Repeats"): checking it would make the contract's `duplicate` case impossible |
| 8 | Med | A lost acknowledgement silenced every retry | Fixed: the real answer is stored; a retry receives it; `TestARetryLearnsTheAnswerWhoseAcknowledgementWasLost` |
| 9 | Med | A malformed command echoed an untrusted id | Fixed: only a wire-safe `command_id` is answered (same test) |
| 10 | Med | Exchange assertions matched any record | Fixed: expectations must be met in order by distinct records |
| 11 | Low | Boot checked before the deadline | Fixed: identify, freshness, boot, then epoch and policy |
| 12 | Low | Unbounded queue and `not_before` | Fixed: at most 1024 queued (`not_ready`); `not_before_ns` at most 24 h ahead (`out_of_range`) |
| 13 | Low | Swap within one batch | Carried (R7c-02) |
| 14 | Low | `lastCallReason` copies the call log | Accepted: refusals are rare; noted |
| 15 | Low | `failed` and `unknown` results never produced | Noted: the simulated device reports `succeeded` or nothing; `unknown` is the host's conclusion |
| 16 | Low | Replay test scope | The script covers a leased setpoint, a state query, a rebirth and a safe stop; the artifact replay checks the world |
| 17 | Low | A hidden narrowing could surface as `out_of_range` at execution | Noted as a second one-bit leak in the placement doc |
