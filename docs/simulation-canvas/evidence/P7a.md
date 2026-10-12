# P7a evidence — plant wire spec v2 and contract

- **Date**: 2026-10-12
- **Change**: [`contracts/plant-wire/`](../../../contracts/plant-wire/SPEC.md): spec v2, fourteen record schemas, 64 fixtures (including 17 scripted exchanges),
  pinned `DIGEST`; `test/contract` proves the fixtures against the schemas, the digests and the safety-field floor;
  `make contract-sync` and `.env.example`. The design proposal
  [plant-wire-protocol.md](../plant-wire-protocol.md) is marked superseded.
- **Applied**: D1 (Sparkplug names: NBIRTH, DBIRTH, DDATA, DDEATH, NDEATH, NCMD, DCMD, STATE, plus DRECEIPT, DRESULT, DQUERY,
  DSTATE, NSYNC, HSYNC), D2 (node dials host), D3 (`fallback_on_lease_loss`, heartbeat), D4 (digests for the bench, envelope
  for plants), D5 (`protocol_min`/`protocol_max`, must-ignore), M5 (spec section 11: module, one effector path, perturbation
  after sequencing), P12 (virtual time with recorded host latency and a wall cap; `NSYNC`/`HSYNC` barrier), P5
  (`authority_epoch` host and gateway side only), R0-05 (coexistence with the device path), R0-13 (sync, pinned digest,
  safety floor). Every review item P1–P24 has a disposition in
  [plant-wire-in-simulator.md](../plant-wire-in-simulator.md), which also holds the simulator placement (M5). The P7a review
  (27 findings, 9 High) was resolved in the spec before the commit.

## Decisions

| Question | Decision | Why |
|---|---|---|
| Record names | Sparkplug B types plus six extensions | D1: a real binding stays possible |
| Sequencing | One sequence per direction for every record; lanes are transmit priority only | P3: a result never overtakes the observations before it |
| Must-ignore | Schemas allow unknown members; the safety floor test guards the required ones | D5 without letting a weakened master pass |
| Refusals on the wire | Classes (`permissive_not_met`, `interlocked`, `not_ready`, …), never the fault | Like a PLC: the class of refusal, not its cause |
| Clock mode | Virtual time; the node waits for `HSYNC` up to a wall cap and records the latency | The quiescence barrier holds with a live LLM host |
| Pinned digest | SHA-256 over file name and canonical JSON of each schema in name order | A schema change needs a reviewed `DIGEST` change |
| Where it lives | A self-contained folder with no reference to the rest of the repository | It moves to its own repository by copying the folder |
| Findings | Design fixes closed by the spec; enforcement moved to the phase that builds it (P4, R6c-02 to P7c; P14 to P7b; P8 to P7d) | The register tracks real work |

## Tests

| Test | Proves |
|---|---|
| `TestEveryValidPlantWireFixtureMatchesItsSchema` | One valid record per type, each accepted |
| `TestEveryInvalidPlantWireFixtureFailsAtTheMemberItNames` | 26 invalid records, each refused at the named member |
| `TestMustIgnoreAndInjectionFixturesAreValidRecords` | Unknown members and hostile strings still validate (handled by the receiver) |
| `TestExchangesHoldValidRecordsInContiguousSequencePerDirection` | 17 exchanges hold valid records with contiguous sequences per direction (gaps only where declared) |
| `TestFixturesCarryTheRealContractAndCapabilityDigests` | The birth fixture announces this folder's digest and the digest of its own command set |
| `TestEveryCommandFixtureCarriesTheIdempotencyKeyOfItsContent` | Every command's key is the digest of its content |
| `TestThePlantWireContractMatchesItsPinnedDigest` | The schemas match `DIGEST` |
| `TestTheSafetyFieldsCannotBeWeakened` | Commands, births, data, state, receipts and results keep their required safety members and values; `DCMD` stays strict |
| `TestTheVendoredContractMatchesItsMasterWhenOneIsConfigured` | The whole folder matches a configured master (skipped when none) |

## Gates

- `make phase-gate`: passed in 1:33; behaviour pin unchanged.
