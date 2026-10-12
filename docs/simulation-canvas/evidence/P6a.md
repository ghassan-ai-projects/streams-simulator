# P6a evidence — `store` and `flow`

- **Date**: 2026-10-12
- **Change**: the dynamics vocabulary gains `store` (an F1 form with a `capacity`) and kind-level `flows` between stores,
  integrated jointly per entity. Spec: [`docs/design/DOMAIN_SPEC_V0_2.md`](../../design/DOMAIN_SPEC_V0_2.md) sections 13 to 15.
  Closes findings R0-11, R2-05, R5b-01, R5b-02 and R5b-03. S14's remaining items (`lockout` naming, closed comparators) move
  to P6c, where those forms are built.
- **Not in this phase**: flows between instances along links, `regime`, `schedule`, `peak` (P6b), `trip` and `lockout` (P6c),
  the event clock (P6e).

## Decisions

| Question | Decision | Why |
|---|---|---|
| Form | `f1.form: "store"`, `capacity` required; level' = (Σ coef × input + inflows − outflows) / capacity | An accumulator is an integrator with a physical scale |
| Flow | `kinds[].flows[{name, from, to?, conductance, scale[]}]`; rate = conductance × Π scale × (from − to); no `to` drains toward zero | Linear conductance scaled by actuator and health states, as the vocabulary table states |
| Integration | Every integrated state of an entity advances together, earliest step first; stores step jointly with one RK4 over the vector of levels; each state keeps snapshots back to the entity's longest step; every store of a kind shares one `dt_ms` | A state read ahead returned its later value to an earlier reader, so results depended on read order (review finding 1, pre-existing for lag chains); the emitting world and the silent oracle world must agree |
| Hybrid rule | Inputs and scale states read at the step boundary and held for the step | The system is linear between steps |
| Faults | Act through scale states (a pipe health), not by overlaying a store level | A flow reads integrated levels; the health state is the modelled path |
| Scope | Flows join stores of one entity; an override may add a store state but not a flow | Smallest correct unit; cross-instance flow needs link aggregation |
| Closed grammar | A schema walk classifies every unconstrained string as prose, identifier or reference and fails on a new one; open objects limited to argument schemas, arguments, parameters and weights; no member named like an expression | R0-11: no expression language, enforced |
| Idempotency numbers | Argument numbers compared by value (`json.Number` → float64) | R2-05: `3` and `3.0` are one request |

## What was built

- `internal/model/flow.go` (`Flow`, `FormStore`, `Kind.IsStore`), `F1Dyn.Capacity`, `Kind.Flows`.
- Contract and embedded schema: `store` form, `capacity`, `flows`, `$defs/flow`.
- Domain layer: `validation_flows.go` (flow and shared-step rules), `override_detectors.go` (R5b-01), duplicate and empty
  narrowing (R5b-02); `domain format` orders `flows` after `dynamics`.
- World: `advance.go` (entity-wide earliest-first integration with step snapshots, for every F1 form), `stores.go` and
  `store_exchange.go` (joint RK4, partial-step probe, clamp to the form and the state bounds), `request_numbers.go` (R2-05),
  `firstTemplatedSet` (R5b-03).
- Fixture `internal/testsupport/testdata/transfer-plant.domain.json`: a feed fills an upper tank, a pumped pipe (scale
  `pump_speed × pipe_health`) transfers to a lower tank, an outlet drains it; `pipe_blockage` lowers the pipe health.

## Oracles

| Test | Independent derivation | Result |
|---|---|---|
| `TestTwoStoresJoinedByAFlowAgreeWithTheClosedForm` (world) | Conservation of C_a·a + C_b·b and exponential decay of a − b, swept over 2 × 2 × 2 × 2 parameters and 3 times | Within 1e-6 |
| `TestADrainedStoreWithAnInflowAgreesWithTheClosedForm` (world) | x∞ + (x0 − x∞)·e^(−kt/C) | Within 1e-6 relative |
| `TestCoupledStoresDoNotDependOnReadOrder` (world) | Same levels whichever store is read first, and with many intermediate reads | Bit-identical |
| `TestTheTransferWorldFollowsTheMatrixExponential` (truth) | 2×2 matrix exponential of the fixture's linear system with the feed | Within 1e-6 over 4 h |
| `TestTheSolverFindsTheClosedFormCrossingOfAPipeBlockage` (truth) | First 60 s grid time where the matrix-exponential gap in the lower level reaches 3σ and 10σ | Exact |
| `TestAStateFedByAnIntegratedStateDoesNotDependOnReadOrder` (world) | A lag fed by a store and a lag fed by a lag, read in three orders | Bit-identical |
| `TestAStoreLevelStaysWithinItsStateBounds` (world) | Overfilled store never commits above its state max | Pass |
| `TestVocabularyFixtureRunsAreByteReproducibleAndReplayToTheSameDigest` (run, transfer case) | Two runs byte-identical; replay digest matches | Pass |

End to end with the binary: `run` of the transfer plant with `cell-1=pipe_blockage@1800` emits 720 records; after the
blockage the upper level rises faster and the lower level falls; `replay` matches the digest.

## Gates

- `make phase-gate`: passed in 1:53; behaviour pin unchanged (no shipped domain uses a store, so `model.SimVersion` stays).
- No comment lines added in Go (`git diff HEAD -- '*.go' | grep '^+\s*//'` prints nothing).
