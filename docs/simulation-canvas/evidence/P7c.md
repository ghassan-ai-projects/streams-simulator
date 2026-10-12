# P7c evidence — plant wire commands

- **Date**: 2026-10-12
- **Change**: the node executes commands (`internal/plantwire/internal/{domain,app}`: `commands.go`, `execution.go`,
  `wire_faults.go`); contract fixtures' boot ids follow `<board>-boot-<n>`; a `pump-board` test domain matches the fixtures.
  Placement and mapping: [plant-wire-in-simulator.md](../plant-wire-in-simulator.md) §1. Closes P4 and R6c-02; P19 is
  partly done (the reconnect-dependent faults move to P7d).
- **Not in this phase**: acting on `fallback_on_lease_loss` (R7c-01), the reference host, the journal and CLI wiring (P7d); split frames, reset mid-command and
  reconnect storms (P7d, with reconnection); reboots (no world model of a board reboot yet).

## Decisions

| Question | Decision | Why |
|---|---|---|
| Check order | Safe stops apart; otherwise the ledger first (a repeat replays or waits for its queued original), then device, command, arguments, deadline, boot, epoch floor, policy | A retry gets its original answer even after its deadline; SPEC §9 order otherwise |
| Ledger | Pending while queued; settled once with the answer; released on a terminal refusal; never overwritten; a lost acknowledgement stores the real answer for retries | One key, one effect, one answer |
| Queued commands | Re-checked in full when due, at the slice end that reaches `not_before_ns`; refused terminally if any check now fails | SPEC §9: no stale or cross-epoch execution |
| Execution time | The slice end's world time; `idempotency_key` is the world command id | Deterministic; one key, one effect at both layers |
| Outcomes | The device's view: a `silent_no_effect` or `partial` call reports `succeeded`; `ack_lost` sends nothing | The benchmark needs results that can lie; verification is the host's job |
| Precondition vs random reject | The world's call log tells `precondition_refused` from a random reject | The node is part of the simulator; the operator surface stays generic |
| Wire faults | A plan of `{seq, kind}` applied after sequencing | SPEC pipeline amendment: faults show as gaps, repeats, swaps or invalid records |

## Tests

| Test | Proves |
|---|---|
| `TestTheContractsCommandExchangesDriveTheNode` | Six contract exchanges (expired, stale epoch, policy, duplicate, duplicate with other content, safe stop from a stale epoch) drive the real node and its receipts match |
| `TestAnAcceptedCommandActsOnTheWorldAndReportsItsResult` | Receipt then result |
| `TestEveryRefusalClassIsReportedByItsCheck` | Seven refusal classes, each from its check |
| `TestARepeatedKeyReplaysItsAnswerAndOtherContentIsADuplicate` | Idempotency by content |
| `TestAQueuedCommandRunsWhenDueAndIsCheckedAgain` | A due command runs; one due after an epoch rise is refused `stale_epoch` |
| `TestASafeStopSkipsEpochAndPolicyAndAPermissiveRefusalIsReported` | Safe stop from a stale epoch and another policy is accepted; a start during anti-recycle is `permissive_not_met` |
| `TestALostAcknowledgementIsSilentAndASilentNoEffectStillReportsSuccess` | Lying and silent devices |
| `TestReplayingTheHostRecordsReproducesTheWireAndTheWorld` | Replaying the recorded host records gives a byte-identical wire transcript; the run's artifact replays |
| `TestWireFaultsActOnSequencedRecords` | Drop, duplicate, swap and garble act on sequenced records |
| `TestAQueuedCommandNeverRunsAfterTheHostIsLost`, `TestARepeatOfAQueuedCommandWaitsForItsOneExecution`, `TestARetryLearnsTheAnswerWhoseAcknowledgementWasLost`, `TestASafeStopWithPreconditionsIsRefusedAtStart`, `TestALeaseInMillisecondsLapsesOnTime` | Review regressions |

## Gates

- `make phase-gate`: passed in 1:40; behaviour pin unchanged. The P7c review (2 High) was resolved before the commit.
