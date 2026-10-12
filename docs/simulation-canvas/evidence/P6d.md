# P6d evidence — supervisory control: setpoints, sequencer, responder, lease

- **Date**: 2026-10-12
- **Change**: conditions may compare against a setpoint state; the `sequencer` form; effector `responder` and `lease`.
  Spec: [`docs/design/DOMAIN_SPEC_V0_2.md`](../../design/DOMAIN_SPEC_V0_2.md) section 18. Closes review item S2 (with D3).
  `precondition` landed in P6c as effector `requires`.
- **Not in this phase**: responder availability, reports and approvals (S6), the air pressure band and generator of the
  reference scenario (S10), both P12; the host heartbeat of the plant wire (P7).

## Decisions

| Question | Decision | Why |
|---|---|---|
| Setpoints | `threshold_state` on a condition, `threshold` as an offset | Supervisory actions assign a state; trips, preconditions and sequencers follow it with no new mechanism |
| Sequencer | Stage count changes by one per boundary once `hold_s` has passed since the last change; start wins | The plant's own staging logic, closed data; the stage count is an ordinary input |
| Responder | Decided when the call is acknowledged from the faults known then and active at arrival; effect at arrival; may clear the fault | A fixed form with no free text; the outcome shows only through states (no report leaks the cause) |
| Lease | A pending fallback is withdrawn when the same effector is acknowledged again before the lapse | Renewal is an ordinary call; a refused or unacknowledged call does not renew |
| Safe state | `safe_off` sets each listed state to its `delta` | A safe state is a constant, whatever the call set |

## What was built

- Model: `Condition.ThresholdState`, sequencer members on `F1Dyn`, `Responder`, `Lease`; contract with `delta_list`.
- Domain layer: shared `condition` check; `validation_calls.go` (responder and lease).
- World: `sequencer.go`; `follow_up.go` (responder dispatch, lease arming, renewal and withdrawal); the form dispatch is two
  switches (vocabulary and protection).
- Fixture: an air system in `utility-plant.domain.json`: a pressure store fed by staged compressors and a leased boost, drawn
  down by a demand a leak raises; a supervisory pressure band; a work-order responder that clears the leak.

## Oracles

| Test | Independent derivation | Result |
|---|---|---|
| `TestASequencerStagesOncePerHoldWithinItsBand` | Stage times from the band crossings (effects strictly after their time) and `hold_s`: up at 102, 162, 222 s (capped at 3), down at 402, 462, 522 s | Exact |
| `TestAResponderActsOnTheFaultItIsToldAbout` | Wrong cause: `on_miss` at arrival, fault stays; right cause: `on_match` at arrival, fault cleared from arrival | Pass |
| `TestALeaseFallsBackUnlessRenewed` | Safe value and revert at the lapse; renewal moves the lapse | Pass |
| `TestSequencerResponderAndLeaseRulesAreLocated` | Each rule refused at its pointer | Pass |
| `TestVocabularyFixtureRunsAreByteReproducibleAndReplayToTheSameDigest` (air case) | Two runs byte-identical; replay matches | Pass |

End to end with the binary: `AIR1=air_leak@1200`, a pressure band set at 600 s, a work order naming the leak at 2400 s and a
leased boost at 4000 s: the sequencer stages up as the leak draws the pressure down, the leak clears at 3300 s, the pressure
stays around the band, replay matches. A first fixture tuning (receiver capacity 20) hunted between 0 and 12 bar; a realistic
capacity (600) keeps the hunting inside the band.

## Gates

- `make phase-gate`: passed in 1:40; behaviour pin unchanged.
- No comment lines added in Go.
