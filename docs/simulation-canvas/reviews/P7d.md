# P7d review record — reference host, journal, CLI

- **Reviewed**: the P7d working-tree changes against `58d189d` (2026-10-12)
- **Reviewer**: one independent Sonnet subagent, as a protocol and safety reviewer.
- **Outcome**: 2 High, 6 Medium, 4 Low. Fixed or documented; none carried.

| # | Sev | Finding | Disposition |
|---|---|---|---|
| 1 | High | A schema-invalid record (the node's own `garble` fault) reached the host's handlers and panicked it | Fixed: the host checks every record against its schema; an invalid one consumes its `seq` and is ignored; `TestAGarbledRecordIsIgnoredButConsumesItsSeq` |
| 2 | High | The receiver never resynchronised after a gap, so every rebirth raised another gap and swallowed a record | Fixed: a gap moves the receiver past it (SPEC §7, an ordered stream) and the record is still handled; `TestAGapAsksForOneRebirthAndTheRebirthCompletes` |
| 3 | Med | Verification passed on a value seen before the command | Fixed: only a sample that arrives after the result verifies; `TestAValueThatAlreadyMatchedDoesNotVerifyACommand` |
| 4 | Med | Commands went to quarantined or unborn devices | Fixed: not sent, reported `denied: not_commissioned`; an empty inventory accepts every device (documented) |
| 5 | Med | Handshake gaps: no `protocol_invalid`, records before the `NBIRTH` handled, the host kept answering after a refusal, the session version not negotiated | Fixed: `protocol_invalid` and a low node maximum are refused; nothing before an accepted `NBIRTH`; silence after a refusal; `UseVersion` of the negotiated version |
| 6 | Med | The socket was briefly reachable under a loose umask; a shared directory was accepted | Fixed: the socket's directory must be owner-only and owned (the node's own check); `TestAHostRefusesASharedDirectory` |
| 7 | Med | Cancellation ignored once connected; a read error ended the session silently; no interrupt handling | Fixed: the connection closes on cancellation, `scanner.Err` is returned, the host stops on interrupt or `SIGTERM` |
| 8 | Med | An unsorted script delayed later steps; steps not validated | Fixed: the script runs in `at_ns` order; `ParseScript` requires a device, a command and a positive window (`ErrInvalidStep`) |
| 9 | Low | A doc comment above the wrong function in `run/operations.go` | Fixed |
| 10 | Low | The journal schema required `type` and `seq` on wire records, which invalid host records lack | Fixed: `wire.record` is any object |
| 11 | Low | A hard-coded start time in the CLI; host unit tests used Go numbers instead of wire numbers | Fixed: `model.DefaultStartTimeNS`; the host tests build records from the contract's valid fixtures through `domain.Parse` |
| 12 | Low | CLI synopsis missed flags; the placement doc said the host closes on refusal | Fixed |

Checked and fine: wire and pacing records are session records outside the stream digest; journal calls are serialised by
the run's command lock; the host report marshals with sorted keys; wall time stays in the declared edges.

## Quality round (after the PR opened)

- **Reviewed**: the branch against `main` with the typed-records change, by three independent Sonnet subagents: protocol
  conformance against SPEC.md, safety under hostile peers, and determinism, truth boundary, architecture and test quality.
- **Outcome**: 2 High, 16 Medium, 20 Low. Fixed, or registered with a phase (R7d-01 to R7d-06).

| # | Sev | Finding | Disposition |
|---|---|---|---|
| Q1 | High | An empty session id panicked `WireID` | Fixed: an empty id becomes `d`, so `validate` refuses it; `TestSequencesAreReadStrictlyAndNeverWrap` |
| Q2 | High | A safe stop to an unknown device or command answered a concatenated, schema-invalid reason | Fixed: the first reason only; `TestASafeStopToAnUnknownDeviceOrCommandIsRefusedForOneReason` |
| Q3 | Med | The ledger ran before the other checks (a stale-epoch conflict answered `duplicate`; an expired retry replayed) | Fixed: SPEC §9 order; `TestEveryCheckRunsBeforeTheIdempotencyLedger` |
| Q4 | Med | Commands read just before the host hung up were executed | Fixed: on a lost connection only a goodbye `STATE` is honoured; `TestCommandsReadBeforeTheHostHangsUpAreNotExecuted` |
| Q5 | Med | A flooding host grew the node's memory and journal without bound | Fixed: at most 4096 records per barrier (`ErrFlooded`, host lost); `TestAFloodingHostIsCutOffAtOneBarrier` |
| Q6 | Med | A stale socket stopped the host from restarting; the socket briefly existed before it was owner-only | Fixed: published by rename after `chmod 0600`; a stale socket is replaced, a live host or a file is not; `TestAHostReplacesAStaleSocketButNotALiveHostOrAFile` |
| Q7 | Med | An idle peer pinned the host; a refused node kept the connection | Fixed: a 10 s handshake deadline, cleared once accepted; the host closes after a refusal; `TestAnIdlePeerCannotHoldTheHostPastTheHandshake`, `TestAHostThatIsDoneClosesTheConnection` |
| Q8 | Med | The `NBIRTH` declared every command `safe_stop: false` | Fixed: the node's safe stops are declared; `TestTheNodeBirthDeclaresItsSafeStops` |
| Q9 | Med | The host commanded quarantined, dead or re-qualified devices; ignored deaths | Fixed: `DDEATH`, a current `NDEATH` and a rebirth as another kind remove the device; `TestADeadOrRequalifiedDeviceGetsNoCommand`, `TestALateNodeDeathIsIgnored` |
| Q10 | Med | Commissioning checked only the kind; no `inventory_mismatch`, no contract version | Fixed: both refused at the handshake; `TestTheHandshakeChecksTheInventoryAndTheContractVersion` |
| Q11 | Med | The queue skipped missed barriers, so a slow host delayed due commands | Fixed: due commands run at every slice end until host loss (SPEC §8); `TestAQueuedCommandRunsAtAMissedBarrierButNeverAfterTheHostIsLost` |
| Q12 | Med | `permissive_not_met` read the world's call log (stale after a replay, O(n)) | Fixed: `InvokeResult.Cause` carries the refusal cause, never serialised |
| Q13 | Med | Behaviours with no test: queue and pending bounds, `not_before` 24 h, `missed: true`, safe-stop repeats | Fixed: `TestTheQueueAndAPendingKeyAreBounded`, `TestEveryCheckRunsBeforeTheIdempotencyLedger`, `TestAHostSlowerThanTheCapIsRecordedAsMissedBarriersThenLost`, `TestEverySafeStopActsEvenWhenItRepeats` |
| Q14 | Med | Wire and pacing records were proven outside the digest only at the plantwire layer | Fixed: `TestWireAndPacingRecordsLeaveTheReplayableStreamUnchanged` in the journal |
| Q15 | Med | No `DDEATH`, quality codes or reboot | Registered R7d-01 (P13): the world has nothing to map from yet |
| Q16 | Med | Host loss not in the command log; no synthesised `NDEATH` | Registered R7d-02 (P13, with reconnection) |
| Q17 | Med | The truth canary scanned few facts; argument schemas kept every non-`description` keyword | Fixed: validation keywords only; the canary also scans hidden state names, descriptions, `stresses` and schema annotations, and a test domain carries argument annotations (a leak of them fails it) |
| Q18 | Low | Seq wrap at the largest value; unparsable `seq` read as 0 | Fixed: the receiver stops at the largest seq; a record without a usable `seq` consumes none |
| Q19 | Low | A command whose content cannot be digested collided in the ledger | Fixed: refused `malformed` |
| Q20 | Low | No interrupt handling for the node | Fixed: both node and host stop on interrupt or `SIGTERM`; the node still writes its run folder |
| Q21 | Low | Close returned noise after a cancel; a reader error was reported once, later waits timed out | Fixed |
| Q22 | Low | Version negotiation accepted any common version; records of another session or version were accepted | Fixed: the highest common version only; other sessions and versions are invalid; `TestTheHostsVersionMustBeTheHighestBothSupport`, `TestARecordFromAnotherSessionIsMalformed` |
| Q23 | Low | No 64 KiB record limit | Fixed on receipt; send-side limits registered R7d-03 |
| Q24 | Low | The host sampled bad-quality birth values and data of another boot | Fixed; `TestOnlyGoodCurrentDataOfTheBornBootCounts` |
| Q25 | Low | Safe stops reused the command id as the world id, so a repeat did not act | Fixed: each safe stop has its own world id |
| Q26 | Low | Unbounded host device map and quarantine list | Fixed: at most 4096 devices |
| Q27 | Low | Wall-time assertions at the edge of their bounds; a vacuous pacing loop | Fixed: lower bounds with margin, the timer starts after the clock read, barrier counts asserted |
| Q28 | Low | No test that ok, partial and silent-no-effect answers are identical | Fixed: `TestAnAcceptedAPartialAndASilentNoEffectCommandLookTheSame` |
| Q29 | Low | Peer credentials not verified; no reference-host heartbeat; `DSTATE` fidelity | Registered R7d-04 (next goal), R7d-05 (next goal), R7d-06 (P13) |
| Q30 | Low | `pacing` journaled before the barrier's host records | Kept: the record closes the wait that delivered them |
