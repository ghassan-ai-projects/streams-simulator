# P7i evidence — world forms for plants

P7i lands in steps, one commit each. Domain-spec v0.2 §22 onwards holds the rules.

## Step 1 — prerequisites (R7g-01, R7e-04, R6a-01)

| Check | Result |
|---|---|
| An absent `noise_scale_at_full` keeps the channel noise; an explicit one scales it | `TestABiasWithoutANoiseScaleKeepsTheChannelNoise` |
| A state bound clamps only the side it names | `TestAStateBoundClampsOnlyTheSideItNames` |
| An assignment approaches its target from the value it replaces (natural, after an additive kick, after an earlier assignment, with a later kick, before it starts) | `TestAnAssignmentApproachesItsTargetFromTheValueItReplaces` |
| Lags, RC networks, regimes and dead times step within their stability bound; stores within their exchange bound, scale bounds included | `TestAnIntegratedStateMustStepWithinItsStabilityBound`, `TestEachIntegratedFormHasItsStabilityLimit` |
| A negative flow scale stops the flow | `TestANegativeScaleStateStopsAFlowInsteadOfReversingIt` |
| Id templates rendering a common id are refused (exact, both orders) | `TestTwoIDTemplatesOverlapWhenOneIDFitsBoth`, `TestAnInstanceSetMayNotRenderAnotherSetsIDs` |
| A lease argument cannot be negative | `TestALeaseArgumentMustNotAllowANegativeLease` |
| Shipped worlds unchanged | Behaviour pin: run, trace and replay hashes identical; only `catalog describe` output changed, because explicit zero bounds now survive the model (re-recorded with the v0.1 runtime equivalence digests) |
