# P7d evidence — reference host, journal, barrier, CLI

- **Date**: 2026-10-12
- **Change**: the reference host (`internal/plantwire/internal/host`, `uds/listen.go`, facade `plantwire.Host`), `wire` and
  `pacing` journal records (`internal/journal`, `run.JournalWire`, `run.JournalPacing`), CLI `plantwire node` and
  `plantwire host`. Closes P8, P12 and P18. Placement: [plant-wire-in-simulator.md](../plant-wire-in-simulator.md).
- **Deferred to P13**: reconnection with its faults (P19) and acting on `fallback_on_lease_loss` (R7c-01).

## Decisions

| Question | Decision | Why |
|---|---|---|
| Where the host lives | A `host` layer of the plantwire module with a listening socket edge | It shares the records, digests and sequencing; the facade only delegates |
| What the host trusts | Only its own inventory and script; births are claims (P8) | Births grant nothing (SPEC §5) |
| Verification | A succeeded command is verified when a named metric reaches the expected value within a window of world time | Receipt ≠ result ≠ observation (P18); catches a lying device |
| Barrier proof | A host that waits before each `HSYNC`: the node waits for it (pacing shows the wait, never a missed barrier) and the host's command lands at its slice | The quiescence barrier with a live host (P12) |
| Journal | `wire` (every record, both directions) and `pacing` (each barrier) as observer session records | They are inputs or wall-time facts, outside the replayable stream |
| CLI | Wiring only: flags, files, facades; the script and inventory are parsed by the module | `cli_wiring_test` |

## Tests

| Test | Proves |
|---|---|
| `TestASlowReferenceHostIsNeverOutrunAndVerifiesItsCommand` | A 30 ms host delay: no barrier missed, every wait recorded ≥ 30 ms, the scripted setpoint accepted, succeeded and verified |
| `TestTheReferenceHostCatchesALyingResult` | Under `silent_no_effect` the device reports success and the host's verification is false |
| `TestTheReferenceHostRefusesAnotherEnvironmentAndQuarantinesTheUncommissioned` | An environment mismatch is refused; a device outside the inventory is quarantined |
| `TestTwoRunsWithTheSameHostGiveTheSameWire` | The journal's wire records are byte-identical across two runs |
| Host unit tests | Handshake and contract refusal, one rebirth per gap, historical data, scripted timing, verification |
| CLI tests | `plantwire node` writes its run folder and matches its contract document; `plantwire host` reports quarantine; usage errors exit 2 |
| Device path | `internal/device`, `internal/deviceworld` and `test/acceptance` unchanged and green in the gate |

## Quality round (after the PR opened)

| Change | Why |
|---|---|
| Typed inbound records: `domain.Decode` validates a line against its schema once and reads it into `domain.Message`; the node and the host read typed fields | No handler type-asserts on wire data; an invalid record keeps only its `seq`, `type` and `command_id` |
| `TestTheContractsExchangesDriveTheReferenceHost` | The host is proven against the contract's eight host-side exchanges (session start, refusal, negotiation, gap and rebirth, the verified and three lying results), not only against this repository's node; each host `DCMD` must match the contract's record, including its content-derived `idempotency_key`. Breaking the verdict, the key, the environment check, the rebirth or the version makes it fail |
| `FuzzTheHostSurvivesAnyLine` | Any line, decoded and received, never panics the host (3.4 million executions locally) |
| A script step may carry its `id` | Commands are named as the contract names them; ids are unique and wire-safe |
| Coverage floors: host 91, uds 88, domain 91 | Encoding failures on both socket sides are tested |
| Three independent reviews (protocol, safety, determinism and tests) | 30 dispositions in [reviews/P7d.md](../reviews/P7d.md); R7d-01 to R7d-06 registered |

## Gates

- `make phase-gate`: see the commit; behaviour pin unchanged.
