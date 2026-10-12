# P7b evidence — plant wire node sessions

- **Date**: 2026-10-12
- **Change**: the `internal/plantwire` module (facade over `domain`, `app` and a `uds` edge; schemas embedded from a
  byte-identical copy of `contracts/plant-wire/schemas`). Contract: [SPEC](../../../contracts/plant-wire/SPEC.md) §1–§8;
  placement: [plant-wire-in-simulator.md](../plant-wire-in-simulator.md). Closes P14 (host liveness) at the node.
- **Not in this phase**: command execution, freshness and refusal mapping (P7c, until then a command is refused
  `not_ready`), wire perturbation (P7c), the reference host and the journal (P7d), CLI wiring (P7d).

## Decisions

| Question | Decision | Why |
|---|---|---|
| Layers | The app layer is single-threaded and reads no clock; the `uds` edge owns the reader goroutine and the barrier's wall-time cap through a `Link` | Domain and app stay deterministic and testable with a scripted host |
| Births | Kinds from the domain through an allow-list (channel name, data type, unit, range, resolution, enum values, writable; effector name, argument schema without descriptions, risk class); no domain description; devices from the run's entities; overrides computed against the base kind | No sealed field or author note can reach a birth; `writable` marks a channel whose state an effector assigns |
| Wire ids | A deterministic wire-safe id per entity (`/` → `:`), distinct on collision, with a reverse map | Five shipped domains use `/` in ids |
| Data source | `run.SetDeliveryObserver`: every delivered event after perturbation, before the adapter | The wire must not depend on the run's adapter |
| Current values | The last emitted value per channel; `null` with `warming_up` before the first sample (contract change) | A birth never shows a noise-free truth value |
| Boards and boots | A device on a domain board shares its board's boot; a device on no board is its own board; `boot_id` is `<board>-boot-1` | Deterministic; the board owns the boot |
| Slices | Advance, send the slice's data, then `NSYNC` with the last data `seq`; wait for `HSYNC`; host records take the slice end's world time | Data before its barrier; deterministic order |
| Host loss | Three consecutive missed barriers | SPEC §8 for a simulated node |
| Ordering | Host records in `seq` order; duplicates dropped; a gap ends the session | SPEC §7 (a gap cannot be reordered once the barrier collected the slice) |

## Tests

| Test | Proves |
|---|---|
| `TestASessionBirthsEveryDeviceBeforeItsDataAndSyncsEachSlice` | Every record valid against the contract; contiguous `seq`; `NBIRTH` first, `NDEATH` last; a `DBIRTH` before each device's data; one `NSYNC` per slice |
| `TestARefusedSessionEndsAfterItsBirth` | An environment refusal stops the node after `NBIRTH` |
| `TestARebirthRepeatsEveryBirthInTheSameSession` | `NBIRTH` with the next `bd_seq`, same session, then every `DBIRTH` |
| `TestAHostSequenceGapEndsTheSession`, `TestThreeMissedBarriersAreHostLoss`, `TestAHostThatNeverAnswersTheBirthIsLost` | Gap and loss end the session |
| `TestDuplicateHostRecordsAreDroppedAndTheEpochOnlyRises` | A duplicate is dropped; a lower epoch does not lower the floor |
| `TestNoWireRecordCarriesAFaultOrItsOnset`, `TestTheWireCarriesNoFaultFailureModeOrOnsetInAnyShippedDomain` | Truth canary over every shipped domain: no fault id, hidden failure mode or onset on the wire |
| `TestAFaultChangesNoBirth` | Births are byte-identical with and without a fault |
| `TestEveryShippedDomainSpeaksValidWire` | A session over each of the eight shipped domains sends only schema-valid records |
| `TestAMalformedCommandIsAnsweredAndTheSessionGoesOn`, `TestAGoodbyeThenHangUpEndsCleanlyAndAHangUpAloneIsHostLoss`, `TestALowerEpochGoodbyeIsIgnored`, `TestInvalidOptionsAreRefused` | Review regressions |
| `TestBirthsOfEveryDomainMatchTheContract` | Births of three domains (overrides, boards, system kinds) validate |
| `TestTheContractDigestIsThePinnedOne`, `TestTheCapabilityDigestMatchesTheContractFixture` | The node announces the pinned contract digest and derives the capability digest as the fixture does |
| `TestANodeServesAWholeSessionOverThePrivateSocket`, `TestASocketOthersCanReachIsRefused` | A real owner-only socket session; a shared socket or directory is refused |
| `FuzzWireParse` | Host record parsing never panics |

Coverage: facade 100 %, app 90 %, domain 88 %, uds 91 % (floors added). The P7b review (4 High) was resolved before the commit.

## Gates

- `make phase-gate`: passed in 1:37; behaviour pin unchanged.
