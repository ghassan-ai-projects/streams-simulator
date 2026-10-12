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
